package executor

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/storage"
	"github.com/yyun543/minidb/internal/types"
)

// DataManager 管理表数据的存储和查询 (v2.0)
// 使用新的 StorageEngine 接口，通过 database/table 访问而非 key
type DataManager struct {
	catalog       *catalog.Catalog
	storageEngine storage.StorageEngine
	mu            sync.RWMutex
}

// NewDataManager 创建新的数据管理器 (v2.0)
func NewDataManager(catalog *catalog.Catalog) *DataManager {
	// Get the storage engine from catalog
	storageEngine := catalog.GetStorageEngine()

	return &DataManager{
		catalog:       catalog,
		storageEngine: storageEngine,
	}
}

// InsertData 插入数据到表中 (v2.0)
func (dm *DataManager) InsertData(dbName, tableName string, columns []string, values []interface{}) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 获取表元数据
	tableMeta, err := dm.catalog.GetTable(dbName, tableName)
	if err != nil {
		return fmt.Errorf("table not found: %w", err)
	}

	// 创建新记录
	newRecord, err := dm.createRecord(tableMeta.Schema, columns, values)
	if err != nil {
		return err
	}
	defer newRecord.Release()

	// 使用 StorageEngine.Write 写入数据
	ctx := context.Background()
	err = dm.storageEngine.Write(ctx, dbName, tableName, newRecord)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

// GetTableData 获取表的所有数据 (v2.0)
func (dm *DataManager) GetTableData(dbName, tableName string) ([]*types.Batch, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// 特殊处理系统表：支持 "sys.table" 或直接 dbName="sys"
	if dbName == "sys" {
		return dm.getSystemTableData(tableName)
	}
	// 兼容处理：如果tableName格式为 "sys.xxx"，解析并调用系统表处理
	if strings.HasPrefix(tableName, "sys.") {
		sysTable := strings.TrimPrefix(tableName, "sys.")
		return dm.getSystemTableData(sysTable)
	}

	// 使用 StorageEngine.Scan 读取数据
	ctx := context.Background()
	iter, err := dm.storageEngine.Scan(ctx, dbName, tableName, []storage.Filter{})
	if err != nil {
		return nil, fmt.Errorf("failed to scan table: %w", err)
	}
	defer iter.Close()

	// 收集所有批次
	var batches []*types.Batch
	for iter.Next() {
		record := iter.Record()
		if record.NumRows() > 0 {
			batch := types.NewBatch(record)
			batches = append(batches, batch)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("iterator error: %w", err)
	}

	return batches, nil
}

// getSystemTableData 获取系统表数据
func (dm *DataManager) getSystemTableData(tableName string) ([]*types.Batch, error) {
	switch tableName {
	case "schemata":
		return dm.getSchemataData()
	case "table_catalog":
		return dm.getTableCatalogData()
	case "columns":
		return dm.getColumnsData()
	case "index_metadata":
		return dm.getIndexMetadataData()
	case "delta_log":
		return dm.getDeltaLogData()
	case "table_files":
		return dm.getTableFilesData()
	default:
		return nil, fmt.Errorf("unknown system table: %s", tableName)
	}
}

// getSchemataData 获取schemata系统表数据（数据库列表）
func (dm *DataManager) getSchemataData() ([]*types.Batch, error) {
	// 获取所有数据库
	databases, err := dm.catalog.GetAllDatabases()
	if err != nil {
		return nil, err
	}

	// 确保系统数据库存在
	systemDbs := []string{"sys", "default"}
	dbSet := make(map[string]bool)

	for _, db := range databases {
		dbSet[db] = true
	}

	for _, sysDb := range systemDbs {
		if !dbSet[sysDb] {
			databases = append(databases, sysDb)
		}
	}

	// 创建schemata表的schema
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "schema_name", Type: arrow.BinaryTypes.String},
	}, nil)

	// 创建Arrow记录
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	nameBuilder := builder.Field(0).(*array.StringBuilder)

	for _, dbName := range databases {
		nameBuilder.Append(dbName)
	}

	record := builder.NewRecord()
	defer record.Release()

	batch := types.NewBatch(record)
	return []*types.Batch{batch}, nil
}

// getTableCatalogData 获取table_catalog系统表数据（表列表）
func (dm *DataManager) getTableCatalogData() ([]*types.Batch, error) {
	// 创建table_catalog表的schema
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "table_schema", Type: arrow.BinaryTypes.String},
		{Name: "table_name", Type: arrow.BinaryTypes.String},
	}, nil)

	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	schemaBuilder := builder.Field(0).(*array.StringBuilder)
	nameBuilder := builder.Field(1).(*array.StringBuilder)

	// 添加系统表
	systemTables := []struct {
		schema string
		table  string
	}{
		{"sys", "schemata"},
		{"sys", "table_catalog"},
		{"sys", "columns"},
		{"sys", "index_metadata"},
		{"sys", "delta_log"},
		{"sys", "table_files"},
	}

	for _, sysTable := range systemTables {
		schemaBuilder.Append(sysTable.schema)
		nameBuilder.Append(sysTable.table)
	}

	// 获取所有数据库
	databases, err := dm.catalog.GetAllDatabases()
	if err != nil {
		return nil, err
	}

	// 确保系统数据库存在
	systemDbs := []string{"sys", "default"}
	dbSet := make(map[string]bool)
	for _, db := range databases {
		dbSet[db] = true
	}
	for _, sysDb := range systemDbs {
		if !dbSet[sysDb] {
			databases = append(databases, sysDb)
		}
	}

	// 为每个数据库获取表列表
	for _, dbName := range databases {
		tables, err := dm.catalog.GetAllTables(dbName)
		if err != nil {
			continue
		}

		// 系统表已经在上面添加过了，这里跳过
		systemTableSet := map[string]bool{
			"schemata": true, "table_catalog": true, "columns": true,
			"index_metadata": true, "delta_log": true, "table_files": true,
		}

		for _, tableName := range tables {
			if dbName == "sys" && systemTableSet[tableName] {
				continue
			}
			schemaBuilder.Append(dbName)
			nameBuilder.Append(tableName)
		}
	}

	record := builder.NewRecord()
	defer record.Release()

	batch := types.NewBatch(record)
	return []*types.Batch{batch}, nil
}

// getColumnsData 获取columns系统表数据（所有表的列信息）
func (dm *DataManager) getColumnsData() ([]*types.Batch, error) {
	// 创建columns表的schema
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "table_schema", Type: arrow.BinaryTypes.String},
		{Name: "table_name", Type: arrow.BinaryTypes.String},
		{Name: "column_name", Type: arrow.BinaryTypes.String},
		{Name: "ordinal_position", Type: arrow.PrimitiveTypes.Int64},
		{Name: "data_type", Type: arrow.BinaryTypes.String},
		{Name: "is_nullable", Type: arrow.BinaryTypes.String},
	}, nil)

	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	schemaBuilder := builder.Field(0).(*array.StringBuilder)
	tableBuilder := builder.Field(1).(*array.StringBuilder)
	columnBuilder := builder.Field(2).(*array.StringBuilder)
	positionBuilder := builder.Field(3).(*array.Int64Builder)
	typeBuilder := builder.Field(4).(*array.StringBuilder)
	nullableBuilder := builder.Field(5).(*array.StringBuilder)

	// 获取所有数据库
	databases, err := dm.catalog.GetAllDatabases()
	if err != nil {
		return nil, err
	}

	// 为每个数据库获取表列表和列信息
	for _, dbName := range databases {
		tables, err := dm.catalog.GetAllTables(dbName)
		if err != nil {
			continue
		}

		for _, tableName := range tables {
			tableMeta, err := dm.catalog.GetTable(dbName, tableName)
			if err != nil {
				continue
			}

			// 遍历表的列
			for i, field := range tableMeta.Schema.Fields() {
				schemaBuilder.Append(dbName)
				tableBuilder.Append(tableName)
				columnBuilder.Append(field.Name)
				positionBuilder.Append(int64(i + 1))
				typeBuilder.Append(field.Type.String())
				if field.Nullable {
					nullableBuilder.Append("YES")
				} else {
					nullableBuilder.Append("NO")
				}
			}
		}
	}

	record := builder.NewRecord()
	defer record.Release()

	batch := types.NewBatch(record)
	return []*types.Batch{batch}, nil
}

// getIndexesData 获取indexes系统表数据（所有索引信息）
func (dm *DataManager) getIndexMetadataData() ([]*types.Batch, error) {
	// 创建indexes表的schema
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "table_schema", Type: arrow.BinaryTypes.String},
		{Name: "table_name", Type: arrow.BinaryTypes.String},
		{Name: "index_name", Type: arrow.BinaryTypes.String},
		{Name: "index_type", Type: arrow.BinaryTypes.String},
		{Name: "column_name", Type: arrow.BinaryTypes.String},
		{Name: "is_unique", Type: arrow.BinaryTypes.String},
	}, nil)

	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	schemaBuilder := builder.Field(0).(*array.StringBuilder)
	tableBuilder := builder.Field(1).(*array.StringBuilder)
	indexNameBuilder := builder.Field(2).(*array.StringBuilder)
	indexTypeBuilder := builder.Field(3).(*array.StringBuilder)
	columnBuilder := builder.Field(4).(*array.StringBuilder)
	uniqueBuilder := builder.Field(5).(*array.StringBuilder)

	// 获取所有数据库
	databases, err := dm.catalog.GetAllDatabases()
	if err != nil {
		return nil, err
	}

	// 为每个数据库获取索引信息
	for _, dbName := range databases {
		tables, err := dm.catalog.GetAllTables(dbName)
		if err != nil {
			continue
		}

		for _, tableName := range tables {
			indexes, err := dm.catalog.GetAllIndexes(dbName, tableName)
			if err != nil {
				continue
			}

			for _, indexInfo := range indexes {
				// 每个索引可能有多个列
				for _, columnName := range indexInfo.Columns {
					schemaBuilder.Append(dbName)
					tableBuilder.Append(indexInfo.Table)
					indexNameBuilder.Append(indexInfo.Name)
					indexTypeBuilder.Append(indexInfo.IndexType)
					columnBuilder.Append(columnName)
					if indexInfo.IsUnique {
						uniqueBuilder.Append("YES")
					} else {
						uniqueBuilder.Append("NO")
					}
				}
			}
		}
	}

	record := builder.NewRecord()
	defer record.Release()

	batch := types.NewBatch(record)
	return []*types.Batch{batch}, nil
}

// getDeltaLogData 获取delta_log系统表数据（Delta Log版本历史）
func (dm *DataManager) getDeltaLogData() ([]*types.Batch, error) {
	// 创建delta_log表的schema
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "table_schema", Type: arrow.BinaryTypes.String},
		{Name: "table_name", Type: arrow.BinaryTypes.String},
		{Name: "version", Type: arrow.PrimitiveTypes.Int64},
		{Name: "operation", Type: arrow.BinaryTypes.String},
		{Name: "file_path", Type: arrow.BinaryTypes.String},
		{Name: "row_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "file_size", Type: arrow.PrimitiveTypes.Int64},
	}, nil)

	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	schemaBuilder := builder.Field(0).(*array.StringBuilder)
	tableBuilder := builder.Field(1).(*array.StringBuilder)
	versionBuilder := builder.Field(2).(*array.Int64Builder)
	operationBuilder := builder.Field(3).(*array.StringBuilder)
	filePathBuilder := builder.Field(4).(*array.StringBuilder)
	rowCountBuilder := builder.Field(5).(*array.Int64Builder)
	fileSizeBuilder := builder.Field(6).(*array.Int64Builder)

	// 从 Storage Engine 获取 Delta Log (使用类型断言访问 ParquetEngine)
	if pe, ok := dm.storageEngine.(*storage.ParquetEngine); ok {
		deltaLog := pe.GetDeltaLog()
		if deltaLog != nil {
			// 获取所有日志条目
			entries := deltaLog.GetAllEntries()

			// 填充数据
			for _, entry := range entries {
				// Parse table_id: "database.table"
				parts := strings.Split(entry.TableID, ".")
				var dbName, tableName string
				if len(parts) == 2 {
					dbName = parts[0]
					tableName = parts[1]
				} else {
					dbName = "default"
					tableName = entry.TableID
				}

				schemaBuilder.Append(dbName)
				tableBuilder.Append(tableName)
				versionBuilder.Append(entry.Version)
				operationBuilder.Append(string(entry.Operation))
				filePathBuilder.Append(entry.FilePath)
				rowCountBuilder.Append(entry.RowCount)
				fileSizeBuilder.Append(entry.FileSize)
			}
		}
	}

	record := builder.NewRecord()
	defer record.Release()

	batch := types.NewBatch(record)
	return []*types.Batch{batch}, nil
}

// getTableFilesData 获取table_files系统表数据（表的物理文件信息）
func (dm *DataManager) getTableFilesData() ([]*types.Batch, error) {
	// 创建table_files表的schema
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "table_schema", Type: arrow.BinaryTypes.String},
		{Name: "table_name", Type: arrow.BinaryTypes.String},
		{Name: "file_path", Type: arrow.BinaryTypes.String},
		{Name: "file_size", Type: arrow.PrimitiveTypes.Int64},
		{Name: "row_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "status", Type: arrow.BinaryTypes.String},
	}, nil)

	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	schemaBuilder := builder.Field(0).(*array.StringBuilder)
	tableBuilder := builder.Field(1).(*array.StringBuilder)
	filePathBuilder := builder.Field(2).(*array.StringBuilder)
	fileSizeBuilder := builder.Field(3).(*array.Int64Builder)
	rowCountBuilder := builder.Field(4).(*array.Int64Builder)
	statusBuilder := builder.Field(5).(*array.StringBuilder)

	// 从 Storage Engine 获取 Delta Log (使用类型断言访问 ParquetEngine)
	if pe, ok := dm.storageEngine.(*storage.ParquetEngine); ok {
		deltaLog := pe.GetDeltaLog()
		if deltaLog != nil {
			// 获取所有表
			tables := deltaLog.ListTables()

			// 为每个表获取活跃文件
			for _, tableID := range tables {
				// Parse table_id: "database.table"
				parts := strings.Split(tableID, ".")
				var dbName, tableName string
				if len(parts) == 2 {
					dbName = parts[0]
					tableName = parts[1]
				} else {
					dbName = "default"
					tableName = tableID
				}

				// 获取最新快照以获取活跃文件
				snapshot, err := deltaLog.GetSnapshot(tableID, -1)
				if err != nil {
					continue
				}

				// 填充活跃文件信息
				for _, file := range snapshot.Files {
					schemaBuilder.Append(dbName)
					tableBuilder.Append(tableName)
					filePathBuilder.Append(file.Path)
					fileSizeBuilder.Append(file.Size)
					rowCountBuilder.Append(file.RowCount)
					statusBuilder.Append("ACTIVE")
				}
			}
		}
	}

	record := builder.NewRecord()
	defer record.Release()

	batch := types.NewBatch(record)
	return []*types.Batch{batch}, nil
}

// UpdateData 更新表中的数据 (v2.0) - DEPRECATED
// Use UpdateDataWithFilters instead
func (dm *DataManager) UpdateData(dbName, tableName string, assignments map[string]interface{}, whereCondition func(arrow.Record, int) bool) error {
	// Fallback to UpdateDataWithFilters with empty filters
	return dm.UpdateDataWithFilters(dbName, tableName, assignments, []storage.Filter{})
}

// UpdateDataWithFilters 使用storage.Filter更新表中的数据 (v2.0)
func (dm *DataManager) UpdateDataWithFilters(dbName, tableName string, assignments map[string]interface{}, filters []storage.Filter) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Get table metadata
	_, err := dm.catalog.GetTable(dbName, tableName)
	if err != nil {
		return fmt.Errorf("table not found: %w", err)
	}

	// Use the storage engine's Update method (Copy-on-Write)
	ctx := context.Background()
	updatedCount, err := dm.storageEngine.Update(ctx, dbName, tableName, filters, assignments)
	if err != nil {
		return fmt.Errorf("failed to update table: %w", err)
	}

	_ = updatedCount // Successfully updated

	return nil
}

// DeleteData 删除表中的数据 (v2.0) - DEPRECATED
// Use DeleteDataWithFilters instead
func (dm *DataManager) DeleteData(dbName, tableName string, whereCondition func(arrow.Record, int) bool) error {
	// Fallback to DeleteDataWithFilters with empty filters
	return dm.DeleteDataWithFilters(dbName, tableName, []storage.Filter{})
}

// DeleteDataWithFilters 使用storage.Filter删除表中的数据 (v2.0)
func (dm *DataManager) DeleteDataWithFilters(dbName, tableName string, filters []storage.Filter) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Get table metadata
	_, err := dm.catalog.GetTable(dbName, tableName)
	if err != nil {
		return fmt.Errorf("table not found: %w", err)
	}

	// Use the storage engine's Delete method (Delta Log integration)
	ctx := context.Background()
	deletedCount, err := dm.storageEngine.Delete(ctx, dbName, tableName, filters)
	if err != nil {
		return fmt.Errorf("failed to delete from table: %w", err)
	}

	_ = deletedCount // Successfully deleted

	return nil
}

// createRecord 创建新记录
func (dm *DataManager) createRecord(schema *arrow.Schema, columns []string, values []interface{}) (arrow.Record, error) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 创建列名到索引的映射
	fieldMap := make(map[string]int)
	for i, field := range schema.Fields() {
		fieldMap[field.Name] = i
	}

	// 为每个字段准备值
	fieldValues := make([]interface{}, len(schema.Fields()))
	for i := range fieldValues {
		fieldValues[i] = nil
	}

	// 设置实际提供的值
	for i, column := range columns {
		if fieldIdx, exists := fieldMap[column]; exists {
			if i < len(values) {
				fieldValues[fieldIdx] = values[i]
			}
		}
	}

	// 为所有字段添加值
	for i, value := range fieldValues {
		field := builder.Field(i)
		err := dm.appendValue(field, value)
		if err != nil {
			return nil, err
		}
	}

	return builder.NewRecord(), nil
}

// updateRecord 更新记录
func (dm *DataManager) updateRecord(oldRecord arrow.Record, assignments map[string]interface{}, whereCondition func(arrow.Record, int) bool) (arrow.Record, error) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, oldRecord.Schema())
	defer builder.Release()

	// 创建字段名到索引的映射
	fieldMap := make(map[string]int)
	for i, field := range oldRecord.Schema().Fields() {
		fieldMap[field.Name] = i
	}

	// 处理每一行
	for rowIdx := int64(0); rowIdx < oldRecord.NumRows(); rowIdx++ {
		shouldUpdate := whereCondition == nil || whereCondition(oldRecord, int(rowIdx))

		for colIdx := int64(0); colIdx < oldRecord.NumCols(); colIdx++ {
			field := builder.Field(int(colIdx))
			fieldName := oldRecord.Schema().Field(int(colIdx)).Name

			var value interface{}
			if shouldUpdate && assignments[fieldName] != nil {
				value = assignments[fieldName]
			} else {
				oldCol := oldRecord.Column(int(colIdx))
				value = dm.getValueFromArray(oldCol, int(rowIdx))
			}

			err := dm.appendValue(field, value)
			if err != nil {
				return nil, err
			}
		}
	}

	return builder.NewRecord(), nil
}

// getValueFromArray 从Array中获取值
func (dm *DataManager) getValueFromArray(arr arrow.Array, index int) interface{} {
	switch arr := arr.(type) {
	case *array.Int64:
		return arr.Value(index)
	case *array.Float64:
		return arr.Value(index)
	case *array.String:
		return arr.Value(index)
	case *array.Boolean:
		return arr.Value(index)
	default:
		return nil
	}
}

// appendValue 向字段添加值
func (dm *DataManager) appendValue(field array.Builder, value interface{}) error {
	switch field := field.(type) {
	case *array.Int64Builder:
		if v, ok := value.(int64); ok {
			field.Append(v)
		} else if v, ok := value.(int); ok {
			field.Append(int64(v))
		} else {
			field.AppendNull()
		}
	case *array.Float64Builder:
		if v, ok := value.(float64); ok {
			field.Append(v)
		} else {
			field.AppendNull()
		}
	case *array.StringBuilder:
		if v, ok := value.(string); ok {
			field.Append(v)
		} else {
			field.AppendNull()
		}
	case *array.BooleanBuilder:
		if v, ok := value.(bool); ok {
			field.Append(v)
		} else {
			field.AppendNull()
		}
	default:
		return fmt.Errorf("unsupported field type")
	}
	return nil
}

package executor

import (
	"fmt"
	"sync"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/storage"
	"github.com/yyun543/minidb/internal/types"
)

// DataManager 管理表数据的存储和查询
type DataManager struct {
	catalog    *catalog.Catalog
	engine     storage.Engine
	keyManager *storage.KeyManager
	mu         sync.RWMutex
}

// NewDataManager 创建新的数据管理器
func NewDataManager(catalog *catalog.Catalog) *DataManager {
	return &DataManager{
		catalog:    catalog,
		engine:     catalog.GetEngine(),
		keyManager: storage.NewKeyManager(),
	}
}

// InsertData 插入数据到表中
func (dm *DataManager) InsertData(dbName, tableName string, columns []string, values []interface{}) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 获取表元数据
	tableMeta, err := dm.catalog.GetTable(dbName, tableName)
	if err != nil {
		return fmt.Errorf("table not found: %w", err)
	}

	// 获取现有数据
	chunks, err := dm.getTableChunks(dbName, tableName, tableMeta.ChunkCount)
	if err != nil {
		return err
	}

	// 创建新记录
	newRecord, err := dm.createRecord(tableMeta.Schema, columns, values)
	if err != nil {
		return err
	}
	defer newRecord.Release()

	// 如果有现有数据，合并记录
	var finalRecord arrow.Record
	if len(chunks) > 0 && chunks[0] != nil {
		finalRecord, err = dm.mergeRecords(chunks[0].Record(), newRecord)
		if err != nil {
			return err
		}
		defer finalRecord.Release()
	} else {
		newRecord.Retain()
		finalRecord = newRecord
	}

	// 创建Chunk并存储
	chunk := types.NewChunk(finalRecord, 0, int64(finalRecord.NumRows()))
	defer chunk.Release()

	key := dm.keyManager.TableChunkKey(dbName, tableName, 0)
	record := chunk.Record()
	err = dm.engine.Put(key, &record)
	if err != nil {
		return err
	}

	// 更新表元数据中的ChunkCount（如果需要）
	if tableMeta.ChunkCount == 0 {
		// 更新catalog中的表信息
		tableMeta.ChunkCount = 1
		err := dm.catalog.UpdateTable(dbName, tableMeta)
		if err != nil {
			return fmt.Errorf("failed to update table metadata: %w", err)
		}
	}

	return nil
}

// GetTableData 获取表的所有数据
func (dm *DataManager) GetTableData(dbName, tableName string) ([]*types.Batch, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// 特殊处理系统表
	if dbName == "sys" {
		return dm.getSystemTableData(tableName)
	}

	// 获取表元数据
	tableMeta, err := dm.catalog.GetTable(dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("table not found: %w", err)
	}

	// 获取所有数据块
	chunks, err := dm.getTableChunks(dbName, tableName, tableMeta.ChunkCount)
	if err != nil {
		return nil, err
	}

	// 转换为Batch
	var batches []*types.Batch
	for _, chunk := range chunks {
		if chunk != nil && chunk.NumRows() > 0 {
			batch := types.NewBatch(chunk.Record())
			batches = append(batches, batch)
		}
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

	// 确保系统数据库存在（防御性编程）
	systemDbs := []string{"sys", "default"}
	dbSet := make(map[string]bool)

	// 将现有数据库加入集合
	for _, db := range databases {
		dbSet[db] = true
	}

	// 添加缺失的系统数据库
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

	// 添加所有数据库名称
	for _, dbName := range databases {
		nameBuilder.Append(dbName)
	}

	record := builder.NewRecord()
	defer record.Release()

	// 转换为Batch
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

	// 创建Arrow记录
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	schemaBuilder := builder.Field(0).(*array.StringBuilder)
	nameBuilder := builder.Field(1).(*array.StringBuilder)

	// 首先添加系统表（确保它们总是出现）
	systemTables := []struct {
		schema string
		table  string
	}{
		{"sys", "schemata"},
		{"sys", "table_catalog"},
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

	// 确保系统数据库存在（防御性编程）
	systemDbs := []string{"sys", "default"}
	dbSet := make(map[string]bool)

	// 将现有数据库加入集合
	for _, db := range databases {
		dbSet[db] = true
	}

	// 添加缺失的系统数据库
	for _, sysDb := range systemDbs {
		if !dbSet[sysDb] {
			databases = append(databases, sysDb)
		}
	}

	// 为每个数据库获取表列表
	for _, dbName := range databases {
		tables, err := dm.catalog.GetAllTables(dbName)
		if err != nil {
			continue // 跳过获取失败的数据库
		}

		// 跳过系统表（已经在上面添加过了）
		for _, tableName := range tables {
			if dbName == "sys" && (tableName == "schemata" || tableName == "table_catalog") {
				continue // 跳过已添加的系统表
			}
			schemaBuilder.Append(dbName)
			nameBuilder.Append(tableName)
		}
	}

	record := builder.NewRecord()
	defer record.Release()

	// 转换为Batch
	batch := types.NewBatch(record)
	return []*types.Batch{batch}, nil
}

// UpdateData 更新表中的数据
func (dm *DataManager) UpdateData(dbName, tableName string, assignments map[string]interface{}, whereCondition func(arrow.Record, int) bool) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 获取表元数据
	tableMeta, err := dm.catalog.GetTable(dbName, tableName)
	if err != nil {
		return fmt.Errorf("table not found: %w", err)
	}

	// 获取现有数据
	chunks, err := dm.getTableChunks(dbName, tableName, tableMeta.ChunkCount)
	if err != nil {
		return err
	}

	if len(chunks) == 0 || chunks[0] == nil {
		return nil // 没有数据可更新
	}

	oldRecord := chunks[0].Record()
	newRecord, err := dm.updateRecord(oldRecord, assignments, whereCondition)
	if err != nil {
		return err
	}
	defer newRecord.Release()

	// 存储更新后的数据
	chunk := types.NewChunk(newRecord, 0, int64(newRecord.NumRows()))
	defer chunk.Release()

	key := dm.keyManager.TableChunkKey(dbName, tableName, 0)
	record := chunk.Record()
	return dm.engine.Put(key, &record)
}

// DeleteData 删除表中的数据
func (dm *DataManager) DeleteData(dbName, tableName string, whereCondition func(arrow.Record, int) bool) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 获取表元数据
	tableMeta, err := dm.catalog.GetTable(dbName, tableName)
	if err != nil {
		return fmt.Errorf("table not found: %w", err)
	}

	// 获取现有数据
	chunks, err := dm.getTableChunks(dbName, tableName, tableMeta.ChunkCount)
	if err != nil {
		return err
	}

	if len(chunks) == 0 || chunks[0] == nil {
		return nil // 没有数据可删除
	}

	oldRecord := chunks[0].Record()
	newRecord, err := dm.filterRecord(oldRecord, whereCondition)
	if err != nil {
		return err
	}
	defer newRecord.Release()

	// 存储过滤后的数据
	chunk := types.NewChunk(newRecord, 0, int64(newRecord.NumRows()))
	defer chunk.Release()

	key := dm.keyManager.TableChunkKey(dbName, tableName, 0)
	record := chunk.Record()
	return dm.engine.Put(key, &record)
}

// 获取表的所有数据块
func (dm *DataManager) getTableChunks(dbName, tableName string, chunkCount int64) ([]*types.Chunk, error) {
	var chunks []*types.Chunk

	// 至少尝试读取第一个chunk
	maxChunks := chunkCount
	if maxChunks == 0 {
		maxChunks = 1
	}

	for i := int64(0); i < maxChunks; i++ {
		key := dm.keyManager.TableChunkKey(dbName, tableName, i)
		record, err := dm.engine.Get(key)
		if err != nil {
			return nil, err
		}

		if record != nil {
			chunk := types.NewChunk(record, i, int64(record.NumRows()))
			chunks = append(chunks, chunk)
		} else {
			chunks = append(chunks, nil)
		}
	}

	return chunks, nil
}

// 创建新记录
func (dm *DataManager) createRecord(schema *arrow.Schema, columns []string, values []interface{}) (arrow.Record, error) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 创建列名到索引的映射
	fieldMap := make(map[string]int)
	for i, field := range schema.Fields() {
		fieldMap[field.Name] = i
	}

	// 为每个字段准备值（使用NULL作为默认值）
	fieldValues := make([]interface{}, len(schema.Fields()))

	// 首先将所有字段设为NULL
	for i := range fieldValues {
		fieldValues[i] = nil
	}

	// 然后设置实际提供的值
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

// 合并两个记录
func (dm *DataManager) mergeRecords(oldRecord, newRecord arrow.Record) (arrow.Record, error) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, oldRecord.Schema())
	defer builder.Release()

	// 复制旧数据
	for colIdx := int64(0); colIdx < oldRecord.NumCols(); colIdx++ {
		oldCol := oldRecord.Column(int(colIdx))
		field := builder.Field(int(colIdx))

		for rowIdx := int64(0); rowIdx < oldRecord.NumRows(); rowIdx++ {
			value := dm.getValueFromArray(oldCol, int(rowIdx))
			err := dm.appendValue(field, value)
			if err != nil {
				return nil, err
			}
		}
	}

	// 复制新数据
	for colIdx := int64(0); colIdx < newRecord.NumCols(); colIdx++ {
		newCol := newRecord.Column(int(colIdx))
		field := builder.Field(int(colIdx))

		for rowIdx := int64(0); rowIdx < newRecord.NumRows(); rowIdx++ {
			value := dm.getValueFromArray(newCol, int(rowIdx))
			err := dm.appendValue(field, value)
			if err != nil {
				return nil, err
			}
		}
	}

	return builder.NewRecord(), nil
}

// 更新记录
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

// 过滤记录
func (dm *DataManager) filterRecord(oldRecord arrow.Record, whereCondition func(arrow.Record, int) bool) (arrow.Record, error) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, oldRecord.Schema())
	defer builder.Release()

	// 处理每一行
	for rowIdx := int64(0); rowIdx < oldRecord.NumRows(); rowIdx++ {
		shouldKeep := whereCondition == nil || !whereCondition(oldRecord, int(rowIdx))

		if shouldKeep {
			for colIdx := int64(0); colIdx < oldRecord.NumCols(); colIdx++ {
				field := builder.Field(int(colIdx))
				oldCol := oldRecord.Column(int(colIdx))
				value := dm.getValueFromArray(oldCol, int(rowIdx))

				err := dm.appendValue(field, value)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return builder.NewRecord(), nil
}

// 从Array中获取值
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

// 向字段添加值
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

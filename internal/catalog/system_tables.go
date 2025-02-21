package catalog

import (
	"fmt"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/storage"
	"github.com/yyun543/minidb/internal/types"
)

// 系统表的 Schema 定义
var SysDatabasesSchema = arrow.NewSchema([]arrow.Field{
	{Name: "id", Type: arrow.PrimitiveTypes.Int64},
	{Name: "name", Type: arrow.BinaryTypes.String},
}, nil)

var SysTablesSchema = arrow.NewSchema([]arrow.Field{
	{Name: "id", Type: arrow.PrimitiveTypes.Int64},
	{Name: "db_id", Type: arrow.PrimitiveTypes.Int64},
	{Name: "db_name", Type: arrow.BinaryTypes.String},
	{Name: "table_name", Type: arrow.BinaryTypes.String},
	{Name: "schema", Type: arrow.BinaryTypes.Binary}, // 使用Binary类型存储Arrow Schema
}, nil)

var SysColumnsSchema = arrow.NewSchema([]arrow.Field{
	{Name: "id", Type: arrow.PrimitiveTypes.Int64},
	{Name: "db_id", Type: arrow.PrimitiveTypes.Int64},
	{Name: "db_name", Type: arrow.BinaryTypes.String},
	{Name: "table_id", Type: arrow.PrimitiveTypes.Int64},
	{Name: "table_name", Type: arrow.BinaryTypes.String},
	{Name: "column_name", Type: arrow.BinaryTypes.String},
	{Name: "column_type", Type: arrow.BinaryTypes.String},
}, nil)

var SysIndexesSchema = arrow.NewSchema([]arrow.Field{
	{Name: "id", Type: arrow.PrimitiveTypes.Int64},
	{Name: "db_id", Type: arrow.PrimitiveTypes.Int64},
	{Name: "db_name", Type: arrow.BinaryTypes.String},
	{Name: "table_id", Type: arrow.PrimitiveTypes.Int64},
	{Name: "table_name", Type: arrow.BinaryTypes.String},
	{Name: "index_name", Type: arrow.BinaryTypes.String},
	{Name: "columns", Type: arrow.BinaryTypes.String}, // 可存储 JSON 数组或逗号分隔的名称序列
}, nil)

// createRecord 创建包含给定数据的 Arrow Record
func createRecord(schema *arrow.Schema, rowData [][]interface{}) arrow.Record {
	// 创建 RecordBuilder
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 遍历每一列
	for i := 0; i < len(schema.Fields()); i++ {
		fieldBuilder := builder.Field(i)
		// 确保 rowData 有足够的数据
		if i >= len(rowData) {
			continue
		}

		// 遍历该列的每一行数据
		for _, val := range rowData[i] {
			switch builder := fieldBuilder.(type) {
			case *array.Int64Builder:
				if v, ok := val.(int64); ok {
					builder.Append(v)
				}
			case *array.StringBuilder:
				if v, ok := val.(string); ok {
					builder.Append(v)
				}
			case *array.BinaryBuilder:
				if v, ok := val.([]byte); ok {
					builder.Append(v)
				}
			}
		}
	}

	// 创建 Record
	record := builder.NewRecord()
	return record
}

// InitializeSystemTables 检查系统表是否存在；如果不存在，则创建初始记录。
func InitializeSystemTables(engine storage.Engine) error {
	km := storage.NewKeyManager()

	// 检查 sys_databases 表
	rec, err := engine.Get(km.TableKey(storage.SYS_DATABASE, storage.SYS_DATABASES))
	if err != nil {
		return fmt.Errorf("failed to get sys_databases: %w", err)
	}
	if rec == nil {
		// 创建包含系统数据库的初始记录
		data := [][]interface{}{
			{int64(1)},             // id
			{storage.SYS_DATABASE}, // name
		}
		initRec := createRecord(SysDatabasesSchema, data)
		if err := engine.Put(km.TableKey(storage.SYS_DATABASE, storage.SYS_DATABASES), &initRec); err != nil {
			initRec.Release()
			return fmt.Errorf("failed to initialize sys_databases: %w", err)
		}
		initRec.Release()
	} else {
		rec.Release()
	}

	// 检查 sys_tables 表
	rec, err = engine.Get(km.TableKey(storage.SYS_DATABASE, storage.SYS_TABLES))
	if err != nil {
		return fmt.Errorf("failed to get sys_tables: %w", err)
	}
	if rec == nil {
		// 创建包含系统表的初始记录
		data := [][]interface{}{
			{int64(1), int64(2), int64(3), int64(4)},                                                 // id
			{int64(1), int64(1), int64(1), int64(1)},                                                 // db_id
			{storage.SYS_DATABASE, storage.SYS_DATABASE, storage.SYS_DATABASE, storage.SYS_DATABASE}, // db_name
			{storage.SYS_DATABASES, storage.SYS_TABLES, storage.SYS_COLUMNS, storage.SYS_INDEXES},    // table_name
			{types.SerializeSchema(SysDatabasesSchema), types.SerializeSchema(SysTablesSchema),
				types.SerializeSchema(SysColumnsSchema), types.SerializeSchema(SysIndexesSchema)}, // schema
		}
		initRec := createRecord(SysTablesSchema, data)
		if err := engine.Put(km.TableKey(storage.SYS_DATABASE, storage.SYS_TABLES), &initRec); err != nil {
			initRec.Release()
			return fmt.Errorf("failed to initialize sys_tables: %w", err)
		}
		initRec.Release()
	} else {
		rec.Release()
	}

	// 检查 sys_columns 表
	rec, err = engine.Get(km.TableKey(storage.SYS_DATABASE, storage.SYS_COLUMNS))
	if err != nil {
		return fmt.Errorf("failed to get sys_columns: %w", err)
	}
	if rec == nil {
		// 创建包含系统表列信息的初始记录
		data := [][]interface{}{
			{int64(1)},              // id
			{int64(1)},              // db_id
			{storage.SYS_DATABASE},  // db_name
			{int64(1)},              // table_id
			{storage.SYS_DATABASES}, // table_name
			{"name"},                // column_name
			{"string"},              // column_type
		}
		initRec := createRecord(SysColumnsSchema, data)
		if err := engine.Put(km.TableKey(storage.SYS_DATABASE, storage.SYS_COLUMNS), &initRec); err != nil {
			initRec.Release()
			return fmt.Errorf("failed to initialize sys_columns: %w", err)
		}
		initRec.Release()
	} else {
		rec.Release()
	}

	// 检查 sys_indexes 表
	rec, err = engine.Get(km.TableKey(storage.SYS_DATABASE, storage.SYS_INDEXES))
	if err != nil {
		return fmt.Errorf("failed to get sys_indexes: %w", err)
	}
	if rec == nil {
		// 创建空的系统索引表
		data := [][]interface{}{} // 初始不包含任何索引
		initRec := createRecord(SysIndexesSchema, data)
		if err := engine.Put(km.TableKey(storage.SYS_DATABASE, storage.SYS_INDEXES), &initRec); err != nil {
			initRec.Release()
			return fmt.Errorf("failed to initialize sys_indexes: %w", err)
		}
		initRec.Release()
	} else {
		rec.Release()
	}

	return nil
}

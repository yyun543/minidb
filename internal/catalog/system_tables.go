package catalog

import (
	"fmt"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/storage"
)

const (
	// 定义系统表的固定 key，用于存储系统级元数据
	KeySysDatabases = "sys:databases"
	KeySysTables    = "sys:tables"
	KeySysColumns   = "sys:columns"
	KeySysIndexes   = "sys:indexes"
)

var SysDatabasesSchema = arrow.NewSchema([]arrow.Field{
	{Name: "name", Type: arrow.BinaryTypes.String},
}, nil)

var SysTablesSchema = arrow.NewSchema([]arrow.Field{
	{Name: "database", Type: arrow.BinaryTypes.String},
	{Name: "table", Type: arrow.BinaryTypes.String},
	{Name: "schema", Type: arrow.BinaryTypes.String}, // 存储 Arrow Schema 的 JSON 表示
}, nil)

var SysColumnsSchema = arrow.NewSchema([]arrow.Field{
	{Name: "database", Type: arrow.BinaryTypes.String},
	{Name: "table", Type: arrow.BinaryTypes.String},
	{Name: "column", Type: arrow.BinaryTypes.String},
	{Name: "type", Type: arrow.BinaryTypes.String},
}, nil)

var SysIndexesSchema = arrow.NewSchema([]arrow.Field{
	{Name: "database", Type: arrow.BinaryTypes.String},
	{Name: "table", Type: arrow.BinaryTypes.String},
	{Name: "index", Type: arrow.BinaryTypes.String},
	{Name: "columns", Type: arrow.BinaryTypes.String}, // 可存储 JSON 数组或逗号分隔的名称序列
}, nil)

// createEmptyRecord 辅助函数：为给定 schema 构造一个空的 Arrow record。
// 这里假设系统表所有的字段均为 string 类型。
func createEmptyRecord(schema *arrow.Schema) arrow.Record {
	pool := memory.NewGoAllocator()
	arrays := make([]arrow.Array, len(schema.Fields()))
	for i := range schema.Fields() {
		builder := array.NewStringBuilder(pool)
		arr := builder.NewArray()
		builder.Release()
		arrays[i] = arr
	}
	return array.NewRecord(schema, arrays, 0)
}

// InitializeSystemTables 检查系统表是否存在；如果不存在，则创建空记录初始化，确保系统自举时 Catalog 元数据可用。
func InitializeSystemTables(engine storage.Engine) error {
	// 检查 sys_databases 表
	rec, err := engine.Get([]byte(KeySysDatabases))
	if err != nil {
		return fmt.Errorf("failed to get sys_databases: %w", err)
	}
	if rec == nil {
		emptyRec := createEmptyRecord(SysDatabasesSchema)
		if err := engine.Put([]byte(KeySysDatabases), &emptyRec); err != nil {
			emptyRec.Release()
			return fmt.Errorf("failed to initialize sys_databases: %w", err)
		}
		emptyRec.Release()
	} else {
		rec.Release()
	}

	// 检查 sys_tables 表
	rec, err = engine.Get([]byte(KeySysTables))
	if err != nil {
		return fmt.Errorf("failed to get sys_tables: %w", err)
	}
	if rec == nil {
		emptyRec := createEmptyRecord(SysTablesSchema)
		if err := engine.Put([]byte(KeySysTables), &emptyRec); err != nil {
			emptyRec.Release()
			return fmt.Errorf("failed to initialize sys_tables: %w", err)
		}
		emptyRec.Release()
	} else {
		rec.Release()
	}

	// 检查 sys_columns 表
	rec, err = engine.Get([]byte(KeySysColumns))
	if err != nil {
		return fmt.Errorf("failed to get sys_columns: %w", err)
	}
	if rec == nil {
		emptyRec := createEmptyRecord(SysColumnsSchema)
		if err := engine.Put([]byte(KeySysColumns), &emptyRec); err != nil {
			emptyRec.Release()
			return fmt.Errorf("failed to initialize sys_columns: %w", err)
		}
		emptyRec.Release()
	} else {
		rec.Release()
	}

	// 检查 sys_indexes 表
	rec, err = engine.Get([]byte(KeySysIndexes))
	if err != nil {
		return fmt.Errorf("failed to get sys_indexes: %w", err)
	}
	if rec == nil {
		emptyRec := createEmptyRecord(SysIndexesSchema)
		if err := engine.Put([]byte(KeySysIndexes), &emptyRec); err != nil {
			emptyRec.Release()
			return fmt.Errorf("failed to initialize sys_indexes: %w", err)
		}
		emptyRec.Release()
	} else {
		rec.Release()
	}

	return nil
}

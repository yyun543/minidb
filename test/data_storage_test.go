package test

import (
	"fmt"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/session"
)

// 测试数据存储和检索的基本功能
func TestDataStorageAndRetrieval(t *testing.T) {
	// 创建catalog
	cat, err := catalog.NewCatalogWithDefaultStorage()
	if err != nil {
		t.Fatalf("Failed to create catalog: %v", err)
	}

	if err := cat.Init(); err != nil {
		t.Fatalf("Failed to initialize catalog: %v", err)
	}

	// 创建执行器（用于后续可能的测试）
	_ = executor.NewExecutor(cat)

	// 创建数据管理器
	dataManager := executor.NewDataManager(cat)

	// 创建会话
	sessMgr, err := session.NewSessionManager()
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	sess := sessMgr.CreateSession()
	sess.CurrentDB = "test_db"

	// 创建数据库
	err = cat.CreateDatabase("test_db")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// 手动创建表schema
	tableMeta := catalog.TableMeta{
		Database:   "test_db",
		Table:      "users",
		ChunkCount: 0,
		Schema:     createTestSchema(),
	}

	err = cat.CreateTable("test_db", tableMeta)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 测试插入数据
	columns := []string{"id", "name", "age"}
	values := []interface{}{1, "John Doe", 25}

	err = dataManager.InsertData("test_db", "users", columns, values)
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// 测试检索数据
	batches, err := dataManager.GetTableData("test_db", "users")
	if err != nil {
		t.Fatalf("Failed to get table data: %v", err)
	}

	if len(batches) == 0 {
		t.Fatalf("No data retrieved after insert")
	}

	if batches[0] == nil {
		t.Fatalf("First batch is nil")
	}

	record := batches[0].Record()
	if record.NumRows() == 0 {
		t.Fatalf("No rows in retrieved data")
	}

	// 验证数据内容
	if record.NumRows() != 1 {
		t.Errorf("Expected 1 row, got %d", record.NumRows())
	}

	if record.NumCols() != 3 {
		t.Errorf("Expected 3 columns, got %d", record.NumCols())
	}

	// 检查第一行数据
	fmt.Printf("Retrieved data:\n")
	for i := int64(0); i < record.NumCols(); i++ {
		column := record.Column(int(i))
		field := record.Schema().Field(int(i))
		fmt.Printf("Column %s: %v\n", field.Name, getColumnValueFromArray(column, 0))
	}
}

func createTestSchema() *arrow.Schema {
	fields := []arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "name", Type: arrow.BinaryTypes.String},
		{Name: "age", Type: arrow.PrimitiveTypes.Int64},
	}
	return arrow.NewSchema(fields, nil)
}

func getColumnValueFromArray(column arrow.Array, index int) interface{} {
	switch col := column.(type) {
	case *array.Int64:
		return col.Value(index)
	case *array.String:
		return col.Value(index)
	default:
		return "unknown"
	}
}

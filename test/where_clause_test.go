package test

import (
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/statistics"
)

// 测试WHERE子句中的列查找
func TestWhereClauseColumnLookup(t *testing.T) {
	// 创建catalog
	cat, err := catalog.NewCatalogWithDefaultStorage()
	if err != nil {
		t.Fatalf("Failed to create catalog: %v", err)
	}

	if err := cat.Init(); err != nil {
		t.Fatalf("Failed to initialize catalog: %v", err)
	}

	// 创建会话
	sessMgr, err := session.NewSessionManager()
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	sess := sessMgr.CreateSession()
	sess.CurrentDB = "ecommerce"

	// 创建数据库
	err = cat.CreateDatabase("ecommerce")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// 创建表 - 确保包含age列
	tableMeta := catalog.TableMeta{
		Database:   "ecommerce",
		Table:      "users",
		ChunkCount: 0,
		Schema:     createUsersSchema(),
	}

	err = cat.CreateTable("ecommerce", tableMeta)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 插入测试数据
	dataManager := executor.NewDataManager(cat)
	columns := []string{"id", "name", "email", "age", "created_at"}
	values := []interface{}{1, "John Doe", "john@example.com", 25, "2024-01-01"}

	err = dataManager.InsertData("ecommerce", "users", columns, values)
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// 验证表schema
	retrievedMeta, err := cat.GetTable("ecommerce", "users")
	if err != nil {
		t.Fatalf("Failed to get table metadata: %v", err)
	}

	t.Logf("Table schema fields:")
	for i, field := range retrievedMeta.Schema.Fields() {
		t.Logf("  %d: %s (%s)", i, field.Name, field.Type)
	}

	// 测试简单的SELECT * FROM users
	sql1 := "SELECT * FROM users"
	ast1, err := parser.Parse(sql1)
	if err != nil {
		t.Fatalf("Failed to parse SQL: %v", err)
	}

	opt := optimizer.NewOptimizer()
	plan1, err := opt.Optimize(ast1)
	if err != nil {
		t.Fatalf("Failed to optimize query: %v", err)
	}

	t.Logf("Plan for 'SELECT * FROM users':")
	t.Logf("  Type: %v", plan1.Type)
	t.Logf("  Children: %d", len(plan1.Children))

	// 测试带WHERE子句的SELECT
	sql2 := "SELECT name, email FROM users WHERE age > 25"
	ast2, err := parser.Parse(sql2)
	if err != nil {
		t.Fatalf("Failed to parse SQL: %v", err)
	}

	plan2, err := opt.Optimize(ast2)
	if err != nil {
		t.Fatalf("Failed to optimize WHERE query: %v", err)
	}

	t.Logf("Plan for 'SELECT name, email FROM users WHERE age > 25':")
	t.Logf("  Type: %v", plan2.Type)
	t.Logf("  Children: %d", len(plan2.Children))

	// 检查计划结构
	if len(plan2.Children) > 0 {
		t.Logf("  Child 0 Type: %v", plan2.Children[0].Type)
		if len(plan2.Children[0].Children) > 0 {
			t.Logf("    Child 0.0 Type: %v", plan2.Children[0].Children[0].Type)
		}
	}

	// 测试向量化执行器
	statsMgr := statistics.NewStatisticsManager()
	vectorizedExec := executor.NewVectorizedExecutor(cat, statsMgr)

	// 手动测试schema推断
	t.Logf("Manual schema inference test:")
	inferredSchema := vectorizedExec.InferSchema(plan2, sess)
	if inferredSchema != nil {
		t.Logf("  Inferred schema fields: %d", inferredSchema.NumFields())
		for i, field := range inferredSchema.Fields() {
			t.Logf("    %d: %s (%s)", i, field.Name, field.Type)
		}
	} else {
		t.Logf("  Inferred schema is nil")
	}

	t.Logf("Testing vectorized execution with WHERE clause...")
	vectorizedResult, err := vectorizedExec.Execute(plan2, sess)
	if err != nil {
		t.Logf("Vectorized execution failed: %v", err)
		// 这个预期会失败，我们要分析为什么
	} else {
		t.Logf("Vectorized executor succeeded: Headers=%v, Batches count=%d",
			vectorizedResult.Headers, len(vectorizedResult.Batches))
	}
}

func createUsersSchema() *arrow.Schema {
	fields := []arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "name", Type: arrow.BinaryTypes.String},
		{Name: "email", Type: arrow.BinaryTypes.String},
		{Name: "age", Type: arrow.PrimitiveTypes.Int64},
		{Name: "created_at", Type: arrow.BinaryTypes.String},
	}
	return arrow.NewSchema(fields, nil)
}

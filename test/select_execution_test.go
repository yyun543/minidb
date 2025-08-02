package test

import (
	"testing"

	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/statistics"
)

// 测试SELECT语句的执行路径
func TestSelectExecution(t *testing.T) {
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
	sess.CurrentDB = "test_db"

	// 创建数据库
	err = cat.CreateDatabase("test_db")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// 创建表
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

	// 插入测试数据
	dataManager := executor.NewDataManager(cat)
	columns := []string{"id", "name", "age"}
	values := []interface{}{1, "John Doe", 25}

	err = dataManager.InsertData("test_db", "users", columns, values)
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// 验证数据确实被插入了
	batches, err := dataManager.GetTableData("test_db", "users")
	if err != nil {
		t.Fatalf("Failed to get table data: %v", err)
	}

	if len(batches) == 0 {
		t.Fatalf("No data retrieved after insert")
	}

	t.Logf("Data inserted successfully, %d batches found", len(batches))

	// 测试常规执行器的SELECT
	regularExec := executor.NewExecutor(cat)

	// 解析SELECT语句
	sql := "SELECT * FROM users"
	ast, err := parser.Parse(sql)
	if err != nil {
		t.Fatalf("Failed to parse SQL: %v", err)
	}

	// 优化查询
	opt := optimizer.NewOptimizer()
	plan, err := opt.Optimize(ast)
	if err != nil {
		t.Fatalf("Failed to optimize query: %v", err)
	}

	t.Logf("Query plan type: %v", plan.Type)
	if plan.Type == optimizer.SelectPlan {
		props := plan.Properties.(*optimizer.SelectProperties)
		t.Logf("SelectProperties: All=%v, Columns=%+v", props.All, props.Columns)
	}

	// 使用常规执行器执行
	regularResult, err := regularExec.Execute(plan, sess)
	if err != nil {
		t.Fatalf("Failed to execute query with regular executor: %v", err)
	}

	t.Logf("Regular executor result: Headers=%v, Batches count=%d",
		regularResult.Headers, len(regularResult.Batches()))

	if len(regularResult.Batches()) == 0 {
		t.Errorf("Regular executor returned no batches")
	}

	// 测试向量化执行器的SELECT
	statsMgr := statistics.NewStatisticsManager()
	vectorizedExec := executor.NewVectorizedExecutor(cat, statsMgr)

	vectorizedResult, err := vectorizedExec.Execute(plan, sess)
	if err != nil {
		t.Fatalf("Failed to execute query with vectorized executor: %v", err)
	}

	t.Logf("Vectorized executor result: Headers=%v, Batches count=%d",
		vectorizedResult.Headers, len(vectorizedResult.Batches))

	if len(vectorizedResult.Batches) == 0 {
		t.Errorf("Vectorized executor returned no batches")
	}
}

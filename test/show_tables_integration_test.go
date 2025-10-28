package test

import (
	"testing"

	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

// TestShowTablesIntegration 测试SHOW TABLES在实际使用场景中的功能
func TestShowTablesIntegration(t *testing.T) {
	storageEngine, err := storage.NewParquetEngine(SetupTestDir(t, "show_tables_integration.wal"))
	assert.NoError(t, err)
	defer storageEngine.Close()
	err = storageEngine.Open()
	assert.NoError(t, err)

	cat := catalog.NewCatalog()
	cat.SetStorageEngine(storageEngine)
	err = cat.Init()
	if err != nil {
		t.Fatalf("Failed to initialize catalog: %v", err)
	}
	sessMgr, err := session.NewSessionManager()
	assert.NoError(t, err)
	sess := sessMgr.CreateSession()
	opt := optimizer.NewOptimizer()
	exec := executor.NewExecutor(cat)

	execSQL := func(sql string) error {
		stmt, err := parser.Parse(sql)
		if err != nil {
			return err
		}
		plan, err := opt.Optimize(stmt)
		if err != nil {
			return err
		}
		_, err = exec.Execute(plan, sess)
		return err
	}

	getTableCount := func() int {
		stmt, err := parser.Parse("SHOW TABLES")
		if err != nil {
			return -1
		}
		plan, err := opt.Optimize(stmt)
		if err != nil {
			return -1
		}
		result, err := exec.Execute(plan, sess)
		if err != nil {
			return -1
		}
		if len(result.Batches()) == 0 {
			return 0
		}
		return int(result.Batches()[0].NumRows())
	}

	// 测试场景：从0个表开始，逐步添加表并验证SHOW TABLES

	// 1. 创建数据库，应该有0个表
	err = execSQL("CREATE DATABASE integration_test")
	assert.NoError(t, err)
	sess.CurrentDB = "integration_test"

	count := getTableCount()
	assert.Equal(t, 0, count, "New database should have 0 tables")
	t.Logf("✓ Empty database shows 0 tables")

	// 2. 创建第一个表
	err = execSQL("CREATE TABLE orders (id INTEGER, total FLOAT)")
	assert.NoError(t, err)

	count = getTableCount()
	assert.Equal(t, 1, count, "Should show 1 table after creating first table")
	t.Logf("✓ After creating 1 table, SHOW TABLES shows 1 table")

	// 3. 创建第二个表
	err = execSQL("CREATE TABLE customers (id INTEGER, name VARCHAR(100))")
	assert.NoError(t, err)

	count = getTableCount()
	assert.Equal(t, 2, count, "Should show 2 tables after creating second table")
	t.Logf("✓ After creating 2 tables, SHOW TABLES shows 2 tables")

	// 4. 再创建第三个表
	err = execSQL("CREATE TABLE products (id INTEGER, name VARCHAR(50), price FLOAT)")
	assert.NoError(t, err)

	count = getTableCount()
	assert.Equal(t, 3, count, "Should show 3 tables after creating third table")
	t.Logf("✓ After creating 3 tables, SHOW TABLES shows 3 tables")

	// 5. 验证表名内容
	stmt, err := parser.Parse("SHOW TABLES")
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	result, err := exec.Execute(plan, sess)
	assert.NoError(t, err)

	assert.True(t, len(result.Batches()) > 0, "Should have result batches")
	batch := result.Batches()[0]
	record := batch.Record()

	// 收集所有表名
	var tableNames []string
	stringArray := record.Column(0).(*array.String)
	for i := 0; i < int(record.NumRows()); i++ {
		tableNames = append(tableNames, stringArray.Value(i))
	}

	// 验证包含我们创建的所有表名
	assert.Contains(t, tableNames, "orders", "Should include 'orders' table")
	assert.Contains(t, tableNames, "customers", "Should include 'customers' table")
	assert.Contains(t, tableNames, "products", "Should include 'products' table")

	t.Logf("✅ SHOW TABLES integration test PASSED - found tables: %v", tableNames)
}

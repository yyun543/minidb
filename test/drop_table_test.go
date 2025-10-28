package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

// TestDropTableFunctionality tests that DROP TABLE correctly removes tables
func TestDropTableFunctionality(t *testing.T) {
	testDir := SetupTestDir(t, "drop_table_test")
	storageEngine, err := storage.NewParquetEngine(testDir)
	assert.NoError(t, err)
	defer storageEngine.Close()
	err = storageEngine.Open()
	assert.NoError(t, err)

	cat := catalog.NewCatalog()
	cat.SetStorageEngine(storageEngine)
	err = cat.Init()
	assert.NoError(t, err)

	sessMgr, err := session.NewSessionManager()
	assert.NoError(t, err)
	sess := sessMgr.CreateSession()

	opt := optimizer.NewOptimizer()
	exec := executor.NewExecutor(cat)

	// Create database
	createDbSQL := "CREATE DATABASE testdb"
	stmt, err := parser.Parse(createDbSQL)
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "testdb"

	// Create table
	createTableSQL := "CREATE TABLE test_table (id INT, name VARCHAR)"
	stmt, err = parser.Parse(createTableSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Verify table exists
	tables, err := cat.GetAllTables("testdb")
	assert.NoError(t, err)
	assert.Contains(t, tables, "test_table", "Table should exist after creation")

	// Drop table
	dropTableSQL := "DROP TABLE test_table"
	stmt, err = parser.Parse(dropTableSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Verify table no longer exists
	tables, err = cat.GetAllTables("testdb")
	assert.NoError(t, err)
	assert.NotContains(t, tables, "test_table", "Table should not exist after DROP")

	t.Log("✅ DROP TABLE functionality works correctly")
}

// TestDropTableWithData tests that DROP TABLE works even when table has data
func TestDropTableWithData(t *testing.T) {
	testDir := SetupTestDir(t, "drop_table_with_data_test")
	storageEngine, err := storage.NewParquetEngine(testDir)
	assert.NoError(t, err)
	defer storageEngine.Close()
	err = storageEngine.Open()
	assert.NoError(t, err)

	cat := catalog.NewCatalog()
	cat.SetStorageEngine(storageEngine)
	err = cat.Init()
	assert.NoError(t, err)

	sessMgr, err := session.NewSessionManager()
	assert.NoError(t, err)
	sess := sessMgr.CreateSession()

	opt := optimizer.NewOptimizer()
	exec := executor.NewExecutor(cat)

	// Create database
	createDbSQL := "CREATE DATABASE testdb"
	stmt, err := parser.Parse(createDbSQL)
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "testdb"

	// Create table
	createTableSQL := "CREATE TABLE data_table (id INT, value VARCHAR)"
	stmt, err = parser.Parse(createTableSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Insert some data
	insertSQL := "INSERT INTO data_table VALUES (1, 'test')"
	stmt, err = parser.Parse(insertSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Verify data exists
	selectSQL := "SELECT * FROM data_table"
	stmt, err = parser.Parse(selectSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	result, err := exec.Execute(plan, sess)
	assert.NoError(t, err)
	assert.Greater(t, len(result.Batches()), 0, "Should have data")

	// Drop table with data
	dropTableSQL := "DROP TABLE data_table"
	stmt, err = parser.Parse(dropTableSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Verify table no longer exists
	tables, err := cat.GetAllTables("testdb")
	assert.NoError(t, err)
	assert.NotContains(t, tables, "data_table", "Table with data should be dropped")

	t.Log("✅ DROP TABLE with data works correctly")
}

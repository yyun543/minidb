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

// TestMultiRowInsert tests that multi-row INSERT statements work correctly
func TestMultiRowInsert(t *testing.T) {
	testDir := SetupTestDir(t, "multi_row_insert_test")
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
	createTableSQL := "CREATE TABLE products (id INT, name VARCHAR, price INT)"
	stmt, err = parser.Parse(createTableSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Multi-row INSERT (user's example from logs)
	insertSQL := "INSERT INTO products VALUES (2, 'Mouse', 25), (3, 'Keyboard', 75), (4, 'Monitor', 300)"
	stmt, err = parser.Parse(insertSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Verify all 3 rows were inserted
	selectSQL := "SELECT * FROM products"
	stmt, err = parser.Parse(selectSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	result, err := exec.Execute(plan, sess)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Count total rows across all batches
	totalRows := 0
	for _, batch := range result.Batches() {
		totalRows += int(batch.NumRows())
	}

	assert.Equal(t, 3, totalRows, "Should have inserted all 3 rows")

	// Verify specific data
	// Check that we have Mouse, Keyboard, Monitor
	foundMouse := false
	foundKeyboard := false
	foundMonitor := false

	for _, batch := range result.Batches() {
		record := batch.Record()
		for i := int64(0); i < record.NumRows(); i++ {
			nameCol := record.Column(1) // name column
			if nameCol.IsValid(int(i)) && !nameCol.IsNull(int(i)) {
				nameStr := nameCol.ValueStr(int(i))
				if nameStr == "Mouse" {
					foundMouse = true
				} else if nameStr == "Keyboard" {
					foundKeyboard = true
				} else if nameStr == "Monitor" {
					foundMonitor = true
				}
			}
		}
	}

	assert.True(t, foundMouse, "Should find Mouse in results")
	assert.True(t, foundKeyboard, "Should find Keyboard in results")
	assert.True(t, foundMonitor, "Should find Monitor in results")

	t.Log("✅ Multi-row INSERT works correctly")
}

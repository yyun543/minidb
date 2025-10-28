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

// TestNumericColumnDisplay tests that numeric columns are displayed correctly in SELECT results
func TestNumericColumnDisplay(t *testing.T) {
	testDir := SetupTestDir(t, "numeric_display_test")
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

	// Create table with multiple column types including numeric
	createTableSQL := "CREATE TABLE products (id INT, name VARCHAR, price INT, quantity INT, in_stock INT, created_date VARCHAR, category VARCHAR)"
	stmt, err = parser.Parse(createTableSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Insert data with numeric values
	insertSQL := "INSERT INTO products VALUES (1, 'Laptop', 999, 10, 1, '2024-01-01', 'Electronics')"
	stmt, err = parser.Parse(insertSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// SELECT to verify numeric columns are displayed
	selectSQL := "SELECT * FROM products"
	stmt, err = parser.Parse(selectSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	result, err := exec.Execute(plan, sess)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify result has data
	assert.Greater(t, len(result.Batches()), 0, "Should have at least one batch")

	batch := result.Batches()[0]
	assert.Greater(t, batch.NumRows(), int64(0), "Should have at least one row")

	record := batch.Record()
	assert.Equal(t, int64(7), record.NumCols(), "Should have 7 columns")

	// Verify numeric column values are not null/empty
	// Column 0: id (INT)
	idCol := record.Column(0)
	assert.False(t, idCol.IsNull(0), "ID column should not be null")

	// Column 2: price (INT)
	priceCol := record.Column(2)
	assert.False(t, priceCol.IsNull(0), "Price column should not be null")

	// Column 3: quantity (INT)
	qtyCol := record.Column(3)
	assert.False(t, qtyCol.IsNull(0), "Quantity column should not be null")

	// Column 4: in_stock (INT)
	inStockCol := record.Column(4)
	assert.False(t, inStockCol.IsNull(0), "In-stock column should not be null")

	t.Log("✅ Numeric columns are displayed correctly")
}

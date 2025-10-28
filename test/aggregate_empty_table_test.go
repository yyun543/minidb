package test

import (
	"os"
	"testing"

	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

// TestAggregateOnEmptyTable tests aggregate functions on empty tables
// Per SQL:2023 standard, aggregate functions without GROUP BY must return
// exactly one row even when the input table is empty:
// - COUNT(*) returns 0
// - SUM/AVG/MIN/MAX return NULL
func TestAggregateOnEmptyTable(t *testing.T) {
	// Setup test environment
	testDir := "./test_data/aggregate_empty_test"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	storageEngine, err := storage.NewParquetEngine(testDir)
	require.NoError(t, err)
	defer storageEngine.Close()
	err = storageEngine.Open()
	require.NoError(t, err)

	cat := catalog.NewCatalog()
	cat.SetStorageEngine(storageEngine)
	err = cat.Init()
	require.NoError(t, err)

	sessMgr, err := session.NewSessionManager()
	require.NoError(t, err)
	sess := sessMgr.CreateSession()

	// Create database and table
	opt := optimizer.NewOptimizer()
	exec := executor.NewExecutor(cat)

	createDBNode, err := parser.Parse("CREATE DATABASE test_db")
	require.NoError(t, err)
	createDBPlan, err := opt.Optimize(createDBNode)
	require.NoError(t, err)
	_, err = exec.Execute(createDBPlan, sess)
	require.NoError(t, err)

	sess.CurrentDB = "test_db"

	createTableSQL := `CREATE TABLE products (
		id INTEGER,
		name VARCHAR(100),
		price DOUBLE,
		quantity INTEGER,
		category VARCHAR
	)`
	createNode, err := parser.Parse(createTableSQL)
	require.NoError(t, err)
	createPlan, err := opt.Optimize(createNode)
	require.NoError(t, err)
	_, err = exec.Execute(createPlan, sess)
	require.NoError(t, err)

	// Test aggregate functions on empty table (no inserts)
	testCases := []struct {
		name           string
		query          string
		expectedRows   int
		expectedCols   int
		validateResult func(t *testing.T, result *executor.ResultSet)
	}{
		{
			name:         "COUNT(*) on empty table",
			query:        "SELECT COUNT(*) FROM products",
			expectedRows: 1,
			expectedCols: 1,
			validateResult: func(t *testing.T, result *executor.ResultSet) {
				batches := result.Batches()
				require.Equal(t, 1, len(batches), "COUNT(*) must return exactly one batch")
				require.Equal(t, int64(1), batches[0].NumRows(), "COUNT(*) must return exactly one row")
				value := batches[0].Record().Column(0).(*array.Int64).Value(0)
				assert.Equal(t, int64(0), value, "COUNT(*) on empty table should return 0")
			},
		},
		{
			name:         "COUNT(column) on empty table",
			query:        "SELECT COUNT(name) FROM products",
			expectedRows: 1,
			expectedCols: 1,
			validateResult: func(t *testing.T, result *executor.ResultSet) {
				batches := result.Batches()
				require.Equal(t, 1, len(batches), "COUNT(column) must return exactly one batch")
				require.Equal(t, int64(1), batches[0].NumRows(), "COUNT(column) must return exactly one row")
				value := batches[0].Record().Column(0).(*array.Int64).Value(0)
				assert.Equal(t, int64(0), value, "COUNT(column) on empty table should return 0")
			},
		},
		{
			name:         "SUM on empty table",
			query:        "SELECT SUM(price) FROM products",
			expectedRows: 1,
			expectedCols: 1,
			validateResult: func(t *testing.T, result *executor.ResultSet) {
				batches := result.Batches()
				require.Equal(t, 1, len(batches), "SUM must return exactly one batch")
				require.Equal(t, int64(1), batches[0].NumRows(), "SUM must return exactly one row")
				isNull := batches[0].Record().Column(0).IsNull(0)
				assert.True(t, isNull, "SUM on empty table should return NULL")
			},
		},
		{
			name:         "AVG on empty table",
			query:        "SELECT AVG(price) FROM products",
			expectedRows: 1,
			expectedCols: 1,
			validateResult: func(t *testing.T, result *executor.ResultSet) {
				batches := result.Batches()
				require.Equal(t, 1, len(batches), "AVG must return exactly one batch")
				require.Equal(t, int64(1), batches[0].NumRows(), "AVG must return exactly one row")
				isNull := batches[0].Record().Column(0).IsNull(0)
				assert.True(t, isNull, "AVG on empty table should return NULL")
			},
		},
		{
			name:         "MIN on empty table",
			query:        "SELECT MIN(price) FROM products",
			expectedRows: 1,
			expectedCols: 1,
			validateResult: func(t *testing.T, result *executor.ResultSet) {
				batches := result.Batches()
				require.Equal(t, 1, len(batches), "MIN must return exactly one batch")
				require.Equal(t, int64(1), batches[0].NumRows(), "MIN must return exactly one row")
				isNull := batches[0].Record().Column(0).IsNull(0)
				assert.True(t, isNull, "MIN on empty table should return NULL")
			},
		},
		{
			name:         "MAX on empty table",
			query:        "SELECT MAX(price) FROM products",
			expectedRows: 1,
			expectedCols: 1,
			validateResult: func(t *testing.T, result *executor.ResultSet) {
				batches := result.Batches()
				require.Equal(t, 1, len(batches), "MAX must return exactly one batch")
				require.Equal(t, int64(1), batches[0].NumRows(), "MAX must return exactly one row")
				isNull := batches[0].Record().Column(0).IsNull(0)
				assert.True(t, isNull, "MAX on empty table should return NULL")
			},
		},
		{
			name:         "Multiple aggregates on empty table",
			query:        "SELECT COUNT(*) AS cnt, SUM(price) AS total, AVG(price) AS avg_price, MIN(price) AS min_p, MAX(price) AS max_p FROM products",
			expectedRows: 1,
			expectedCols: 5,
			validateResult: func(t *testing.T, result *executor.ResultSet) {
				batches := result.Batches()
				require.Equal(t, 1, len(batches), "Multiple aggregates must return exactly one batch")
				require.Equal(t, int64(1), batches[0].NumRows(), "Multiple aggregates must return exactly one row")

				record := batches[0].Record()
				assert.Equal(t, int64(0), record.Column(0).(*array.Int64).Value(0), "COUNT(*) should be 0")
				assert.True(t, record.Column(1).IsNull(0), "SUM should be NULL")
				assert.True(t, record.Column(2).IsNull(0), "AVG should be NULL")
				assert.True(t, record.Column(3).IsNull(0), "MIN should be NULL")
				assert.True(t, record.Column(4).IsNull(0), "MAX should be NULL")
			},
		},
		{
			name:         "GROUP BY on empty table",
			query:        "SELECT quantity, COUNT(*) FROM products GROUP BY quantity",
			expectedRows: 0,
			expectedCols: 2,
			validateResult: func(t *testing.T, result *executor.ResultSet) {
				batches := result.Batches()
				assert.Equal(t, 0, len(batches), "GROUP BY on empty table should return empty set")
			},
		},
		{
			name:         "GROUP BY with HAVING on empty table",
			query:        "SELECT quantity, COUNT(*) AS cnt FROM products GROUP BY quantity HAVING cnt > 1",
			expectedRows: 0,
			expectedCols: 2,
			validateResult: func(t *testing.T, result *executor.ResultSet) {
				batches := result.Batches()
				assert.Equal(t, 0, len(batches), "GROUP BY with HAVING on empty table should return empty set")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse query
			node, err := parser.Parse(tc.query)
			require.NoError(t, err, "Failed to parse query: %s", tc.query)

			// Optimize
			plan, err := opt.Optimize(node)
			require.NoError(t, err, "Failed to optimize query: %s", tc.query)

			// Execute
			result, err := exec.Execute(plan, sess)
			require.NoError(t, err, "Failed to execute query: %s", tc.query)
			require.NotNil(t, result, "Result should not be nil")

			// Validate expected structure
			assert.Equal(t, tc.expectedCols, len(result.Headers), "Expected %d columns", tc.expectedCols)

			// Validate specific results
			tc.validateResult(t, result)
		})
	}
}

// TestAggregateAfterDelete tests aggregate functions after DELETE operations
// This tests the specific scenario from the bug report
func TestAggregateAfterDelete(t *testing.T) {
	// Setup test environment
	testDir := "./test_data/aggregate_after_delete_test"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	storageEngine, err := storage.NewParquetEngine(testDir)
	require.NoError(t, err)
	defer storageEngine.Close()
	err = storageEngine.Open()
	require.NoError(t, err)

	cat := catalog.NewCatalog()
	cat.SetStorageEngine(storageEngine)
	err = cat.Init()
	require.NoError(t, err)

	sessMgr, err := session.NewSessionManager()
	require.NoError(t, err)
	sess := sessMgr.CreateSession()

	opt := optimizer.NewOptimizer()
	exec := executor.NewExecutor(cat)

	// Create database and table
	createDBNode, err := parser.Parse("CREATE DATABASE test_db")
	require.NoError(t, err)
	createDBPlan, err := opt.Optimize(createDBNode)
	require.NoError(t, err)
	_, err = exec.Execute(createDBPlan, sess)
	require.NoError(t, err)

	sess.CurrentDB = "test_db"

	createTableSQL := `CREATE TABLE products (
		id INTEGER,
		name VARCHAR(100),
		price DOUBLE,
		quantity INTEGER,
		category VARCHAR
	)`
	createNode, err := parser.Parse(createTableSQL)
	require.NoError(t, err)
	createPlan, err := opt.Optimize(createNode)
	require.NoError(t, err)
	_, err = exec.Execute(createPlan, sess)
	require.NoError(t, err)

	// Insert some data
	insertSQL := `INSERT INTO products VALUES
		(1, 'Laptop', 1099.00, 20, 'Electronics'),
		(2, 'Mouse', 29.99, 60, 'Electronics'),
		(3, 'Desk', 349.99, 25, 'Furniture'),
		(4, 'Chair', 199.99, 30, 'Furniture')`
	insertNode, err := parser.Parse(insertSQL)
	require.NoError(t, err)
	insertPlan, err := opt.Optimize(insertNode)
	require.NoError(t, err)
	_, err = exec.Execute(insertPlan, sess)
	require.NoError(t, err)

	// Verify data exists
	selectNode, err := parser.Parse("SELECT COUNT(*) FROM products")
	require.NoError(t, err)
	selectPlan, err := opt.Optimize(selectNode)
	require.NoError(t, err)
	result, err := exec.Execute(selectPlan, sess)
	require.NoError(t, err)
	batches := result.Batches()
	require.Greater(t, len(batches), 0)
	assert.Equal(t, int64(4), batches[0].Record().Column(0).(*array.Int64).Value(0), "Should have 4 rows initially")

	// Delete all data
	deleteNode, err := parser.Parse("DELETE FROM products")
	require.NoError(t, err)
	deletePlan, err := opt.Optimize(deleteNode)
	require.NoError(t, err)
	_, err = exec.Execute(deletePlan, sess)
	require.NoError(t, err)

	// Now test aggregate functions - should behave same as empty table
	testCases := []struct {
		name           string
		query          string
		validateResult func(t *testing.T, result *executor.ResultSet)
	}{
		{
			name:  "COUNT(*) after DELETE",
			query: "SELECT COUNT(*) FROM products",
			validateResult: func(t *testing.T, result *executor.ResultSet) {
				batches := result.Batches()
				require.Equal(t, 1, len(batches), "COUNT(*) must return a batch")
				require.Equal(t, int64(1), batches[0].NumRows(), "COUNT(*) must return exactly one row")
				assert.Equal(t, int64(0), batches[0].Record().Column(0).(*array.Int64).Value(0), "COUNT(*) should be 0 after DELETE")
			},
		},
		{
			name:  "SUM after DELETE",
			query: "SELECT SUM(price) FROM products",
			validateResult: func(t *testing.T, result *executor.ResultSet) {
				batches := result.Batches()
				require.Equal(t, 1, len(batches), "SUM must return a batch")
				require.Equal(t, int64(1), batches[0].NumRows(), "SUM must return exactly one row")
				isNull := batches[0].Record().Column(0).IsNull(0)
				assert.True(t, isNull, "SUM should be NULL after DELETE")
			},
		},
		{
			name:  "GROUP BY after DELETE",
			query: "SELECT category, COUNT(*) FROM products GROUP BY category",
			validateResult: func(t *testing.T, result *executor.ResultSet) {
				batches := result.Batches()
				assert.Equal(t, 0, len(batches), "GROUP BY should return empty set after DELETE")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node, err := parser.Parse(tc.query)
			require.NoError(t, err)
			plan, err := opt.Optimize(node)
			require.NoError(t, err)
			result, err := exec.Execute(plan, sess)
			require.NoError(t, err)
			require.NotNil(t, result)
			tc.validateResult(t, result)
		})
	}
}

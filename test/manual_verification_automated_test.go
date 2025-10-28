package test

import (
	"fmt"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
	"github.com/yyun543/minidb/internal/types"
)

// ManualVerificationExpectedResult represents the expected result of a query
type ManualVerificationExpectedResult struct {
	Description  string
	Headers      []string
	DataTypes    []arrow.Type // Expected Arrow types
	RowCount     int
	Values       [][]interface{} // Expected values [row][col]
	OrderMatters bool            // Whether row order matters
	CheckValues  bool            // Whether to check exact values
}

// ManualVerificationTestContext holds the test execution context
type ManualVerificationTestContext struct {
	t             *testing.T
	cat           *catalog.Catalog
	sessMgr       *session.SessionManager
	sess          *session.Session
	opt           *optimizer.Optimizer
	exec          *executor.ExecutorImpl
	storageEngine *storage.ParquetEngine
}

// NewManualVerificationTestContext creates a new test context
func NewManualVerificationTestContext(t *testing.T, testName string) *ManualVerificationTestContext {
	testDir := SetupTestDir(t, testName)
	storageEngine, err := storage.NewParquetEngine(testDir)
	require.NoError(t, err)

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

	return &ManualVerificationTestContext{
		t:             t,
		cat:           cat,
		sessMgr:       sessMgr,
		sess:          sess,
		opt:           opt,
		exec:          exec,
		storageEngine: storageEngine,
	}
}

// Cleanup closes the test context
func (tc *ManualVerificationTestContext) Cleanup() {
	if tc.storageEngine != nil {
		tc.storageEngine.Close()
	}
}

// ExecuteSQL executes a SQL statement
func (tc *ManualVerificationTestContext) ExecuteSQL(sql string) (*executor.ResultSet, error) {
	stmt, err := parser.Parse(sql)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	plan, err := tc.opt.Optimize(stmt)
	if err != nil {
		return nil, fmt.Errorf("optimize error: %w", err)
	}
	return tc.exec.Execute(plan, tc.sess)
}

// ExecuteSQLMustSucceed executes SQL and requires it to succeed
func (tc *ManualVerificationTestContext) ExecuteSQLMustSucceed(sql string) *executor.ResultSet {
	result, err := tc.ExecuteSQL(sql)
	require.NoError(tc.t, err, "SQL should succeed: %s", sql)
	return result
}

// ExecuteSQLMustFail executes SQL and requires it to fail
func (tc *ManualVerificationTestContext) ExecuteSQLMustFail(sql string) {
	_, err := tc.ExecuteSQL(sql)
	require.Error(tc.t, err, "SQL should fail: %s", sql)
}

// GetCellValue extracts a cell value from a batch
func GetCellValue(batch *types.Batch, rowIdx, colIdx int) interface{} {
	col := batch.Column(colIdx)
	if col.IsNull(rowIdx) {
		return nil
	}
	switch col.DataType().ID() {
	case arrow.INT64:
		return col.(*array.Int64).Value(rowIdx)
	case arrow.STRING:
		return col.(*array.String).Value(rowIdx)
	case arrow.FLOAT64:
		return col.(*array.Float64).Value(rowIdx)
	case arrow.BOOL:
		return col.(*array.Boolean).Value(rowIdx)
	case arrow.TIMESTAMP:
		return col.(*array.Timestamp).Value(rowIdx)
	default:
		return nil
	}
}

// VerifyResult verifies a query result against expected values
func (tc *ManualVerificationTestContext) VerifyResult(result *executor.ResultSet, expected ManualVerificationExpectedResult) {
	tc.t.Helper()

	// Log the test description
	tc.t.Logf("Verifying: %s", expected.Description)

	// Verify result is not nil
	require.NotNil(tc.t, result, "Result should not be nil for: %s", expected.Description)

	// Verify headers
	if len(expected.Headers) > 0 {
		require.Equal(tc.t, len(expected.Headers), len(result.Headers),
			"Header count mismatch for: %s", expected.Description)
		for i, expectedHeader := range expected.Headers {
			assert.Equal(tc.t, expectedHeader, result.Headers[i],
				"Header[%d] mismatch for: %s", i, expected.Description)
		}
	}

	// Verify row count
	totalRows := 0
	batches := result.Batches()
	for _, batch := range batches {
		totalRows += int(batch.NumRows())
	}

	if expected.RowCount >= 0 {
		assert.Equal(tc.t, expected.RowCount, totalRows,
			"Row count mismatch for: %s", expected.Description)
	}

	// Verify data types
	if len(expected.DataTypes) > 0 && len(batches) > 0 {
		batch := batches[0]
		require.Equal(tc.t, int64(len(expected.DataTypes)), batch.NumCols(),
			"Column count mismatch for: %s", expected.Description)
		for i, expectedType := range expected.DataTypes {
			actualType := batch.Column(i).DataType().ID()
			assert.Equal(tc.t, expectedType, actualType,
				"Data type mismatch for column %d in: %s", i, expected.Description)
		}
	}

	// Verify values
	if expected.CheckValues && len(expected.Values) > 0 && len(batches) > 0 {
		allRows := [][]interface{}{}
		for _, batch := range batches {
			for rowIdx := 0; rowIdx < int(batch.NumRows()); rowIdx++ {
				row := make([]interface{}, int(batch.NumCols()))
				for colIdx := 0; colIdx < int(batch.NumCols()); colIdx++ {
					row[colIdx] = GetCellValue(batch, rowIdx, colIdx)
				}
				allRows = append(allRows, row)
			}
		}

		if expected.OrderMatters {
			// Verify rows in exact order
			require.Equal(tc.t, len(expected.Values), len(allRows),
				"Row count mismatch when checking values for: %s", expected.Description)
			for rowIdx, expectedRow := range expected.Values {
				for colIdx, expectedValue := range expectedRow {
					actualValue := allRows[rowIdx][colIdx]
					assert.Equal(tc.t, expectedValue, actualValue,
						"Value mismatch at [%d][%d] for: %s", rowIdx, colIdx, expected.Description)
				}
			}
		} else {
			// Verify all expected rows exist (order doesn't matter)
			for _, expectedRow := range expected.Values {
				found := false
				for _, actualRow := range allRows {
					if rowsEqual(expectedRow, actualRow) {
						found = true
						break
					}
				}
				assert.True(tc.t, found,
					"Expected row %v not found in results for: %s", expectedRow, expected.Description)
			}
		}
	}

	tc.t.Logf("✅ Verified: %s", expected.Description)
}

// rowsEqual checks if two rows are equal
func rowsEqual(row1, row2 []interface{}) bool {
	if len(row1) != len(row2) {
		return false
	}
	for i := range row1 {
		if row1[i] != row2[i] {
			return false
		}
	}
	return true
}

// TestManualVerification tests all scenarios from manual_verification.sql with detailed verification
func TestManualVerification(t *testing.T) {
	tc := NewManualVerificationTestContext(t, "manual_verification_automated")
	defer tc.Cleanup()

	// Setup: Create database and tables
	tc.ExecuteSQLMustSucceed("CREATE DATABASE testdb")
	tc.sess.CurrentDB = "testdb"

	// Test 1: Multi-row INSERT
	t.Run("Test1_MultiRowInsert", func(t *testing.T) {
		tc.ExecuteSQLMustSucceed("CREATE TABLE products (id INT, name VARCHAR, price INT)")
		tc.ExecuteSQLMustSucceed("INSERT INTO products VALUES (1, 'Laptop', 1200), (2, 'Mouse', 25), (3, 'Keyboard', 75)")

		result := tc.ExecuteSQLMustSucceed("SELECT * FROM products")
		tc.VerifyResult(result, ManualVerificationExpectedResult{
			Description:  "Multi-row INSERT should return 3 products",
			Headers:      []string{"id", "name", "price"},
			DataTypes:    []arrow.Type{arrow.INT64, arrow.STRING, arrow.INT64},
			RowCount:     3,
			CheckValues:  true,
			OrderMatters: false,
			Values: [][]interface{}{
				{int64(1), "Laptop", int64(1200)},
				{int64(2), "Mouse", int64(25)},
				{int64(3), "Keyboard", int64(75)},
			},
		})
	})

	// Test 2: Numeric data integrity
	t.Run("Test2_NumericDataIntegrity", func(t *testing.T) {
		tc.ExecuteSQLMustSucceed("CREATE TABLE numbers (id INT, value INT)")
		tc.ExecuteSQLMustSucceed("INSERT INTO numbers VALUES (1, 100)")
		tc.ExecuteSQLMustSucceed("INSERT INTO numbers VALUES (2, 200)")
		tc.ExecuteSQLMustSucceed("INSERT INTO numbers VALUES (3, 300)")

		result := tc.ExecuteSQLMustSucceed("SELECT * FROM numbers")
		tc.VerifyResult(result, ManualVerificationExpectedResult{
			Description:  "Numeric values should be preserved correctly",
			Headers:      []string{"id", "value"},
			DataTypes:    []arrow.Type{arrow.INT64, arrow.INT64},
			RowCount:     3,
			CheckValues:  true,
			OrderMatters: false,
			Values: [][]interface{}{
				{int64(1), int64(100)},
				{int64(2), int64(200)},
				{int64(3), int64(300)},
			},
		})
	})

	// Test 3: DELETE without WHERE
	t.Run("Test3_DeleteWithoutWhere", func(t *testing.T) {
		tc.ExecuteSQLMustSucceed("CREATE TABLE temp_table (id INT, name VARCHAR)")
		tc.ExecuteSQLMustSucceed("INSERT INTO temp_table VALUES (1, 'Test1')")
		tc.ExecuteSQLMustSucceed("INSERT INTO temp_table VALUES (2, 'Test2')")
		tc.ExecuteSQLMustSucceed("INSERT INTO temp_table VALUES (3, 'Test3')")

		// Verify 3 rows before deletion
		result := tc.ExecuteSQLMustSucceed("SELECT * FROM temp_table")
		tc.VerifyResult(result, ManualVerificationExpectedResult{
			Description: "Should have 3 rows before DELETE",
			Headers:     []string{"id", "name"},
			RowCount:    3,
		})

		// DELETE without WHERE should not panic
		tc.ExecuteSQLMustSucceed("DELETE FROM temp_table")

		// Verify all rows deleted
		result = tc.ExecuteSQLMustSucceed("SELECT * FROM temp_table")
		tc.VerifyResult(result, ManualVerificationExpectedResult{
			Description: "Should have 0 rows after DELETE without WHERE",
			Headers:     []string{"id", "name"},
			RowCount:    0,
		})
	})

	// Test 4: DROP TABLE
	t.Run("Test4_DropTable", func(t *testing.T) {
		tc.ExecuteSQLMustSucceed("CREATE TABLE drop_test (id INT)")
		tc.ExecuteSQLMustSucceed("INSERT INTO drop_test VALUES (1)")

		// Verify table exists
		result := tc.ExecuteSQLMustSucceed("SELECT * FROM drop_test")
		tc.VerifyResult(result, ManualVerificationExpectedResult{
			Description: "Table should exist with 1 row",
			Headers:     []string{"id"},
			RowCount:    1,
			CheckValues: true,
			Values: [][]interface{}{
				{int64(1)},
			},
		})

		// Drop table
		tc.ExecuteSQLMustSucceed("DROP TABLE drop_test")

		// Verify table is gone
		tc.ExecuteSQLMustFail("SELECT * FROM drop_test")
	})

	// Test 5: Index operations
	t.Run("Test5_IndexOperations", func(t *testing.T) {
		tc.ExecuteSQLMustSucceed("CREATE TABLE indexed_table (id INT, name VARCHAR)")
		tc.ExecuteSQLMustSucceed("INSERT INTO indexed_table VALUES (1, 'Alice')")
		tc.ExecuteSQLMustSucceed("CREATE INDEX idx_name ON indexed_table (name)")

		// SHOW INDEXES should return the index
		result := tc.ExecuteSQLMustSucceed("SHOW INDEXES ON indexed_table")
		tc.VerifyResult(result, ManualVerificationExpectedResult{
			Description: "Should show idx_name index",
			RowCount:    1,
		})

		tc.ExecuteSQLMustSucceed("DROP INDEX idx_name ON indexed_table")

		// After drop, no indexes should remain
		result = tc.ExecuteSQLMustSucceed("SHOW INDEXES ON indexed_table")
		tc.VerifyResult(result, ManualVerificationExpectedResult{
			Description: "Should show no indexes after DROP INDEX",
			RowCount:    0,
		})
	})

	// Test 6: Complex workflow
	t.Run("Test6_ComplexWorkflow", func(t *testing.T) {
		tc.ExecuteSQLMustSucceed("CREATE TABLE orders (order_id INT, customer VARCHAR, amount INT)")
		tc.ExecuteSQLMustSucceed("INSERT INTO orders VALUES (1, 'Alice', 100), (2, 'Bob', 200), (3, 'Charlie', 150)")

		// Verify 3 rows
		result := tc.ExecuteSQLMustSucceed("SELECT * FROM orders")
		tc.VerifyResult(result, ManualVerificationExpectedResult{
			Description: "Should have 3 orders initially",
			Headers:     []string{"order_id", "customer", "amount"},
			RowCount:    3,
		})

		tc.ExecuteSQLMustSucceed("CREATE INDEX idx_customer ON orders (customer)")
		tc.ExecuteSQLMustSucceed("INSERT INTO orders VALUES (4, 'David', 300), (5, 'Eve', 250)")

		// Verify 5 rows
		result = tc.ExecuteSQLMustSucceed("SELECT * FROM orders")
		tc.VerifyResult(result, ManualVerificationExpectedResult{
			Description: "Should have 5 orders after additional inserts",
			Headers:     []string{"order_id", "customer", "amount"},
			RowCount:    5,
		})

		// DELETE with WHERE
		tc.ExecuteSQLMustSucceed("DELETE FROM orders WHERE order_id = 1")

		// Verify 4 rows (Alice's order removed)
		result = tc.ExecuteSQLMustSucceed("SELECT * FROM orders")
		tc.VerifyResult(result, ManualVerificationExpectedResult{
			Description: "Should have 4 orders after deleting order_id=1",
			Headers:     []string{"order_id", "customer", "amount"},
			RowCount:    4,
		})

		// DELETE without WHERE
		tc.ExecuteSQLMustSucceed("DELETE FROM orders")

		// Verify all rows deleted
		result = tc.ExecuteSQLMustSucceed("SELECT * FROM orders")
		tc.VerifyResult(result, ManualVerificationExpectedResult{
			Description: "Should have 0 orders after DELETE without WHERE",
			Headers:     []string{"order_id", "customer", "amount"},
			RowCount:    0,
		})

		tc.ExecuteSQLMustSucceed("DROP INDEX idx_customer ON orders")
		tc.ExecuteSQLMustSucceed("DROP TABLE orders")
	})

	// Cleanup
	tc.ExecuteSQLMustSucceed("DROP TABLE products")
	tc.ExecuteSQLMustSucceed("DROP TABLE numbers")
	tc.ExecuteSQLMustSucceed("DROP TABLE temp_table")
	tc.ExecuteSQLMustSucceed("DROP TABLE indexed_table")
	tc.ExecuteSQLMustSucceed("DROP DATABASE testdb")

	t.Log("✅ All manual verification tests passed with detailed validation")
}

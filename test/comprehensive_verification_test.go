package test

import (
	"fmt"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
	"github.com/yyun543/minidb/internal/types"
)

// TestComprehensiveVerification tests all SQL features with detailed result verification
func TestComprehensiveVerification(t *testing.T) {
	testDir := SetupTestDir(t, "comprehensive_verification_test")
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

	executeSQL := func(sql string) (*executor.ResultSet, error) {
		stmt, err := parser.Parse(sql)
		if err != nil {
			return nil, err
		}
		plan, err := opt.Optimize(stmt)
		if err != nil {
			return nil, err
		}
		return exec.Execute(plan, sess)
	}

	// Helper function to verify result headers
	verifyHeaders := func(t *testing.T, result *executor.ResultSet, expectedHeaders []string) {
		assert.Equal(t, len(expectedHeaders), len(result.Headers), "Header count mismatch")
		for i, expected := range expectedHeaders {
			assert.Equal(t, expected, result.Headers[i], fmt.Sprintf("Header[%d] mismatch", i))
		}
	}

	// Helper function to verify row count
	verifyRowCount := func(t *testing.T, result *executor.ResultSet, expectedCount int) int {
		totalRows := 0
		for _, batch := range result.Batches() {
			totalRows += int(batch.NumRows())
		}
		assert.Equal(t, expectedCount, totalRows, "Row count mismatch")
		return totalRows
	}

	// Helper function to get cell value from batch
	getCellValue := func(batch *types.Batch, rowIdx, colIdx int) interface{} {
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
		default:
			return nil
		}
	}

	// Helper function to safely convert to int64
	toInt64 := func(val interface{}) int64 {
		switch v := val.(type) {
		case int64:
			return v
		case float64:
			return int64(v)
		case int:
			return int64(v)
		case int32:
			return int64(v)
		default:
			return 0
		}
	}

	// Setup test database
	_, err = executeSQL("CREATE DATABASE testdb")
	assert.NoError(t, err)
	sess.CurrentDB = "testdb"

	// Test 1: Multi-row INSERT with detailed verification
	t.Run("MultiRowInsertDetailed", func(t *testing.T) {
		_, err = executeSQL("CREATE TABLE products (id INT, name VARCHAR, price INT)")
		assert.NoError(t, err)

		// Insert 3 rows
		_, err = executeSQL("INSERT INTO products VALUES (1, 'Laptop', 1200), (2, 'Mouse', 25), (3, 'Keyboard', 75)")
		assert.NoError(t, err)

		// Verify SELECT results with ORDER BY to ensure consistent order
		result, err := executeSQL("SELECT * FROM products ORDER BY id ASC")
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify headers
		verifyHeaders(t, result, []string{"id", "name", "price"})

		// Verify row count
		verifyRowCount(t, result, 3)

		// Verify specific values
		batches := result.Batches()
		if len(batches) > 0 {
			batch := batches[0]

			// First row: id=1, name='Laptop', price=1200
			if batch.NumRows() >= 1 {
				assert.Equal(t, int64(1), getCellValue(batch, 0, 0), "Row 0, id mismatch")
				assert.Equal(t, "Laptop", getCellValue(batch, 0, 1), "Row 0, name mismatch")
				assert.Equal(t, int64(1200), getCellValue(batch, 0, 2), "Row 0, price mismatch")
			}

			// Second row: id=2, name='Mouse', price=25
			if batch.NumRows() >= 2 {
				assert.Equal(t, int64(2), getCellValue(batch, 1, 0), "Row 1, id mismatch")
				assert.Equal(t, "Mouse", getCellValue(batch, 1, 1), "Row 1, name mismatch")
				assert.Equal(t, int64(25), getCellValue(batch, 1, 2), "Row 1, price mismatch")
			}

			// Third row: id=3, name='Keyboard', price=75
			if batch.NumRows() >= 3 {
				assert.Equal(t, int64(3), getCellValue(batch, 2, 0), "Row 2, id mismatch")
				assert.Equal(t, "Keyboard", getCellValue(batch, 2, 1), "Row 2, name mismatch")
				assert.Equal(t, int64(75), getCellValue(batch, 2, 2), "Row 2, price mismatch")
			}
		}

		t.Log("✅ Multi-row INSERT with detailed verification passed")
	})

	// Test 2: WHERE clause with detailed verification
	t.Run("WhereClauseDetailed", func(t *testing.T) {
		// Query with WHERE
		result, err := executeSQL("SELECT * FROM products WHERE price > 50")
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify headers
		verifyHeaders(t, result, []string{"id", "name", "price"})

		// Should return 2 rows (Laptop: 1200, Keyboard: 75)
		verifyRowCount(t, result, 2)

		// Verify values
		batches := result.Batches()
		if len(batches) > 0 {
			batch := batches[0]
			for i := 0; i < int(batch.NumRows()); i++ {
				price := toInt64(getCellValue(batch, i, 2))
				assert.Greater(t, price, int64(50), fmt.Sprintf("Row %d price should be > 50", i))
			}
		}

		t.Log("✅ WHERE clause with detailed verification passed")
	})

	// Test 3: ORDER BY with detailed verification
	t.Run("OrderByDetailed", func(t *testing.T) {
		result, err := executeSQL("SELECT * FROM products ORDER BY price ASC")
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify headers
		verifyHeaders(t, result, []string{"id", "name", "price"})

		// Verify row count
		verifyRowCount(t, result, 3)

		// Verify order: Mouse (25), Keyboard (75), Laptop (1200)
		batches := result.Batches()
		if len(batches) > 0 {
			batch := batches[0]
			if batch.NumRows() >= 3 {
				price0 := getCellValue(batch, 0, 2).(int64)
				price1 := getCellValue(batch, 1, 2).(int64)
				price2 := getCellValue(batch, 2, 2).(int64)

				assert.Equal(t, int64(25), price0, "First row should be Mouse (25)")
				assert.Equal(t, int64(75), price1, "Second row should be Keyboard (75)")
				assert.Equal(t, int64(1200), price2, "Third row should be Laptop (1200)")

				// Verify ascending order
				assert.Less(t, price0, price1, "Prices should be in ascending order")
				assert.Less(t, price1, price2, "Prices should be in ascending order")
			}
		}

		t.Log("✅ ORDER BY with detailed verification passed")
	})

	// Test 4: Aggregate functions with detailed verification
	t.Run("AggregateDetailed", func(t *testing.T) {
		result, err := executeSQL("SELECT COUNT(*) AS cnt, SUM(price) AS total, AVG(price) AS avg_price FROM products")
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify headers
		verifyHeaders(t, result, []string{"cnt", "total", "avg_price"})

		// Verify single row result
		verifyRowCount(t, result, 1)

		// Verify aggregate values
		batches := result.Batches()
		if len(batches) > 0 {
			batch := batches[0]
			if batch.NumRows() >= 1 {
				countVal := getCellValue(batch, 0, 0)
				sumVal := getCellValue(batch, 0, 1)
				avgVal := getCellValue(batch, 0, 2)

				// COUNT(*) result
				if count, ok := countVal.(int64); ok {
					assert.Equal(t, int64(3), count, "COUNT(*) should be 3")
				} else if countFloat, ok := countVal.(float64); ok {
					assert.Equal(t, 3.0, countFloat, "COUNT(*) should be 3")
				}

				// SUM(price) result - could be int64 or float64 depending on implementation
				if sum, ok := sumVal.(int64); ok {
					assert.Equal(t, int64(1300), sum, "SUM(price) should be 1300 (1200+25+75)")
				} else if sumFloat, ok := sumVal.(float64); ok {
					assert.InDelta(t, 1300.0, sumFloat, 0.01, "SUM(price) should be 1300 (1200+25+75)")
				}

				// AVG(price) result - should be float64
				if avg, ok := avgVal.(float64); ok {
					assert.InDelta(t, 433.33, avg, 0.01, "AVG(price) should be ~433.33")
				} else if avgInt, ok := avgVal.(int64); ok {
					assert.InDelta(t, 433.33, float64(avgInt), 0.01, "AVG(price) should be ~433.33")
				}
			}
		}

		t.Log("✅ Aggregate functions with detailed verification passed")
	})

	// Test 5: DELETE without WHERE with detailed verification
	t.Run("DeleteWithoutWhereDetailed", func(t *testing.T) {
		_, err = executeSQL("CREATE TABLE temp_table (id INT, name VARCHAR)")
		assert.NoError(t, err)

		// Insert test data
		_, err = executeSQL("INSERT INTO temp_table VALUES (1, 'Test1'), (2, 'Test2'), (3, 'Test3')")
		assert.NoError(t, err)

		// Verify 3 rows exist
		result, err := executeSQL("SELECT * FROM temp_table")
		assert.NoError(t, err)
		verifyRowCount(t, result, 3)

		// DELETE without WHERE
		_, err = executeSQL("DELETE FROM temp_table")
		assert.NoError(t, err, "DELETE without WHERE should not panic")

		// Verify all rows deleted
		result, err = executeSQL("SELECT * FROM temp_table")
		assert.NoError(t, err)
		verifyRowCount(t, result, 0)

		t.Log("✅ DELETE without WHERE with detailed verification passed")
	})

	// Test 6: GROUP BY with detailed verification
	t.Run("GroupByDetailed", func(t *testing.T) {
		_, err = executeSQL("CREATE TABLE sales (category VARCHAR, amount INT)")
		assert.NoError(t, err)

		_, err = executeSQL("INSERT INTO sales VALUES ('Electronics', 100), ('Electronics', 200), ('Furniture', 150)")
		assert.NoError(t, err)

		result, err := executeSQL("SELECT category, COUNT(*) AS cnt, SUM(amount) AS total FROM sales GROUP BY category")
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify headers
		verifyHeaders(t, result, []string{"category", "cnt", "total"})

		// Should have 2 groups
		verifyRowCount(t, result, 2)

		// Verify group values
		batches := result.Batches()
		if len(batches) > 0 {
			batch := batches[0]
			foundElectronics := false
			foundFurniture := false

			for i := 0; i < int(batch.NumRows()); i++ {
				category := getCellValue(batch, i, 0).(string)
				count := toInt64(getCellValue(batch, i, 1))
				total := toInt64(getCellValue(batch, i, 2))

				if category == "Electronics" {
					foundElectronics = true
					assert.Equal(t, int64(2), count, "Electronics should have 2 items")
					assert.Equal(t, int64(300), total, "Electronics total should be 300")
				} else if category == "Furniture" {
					foundFurniture = true
					assert.Equal(t, int64(1), count, "Furniture should have 1 item")
					assert.Equal(t, int64(150), total, "Furniture total should be 150")
				}
			}

			assert.True(t, foundElectronics, "Should find Electronics group")
			assert.True(t, foundFurniture, "Should find Furniture group")
		}

		t.Log("✅ GROUP BY with detailed verification passed")
	})

	// Test 7: DROP TABLE with detailed verification
	t.Run("DropTableDetailed", func(t *testing.T) {
		_, err = executeSQL("CREATE TABLE drop_test (id INT)")
		assert.NoError(t, err)

		_, err = executeSQL("INSERT INTO drop_test VALUES (1)")
		assert.NoError(t, err)

		// Verify table exists
		result, err := executeSQL("SELECT * FROM drop_test")
		assert.NoError(t, err)
		verifyRowCount(t, result, 1)

		// Drop table
		_, err = executeSQL("DROP TABLE drop_test")
		assert.NoError(t, err)

		// Verify table is gone
		_, err = executeSQL("SELECT * FROM drop_test")
		assert.Error(t, err, "Table should not exist after DROP")

		t.Log("✅ DROP TABLE with detailed verification passed")
	})

	t.Log("✅ All comprehensive verification tests passed")
}

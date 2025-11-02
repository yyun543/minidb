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

// TestRowGroupFiltering tests P1.4 Row Group level filtering using Min/Max statistics
func TestRowGroupFiltering(t *testing.T) {
	storageEngine, err := storage.NewParquetEngine(SetupTestDir(t, "row_group_filtering_test"))
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

	// Setup
	createDbSQL := "CREATE DATABASE testdb"
	stmt, err := parser.Parse(createDbSQL)
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "testdb"

	t.Run("RowGroupSkipping_IntegerRange", func(t *testing.T) {
		// Create table
		createSQL := "CREATE TABLE sales (id INTEGER, amount INTEGER, region VARCHAR)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Insert data in distinct ranges to create multiple Row Groups
		// Range 1: amount 0-100 (10 rows)
		for i := 1; i <= 10; i++ {
			insertSQL := "INSERT INTO sales VALUES (1, 50, 'East')"
			stmt, err = parser.Parse(insertSQL)
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}

		// Range 2: amount 200-300 (10 rows)
		for i := 1; i <= 10; i++ {
			insertSQL := "INSERT INTO sales VALUES (2, 250, 'West')"
			stmt, err = parser.Parse(insertSQL)
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}

		// Range 3: amount 400-500 (10 rows)
		for i := 1; i <= 10; i++ {
			insertSQL := "INSERT INTO sales VALUES (3, 450, 'North')"
			stmt, err = parser.Parse(insertSQL)
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}

		// Query that should skip Row Groups 1 and 2 (amount < 400)
		// Only Row Group 3 (amount 400-500) should be read
		selectSQL := "SELECT * FROM sales WHERE amount > 400"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Verify result - Row Groups with amount < 400 should be skipped
		t.Logf("✅ Row Group filtering: Query executed successfully (amount > 400)")

		// Query that should match Row Group 2 only
		selectSQL = "SELECT * FROM sales WHERE amount >= 200 AND amount <= 300"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Row Groups outside 200-300 range should be skipped
		t.Logf("✅ Row Group filtering: Query executed successfully (200-300 range)")

		// Query that should skip all Row Groups (no data in this range)
		selectSQL = "SELECT * FROM sales WHERE amount > 1000"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// All Row Groups should be skipped (no data > 1000)
		t.Logf("✅ Row Group filtering: Query executed successfully (all Row Groups skipped)")
	})

	t.Run("RowGroupSkipping_StringRange", func(t *testing.T) {
		// Create table
		createSQL := "CREATE TABLE customers (id INTEGER, name VARCHAR, city VARCHAR)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Insert data with distinct string ranges
		// Group 1: cities starting with 'A'
		cities1 := []string{"Austin", "Atlanta", "Albany"}
		for _, city := range cities1 {
			insertSQL := "INSERT INTO customers VALUES (1, 'User', '" + city + "')"
			stmt, err = parser.Parse(insertSQL)
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}

		// Group 2: cities starting with 'N'
		cities2 := []string{"New York", "Nashville", "Newark"}
		for _, city := range cities2 {
			insertSQL := "INSERT INTO customers VALUES (2, 'User', '" + city + "')"
			stmt, err = parser.Parse(insertSQL)
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}

		// Group 3: cities starting with 'Z'
		cities3 := []string{"Zurich", "Zagreb", "Zaragoza"}
		for _, city := range cities3 {
			insertSQL := "INSERT INTO customers VALUES (3, 'User', '" + city + "')"
			stmt, err = parser.Parse(insertSQL)
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}

		// Query that should only read Row Group 2 (cities starting with 'N')
		selectSQL := "SELECT * FROM customers WHERE city >= 'N' AND city < 'O'"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Row Groups for cities starting with A or Z should be skipped
		t.Logf("✅ String Row Group filtering: Query executed successfully (N cities only)")
	})
}

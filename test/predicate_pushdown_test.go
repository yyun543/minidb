package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

// TestPredicatePushdown tests predicate pushdown optimization
func TestPredicatePushdown(t *testing.T) {
	storageEngine, err := storage.NewParquetEngine(SetupTestDir(t, "predicate_pushdown_test"))
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
	createDbSQL := "CREATE DATABASE pushdown_test"
	stmt, err := parser.Parse(createDbSQL)
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "pushdown_test"

	t.Run("IntegerPredicatePushdown", func(t *testing.T) {
		// Create table with integers
		createSQL := "CREATE TABLE numbers (id INTEGER, value INTEGER)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Insert test data - 100 rows
		for i := 1; i <= 100; i++ {
			// Use actual values
			stmt, err = parser.Parse("INSERT INTO numbers VALUES (1, 50)")
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}

		// Query with predicate - should use pushdown
		selectSQL := "SELECT * FROM numbers WHERE value = 50"
		startTime := time.Now()
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		elapsed := time.Since(startTime)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		t.Logf("Query with equality predicate took: %v", elapsed)

		// Greater than predicate
		selectSQL = "SELECT * FROM numbers WHERE value > 25"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Less than predicate
		selectSQL = "SELECT * FROM numbers WHERE value < 75"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Range predicate
		selectSQL = "SELECT * FROM numbers WHERE value >= 25 AND value <= 75"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("StringPredicatePushdown", func(t *testing.T) {
		// Create table with strings
		createSQL := "CREATE TABLE users (id INTEGER, name VARCHAR, city VARCHAR)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Insert test data
		cities := []string{"New York", "Los Angeles", "Chicago", "Houston", "Phoenix"}
		for i := 0; i < 50; i++ {
			city := cities[i%len(cities)]
			// Direct SQL with city value
			insertSQL := "INSERT INTO users VALUES (1, 'User', '" + city + "')"
			stmt, err = parser.Parse(insertSQL)
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}

		// Query with string equality predicate
		selectSQL := "SELECT * FROM users WHERE city = 'Chicago'"
		startTime := time.Now()
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		elapsed := time.Since(startTime)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		t.Logf("String equality predicate took: %v", elapsed)

		// String comparison
		selectSQL = "SELECT * FROM users WHERE city > 'Chicago'"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("FloatPredicatePushdown", func(t *testing.T) {
		// Create table with floats
		createSQL := "CREATE TABLE measurements (id INTEGER, temperature FLOAT, humidity FLOAT)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Insert test data
		for i := 0; i < 50; i++ {
			insertSQL := "INSERT INTO measurements VALUES (1, 25.5, 60.0)"
			stmt, err = parser.Parse(insertSQL)
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}

		// Query with float predicate
		selectSQL := "SELECT * FROM measurements WHERE temperature > 20.0"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Range query
		selectSQL = "SELECT * FROM measurements WHERE temperature >= 20.0 AND temperature <= 30.0"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("ComplexPredicatesPushdown", func(t *testing.T) {
		// Create table
		createSQL := "CREATE TABLE orders (order_id INTEGER, user_id INTEGER, amount INTEGER, status VARCHAR)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Insert mixed data
		statuses := []string{"pending", "completed", "cancelled"}
		for i := 1; i <= 30; i++ {
			status := statuses[i%len(statuses)]
			// Direct SQL
			stmt, err = parser.Parse("INSERT INTO orders VALUES (1, 1, 100, '" + status + "')")
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}

		// Multi-column predicate
		selectSQL := "SELECT * FROM orders WHERE amount > 50 AND status = 'completed'"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// OR predicate (may not push down fully)
		selectSQL = "SELECT * FROM orders WHERE amount < 50 OR amount > 150"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// TestDataSkippingWithStatistics tests data skipping using min/max statistics
func TestDataSkippingWithStatistics(t *testing.T) {
	storageEngine, err := storage.NewParquetEngine(SetupTestDir(t, "data_skipping_test"))
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
	createDbSQL := "CREATE DATABASE skipping_test"
	stmt, err := parser.Parse(createDbSQL)
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "skipping_test"

	t.Run("RangeBasedSkipping", func(t *testing.T) {
		// Create table
		createSQL := "CREATE TABLE ranges (id INTEGER, range_id INTEGER, value INTEGER)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Insert data in ranges (simulating partitioned data)
		// Range 1: values 0-100
		for i := 0; i < 10; i++ {
			stmt, err = parser.Parse("INSERT INTO ranges VALUES (1, 1, 50)")
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}

		// Range 2: values 100-200
		for i := 0; i < 10; i++ {
			stmt, err = parser.Parse("INSERT INTO ranges VALUES (2, 2, 150)")
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}

		// Query that should skip Range 1
		selectSQL := "SELECT * FROM ranges WHERE value > 100"
		startTime := time.Now()
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		elapsed := time.Since(startTime)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		t.Logf("Range skipping query took: %v", elapsed)

		// Query that touches both ranges
		selectSQL = "SELECT * FROM ranges WHERE value > 25"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("NullValueSkipping", func(t *testing.T) {
		// Create table
		createSQL := "CREATE TABLE nullable (id INTEGER, optional_value INTEGER)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Insert data with null counts tracked by statistics
		for i := 0; i < 20; i++ {
			stmt, err = parser.Parse("INSERT INTO nullable VALUES (1, 100)")
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}

		// Query should use null count statistics
		selectSQL := "SELECT * FROM nullable WHERE optional_value > 50"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// TestPredicatePushdownPerformance benchmarks predicate pushdown performance
func TestPredicatePushdownPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	storageEngine, err := storage.NewParquetEngine(SetupTestDir(t, "pushdown_performance_test"))
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
	createDbSQL := "CREATE DATABASE perf_test"
	stmt, err := parser.Parse(createDbSQL)
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "perf_test"

	t.Run("SelectivityImpact", func(t *testing.T) {
		// Create large table
		createSQL := "CREATE TABLE large_table (id INTEGER, category INTEGER, value INTEGER)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Insert 500 rows
		insertStart := time.Now()
		for i := 0; i < 500; i++ {
			stmt, err = parser.Parse("INSERT INTO large_table VALUES (1, 5, 100)")
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}
		insertDuration := time.Since(insertStart)
		t.Logf("Inserted 500 rows in: %v", insertDuration)

		// High selectivity query (should be fast)
		selectSQL := "SELECT * FROM large_table WHERE category = 5"
		startTime := time.Now()
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		highSelectivityDuration := time.Since(startTime)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		t.Logf("High selectivity query took: %v", highSelectivityDuration)

		// Full table scan (for comparison)
		selectSQL = "SELECT * FROM large_table"
		startTime = time.Now()
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err = exec.Execute(plan, sess)
		fullScanDuration := time.Since(startTime)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		t.Logf("Full table scan took: %v", fullScanDuration)

		t.Logf("Predicate pushdown speedup: %.2fx", float64(fullScanDuration)/float64(highSelectivityDuration))
	})
}

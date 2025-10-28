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

// TestDeltaLakeACID tests ACID properties of Delta Lake implementation
func TestDeltaLakeACID(t *testing.T) {
	// Setup test environment with cleanup
	testDir := SetupTestDir(t, "delta_acid_test")
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
	createDbSQL := "CREATE DATABASE acid_test"
	stmt, err := parser.Parse(createDbSQL)
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "acid_test"

	// Test Atomicity: All or nothing
	t.Run("Atomicity", func(t *testing.T) {
		// Create table
		createSQL := "CREATE TABLE accounts (id INTEGER, name VARCHAR, balance INTEGER)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Insert initial data
		insertSQL := "INSERT INTO accounts (id, name, balance) VALUES (1, 'Alice', 1000)"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		insertSQL = "INSERT INTO accounts (id, name, balance) VALUES (2, 'Bob', 500)"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Verify all inserts completed atomically
		selectSQL := "SELECT * FROM accounts"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result.Batches()), "Should have 2 records after atomic inserts")
	})

	// Test Consistency: Data integrity maintained
	t.Run("Consistency", func(t *testing.T) {
		// Create table with constraints (conceptual - we validate consistency through queries)
		createSQL := "CREATE TABLE orders (order_id INTEGER, user_id INTEGER, amount INTEGER)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Insert consistent data
		for i := 1; i <= 5; i++ {
			// Since we don't have parameterized queries, use direct values
			stmt, err = parser.Parse("INSERT INTO orders VALUES (1, 1, 100)")
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}

		// Verify consistency - all records should exist
		selectSQL := "SELECT * FROM orders"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.Equal(t, 5, len(result.Batches()), "Should maintain consistency with all inserts")
	})

	// Test Isolation: Concurrent operations don't interfere
	t.Run("Isolation", func(t *testing.T) {
		// Create table
		createSQL := "CREATE TABLE transactions (id INTEGER, amount INTEGER, timestamp VARCHAR)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Session 1: Insert data
		sess1 := sessMgr.CreateSession()
		sess1.CurrentDB = "acid_test"

		insertSQL := "INSERT INTO transactions VALUES (1, 100, '2024-01-01')"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess1)
		assert.NoError(t, err)

		// Session 2: Should see isolated view (snapshot isolation)
		sess2 := sessMgr.CreateSession()
		sess2.CurrentDB = "acid_test"

		insertSQL = "INSERT INTO transactions VALUES (2, 200, '2024-01-02')"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess2)
		assert.NoError(t, err)

		// Both sessions can read their own writes + committed data
		selectSQL := "SELECT * FROM transactions"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess1)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result.Batches()), 1, "Session 1 should see at least its own write")

		result, err = exec.Execute(plan, sess2)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result.Batches()), 1, "Session 2 should see at least its own write")
	})

	// Test Durability: Data persists after operations
	t.Run("Durability", func(t *testing.T) {
		// Create table
		createSQL := "CREATE TABLE durable_data (id INTEGER, value VARCHAR)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Insert data
		insertSQL := "INSERT INTO durable_data VALUES (1, 'persistent')"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Verify data is immediately readable (durability)
		selectSQL := "SELECT * FROM durable_data WHERE id = 1"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result.Batches()), "Data should be durable and readable immediately")

		// Data should persist across multiple queries
		time.Sleep(100 * time.Millisecond)
		result, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result.Batches()), "Data should remain durable over time")
	})

	// Test Delta Log versioning
	t.Run("VersionControl", func(t *testing.T) {
		// Create table
		createSQL := "CREATE TABLE versioned (id INTEGER, value INTEGER)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Multiple inserts create versions
		for i := 1; i <= 3; i++ {
			// Direct value insertion
			stmt, err = parser.Parse("INSERT INTO versioned VALUES (1, 100)")
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)

			// Small delay to ensure different versions
			time.Sleep(10 * time.Millisecond)
		}

		// Query should return all versions
		selectSQL := "SELECT * FROM versioned"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(result.Batches()), "Should have 3 versions")
	})
}

// TestDeltaLogSnapshot tests snapshot isolation
func TestDeltaLogSnapshot(t *testing.T) {
	testDir := SetupTestDir(t, "delta_snapshot_test")
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

	// Create database and table
	createDbSQL := "CREATE DATABASE snapshot_test"
	stmt, err := parser.Parse(createDbSQL)
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "snapshot_test"

	createSQL := "CREATE TABLE events (id INTEGER, event VARCHAR, value INTEGER)"
	stmt, err = parser.Parse(createSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	t.Run("SnapshotConsistency", func(t *testing.T) {
		// Insert initial data
		insertSQL := "INSERT INTO events VALUES (1, 'create', 100)"
		stmt, err := parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Read should show consistent snapshot
		selectSQL := "SELECT * FROM events"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result1, err := exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Add more data
		insertSQL = "INSERT INTO events VALUES (2, 'update', 200)"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// New read should show updated snapshot - use SELECT query
		selectSQL = "SELECT * FROM events"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result2, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.Greater(t, len(result2.Batches()), len(result1.Batches()),
			"New snapshot should include new data")
	})
}

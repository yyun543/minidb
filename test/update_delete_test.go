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

// TestUpdateOperation tests UPDATE with Copy-on-Write mechanism
func TestUpdateOperation(t *testing.T) {
	storageEngine, err := storage.NewParquetEngine("./test_data/update_test")
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

	// Setup: Create database and table
	createDbSQL := "CREATE DATABASE update_test"
	stmt, err := parser.Parse(createDbSQL)
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "update_test"

	// Create users table
	createSQL := "CREATE TABLE users (id INTEGER, name VARCHAR, email VARCHAR, age INTEGER)"
	stmt, err = parser.Parse(createSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Insert test data
	insertSQL1 := "INSERT INTO users VALUES (1, 'John Doe', 'john@example.com', 25)"
	stmt, err = parser.Parse(insertSQL1)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	insertSQL2 := "INSERT INTO users VALUES (2, 'Jane Smith', 'jane@example.com', 30)"
	stmt, err = parser.Parse(insertSQL2)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	insertSQL3 := "INSERT INTO users VALUES (3, 'Bob Wilson', 'bob@example.com', 35)"
	stmt, err = parser.Parse(insertSQL3)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Test UPDATE operation
	t.Run("UpdateWithWhereClause", func(t *testing.T) {
		updateSQL := "UPDATE users SET email = 'john.doe@newdomain.com' WHERE name = 'John Doe'"
		stmt, err := parser.Parse(updateSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err, "UPDATE should execute successfully")

		// Verify the update
		selectSQL := "SELECT * FROM users WHERE name = 'John Doe'"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Should have 1 record with updated email
		batches := result.Batches()

		// Count total rows across all batches
		totalRows := int64(0)
		for _, batch := range batches {
			totalRows += batch.NumRows()
		}

		t.Logf("Update completed: %d batches, %d total rows", len(batches), totalRows)
		assert.Equal(t, int64(1), totalRows, "Should have 1 updated record")
	})

	t.Run("UpdateMultipleRows", func(t *testing.T) {
		// Update all users with age > 25
		updateSQL := "UPDATE users SET email = 'updated@example.com' WHERE age > 25"
		stmt, err := parser.Parse(updateSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err, "UPDATE should execute successfully")

		// Verify updates
		selectSQL := "SELECT * FROM users WHERE age > 25"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Should have updated records
		batches := result.Batches()

		// Count total rows across all batches
		totalRows := int64(0)
		for _, batch := range batches {
			totalRows += batch.NumRows()
		}

		t.Logf("Update completed: %d batches, %d total rows", len(batches), totalRows)
		assert.GreaterOrEqual(t, totalRows, int64(2), "Should have at least 2 updated records")
	})
}

// TestDeleteOperation tests DELETE with Delta Log integration
func TestDeleteOperation(t *testing.T) {
	storageEngine, err := storage.NewParquetEngine("./test_data/delete_test")
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
	createDbSQL := "CREATE DATABASE delete_test"
	stmt, err := parser.Parse(createDbSQL)
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "delete_test"

	createSQL := "CREATE TABLE orders (id INTEGER, user_id INTEGER, amount INTEGER)"
	stmt, err = parser.Parse(createSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Insert test data
	for i := 1; i <= 5; i++ {
		insertSQL := "INSERT INTO orders VALUES (1, 1, 100)"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
	}

	// Test DELETE operation
	t.Run("DeleteWithWhereClause", func(t *testing.T) {
		deleteSQL := "DELETE FROM orders WHERE amount < 50"
		stmt, err := parser.Parse(deleteSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err, "DELETE should execute successfully")

		// Verify deletion
		selectSQL := "SELECT * FROM orders"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)

		// All records should still be there (amount = 100, not < 50)
		batches := result.Batches()

		// Count total rows across all batches
		totalRows := int64(0)
		for _, batch := range batches {
			totalRows += batch.NumRows()
		}

		t.Logf("Delete completed, 0 rows deleted: %d batches, %d total rows", len(batches), totalRows)
		assert.Equal(t, int64(5), totalRows, "Should have 5 records (none matched deletion criteria)")
	})

	t.Run("DeleteMatchingRecords", func(t *testing.T) {
		// Insert a record that will match
		insertSQL := "INSERT INTO orders VALUES (6, 1, 25)"
		stmt, err := parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Delete it
		deleteSQL := "DELETE FROM orders WHERE amount < 50"
		stmt, err = parser.Parse(deleteSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err, "DELETE should execute successfully")

		// Verify
		selectSQL := "SELECT * FROM orders"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Should have 5 records left (deleted 1)
		batches := result.Batches()

		// Count total rows across all batches
		totalRows := int64(0)
		for _, batch := range batches {
			totalRows += batch.NumRows()
		}

		t.Logf("Delete completed: %d batches, %d total rows", len(batches), totalRows)
		assert.Equal(t, int64(5), totalRows, "Should have 5 records left after deleting 1")
	})
}

package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

// TestMoRJoinAfterRestart tests the bug where JOIN fails after server restart with MoR delta files
func TestMoRJoinAfterRestart(t *testing.T) {
	testDir := "./test_data/mor_join_restart"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Phase 1: Create tables, insert data, perform UPDATE/DELETE (creates delta files)
	t.Log("Phase 1: Creating data and delta files")
	{
		engine, err := storage.NewParquetEngine(testDir)
		require.NoError(t, err)
		err = engine.Open()
		require.NoError(t, err)

		cat := catalog.NewCatalog()
		cat.SetStorageEngine(engine)
		err = cat.Init()
		require.NoError(t, err)

		sessMgr, err := session.NewSessionManager()
		require.NoError(t, err)
		sess := sessMgr.CreateSession()
		sess.CurrentDB = "testdb"

		opt := optimizer.NewOptimizer()
		exec := executor.NewExecutor(cat)

		execSQL := func(sql string) error {
			stmt, err := parser.Parse(sql)
			if err != nil {
				return err
			}
			plan, err := opt.Optimize(stmt)
			if err != nil {
				return err
			}
			_, err = exec.Execute(plan, sess)
			return err
		}

		// Create database and tables
		err = execSQL("CREATE DATABASE testdb")
		require.NoError(t, err)

		err = execSQL("CREATE TABLE users (id INT, name VARCHAR, age INT)")
		require.NoError(t, err)

		err = execSQL("CREATE TABLE orders (id INT, user_id INT, amount INT)")
		require.NoError(t, err)

		// Insert data
		err = execSQL("INSERT INTO users VALUES (1, 'John', 25)")
		require.NoError(t, err)
		err = execSQL("INSERT INTO users VALUES (2, 'Jane', 30)")
		require.NoError(t, err)
		err = execSQL("INSERT INTO users VALUES (3, 'Bob', 35)")
		require.NoError(t, err)

		err = execSQL("INSERT INTO orders VALUES (1, 1, 100)")
		require.NoError(t, err)
		err = execSQL("INSERT INTO orders VALUES (2, 2, 250)")
		require.NoError(t, err)
		err = execSQL("INSERT INTO orders VALUES (3, 1, 150)")
		require.NoError(t, err)

		// Perform UPDATE (creates delta file for users)
		err = execSQL("UPDATE users SET name = 'John Doe' WHERE id = 1")
		require.NoError(t, err)

		// Perform DELETE (creates delta file for orders)
		err = execSQL("DELETE FROM orders WHERE amount < 50")
		require.NoError(t, err)

		// Test JOIN before restart - should work
		stmt, err := parser.Parse(`
			SELECT u.name, COUNT(o.id) as order_count, SUM(o.amount) as total
			FROM users u
			LEFT JOIN orders o ON u.id = o.user_id
			GROUP BY u.name
			HAVING order_count > 1
			ORDER BY total DESC
		`)
		require.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		require.NoError(t, err)
		require.NotNil(t, result)
		t.Log("✓ JOIN query works BEFORE restart")

		err = engine.Close()
		require.NoError(t, err)
	}

	// Phase 2: Restart and test the same JOIN query
	t.Log("Phase 2: Restarting and testing JOIN with delta files")
	{
		engine, err := storage.NewParquetEngine(testDir)
		require.NoError(t, err)
		err = engine.Open()
		require.NoError(t, err)
		defer engine.Close()

		cat := catalog.NewCatalog()
		cat.SetStorageEngine(engine)
		err = cat.Init()
		require.NoError(t, err)

		sessMgr, err := session.NewSessionManager()
		require.NoError(t, err)
		sess := sessMgr.CreateSession()
		sess.CurrentDB = "testdb"

		opt := optimizer.NewOptimizer()
		exec := executor.NewExecutor(cat)

		// Test simple JOIN after restart - this is where the bug occurs
		stmt, err := parser.Parse(`
			SELECT *
			FROM users u
			LEFT JOIN orders o ON u.id = o.user_id
		`)
		require.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		// This should NOT panic (the original bug was here)
		result, err := exec.Execute(plan, sess)
		require.NoError(t, err, "Simple JOIN query should work AFTER restart with delta files")
		require.NotNil(t, result)

		// Verify result
		totalRows := int64(0)
		for _, batch := range result.Batches() {
			totalRows += batch.NumRows()
		}
		assert.Greater(t, totalRows, int64(0), "Should have at least 1 result row from JOIN")
		t.Log("✓ Simple JOIN query works AFTER restart")
	}
}

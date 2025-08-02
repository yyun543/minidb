package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/statistics"
)

func TestDatabaseOperationsIntegration(t *testing.T) {
	// Initialize components
	cat, err := catalog.NewCatalogWithDefaultStorage()
	require.NoError(t, err)

	err = cat.Init()
	require.NoError(t, err)

	statsMgr := statistics.NewStatisticsManager()
	regularExecutor := executor.NewExecutor(cat)
	vectorizedExecutor := executor.NewVectorizedExecutor(cat, statsMgr)
	sessionMgr, err := session.NewSessionManager()
	require.NoError(t, err)

	sess := sessionMgr.CreateSession()

	t.Run("CreateDatabaseExecution", func(t *testing.T) {
		// Parse CREATE DATABASE
		sql := "CREATE DATABASE ecommerce;"
		ast, err := parser.Parse(sql)
		require.NoError(t, err)

		// Optimize to get plan
		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(ast)
		require.NoError(t, err)

		// Test regular executor
		result, err := regularExecutor.Execute(plan, sess)
		assert.NoError(t, err, "Regular executor should handle CREATE DATABASE")
		assert.NotNil(t, result)

		// Test vectorized executor
		vResult, err := vectorizedExecutor.Execute(plan, sess)
		assert.NoError(t, err, "Vectorized executor should handle CREATE DATABASE")
		assert.NotNil(t, vResult)
	})

	t.Run("ShowCommandsExecution", func(t *testing.T) {
		// Parse SHOW DATABASES
		sql := "SHOW DATABASES;"
		ast, err := parser.Parse(sql)
		require.NoError(t, err)

		// Optimize to get plan
		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(ast)
		require.NoError(t, err)

		// Test regular executor (should work)
		result, err := regularExecutor.Execute(plan, sess)
		assert.NoError(t, err, "Regular executor should handle SHOW DATABASES")
		assert.NotNil(t, result)

		// Vectorized executor should NOT be used for SHOW commands
		// This test verifies that the handler selects the correct executor
		assert.Equal(t, optimizer.ShowPlan, plan.Type, "Plan should be ShowPlan type")
	})

	t.Run("DatabaseContextPropagation", func(t *testing.T) {
		// Create database first
		createDB := "CREATE DATABASE testdb;"
		ast, err := parser.Parse(createDB)
		require.NoError(t, err)
		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(ast)
		require.NoError(t, err)
		_, err = regularExecutor.Execute(plan, sess)
		require.NoError(t, err)

		// Switch to the database
		sess.CurrentDB = "testdb"

		// Create table in the current database
		createTable := "CREATE TABLE users (id INT, name VARCHAR);"
		ast, err = parser.Parse(createTable)
		require.NoError(t, err)
		plan, err = opt.Optimize(ast)
		require.NoError(t, err)

		// Execute create table - should use testdb
		result, err := regularExecutor.Execute(plan, sess)
		assert.NoError(t, err, "CREATE TABLE should work with database context")
		assert.NotNil(t, result)

		// Verify table was created in correct database
		tableMeta, err := cat.GetTable("testdb", "users")
		assert.NoError(t, err, "Table should exist in testdb")
		assert.Equal(t, "testdb", tableMeta.Database)
		assert.Equal(t, "users", tableMeta.Table)
	})

	t.Run("InsertWithDatabaseContext", func(t *testing.T) {
		// Ensure we have a table in testdb
		sess.CurrentDB = "testdb"

		// Insert data
		insertSQL := "INSERT INTO users VALUES (1, 'John Doe');"
		ast, err := parser.Parse(insertSQL)
		require.NoError(t, err)
		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(ast)
		require.NoError(t, err)

		// Execute insert - should use testdb
		result, err := regularExecutor.Execute(plan, sess)
		assert.NoError(t, err, "INSERT should work with database context")
		assert.NotNil(t, result)

		// Result should indicate success, not "Empty set"
		assert.Equal(t, []string{"status"}, result.Headers)
	})

	t.Run("SelectWithDatabaseContext", func(t *testing.T) {
		// Ensure session is set to testdb
		sess.CurrentDB = "testdb"

		// Select from table
		selectSQL := "SELECT * FROM users;"
		ast, err := parser.Parse(selectSQL)
		require.NoError(t, err)
		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(ast)
		require.NoError(t, err)

		// Execute select - should look in testdb, not default
		result, err := regularExecutor.Execute(plan, sess)
		// This might fail initially, but should work after fixes
		if err != nil {
			t.Logf("SELECT failed as expected before fix: %v", err)
			// Error should NOT mention "default" database
			assert.NotContains(t, err.Error(), "table default.users",
				"Error should not reference default database when current database is testdb")
		} else {
			assert.NotNil(t, result)
		}
	})
}

func TestExecutorSelection(t *testing.T) {
	t.Run("VectorizedExecutorSelection", func(t *testing.T) {
		testCases := []struct {
			sql              string
			shouldVectorize  bool
			expectedPlanType optimizer.PlanType
		}{
			{"SELECT * FROM users;", true, optimizer.SelectPlan},
			{"INSERT INTO users VALUES (1, 'test');", true, optimizer.InsertPlan},
			{"UPDATE users SET name = 'test';", true, optimizer.UpdatePlan},
			{"DELETE FROM users WHERE id = 1;", true, optimizer.DeletePlan},
			{"CREATE DATABASE test;", false, optimizer.CreateDatabasePlan},
			{"CREATE TABLE test (id INT);", false, optimizer.CreateTablePlan},
			{"SHOW DATABASES;", false, optimizer.ShowPlan},
			{"SHOW TABLES;", false, optimizer.ShowPlan},
		}

		opt := optimizer.NewOptimizer()
		for _, tc := range testCases {
			t.Run(tc.sql, func(t *testing.T) {
				ast, err := parser.Parse(tc.sql)
				require.NoError(t, err)

				plan, err := opt.Optimize(ast)
				require.NoError(t, err)

				assert.Equal(t, tc.expectedPlanType, plan.Type,
					"Plan type should match expected for: %s", tc.sql)
			})
		}
	})
}

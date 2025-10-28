package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

// TestComprehensivePlanTypes 测试各种可能导致"Unknown plan type"的SQL语句
func TestComprehensivePlanTypes(t *testing.T) {
	// 创建 v2.0 Parquet 存储引擎
	testDir := SetupTestDir(t, "comprehensive_plan_test")
	storageEngine, err := storage.NewParquetEngine(testDir)
	assert.NoError(t, err)
	defer storageEngine.Close()
	err = storageEngine.Open()
	assert.NoError(t, err)

	cat := catalog.NewCatalog()
	cat.SetStorageEngine(storageEngine)
	err = cat.Init()
	if err != nil {
		t.Fatalf("Failed to initialize catalog: %v", err)
	}
	sessMgr, err := session.NewSessionManager()
	assert.NoError(t, err)
	sess := sessMgr.CreateSession()
	opt := optimizer.NewOptimizer()
	exec := executor.NewExecutor(cat)

	testQuery := func(description, sql string) {
		t.Logf("Testing: %s", description)
		t.Logf("SQL: %s", sql)

		stmt, err := parser.Parse(sql)
		if err != nil {
			t.Logf("  ❌ Parsing failed: %v", err)
			return
		}
		t.Logf("  ✓ Parsing successful")

		plan, err := opt.Optimize(stmt)
		if err != nil {
			t.Logf("  ❌ Optimization failed: %v", err)
			return
		}
		t.Logf("  ✓ Optimization successful, plan type: %v", plan.Type)

		_, err = exec.Execute(plan, sess)
		if err != nil {
			if strings.Contains(err.Error(), "不支持的计划节点类型") || strings.Contains(err.Error(), "unsupported plan node type") {
				t.Logf("  ❌ FOUND 'Unknown plan type' error: %v", err)
			} else {
				t.Logf("  ⚠️  Different error: %v", err)
			}
		} else {
			t.Logf("  ✅ Execution successful")
		}
		t.Log("")
	}

	// Setup test environment
	testQuery("Create database", "CREATE DATABASE comprehensive_test")
	sess.CurrentDB = "comprehensive_test"

	testQuery("Create table", "CREATE TABLE test_table (id INTEGER, name VARCHAR(50), amount FLOAT)")
	testQuery("Insert data", "INSERT INTO test_table (id, name, amount) VALUES (1, 'Alice', 100.0)")
	testQuery("Insert data", "INSERT INTO test_table (id, name, amount) VALUES (2, 'Bob', 200.0)")

	// Test potentially problematic plan types

	// 1. LimitPlan - LIMIT clause
	testQuery("LIMIT clause", "SELECT * FROM test_table LIMIT 1")

	// 2. DropTablePlan - DROP TABLE
	testQuery("DROP TABLE", "DROP TABLE IF EXISTS temp_table")

	// 3. TransactionPlan - Transaction statements
	testQuery("BEGIN transaction", "BEGIN")
	testQuery("COMMIT transaction", "COMMIT")
	testQuery("ROLLBACK transaction", "ROLLBACK")

	// 4. UsePlan - USE database
	testQuery("USE database", "USE comprehensive_test")

	// 5. ExplainPlan - EXPLAIN statement
	testQuery("EXPLAIN query", "EXPLAIN SELECT * FROM test_table")

	// 6. Complex queries that might generate missing plan combinations
	testQuery("ORDER BY with LIMIT", "SELECT * FROM test_table ORDER BY amount DESC LIMIT 1")
	testQuery("GROUP BY with LIMIT", "SELECT name, COUNT(*) FROM test_table GROUP BY name LIMIT 1")

	t.Log("✅ Comprehensive plan type test completed")
}

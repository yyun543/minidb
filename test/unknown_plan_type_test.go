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

// TestUnknownPlanTypeIssue 测试导致"不支持的计划节点类型: Unknown"错误的查询
func TestUnknownPlanTypeIssue(t *testing.T) {
	storageEngine, err := storage.NewParquetEngine(SetupTestDir(t, "unknown_plan_test.wal"))
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

	execSQL := func(sql string) (string, error) {
		t.Logf("Executing SQL: %s", sql)

		stmt, err := parser.Parse(sql)
		if err != nil {
			return "", err
		}
		t.Logf("✓ Parsing successful")

		plan, err := opt.Optimize(stmt)
		if err != nil {
			return "", err
		}
		t.Logf("✓ Optimization successful, plan type: %v", plan.Type)

		_, err = exec.Execute(plan, sess)
		if err != nil {
			return "", err
		}
		t.Logf("✓ Execution successful")

		return "success", nil
	}

	// 设置测试环境
	_, err = execSQL("CREATE DATABASE test_unknown_plan")
	assert.NoError(t, err)
	sess.CurrentDB = "test_unknown_plan"

	_, err = execSQL("CREATE TABLE sales (id INTEGER, product VARCHAR(50), amount FLOAT, region VARCHAR(50))")
	assert.NoError(t, err)

	_, err = execSQL("INSERT INTO sales (id, product, amount, region) VALUES (1, 'Widget A', 100.0, 'North')")
	assert.NoError(t, err)

	_, err = execSQL("INSERT INTO sales (id, product, amount, region) VALUES (2, 'Widget B', 200.0, 'South')")
	assert.NoError(t, err)

	// 测试可能导致"Unknown"错误的复杂查询

	// 1. 测试带HAVING子句的GROUP BY查询
	t.Log("Testing GROUP BY with HAVING clause...")
	result1, err := execSQL("SELECT region, SUM(amount) FROM sales GROUP BY region HAVING SUM(amount) > 150")
	if err != nil {
		if strings.Contains(err.Error(), "不支持的计划节点类型: Unknown") || strings.Contains(err.Error(), "unsupported plan node type: Unknown") {
			t.Logf("✗ Found 'Unknown plan type' error in GROUP BY HAVING: %v", err)
		} else {
			t.Logf("Different error in GROUP BY HAVING: %v", err)
		}
	} else {
		t.Logf("✓ GROUP BY HAVING executed successfully: %s", result1)
	}

	// 2. 测试复合查询
	t.Log("Testing compound query...")
	result2, err := execSQL("SELECT * FROM sales WHERE id IN (SELECT id FROM sales WHERE amount > 150)")
	if err != nil {
		if strings.Contains(err.Error(), "不支持的计划节点类型: Unknown") || strings.Contains(err.Error(), "unsupported plan node type: Unknown") {
			t.Logf("✗ Found 'Unknown plan type' error in subquery: %v", err)
		} else {
			t.Logf("Different error in subquery: %v", err)
		}
	} else {
		t.Logf("✓ Subquery executed successfully: %s", result2)
	}

	// 3. 测试窗口函数（如果支持）
	t.Log("Testing window function...")
	result3, err := execSQL("SELECT id, amount, ROW_NUMBER() OVER (ORDER BY amount DESC) FROM sales")
	if err != nil {
		if strings.Contains(err.Error(), "不支持的计划节点类型: Unknown") || strings.Contains(err.Error(), "unsupported plan node type: Unknown") {
			t.Logf("✗ Found 'Unknown plan type' error in window function: %v", err)
		} else {
			t.Logf("Different error in window function: %v", err)
		}
	} else {
		t.Logf("✓ Window function executed successfully: %s", result3)
	}

	// 4. 测试CTE（公用表表达式）
	t.Log("Testing CTE...")
	result4, err := execSQL("WITH high_sales AS (SELECT * FROM sales WHERE amount > 150) SELECT * FROM high_sales")
	if err != nil {
		if strings.Contains(err.Error(), "不支持的计划节点类型: Unknown") || strings.Contains(err.Error(), "unsupported plan node type: Unknown") {
			t.Logf("✗ Found 'Unknown plan type' error in CTE: %v", err)
		} else {
			t.Logf("Different error in CTE: %v", err)
		}
	} else {
		t.Logf("✓ CTE executed successfully: %s", result4)
	}

	t.Log("✅ Unknown plan type test completed - check logs above for any 'Unknown plan type' errors")
}

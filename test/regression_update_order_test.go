package test

import (
	"testing"

	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

// setupRegressionTest 创建回归测试环境
func setupRegressionTest(t *testing.T, testName string) (*executor.ExecutorImpl, *session.Session, func()) {
	storageEngine, err := storage.NewParquetEngine(SetupTestDir(t, testName))
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
	exec := executor.NewExecutor(cat)

	cleanup := func() {
		storageEngine.Close()
	}

	return exec, sess, cleanup
}

// execSQL 执行SQL语句的辅助函数
func execSQL(t *testing.T, exec *executor.ExecutorImpl, sess *session.Session, sql string) (*executor.ResultSet, error) {
	t.Helper()
	stmt, err := parser.Parse(sql)
	if err != nil {
		return nil, err
	}
	opt := optimizer.NewOptimizer()
	plan, err := opt.Optimize(stmt)
	if err != nil {
		return nil, err
	}
	return exec.Execute(plan, sess)
}

// TestOrderByWithBooleanType 回归测试: ORDER BY 支持 BOOLEAN 类型
// 问题: ORDER BY 在处理 BOOLEAN 列时会导致服务器崩溃
// 修复: internal/executor/operators/order_by.go 添加 Boolean 类型支持
// 日期: 2025-10-28
func TestOrderByWithBooleanType(t *testing.T) {
	exec, sess, cleanup := setupRegressionTest(t, "order_by_boolean_regression")
	defer cleanup()

	sess.CurrentDB = "default"

	_, err := execSQL(t, exec, sess, `
		CREATE TABLE test_order_bool (
			id INTEGER,
			name VARCHAR(50),
			is_active BOOLEAN,
			score DOUBLE
		)
	`)
	require.NoError(t, err, "Failed to create table")

	// 插入测试数据
	insertData := []string{
		"INSERT INTO test_order_bool VALUES (1, 'Alice', true, 95.5)",
		"INSERT INTO test_order_bool VALUES (2, 'Bob', false, 87.3)",
		"INSERT INTO test_order_bool VALUES (3, 'Charlie', true, 92.1)",
	}
	for _, sql := range insertData {
		_, err := execSQL(t, exec, sess, sql)
		require.NoError(t, err, "Failed to insert data")
	}

	// 测试 ORDER BY BOOLEAN 列 - 关键测试，之前会崩溃
	resultSet, err := execSQL(t, exec, sess, "SELECT * FROM test_order_bool ORDER BY is_active")
	require.NoError(t, err, "ORDER BY with BOOLEAN should not crash - THIS WAS THE BUG")
	require.NotNil(t, resultSet, "Result should not be nil")
	assert.True(t, len(resultSet.Batches()) > 0, "Should have result batches")
	t.Log("✅ REGRESSION TEST PASSED: ORDER BY BOOLEAN works without crashing")

	// 测试 ORDER BY BOOLEAN DESC
	resultSet, err = execSQL(t, exec, sess, "SELECT * FROM test_order_bool ORDER BY is_active DESC")
	require.NoError(t, err, "ORDER BY BOOLEAN DESC should work")
	assert.True(t, len(resultSet.Batches()) > 0, "Should have result batches")
	t.Log("✅ ORDER BY BOOLEAN DESC also works")
}

// TestUpdateWithArithmeticExpressions 回归测试: UPDATE 支持算术表达式
// 问题: UPDATE SET price = price * 1.1 不生效，值未变化
// 根本原因: ANTLR 语法缺少算术表达式规则
// 修复:
//  1. internal/parser/MiniQL.g4 添加 multiplicativeExpression 和 additiveExpression
//  2. internal/parser/parser.go 添加访问者方法
//  3. internal/executor/executor.go 实现表达式求值器
//
// 日期: 2025-10-28
func TestUpdateWithArithmeticExpressions(t *testing.T) {
	exec, sess, cleanup := setupRegressionTest(t, "update_expr_regression")
	defer cleanup()

	sess.CurrentDB = "default"

	_, err := execSQL(t, exec, sess, `
		CREATE TABLE products (
			id INTEGER,
			name VARCHAR(50),
			price DOUBLE
		)
	`)
	require.NoError(t, err, "Failed to create table")

	_, err = execSQL(t, exec, sess, "INSERT INTO products VALUES (1, 'Laptop', 100.0)")
	require.NoError(t, err, "Failed to insert data")

	// 测试: 乘法表达式 - 这是原始bug
	_, err = execSQL(t, exec, sess, "UPDATE products SET price = price * 2 WHERE id = 1")
	require.NoError(t, err, "UPDATE with multiplication should work - THIS WAS THE BUG")

	resultSet, err := execSQL(t, exec, sess, "SELECT price FROM products WHERE id = 1")
	require.NoError(t, err)
	require.True(t, len(resultSet.Batches()) > 0 && resultSet.Batches()[0].NumRows() > 0, "Should have result")

	// 从Arrow列中提取值
	batch := resultSet.Batches()[0]
	priceCol := batch.Column(0).(*array.Float64)
	price := priceCol.Value(0)

	assert.InDelta(t, 200.0, price, 0.01, "Price should be doubled: 100 * 2 = 200")
	t.Log("✅ REGRESSION TEST PASSED: UPDATE with multiplication expression works")

	// 测试: 百分比增加 (常见业务场景) - 用户报告的原始问题
	_, err = execSQL(t, exec, sess, "INSERT INTO products VALUES (2, 'Mouse', 100.0)")
	require.NoError(t, err)

	_, err = execSQL(t, exec, sess, "UPDATE products SET price = price * 1.1 WHERE id = 2")
	require.NoError(t, err, "UPDATE with percentage increase should work")

	resultSet, err = execSQL(t, exec, sess, "SELECT price FROM products WHERE id = 2")
	require.NoError(t, err)
	require.True(t, len(resultSet.Batches()) > 0 && resultSet.Batches()[0].NumRows() > 0)

	batch = resultSet.Batches()[0]
	priceCol = batch.Column(0).(*array.Float64)
	price = priceCol.Value(0)

	assert.InDelta(t, 110.0, price, 0.01, "Price should be 110%: 100 * 1.1 = 110")
	t.Log("✅ UPDATE with percentage (price * 1.1) - ORIGINAL USER-REPORTED BUG FIXED")
}

// TestUpdateWithoutWhereClause 回归测试: UPDATE 不带 WHERE 子句应更新所有行
// 问题: UPDATE table SET col = val (无WHERE) 只更新部分行或不更新
// 根本原因: merge_on_read.go 的 applyUpdateDelta 在 filterColumn == "" 时返回
// 修复: internal/storage/merge_on_read.go 添加无过滤条件时更新所有行的逻辑
// 日期: 2025-10-28
func TestUpdateWithoutWhereClause(t *testing.T) {
	exec, sess, cleanup := setupRegressionTest(t, "update_no_where_regression")
	defer cleanup()

	sess.CurrentDB = "default"

	_, err := execSQL(t, exec, sess, `
		CREATE TABLE items (
			id INTEGER,
			status VARCHAR(20)
		)
	`)
	require.NoError(t, err, "Failed to create table")

	// 插入测试数据
	insertData := []string{
		"INSERT INTO items VALUES (1, 'pending')",
		"INSERT INTO items VALUES (2, 'active')",
		"INSERT INTO items VALUES (3, 'pending')",
	}
	for _, sql := range insertData {
		_, err := execSQL(t, exec, sess, sql)
		require.NoError(t, err, "Failed to insert")
	}

	// 测试: UPDATE 不带 WHERE - 原始bug，应该更新所有行
	_, err = execSQL(t, exec, sess, "UPDATE items SET status = 'archived'")
	require.NoError(t, err, "UPDATE without WHERE should work - THIS WAS THE BUG")
	t.Log("✅ UPDATE without WHERE executed successfully")

	// 验证所有行都被更新
	resultSet, err := execSQL(t, exec, sess, "SELECT COUNT(*) FROM items WHERE status = 'archived'")
	require.NoError(t, err)
	require.True(t, len(resultSet.Batches()) > 0 && resultSet.Batches()[0].NumRows() > 0)

	batch := resultSet.Batches()[0]
	countCol := batch.Column(0).(*array.Int64)
	archivedCount := countCol.Value(0)
	assert.Equal(t, int64(3), archivedCount, "All 3 rows should be archived - ORIGINAL BUG FIXED")
	t.Log("✅ REGRESSION TEST PASSED: UPDATE without WHERE correctly updated all rows")
}

// TestMultiColumnUpdateRegression 回归测试: 多列UPDATE应该正常工作
// 问题: 多列同时更新可能执行异常
// 验证: 确保多列更新能够正确执行
// 日期: 2025-10-28
func TestMultiColumnUpdateRegression(t *testing.T) {
	exec, sess, cleanup := setupRegressionTest(t, "multi_column_update_regression")
	defer cleanup()

	sess.CurrentDB = "default"

	_, err := execSQL(t, exec, sess, `
		CREATE TABLE products (
			id INTEGER,
			price DOUBLE,
			quantity INTEGER,
			status VARCHAR(20)
		)
	`)
	require.NoError(t, err, "Failed to create table")

	_, err = execSQL(t, exec, sess, "INSERT INTO products VALUES (1, 100.0, 10, 'active')")
	require.NoError(t, err)

	// 测试: 多列同时更新
	_, err = execSQL(t, exec, sess, "UPDATE products SET price = 150.0, quantity = 20, status = 'updated' WHERE id = 1")
	require.NoError(t, err, "Multi-column UPDATE should work")

	// 验证所有列都被更新
	resultSet, err := execSQL(t, exec, sess, "SELECT price, quantity, status FROM products WHERE id = 1")
	require.NoError(t, err)
	require.True(t, len(resultSet.Batches()) > 0 && resultSet.Batches()[0].NumRows() > 0)

	batch := resultSet.Batches()[0]
	price := batch.Column(0).(*array.Float64).Value(0)
	quantity := batch.Column(1).(*array.Int64).Value(0)
	status := batch.Column(2).(*array.String).Value(0)

	assert.InDelta(t, 150.0, price, 0.01, "Price should be updated")
	assert.Equal(t, int64(20), quantity, "Quantity should be updated")
	assert.Equal(t, "updated", status, "Status should be updated")
	t.Log("✅ Multi-column UPDATE works correctly")
}

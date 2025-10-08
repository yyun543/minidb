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

func TestExecutor(t *testing.T) {
	// 创建 v2.0 Parquet 存储引擎
	storageEngine, err := storage.NewParquetEngine("./test_data/executor_test")
	assert.NoError(t, err)
	defer storageEngine.Close()
	err = storageEngine.Open()
	assert.NoError(t, err)

	// 创建目录和会话
	cat := catalog.NewCatalog()
	cat.SetStorageEngine(storageEngine)
	err = cat.Init()
	if err != nil {
		t.Fatalf("Failed to initialize catalog: %v", err)
	}

	sessMgr, err := session.NewSessionManager()
	assert.NoError(t, err)
	sess := sessMgr.CreateSession()

	// 创建优化器和执行器
	opt := optimizer.NewOptimizer()
	exec := executor.NewExecutor(cat)

	// 创建默认数据库
	createDbSQL := "CREATE DATABASE default"
	stmt, err := parser.Parse(createDbSQL)
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// 设置当前数据库上下文
	sess.CurrentDB = "default"

	// 测试执行器创建
	t.Run("TestNewExecutor", func(t *testing.T) {
		assert.NotNil(t, exec)
	})

	// 测试SELECT语句执行
	t.Run("TestExecuteSelect", func(t *testing.T) {
		// 基本SELECT
		t.Run("BasicSelect", func(t *testing.T) {
			// 先创建表并插入数据
			createSQL := "CREATE TABLE users (id INTEGER, name VARCHAR(255))"
			stmt, err := parser.Parse(createSQL)
			assert.NoError(t, err)
			plan, err := opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)

			// 插入测试数据
			insertSQL := "INSERT INTO users (id, name) VALUES (1, 'test')"
			stmt, err = parser.Parse(insertSQL)
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)

			// 执行SELECT
			sql := "SELECT id, name FROM users"
			stmt, err = parser.Parse(sql)
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			result, err := exec.Execute(plan, sess)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			// 验证结果集
			assert.Equal(t, []string{"id", "name"}, result.Headers)
			assert.Equal(t, 1, len(result.Batches()))
		})

		// 带WHERE条件的SELECT
		t.Run("SelectWithWhere", func(t *testing.T) {
			sql := "SELECT id, name FROM users WHERE id = 1"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)
			plan, err := opt.Optimize(stmt)
			assert.NoError(t, err)
			result, err := exec.Execute(plan, sess)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			// 验证结果集
			assert.Equal(t, []string{"id", "name"}, result.Headers)
			assert.Equal(t, 1, len(result.Batches()))
		})

		// 带JOIN的SELECT
		t.Run("SelectWithJoin", func(t *testing.T) {
			// 创建orders表
			createSQL := "CREATE TABLE orders (order_id INTEGER, user_id INTEGER)"
			stmt, err := parser.Parse(createSQL)
			assert.NoError(t, err)
			plan, err := opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)

			// 插入测试数据
			insertSQL := "INSERT INTO orders (order_id, user_id) VALUES (1, 1)"
			stmt, err = parser.Parse(insertSQL)
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)

			// 执行JOIN查询
			sql := "SELECT u.id, u.name, o.order_id FROM users u JOIN orders o ON u.id = o.user_id"
			stmt, err = parser.Parse(sql)
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			result, err := exec.Execute(plan, sess)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			// 验证结果集
			assert.Equal(t, []string{"u.id", "u.name", "o.order_id"}, result.Headers)
			assert.Equal(t, 1, len(result.Batches()))
		})
	})

	// 测试INSERT语句执行
	t.Run("TestExecuteInsert", func(t *testing.T) {
		// 确保表存在（如果不存在则创建）
		createSQL := "CREATE TABLE users (id INTEGER, name VARCHAR(255))"
		stmt, err := parser.Parse(createSQL)
		if err == nil {
			plan, err := opt.Optimize(stmt)
			if err == nil {
				exec.Execute(plan, sess) // 忽略错误，因为表可能已存在
			}
		}

		sql := "INSERT INTO users (id, name) VALUES (2, 'test2')"
		stmt, err = parser.Parse(sql)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 验证插入是否成功
		sql = "SELECT id, name FROM users WHERE id = 2"
		stmt, err = parser.Parse(sql)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result.Batches()))
	})

	// 测试UPDATE语句执行
	t.Run("TestExecuteUpdate", func(t *testing.T) {
		// 确保表存在（如果不存在则创建）
		createSQL := "CREATE TABLE users (id INTEGER, name VARCHAR(255))"
		stmt, err := parser.Parse(createSQL)
		if err == nil {
			plan, err := opt.Optimize(stmt)
			if err == nil {
				exec.Execute(plan, sess) // 忽略错误，因为表可能已存在
			}
		}

		// 确保有数据可更新：先插入id=2的记录
		insertSQL := "INSERT INTO users (id, name) VALUES (2, 'before_update')"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// 执行UPDATE操作
		sql := "UPDATE users SET name = 'updated' WHERE id = 2"
		stmt, err = parser.Parse(sql)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 验证更新是否成功
		sql = "SELECT name FROM users WHERE id = 2"
		stmt, err = parser.Parse(sql)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
		if len(result.Batches()) > 0 && result.Batches()[0].NumRows() > 0 {
			actualValue := result.Batches()[0].GetString(0, 0)
			t.Logf("Expected: 'updated', Got: '%s'", actualValue)
			// UPDATE fix has been implemented - test should now pass
			assert.Equal(t, "updated", actualValue)
		} else {
			t.Logf("No data returned from UPDATE verification query")
		}
	})

	// 测试DELETE语句执行
	t.Run("TestExecuteDelete", func(t *testing.T) {
		sql := "DELETE FROM users WHERE id = 2"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 验证删除是否成功
		sql = "SELECT id FROM users WHERE id = 2"
		stmt, err = parser.Parse(sql)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(result.Batches()))
	})

	// 测试GROUP BY功能
	t.Run("TestGroupByFunctionality", func(t *testing.T) {
		// 创建测试表
		createSQL := "CREATE TABLE sales (region VARCHAR, amount INT)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// 插入测试数据
		insertData := []string{
			"INSERT INTO sales VALUES ('North', 100)",
			"INSERT INTO sales VALUES ('South', 150)",
			"INSERT INTO sales VALUES ('North', 200)",
			"INSERT INTO sales VALUES ('East', 300)",
			"INSERT INTO sales VALUES ('North', 50)",
		}

		for _, insertSQL := range insertData {
			stmt, err := parser.Parse(insertSQL)
			assert.NoError(t, err)
			plan, err := opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)
		}

		// 测试基本GROUP BY with 别名
		t.Run("BasicGroupByWithAliases", func(t *testing.T) {
			sql := "SELECT region, COUNT(*) AS orders, SUM(amount) AS total FROM sales GROUP BY region"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)
			plan, err := opt.Optimize(stmt)
			assert.NoError(t, err)
			result, err := exec.Execute(plan, sess)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			// 验证表头别名显示正确
			expectedHeaders := []string{"region", "orders", "total"}
			assert.Equal(t, expectedHeaders, result.Headers)
			assert.Greater(t, len(result.Batches()), 0)
		})

		// 测试AVG聚合函数
		t.Run("AvgAggregationFunction", func(t *testing.T) {
			sql := "SELECT region, AVG(amount) AS avg_amount FROM sales GROUP BY region"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)
			plan, err := opt.Optimize(stmt)
			assert.NoError(t, err)
			result, err := exec.Execute(plan, sess)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			expectedHeaders := []string{"region", "avg_amount"}
			assert.Equal(t, expectedHeaders, result.Headers)
		})

		// 测试HAVING子句
		t.Run("HavingClause", func(t *testing.T) {
			sql := "SELECT region, COUNT(*) AS cnt FROM sales GROUP BY region HAVING cnt >= 2"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)
			plan, err := opt.Optimize(stmt)
			assert.NoError(t, err)
			result, err := exec.Execute(plan, sess)
			assert.NoError(t, err)
			assert.NotNil(t, result)

			expectedHeaders := []string{"region", "cnt"}
			assert.Equal(t, expectedHeaders, result.Headers)
		})
	})
}

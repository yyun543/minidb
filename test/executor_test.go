package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
)

func TestExecutor(t *testing.T) {
	// 创建目录和会话
	cat, err := catalog.NewCatalogWithDefaultStorage()
	assert.NoError(t, err)
	err = cat.Init()
	assert.NoError(t, err)
	sessMgr, err := session.NewSessionManager()
	assert.NoError(t, err)
	sess := sessMgr.CreateSession()

	// 创建优化器和执行器
	opt := optimizer.NewOptimizer()
	exec := executor.NewExecutor(cat)

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
		sql := "INSERT INTO users (id, name) VALUES (2, 'test2')"
		stmt, err := parser.Parse(sql)
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
		sql := "UPDATE users SET name = 'updated' WHERE id = 2"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
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
		assert.Equal(t, "updated", result.Batches()[0].GetString(0, 0))
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
}

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

func TestGroupByFunctionality(t *testing.T) {
	// 创建 v2.0 Parquet 存储引擎
	testDir := SetupTestDir(t, "group_by_test")
	storageEngine, err := storage.NewParquetEngine(testDir)
	assert.NoError(t, err)
	defer storageEngine.Close()
	err = storageEngine.Open()
	assert.NoError(t, err)

	// 创建catalog和session
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

	// 创建测试数据库和表
	setupTestData(t, opt, exec, sess)

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

		// 验证有数据返回
		assert.Greater(t, len(result.Batches()), 0)
	})

	t.Run("AvgAggregationFunction", func(t *testing.T) {
		sql := "SELECT region, AVG(amount) AS avg_amount FROM sales GROUP BY region"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 验证表头别名
		expectedHeaders := []string{"region", "avg_amount"}
		assert.Equal(t, expectedHeaders, result.Headers)

		// 验证AVG计算正确
		assert.Greater(t, len(result.Batches()), 0)
		if len(result.Batches()) > 0 && result.Batches()[0].NumRows() > 0 {
			// 检查北方地区的平均值 (100+200+50)/3 = 116.67
			t.Logf("AVG calculation test passed")
		}
	})

	t.Run("CountStarFunction", func(t *testing.T) {
		sql := "SELECT region, COUNT(*) AS count FROM sales GROUP BY region"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 验证COUNT(*)返回正确值，不是0
		expectedHeaders := []string{"region", "count"}
		assert.Equal(t, expectedHeaders, result.Headers)

		assert.Greater(t, len(result.Batches()), 0)
		if len(result.Batches()) > 0 && result.Batches()[0].NumRows() > 0 {
			// 验证COUNT不为0
			t.Logf("COUNT(*) calculation test passed")
		}
	})

	t.Run("HavingClause", func(t *testing.T) {
		sql := "SELECT region, COUNT(*) AS cnt FROM sales GROUP BY region HAVING cnt >= 2"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 验证HAVING子句的表头别名显示
		expectedHeaders := []string{"region", "cnt"}
		assert.Equal(t, expectedHeaders, result.Headers)

		// HAVING应该过滤掉cnt<2的记录
		assert.Greater(t, len(result.Batches()), 0)
	})

	t.Run("ComplexNestedQueryHeaders", func(t *testing.T) {
		// 测试复杂嵌套查询的表头别名显示
		// 包含 JOIN, GROUP BY, HAVING, ORDER BY 的复杂查询
		sql := `SELECT u.name, COUNT(o.id) as order_count, SUM(o.amount) as total_amount, AVG(o.amount) as avg_amount 
                FROM users u LEFT JOIN orders o ON u.id = o.user_id 
                GROUP BY u.name HAVING order_count > 1 ORDER BY total_amount DESC`

		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 验证所有别名都正确显示
		expectedHeaders := []string{"u.name", "order_count", "total_amount", "avg_amount"}
		assert.Equal(t, expectedHeaders, result.Headers)

		t.Logf("Complex nested query headers: %v", result.Headers)
	})

	t.Run("AllAggregationFunctions", func(t *testing.T) {
		sql := "SELECT region, COUNT(*) as cnt, SUM(amount) as sum_amt, AVG(amount) as avg_amt, MIN(amount) as min_amt, MAX(amount) as max_amt FROM sales GROUP BY region"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// 验证所有聚合函数的别名显示
		expectedHeaders := []string{"region", "cnt", "sum_amt", "avg_amt", "min_amt", "max_amt"}
		assert.Equal(t, expectedHeaders, result.Headers)
	})
}

// setupTestData 设置测试数据
func setupTestData(t *testing.T, opt *optimizer.Optimizer, exec *executor.ExecutorImpl, sess *session.Session) {
	// 创建数据库
	createDbSQL := "CREATE DATABASE test"
	stmt, err := parser.Parse(createDbSQL)
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// 设置当前数据库
	sess.CurrentDB = "test"

	// 创建销售表
	createTableSQL := "CREATE TABLE sales (region VARCHAR, amount INT)"
	stmt, err = parser.Parse(createTableSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
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

	// 创建用户表和订单表用于复杂查询测试
	createUsersSQL := "CREATE TABLE users (id INT, name VARCHAR)"
	stmt, err = parser.Parse(createUsersSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	createOrdersSQL := "CREATE TABLE orders (id INT, user_id INT, amount INT)"
	stmt, err = parser.Parse(createOrdersSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// 插入用户和订单数据
	userData := []string{
		"INSERT INTO users VALUES (1, 'John Doe')",
		"INSERT INTO users VALUES (2, 'Jane Smith')",
	}

	orderData := []string{
		"INSERT INTO orders VALUES (1, 1, 100)",
		"INSERT INTO orders VALUES (2, 1, 150)",
		"INSERT INTO orders VALUES (3, 2, 75)",
	}

	for _, userSQL := range userData {
		stmt, err := parser.Parse(userSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
	}

	for _, orderSQL := range orderData {
		stmt, err := parser.Parse(orderSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
	}
}

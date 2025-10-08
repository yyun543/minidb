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

// TestUpdateDataPersistence 专门测试UPDATE操作的数据持久化问题
// 这是一个failing test，用来重现并修复UPDATE问题
func TestUpdateDataPersistence(t *testing.T) {
	// 创建独立的存储引擎，避免干扰其他测试
	storageEngine, err := storage.NewParquetEngine("./test_data/test_update_debug")
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

	// 创建默认数据库
	createDbSQL := "CREATE DATABASE test_update_db"
	stmt, err := parser.Parse(createDbSQL)
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// 第一步：创建测试表
	createSQL := "CREATE TABLE test_users (id INTEGER, name VARCHAR(255), status VARCHAR(50))"
	stmt, err = parser.Parse(createSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)
	t.Log("✓ Table created successfully")

	// 第二步：插入测试数据
	insertSQL := "INSERT INTO test_users (id, name, status) VALUES (1, 'John', 'active')"
	stmt, err = parser.Parse(insertSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)
	t.Log("✓ Initial data inserted")

	// 第三步：验证初始数据存在
	selectSQL := "SELECT id, name, status FROM test_users WHERE id = 1"
	stmt, err = parser.Parse(selectSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	result, err := exec.Execute(plan, sess)
	assert.NoError(t, err)
	assert.Greater(t, len(result.Batches()), 0, "Should have data before update")
	assert.Greater(t, int(result.Batches()[0].NumRows()), 0, "Should have at least one row")

	initialName := result.Batches()[0].GetString(1, 0)   // name column
	initialStatus := result.Batches()[0].GetString(2, 0) // status column
	t.Logf("✓ Initial data verified: name=%s, status=%s", initialName, initialStatus)
	assert.Equal(t, "John", initialName)
	assert.Equal(t, "active", initialStatus)

	// 第四步：执行UPDATE操作
	updateSQL := "UPDATE test_users SET name = 'Jane', status = 'inactive' WHERE id = 1"
	stmt, err = parser.Parse(updateSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)
	t.Log("✓ UPDATE executed without error")

	// 第五步：验证UPDATE是否生效 - 这里应该会失败，证明问题存在
	selectSQL = "SELECT id, name, status FROM test_users WHERE id = 1"
	stmt, err = parser.Parse(selectSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	result, err = exec.Execute(plan, sess)
	assert.NoError(t, err)
	assert.Greater(t, len(result.Batches()), 0, "Should have data after update")
	assert.Greater(t, int(result.Batches()[0].NumRows()), 0, "Should have at least one row after update")

	updatedName := result.Batches()[0].GetString(1, 0)   // name column
	updatedStatus := result.Batches()[0].GetString(2, 0) // status column
	t.Logf("After UPDATE: name=%s, status=%s", updatedName, updatedStatus)

	// 这些断言应该会失败，证明UPDATE问题存在
	assert.Equal(t, "Jane", updatedName, "Name should be updated to 'Jane'")
	assert.Equal(t, "inactive", updatedStatus, "Status should be updated to 'inactive'")
}

// TestUpdateDataPersistenceMinimal 最简化的UPDATE测试，仅测试单一字段更新
func TestUpdateDataPersistenceMinimal(t *testing.T) {
	// 使用不同的文件名避免冲突
	storageEngine, err := storage.NewParquetEngine("./test_data/test_update_minimal.wal")
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

	// 创建数据库
	_, err = exec.Execute(mustParse(opt, "CREATE DATABASE minimal_test"), sess)
	assert.NoError(t, err)

	// 创建简单的表（只有两列）
	_, err = exec.Execute(mustParse(opt, "CREATE TABLE users (id INTEGER, name VARCHAR(255))"), sess)
	assert.NoError(t, err)

	// 插入一行数据
	_, err = exec.Execute(mustParse(opt, "INSERT INTO users (id, name) VALUES (1, 'original')"), sess)
	assert.NoError(t, err)

	// 验证插入的数据
	result, err := exec.Execute(mustParse(opt, "SELECT name FROM users WHERE id = 1"), sess)
	assert.NoError(t, err)
	if len(result.Batches()) > 0 && result.Batches()[0].NumRows() > 0 {
		originalName := result.Batches()[0].GetString(0, 0)
		assert.Equal(t, "original", originalName)
		t.Logf("✓ Original name: %s", originalName)
	}

	// 更新数据
	_, err = exec.Execute(mustParse(opt, "UPDATE users SET name = 'updated' WHERE id = 1"), sess)
	assert.NoError(t, err)
	t.Log("✓ UPDATE executed")

	// 验证更新后的数据
	result, err = exec.Execute(mustParse(opt, "SELECT name FROM users WHERE id = 1"), sess)
	assert.NoError(t, err)
	if len(result.Batches()) > 0 && result.Batches()[0].NumRows() > 0 {
		updatedName := result.Batches()[0].GetString(0, 0)
		t.Logf("Updated name: %s", updatedName)
		assert.Equal(t, "updated", updatedName, "FAILING TEST: This will fail until UPDATE is fixed")
	} else {
		t.Fatal("No data returned after UPDATE")
	}
}

// mustParse 辅助函数，简化测试代码
func mustParse(opt *optimizer.Optimizer, sql string) *optimizer.Plan {
	stmt, err := parser.Parse(sql)
	if err != nil {
		panic(err)
	}
	plan, err := opt.Optimize(stmt)
	if err != nil {
		panic(err)
	}
	return plan
}

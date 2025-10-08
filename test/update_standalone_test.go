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

// TestUpdateStandalone 独立的UPDATE测试，验证UPDATE修复是否有效
func TestUpdateStandalone(t *testing.T) {
	// 创建完全独立的测试环境
	storageEngine, err := storage.NewParquetEngine("./test_data/update_standalone.wal")
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

	// 执行SQL辅助函数
	execSQL := func(sql string) (*executor.ResultSet, error) {
		stmt, err := parser.Parse(sql)
		if err != nil {
			return nil, err
		}
		plan, err := opt.Optimize(stmt)
		if err != nil {
			return nil, err
		}
		return exec.Execute(plan, sess)
	}

	// 1. 创建数据库
	_, err = execSQL("CREATE DATABASE update_test")
	assert.NoError(t, err)
	t.Log("✓ Database created")

	// 手动设置当前数据库（因为USE语句未实现）
	sess.CurrentDB = "update_test"
	t.Log("✓ Current database set to update_test")

	// 2. 创建表
	_, err = execSQL("CREATE TABLE test_table (id INTEGER, name VARCHAR(255))")
	assert.NoError(t, err)
	t.Log("✓ Table created")

	// 3. 插入初始数据
	_, err = execSQL("INSERT INTO test_table (id, name) VALUES (1, 'original')")
	assert.NoError(t, err)
	t.Log("✓ Initial data inserted")

	// 4. 验证初始数据
	result, err := execSQL("SELECT id, name FROM test_table WHERE id = 1")
	assert.NoError(t, err)
	assert.True(t, len(result.Batches()) > 0, "Should have initial data")
	if len(result.Batches()) > 0 && result.Batches()[0].NumRows() > 0 {
		initialName := result.Batches()[0].GetString(1, 0) // name column (index 1)
		assert.Equal(t, "original", initialName, "Initial name should be 'original'")
		t.Logf("✓ Initial data verified: name = %s", initialName)
	}

	// 5. 执行UPDATE
	_, err = execSQL("UPDATE test_table SET name = 'updated' WHERE id = 1")
	assert.NoError(t, err)
	t.Log("✓ UPDATE executed")

	// 6. 验证UPDATE结果 - 这是关键测试
	result, err = execSQL("SELECT id, name FROM test_table WHERE id = 1")
	assert.NoError(t, err)
	assert.True(t, len(result.Batches()) > 0, "Should have data after update")
	if len(result.Batches()) > 0 && result.Batches()[0].NumRows() > 0 {
		updatedName := result.Batches()[0].GetString(1, 0) // name column (index 1)
		t.Logf("After UPDATE: name = %s", updatedName)
		assert.Equal(t, "updated", updatedName, "UPDATE should have changed name to 'updated'")
		t.Log("✅ UPDATE test PASSED - fix is working!")
	} else {
		t.Fatal("❌ No data returned after UPDATE - UPDATE fix failed")
	}

	// 7. 验证未匹配行没有被更新
	_, err = execSQL("INSERT INTO test_table (id, name) VALUES (2, 'unchanged')")
	assert.NoError(t, err)

	result, err = execSQL("SELECT name FROM test_table WHERE id = 2")
	assert.NoError(t, err)
	if len(result.Batches()) > 0 && result.Batches()[0].NumRows() > 0 {
		unchangedName := result.Batches()[0].GetString(0, 0)
		assert.Equal(t, "unchanged", unchangedName, "Non-matching rows should not be affected")
		t.Log("✓ Non-matching rows correctly preserved")
	}
}

// TestUpdateMultipleColumns 测试多列UPDATE
func TestUpdateMultipleColumns(t *testing.T) {
	storageEngine, err := storage.NewParquetEngine("./test_data/update_multi.wal")
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

	execSQL := func(sql string) (*executor.ResultSet, error) {
		stmt, err := parser.Parse(sql)
		if err != nil {
			return nil, err
		}
		plan, err := opt.Optimize(stmt)
		if err != nil {
			return nil, err
		}
		return exec.Execute(plan, sess)
	}

	// 准备测试环境
	_, err = execSQL("CREATE DATABASE multi_test")
	assert.NoError(t, err)
	sess.CurrentDB = "multi_test"

	_, err = execSQL("CREATE TABLE users (id INTEGER, name VARCHAR(255), status VARCHAR(50))")
	assert.NoError(t, err)

	_, err = execSQL("INSERT INTO users (id, name, status) VALUES (1, 'John', 'active')")
	assert.NoError(t, err)

	// 执行多列UPDATE - 这个测试在原来的executor_test.go中失败了
	_, err = execSQL("UPDATE users SET name = 'Jane', status = 'inactive' WHERE id = 1")
	assert.NoError(t, err)
	t.Log("✓ Multi-column UPDATE executed")

	// 验证多列UPDATE结果
	result, err := execSQL("SELECT id, name, status FROM users WHERE id = 1")
	assert.NoError(t, err)
	assert.True(t, len(result.Batches()) > 0, "Should have data after multi-column update")

	if len(result.Batches()) > 0 && result.Batches()[0].NumRows() > 0 {
		batch := result.Batches()[0]
		updatedName := batch.GetString(1, 0)   // name column
		updatedStatus := batch.GetString(2, 0) // status column

		t.Logf("After multi-column UPDATE: name = %s, status = %s", updatedName, updatedStatus)
		assert.Equal(t, "Jane", updatedName, "Name should be updated to 'Jane'")
		assert.Equal(t, "inactive", updatedStatus, "Status should be updated to 'inactive'")
		t.Log("✅ Multi-column UPDATE test PASSED!")
	} else {
		t.Fatal("❌ Multi-column UPDATE test FAILED")
	}
}

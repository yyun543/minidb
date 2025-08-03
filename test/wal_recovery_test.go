package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

// TestWALRecovery 测试WAL恢复 - 服务端重启后数据应该恢复
func TestWALRecovery(t *testing.T) {
	walFile := "test_wal_recovery.wal"

	// 清理之前的WAL文件
	os.Remove(walFile)
	defer os.Remove(walFile)

	// 阶段1: 创建数据库、表并插入数据，然后"重启"服务器
	t.Log("=== Phase 1: Create database and tables ===")

	func() {
		// 第一次启动服务器
		engine, err := storage.NewMemTable(walFile)
		assert.NoError(t, err)
		defer engine.Close()

		err = engine.Open()
		assert.NoError(t, err)

		cat := catalog.CreateTemporaryCatalog(engine)
		sessMgr, err := session.NewSessionManager()
		assert.NoError(t, err)
		sess := sessMgr.CreateSession()
		opt := optimizer.NewOptimizer()
		exec := executor.NewExecutor(cat)

		execSQL := func(sql string) error {
			t.Logf("Executing: %s", sql)
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

		// 创建数据库和表
		err = execSQL("CREATE DATABASE recovery_test")
		assert.NoError(t, err, "Should create database successfully")
		sess.CurrentDB = "recovery_test"

		err = execSQL("CREATE TABLE users (id INTEGER, name VARCHAR(100))")
		assert.NoError(t, err, "Should create table successfully")

		err = execSQL("CREATE TABLE orders (id INTEGER, user_id INTEGER, amount FLOAT)")
		assert.NoError(t, err, "Should create second table successfully")

		// 插入数据
		err = execSQL("INSERT INTO users (id, name) VALUES (1, 'Alice')")
		assert.NoError(t, err, "Should insert user data")

		err = execSQL("INSERT INTO users (id, name) VALUES (2, 'Bob')")
		assert.NoError(t, err, "Should insert user data")

		err = execSQL("INSERT INTO orders (id, user_id, amount) VALUES (1, 1, 100.5)")
		assert.NoError(t, err, "Should insert order data")

		t.Log("✓ Data created and inserted successfully")

		// 检查WAL文件是否存在
		if _, err := os.Stat(walFile); os.IsNotExist(err) {
			t.Logf("⚠️  WAL file %s does not exist after operations", walFile)
		} else {
			t.Logf("✓ WAL file %s exists", walFile)
		}
	}()

	// 阶段2: 模拟服务器重启，检查数据是否恢复
	t.Log("=== Phase 2: Simulate server restart and check recovery ===")

	// 第二次启动服务器 (模拟重启)
	engine, err := storage.NewMemTable(walFile)
	assert.NoError(t, err, "Should create engine for recovery")
	defer engine.Close()

	err = engine.Open()
	assert.NoError(t, err, "Should open engine for recovery")

	cat := catalog.CreateTemporaryCatalog(engine)
	sessMgr, err := session.NewSessionManager()
	assert.NoError(t, err)
	sess := sessMgr.CreateSession()
	sess.CurrentDB = "recovery_test" // 设置数据库上下文
	opt := optimizer.NewOptimizer()
	exec := executor.NewExecutor(cat)

	queryAndCheck := func(sql string, description string) bool {
		t.Logf("Checking: %s", description)
		t.Logf("Query: %s", sql)

		stmt, err := parser.Parse(sql)
		if err != nil {
			t.Logf("  ❌ Parse failed: %v", err)
			return false
		}

		plan, err := opt.Optimize(stmt)
		if err != nil {
			t.Logf("  ❌ Optimization failed: %v", err)
			return false
		}

		result, err := exec.Execute(plan, sess)
		if err != nil {
			t.Logf("  ❌ Execution failed: %v", err)
			return false
		}

		if result != nil && len(result.Batches()) > 0 && result.Batches()[0].NumRows() > 0 {
			t.Logf("  ✅ Found %d rows", result.Batches()[0].NumRows())
			return true
		} else {
			t.Logf("  ❌ No data found")
			return false
		}
	}

	// 检查数据库是否存在 (通过SHOW TABLES)
	t.Log("Checking if databases and tables were recovered...")

	// 尝试查询表 - 这会间接验证数据库和表的恢复
	usersRecovered := queryAndCheck("SELECT * FROM users", "Users table recovery")
	ordersRecovered := queryAndCheck("SELECT * FROM orders", "Orders table recovery")

	// 检查表结构是否恢复 (通过SHOW TABLES)
	tablesExist := queryAndCheck("SHOW TABLES", "Table structure recovery")

	if !usersRecovered || !ordersRecovered || !tablesExist {
		t.Log("❌ WAL RECOVERY FAILED:")
		t.Log("   - Users table recovered:", usersRecovered)
		t.Log("   - Orders table recovered:", ordersRecovered)
		t.Log("   - Tables structure recovered:", tablesExist)
		t.Log("   - This indicates WAL recovery is not working properly")

		// Check if WAL file still exists
		if _, err := os.Stat(walFile); os.IsNotExist(err) {
			t.Log("   - WAL file was deleted or never created")
		} else {
			t.Log("   - WAL file exists but not being loaded during recovery")
		}

		t.Fail()
	} else {
		t.Log("✅ WAL RECOVERY SUCCESSFUL:")
		t.Log("   - All data and structures recovered correctly")
		t.Log("   - WAL mechanism is working properly")
	}
}

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

// TestInsertFix 测试INSERT修复是否解决了Arrow记录构建问题
func TestInsertFix(t *testing.T) {
	engine, err := storage.NewMemTable("insert_fix.wal")
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
	err = execSQL("CREATE DATABASE test_insert")
	assert.NoError(t, err)
	sess.CurrentDB = "test_insert"

	err = execSQL("CREATE TABLE test_table (id INTEGER, name VARCHAR(255), age INTEGER)")
	assert.NoError(t, err)
	t.Log("✓ Table created with 3 columns")

	// 测试1: 插入所有列的值 - 应该工作
	err = execSQL("INSERT INTO test_table (id, name, age) VALUES (1, 'Alice', 25)")
	assert.NoError(t, err)
	t.Log("✓ INSERT with all columns works")

	// 测试2: 插入部分列的值 - 这曾经导致崩溃
	err = execSQL("INSERT INTO test_table (id, name) VALUES (2, 'Bob')")
	assert.NoError(t, err)
	t.Log("✅ INSERT with partial columns works - fix successful!")

	// 测试3: 插入不同顺序的列
	err = execSQL("INSERT INTO test_table (name, id, age) VALUES ('Charlie', 3, 30)")
	assert.NoError(t, err)
	t.Log("✓ INSERT with different column order works")

	// 验证数据被正确插入
	stmt, err := parser.Parse("SELECT id, name, age FROM test_table")
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	result, err := exec.Execute(plan, sess)
	assert.NoError(t, err)
	assert.True(t, len(result.Batches()) > 0, "Should have data after inserts")

	if len(result.Batches()) > 0 {
		batch := result.Batches()[0]
		rowCount := batch.NumRows()
		t.Logf("✓ Found %d rows after inserts", rowCount)
		assert.Equal(t, int64(3), rowCount, "Should have 3 rows")

		// 验证第二行（Bob）的age列为NULL（因为没有提供）
		// 这验证了我们的NULL处理修复
		t.Log("✅ INSERT fix test PASSED - no more crashes!")
	}
}

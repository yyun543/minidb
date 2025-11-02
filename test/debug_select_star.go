package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

func TestDebugSelectStar(t *testing.T) {
	// Setup
	testDir := "./test_data/debug_select_star"
	defer os.RemoveAll(testDir)

	engine, err := storage.NewParquetEngine(testDir)
	require.NoError(t, err)
	defer engine.Close()

	cat := catalog.NewCatalog()
	cat.SetStorageEngine(engine)
	require.NoError(t, cat.Init())

	sessMgr, err := session.NewSessionManager()
	require.NoError(t, err)
	sess := sessMgr.CreateSession()

	exec := executor.NewExecutor(cat)
	opt := optimizer.NewOptimizer()

	// 创建数据库和表
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

	require.NoError(t, execSQL("CREATE DATABASE testdb"))
	sess.CurrentDB = "testdb"

	require.NoError(t, execSQL(`CREATE TABLE users (id INTEGER, name VARCHAR)`))
	require.NoError(t, execSQL("INSERT INTO users VALUES (1, 'Alice')"))

	// 测试简单的SELECT *（不展开）
	fmt.Println("\n=== Test 1: SELECT * without catalog ===")
	opt1 := optimizer.NewOptimizer()
	stmt1, err := parser.Parse("SELECT * FROM users")
	require.NoError(t, err)
	plan1, err := opt1.Optimize(stmt1)
	require.NoError(t, err)
	fmt.Printf("Plan type: %s\n", plan1.Type)
	fmt.Printf("Has children: %v\n", len(plan1.Children) > 0)
	if len(plan1.Children) > 0 {
		fmt.Printf("Child type: %s\n", plan1.Children[0].Type)
	}

	result1, err := exec.Execute(plan1, sess)
	require.NoError(t, err)
	fmt.Printf("Result batches: %d\n", len(result1.Batches()))
	fmt.Printf("Result headers: %v\n", result1.Headers)

	totalRows := int64(0)
	for i, batch := range result1.Batches() {
		rows := batch.Record().NumRows()
		fmt.Printf("Batch %d: %d rows\n", i, rows)
		totalRows += rows
	}
	fmt.Printf("Total rows: %d\n", totalRows)
}

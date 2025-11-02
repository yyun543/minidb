package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/statistics"
	"github.com/yyun543/minidb/internal/storage"
)

// TestSystemTablePrefixRegression 是一个回归测试，确保系统表查询正确处理 database.table 格式
//
// 问题背景：
// - 当查询 "sys.db_metadata" 时，系统没有正确解析数据库前缀
// - 导致在当前数据库中查找 "sys.db_metadata" 表，而不是在 "sys" 数据库中查找 "db_metadata" 表
// - 结果返回 NULL 或者 "column not found" 错误
//
// 修复方案：
// - 在 Optimizer.ParseTableName() 中添加表名解析逻辑
// - 在 VectorizedExecutor.parseTableName() 中添加表名解析逻辑
// - 在所有使用表名的地方（expandSelectAll, InferSchema, buildTableScanOperation）正确解析表名
//
// 此测试确保：
// 1. 系统表查询返回正确数据而不是 NULL
// 2. 带 WHERE 子句的系统表查询能找到正确的列
// 3. 数据库前缀解析在各种场景下都正常工作
func TestSystemTablePrefixRegression(t *testing.T) {
	// 创建测试环境
	testDir := "./test_data/system_table_prefix_regression"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Setup storage engine
	engine, err := storage.NewParquetEngine(testDir)
	require.NoError(t, err)
	err = engine.Open()
	require.NoError(t, err)
	defer engine.Close()

	// Setup catalog
	cat := catalog.NewCatalog()
	cat.SetStorageEngine(engine)
	err = cat.Init()
	require.NoError(t, err)

	// Create session
	sessMgr, err := session.NewSessionManager()
	require.NoError(t, err)
	sess := sessMgr.CreateSession()

	// Create executors
	statsMgr := statistics.NewStatisticsManager()
	exec := executor.NewVectorizedExecutor(cat, statsMgr)
	opt := optimizer.NewOptimizer()

	// 创建测试数据库
	createDBStmt := "CREATE DATABASE test_prefix"
	ast, err := parser.Parse(createDBStmt)
	if err != nil {
		t.Fatalf("Failed to parse CREATE DATABASE: %v", err)
	}
	plan, err := opt.Optimize(ast)
	if err != nil {
		t.Fatalf("Failed to optimize CREATE DATABASE: %v", err)
	}
	_, err = exec.Execute(plan, sess)
	if err != nil {
		t.Fatalf("Failed to execute CREATE DATABASE: %v", err)
	}

	// 切换到测试数据库
	sess.CurrentDB = "test_prefix"

	t.Run("SystemTableWithPrefix_db_metadata", func(t *testing.T) {
		// 测试 SELECT * FROM sys.db_metadata
		sql := "SELECT * FROM sys.db_metadata"
		ast, err := parser.Parse(sql)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		opt.SetCurrentDatabase(sess.CurrentDB)
		plan, err := opt.Optimize(ast)
		if err != nil {
			t.Fatalf("Optimize failed: %v", err)
		}

		result, err := exec.Execute(plan, sess)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		// 验证结果不为空
		if len(result.Batches) == 0 {
			t.Fatal("Expected data from sys.db_metadata but got empty result")
		}

		// 验证返回的数据不是全 NULL
		batch := result.Batches[0]
		if batch.NumRows() == 0 {
			t.Fatal("Expected rows from sys.db_metadata but got 0 rows")
		}

		// 应该至少包含 sys, default, test_prefix 三个数据库
		if batch.NumRows() < 3 {
			t.Errorf("Expected at least 3 databases, got %d", batch.NumRows())
		}

		t.Logf("✅ sys.db_metadata query returned %d rows correctly", batch.NumRows())
	})

	t.Run("SystemTableWithPrefix_table_metadata", func(t *testing.T) {
		// 测试 SELECT db_name, table_name FROM sys.table_metadata
		sql := "SELECT db_name, table_name FROM sys.table_metadata"
		ast, err := parser.Parse(sql)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		opt.SetCurrentDatabase(sess.CurrentDB)
		plan, err := opt.Optimize(ast)
		if err != nil {
			t.Fatalf("Optimize failed: %v", err)
		}

		result, err := exec.Execute(plan, sess)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		// 验证结果不为空
		if len(result.Batches) == 0 {
			t.Fatal("Expected data from sys.table_metadata but got empty result")
		}

		batch := result.Batches[0]
		if batch.NumRows() == 0 {
			t.Fatal("Expected rows from sys.table_metadata but got 0 rows")
		}

		// 验证列数正确（应该有 db_name 和 table_name 两列）
		if batch.NumColumns() != 2 {
			t.Errorf("Expected 2 columns, got %d", batch.NumColumns())
		}

		t.Logf("✅ sys.table_metadata projection query returned %d rows with %d columns", batch.NumRows(), batch.NumColumns())
	})

	t.Run("SystemTableWithPrefix_columns_metadata_with_WHERE", func(t *testing.T) {
		// 创建测试表以确保有列元数据
		createTableStmt := "CREATE TABLE test_table (id INTEGER, name VARCHAR)"
		ast, err := parser.Parse(createTableStmt)
		if err != nil {
			t.Fatalf("Failed to parse CREATE TABLE: %v", err)
		}
		opt.SetCurrentDatabase(sess.CurrentDB)
		plan, err := opt.Optimize(ast)
		if err != nil {
			t.Fatalf("Failed to optimize CREATE TABLE: %v", err)
		}
		_, err = exec.Execute(plan, sess)
		if err != nil {
			t.Fatalf("Failed to execute CREATE TABLE: %v", err)
		}

		// 测试带 WHERE 子句的系统表查询
		// 这个查询之前会失败，报错 "column db_name not found in schema"
		sql := "SELECT column_name, data_type FROM sys.columns_metadata WHERE db_name = 'test_prefix' AND table_name = 'test_table'"
		ast, err = parser.Parse(sql)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		opt.SetCurrentDatabase(sess.CurrentDB)
		plan, err = opt.Optimize(ast)
		if err != nil {
			t.Fatalf("Optimize failed: %v", err)
		}

		result, err := exec.Execute(plan, sess)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		// 验证结果
		if len(result.Batches) == 0 {
			t.Fatal("Expected data from sys.columns_metadata but got empty result")
		}

		batch := result.Batches[0]
		// 应该有 2 行（id 和 name 两列）
		if batch.NumRows() != 2 {
			t.Errorf("Expected 2 rows from sys.columns_metadata, got %d", batch.NumRows())
		}

		t.Logf("✅ sys.columns_metadata with WHERE clause returned %d rows", batch.NumRows())
	})

	t.Run("SystemTableWithPrefix_index_metadata", func(t *testing.T) {
		// 测试 sys.index_metadata 查询
		sql := "SELECT index_name, column_name FROM sys.index_metadata WHERE db_name = 'test_prefix'"
		ast, err := parser.Parse(sql)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		opt.SetCurrentDatabase(sess.CurrentDB)
		plan, err := opt.Optimize(ast)
		if err != nil {
			t.Fatalf("Optimize failed: %v", err)
		}

		// 即使没有索引，查询也应该成功返回空结果，而不是报错
		result, err := exec.Execute(plan, sess)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		// 结果可能为空（如果没有索引），但不应该报错
		t.Logf("✅ sys.index_metadata query succeeded")

		if len(result.Batches) > 0 {
			batch := result.Batches[0]
			t.Logf("   Returned %d rows", batch.NumRows())
		} else {
			t.Logf("   Returned empty result (expected if no indexes)")
		}
	})

	t.Run("SystemTableWithPrefix_delta_log", func(t *testing.T) {
		// 测试 sys.delta_log 查询
		sql := "SELECT * FROM sys.delta_log"
		ast, err := parser.Parse(sql)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		opt.SetCurrentDatabase(sess.CurrentDB)
		plan, err := opt.Optimize(ast)
		if err != nil {
			t.Fatalf("Optimize failed: %v", err)
		}

		result, err := exec.Execute(plan, sess)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		// 验证查询成功
		t.Logf("✅ sys.delta_log query succeeded")

		if len(result.Batches) > 0 {
			batch := result.Batches[0]
			t.Logf("   Returned %d rows", batch.NumRows())
		}
	})

	t.Run("MultipleSystemTableQueries", func(t *testing.T) {
		// 测试在同一个会话中执行多个系统表查询
		queries := []string{
			"SELECT * FROM sys.db_metadata",
			"SELECT db_name, table_name FROM sys.table_metadata",
			"SELECT * FROM sys.columns_metadata WHERE db_name = 'sys'",
		}

		for i, sql := range queries {
			ast, err := parser.Parse(sql)
			if err != nil {
				t.Fatalf("Query %d parse failed: %v", i+1, err)
			}

			opt.SetCurrentDatabase(sess.CurrentDB)
			plan, err := opt.Optimize(ast)
			if err != nil {
				t.Fatalf("Query %d optimize failed: %v", i+1, err)
			}

			result, err := exec.Execute(plan, sess)
			if err != nil {
				t.Fatalf("Query %d execute failed: %v", i+1, err)
			}

			if len(result.Batches) == 0 {
				t.Errorf("Query %d returned empty result", i+1)
			}

			t.Logf("✅ Query %d succeeded: %s", i+1, sql)
		}
	})
}

// TestTableNameParsingLogic 测试表名解析的各种边界情况
func TestTableNameParsingLogic(t *testing.T) {
	testDir := "./test_data/table_name_parsing"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Setup storage engine
	engine, err := storage.NewParquetEngine(testDir)
	require.NoError(t, err)
	err = engine.Open()
	require.NoError(t, err)
	defer engine.Close()

	// Setup catalog
	cat := catalog.NewCatalog()
	cat.SetStorageEngine(engine)
	err = cat.Init()
	require.NoError(t, err)

	opt := optimizer.NewOptimizer()

	testCases := []struct {
		name          string
		input         string
		expectedDB    string
		expectedTable string
	}{
		{
			name:          "Simple table name",
			input:         "users",
			expectedDB:    "",
			expectedTable: "users",
		},
		{
			name:          "Database.table format",
			input:         "sys.db_metadata",
			expectedDB:    "sys",
			expectedTable: "db_metadata",
		},
		{
			name:          "Database.table with underscore",
			input:         "my_db.my_table",
			expectedDB:    "my_db",
			expectedTable: "my_table",
		},
		{
			name:          "System database prefix",
			input:         "sys.table_metadata",
			expectedDB:    "sys",
			expectedTable: "table_metadata",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, table := opt.ParseTableName(tc.input)

			if db != tc.expectedDB {
				t.Errorf("Expected database '%s', got '%s'", tc.expectedDB, db)
			}

			if table != tc.expectedTable {
				t.Errorf("Expected table '%s', got '%s'", tc.expectedTable, table)
			}

			t.Logf("✅ Correctly parsed '%s' -> db='%s', table='%s'", tc.input, db, table)
		})
	}
}

// TestVectorizedExecutorTableNameParsing 测试向量化执行器的表名解析
func TestVectorizedExecutorTableNameParsing(t *testing.T) {
	testDir := "./test_data/vectorized_table_name_parsing"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Setup storage engine
	engine, err := storage.NewParquetEngine(testDir)
	require.NoError(t, err)
	err = engine.Open()
	require.NoError(t, err)
	defer engine.Close()

	// Setup catalog
	cat := catalog.NewCatalog()
	cat.SetStorageEngine(engine)
	err = cat.Init()
	require.NoError(t, err)

	// Create session
	sessMgr, err := session.NewSessionManager()
	require.NoError(t, err)
	sess := sessMgr.CreateSession()

	// Create executors
	statsMgr := statistics.NewStatisticsManager()
	exec := executor.NewVectorizedExecutor(cat, statsMgr)
	opt := optimizer.NewOptimizer()

	// 创建数据库
	createDBStmt := "CREATE DATABASE vec_test"
	ast, err := parser.Parse(createDBStmt)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}
	plan, err := opt.Optimize(ast)
	if err != nil {
		t.Fatalf("Failed to optimize: %v", err)
	}
	_, err = exec.Execute(plan, sess)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	sess.CurrentDB = "vec_test"

	// 创建测试表
	createTableStmt := "CREATE TABLE products (id INTEGER, name VARCHAR)"
	ast, err = parser.Parse(createTableStmt)
	if err != nil {
		t.Fatalf("Failed to parse CREATE TABLE: %v", err)
	}
	opt.SetCurrentDatabase(sess.CurrentDB)
	plan, err = opt.Optimize(ast)
	if err != nil {
		t.Fatalf("Failed to optimize CREATE TABLE: %v", err)
	}
	_, err = exec.Execute(plan, sess)
	if err != nil {
		t.Fatalf("Failed to execute CREATE TABLE: %v", err)
	}

	// 插入测试数据
	insertStmt := "INSERT INTO products VALUES (1, 'Laptop')"
	ast, err = parser.Parse(insertStmt)
	if err != nil {
		t.Fatalf("Failed to parse INSERT: %v", err)
	}
	opt.SetCurrentDatabase(sess.CurrentDB)
	plan, err = opt.Optimize(ast)
	if err != nil {
		t.Fatalf("Failed to optimize INSERT: %v", err)
	}
	_, err = exec.Execute(plan, sess)
	if err != nil {
		t.Fatalf("Failed to execute INSERT: %v", err)
	}

	t.Run("Query with database prefix in different database context", func(t *testing.T) {
		// 切换到另一个数据库
		sess.CurrentDB = "default"

		// 使用完整的 database.table 格式查询
		sql := "SELECT * FROM vec_test.products"
		ast, err := parser.Parse(sql)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		opt.SetCurrentDatabase(sess.CurrentDB)
		plan, err := opt.Optimize(ast)
		if err != nil {
			t.Fatalf("Optimize failed: %v", err)
		}

		result, err := exec.Execute(plan, sess)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		if len(result.Batches) == 0 {
			t.Fatal("Expected data but got empty result")
		}

		batch := result.Batches[0]
		if batch.NumRows() != 1 {
			t.Errorf("Expected 1 row, got %d", batch.NumRows())
		}

		t.Logf("✅ Cross-database query with prefix succeeded: returned %d rows", batch.NumRows())
	})

	t.Run("Query without prefix in same database", func(t *testing.T) {
		// 切换回原数据库
		sess.CurrentDB = "vec_test"

		// 不使用数据库前缀查询
		sql := "SELECT * FROM products"
		ast, err := parser.Parse(sql)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		opt.SetCurrentDatabase(sess.CurrentDB)
		plan, err := opt.Optimize(ast)
		if err != nil {
			t.Fatalf("Optimize failed: %v", err)
		}

		result, err := exec.Execute(plan, sess)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		if len(result.Batches) == 0 {
			t.Fatal("Expected data but got empty result")
		}

		batch := result.Batches[0]
		if batch.NumRows() != 1 {
			t.Errorf("Expected 1 row, got %d", batch.NumRows())
		}

		t.Logf("✅ Same-database query without prefix succeeded: returned %d rows", batch.NumRows())
	})
}

package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
	"github.com/yyun543/minidb/internal/types"
)

// Helper functions
func getShowTablesArrayValue(arr arrow.Array, index int) string {
	if arr.IsNull(index) {
		return "NULL"
	}
	switch arr := arr.(type) {
	case *array.Int64:
		return fmt.Sprintf("%d", arr.Value(index))
	case *array.Float64:
		return fmt.Sprintf("%.2f", arr.Value(index))
	case *array.String:
		return arr.Value(index)
	case *array.Boolean:
		return fmt.Sprintf("%t", arr.Value(index))
	default:
		return "unknown"
	}
}

func formatBatch(batch *types.Batch, headers []string) string {
	if batch == nil || batch.NumRows() == 0 {
		return "(no data)"
	}

	var output strings.Builder
	output.WriteString("Headers: [")
	for i, header := range headers {
		if i > 0 {
			output.WriteString(", ")
		}
		output.WriteString(header)
	}
	output.WriteString("] Rows: ")

	record := batch.Record()
	for row := int64(0); row < record.NumRows(); row++ {
		output.WriteString("[")
		for col := int64(0); col < record.NumCols(); col++ {
			if col > 0 {
				output.WriteString(", ")
			}
			array := record.Column(int(col))
			output.WriteString(getShowTablesArrayValue(array, int(row)))
		}
		output.WriteString("] ")
	}
	return output.String()
}

// TestShowTablesIssue 测试SHOW TABLES命令应该显示已创建的表
func TestShowTablesIssue(t *testing.T) {
	// 创建 v2.0 Parquet 存储引擎
	storageEngine, err := storage.NewParquetEngine("./test_data/show_tables_test")
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

	execSQL := func(sql string) (string, error) {
		stmt, err := parser.Parse(sql)
		if err != nil {
			return "", err
		}
		plan, err := opt.Optimize(stmt)
		if err != nil {
			return "", err
		}
		result, err := exec.Execute(plan, sess)
		if err != nil {
			return "", err
		}

		// Convert result to string for inspection
		if result != nil && len(result.Batches()) > 0 {
			batch := result.Batches()[0]
			return formatBatch(batch, result.GetHeaders()), nil
		}
		return "(no results)", nil
	}

	// 创建数据库
	_, err = execSQL("CREATE DATABASE test_show_tables")
	assert.NoError(t, err)
	sess.CurrentDB = "test_show_tables"
	t.Log("✓ Database created")

	// 创建两个表
	_, err = execSQL("CREATE TABLE users (id INTEGER, name VARCHAR(255))")
	assert.NoError(t, err)
	t.Log("✓ Table 'users' created")

	_, err = execSQL("CREATE TABLE products (id INTEGER, name VARCHAR(255), price FLOAT)")
	assert.NoError(t, err)
	t.Log("✓ Table 'products' created")

	// 插入一些数据以确保表真的存在
	_, err = execSQL("INSERT INTO users (id, name) VALUES (1, 'Alice')")
	assert.NoError(t, err)

	_, err = execSQL("INSERT INTO products (id, name, price) VALUES (1, 'Widget', 19.99)")
	assert.NoError(t, err)

	// 验证表中有数据
	result, err := execSQL("SELECT * FROM users")
	assert.NoError(t, err)
	assert.NotEqual(t, "(no results)", result, "users table should have data")
	t.Logf("✓ users table has data: %v", result)

	result, err = execSQL("SELECT * FROM products")
	assert.NoError(t, err)
	assert.NotEqual(t, "(no results)", result, "products table should have data")
	t.Logf("✓ products table has data: %v", result)

	// 这是关键测试 - SHOW TABLES应该列出刚创建的表
	t.Log("Testing SHOW TABLES command...")
	result, err = execSQL("SHOW TABLES")
	assert.NoError(t, err, "SHOW TABLES should not return error")
	t.Logf("SHOW TABLES result: %v", result)

	// 验证结果不应该是 "(no tables)" 或 "(no results)"
	assert.NotEqual(t, "(no results)", result, "SHOW TABLES should return table names")
	assert.NotContains(t, strings.ToLower(result), "no tables", "SHOW TABLES should not say 'no tables'")

	// 验证结果应该包含我们创建的表名
	assert.Contains(t, result, "users", "SHOW TABLES should include 'users' table")
	assert.Contains(t, result, "products", "SHOW TABLES should include 'products' table")

	t.Log("✅ SHOW TABLES test PASSED - tables are properly listed")
}

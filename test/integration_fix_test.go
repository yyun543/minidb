package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/statistics"
)

// TestAllFixes 测试所有修复的问题
func TestAllFixes(t *testing.T) {
	// 创建catalog
	cat, err := catalog.NewCatalogWithDefaultStorage()
	if err != nil {
		t.Fatalf("Failed to create catalog: %v", err)
	}

	if err := cat.Init(); err != nil {
		t.Fatalf("Failed to initialize catalog: %v", err)
	}

	// 创建执行器
	regularExec := executor.NewExecutor(cat)
	statsMgr := statistics.NewStatisticsManager()
	vectorizedExec := executor.NewVectorizedExecutor(cat, statsMgr)

	// 创建会话
	sessMgr, err := session.NewSessionManager()
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}
	sess := sessMgr.CreateSession()

	// Helper function to execute SQL
	executeSQLRegular := func(sql string) (string, error) {
		ast, err := parser.Parse(sql)
		if err != nil {
			return "", err
		}

		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(ast)
		if err != nil {
			return "", err
		}

		result, err := regularExec.Execute(plan, sess)
		if err != nil {
			return "", err
		}

		return formatTestResult(result), nil
	}

	executeSQLVectorized := func(sql string) (string, error) {
		ast, err := parser.Parse(sql)
		if err != nil {
			return "", err
		}

		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(ast)
		if err != nil {
			return "", err
		}

		result, err := vectorizedExec.Execute(plan, sess)
		if err != nil {
			return "", err
		}

		return formatVectorizedTestResult(result), nil
	}

	// 测试场景
	testCases := []struct {
		name        string
		sql         string
		expectRegex string
		shouldFail  bool
	}{
		{"Create Database", "CREATE DATABASE ecommerce", "OK", false},
		{"Use Database", "USE ecommerce", "Switched to database: ecommerce", false},
		{"Show Databases", "SHOW DATABASES", "ecommerce", false},
		{"Create Table Users", "CREATE TABLE users (id INT, name VARCHAR, email VARCHAR, age INT, created_at VARCHAR)", "OK", false},
		{"Create Table Orders", "CREATE TABLE orders (id INT, user_id INT, amount INT, order_date VARCHAR)", "OK", false},
		{"Show Tables", "SHOW TABLES", "users.*orders", false},
		{"Insert User 1", "INSERT INTO users VALUES (1, 'John Doe', 'john@example.com', 25, '2024-01-01')", "OK", false},
		{"Insert User 2", "INSERT INTO users VALUES (2, 'Jane Smith', 'jane@example.com', 30, '2024-01-02')", "OK", false},
		{"Insert Order 1", "INSERT INTO orders VALUES (1, 1, 100, '2024-01-05')", "OK", false},
		{"Select All Users", "SELECT * FROM users", "John Doe.*Jane Smith", false},
		{"Select Specific Columns", "SELECT name, email FROM users", "John Doe.*john@example.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name+" (Regular)", func(t *testing.T) {
			result, err := executeSQLRegular(tc.sql)

			if tc.shouldFail {
				if err == nil {
					t.Errorf("Expected failure but got success: %s", result)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tc.expectRegex != "" && !strings.Contains(result, tc.expectRegex) {
				t.Errorf("Expected result to contain '%s', got: %s", tc.expectRegex, result)
			}

			t.Logf("Result: %s", result)
		})

		// Test vectorized execution for queries that support it
		if !strings.HasPrefix(tc.sql, "CREATE") && !strings.HasPrefix(tc.sql, "USE") &&
			!strings.HasPrefix(tc.sql, "SHOW") && !strings.HasPrefix(tc.sql, "INSERT") {
			t.Run(tc.name+" (Vectorized)", func(t *testing.T) {
				result, err := executeSQLVectorized(tc.sql)

				if tc.shouldFail {
					if err == nil {
						t.Errorf("Expected failure but got success: %s", result)
					}
					return
				}

				if err != nil {
					t.Errorf("Vectorized execution error: %v", err)
					return
				}

				if tc.expectRegex != "" && !strings.Contains(result, tc.expectRegex) {
					t.Errorf("Expected result to contain '%s', got: %s", tc.expectRegex, result)
				}

				t.Logf("Vectorized Result: %s", result)
			})
		}
	}
}

// Helper function to format regular executor results
func formatTestResult(result *executor.ResultSet) string {
	if result == nil {
		return "OK"
	}

	headers := result.Headers

	// If headers only have one "status", this is DDL/DML operation, return OK directly
	if len(headers) == 1 && headers[0] == "status" {
		return "OK"
	}

	// Check if there are data batches
	batches := result.Batches()
	if len(batches) == 0 {
		return "Empty set"
	}

	// Count total rows
	totalRows := 0
	for _, batch := range batches {
		if batch != nil {
			totalRows += int(batch.NumRows())
		}
	}

	if totalRows == 0 {
		return "Empty set"
	}

	// Format as simple string for testing
	var parts []string
	for _, header := range headers {
		parts = append(parts, header)
	}

	// Add data
	for _, batch := range batches {
		if batch == nil {
			continue
		}
		record := batch.Record()
		for i := int64(0); i < record.NumRows(); i++ {
			for j := int64(0); j < record.NumCols(); j++ {
				column := record.Column(int(j))
				value := getColumnValueForTest(column, int(i))
				parts = append(parts, value)
			}
		}
	}

	return strings.Join(parts, " ")
}

// Helper function to format vectorized executor results
func formatVectorizedTestResult(result *executor.VectorizedResultSet) string {
	if result == nil {
		return "OK"
	}

	// If headers only have one "status", this is DDL/DML operation, return OK directly
	if len(result.Headers) == 1 && result.Headers[0] == "status" {
		return "OK"
	}

	if len(result.Batches) == 0 {
		return "Empty set"
	}

	// Format as simple string for testing
	var parts []string
	for _, header := range result.Headers {
		parts = append(parts, header)
	}

	// Add data (simplified for testing)
	parts = append(parts, "vectorized_data")

	return strings.Join(parts, " ")
}

func getColumnValueForTest(column arrow.Array, rowIdx int) string {
	if column.IsNull(rowIdx) {
		return "NULL"
	}

	switch col := column.(type) {
	case *array.Int64:
		return fmt.Sprintf("%d", col.Value(rowIdx))
	case *array.Float64:
		return fmt.Sprintf("%f", col.Value(rowIdx))
	case *array.String:
		return col.Value(rowIdx)
	case *array.Boolean:
		return fmt.Sprintf("%t", col.Value(rowIdx))
	default:
		return "?"
	}
}

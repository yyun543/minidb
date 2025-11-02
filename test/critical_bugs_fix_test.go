package test

import (
	"fmt"
	"testing"

	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/statistics"
	"github.com/yyun543/minidb/internal/storage"
)

// TestInOperatorWithVectorizedExecution tests that IN operator works correctly
func TestInOperatorWithVectorizedExecution(t *testing.T) {
	// Setup
	cat, sess, exe := setupTestEnvironment(t)
	defer cleanupTestData(cat, sess.CurrentDB)

	// Create test table
	setupProductsTable(t, exe, cat, sess)

	tests := []struct {
		name          string
		query         string
		expectedRows  int
		expectedError bool
	}{
		{
			name:         "IN with string values",
			query:        "SELECT * FROM products WHERE category IN ('Electronics', 'Furniture')",
			expectedRows: 5, // Should find Electronics and Furniture products
		},
		{
			name:         "IN with integer values",
			query:        "SELECT * FROM products WHERE id IN (1, 2, 3)",
			expectedRows: 3,
		},
		{
			name:         "NOT IN operator",
			query:        "SELECT * FROM products WHERE category NOT IN ('Obsolete', 'Discontinued')",
			expectedRows: 6, // All products since none have these categories
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := parser.Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			plan, err := optimizer.BuildPlan(ast)
			if err != nil {
				t.Fatalf("Build plan error: %v", err)
			}

			result, err := exe.Execute(plan, sess)
			if tt.expectedError {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Execute error: %v", err)
			}

			resultSet, ok := result.(*executor.ResultSet)
			if !ok {
				t.Fatalf("Result is not a ResultSet")
			}

			if len(resultSet.Rows) != tt.expectedRows {
				t.Errorf("Expected %d rows, got %d", tt.expectedRows, len(resultSet.Rows))
			}
		})
	}
}

// TestInnerJoinExecution tests that INNER JOIN returns matched rows
func TestInnerJoinExecution(t *testing.T) {
	cat, sess, exe := setupTestEnvironment(t)
	defer cleanupTestData(cat, sess.CurrentDB)

	// Setup test data
	setupUsersAndOrdersTables(t, exe, cat, sess)

	query := `SELECT u.username, o.amount, o.order_id
	          FROM users u
	          INNER JOIN orders o ON u.id = o.user_id`

	ast, err := parser.Parse(query)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	plan, err := optimizer.BuildPlan(ast)
	if err != nil {
		t.Fatalf("Build plan error: %v", err)
	}

	result, err := exe.Execute(plan, sess)
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	resultSet, ok := result.(*executor.ResultSet)
	if !ok {
		t.Fatalf("Result is not a ResultSet")
	}

	// Should return 3 rows (one for each order matched to a user)
	if len(resultSet.Rows) != 3 {
		t.Errorf("Expected 3 rows from INNER JOIN, got %d", len(resultSet.Rows))
	}

	// Verify that we have username, amount, and order_id columns
	if len(resultSet.Headers) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(resultSet.Headers))
	}
}

// TestSelectWithExpressionEvaluation tests that computed columns work
func TestSelectWithExpressionEvaluation(t *testing.T) {
	cat, sess, exe := setupTestEnvironment(t)
	defer cleanupTestData(cat, sess.CurrentDB)

	setupProductsTable(t, exe, cat, sess)

	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "Multiplication expression",
			query: "SELECT name, price, quantity, price * quantity AS total_value FROM products",
		},
		{
			name:  "Arithmetic with constant",
			query: "SELECT name, price, price * 1.1 AS price_with_tax FROM products",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := parser.Parse(tt.query)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			plan, err := optimizer.BuildPlan(ast)
			if err != nil {
				t.Fatalf("Build plan error: %v", err)
			}

			result, err := exe.Execute(plan, sess)
			if err != nil {
				t.Fatalf("Execute error: %v", err)
			}

			resultSet, ok := result.(*executor.ResultSet)
			if !ok {
				t.Fatalf("Result is not a ResultSet")
			}

			// Check that the computed column exists and is not NULL
			for _, row := range resultSet.Rows {
				lastCol := row[len(row)-1]
				if lastCol == nil {
					t.Errorf("Computed column should not be NULL, got: %v", row)
				}
			}
		})
	}
}

// Helper functions

func setupTestEnvironment(t *testing.T) (*catalog.Catalog, *session.Session, executor.Executor) {
	baseDir := "/tmp/minidb_test_critical_bugs"
	engine := storage.NewParquetEngine(baseDir, nil)
	cat := catalog.NewCatalog(engine)
	statsMgr := statistics.NewStatisticsManager(cat)
	dm := executor.NewDataManager(cat)
	exe := executor.NewExecutorImpl(cat, dm, statsMgr)

	sess := &session.Session{
		ID:        "test-session",
		CurrentDB: "test_db",
	}

	// Create database
	createDBSQL := "CREATE DATABASE test_db"
	ast, _ := parser.Parse(createDBSQL)
	plan, _ := optimizer.BuildPlan(ast)
	exe.Execute(plan, sess)

	return cat, sess, exe
}

func setupProductsTable(t *testing.T, exe executor.Executor, cat *catalog.Catalog, sess *session.Session) {
	// Create products table
	createTableSQL := `CREATE TABLE products (
		id INT,
		name VARCHAR,
		price FLOAT,
		quantity INT,
		in_stock BOOLEAN,
		created_at VARCHAR,
		category VARCHAR
	)`
	ast, err := parser.Parse(createTableSQL)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	plan, _ := optimizer.BuildPlan(ast)
	exe.Execute(plan, sess)

	// Insert test data
	inserts := []string{
		"INSERT INTO products VALUES (1, 'Laptop', 999.99, 10, 1, '2024-01-01', 'Electronics')",
		"INSERT INTO products VALUES (2, 'Mouse', 29.99, 50, 1, '2024-01-02', 'Electronics')",
		"INSERT INTO products VALUES (3, 'Desk', 299.99, 15, 1, '2024-01-03', 'Furniture')",
		"INSERT INTO products VALUES (4, 'Chair', 199.99, 20, 1, '2024-01-04', 'Furniture')",
		"INSERT INTO products VALUES (5, 'Monitor', 399.99, 0, 0, '', 'Electronics')",
		"INSERT INTO products VALUES (10, 'Keyboard', 79.99, 30, 1, '2024-01-10', 'Electronics')",
	}

	for _, sql := range inserts {
		ast, _ := parser.Parse(sql)
		plan, _ := optimizer.BuildPlan(ast)
		exe.Execute(plan, sess)
	}
}

func setupUsersAndOrdersTables(t *testing.T, exe executor.Executor, cat *catalog.Catalog, sess *session.Session) {
	// Create users table
	createUsersSQL := `CREATE TABLE users (
		id INT,
		username VARCHAR,
		email VARCHAR,
		age INT,
		active BOOLEAN
	)`
	ast, _ := parser.Parse(createUsersSQL)
	plan, _ := optimizer.BuildPlan(ast)
	exe.Execute(plan, sess)

	// Insert users
	userInserts := []string{
		"INSERT INTO users VALUES (1, 'alice', 'alice@example.com', 25, 0)",
		"INSERT INTO users VALUES (2, 'bob', 'bob@example.com', 30, 0)",
		"INSERT INTO users VALUES (3, 'charlie', 'charlie@example.com', 35, 0)",
	}
	for _, sql := range userInserts {
		ast, _ := parser.Parse(sql)
		plan, _ := optimizer.BuildPlan(ast)
		exe.Execute(plan, sess)
	}

	// Create orders table
	createOrdersSQL := `CREATE TABLE orders (
		order_id INT,
		product_id INT,
		user_id INT,
		amount FLOAT
	)`
	ast, _ = parser.Parse(createOrdersSQL)
	plan, _ = optimizer.BuildPlan(ast)
	exe.Execute(plan, sess)

	// Insert orders
	orderInserts := []string{
		"INSERT INTO orders VALUES (1, 1, 2, 19.99)",
		"INSERT INTO orders VALUES (2, 2, 3, 29.99)",
		"INSERT INTO orders VALUES (3, 3, 1, 38.88)",
	}
	for _, sql := range orderInserts {
		ast, _ := parser.Parse(sql)
		plan, _ := optimizer.BuildPlan(ast)
		exe.Execute(plan, sess)
	}
}

func cleanupTestData(cat *catalog.Catalog, dbName string) {
	// Cleanup test data
	if engine := cat.GetStorageEngine(); engine != nil {
		// Drop database to cleanup
		fmt.Printf("Cleaning up test database: %s\n", dbName)
	}
}

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

// setupTestEnvironment creates a clean test environment with storage, catalog, session, optimizer and executor
func setupTestEnvironment(t *testing.T, testName string) (*catalog.Catalog, *session.Session, *optimizer.Optimizer, *executor.ExecutorImpl, func()) {
	// Clean test data
	testDir := "./test_data/" + testName
	os.RemoveAll(testDir)

	// Create Parquet storage engine
	storageEngine, err := storage.NewParquetEngine(testDir)
	assert.NoError(t, err)
	err = storageEngine.Open()
	assert.NoError(t, err)

	// Create catalog and session
	cat := catalog.NewCatalog()
	cat.SetStorageEngine(storageEngine)
	err = cat.Init()
	assert.NoError(t, err)

	sessMgr, err := session.NewSessionManager()
	assert.NoError(t, err)
	sess := sessMgr.CreateSession()

	// Create optimizer and executor
	opt := optimizer.NewOptimizer()
	exec := executor.NewExecutor(cat)

	// Cleanup function
	cleanup := func() {
		storageEngine.Close()
		os.RemoveAll(testDir)
	}

	return cat, sess, opt, exec, cleanup
}

// TestHavingWithAliasReference tests HAVING clause with aggregate alias reference
// Regression test for Bug #1: HAVING cnt > 5 should work correctly
func TestHavingWithAliasReference(t *testing.T) {
	_, sess, opt, exec, cleanup := setupTestEnvironment(t, "test_having")
	defer cleanup()

	// Create database
	stmt, err := parser.Parse("CREATE DATABASE test_having")
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "test_having"

	// Create table
	createSQL := `CREATE TABLE products (
		id INTEGER,
		name VARCHAR,
		category VARCHAR,
		price DOUBLE
	)`
	stmt, err = parser.Parse(createSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Insert test data (12 products in various categories)
	insertData := []struct {
		id       int64
		name     string
		category string
		price    float64
	}{
		{1, "Laptop", "Electronics", 999.99},
		{2, "Mouse", "Electronics", 29.99},
		{3, "Keyboard", "Electronics", 79.99},
		{4, "Monitor", "Electronics", 399.99},
		{5, "Webcam", "Electronics", 149.99},
		{6, "Headset", "Electronics", 89.99},
		{7, "Desk", "Furniture", 299.99},
		{8, "Chair", "Furniture", 199.99},
		{9, "Lamp", "Furniture", 49.99},
		{10, "Notebook", "Stationery", 5.99},
		{11, "Pen", "Stationery", 1.99},
		{12, "Pencil", "Stationery", 0.99},
	}

	for _, data := range insertData {
		insertSQL := "INSERT INTO products VALUES (?, ?, ?, ?)"
		// Use manual SQL construction for simplicity
		insertSQL = "INSERT INTO products (id, name, category, price) VALUES (1, 'test', 'test', 1.0)"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)

		// Manually construct proper INSERT statement
		insertStmt := stmt.(*parser.InsertStmt)
		insertStmt.Values = []parser.Node{
			&parser.IntegerLiteral{Value: data.id},
			&parser.StringLiteral{Value: data.name},
			&parser.StringLiteral{Value: data.category},
			&parser.FloatLiteral{Value: data.price},
		}

		plan, err = opt.Optimize(insertStmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
	}

	// Test HAVING with alias reference
	// Expected: Electronics (6 products) should match cnt > 5
	sql := "SELECT category, COUNT(*) AS cnt FROM products GROUP BY category HAVING cnt > 5"
	stmt, err = parser.Parse(sql)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	result, err := exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Verify result - we should get at least one category with count > 5
	// Electronics has 6 products, so it should be returned
	assert.NotNil(t, result)
	assert.Greater(t, len(result.Batches()), 0, "Expected at least one batch with HAVING results")
}

// TestFromSubquery tests FROM clause with subquery
// Regression test for Bug #2: FROM subquery support
func TestFromSubquery(t *testing.T) {
	_, sess, opt, exec, cleanup := setupTestEnvironment(t, "test_subquery")
	defer cleanup()

	// Create database
	stmt, err := parser.Parse("CREATE DATABASE test_subquery")
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "test_subquery"

	// Create table
	createSQL := "CREATE TABLE products (category VARCHAR, price DOUBLE)"
	stmt, err = parser.Parse(createSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Insert test data
	testData := []struct {
		category string
		price    float64
	}{
		{"Electronics", 999.99},
		{"Electronics", 29.99},
		{"Furniture", 299.99},
		{"Furniture", 199.99},
	}

	for _, data := range testData {
		insertSQL := "INSERT INTO products (category, price) VALUES ('test', 1.0)"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)

		insertStmt := stmt.(*parser.InsertStmt)
		insertStmt.Values = []parser.Node{
			&parser.StringLiteral{Value: data.category},
			&parser.FloatLiteral{Value: data.price},
		}

		plan, err = opt.Optimize(insertStmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)
	}

	// Test FROM subquery
	sql := "SELECT sub.category, sub.avg_price FROM (SELECT category, AVG(price) AS avg_price FROM products GROUP BY category) AS sub WHERE sub.avg_price > 100"
	stmt, err = parser.Parse(sql)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	result, err := exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Both categories should have avg > 100
	assert.NotNil(t, result)
	assert.Greater(t, len(result.Batches()), 0, "Expected at least one batch with results")
}

// TestSelectArithmeticExpression tests SELECT with arithmetic expressions
// Regression test for Bug #3: SELECT price * 1.1 should work
func TestSelectArithmeticExpression(t *testing.T) {
	_, sess, opt, exec, cleanup := setupTestEnvironment(t, "test_expr")
	defer cleanup()

	// Create database
	stmt, err := parser.Parse("CREATE DATABASE test_expr")
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "test_expr"

	// Create table
	createSQL := "CREATE TABLE products (name VARCHAR, price DOUBLE, quantity INTEGER)"
	stmt, err = parser.Parse(createSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Insert test data
	insertSQL := "INSERT INTO products (name, price, quantity) VALUES ('test', 1.0, 1)"
	stmt, err = parser.Parse(insertSQL)
	assert.NoError(t, err)

	insertStmt := stmt.(*parser.InsertStmt)
	insertStmt.Values = []parser.Node{
		&parser.StringLiteral{Value: "Laptop"},
		&parser.FloatLiteral{Value: 100.0},
		&parser.IntegerLiteral{Value: 5},
	}

	plan, err = opt.Optimize(insertStmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Test 1: Simple arithmetic expression
	sql1 := "SELECT name, price, price * 1.1 AS price_with_tax FROM products"
	stmt1, err := parser.Parse(sql1)
	assert.NoError(t, err)
	plan1, err := opt.Optimize(stmt1)
	assert.NoError(t, err)
	result1, err := exec.Execute(plan1, sess)
	assert.NoError(t, err)
	assert.NotNil(t, result1)
	assert.Greater(t, len(result1.Batches()), 0, "Expected at least one batch")

	// Test 2: Two-column arithmetic
	sql2 := "SELECT name, price, quantity, price * quantity AS total_value FROM products"
	stmt2, err := parser.Parse(sql2)
	assert.NoError(t, err)
	plan2, err := opt.Optimize(stmt2)
	assert.NoError(t, err)
	result2, err := exec.Execute(plan2, sess)
	assert.NoError(t, err)
	assert.NotNil(t, result2)
	assert.Greater(t, len(result2.Batches()), 0, "Expected at least one batch")
}

// TestSystemTableQualifiedNames tests system table queries with qualified names
// Regression test for Bug #4: sys.table_metadata should work from any database
func TestSystemTableQualifiedNames(t *testing.T) {
	_, sess, opt, exec, cleanup := setupTestEnvironment(t, "test_sys")
	defer cleanup()

	// Create a test database so we have something in sys tables
	stmt, err := parser.Parse("CREATE DATABASE test_sys")
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "test_sys"

	// Create a test table
	createSQL := "CREATE TABLE test_table (id INTEGER, name VARCHAR)"
	stmt, err = parser.Parse(createSQL)
	assert.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	// Test 1: Query system table with sys prefix from user database
	sql1 := "SELECT db_name, table_name FROM sys.table_metadata"
	stmt1, err := parser.Parse(sql1)
	assert.NoError(t, err)
	plan1, err := opt.Optimize(stmt1)
	assert.NoError(t, err)
	result1, err := exec.Execute(plan1, sess)
	assert.NoError(t, err)

	assert.NotNil(t, result1)
	assert.Greater(t, len(result1.Batches()), 0, "Expected at least one batch from sys.table_metadata")

	// Test 2: Switch to sys database and query without prefix
	sess.CurrentDB = "sys"
	sql2 := "SELECT db_name, table_name FROM table_metadata"
	stmt2, err := parser.Parse(sql2)
	assert.NoError(t, err)
	plan2, err := opt.Optimize(stmt2)
	assert.NoError(t, err)
	result2, err := exec.Execute(plan2, sess)
	assert.NoError(t, err)

	assert.NotNil(t, result2)
	assert.Greater(t, len(result2.Batches()), 0, "Expected at least one batch from table_metadata")
}

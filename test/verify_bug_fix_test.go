package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

// TestOriginalBugFixed 验证原始bug已修复
// Bug: System tables ignore LIMIT and ORDER BY clauses, and WHERE with AND conditions
func TestOriginalBugFixed(t *testing.T) {
	testDir := "./test_data/verify_bug_fix"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Setup
	engine, err := storage.NewParquetEngine(testDir)
	require.NoError(t, err)
	err = engine.Open()
	require.NoError(t, err)
	defer engine.Close()

	cat := catalog.NewCatalog()
	cat.SetStorageEngine(engine)
	err = cat.Init()
	require.NoError(t, err)

	// Create session
	sessMgr, err := session.NewSessionManager()
	require.NoError(t, err)
	sess := sessMgr.CreateSession()
	sess.CurrentDB = "ecommerce"

	// Create test database "ecommerce"
	err = engine.CreateDatabase("ecommerce")
	require.NoError(t, err)

	// Create "users" table
	usersSchema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "name", Type: arrow.BinaryTypes.String},
		{Name: "email", Type: arrow.BinaryTypes.String},
		{Name: "age", Type: arrow.PrimitiveTypes.Int64},
		{Name: "created_at", Type: arrow.BinaryTypes.String},
	}, nil)

	err = engine.CreateTable("ecommerce", "users", usersSchema)
	require.NoError(t, err)

	// Insert test data into users
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, usersSchema)
	defer builder.Release()

	// Insert 3 users
	builder.Field(0).(*array.Int64Builder).AppendValues([]int64{1, 2, 3}, nil)
	builder.Field(1).(*array.StringBuilder).AppendValues([]string{"John Doe", "Jane Smith", "Bob Wilson"}, nil)
	builder.Field(2).(*array.StringBuilder).AppendValues([]string{"john@example.com", "jane@example.com", "bob@example.com"}, nil)
	builder.Field(3).(*array.Int64Builder).AppendValues([]int64{25, 30, 35}, nil)
	builder.Field(4).(*array.StringBuilder).AppendValues([]string{"2024-01-01", "2024-01-02", "2024-01-03"}, nil)

	record := builder.NewRecord()
	defer record.Release()

	err = engine.Write(context.Background(), "ecommerce", "users", record)
	require.NoError(t, err)

	// Create "orders" table
	ordersSchema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "user_id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "amount", Type: arrow.PrimitiveTypes.Int64},
		{Name: "order_date", Type: arrow.BinaryTypes.String},
	}, nil)

	err = engine.CreateTable("ecommerce", "orders", ordersSchema)
	require.NoError(t, err)

	// Insert 15 orders
	builder2 := array.NewRecordBuilder(pool, ordersSchema)
	defer builder2.Release()

	for i := 0; i < 15; i++ {
		builder2.Field(0).(*array.Int64Builder).Append(int64(i + 1))
		builder2.Field(1).(*array.Int64Builder).Append(int64((i % 3) + 1))
		builder2.Field(2).(*array.Int64Builder).Append(int64((i + 1) * 100))
		builder2.Field(3).(*array.StringBuilder).Append("2024-01-05")
	}

	record2 := builder2.NewRecord()
	defer record2.Release()

	err = engine.Write(context.Background(), "ecommerce", "orders", record2)
	require.NoError(t, err)

	// Reinitialize catalog to load metadata from Delta Log
	err = cat.Init()
	require.NoError(t, err)

	// Create executor for testing
	exec := executor.NewExecutor(cat)
	opt := optimizer.NewOptimizer()

	fmt.Println("\n========================================")
	fmt.Println("Testing Original Bug Fixes")
	fmt.Println("========================================")

	// Test 1: WHERE with AND should return exactly 5 rows (not 36)
	fmt.Println("Test 1: WHERE with AND condition")
	fmt.Println("Query: SELECT column_name, data_type, is_nullable FROM sys.columns_metadata WHERE db_name = 'ecommerce' AND table_name = 'users'")
	sql1 := "SELECT column_name, data_type, is_nullable FROM sys.columns_metadata WHERE db_name = 'ecommerce' AND table_name = 'users'"
	stmt1, err := parser.Parse(sql1)
	require.NoError(t, err)
	plan1, err := opt.Optimize(stmt1)
	require.NoError(t, err)
	result1, err := exec.Execute(plan1, sess)
	require.NoError(t, err)

	totalRows1 := 0
	for _, batch := range result1.Batches() {
		totalRows1 += int(batch.NumRows())
	}
	fmt.Printf("Expected: 5 rows (columns of users table)\n")
	fmt.Printf("Got: %d rows\n", totalRows1)
	require.Equal(t, 5, totalRows1, "Bug: Should return exactly 5 rows for users table columns")
	fmt.Println("✓ PASS: WHERE with AND works correctly")

	// Test 2: LIMIT should return at most 10 rows (not 15)
	fmt.Println("Test 2: ORDER BY with LIMIT")
	fmt.Println("Query: SELECT version, operation, db_name, table_name, file_path FROM sys.delta_log WHERE db_name = 'ecommerce' ORDER BY version DESC LIMIT 10")
	sql2 := "SELECT version, operation, db_name, table_name, file_path FROM sys.delta_log WHERE db_name = 'ecommerce' ORDER BY version DESC LIMIT 10"
	stmt2, err := parser.Parse(sql2)
	require.NoError(t, err)
	plan2, err := opt.Optimize(stmt2)
	require.NoError(t, err)
	result2, err := exec.Execute(plan2, sess)
	require.NoError(t, err)

	totalRows2 := 0
	var versions []int64
	for _, batch := range result2.Batches() {
		totalRows2 += int(batch.NumRows())
		// Get version column
		versionCol := batch.Column(0).(*array.Int64)
		for i := 0; i < int(batch.NumRows()); i++ {
			versions = append(versions, versionCol.Value(i))
		}
	}
	fmt.Printf("Expected: At most 10 rows\n")
	fmt.Printf("Got: %d rows\n", totalRows2)
	require.LessOrEqual(t, totalRows2, 10, "Bug: LIMIT 10 should return at most 10 rows")
	fmt.Println("✓ PASS: LIMIT works correctly")

	// Verify ORDER BY DESC
	for i := 1; i < len(versions); i++ {
		require.GreaterOrEqual(t, versions[i-1], versions[i], "Bug: ORDER BY DESC not working")
	}
	fmt.Println("✓ PASS: ORDER BY DESC works correctly")

	// Test 3: WHERE should filter table_files correctly
	fmt.Println("Test 3: WHERE filtering on sys.table_files")
	fmt.Println("Query: SELECT file_path, file_size, row_count, status FROM sys.table_files WHERE db_name = 'ecommerce' AND table_name = 'orders'")
	sql3 := "SELECT file_path, file_size, row_count, status FROM sys.table_files WHERE db_name = 'ecommerce' AND table_name = 'orders'"
	stmt3, err := parser.Parse(sql3)
	require.NoError(t, err)
	plan3, err := opt.Optimize(stmt3)
	require.NoError(t, err)
	result3, err := exec.Execute(plan3, sess)
	require.NoError(t, err)

	totalRows3 := 0
	allFilesAreOrders := true
	for _, batch := range result3.Batches() {
		totalRows3 += int(batch.NumRows())
		for i := 0; i < int(batch.NumRows()); i++ {
			filePath := batch.GetString(0, i)
			if !contains(filePath, "/orders/") {
				allFilesAreOrders = false
				fmt.Printf("Error: Found file not for orders table: %s\n", filePath)
			}
		}
	}
	fmt.Printf("Expected: Only files for orders table\n")
	fmt.Printf("Got: %d rows, all for orders: %v\n", totalRows3, allFilesAreOrders)
	require.Greater(t, totalRows3, 0, "Should have at least 1 file")
	require.True(t, allFilesAreOrders, "Bug: All files should be for orders table only")
	fmt.Println("✓ PASS: WHERE filtering works correctly")

	fmt.Println("========================================")
	fmt.Println("All Original Bugs FIXED!")
	fmt.Println("========================================")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

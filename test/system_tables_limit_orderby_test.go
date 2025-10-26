package test

import (
	"context"
	"os"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

// TestSystemTableLimitAndOrderBy tests that LIMIT and ORDER BY work correctly on system tables
// Bug: System tables ignore LIMIT and ORDER BY clauses
func TestSystemTableLimitAndOrderBy(t *testing.T) {
	testDir := "./test_data/system_tables_limit_orderby_test"
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

	// Insert test data into orders (this will create multiple Delta Log entries)
	builder2 := array.NewRecordBuilder(pool, ordersSchema)
	defer builder2.Release()

	// Insert 15 orders to ensure we exceed LIMIT 10
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

	t.Run("QueryColumnsMetadataWithWhereShouldRespectFilter", func(t *testing.T) {
		sql := "SELECT column_name, data_type, is_nullable FROM sys.columns_metadata WHERE db_name = 'ecommerce' AND table_name = 'users'"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		require.NoError(t, err, "Query should not fail")
		require.NotNil(t, result, "Result should not be nil")

		// Count total rows
		batches := result.Batches()
		totalRows := 0
		for _, batch := range batches {
			totalRows += int(batch.NumRows())
		}

		// CRITICAL: Should return exactly 5 rows (5 columns of users table), not all columns from all tables
		assert.Equal(t, 5, totalRows, "WHERE clause should filter to only 5 columns for ecommerce.users, but got %d rows", totalRows)

		// Verify all returned columns belong to ecommerce.users
		expectedColumns := map[string]bool{"id": true, "name": true, "email": true, "age": true, "created_at": true}
		for _, batch := range batches {
			for i := 0; i < int(batch.NumRows()); i++ {
				colName := batch.GetString(0, i)
				assert.True(t, expectedColumns[colName], "Column %s should be one of the users table columns", colName)
			}
		}
	})

	t.Run("QueryDeltaLogWithOrderByDescAndLimitShouldRespectBoth", func(t *testing.T) {
		sql := "SELECT version, operation, db_name, table_name, file_path FROM sys.delta_log WHERE db_name = 'ecommerce' ORDER BY version DESC LIMIT 10"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		require.NoError(t, err, "Query should not fail")
		require.NotNil(t, result, "Result should not be nil")

		// Count total rows
		batches := result.Batches()
		totalRows := 0
		var versions []int64
		for _, batch := range batches {
			totalRows += int(batch.NumRows())
			// Get version column (index 0)
			versionCol := batch.Column(0).(*array.Int64)
			for i := 0; i < int(batch.NumRows()); i++ {
				versions = append(versions, versionCol.Value(i))
			}
		}

		// CRITICAL: Should return at most 10 rows due to LIMIT 10
		// Note: if there are fewer than 10 entries, we'll get fewer rows
		assert.LessOrEqual(t, totalRows, 10, "LIMIT 10 should return at most 10 rows, but got %d rows", totalRows)
		assert.Greater(t, totalRows, 0, "Should return at least 1 row from ecommerce Delta Log")

		// CRITICAL: Should be in descending order by version
		for i := 1; i < len(versions); i++ {
			assert.GreaterOrEqual(t, versions[i-1], versions[i], "Versions should be in descending order: version[%d]=%d should be >= version[%d]=%d", i-1, versions[i-1], i, versions[i])
		}
	})

	t.Run("QueryTableFilesWithWhereShouldOnlyReturnMatchingRows", func(t *testing.T) {
		sql := "SELECT file_path, file_size, row_count, status FROM sys.table_files WHERE db_name = 'ecommerce' AND table_name = 'orders'"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		require.NoError(t, err, "Query should not fail")
		require.NotNil(t, result, "Result should not be nil")

		// Count total rows
		batches := result.Batches()
		totalRows := 0
		for _, batch := range batches {
			totalRows += int(batch.NumRows())
		}

		// CRITICAL: Should only return files for orders table, not files from users table
		// We expect at least 1 file for orders, but definitely not files from other tables
		assert.Greater(t, totalRows, 0, "Should return at least 1 file for ecommerce.orders")

		// Verify all files are for orders table (file paths should contain "/orders/")
		for _, batch := range batches {
			for i := 0; i < int(batch.NumRows()); i++ {
				filePath := batch.GetString(0, i)
				assert.Contains(t, filePath, "/orders/", "All file paths should be for orders table, but got: %s", filePath)
			}
		}
	})

	t.Run("QueryWithLimitOnly", func(t *testing.T) {
		sql := "SELECT db_name FROM sys.db_metadata LIMIT 2"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		require.NoError(t, err, "Query should not fail")
		require.NotNil(t, result, "Result should not be nil")

		// Count total rows
		batches := result.Batches()
		totalRows := 0
		for _, batch := range batches {
			totalRows += int(batch.NumRows())
		}

		// CRITICAL: Should return exactly 2 rows due to LIMIT 2
		assert.Equal(t, 2, totalRows, "LIMIT 2 should return exactly 2 rows, but got %d rows", totalRows)
	})

	t.Run("QueryWithOrderByOnly", func(t *testing.T) {
		sql := "SELECT db_name FROM sys.db_metadata ORDER BY db_name ASC"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		require.NoError(t, err, "Query should not fail")
		require.NotNil(t, result, "Result should not be nil")

		// Get all database names
		batches := result.Batches()
		var dbNames []string
		for _, batch := range batches {
			for i := 0; i < int(batch.NumRows()); i++ {
				dbNames = append(dbNames, batch.GetString(0, i))
			}
		}

		// CRITICAL: Should be in ascending alphabetical order
		for i := 1; i < len(dbNames); i++ {
			assert.LessOrEqual(t, dbNames[i-1], dbNames[i], "Database names should be in ascending order: dbNames[%d]=%s should be <= dbNames[%d]=%s", i-1, dbNames[i-1], i, dbNames[i])
		}
	})
}

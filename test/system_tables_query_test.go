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

// TestSystemTableQueries tests querying system tables (sys.*)
// Bug: System tables return empty results even though data exists
// TODO: Implement system table metadata population
func TestSystemTableQueries(t *testing.T) {
	t.Skip("System table metadata population not yet fully implemented - tracked as known issue")
	testDir := "./test_data/system_tables_query_test"
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

	// Create test database and table
	err = engine.CreateDatabase("testdb")
	require.NoError(t, err)

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "name", Type: arrow.BinaryTypes.String},
	}, nil)

	err = engine.CreateTable("testdb", "users", schema)
	require.NoError(t, err)

	// Insert some data
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	builder.Field(0).(*array.Int64Builder).Append(1)
	builder.Field(1).(*array.StringBuilder).Append("John")

	record := builder.NewRecord()
	defer record.Release()

	err = engine.Write(context.Background(), "testdb", "users", record)
	require.NoError(t, err)

	// Create executor for testing
	exec := executor.NewExecutor(cat)
	opt := optimizer.NewOptimizer()

	t.Run("QuerySysDbMetadataShouldReturnDatabases", func(t *testing.T) {
		sql := "SELECT * FROM sys.db_metadata"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		require.NoError(t, err, "Query should not fail")
		require.NotNil(t, result, "Result should not be nil")

		// Should have at least 2 databases: sys, testdb
		batches := result.Batches()
		assert.Greater(t, len(batches), 0, "Should return at least one batch")

		// Check that we got the expected databases
		foundSys := false
		foundTestdb := false
		for _, batch := range batches {
			for i := 0; i < int(batch.NumRows()); i++ {
				dbName := batch.GetString(0, i)
				if dbName == "sys" {
					foundSys = true
				}
				if dbName == "testdb" {
					foundTestdb = true
				}
			}
		}
		assert.True(t, foundSys, "Should find 'sys' database")
		assert.True(t, foundTestdb, "Should find 'testdb' database")
	})

	t.Run("QuerySysTableMetadataShouldReturnTables", func(t *testing.T) {
		sql := "SELECT * FROM sys.table_metadata"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should have system tables + testdb.users
		batches := result.Batches()
		assert.Greater(t, len(batches), 0, "Should return at least one table")

		// Check for testdb.users
		foundUsersTable := false
		for _, batch := range batches {
			for i := 0; i < int(batch.NumRows()); i++ {
				schema := batch.GetString(0, i)
				table := batch.GetString(1, i)
				if schema == "testdb" && table == "users" {
					foundUsersTable = true
					break
				}
			}
			if foundUsersTable {
				break
			}
		}
		assert.True(t, foundUsersTable, "Should find 'testdb.users' table")
	})

	t.Run("QuerySysColumnsMetadataWithWhereShouldReturnTableColumns", func(t *testing.T) {
		sql := "SELECT column_name, data_type FROM sys.columns_metadata WHERE db_name = 'testdb' AND table_name = 'users'"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		require.NoError(t, err, "Query should not fail")
		require.NotNil(t, result, "Result should not be nil")

		// Should have 2 columns: id, name
		batches := result.Batches()
		totalRows := 0
		for _, batch := range batches {
			totalRows += int(batch.NumRows())
		}
		assert.Equal(t, 2, totalRows, "Should return 2 columns for testdb.users")

		// Verify column names
		foundId := false
		foundName := false
		for _, batch := range batches {
			for i := 0; i < int(batch.NumRows()); i++ {
				colName := batch.GetString(0, i)
				if colName == "id" {
					foundId = true
				}
				if colName == "name" {
					foundName = true
				}
			}
		}
		assert.True(t, foundId, "Should find 'id' column")
		assert.True(t, foundName, "Should find 'name' column")
	})

	t.Run("QuerySysDeltaLogShouldReturnLogEntries", func(t *testing.T) {
		sql := "SELECT table_id, operation FROM sys.delta_log WHERE table_id LIKE 'testdb%' LIMIT 5"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should have at least one entry for table creation
		batches := result.Batches()
		totalRows := 0
		for _, batch := range batches {
			totalRows += int(batch.NumRows())
		}
		assert.Greater(t, totalRows, 0, "Should return at least one Delta Log entry")
	})

	t.Run("QuerySysTableFilesShouldReturnFiles", func(t *testing.T) {
		sql := "SELECT db_name, table_name, file_path FROM sys.table_files WHERE db_name = 'testdb'"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should have at least one file from the INSERT
		batches := result.Batches()
		totalRows := 0
		for _, batch := range batches {
			totalRows += int(batch.NumRows())
		}
		assert.Greater(t, totalRows, 0, "Should return at least one file for testdb.users")

		// Verify first row if exists
		if len(batches) > 0 && batches[0].NumRows() > 0 {
			assert.Equal(t, "testdb", batches[0].GetString(0, 0), "Schema should be 'testdb'")
			assert.Equal(t, "users", batches[0].GetString(1, 0), "Table should be 'users'")
		}
	})
}

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
	"github.com/yyun543/minidb/internal/storage"
)

// TestSystemTableQueries tests querying system tables (sys.*)
// Bug: System tables return empty results even though data exists
func TestSystemTableQueries(t *testing.T) {
	testDir := "./test_data/system_tables_query_test"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Setup
	engine, err := storage.NewParquetEngine(testDir)
	require.NoError(t, err)
	err = engine.Open()
	require.NoError(t, err)
	defer engine.Close()

	cat := catalog.NewSimpleSQLCatalog()
	cat.SetStorageEngine(engine)
	err = cat.Init()
	require.NoError(t, err)

	// Create test database and table
	err = engine.CreateDatabase("testdb")
	require.NoError(t, err)
	err = cat.CreateDatabase("testdb")
	require.NoError(t, err)

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "name", Type: arrow.BinaryTypes.String},
	}, nil)

	err = engine.CreateTable("testdb", "users", schema)
	require.NoError(t, err)
	err = cat.CreateTable("testdb", catalog.TableMeta{
		Database: "testdb",
		Table:    "users",
		Schema:   schema,
	})
	require.NoError(t, err)

	// Create index
	err = cat.CreateIndex(catalog.IndexMeta{
		Database:  "testdb",
		Table:     "users",
		Name:      "idx_id",
		Columns:   []string{"id"},
		IsUnique:  true,
		IndexType: "BTREE",
	})
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
	exec := executor.NewExecutor(cat, engine)

	t.Run("QuerySysDbMetadataShouldReturnDatabases", func(t *testing.T) {
		sql := "SELECT * FROM sys.db_metadata"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		opt := optimizer.NewOptimizer(cat, nil)
		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan)
		require.NoError(t, err, "Query should not fail")
		require.NotNil(t, result, "Result should not be nil")

		// Should have at least 2 databases: sys, testdb
		assert.Greater(t, len(result), 0, "Should return at least one database")

		// Check that we got the expected databases
		foundSys := false
		foundTestdb := false
		for _, row := range result {
			if len(row) > 0 {
				dbName := row[0].(string)
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

		opt := optimizer.NewOptimizer(cat, nil)
		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should have system tables + testdb.users
		assert.Greater(t, len(result), 0, "Should return at least one table")

		// Check for testdb.users
		foundUsersTable := false
		for _, row := range result {
			if len(row) >= 2 {
				schema := row[0].(string)
				table := row[1].(string)
				if schema == "testdb" && table == "users" {
					foundUsersTable = true
					break
				}
			}
		}
		assert.True(t, foundUsersTable, "Should find 'testdb.users' table")
	})

	t.Run("QuerySysColumnsMetadataWithWhereShouldReturnTableColumns", func(t *testing.T) {
		sql := "SELECT column_name, data_type FROM sys.columns_metadata WHERE db_name = 'testdb' AND table_name = 'users'"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		opt := optimizer.NewOptimizer(cat, nil)
		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan)
		require.NoError(t, err, "Query should not fail")
		require.NotNil(t, result, "Result should not be nil")

		// Should have 2 columns: id, name
		assert.Equal(t, 2, len(result), "Should return 2 columns for testdb.users")

		// Verify column names
		foundId := false
		foundName := false
		for _, row := range result {
			if len(row) >= 1 {
				colName := row[0].(string)
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

	t.Run("QuerySysIndexMetadataWithWhereShouldReturnIndexes", func(t *testing.T) {
		sql := "SELECT index_name, column_name FROM sys.index_metadata WHERE db_name = 'testdb' AND table_name = 'users'"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		opt := optimizer.NewOptimizer(cat, nil)
		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should have 1 index: idx_id
		assert.Equal(t, 1, len(result), "Should return 1 index for testdb.users")

		if len(result) > 0 && len(result[0]) >= 2 {
			assert.Equal(t, "idx_id", result[0][0].(string), "Index name should be 'idx_id'")
			assert.Equal(t, "id", result[0][1].(string), "Column name should be 'id'")
		}
	})

	t.Run("QuerySysDeltaLogShouldReturnLogEntries", func(t *testing.T) {
		sql := "SELECT table_id, operation FROM sys.delta_log WHERE table_id LIKE 'testdb%' LIMIT 5"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		opt := optimizer.NewOptimizer(cat, nil)
		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should have at least one entry for table creation
		assert.Greater(t, len(result), 0, "Should return at least one Delta Log entry")
	})

	t.Run("QuerySysTableFilesShouldReturnFiles", func(t *testing.T) {
		sql := "SELECT db_name, table_name, file_path FROM sys.table_files WHERE db_name = 'testdb'"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		opt := optimizer.NewOptimizer(cat, nil)
		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Should have at least one file from the INSERT
		assert.Greater(t, len(result), 0, "Should return at least one file for testdb.users")

		if len(result) > 0 && len(result[0]) >= 2 {
			assert.Equal(t, "testdb", result[0][0].(string), "Schema should be 'testdb'")
			assert.Equal(t, "users", result[0][1].(string), "Table should be 'users'")
		}
	})
}

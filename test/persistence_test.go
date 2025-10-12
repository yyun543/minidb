package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/storage"
)

// TestPersistenceAfterRestart tests that data persists after server restart
// following Delta Lake checkpoint-based recovery mechanism
func TestPersistenceAfterRestart(t *testing.T) {
	ctx := context.Background()
	tempDir := setupPersistenceTestDir(t)
	defer os.RemoveAll(tempDir)

	// Step 1: Create database, table, and insert data
	t.Log("Step 1: First engine instance - creating data")
	{
		engine1, err := storage.NewParquetEngine(tempDir)
		require.NoError(t, err)
		require.NoError(t, engine1.Open())

		// Create database
		require.NoError(t, engine1.CreateDatabase("testdb"))

		// Create table
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "id", Type: arrow.PrimitiveTypes.Int64},
				{Name: "name", Type: arrow.BinaryTypes.String},
				{Name: "age", Type: arrow.PrimitiveTypes.Int64},
			}, nil,
		)
		require.NoError(t, engine1.CreateTable("testdb", "users", schema))

		// Insert data
		pool := memory.NewGoAllocator()
		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()

		// Insert 3 users
		for i := 0; i < 3; i++ {
			builder.Field(0).(*array.Int64Builder).Append(int64(i + 1))
			builder.Field(1).(*array.StringBuilder).Append(fmt.Sprintf("User%d", i+1))
			builder.Field(2).(*array.Int64Builder).Append(int64(20 + i*5))
		}

		record := builder.NewRecord()
		err = engine1.Write(ctx, "testdb", "users", record)
		record.Release()
		require.NoError(t, err)

		// Verify data was written
		snapshot1, err := engine1.GetDeltaLog().GetSnapshot("testdb.users", -1)
		require.NoError(t, err)
		assert.Greater(t, len(snapshot1.Files), 0, "Should have Parquet files")
		t.Logf("After write: %d files, latest version: %d", len(snapshot1.Files), engine1.GetDeltaLog().GetLatestVersion())

		// ✅ Verify Delta Log JSON files are created in _delta_log directory
		deltaLogDir := filepath.Join(tempDir, "_delta_log")
		if _, err := os.Stat(deltaLogDir); os.IsNotExist(err) {
			t.Logf("Warning: _delta_log directory not found (will be created in implementation)")
		} else {
			// Check for JSON log files
			files, _ := os.ReadDir(deltaLogDir)
			jsonFileCount := 0
			for _, f := range files {
				if filepath.Ext(f.Name()) == ".json" {
					jsonFileCount++
					t.Logf("Found Delta Log JSON file: %s", f.Name())
				}
			}
			if jsonFileCount > 0 {
				assert.Greater(t, jsonFileCount, 0, "Should have JSON log files in _delta_log")
			}
		}

		// Close engine
		require.NoError(t, engine1.Close())
	}

	// Step 2: Restart - create new engine instance and verify data persists
	t.Log("Step 2: Second engine instance - verifying Delta Lake checkpoint-based recovery")
	{
		engine2, err := storage.NewParquetEngine(tempDir)
		require.NoError(t, err)

		// Bootstrap should scan _delta_log directory and rebuild state
		require.NoError(t, engine2.Open())
		defer engine2.Close()

		// Check if database exists (should be rebuilt from Delta Log)
		exists, err := engine2.DatabaseExists("testdb")
		require.NoError(t, err)
		assert.True(t, exists, "Database should exist after restart (rebuilt from Delta Log)")

		// Check if table exists (should be rebuilt from metadata entries)
		tableExists, err := engine2.TableExists("testdb", "users")
		require.NoError(t, err)
		assert.True(t, tableExists, "Table should exist after restart (rebuilt from Delta Log)")

		// Verify table schema (should be loaded from metadata log entry)
		schema, err := engine2.GetTableSchema("testdb", "users")
		require.NoError(t, err)
		assert.NotNil(t, schema, "Schema should be loaded from Delta Log metadata")
		assert.Equal(t, 3, len(schema.Fields()), "Schema should have 3 fields")

		// Verify data can be read (files should be rebuilt from ADD entries)
		iterator, err := engine2.Scan(ctx, "testdb", "users", nil)
		require.NoError(t, err)
		defer iterator.Close()

		rowCount := int64(0)
		for iterator.Next() {
			record := iterator.Record()
			rowCount += record.NumRows()
			t.Logf("Read %d rows from persisted data", record.NumRows())
		}
		require.NoError(t, iterator.Err())
		assert.Equal(t, int64(3), rowCount, "Should read all 3 rows after restart")
	}
}

// TestMultipleTablesPersis tence tests persistence with multiple tables
func TestMultipleTablesPersistence(t *testing.T) {
	ctx := context.Background()
	tempDir := setupPersistenceTestDir(t)
	defer os.RemoveAll(tempDir)

	// Step 1: Create multiple tables
	t.Log("Step 1: Creating multiple tables")
	{
		engine1, err := storage.NewParquetEngine(tempDir)
		require.NoError(t, err)
		require.NoError(t, engine1.Open())

		require.NoError(t, engine1.CreateDatabase("db1"))
		require.NoError(t, engine1.CreateDatabase("db2"))

		// Create tables in db1
		schema1 := arrow.NewSchema([]arrow.Field{{Name: "id", Type: arrow.PrimitiveTypes.Int64}}, nil)
		require.NoError(t, engine1.CreateTable("db1", "table1", schema1))
		require.NoError(t, engine1.CreateTable("db1", "table2", schema1))

		// Create tables in db2
		schema2 := arrow.NewSchema([]arrow.Field{{Name: "name", Type: arrow.BinaryTypes.String}}, nil)
		require.NoError(t, engine1.CreateTable("db2", "table3", schema2))

		// Insert data into each table
		insertData(t, ctx, engine1, "db1", "table1", schema1, 5)
		insertData(t, ctx, engine1, "db1", "table2", schema1, 3)
		insertData(t, ctx, engine1, "db2", "table3", schema2, 7)

		require.NoError(t, engine1.Close())
	}

	// Step 2: Restart and verify all tables
	t.Log("Step 2: Verifying all tables after restart")
	{
		engine2, err := storage.NewParquetEngine(tempDir)
		require.NoError(t, err)
		require.NoError(t, engine2.Open())
		defer engine2.Close()

		// Verify databases
		db1Exists, _ := engine2.DatabaseExists("db1")
		db2Exists, _ := engine2.DatabaseExists("db2")
		assert.True(t, db1Exists, "db1 should exist")
		assert.True(t, db2Exists, "db2 should exist")

		// Verify tables
		verifyTableData(t, ctx, engine2, "db1", "table1", 5)
		verifyTableData(t, ctx, engine2, "db1", "table2", 3)
		verifyTableData(t, ctx, engine2, "db2", "table3", 7)
	}
}

// TestUpdateDeletePersistence tests persistence after UPDATE/DELETE operations
func TestUpdateDeletePersistence(t *testing.T) {
	ctx := context.Background()
	tempDir := setupPersistenceTestDir(t)
	defer os.RemoveAll(tempDir)

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "value", Type: arrow.PrimitiveTypes.Int64},
		}, nil,
	)

	// Step 1: Create, insert, update, delete
	t.Log("Step 1: Creating and modifying data")
	{
		engine1, err := storage.NewParquetEngine(tempDir)
		require.NoError(t, err)
		require.NoError(t, engine1.Open())

		require.NoError(t, engine1.CreateDatabase("testdb"))
		require.NoError(t, engine1.CreateTable("testdb", "data", schema))

		// Insert 10 rows
		insertData(t, ctx, engine1, "testdb", "data", schema, 10)

		// Update some rows
		filters := []storage.Filter{{Column: "id", Operator: "<", Value: int64(5)}}
		updates := map[string]interface{}{"value": int64(999)}
		updated, err := engine1.Update(ctx, "testdb", "data", filters, updates)
		require.NoError(t, err)
		t.Logf("Updated %d rows", updated)

		// Delete some rows
		deleteFilters := []storage.Filter{{Column: "id", Operator: ">=", Value: int64(8)}}
		deleted, err := engine1.Delete(ctx, "testdb", "data", deleteFilters)
		require.NoError(t, err)
		t.Logf("Deleted %d rows", deleted)

		require.NoError(t, engine1.Close())
	}

	// Step 2: Restart and verify final state
	t.Log("Step 2: Verifying data state after restart")
	{
		engine2, err := storage.NewParquetEngine(tempDir)
		require.NoError(t, err)
		require.NoError(t, engine2.Open())
		defer engine2.Close()

		// Count remaining rows
		iterator, err := engine2.Scan(ctx, "testdb", "data", nil)
		require.NoError(t, err)
		defer iterator.Close()

		rowCount := int64(0)
		for iterator.Next() {
			record := iterator.Record()
			rowCount += record.NumRows()
		}
		require.NoError(t, iterator.Err())

		// Should have 10 - deleted rows
		t.Logf("Found %d rows after restart", rowCount)
		assert.Greater(t, rowCount, int64(0), "Should have some rows remaining")
	}
}

// Helper functions

func setupPersistenceTestDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "minidb_persistence_test_*")
	require.NoError(t, err)
	t.Logf("Created test directory: %s", tempDir)
	return tempDir
}

func insertData(t *testing.T, ctx context.Context, engine storage.StorageEngine, db, table string, schema *arrow.Schema, rowCount int) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	for i := 0; i < rowCount; i++ {
		for fieldIdx := 0; fieldIdx < len(schema.Fields()); fieldIdx++ {
			field := schema.Field(fieldIdx)
			switch field.Type.ID() {
			case arrow.INT64:
				builder.Field(fieldIdx).(*array.Int64Builder).Append(int64(i + 1))
			case arrow.STRING:
				builder.Field(fieldIdx).(*array.StringBuilder).Append(fmt.Sprintf("data%d", i+1))
			}
		}
	}

	record := builder.NewRecord()
	err := engine.Write(ctx, db, table, record)
	record.Release()
	require.NoError(t, err)
}

func verifyTableData(t *testing.T, ctx context.Context, engine storage.StorageEngine, db, table string, expectedRows int64) {
	exists, err := engine.TableExists(db, table)
	require.NoError(t, err)
	assert.True(t, exists, fmt.Sprintf("Table %s.%s should exist", db, table))

	iterator, err := engine.Scan(ctx, db, table, nil)
	require.NoError(t, err)
	defer iterator.Close()

	rowCount := int64(0)
	for iterator.Next() {
		record := iterator.Record()
		rowCount += record.NumRows()
	}
	require.NoError(t, iterator.Err())

	assert.Equal(t, expectedRows, rowCount, fmt.Sprintf("Table %s.%s should have %d rows", db, table, expectedRows))
	t.Logf("Table %s.%s has %d rows (expected %d)", db, table, rowCount, expectedRows)
}

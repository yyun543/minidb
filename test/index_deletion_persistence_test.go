package test

import (
	"os"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/storage"
)

// TestIndexDeletionPersistence is a critical regression test for index deletion persistence.
// This test verifies that when an index is dropped, the deletion is persisted to Delta Log
// and the index does NOT reappear after server restart.
//
// Bug Report: After creating and then deleting an index, the index reappears after server
// restart, showing that DropIndex does not persist the deletion to Delta Log.
func TestIndexDeletionPersistence(t *testing.T) {
	testDir := "./test_data/index_deletion_persistence_test"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Phase 1: Create database, table, and indexes, then delete one index
	t.Log("Phase 1: Creating indexes and deleting one")
	{
		engine, err := storage.NewParquetEngine(testDir)
		require.NoError(t, err, "Failed to create storage engine")

		err = engine.Open()
		require.NoError(t, err, "Failed to open storage engine")

		cat := catalog.NewSimpleSQLCatalog()
		cat.SetStorageEngine(engine)
		err = cat.Init()
		require.NoError(t, err, "Failed to initialize catalog")

		// Create database
		err = engine.CreateDatabase("testdb")
		require.NoError(t, err, "Failed to create database")
		err = cat.CreateDatabase("testdb")
		require.NoError(t, err, "Failed to register database")

		// Create table
		schema := arrow.NewSchema([]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "name", Type: arrow.BinaryTypes.String},
		}, nil)

		err = engine.CreateTable("testdb", "users", schema)
		require.NoError(t, err, "Failed to create table")
		err = cat.CreateTable("testdb", catalog.TableMeta{
			Database: "testdb",
			Table:    "users",
			Schema:   schema,
		})
		require.NoError(t, err, "Failed to register table")

		// Create 3 indexes
		indexes := []catalog.IndexMeta{
			{Database: "testdb", Table: "users", Name: "idx_id", Columns: []string{"id"}, IsUnique: false, IndexType: "BTREE"},
			{Database: "testdb", Table: "users", Name: "idx_name", Columns: []string{"name"}, IsUnique: false, IndexType: "BTREE"},
			{Database: "testdb", Table: "users", Name: "idx_to_delete", Columns: []string{"id"}, IsUnique: false, IndexType: "BTREE"},
		}

		for _, idx := range indexes {
			err = cat.CreateIndex(idx)
			require.NoError(t, err, "Failed to create index %s", idx.Name)
			t.Logf("✓ Created index: %s", idx.Name)
		}

		// Verify all 3 indexes exist
		allIndexes, err := cat.GetAllIndexes("testdb", "users")
		require.NoError(t, err)
		assert.Equal(t, 3, len(allIndexes), "Should have 3 indexes before deletion")

		// Delete one index
		err = cat.DropIndex("testdb", "users", "idx_to_delete")
		require.NoError(t, err, "Failed to drop index")
		t.Log("✓ Deleted index: idx_to_delete")

		// Verify only 2 indexes remain
		allIndexes, err = cat.GetAllIndexes("testdb", "users")
		require.NoError(t, err)
		assert.Equal(t, 2, len(allIndexes), "Should have 2 indexes after deletion")

		// Verify the deleted index cannot be retrieved
		_, err = cat.GetIndex("testdb", "users", "idx_to_delete")
		assert.Error(t, err, "Deleted index should not exist")

		err = engine.Close()
		require.NoError(t, err)
		t.Log("Phase 1 completed: Created 3 indexes, deleted 1, verified 2 remain")
	}

	// Phase 2: Restart and verify the deleted index does NOT reappear
	t.Log("Phase 2: Restarting and verifying deleted index stays deleted")
	{
		engine, err := storage.NewParquetEngine(testDir)
		require.NoError(t, err)

		err = engine.Open()
		require.NoError(t, err)
		defer engine.Close()

		cat := catalog.NewSimpleSQLCatalog()
		cat.SetStorageEngine(engine)
		err = cat.Init()
		require.NoError(t, err)

		// CRITICAL: Verify only 2 indexes exist after restart
		allIndexes, err := cat.GetAllIndexes("testdb", "users")
		require.NoError(t, err)

		if len(allIndexes) != 2 {
			t.Errorf("FAILED: Expected 2 indexes after restart, got %d", len(allIndexes))
			for _, idx := range allIndexes {
				t.Logf("  Found index: %s", idx.Name)
			}
			t.Fatal("BUG CONFIRMED: Deleted index reappeared after restart!")
		}

		// Verify the deleted index is really gone
		_, err = cat.GetIndex("testdb", "users", "idx_to_delete")
		if err == nil {
			t.Fatal("FAILED: Deleted index 'idx_to_delete' still exists after restart!")
		}

		// Verify the remaining indexes are correct
		_, err = cat.GetIndex("testdb", "users", "idx_id")
		require.NoError(t, err, "idx_id should still exist")

		_, err = cat.GetIndex("testdb", "users", "idx_name")
		require.NoError(t, err, "idx_name should still exist")

		t.Log("✓ SUCCESS: Deleted index did not reappear after restart")
		t.Log("Phase 2 completed")
	}

	t.Log("========================================")
	t.Log("✓ Index Deletion Persistence Test PASSED")
	t.Log("========================================")
}

// TestTableDeletionPersistence verifies that table deletion is persisted correctly
func TestTableDeletionPersistence(t *testing.T) {
	testDir := "./test_data/table_deletion_persistence_test"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Phase 1: Create tables and delete one
	t.Log("Phase 1: Creating tables and deleting one")
	{
		engine, err := storage.NewParquetEngine(testDir)
		require.NoError(t, err)

		err = engine.Open()
		require.NoError(t, err)

		cat := catalog.NewSimpleSQLCatalog()
		cat.SetStorageEngine(engine)
		err = cat.Init()
		require.NoError(t, err)

		// Create database
		err = engine.CreateDatabase("testdb")
		require.NoError(t, err)
		err = cat.CreateDatabase("testdb")
		require.NoError(t, err)

		// Create 3 tables
		schema := arrow.NewSchema([]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		}, nil)

		tables := []string{"table1", "table2", "table_to_delete"}
		for _, tableName := range tables {
			err = engine.CreateTable("testdb", tableName, schema)
			require.NoError(t, err)
			err = cat.CreateTable("testdb", catalog.TableMeta{
				Database: "testdb",
				Table:    tableName,
				Schema:   schema,
			})
			require.NoError(t, err)
			t.Logf("✓ Created table: %s", tableName)
		}

		// Verify 3 tables exist
		allTables, err := cat.GetAllTables("testdb")
		require.NoError(t, err)
		assert.Equal(t, 3, len(allTables), "Should have 3 tables before deletion")

		// Delete one table
		err = cat.DropTable("testdb", "table_to_delete")
		require.NoError(t, err)
		t.Log("✓ Deleted table: table_to_delete")

		// Verify only 2 tables remain
		allTables, err = cat.GetAllTables("testdb")
		require.NoError(t, err)
		assert.Equal(t, 2, len(allTables), "Should have 2 tables after deletion")

		err = engine.Close()
		require.NoError(t, err)
		t.Log("Phase 1 completed")
	}

	// Phase 2: Restart and verify deleted table stays deleted
	t.Log("Phase 2: Restarting and verifying deleted table stays deleted")
	{
		engine, err := storage.NewParquetEngine(testDir)
		require.NoError(t, err)

		err = engine.Open()
		require.NoError(t, err)
		defer engine.Close()

		cat := catalog.NewSimpleSQLCatalog()
		cat.SetStorageEngine(engine)
		err = cat.Init()
		require.NoError(t, err)

		// Verify only 2 tables exist
		allTables, err := cat.GetAllTables("testdb")
		require.NoError(t, err)

		if len(allTables) != 2 {
			t.Errorf("FAILED: Expected 2 tables after restart, got %d", len(allTables))
			for _, tbl := range allTables {
				t.Logf("  Found table: %s", tbl)
			}
		}
		assert.Equal(t, 2, len(allTables), "Should have 2 tables after restart")

		// Verify deleted table is gone
		_, err = cat.GetTable("testdb", "table_to_delete")
		assert.Error(t, err, "Deleted table should not exist")

		t.Log("✓ SUCCESS: Deleted table did not reappear")
		t.Log("Phase 2 completed")
	}

	t.Log("========================================")
	t.Log("✓ Table Deletion Persistence Test PASSED")
	t.Log("========================================")
}

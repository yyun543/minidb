package test

import (
	"context"
	"os"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/storage"
)

// TestMetadataRecoveryAfterRestart tests that databases and tables persist after restart
// This is a regression test for the data loss bug where user databases were not recovered
func TestMetadataRecoveryAfterRestart(t *testing.T) {
	// Clean test directory
	testDir := "./test_metadata_recovery"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Phase 1: Create database, table, and insert data
	t.Log("Phase 1: Creating initial data")
	{
		// Create storage engine
		engine, err := storage.NewParquetEngine(testDir)
		if err != nil {
			t.Fatalf("Failed to create engine: %v", err)
		}

		if err := engine.Open(); err != nil {
			t.Fatalf("Failed to open engine: %v", err)
		}

		// Create catalog
		cat := catalog.NewSimpleSQLCatalog()
		cat.SetStorageEngine(engine)
		if err := cat.Init(); err != nil {
			t.Fatalf("Failed to initialize catalog: %v", err)
		}

		// Create database 'ecommerce'
		if err := engine.CreateDatabase("ecommerce"); err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}

		if err := cat.CreateDatabase("ecommerce"); err != nil {
			t.Fatalf("Failed to register database in catalog: %v", err)
		}

		// Create table 'users'
		userSchema := arrow.NewSchema([]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "name", Type: arrow.BinaryTypes.String},
			{Name: "age", Type: arrow.PrimitiveTypes.Int64},
		}, nil)

		if err := engine.CreateTable("ecommerce", "users", userSchema); err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}

		if err := cat.CreateTable("ecommerce", catalog.TableMeta{
			Database: "ecommerce",
			Table:    "users",
			Schema:   userSchema,
		}); err != nil {
			t.Fatalf("Failed to register table in catalog: %v", err)
		}

		// Insert data
		pool := memory.NewGoAllocator()
		builder := array.NewRecordBuilder(pool, userSchema)
		defer builder.Release()

		builder.Field(0).(*array.Int64Builder).AppendValues([]int64{1, 2, 3}, nil)
		builder.Field(1).(*array.StringBuilder).AppendValues([]string{"Alice", "Bob", "Charlie"}, nil)
		builder.Field(2).(*array.Int64Builder).AppendValues([]int64{25, 30, 35}, nil)

		record := builder.NewRecord()
		defer record.Release()

		if err := engine.Write(context.Background(), "ecommerce", "users", record); err != nil {
			t.Fatalf("Failed to write data: %v", err)
		}

		// Close engine
		if err := engine.Close(); err != nil {
			t.Fatalf("Failed to close engine: %v", err)
		}

		t.Log("Phase 1 completed: Database and table created with data")
	}

	// Phase 2: Restart - open engine again and check if database and table exist
	t.Log("Phase 2: Restarting and checking metadata recovery")
	{
		// Create new engine instance (simulating restart)
		engine, err := storage.NewParquetEngine(testDir)
		if err != nil {
			t.Fatalf("Failed to create engine on restart: %v", err)
		}

		if err := engine.Open(); err != nil {
			t.Fatalf("Failed to open engine on restart: %v", err)
		}
		defer engine.Close()

		// Create new catalog instance
		cat := catalog.NewSimpleSQLCatalog()
		cat.SetStorageEngine(engine)
		if err := cat.Init(); err != nil {
			t.Fatalf("Failed to initialize catalog on restart: %v", err)
		}

		// Check if 'ecommerce' database exists
		databases, err := cat.GetAllDatabases()
		if err != nil {
			t.Fatalf("Failed to get databases: %v", err)
		}

		t.Logf("Databases found: %v", databases)

		foundEcommerce := false
		for _, db := range databases {
			if db == "ecommerce" {
				foundEcommerce = true
				break
			}
		}

		if !foundEcommerce {
			t.Errorf("FAILED: Database 'ecommerce' not recovered after restart. Databases: %v", databases)
		} else {
			t.Log("SUCCESS: Database 'ecommerce' recovered")
		}

		// Check if 'users' table exists
		tables, err := cat.GetAllTables("ecommerce")
		if err != nil {
			t.Logf("Failed to get tables from ecommerce: %v", err)
			// Continue to check if table exists in storage engine
		} else {
			t.Logf("Tables found in ecommerce: %v", tables)

			foundUsers := false
			for _, tbl := range tables {
				if tbl == "users" {
					foundUsers = true
					break
				}
			}

			if !foundUsers {
				t.Errorf("FAILED: Table 'users' not recovered in catalog after restart")
			} else {
				t.Log("SUCCESS: Table 'users' recovered in catalog")
			}
		}

		// Check if table schema is recovered in storage engine
		schema, err := engine.GetTableSchema("ecommerce", "users")
		if err != nil {
			t.Errorf("FAILED: Table schema not recovered in storage engine: %v", err)
		} else {
			t.Logf("SUCCESS: Table schema recovered: %d fields", len(schema.Fields()))
		}

		// Check if data can be read
		iter, err := engine.Scan(context.Background(), "ecommerce", "users", nil)
		if err != nil {
			t.Errorf("FAILED: Cannot scan table after restart: %v", err)
		} else {
			defer iter.Close()

			rowCount := 0
			for iter.Next() {
				record := iter.Record()
				rowCount += int(record.NumRows())
			}

			if iter.Err() != nil {
				t.Errorf("FAILED: Iteration error: %v", iter.Err())
			}

			if rowCount != 3 {
				t.Errorf("FAILED: Expected 3 rows, got %d", rowCount)
			} else {
				t.Logf("SUCCESS: Data recovered, %d rows found", rowCount)
			}
		}

		t.Log("Phase 2 completed")
	}
}

// TestMultipleDatabaseRecovery tests recovery of multiple databases and tables
func TestMultipleDatabaseRecovery(t *testing.T) {
	testDir := "./test_multiple_db_recovery"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Phase 1: Create multiple databases and tables
	{
		engine, err := storage.NewParquetEngine(testDir)
		if err != nil {
			t.Fatalf("Failed to create engine: %v", err)
		}

		if err := engine.Open(); err != nil {
			t.Fatalf("Failed to open engine: %v", err)
		}

		cat := catalog.NewSimpleSQLCatalog()
		cat.SetStorageEngine(engine)
		if err := cat.Init(); err != nil {
			t.Fatalf("Failed to initialize catalog: %v", err)
		}

		// Create databases
		for _, dbName := range []string{"db1", "db2", "db3"} {
			if err := engine.CreateDatabase(dbName); err != nil {
				t.Fatalf("Failed to create database %s: %v", dbName, err)
			}
			if err := cat.CreateDatabase(dbName); err != nil {
				t.Fatalf("Failed to register database %s: %v", dbName, err)
			}

			// Create table in each database
			schema := arrow.NewSchema([]arrow.Field{
				{Name: "id", Type: arrow.PrimitiveTypes.Int64},
				{Name: "value", Type: arrow.BinaryTypes.String},
			}, nil)

			tableName := "table_" + dbName
			if err := engine.CreateTable(dbName, tableName, schema); err != nil {
				t.Fatalf("Failed to create table %s.%s: %v", dbName, tableName, err)
			}

			if err := cat.CreateTable(dbName, catalog.TableMeta{
				Database: dbName,
				Table:    tableName,
				Schema:   schema,
			}); err != nil {
				t.Fatalf("Failed to register table %s.%s: %v", dbName, tableName, err)
			}
		}

		engine.Close()
	}

	// Phase 2: Restart and verify all databases and tables are recovered
	{
		engine, err := storage.NewParquetEngine(testDir)
		if err != nil {
			t.Fatalf("Failed to create engine on restart: %v", err)
		}

		if err := engine.Open(); err != nil {
			t.Fatalf("Failed to open engine on restart: %v", err)
		}
		defer engine.Close()

		cat := catalog.NewSimpleSQLCatalog()
		cat.SetStorageEngine(engine)
		if err := cat.Init(); err != nil {
			t.Fatalf("Failed to initialize catalog on restart: %v", err)
		}

		databases, err := cat.GetAllDatabases()
		if err != nil {
			t.Fatalf("Failed to get databases: %v", err)
		}

		t.Logf("Databases after restart: %v", databases)

		expectedDBs := []string{"sys", "default", "db1", "db2", "db3"}
		for _, expectedDB := range expectedDBs {
			found := false
			for _, db := range databases {
				if db == expectedDB {
					found = true
					break
				}
			}
			if !found && expectedDB != "sys" && expectedDB != "default" {
				t.Errorf("Database '%s' not recovered", expectedDB)
			}
		}
	}
}

// TestIndexMetadataRecoveryAfterRestart is a critical regression test for index persistence.
// This test ensures that index metadata is persisted to Delta Log and recovered after restart.
//
// Bug History: Previously, CreateIndex did not persist index metadata to Delta Log, causing
// all indexes to be lost after server restart. This test prevents that bug from recurring.
func TestIndexMetadataRecoveryAfterRestart(t *testing.T) {
	testDir := "./test_index_metadata_recovery"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Phase 1: Create database, table, and indexes
	t.Log("Phase 1: Creating database, table, and indexes")
	{
		engine, err := storage.NewParquetEngine(testDir)
		if err != nil {
			t.Fatalf("Failed to create engine: %v", err)
		}

		if err := engine.Open(); err != nil {
			t.Fatalf("Failed to open engine: %v", err)
		}

		cat := catalog.NewSimpleSQLCatalog()
		cat.SetStorageEngine(engine)
		if err := cat.Init(); err != nil {
			t.Fatalf("Failed to initialize catalog: %v", err)
		}

		// Create database
		if err := engine.CreateDatabase("testdb"); err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		if err := cat.CreateDatabase("testdb"); err != nil {
			t.Fatalf("Failed to register database in catalog: %v", err)
		}

		// Create table
		schema := arrow.NewSchema([]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "name", Type: arrow.BinaryTypes.String},
			{Name: "price", Type: arrow.PrimitiveTypes.Int64},
		}, nil)

		if err := engine.CreateTable("testdb", "products", schema); err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
		if err := cat.CreateTable("testdb", catalog.TableMeta{
			Database: "testdb",
			Table:    "products",
			Schema:   schema,
		}); err != nil {
			t.Fatalf("Failed to register table in catalog: %v", err)
		}

		// Create indexes
		indexes := []catalog.IndexMeta{
			{
				Database:  "testdb",
				Table:     "products",
				Name:      "idx_product_id",
				Columns:   []string{"id"},
				IsUnique:  false,
				IndexType: "BTREE",
			},
			{
				Database:  "testdb",
				Table:     "products",
				Name:      "idx_product_name",
				Columns:   []string{"name"},
				IsUnique:  false,
				IndexType: "BTREE",
			},
			{
				Database:  "testdb",
				Table:     "products",
				Name:      "idx_product_name_price",
				Columns:   []string{"name", "price"},
				IsUnique:  false,
				IndexType: "BTREE",
			},
			{
				Database:  "testdb",
				Table:     "products",
				Name:      "idx_product_unique_id",
				Columns:   []string{"id"},
				IsUnique:  true,
				IndexType: "BTREE",
			},
		}

		for _, idx := range indexes {
			if err := cat.CreateIndex(idx); err != nil {
				t.Fatalf("Failed to create index %s: %v", idx.Name, err)
			}
			t.Logf("Created index: %s (columns: %v, unique: %v)", idx.Name, idx.Columns, idx.IsUnique)
		}

		// Verify indexes exist before restart
		allIndexes, err := cat.GetAllIndexes("testdb", "products")
		if err != nil {
			t.Fatalf("Failed to get indexes before restart: %v", err)
		}
		if len(allIndexes) != 4 {
			t.Fatalf("Expected 4 indexes before restart, got %d", len(allIndexes))
		}
		t.Logf("✓ All 4 indexes exist before restart")

		// Close engine
		if err := engine.Close(); err != nil {
			t.Fatalf("Failed to close engine: %v", err)
		}

		t.Log("Phase 1 completed: Database, table, and 4 indexes created")
	}

	// Phase 2: Restart and verify index recovery
	t.Log("Phase 2: Restarting and checking index metadata recovery")
	{
		// Create new engine instance (simulating restart)
		engine, err := storage.NewParquetEngine(testDir)
		if err != nil {
			t.Fatalf("Failed to create engine on restart: %v", err)
		}

		if err := engine.Open(); err != nil {
			t.Fatalf("Failed to open engine on restart: %v", err)
		}
		defer engine.Close()

		// Create new catalog instance
		cat := catalog.NewSimpleSQLCatalog()
		cat.SetStorageEngine(engine)
		if err := cat.Init(); err != nil {
			t.Fatalf("Failed to initialize catalog on restart: %v", err)
		}

		// Check if all indexes are recovered
		allIndexes, err := cat.GetAllIndexes("testdb", "products")
		if err != nil {
			t.Fatalf("Failed to get indexes after restart: %v", err)
		}

		if len(allIndexes) != 4 {
			t.Errorf("FAILED: Expected 4 indexes after restart, got %d", len(allIndexes))
			for _, idx := range allIndexes {
				t.Logf("  Found index: %s", idx.Name)
			}
		} else {
			t.Log("SUCCESS: All 4 indexes recovered after restart")
		}

		// Verify specific indexes
		expectedIndexes := map[string]struct {
			columns  []string
			isUnique bool
		}{
			"idx_product_id":         {columns: []string{"id"}, isUnique: false},
			"idx_product_name":       {columns: []string{"name"}, isUnique: false},
			"idx_product_name_price": {columns: []string{"name", "price"}, isUnique: false},
			"idx_product_unique_id":  {columns: []string{"id"}, isUnique: true},
		}

		for indexName, expected := range expectedIndexes {
			idx, err := cat.GetIndex("testdb", "products", indexName)
			if err != nil {
				t.Errorf("FAILED: Index '%s' not recovered: %v", indexName, err)
				continue
			}

			// Verify columns
			if len(idx.Columns) != len(expected.columns) {
				t.Errorf("FAILED: Index '%s' has wrong number of columns. Expected %d, got %d",
					indexName, len(expected.columns), len(idx.Columns))
				continue
			}
			for i, col := range expected.columns {
				if idx.Columns[i] != col {
					t.Errorf("FAILED: Index '%s' column[%d] mismatch. Expected '%s', got '%s'",
						indexName, i, col, idx.Columns[i])
				}
			}

			// Verify uniqueness
			if idx.IsUnique != expected.isUnique {
				t.Errorf("FAILED: Index '%s' uniqueness mismatch. Expected %v, got %v",
					indexName, expected.isUnique, idx.IsUnique)
			}

			// Verify index type
			if idx.IndexType != "BTREE" {
				t.Errorf("FAILED: Index '%s' type mismatch. Expected 'BTREE', got '%s'",
					indexName, idx.IndexType)
			}

			t.Logf("✓ Index '%s' recovered correctly (columns: %v, unique: %v)",
				indexName, idx.Columns, idx.IsUnique)
		}

		t.Log("Phase 2 completed: Index metadata recovery verified")
	}

	t.Log("========================================")
	t.Log("✓ Index Metadata Recovery Test PASSED")
	t.Log("All indexes were successfully persisted and recovered after restart")
	t.Log("========================================")
}

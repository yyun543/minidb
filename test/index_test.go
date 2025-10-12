package test

import (
	"os"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

// TestIndexOperations tests comprehensive index DDL and DQL operations
func TestIndexOperations(t *testing.T) {
	// Setup test environment - v2.0 Parquet
	storageEngine, err := storage.NewParquetEngine("./test_data/index_test")
	require.NoError(t, err)
	defer storageEngine.Close()

	err = storageEngine.Open()
	require.NoError(t, err)

	cat := catalog.NewCatalog()
	cat.SetStorageEngine(storageEngine)
	err = cat.Init()
	if err != nil {
		t.Fatalf("Failed to initialize catalog: %v", err)
	}
	exec := executor.NewExecutor(cat)
	opt := optimizer.NewOptimizer()

	sess := &session.Session{
		ID:        1,
		CurrentDB: "default",
	}

	t.Run("CreateIndex_Basic", func(t *testing.T) {
		// Create a test table first
		createTableSQL := "CREATE TABLE users (id INT, name VARCHAR, email VARCHAR);"
		ast, err := parser.Parse(createTableSQL)
		require.NoError(t, err)

		plan, err := opt.Optimize(ast)
		require.NoError(t, err)

		_, err = exec.Execute(plan, sess)
		require.NoError(t, err)

		// Test CREATE INDEX
		createIndexSQL := "CREATE INDEX idx_users_email ON users (email);"
		ast, err = parser.Parse(createIndexSQL)
		require.NoError(t, err)

		createIndexStmt, ok := ast.(*parser.CreateIndexStmt)
		require.True(t, ok, "Should parse as CreateIndexStmt")
		assert.Equal(t, "idx_users_email", createIndexStmt.Name)
		assert.Equal(t, "users", createIndexStmt.Table)
		assert.Equal(t, []string{"email"}, createIndexStmt.Columns)
		assert.False(t, createIndexStmt.IsUnique)

		// Execute CREATE INDEX
		plan, err = opt.Optimize(ast)
		require.NoError(t, err)

		_, err = exec.Execute(plan, sess)
		require.NoError(t, err)
	})

	t.Run("CreateIndex_Composite", func(t *testing.T) {
		// Create composite index
		createIndexSQL := "CREATE INDEX idx_users_name_email ON users (name, email);"
		ast, err := parser.Parse(createIndexSQL)
		require.NoError(t, err)

		createIndexStmt, ok := ast.(*parser.CreateIndexStmt)
		require.True(t, ok)
		assert.Equal(t, "idx_users_name_email", createIndexStmt.Name)
		assert.Equal(t, "users", createIndexStmt.Table)
		assert.Equal(t, []string{"name", "email"}, createIndexStmt.Columns)
	})

	t.Run("CreateIndex_Unique", func(t *testing.T) {
		// Create unique index
		createIndexSQL := "CREATE UNIQUE INDEX idx_users_id ON users (id);"
		ast, err := parser.Parse(createIndexSQL)
		require.NoError(t, err)

		createIndexStmt, ok := ast.(*parser.CreateIndexStmt)
		require.True(t, ok)
		assert.Equal(t, "idx_users_id", createIndexStmt.Name)
		assert.True(t, createIndexStmt.IsUnique)
	})

	t.Run("DropIndex_Basic", func(t *testing.T) {
		// First create an index
		createIndexSQL := "CREATE INDEX idx_to_drop ON users (name);"
		ast, err := parser.Parse(createIndexSQL)
		require.NoError(t, err)

		plan, err := opt.Optimize(ast)
		require.NoError(t, err)

		_, err = exec.Execute(plan, sess)
		require.NoError(t, err)

		// Test DROP INDEX
		dropIndexSQL := "DROP INDEX idx_to_drop ON users;"
		ast, err = parser.Parse(dropIndexSQL)
		require.NoError(t, err)

		dropIndexStmt, ok := ast.(*parser.DropIndexStmt)
		require.True(t, ok, "Should parse as DropIndexStmt")
		assert.Equal(t, "idx_to_drop", dropIndexStmt.Name)
		assert.Equal(t, "users", dropIndexStmt.Table)

		// Execute DROP INDEX
		plan, err = opt.Optimize(ast)
		require.NoError(t, err)

		_, err = exec.Execute(plan, sess)
		require.NoError(t, err)
	})

	t.Run("ShowIndexes_Basic", func(t *testing.T) {
		// Test SHOW INDEXES
		showIndexesSQL := "SHOW INDEXES ON users;"
		ast, err := parser.Parse(showIndexesSQL)
		require.NoError(t, err)

		showIndexesStmt, ok := ast.(*parser.ShowIndexesStmt)
		require.True(t, ok, "Should parse as ShowIndexesStmt")
		assert.Equal(t, "users", showIndexesStmt.Table)

		// Execute SHOW INDEXES
		plan, err := opt.Optimize(ast)
		require.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("CreateIndex_DuplicateName", func(t *testing.T) {
		// Create an index
		createIndexSQL := "CREATE INDEX idx_duplicate ON users (email);"
		ast, err := parser.Parse(createIndexSQL)
		require.NoError(t, err)

		plan, err := opt.Optimize(ast)
		require.NoError(t, err)

		_, err = exec.Execute(plan, sess)
		require.NoError(t, err)

		// Try to create index with same name - should fail
		ast, err = parser.Parse(createIndexSQL)
		require.NoError(t, err)

		plan, err = opt.Optimize(ast)
		require.NoError(t, err)

		_, err = exec.Execute(plan, sess)
		assert.Error(t, err, "Should fail when creating duplicate index")
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("DropIndex_NonExistent", func(t *testing.T) {
		// Try to drop non-existent index - should fail
		dropIndexSQL := "DROP INDEX idx_nonexistent ON users;"
		ast, err := parser.Parse(dropIndexSQL)
		require.NoError(t, err)

		plan, err := opt.Optimize(ast)
		require.NoError(t, err)

		_, err = exec.Execute(plan, sess)
		assert.Error(t, err, "Should fail when dropping non-existent index")
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("CreateIndex_InvalidColumn", func(t *testing.T) {
		// Try to create index on non-existent column - should fail
		createIndexSQL := "CREATE INDEX idx_invalid ON users (nonexistent_column);"
		ast, err := parser.Parse(createIndexSQL)
		require.NoError(t, err)

		plan, err := opt.Optimize(ast)
		require.NoError(t, err)

		_, err = exec.Execute(plan, sess)
		assert.Error(t, err, "Should fail when creating index on invalid column")
	})

	t.Run("CreateIndex_InvalidTable", func(t *testing.T) {
		// Try to create index on non-existent table - should fail
		createIndexSQL := "CREATE INDEX idx_invalid ON nonexistent_table (column);"
		ast, err := parser.Parse(createIndexSQL)
		require.NoError(t, err)

		plan, err := opt.Optimize(ast)
		require.NoError(t, err)

		_, err = exec.Execute(plan, sess)
		assert.Error(t, err, "Should fail when creating index on invalid table")
		assert.Contains(t, err.Error(), "does not exist")
	})
}

// TestIndexParser tests index SQL parsing
func TestIndexParser(t *testing.T) {
	t.Run("ParseCreateIndex", func(t *testing.T) {
		sql := "CREATE INDEX idx_name ON table_name (column1);"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		createIndexStmt, ok := stmt.(*parser.CreateIndexStmt)
		require.True(t, ok)
		assert.Equal(t, "idx_name", createIndexStmt.Name)
		assert.Equal(t, "table_name", createIndexStmt.Table)
		assert.Equal(t, []string{"column1"}, createIndexStmt.Columns)
		assert.False(t, createIndexStmt.IsUnique)
	})

	t.Run("ParseCreateUniqueIndex", func(t *testing.T) {
		sql := "CREATE UNIQUE INDEX idx_name ON table_name (column1);"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		createIndexStmt, ok := stmt.(*parser.CreateIndexStmt)
		require.True(t, ok)
		assert.True(t, createIndexStmt.IsUnique)
	})

	t.Run("ParseCreateCompositeIndex", func(t *testing.T) {
		sql := "CREATE INDEX idx_name ON table_name (column1, column2, column3);"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		createIndexStmt, ok := stmt.(*parser.CreateIndexStmt)
		require.True(t, ok)
		assert.Equal(t, []string{"column1", "column2", "column3"}, createIndexStmt.Columns)
	})

	t.Run("ParseDropIndex", func(t *testing.T) {
		sql := "DROP INDEX idx_name ON table_name;"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		dropIndexStmt, ok := stmt.(*parser.DropIndexStmt)
		require.True(t, ok)
		assert.Equal(t, "idx_name", dropIndexStmt.Name)
		assert.Equal(t, "table_name", dropIndexStmt.Table)
	})

	t.Run("ParseShowIndexes", func(t *testing.T) {
		sql := "SHOW INDEXES ON table_name;"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		showIndexesStmt, ok := stmt.(*parser.ShowIndexesStmt)
		require.True(t, ok)
		assert.Equal(t, "table_name", showIndexesStmt.Table)
	})

	t.Run("ParseShowIndexesAlt", func(t *testing.T) {
		sql := "SHOW INDEXES FROM table_name;"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		showIndexesStmt, ok := stmt.(*parser.ShowIndexesStmt)
		require.True(t, ok)
		assert.Equal(t, "table_name", showIndexesStmt.Table)
	})
}

// TestIndexCatalogIntegration tests index metadata management in catalog
func TestIndexCatalogIntegration(t *testing.T) {
	// v2.0 Parquet engine
	storageEngine, err := storage.NewParquetEngine("./test_data/index_catalog_test")
	require.NoError(t, err)
	defer storageEngine.Close()

	err = storageEngine.Open()
	require.NoError(t, err)

	cat := catalog.NewCatalog()
	cat.SetStorageEngine(storageEngine)
	err = cat.Init()
	if err != nil {
		t.Fatalf("Failed to initialize catalog: %v", err)
	}

	t.Run("CreateAndGetIndex", func(t *testing.T) {
		// First create a table
		err := cat.CreateDatabase("testdb")
		require.NoError(t, err)

		// Create index metadata
		indexMeta := catalog.IndexMeta{
			Database:  "testdb",
			Table:     "users",
			Name:      "idx_email",
			Columns:   []string{"email"},
			IsUnique:  false,
			IndexType: "BTREE",
		}

		err = cat.CreateIndex(indexMeta)
		require.NoError(t, err)

		// Retrieve index
		retrievedIndex, err := cat.GetIndex("testdb", "users", "idx_email")
		require.NoError(t, err)
		assert.Equal(t, "idx_email", retrievedIndex.Name)
		assert.Equal(t, []string{"email"}, retrievedIndex.Columns)
		assert.False(t, retrievedIndex.IsUnique)
	})

	t.Run("GetAllIndexesForTable", func(t *testing.T) {
		// Create multiple indexes
		indexes := []catalog.IndexMeta{
			{Database: "testdb", Table: "users", Name: "idx1", Columns: []string{"col1"}, IsUnique: false, IndexType: "BTREE"},
			{Database: "testdb", Table: "users", Name: "idx2", Columns: []string{"col2"}, IsUnique: true, IndexType: "BTREE"},
			{Database: "testdb", Table: "users", Name: "idx3", Columns: []string{"col1", "col2"}, IsUnique: false, IndexType: "BTREE"},
		}

		for _, idx := range indexes {
			err := cat.CreateIndex(idx)
			require.NoError(t, err)
		}

		// Get all indexes
		allIndexes, err := cat.GetAllIndexes("testdb", "users")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(allIndexes), 3)
	})

	t.Run("DropIndex", func(t *testing.T) {
		// Create index
		indexMeta := catalog.IndexMeta{
			Database:  "testdb",
			Table:     "users",
			Name:      "idx_to_drop",
			Columns:   []string{"name"},
			IsUnique:  false,
			IndexType: "BTREE",
		}

		err := cat.CreateIndex(indexMeta)
		require.NoError(t, err)

		// Drop index
		err = cat.DropIndex("testdb", "users", "idx_to_drop")
		require.NoError(t, err)

		// Verify it's gone
		_, err = cat.GetIndex("testdb", "users", "idx_to_drop")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})
}

// TestIndexPersistenceAndRecovery is a critical regression test for index metadata persistence.
// This test verifies the complete lifecycle: index creation → persistence to Delta Log →
// server restart simulation → index recovery from Delta Log.
//
// Bug History: Previously, CreateIndex had a misleading comment claiming "Index metadata is
// persisted in Delta Log" but no actual persistence code existed, causing indexes to disappear
// after server restart. This test ensures that bug never returns.
//
// Note: This test uses the catalog API directly (not SQL execution) to keep it simple and focused.
func TestIndexPersistenceAndRecovery(t *testing.T) {
	// This test follows the same pattern as TestIndexMetadataRecoveryAfterRestart in
	// metadata_recovery_test.go, which proved to work correctly.
	testDir := "./test_data/index_persistence_test"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// ========== Phase 1: Create indexes and verify persistence ==========
	t.Log("Phase 1: Creating database, table, and indexes")
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
		require.NoError(t, err, "Failed to create database in engine")
		err = cat.CreateDatabase("testdb")
		require.NoError(t, err, "Failed to register database in catalog")

		// Create table
		schema := arrow.NewSchema([]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "name", Type: arrow.BinaryTypes.String},
			{Name: "price", Type: arrow.PrimitiveTypes.Int64},
		}, nil)

		err = engine.CreateTable("testdb", "products", schema)
		require.NoError(t, err, "Failed to create table in engine")
		err = cat.CreateTable("testdb", catalog.TableMeta{
			Database: "testdb",
			Table:    "products",
			Schema:   schema,
		})
		require.NoError(t, err, "Failed to register table in catalog")

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
			err = cat.CreateIndex(idx)
			require.NoError(t, err, "Failed to create index %s", idx.Name)
			t.Logf("✓ Created index: %s (columns: %v, unique: %v)", idx.Name, idx.Columns, idx.IsUnique)
		}

		// Verify indexes exist before restart
		allIndexes, err := cat.GetAllIndexes("testdb", "products")
		require.NoError(t, err, "Failed to get indexes before restart")
		assert.Equal(t, 3, len(allIndexes), "Should have 3 indexes before restart")
		t.Log("✓ All 3 indexes exist in catalog before restart")

		// Close engine
		err = engine.Close()
		require.NoError(t, err, "Failed to close storage engine")
		t.Log("Phase 1 completed: Database, table, and 3 indexes created")
	}

	// ========== Phase 2: Reopen engine and verify recovery ==========
	t.Log("Phase 2: Reopening storage engine (simulating server restart)")
	{
		// Create new engine instance (simulating restart)
		engine, err := storage.NewParquetEngine(testDir)
		require.NoError(t, err, "Failed to create storage engine on restart")

		err = engine.Open()
		require.NoError(t, err, "Failed to open storage engine on restart")
		defer engine.Close()

		// Create new catalog instance
		cat := catalog.NewSimpleSQLCatalog()
		cat.SetStorageEngine(engine)
		err = cat.Init()
		require.NoError(t, err, "Failed to initialize catalog on restart")

		t.Log("✓ Storage engine reopened and catalog initialized")

		// Verify indexes were recovered from Delta Log
		idx1, err := cat.GetIndex("testdb", "products", "idx_product_id")
		require.NoError(t, err, "Index idx_product_id should be recovered")
		assert.Equal(t, "idx_product_id", idx1.Name)
		assert.Equal(t, []string{"id"}, idx1.Columns)
		assert.False(t, idx1.IsUnique)
		assert.Equal(t, "BTREE", idx1.IndexType)
		t.Log("✓ Verified idx_product_id recovered correctly")

		idx2, err := cat.GetIndex("testdb", "products", "idx_product_name_price")
		require.NoError(t, err, "Index idx_product_name_price should be recovered")
		assert.Equal(t, "idx_product_name_price", idx2.Name)
		assert.Equal(t, []string{"name", "price"}, idx2.Columns)
		assert.False(t, idx2.IsUnique)
		t.Log("✓ Verified idx_product_name_price recovered correctly")

		idx3, err := cat.GetIndex("testdb", "products", "idx_product_unique_id")
		require.NoError(t, err, "Index idx_product_unique_id should be recovered")
		assert.Equal(t, "idx_product_unique_id", idx3.Name)
		assert.Equal(t, []string{"id"}, idx3.Columns)
		assert.True(t, idx3.IsUnique)
		t.Log("✓ Verified idx_product_unique_id recovered correctly")

		// Verify all indexes for the table
		allIndexes, err := cat.GetAllIndexes("testdb", "products")
		require.NoError(t, err, "Should be able to get all indexes after restart")
		assert.Equal(t, 3, len(allIndexes), "Should have exactly 3 indexes after restart")
		t.Log("✓ Verified GetAllIndexes returns 3 indexes")

		t.Log("Phase 2 completed: All index metadata verified after restart")
	}

	t.Log("========================================")
	t.Log("✓ Index Persistence and Recovery Test PASSED")
	t.Log("All indexes were successfully persisted to Delta Log and recovered after restart")
	t.Log("========================================")
}

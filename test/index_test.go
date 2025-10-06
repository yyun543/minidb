package test

import (
	"testing"

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
	// Setup test environment
	storageEngine, err := storage.NewMemTable("test_index.wal")
	require.NoError(t, err)
	defer storageEngine.Close()

	err = storageEngine.Open()
	require.NoError(t, err)

	cat := catalog.CreateTemporaryCatalog(storageEngine)
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
	storageEngine, err := storage.NewMemTable("test_index_catalog.wal")
	require.NoError(t, err)
	defer storageEngine.Close()

	err = storageEngine.Open()
	require.NoError(t, err)

	cat := catalog.CreateTemporaryCatalog(storageEngine)

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

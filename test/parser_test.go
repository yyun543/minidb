package test

// TODO Parser 单元测试

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/parser"
)

func TestParser(t *testing.T) {
	// 测试CREATE DATABASE语句解析
	t.Run("ParseCreateDatabase", func(t *testing.T) {
		sql := "CREATE DATABASE test_db"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)
		assert.NotNil(t, stmt)

		createDbStmt, ok := stmt.(*parser.CreateDatabaseStmt)
		assert.True(t, ok)
		assert.Equal(t, "test_db", createDbStmt.Database)
	})

	// 测试CREATE TABLE语句解析
	t.Run("ParseCreateTable", func(t *testing.T) {
		sql := `CREATE TABLE users (
			id INTEGER NOT NULL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			age INTEGER,
			balance DOUBLE,
			locked BOOLEAN,
			created_at TIMESTAMP,
			PRIMARY KEY (id, name)
		)`
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)
		assert.NotNil(t, stmt)

		createStmt, ok := stmt.(*parser.CreateTableStmt)
		assert.True(t, ok)
		assert.Equal(t, "users", createStmt.Table)
		assert.Len(t, createStmt.Columns, 6)
	})

	// 测试SELECT语句解析
	t.Run("ParseSelect", func(t *testing.T) {
		sql := "SELECT id, name FROM users WHERE age > 18"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)
		assert.NotNil(t, stmt)

		selectStmt, ok := stmt.(*parser.SelectStmt)
		assert.True(t, ok)
		assert.Equal(t, []string{"id", "name"}, selectStmt.Columns)
		assert.Equal(t, "users", selectStmt.From)
		assert.NotNil(t, selectStmt.Where)
	})

	// 测试INSERT语句解析
	t.Run("ParseInsert", func(t *testing.T) {
		sql := "INSERT INTO users (id, name) VALUES (1, 'test')"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)
		assert.NotNil(t, stmt)

		insertStmt, ok := stmt.(*parser.InsertStmt)
		assert.True(t, ok)
		assert.Equal(t, "users", insertStmt.Table)
		assert.Equal(t, []string{"id", "name"}, insertStmt.Columns)
		assert.Len(t, insertStmt.Values, 2)
	})

	// 测试UPDATE语句解析
	t.Run("ParseUpdate", func(t *testing.T) {
		sql := "UPDATE users SET name = 'updated' WHERE id = 1"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)
		assert.NotNil(t, stmt)

		updateStmt, ok := stmt.(*parser.UpdateStmt)
		assert.True(t, ok)
		assert.Equal(t, "users", updateStmt.Table)
		assert.Len(t, updateStmt.Assignments, 1)
		assert.NotNil(t, updateStmt.Where)
	})

	// 测试DELETE语句解析
	t.Run("ParseDelete", func(t *testing.T) {
		sql := "DELETE FROM users WHERE id = 1"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)
		assert.NotNil(t, stmt)

		deleteStmt, ok := stmt.(*parser.DeleteStmt)
		assert.True(t, ok)
		assert.Equal(t, "users", deleteStmt.Table)
		assert.NotNil(t, deleteStmt.Where)
	})
}

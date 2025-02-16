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

	// 测试CREATE INDEX语句解析
	t.Run("ParseCreateIndex", func(t *testing.T) {
		sql := "CREATE INDEX idx_name ON users (id, name);"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)
		assert.NotNil(t, stmt)

		createIndexStmt, ok := stmt.(*parser.CreateIndexStmt)
		assert.True(t, ok)
		assert.Equal(t, "idx_name", createIndexStmt.Name)
		assert.Equal(t, "users", createIndexStmt.Table)
		assert.Equal(t, []string{"id", "name"}, createIndexStmt.Columns)
	})

	// 测试DROP TABLE语句解析
	t.Run("ParseDropTable", func(t *testing.T) {
		sql := "DROP TABLE users;"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)
		assert.NotNil(t, stmt)

		dropTableStmt, ok := stmt.(*parser.DropTableStmt)
		assert.True(t, ok)
		assert.Equal(t, "users", dropTableStmt.Table)
	})

	// 测试DROP DATABASE语句解析
	t.Run("ParseDropDatabase", func(t *testing.T) {
		sql := "DROP DATABASE test_db;"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)
		assert.NotNil(t, stmt)

		dropDbStmt, ok := stmt.(*parser.DropDatabaseStmt)
		assert.True(t, ok)
		assert.Equal(t, "test_db", dropDbStmt.Database)
	})

	// 测试INSERT语句解析
	t.Run("ParseInsert", func(t *testing.T) {
		// INSERT INTO users (id, name, age) VALUES (1, 'test', 20);
		sql := "INSERT INTO users (id, name, age) VALUES (1, 'test', 20)"
		stmt, err := parser.Parse(sql)
		assert.NoError(t, err)
		assert.NotNil(t, stmt)

		insertStmt, ok := stmt.(*parser.InsertStmt)
		assert.True(t, ok)
		assert.Equal(t, "users", insertStmt.Table)
		assert.Equal(t, []string{"id", "name", "age"}, insertStmt.Columns)
		assert.Len(t, insertStmt.Values, 3)
		// INSERT INTO users VALUES (1, 'test', 20);
		sql2 := "INSERT INTO users VALUES (1, 'test', 20);"
		stmt2, err := parser.Parse(sql2)
		assert.NoError(t, err)
		assert.NotNil(t, stmt2)
		insertStmt2, ok := stmt2.(*parser.InsertStmt)
		assert.True(t, ok)
		assert.Equal(t, "users", insertStmt2.Table)
		// 检查Values的类型与值
		assert.IsType(t, &parser.IntegerLiteral{}, insertStmt2.Values[0])
		assert.Equal(t, int64(1), insertStmt2.Values[0].(*parser.IntegerLiteral).Value)
		assert.IsType(t, &parser.StringLiteral{}, insertStmt2.Values[1])
		assert.Equal(t, "test", insertStmt2.Values[1].(*parser.StringLiteral).Value)
		assert.IsType(t, &parser.IntegerLiteral{}, insertStmt2.Values[2])
		assert.Equal(t, int64(20), insertStmt2.Values[2].(*parser.IntegerLiteral).Value)
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

	// 测试SELECT语句解析
	t.Run("ParseSelect", func(t *testing.T) {
		// 测试基本的SELECT语句
		t.Run("BasicSelect", func(t *testing.T) {
			sql := "SELECT id, name FROM users;"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)
			assert.NotNil(t, stmt)

			selectStmt, ok := stmt.(*parser.SelectStmt)
			assert.True(t, ok)
			expectedColumns := []*parser.ColumnItem{
				{Column: "id", Table: "", Alias: ""},
				{Column: "name", Table: "", Alias: ""},
			}
			assert.Equal(t, len(expectedColumns), len(selectStmt.Columns))
			for i, expected := range expectedColumns {
				actual := selectStmt.Columns[i]
				assert.Equal(t, expected.Column, actual.Column, "Column %d: Column mismatch", i)
				assert.Equal(t, expected.Table, actual.Table, "Column %d: Table mismatch", i)
				assert.Equal(t, expected.Alias, actual.Alias, "Column %d: Alias mismatch", i)
			}
			assert.Equal(t, "users", selectStmt.From)
			assert.Nil(t, selectStmt.Where)
		})

		// 测试带WHERE子句的SELECT
		t.Run("SelectWithWhere", func(t *testing.T) {
			sql := "SELECT id, name FROM users WHERE age > 18;"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)
			assert.NotNil(t, stmt)

			selectStmt, ok := stmt.(*parser.SelectStmt)
			assert.True(t, ok)
			expectedColumns := []*parser.ColumnItem{
				{Column: "id", Table: "", Alias: ""},
				{Column: "name", Table: "", Alias: ""},
			}
			assert.Equal(t, len(expectedColumns), len(selectStmt.Columns))
			for i, expected := range expectedColumns {
				actual := selectStmt.Columns[i]
				assert.Equal(t, expected.Column, actual.Column, "Column %d: Column mismatch", i)
				assert.Equal(t, expected.Table, actual.Table, "Column %d: Table mismatch", i)
				assert.Equal(t, expected.Alias, actual.Alias, "Column %d: Alias mismatch", i)
			}
			assert.Equal(t, "users", selectStmt.From)
			assert.NotNil(t, selectStmt.Where)

			// 验证WHERE条件
			whereExpr, ok := selectStmt.Where.Condition.(*parser.BinaryExpr)
			assert.True(t, ok)
			assert.Equal(t, ">", whereExpr.Operator)

			// 验证左操作数（列引用）
			left, ok := whereExpr.Left.(*parser.ColumnRef)
			assert.True(t, ok)
			assert.Equal(t, "age", left.Column)

			// 验证右操作数（数字字面量）
			right, ok := whereExpr.Right.(*parser.IntegerLiteral)
			assert.True(t, ok)
			assert.Equal(t, int64(18), right.Value)
		})

		// 测试SELECT *
		t.Run("SelectAll", func(t *testing.T) {
			sql := "SELECT * FROM users;"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)
			assert.NotNil(t, stmt)

			selectStmt, ok := stmt.(*parser.SelectStmt)
			assert.True(t, ok)
			assert.True(t, selectStmt.All)
			assert.Equal(t, "users", selectStmt.From)
		})

		// 测试带JOIN的SELECT
		t.Run("SelectWithJoin", func(t *testing.T) {
			sql := "SELECT u.id, u.name, o.order_id FROM users u JOIN orders o ON u.id = o.user_id;"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)
			assert.NotNil(t, stmt)

			selectStmt, ok := stmt.(*parser.SelectStmt)
			assert.True(t, ok)
			expectedColumns := []*parser.ColumnItem{
				{Column: "id", Table: "u", Alias: ""},
				{Column: "name", Table: "u", Alias: ""},
				{Column: "order_id", Table: "o", Alias: ""},
			}
			assert.Equal(t, len(expectedColumns), len(selectStmt.Columns))
			for i, expected := range expectedColumns {
				actual := selectStmt.Columns[i]
				assert.Equal(t, expected.Column, actual.Column, "Column %d: Column mismatch", i)
				assert.Equal(t, expected.Table, actual.Table, "Column %d: Table mismatch", i)
				assert.Equal(t, expected.Alias, actual.Alias, "Column %d: Alias mismatch", i)
			}
			assert.Equal(t, "users", selectStmt.From)
			assert.Len(t, selectStmt.Joins, 1)

			// 验证JOIN
			join := selectStmt.Joins[0]
			assert.Equal(t, "INNER", join.JoinType)
			assert.Equal(t, "orders", join.Right.Table)

			// 验证JOIN条件
			joinCond, ok := join.Condition.(*parser.BinaryExpr)
			assert.True(t, ok)
			assert.Equal(t, "=", joinCond.Operator)
		})

		// 测试带GROUP BY的SELECT
		t.Run("SelectWithGroupBy", func(t *testing.T) {
			sql := "SELECT department, COUNT(*) as count FROM employees GROUP BY department;"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)
			assert.NotNil(t, stmt)

			selectStmt, ok := stmt.(*parser.SelectStmt)
			assert.True(t, ok)
			expectedColumns := []*parser.ColumnItem{
				{Column: "department", Table: "", Alias: "", Kind: parser.ColumnItemColumn},
				{Column: "", Table: "", Alias: "count", Kind: parser.ColumnItemFunction},
			}
			assert.Equal(t, len(expectedColumns), len(selectStmt.Columns))
			for i, expected := range expectedColumns {
				actual := selectStmt.Columns[i]
				assert.Equal(t, expected.Column, actual.Column, "Column %d: Column mismatch", i)
				assert.Equal(t, expected.Table, actual.Table, "Column %d: Table mismatch", i)
				assert.Equal(t, expected.Alias, actual.Alias, "Column %d: Alias mismatch", i)
				assert.Equal(t, expected.Kind, actual.Kind, "Column %d: Kind mismatch", i)
			}
			assert.Equal(t, "employees", selectStmt.From)
			assert.Len(t, selectStmt.GroupBy, 1)

			// 验证GROUP BY
			groupByExpr, ok := selectStmt.GroupBy[0].(*parser.ColumnRef)
			assert.True(t, ok)
			assert.Equal(t, "department", groupByExpr.Column)
		})

		// 测试带HAVING的SELECT
		t.Run("SelectWithHaving", func(t *testing.T) {
			sql := "SELECT department, COUNT(*) as count FROM employees GROUP BY department HAVING count > 5;"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)
			assert.NotNil(t, stmt)

			selectStmt, ok := stmt.(*parser.SelectStmt)
			assert.True(t, ok)
			assert.NotNil(t, selectStmt.Having)

			// 验证HAVING条件
			havingExpr, ok := selectStmt.Having.(*parser.BinaryExpr)
			assert.True(t, ok)
			assert.Equal(t, ">", havingExpr.Operator)
		})

		// 测试带ORDER BY的SELECT
		t.Run("SelectWithOrderBy", func(t *testing.T) {
			sql := "SELECT id, name FROM users ORDER BY age DESC, name ASC;"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)
			assert.NotNil(t, stmt)

			selectStmt, ok := stmt.(*parser.SelectStmt)
			assert.True(t, ok)
			assert.Len(t, selectStmt.OrderBy, 2)

			// 验证ORDER BY
			assert.Equal(t, "age", selectStmt.OrderBy[0].Expr.(*parser.ColumnRef).Column)
			assert.Equal(t, "DESC", selectStmt.OrderBy[0].Direction)
			assert.Equal(t, "name", selectStmt.OrderBy[1].Expr.(*parser.ColumnRef).Column)
			assert.Equal(t, "ASC", selectStmt.OrderBy[1].Direction)
		})

		// 测试带LIMIT的SELECT
		t.Run("SelectWithLimit", func(t *testing.T) {
			sql := "SELECT id, name FROM users LIMIT 10;"
			stmt, err := parser.Parse(sql)
			assert.NoError(t, err)
			assert.NotNil(t, stmt)

			selectStmt, ok := stmt.(*parser.SelectStmt)
			assert.True(t, ok)
			assert.Equal(t, int64(10), selectStmt.Limit)
		})
	})
}

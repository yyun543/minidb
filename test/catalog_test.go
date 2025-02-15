package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/catalog"
)

// TODO Catalog 单元测试

func TestCatalog(t *testing.T) {
	// 创建Catalog实例
	cat := catalog.NewCatalog()

	// 测试创建数据库
	t.Run("CreateDatabase", func(t *testing.T) {
		err := cat.CreateDatabase("testdb")
		assert.NoError(t, err)

		// 重复创建应该报错
		err = cat.CreateDatabase("testdb")
		assert.Error(t, err)
	})

	// 测试切换数据库
	t.Run("UseDatabase", func(t *testing.T) {
		err := cat.UseDatabase("testdb")
		assert.NoError(t, err)

		// 切换到不存在的数据库应该报错
		err = cat.UseDatabase("notexist")
		assert.Error(t, err)
	})

	// 测试创建表
	t.Run("CreateTable", func(t *testing.T) {
		table := &catalog.TableMeta{
			ID:   time.Now().UnixNano(),
			Name: "users",
			Columns: []catalog.ColumnMeta{
				{
					ID:      1,
					Name:    "id",
					Type:    "INT64",
					NotNull: true,
				},
				{
					ID:      2,
					Name:    "name",
					Type:    "STRING",
					NotNull: true,
				},
			},
			Constraints: []catalog.Constraint{
				{
					Name:    "pk_users",
					Type:    "PRIMARY",
					Columns: []string{"id"},
				},
			},
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}

		err := cat.CreateTable(table)
		assert.NoError(t, err)

		// 重复创建应该报错
		err = cat.CreateTable(table)
		assert.Error(t, err)
	})

	// 测试获取表
	t.Run("GetTable", func(t *testing.T) {
		table, err := cat.GetTable("users")
		assert.NoError(t, err)
		assert.NotNil(t, table)
		assert.Equal(t, "users", table.Name)
		assert.Len(t, table.Columns, 2)

		// 获取不存在的表应该报错
		table, err = cat.GetTable("notexist")
		assert.Error(t, err)
		assert.Nil(t, table)
	})

	// 测试删除表
	t.Run("DropTable", func(t *testing.T) {
		err := cat.DropTable("users")
		assert.NoError(t, err)

		// 删除不存在的表应该报错
		err = cat.DropTable("users")
		assert.Error(t, err)
	})

	// 测试系统表
	t.Run("SystemTables", func(t *testing.T) {
		// 获取系统表定义
		table, err := cat.GetTable("sys_databases")
		assert.NoError(t, err)
		assert.NotNil(t, table)
		assert.Equal(t, "sys_databases", table.Name)

		table, err = cat.GetTable("sys_tables")
		assert.NoError(t, err)
		assert.NotNil(t, table)
		assert.Equal(t, "sys_tables", table.Name)

		table, err = cat.GetTable("sys_columns")
		assert.NoError(t, err)
		assert.NotNil(t, table)
		assert.Equal(t, "sys_columns", table.Name)
	})
}

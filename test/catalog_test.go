package test

import (
	"github.com/yyun543/minidb/internal/storage"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/catalog"
)

// Catalog 单元测试

func TestCatalog(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "minidb-catalog-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// 设置数据目录
	os.Setenv("MINIDB_DATA_DIR", tmpDir)

	// 创建Catalog实例
	cat, err := catalog.NewCatalog()
	assert.NoError(t, err)
	defer cat.Close()

	// 测试数据库操作
	t.Run("Database Operations", func(t *testing.T) {
		// 测试创建数据库
		t.Run("CreateDatabase", func(t *testing.T) {
			err := cat.CreateDatabase("test_db")
			assert.NoError(t, err)

			// 验证数据库是否创建成功
			db, err := cat.GetDatabase("test_db")
			assert.NoError(t, err)
			assert.NotNil(t, db)
			assert.Equal(t, "test_db", db.Name)
			assert.NotZero(t, db.ID)
			assert.NotZero(t, db.CreateTime)
			assert.NotZero(t, db.UpdateTime)

			// 测试重复创建
			err = cat.CreateDatabase("test_db")
			assert.Error(t, err)
		})

		// 测试获取不存在的数据库
		t.Run("GetNonExistentDatabase", func(t *testing.T) {
			db, err := cat.GetDatabase("non_existent_db")
			assert.Error(t, err)
			assert.Nil(t, db)
		})
	})

	// 测试表操作
	t.Run("Table Operations", func(t *testing.T) {
		// 测试创建表
		t.Run("CreateTable", func(t *testing.T) {
			table := &catalog.TableMeta{
				Name: "users",
				Columns: []catalog.ColumnMeta{
					{
						Name:       "id",
						Type:       "INTEGER",
						NotNull:    true,
						CreateTime: time.Now(),
						UpdateTime: time.Now(),
					},
					{
						Name:       "name",
						Type:       "VARCHAR",
						NotNull:    true,
						CreateTime: time.Now(),
						UpdateTime: time.Now(),
					},
					{
						Name:       "age",
						Type:       "INTEGER",
						NotNull:    false,
						CreateTime: time.Now(),
						UpdateTime: time.Now(),
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

			// 在不存在的数据库中创建表
			err := cat.CreateTable("non_existent_db", table)
			assert.Error(t, err)

			// 在存在的数据库中创建表
			err = cat.CreateTable("test_db", table)
			assert.NoError(t, err)

			// 验证表是否创建成功
			createdTable, err := cat.GetTable("test_db", "users")
			assert.NoError(t, err)
			assert.NotNil(t, createdTable)
			assert.Equal(t, "users", createdTable.Name)
			assert.Len(t, createdTable.Columns, 3)
			assert.Len(t, createdTable.Constraints, 1)

			// 测试重复创建表
			err = cat.CreateTable("test_db", table)
			assert.Error(t, err)
		})

		// 测试获取不存在的表
		t.Run("GetNonExistentTable", func(t *testing.T) {
			table, err := cat.GetTable("test_db", "non_existent_table")
			assert.Error(t, err)
			assert.Nil(t, table)
		})

		// 测试删除表
		t.Run("DropTable", func(t *testing.T) {
			// 删除不存在的表
			err := cat.DropTable("test_db", "non_existent_table")
			assert.Error(t, err)

			// 删除存在的表
			err = cat.DropTable("test_db", "users")
			assert.NoError(t, err)

			// 验证表是否已删除
			table, err := cat.GetTable("test_db", "users")
			assert.Error(t, err)
			assert.Nil(t, table)
		})

		// 测试系统表操作
		t.Run("SystemTables", func(t *testing.T) {
			// 尝试删除系统表（应该失败）
			err := cat.DropTable(storage.SYS_DATABASE, storage.SYS_TABLES)
			assert.Error(t, err)

			// 验证系统表是否存在
			sysTables := []string{
				storage.SYS_DATABASES,
				storage.SYS_TABLES,
				storage.SYS_COLUMNS,
				storage.SYS_INDEXES,
			}

			for _, tableName := range sysTables {
				table, err := cat.GetTable(storage.SYS_DATABASE, tableName)
				assert.NoError(t, err)
				assert.NotNil(t, table)
				assert.Equal(t, tableName, table.Name)
			}
		})
	})

	// 测试删除数据库
	t.Run("DropDatabase", func(t *testing.T) {
		// 删除不存在的数据库
		err := cat.DropDatabase("non_existent_db")
		assert.Error(t, err)

		// 尝试删除系统数据库（应该失败）
		err = cat.DropDatabase(storage.SYS_DATABASE)
		assert.Error(t, err)

		// 删除存在的数据库
		err = cat.DropDatabase("test_db")
		assert.NoError(t, err)

		// 验证数据库是否已删除
		db, err := cat.GetDatabase("test_db")
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

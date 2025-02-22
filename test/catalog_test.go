package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/storage"
)

// setupEngine 初始化临时的存储引擎，用于 Catalog 测试。
func setupEngine(t *testing.T) storage.Engine {
	tmpDir, err := os.MkdirTemp("", "catalog_test")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	walPath := filepath.Join(tmpDir, "test.catalog")
	engine, err := storage.NewMemTable(walPath)
	require.NoError(t, err)
	err = engine.Open()
	require.NoError(t, err)
	t.Cleanup(func() {
		engine.Close()
	})
	return engine
}

// TestCatalog_Init 测试Catalog初始化
func TestCatalog_Init(t *testing.T) {
	engine := setupEngine(t)
	c := catalog.NewCatalog(engine)

	// 测试初始化
	err := c.Init()
	require.NoError(t, err)

	// 验证系统表是否创建
	dbs, err := c.GetAllDatabases()
	require.NoError(t, err)
	assert.Contains(t, dbs, catalog.DatabaseMeta{Name: storage.SYS_DATABASE})

	// 验证系统表
	tables := []string{
		storage.SYS_DATABASES,
		storage.SYS_TABLES,
		storage.SYS_COLUMNS,
		storage.SYS_INDEXES,
	}

	for _, tableName := range tables {
		table, err := c.GetTable(storage.SYS_DATABASE, tableName)
		require.NoError(t, err)
		assert.Equal(t, storage.SYS_DATABASE, table.Database)
		assert.Equal(t, tableName, table.Table)
		assert.NotNil(t, table.Schema)
	}
}

// TestCatalog_DatabaseLifecycle 测试数据库的创建、读取、扫描和删除操作。
func TestCatalog_DatabaseLifecycle(t *testing.T) {
	engine := setupEngine(t)
	c := catalog.NewCatalog(engine)

	// 初始化 Catalog
	err := c.Init()
	require.NoError(t, err)

	// 创建数据库
	dbName := "testdb"
	err = c.CreateDatabase(dbName)
	require.NoError(t, err)

	// 读取数据库元数据
	db, err := c.GetDatabase(dbName)
	require.NoError(t, err)
	assert.Equal(t, dbName, db.Name)

	// 获取所有数据库
	dbs, err := c.GetAllDatabases()
	require.NoError(t, err)
	assert.Contains(t, dbs, catalog.DatabaseMeta{Name: dbName})
	assert.Contains(t, dbs, catalog.DatabaseMeta{Name: storage.SYS_DATABASE})

	// 删除数据库
	err = c.DeleteDatabase(dbName)
	require.NoError(t, err)

	// 验证删除
	_, err = c.GetDatabase(dbName)
	assert.Error(t, err)

	// 验证系统数据库不能删除
	err = c.DeleteDatabase(storage.SYS_DATABASE)
	assert.Error(t, err)
}

// TestCatalog_TableLifecycle 测试表的创建、读取及删除操作。
func TestCatalog_TableLifecycle(t *testing.T) {
	engine := setupEngine(t)
	c := catalog.NewCatalog(engine)

	// 初始化 Catalog
	err := c.Init()
	require.NoError(t, err)

	// 创建测试数据库
	dbName := "testdb"
	err = c.CreateDatabase(dbName)
	require.NoError(t, err)

	// 创建表
	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "name", Type: arrow.BinaryTypes.String},
		},
		nil,
	)

	tableMeta := catalog.TableMeta{
		Database:   dbName,
		Table:      "testtable",
		ChunkCount: 1,
		Schema:     schema,
	}

	err = c.CreateTable(dbName, tableMeta)
	require.NoError(t, err)

	// 读取表元数据
	table, err := c.GetTable(dbName, "testtable")
	require.NoError(t, err)
	assert.Equal(t, tableMeta.Database, table.Database)
	assert.Equal(t, tableMeta.Table, table.Table)
	assert.Equal(t, tableMeta.Schema.String(), table.Schema.String())

	// 删除表
	err = c.DeleteTable(dbName, "testtable")
	require.NoError(t, err)

	// 验证删除
	_, err = c.GetTable(dbName, "testtable")
	assert.Error(t, err)

	// 验证系统表不能删除
	err = c.DeleteTable(storage.SYS_DATABASE, storage.SYS_TABLES)
	assert.Error(t, err)
}

// TestCatalog_ErrorCases 测试错误情况
func TestCatalog_ErrorCases(t *testing.T) {
	engine := setupEngine(t)
	c := catalog.NewCatalog(engine)

	// 初始化 Catalog
	err := c.Init()
	require.NoError(t, err)

	// 测试空数据库名
	err = c.CreateDatabase("")
	assert.Error(t, err)

	// 测试重复创建数据库
	err = c.CreateDatabase("testdb")
	require.NoError(t, err)
	err = c.CreateDatabase("testdb")
	assert.Error(t, err)

	// 测试在不存在的数据库中创建表
	schema := arrow.NewSchema([]arrow.Field{{Name: "id", Type: arrow.PrimitiveTypes.Int64}}, nil)
	err = c.CreateTable("nonexistent", catalog.TableMeta{
		Database: "nonexistent",
		Table:    "test",
		Schema:   schema,
	})
	assert.Error(t, err)

	// 测试删除不存在的数据库
	err = c.DeleteDatabase("nonexistent")
	assert.Error(t, err)

	// 测试删除不存在的表
	err = c.DeleteTable("testdb", "nonexistent")
	assert.Error(t, err)
}

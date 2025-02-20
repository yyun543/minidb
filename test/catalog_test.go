package test

import (
	"os"
	"path/filepath"
	"testing"

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

// TestCatalog_DatabaseLifecycle 测试数据库的创建、读取、扫描和删除操作。
func TestCatalog_DatabaseLifecycle(t *testing.T) {
	engine := setupEngine(t)
	c := catalog.NewCatalog(engine)

	// 初始化 Catalog（会自动初始化系统表）
	err := c.Init()
	require.NoError(t, err)

	// 创建数据库
	dbName := "testdb"
	err = c.CreateDatabase(dbName)
	require.NoError(t, err)

	// 读取数据库元数据
	dbMeta, err := c.GetDatabase(dbName)
	require.NoError(t, err)
	assert.Equal(t, dbName, dbMeta.Name)

	// 获取所有数据库的元数据，检查是否包含新创建的数据库
	dbs, err := c.GetAllDatabases()
	require.NoError(t, err)
	found := false
	for _, meta := range dbs {
		if meta.Name == dbName {
			found = true
			break
		}
	}
	assert.True(t, found, "database %s should appear in GetAllDatabases", dbName)

	// 删除数据库
	err = c.DeleteDatabase(dbName)
	require.NoError(t, err)

	// 再次读取数据库，应返回错误
	_, err = c.GetDatabase(dbName)
	require.Error(t, err)
}

// TestCatalog_TableLifecycle 测试表的创建、读取及删除操作。
func TestCatalog_TableLifecycle(t *testing.T) {
	engine := setupEngine(t)
	c := catalog.NewCatalog(engine)

	// 初始化 Catalog
	err := c.Init()
	require.NoError(t, err)

	// 创建所属数据库（表元数据依赖数据库存在）
	dbName := "testdb"
	err = c.CreateDatabase(dbName)
	require.NoError(t, err)

	// 构造表元数据，Schema 字段存储的是 Arrow Schema 的 JSON 表示（这里示例为简单结构）
	tableMeta := catalog.TableMeta{
		Database: dbName,
		Table:    "testtable",
		Schema:   "{\"fields\": [{\"name\": \"id\", \"type\": \"int64\"}, {\"name\": \"name\", \"type\": \"string\"}]}",
	}
	err = c.CreateTable(dbName, tableMeta)
	require.NoError(t, err)

	// 读取表元数据并进行校验
	retTable, err := c.GetTable(dbName, "testtable")
	require.NoError(t, err)
	assert.Equal(t, tableMeta.Database, retTable.Database)
	assert.Equal(t, tableMeta.Table, retTable.Table)
	assert.Equal(t, tableMeta.Schema, retTable.Schema)

	// 删除表元数据
	err = c.DeleteTable(dbName, "testtable")
	require.NoError(t, err)

	// 删除后再查询表元数据，应返回错误
	_, err = c.GetTable(dbName, "testtable")
	require.Error(t, err)
}

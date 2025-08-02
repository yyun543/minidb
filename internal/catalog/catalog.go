package catalog

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yyun543/minidb/internal/storage"
)

// Catalog 是系统 Catalog 的入口，负责管理数据库、表及其他元数据。
type Catalog struct {
	engine      storage.Engine
	metadataMgr *MetadataManager
}

// NewCatalog 创建一个新的 Catalog 实例，依赖底层 storage 引擎。
func NewCatalog(engine storage.Engine) *Catalog {
	return &Catalog{
		engine:      engine,
		metadataMgr: NewMetadataManager(engine),
	}
}

// NewCatalogWithDefaultStorage 创建一个使用默认内存存储的 Catalog 实例，用于测试
func NewCatalogWithDefaultStorage() (*Catalog, error) {
	// 使用临时内存存储引擎
	tmpDir, err := os.MkdirTemp("", "test_catalog")
	if err != nil {
		return nil, err
	}
	walPath := filepath.Join(tmpDir, "test.wal")
	tmpEngine, err := storage.NewMemTable(walPath)
	if err != nil {
		return nil, err
	}
	err = tmpEngine.Open()
	if err != nil {
		return nil, err
	}
	return &Catalog{
		engine:      tmpEngine,
		metadataMgr: NewMetadataManager(tmpEngine),
	}, nil
}

// Init 初始化 Catalog，主要是检查并初始化系统表（sys_databases、sys_tables、…）。
func (c *Catalog) Init() error {
	if err := InitializeSystemTables(c.engine); err != nil {
		return fmt.Errorf("catalog initialization failed: %w", err)
	}

	// 创建默认数据库
	err := c.CreateDatabase("default")
	if err != nil && err.Error() != "database default already exists" {
		return fmt.Errorf("failed to create default database: %w", err)
	}

	return nil
}

// CreateDatabase 通过 MetadataManager 创建一个新的数据库记录。
func (c *Catalog) CreateDatabase(name string) error {
	if name == "" {
		return fmt.Errorf("database name is empty")
	}
	return c.metadataMgr.CreateDatabase(name)
}

// CreateTable 通过 MetadataManager 创建一个新的表记录。
func (c *Catalog) CreateTable(dbName string, table TableMeta) error {
	if dbName == "" || table.Table == "" {
		return fmt.Errorf("database name or table name is empty")
	}
	return c.metadataMgr.CreateTable(dbName, table)
}

// DeleteDatabase 通过 MetadataManager 删除指定数据库的记录。
func (c *Catalog) DeleteDatabase(name string) error {
	if name == "" {
		return fmt.Errorf("database name is empty")
	}
	return c.metadataMgr.DeleteDatabase(name)
}

// DeleteTable 通过 MetadataManager 删除指定表的记录。
func (c *Catalog) DeleteTable(dbName, tableName string) error {
	if dbName == "" || tableName == "" {
		return fmt.Errorf("database name or table name is empty")
	}
	return c.metadataMgr.DeleteTable(dbName, tableName)
}

// GetAllDatabases 获取所有数据库的元数据。
func (c *Catalog) GetAllDatabases() ([]DatabaseMeta, error) {
	return c.metadataMgr.GetAllDatabases()
}

// GetDatabase 读取指定数据库的元数据。
func (c *Catalog) GetDatabase(name string) (DatabaseMeta, error) {
	if name == "" {
		return DatabaseMeta{}, fmt.Errorf("database name is empty")
	}
	return c.metadataMgr.GetDatabase(name)
}

// GetTable 读取指定表的元数据。
func (c *Catalog) GetTable(dbName, tableName string) (TableMeta, error) {
	if dbName == "" || tableName == "" {
		return TableMeta{}, fmt.Errorf("database name or table name is empty")
	}
	return c.metadataMgr.GetTable(dbName, tableName)
}

// UpdateTable 更新指定表的元数据。
func (c *Catalog) UpdateTable(dbName string, table TableMeta) error {
	if dbName == "" || table.Table == "" {
		return fmt.Errorf("database name or table name is empty")
	}
	return c.metadataMgr.UpdateTable(dbName, table)
}

// GetEngine 获取存储引擎，供其他组件使用
func (c *Catalog) GetEngine() storage.Engine {
	return c.engine
}

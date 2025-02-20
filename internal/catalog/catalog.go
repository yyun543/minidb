package catalog

import (
	"fmt"

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

// Init 初始化 Catalog，主要是检查并初始化系统表（sys_databases、sys_tables、…）。
func (c *Catalog) Init() error {
	if err := InitializeSystemTables(c.engine); err != nil {
		return fmt.Errorf("catalog initialization failed: %w", err)
	}
	return nil
}

// CreateDatabase 通过 MetadataManager 创建一个新的数据库记录。
func (c *Catalog) CreateDatabase(name string) error {
	return c.metadataMgr.CreateDatabase(name)
}

// CreateTable 通过 MetadataManager 创建一个新的表记录。
func (c *Catalog) CreateTable(dbName string, table TableMeta) error {
	return c.metadataMgr.CreateTable(dbName, table)
}

// DeleteDatabase 通过 MetadataManager 删除指定数据库的记录。
func (c *Catalog) DeleteDatabase(name string) error {
	return c.metadataMgr.DeleteDatabase(name)
}

// DeleteTable 通过 MetadataManager 删除指定表的记录。
func (c *Catalog) DeleteTable(dbName, tableName string) error {
	return c.metadataMgr.DeleteTable(dbName, tableName)
}

// GetAllDatabases 获取所有数据库的元数据。
func (c *Catalog) GetAllDatabases() ([]DatabaseMeta, error) {
	return c.metadataMgr.GetAllDatabases()
}

// GetDatabase 读取指定数据库的元数据。
func (c *Catalog) GetDatabase(name string) (DatabaseMeta, error) {
	return c.metadataMgr.GetDatabase(name)
}

// GetTable 读取指定表的元数据。
func (c *Catalog) GetTable(dbName, tableName string) (TableMeta, error) {
	return c.metadataMgr.GetTable(dbName, tableName)
}

package catalog

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/yyun543/minidb/internal/storage"
)

// Catalog 元数据管理器
type Catalog struct {
	storage    *storage.MemTable
	mu         sync.RWMutex
	keyManager *storage.KeyManager // 新增 keyManager 字段
}

// NewCatalog 创建Catalog实例
func NewCatalog() (*Catalog, error) {
	store, err := storage.NewMemTable("data/catalog")
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %v", err)
	}

	cat := &Catalog{
		storage:    store,
		keyManager: storage.NewKeyManager(), // 初始化 keyManager
	}

	// 初始化系统表
	if err := cat.initSystemTables(); err != nil {
		err := store.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to init system tables: %v", err)
	}

	return cat, nil
}

// CreateDatabase 创建数据库
func (c *Catalog) CreateDatabase(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查数据库是否已存在
	if _, err := c.getDatabase(name); err == nil {
		return fmt.Errorf("database %s already exists", name)
	}

	// 创建数据库元数据
	db := &DatabaseMeta{
		ID:         time.Now().UnixNano(),
		Name:       name,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	// 使用 keyManager 生成key
	if err := c.storage.Put(c.keyManager.DatabaseKey(name), encodeDatabase(db)); err != nil {
		return fmt.Errorf("failed to store database metadata: %v", err)
	}

	// 写入系统表
	return c.insertIntoSysDatabases(db)
}

// GetDatabase 获取数据库元数据
func (c *Catalog) GetDatabase(name string) (*DatabaseMeta, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.getDatabase(name)
}

// CreateTable 创建表
func (c *Catalog) CreateTable(dbName string, table *TableMeta) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查数据库是否存在
	db, err := c.getDatabase(dbName)
	if err != nil {
		return fmt.Errorf("database %s not found", dbName)
	}

	// 检查表是否已存在
	if _, err := c.getTable(dbName, table.Name); err == nil {
		return fmt.Errorf("table %s already exists in database %s", table.Name, dbName)
	}

	// 设置数据库ID和时间戳
	table.DatabaseID = db.ID
	table.CreateTime = time.Now()
	table.UpdateTime = time.Now()

	// 写入存储
	key := c.keyManager.TableKey(dbName, table.Name)
	if err := c.storage.Put(key, encodeTable(table)); err != nil {
		return fmt.Errorf("failed to store table metadata: %v", err)
	}

	// 写入系统表
	return c.insertIntoSysTables(table)
}

// GetTable 获取表元数据
func (c *Catalog) GetTable(dbName, tableName string) (*TableMeta, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.getTable(dbName, tableName)
}

// DropTable 删除表
func (c *Catalog) DropTable(dbName, tableName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查表是否存在
	table, err := c.getTable(dbName, tableName)
	if err != nil {
		return err
	}

	// 不允许删除系统表
	if dbName == storage.SYS_DATABASE {
		return fmt.Errorf("cannot drop system table %s", tableName)
	}

	// 从存储中删除
	key := c.keyManager.TableKey(dbName, tableName)
	if err := c.storage.Delete(key); err != nil {
		return fmt.Errorf("failed to delete table metadata: %v", err)
	}

	// 从系统表中删除
	return c.deleteFromSysTables(table.ID)
}

// DropDatabase 删除数据库
func (c *Catalog) DropDatabase(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查数据库是否存在
	db, err := c.getDatabase(name)
	if err != nil {
		return fmt.Errorf("database %s not found", name)
	}

	// 不允许删除系统数据库
	if name == storage.SYS_DATABASE {
		return fmt.Errorf("cannot drop system database")
	}

	// 从存储中删除数据库元数据
	if err := c.storage.Delete(c.keyManager.DatabaseKey(name)); err != nil {
		return fmt.Errorf("failed to delete database metadata: %v", err)
	}

	// 从系统表中删除数据库记录
	if err := c.deleteFromSysDatabases(db.ID); err != nil {
		return fmt.Errorf("failed to delete from sys_databases: %v", err)
	}

	return nil
}

// Close 关闭Catalog
func (c *Catalog) Close() error {
	return c.storage.Close()
}

// 内部辅助方法

func (c *Catalog) getDatabase(name string) (*DatabaseMeta, error) {
	data, err := c.storage.Get(c.keyManager.DatabaseKey(name))
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("database %s not found", name)
	}
	return decodeDatabase(data)
}

func (c *Catalog) getTable(dbName, tableName string) (*TableMeta, error) {
	key := c.keyManager.TableKey(dbName, tableName)
	data, err := c.storage.Get(key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("table %s not found in database %s", tableName, dbName)
	}
	return decodeTable(data)
}

// 系统表操作

func (c *Catalog) insertIntoSysDatabases(db *DatabaseMeta) error {
	key := c.keyManager.TableKey(storage.SYS_DATABASE, db.Name)
	return c.storage.Put(key, encodeDatabase(db))
}

func (c *Catalog) insertIntoSysTables(table *TableMeta) error {
	key := c.keyManager.TableKey(storage.SYS_DATABASE, table.Name)
	return c.storage.Put(key, encodeTable(table))
}

func (c *Catalog) deleteFromSysTables(tableID int64) error {
	key := c.keyManager.SysTableKey(tableID)
	return c.storage.Delete(key)
}

// deleteFromSysDatabases 从系统表中删除数据库记录
func (c *Catalog) deleteFromSysDatabases(dbID int64) error {
	key := c.keyManager.TableKey(storage.SYS_DATABASE, fmt.Sprintf("%d", dbID))
	return c.storage.Delete(key)
}

// 编码/解码方法

func encodeDatabase(db *DatabaseMeta) []byte {
	data, _ := json.Marshal(db)
	return data
}

func decodeDatabase(data []byte) (*DatabaseMeta, error) {
	var db DatabaseMeta
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("failed to decode database metadata: %v", err)
	}
	return &db, nil
}

func encodeTable(table *TableMeta) []byte {
	data, _ := json.Marshal(table)
	return data
}

func decodeTable(data []byte) (*TableMeta, error) {
	var table TableMeta
	if err := json.Unmarshal(data, &table); err != nil {
		return nil, fmt.Errorf("failed to decode table metadata: %v", err)
	}
	return &table, nil
}

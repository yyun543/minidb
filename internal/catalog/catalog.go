package catalog

import (
	"fmt"
	"sync"
	"time"
)

// TODO Catalog 结构体，数据库/表管理

// Catalog 管理数据库元数据
type Catalog struct {
	mu sync.RWMutex

	// 数据库元数据
	databases map[string]*DatabaseMeta

	// 当前数据库
	currentDB string

	// 系统表
	systemTables *SystemTables

	// 表元数据缓存
	tableCache map[string]*TableMeta
}

// NewCatalog 创建新的Catalog实例
func NewCatalog() *Catalog {
	return &Catalog{
		databases:    make(map[string]*DatabaseMeta),
		systemTables: NewSystemTables(),
		tableCache:   make(map[string]*TableMeta),
	}
}

// CreateDatabase 创建数据库
func (c *Catalog) CreateDatabase(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.databases[name]; exists {
		return fmt.Errorf("数据库已存在: %s", name)
	}

	c.databases[name] = &DatabaseMeta{
		ID:         time.Now().UnixNano(),
		Name:       name,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	return nil
}

// UseDatabase 切换当前数据库
func (c *Catalog) UseDatabase(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.databases[name]; !exists {
		return fmt.Errorf("数据库不存在: %s", name)
	}

	c.currentDB = name
	return nil
}

// CreateTable 创建表
func (c *Catalog) CreateTable(table *TableMeta) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.currentDB == "" {
		return fmt.Errorf("未选择数据库")
	}

	db := c.databases[c.currentDB]
	if db == nil {
		return fmt.Errorf("当前数据库不存在")
	}

	// 检查表是否已存在
	for _, t := range db.Tables {
		if t.Name == table.Name {
			return fmt.Errorf("表已存在: %s", table.Name)
		}
	}

	// 添加表
	db.Tables = append(db.Tables, *table)
	c.tableCache[table.Name] = table
	db.UpdateTime = time.Now()

	return nil
}

// GetTable 获取表定义
func (c *Catalog) GetTable(name string) (*TableMeta, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 先查找系统表
	if sysTable, _ := c.systemTables.GetTable(name); sysTable != nil {
		return sysTable, nil
	}

	// 再查找用户表
	if table, ok := c.tableCache[name]; ok {
		return table, nil
	}

	return nil, fmt.Errorf("表不存在: %s", name)
}

// DropTable 删除表
func (c *Catalog) DropTable(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.currentDB == "" {
		return fmt.Errorf("未选择数据库")
	}

	db := c.databases[c.currentDB]
	if db == nil {
		return fmt.Errorf("当前数据库不存在")
	}

	// 删除表
	for i, t := range db.Tables {
		if t.Name == name {
			db.Tables = append(db.Tables[:i], db.Tables[i+1:]...)
			delete(c.tableCache, name)
			db.UpdateTime = time.Now()
			return nil
		}
	}

	return fmt.Errorf("表不存在: %s", name)
}

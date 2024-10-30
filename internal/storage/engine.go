package storage

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Row 表示表中的一行数据
type Row struct {
	ID      int                    // 行ID
	Data    map[string]interface{} // 列数据
	Created time.Time              // 创建时间
	Updated time.Time              // 更新时间
}

// Column 表示表的列定义
type Column struct {
	Type     string // 数据类型
	Nullable bool   // 是否可为空
}

// Schema 表示表结构
type Schema map[string]Column

// Table 表示数据表
type Table struct {
	Name    string
	Schema  Schema
	Rows    map[int]Row
	LastID  int
	mu      sync.RWMutex
	rowLock sync.Map
}

// 实现行级锁
func (t *Table) lockRow(id int) func() {
	value, _ := t.rowLock.LoadOrStore(id, &sync.Mutex{})
	mu := value.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// 使用行级锁的更新操作
func (t *Table) UpdateRow(id int, values map[string]interface{}) error {
	unlock := t.lockRow(id)
	defer unlock()

	row, exists := t.Rows[id]
	if !exists {
		return fmt.Errorf("row %d does not exist", id)
	}

	// 更新数据
	for k, v := range values {
		row.Data[k] = v
	}
	row.Updated = time.Now()
	t.Rows[id] = row

	return nil
}

// 添加死锁检测
type DeadlockDetector struct {
	waitFor sync.Map // 记录等待关系
	mu      sync.Mutex
}

func (d *DeadlockDetector) CheckDeadlock(from, to int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	visited := make(map[int]bool)
	return d.dfs(to, from, visited)
}

func (d *DeadlockDetector) dfs(current, target int, visited map[int]bool) bool {
	if current == target {
		return true
	}

	visited[current] = true
	if waitList, ok := d.waitFor.Load(current); ok {
		for _, next := range waitList.([]int) {
			if !visited[next] && d.dfs(next, target, visited) {
				return true
			}
		}
	}
	return false
}

// StorageEngine 定义存储引擎接口
type StorageEngine interface {
	CreateTable(name string, schema Schema) error
	DropTable(name string) error
	Insert(tableName string, values map[string]interface{}) error
	Select(tableName string, columns []string, where string) ([]Row, error)
	Update(tableName string, values map[string]interface{}, where string) (int, error)
	Delete(tableName string, where string) (int, error)
	Backup() ([]byte, error)
	Restore(data []byte) error
}

// Engine 混合存储引擎，同时支持行存储和列存储
type Engine struct {
	rowStore    *RowStore    // OLTP行存储
	columnStore *ColumnStore // OLAP列存储
	tables      map[string]*Table
	mu          sync.RWMutex
}

// NewEngine 创建新的混合存储引擎
func NewEngine() (*Engine, error) {
	rowStore, err := NewRowStore("data/row_store.db")
	if err != nil {
		return nil, fmt.Errorf("failed to create row store: %v", err)
	}

	columnStore, err := NewColumnStore("data/column_store.db")
	if err != nil {
		return nil, fmt.Errorf("failed to create column store: %v", err)
	}

	return &Engine{
		rowStore:    rowStore,
		columnStore: columnStore,
		tables:      make(map[string]*Table),
	}, nil
}

// CreateTable 创建新表
func (e *Engine) CreateTable(name string, schema Schema) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.tables[name]; exists {
		return fmt.Errorf("table %s already exists", name)
	}

	e.tables[name] = &Table{
		Name:   name,
		Schema: schema,
		Rows:   make(map[int]Row),
	}
	return nil
}

// DropTable 删除表
func (e *Engine) DropTable(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.tables[name]; !exists {
		return fmt.Errorf("table %s does not exist", name)
	}

	delete(e.tables, name)
	return nil
}

// Insert 插入数据
func (e *Engine) Insert(tableName string, values map[string]interface{}) error {
	tx, err := e.Begin()
	if err != nil {
		return err
	}

	err = tx.Insert(tableName, values)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Select 智能选择存储引擎执行查询
func (e *Engine) Select(tableName string, columns []string, where string) ([]Row, error) {
	// 分析查询特征，选择合适的存储引擎
	if isAnalyticalQuery(columns, where) {
		// OLAP查询使用列存储
		return e.columnStore.Select(tableName, columns, where)
	} else {
		// OLTP查询使用行存储
		return e.rowStore.Select(tableName, columns, where)
	}
}

// isAnalyticalQuery 判断是否是分析型查询
func isAnalyticalQuery(columns []string, where string) bool {
	// 简单判断逻辑，实际应该更复杂
	// 1. 是否包含聚合函数
	// 2. 是否查询大量列
	// 3. 是否有复杂的WHERE条件
	// 4. 是否包含GROUP BY
	return len(columns) > 5 || strings.Contains(where, "GROUP BY")
}

// Update 更新数据
func (e *Engine) Update(tableName string, values map[string]interface{}, where string) (int, error) {
	e.mu.RLock()
	table, exists := e.tables[tableName]
	e.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("table %s does not exist", tableName)
	}

	// 验证数据类型
	if err := e.validateValues(table.Schema, values); err != nil {
		return 0, err
	}

	table.mu.Lock()
	defer table.mu.Unlock()

	updateCount := 0
	now := time.Now()

	// 遍历所有行
	for id, row := range table.Rows {
		// 如果有WHERE条件，进行过滤
		if where != "" && !e.evaluateWhere(row, where) {
			continue
		}

		// 更新数据
		for key, value := range values {
			row.Data[key] = value
		}
		row.Updated = now
		table.Rows[id] = row
		updateCount++
	}

	return updateCount, nil
}

// Delete 删除数据
func (e *Engine) Delete(tableName string, where string) (int, error) {
	e.mu.RLock()
	table, exists := e.tables[tableName]
	e.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("table %s does not exist", tableName)
	}

	table.mu.Lock()
	defer table.mu.Unlock()

	deleteCount := 0

	// 遍历所有行
	for id, row := range table.Rows {
		// 如果有WHERE条件，进行过滤
		if where != "" && !e.evaluateWhere(row, where) {
			continue
		}

		delete(table.Rows, id)
		deleteCount++
	}

	return deleteCount, nil
}

// 辅助方法

// validateValues 验证数据类型
func (e *Engine) validateValues(schema Schema, values map[string]interface{}) error {
	for col, val := range values {
		colDef, exists := schema[col]
		if !exists {
			return fmt.Errorf("column %s does not exist", col)
		}

		if val == nil && !colDef.Nullable {
			return fmt.Errorf("column %s cannot be null", col)
		}

		if val != nil {
			if err := e.validateType(colDef.Type, val); err != nil {
				return fmt.Errorf("invalid value for column %s: %v", col, err)
			}
		}
	}
	return nil
}

// validateType 验证数据类型
func (e *Engine) validateType(expectedType string, value interface{}) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "int":
		if _, ok := value.(int); !ok {
			return fmt.Errorf("expected int, got %T", value)
		}
	case "float":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("expected float, got %T", value)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
	default:
		return fmt.Errorf("unsupported type: %s", expectedType)
	}
	return nil
}

// evaluateWhere 评估WHERE条件
func (e *Engine) evaluateWhere(row Row, condition string) bool {
	// 简化的WHERE条件评估
	// 实际实现应该解析和评估复杂的条件表达式
	return true
}

// Backup 创建数据库备份
func (e *Engine) Backup() ([]byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return json.Marshal(e.tables)
}

// Restore 从备份恢复数据库
func (e *Engine) Restore(data []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return json.Unmarshal(data, &e.tables)
}

// RowStore 行存储引擎
type RowStore struct {
	path string
	// ... 其他字段
}

// NewRowStore 创建新的行存储引擎
func NewRowStore(path string) (*RowStore, error) {
	return &RowStore{path: path}, nil
}

// ColumnStore 列存储引擎
type ColumnStore struct {
	path string
	// ... 其他字段
}

// NewColumnStore 创建新的列存储引擎
func NewColumnStore(path string) (*ColumnStore, error) {
	return &ColumnStore{path: path}, nil
}

// 添加事务支持
type Transaction struct {
	id      int64
	changes map[string][]Row
	engine  *Engine
	mu      sync.RWMutex
}

// 开始事务
func (e *Engine) Begin() (*Transaction, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	txID := time.Now().UnixNano()
	return &Transaction{
		id:      txID,
		changes: make(map[string][]Row),
		engine:  e,
	}, nil
}

// 实现完整的CRUD操作
func (tx *Transaction) Insert(tableName string, values map[string]interface{}) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	table, exists := tx.engine.tables[tableName]
	if !exists {
		return fmt.Errorf("table %s does not exist", tableName)
	}

	// 验证数据
	if err := tx.engine.validateValues(table.Schema, values); err != nil {
		return err
	}

	// 生成新行
	row := Row{
		ID:      table.LastID + 1,
		Data:    values,
		Created: time.Now(),
		Updated: time.Now(),
	}

	// 记录变更
	tx.changes[tableName] = append(tx.changes[tableName], row)
	return nil
}

// 添加批量操作支持
type Batch struct {
	operations []operation
	engine     *Engine
}

type operation struct {
	op    string
	table string
	data  interface{}
}

func (e *Engine) NewBatch() *Batch {
	return &Batch{
		engine: e,
	}
}

func (b *Batch) Insert(table string, values map[string]interface{}) {
	b.operations = append(b.operations, operation{
		op:    "insert",
		table: table,
		data:  values,
	})
}

func (b *Batch) Execute() error {
	tx, err := b.engine.Begin()
	if err != nil {
		return err
	}

	for _, op := range b.operations {
		switch op.op {
		case "insert":
			if err := tx.Insert(op.table, op.data.(map[string]interface{})); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

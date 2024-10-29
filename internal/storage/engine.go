package storage

import (
	"fmt"
	"strings"
	"sync"
)

type Engine struct {
	Tables      map[string]*Table       // 行存储
	ColumnStore map[string]*ColumnStore // 列存储
	mu          sync.RWMutex
}

func NewEngine() *Engine {
	return &Engine{
		Tables:      make(map[string]*Table),
		ColumnStore: make(map[string]*ColumnStore),
	}
}

func (e *Engine) CreateTable(name string, schema Row) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	name = strings.ToLower(name)
	if _, exists := e.Tables[name]; exists {
		return fmt.Errorf("table %s already exists", name)
	}

	e.Tables[name] = NewTable(schema)
	e.ColumnStore[name] = NewColumnStore(schema)
	return nil
}

func (e *Engine) DropTable(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	name = strings.ToLower(name)
	if _, exists := e.Tables[name]; !exists {
		return fmt.Errorf("table %s does not exist", name)
	}

	delete(e.Tables, name)
	delete(e.ColumnStore, name)
	return nil
}

func (e *Engine) Insert(table string, values []string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	table = strings.ToLower(table)
	rowStore, exists := e.Tables[table]
	if !exists {
		return fmt.Errorf("table %s does not exist", table)
	}

	// 插入到行存储
	if err := rowStore.Insert(values); err != nil {
		return err
	}

	// 同步到列存储
	colStore := e.ColumnStore[table]
	return colStore.Insert(values)
}

func (e *Engine) Select(table string, columns []string, where string, isAnalytical bool) ([]Row, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	table = strings.ToLower(table)

	if isAnalytical {
		store, exists := e.ColumnStore[table]
		if !exists {
			return nil, fmt.Errorf("table %s does not exist", table)
		}
		return store.Select(columns, where)
	}

	store, exists := e.Tables[table]
	if !exists {
		return nil, fmt.Errorf("table %s does not exist", table)
	}
	return store.Select(columns, where)
}

func (e *Engine) Update(table, column, value, where string) (int, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	table = strings.ToLower(table)
	rowStore, exists := e.Tables[table]
	if !exists {
		return 0, fmt.Errorf("table %s does not exist", table)
	}

	// 更新行存储
	count, err := rowStore.Update(column, value, where)
	if err != nil {
		return 0, err
	}

	// 同步到列存储
	colStore := e.ColumnStore[table]
	_, err = colStore.Update(column, value, where)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (e *Engine) Delete(table string, where string) (int, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	table = strings.ToLower(table)
	rowStore, exists := e.Tables[table]
	if !exists {
		return 0, fmt.Errorf("table %s does not exist", table)
	}

	// 从行存储删除
	count, err := rowStore.Delete(where)
	if err != nil {
		return 0, err
	}

	// 同步到列存储
	colStore := e.ColumnStore[table]
	_, err = colStore.Delete(where)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (e *Engine) GetTables() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	tables := make([]string, 0, len(e.Tables))
	for name := range e.Tables {
		tables = append(tables, name)
	}
	return tables
}

type Transaction struct {
	engine     *Engine
	operations []Operation
	mu         sync.Mutex
}

func (e *Engine) Begin() *Transaction {
	return &Transaction{
		engine:     e,
		operations: make([]Operation, 0),
	}
}

func (tx *Transaction) Commit() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	// 执行所有操作
	for _, op := range tx.operations {
		if err := op.Execute(tx.engine); err != nil {
			return err
		}
	}
	return nil
}

func (tx *Transaction) Rollback() {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	tx.operations = nil
}

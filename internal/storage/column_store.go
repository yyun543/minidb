package storage

import (
	"fmt"
	"sync"
)

// ColumnStore 实现基于列的存储引擎
type ColumnStore struct {
	store  *FileStore
	tables map[string]*ColumnTable
	mu     sync.RWMutex
}

// ColumnTable 列式存储的表结构
type ColumnTable struct {
	Name    string
	Schema  Schema
	Columns map[string]*Column
}

// Column 列数据
type Column struct {
	Name   string
	Type   string
	Values []interface{}
}

// NewColumnStore 创建新的列存储引擎
func NewColumnStore(path string) (*ColumnStore, error) {
	store, err := NewFileStore(path)
	if err != nil {
		return nil, err
	}

	return &ColumnStore{
		store:  store,
		tables: make(map[string]*ColumnTable),
	}, nil
}

// Insert 插入列数据
func (cs *ColumnStore) Insert(tableName string, values map[string]interface{}) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	table, exists := cs.tables[tableName]
	if !exists {
		return fmt.Errorf("table %s does not exist", tableName)
	}

	// 验证数据
	if err := validateValues(table.Schema, values); err != nil {
		return err
	}

	// 将数据添加到对应的列
	for colName, value := range values {
		col := table.Columns[colName]
		col.Values = append(col.Values, value)
	}

	return nil
}

// Select 查询列数据
func (cs *ColumnStore) Select(tableName string, columns []string, where string) ([]Row, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	table, exists := cs.tables[tableName]
	if !exists {
		return nil, fmt.Errorf("table %s does not exist", tableName)
	}

	// 如果没有指定列，返回所有列
	if len(columns) == 0 || (len(columns) == 1 && columns[0] == "*") {
		columns = make([]string, 0, len(table.Schema))
		for col := range table.Schema {
			columns = append(columns, col)
		}
	}

	// 获取满足条件的行索引
	rowIndices := cs.evaluateWhere(table, where)

	// 构建结果集
	var result []Row
	for _, rowIdx := range rowIndices {
		row := Row{
			ID:   rowIdx,
			Data: make(map[string]interface{}),
		}
		for _, colName := range columns {
			if col, ok := table.Columns[colName]; ok {
				row.Data[colName] = col.Values[rowIdx]
			}
		}
		result = append(result, row)
	}

	return result, nil
}

// 其他方法实现...

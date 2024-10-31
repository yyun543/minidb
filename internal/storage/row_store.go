package storage

import (
	"sync"
)

// RowTable 实现基于行的存储引擎
type RowTable struct {
	Name   string
	Schema Schema
	Rows   map[int]Row
	LastID int
	mu     sync.RWMutex
	store  *FileStore
}

// NewRowTable 创建新的行存储引擎
func NewRowTable(name string, schema Schema) *RowTable {
	return &RowTable{
		Name:   name,
		Schema: schema,
		Rows:   make(map[int]Row),
		LastID: 0,
	}
}

// Insert 插入行数据
func (rt *RowTable) Insert(values map[string]interface{}) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	// 验证和转换数据
	if err := validateValues(rt.Schema, values); err != nil {
		return err
	}

	// 生成新的行ID
	rt.LastID++
	rowID := rt.LastID

	// 创建新行
	rt.Rows[rowID] = Row{
		ID:     rowID,
		Values: values,
	}

	return nil
}

// 其他方法实现...

package storage

import (
	"encoding/json"
	"fmt"
	"sync"
)

// RowStore 实现基于行的存储引擎
type RowStore struct {
	store  *FileStore
	tables map[string]*Table
	mu     sync.RWMutex
}

// NewRowStore 创建新的行存储引擎
func NewRowStore(path string) (*RowStore, error) {
	store, err := NewFileStore(path)
	if err != nil {
		return nil, err
	}

	return &RowStore{
		store:  store,
		tables: make(map[string]*Table),
	}, nil
}

// Insert 插入行数据
func (rs *RowStore) Insert(tableName string, values map[string]interface{}) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	table, exists := rs.tables[tableName]
	if !exists {
		return fmt.Errorf("table %s does not exist", tableName)
	}

	// 验证和转换数据
	if err := validateValues(table.Schema, values); err != nil {
		return err
	}

	// 序列化行数据
	rowData, err := json.Marshal(values)
	if err != nil {
		return err
	}

	// 写入存储
	return rs.store.Put([]byte(fmt.Sprintf("%s:%d", tableName, table.LastID+1)), rowData)
}

// 其他方法实现...

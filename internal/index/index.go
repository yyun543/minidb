package index

import (
	"sync"
)

type Index struct {
	data map[string][]int // 值到行号的映射
	mu   sync.RWMutex
}

func NewIndex() *Index {
	return &Index{
		data: make(map[string][]int),
	}
}

func (idx *Index) Add(value string, rowID int) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if _, exists := idx.data[value]; !exists {
		idx.data[value] = make([]int, 0)
	}
	idx.data[value] = append(idx.data[value], rowID)
}

func (idx *Index) Find(value string) []int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if rows, exists := idx.data[value]; exists {
		result := make([]int, len(rows))
		copy(result, rows)
		return result
	}
	return nil
}

func (idx *Index) Remove(value string, rowID int) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if rows, exists := idx.data[value]; exists {
		newRows := make([]int, 0)
		for _, id := range rows {
			if id != rowID {
				newRows = append(newRows, id)
			}
		}
		if len(newRows) == 0 {
			delete(idx.data, value)
		} else {
			idx.data[value] = newRows
		}
	}
}

type IndexManager struct {
	indexes map[string]*Index // 表名.列名 到索引的映射
	mu      sync.RWMutex
}

func NewIndexManager() *IndexManager {
	return &IndexManager{
		indexes: make(map[string]*Index),
	}
}

func (im *IndexManager) CreateIndex(tableName, columnName string) *Index {
	im.mu.Lock()
	defer im.mu.Unlock()

	key := tableName + "." + columnName
	if _, exists := im.indexes[key]; !exists {
		im.indexes[key] = NewIndex()
	}
	return im.indexes[key]
}

func (im *IndexManager) GetIndex(tableName, columnName string) (*Index, bool) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	key := tableName + "." + columnName
	idx, exists := im.indexes[key]
	return idx, exists
}

func (im *IndexManager) DropIndex(tableName, columnName string) {
	im.mu.Lock()
	defer im.mu.Unlock()

	key := tableName + "." + columnName
	delete(im.indexes, key)
}

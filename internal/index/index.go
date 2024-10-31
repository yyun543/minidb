package index

import (
	"strings"
	"sync"

	"github.com/yyun543/minidb/internal/parser"
)

// Index 表示单个索引的数据结构
type Index struct {
	name   string           // 索引名称
	table  string           // 表名
	column string           // 列名
	values map[string][]int // 索引数据: 值 -> 行ID列表
	mu     sync.RWMutex     // 用于并发控制的读写锁
}

// Manager 管理数据库中的所有索引
type Manager struct {
	indexes map[string]*Index // 所有索引的映射表 (表名.列名 -> 索引)
	mu      sync.RWMutex      // 用于并发控制的读写锁
}

// NewManager 创建新的索引管理器
func NewManager() *Manager {
	return &Manager{
		indexes: make(map[string]*Index),
	}
}

// CreateIndex 创建新索引
func (m *Manager) CreateIndex(table, column string) *Index {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := makeKey(table, column)
	if idx, exists := m.indexes[key]; exists {
		return idx
	}

	idx := &Index{
		name:   key,
		table:  table,
		column: column,
		values: make(map[string][]int),
	}
	m.indexes[key] = idx
	return idx
}

// GetIndex 获取指定的索引
func (m *Manager) GetIndex(table, column string) (*Index, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	idx, exists := m.indexes[makeKey(table, column)]
	return idx, exists
}

// DropIndex 删除索引
func (m *Manager) DropIndex(table, column string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.indexes, makeKey(table, column))
}

// FindBestIndex 为查询选择最合适的索引
func (m *Manager) FindBestIndex(table string, where parser.Expression) *Index {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 简化的索引选择逻辑
	// 实际实现应该分析WHERE子句并选择最优索引
	for key, idx := range m.indexes {
		if strings.HasPrefix(key, table+".") && strings.Contains(where.String(), idx.column) {
			return idx
		}
	}
	return nil
}

// GetTableIndexes 获取表的所有索引
func (m *Manager) GetTableIndexes(table string) []*Index {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Index
	for _, idx := range m.indexes {
		if idx.table == table {
			result = append(result, idx)
		}
	}
	return result
}

// Index 方法

// Add 向索引添加一个值
func (idx *Index) Add(value string, rowID int) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if _, exists := idx.values[value]; !exists {
		idx.values[value] = make([]int, 0)
	}
	idx.values[value] = append(idx.values[value], rowID)
}

// Remove 从索引中移除一个值
func (idx *Index) Remove(value string, rowID int) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if rows, exists := idx.values[value]; exists {
		newRows := make([]int, 0, len(rows)-1)
		for _, id := range rows {
			if id != rowID {
				newRows = append(newRows, id)
			}
		}
		if len(newRows) == 0 {
			delete(idx.values, value)
		} else {
			idx.values[value] = newRows
		}
	}
}

// Find 在索引中查找值对应的行ID
func (idx *Index) Find(value string) []int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if rows, exists := idx.values[value]; exists {
		result := make([]int, len(rows))
		copy(result, rows)
		return result
	}
	return nil
}

// Clear 清空索引
func (idx *Index) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.values = make(map[string][]int)
}

// Update 批量更新索引
func (idx *Index) Update(values []string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// 清空现有索引
	idx.Clear()

	// 重建索引
	for i, value := range values {
		idx.Add(value, i)
	}

	return nil
}

// BatchUpdate 批量更新多个值
func (idx *Index) BatchUpdate(updates map[int]string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for rowID, newValue := range updates {
		// 找到并删除旧值
		for value, rows := range idx.values {
			for i, id := range rows {
				if id == rowID {
					// 删除旧值
					idx.values[value] = append(rows[:i], rows[i+1:]...)
					break
				}
			}
		}
		// 添加新值
		idx.Add(newValue, rowID)
	}

	return nil
}

// 辅助函数

// makeKey 生成索引的键名
func makeKey(table, column string) string {
	return table + "." + column
}

// 用于范围查询的数据结构
type IndexRange struct {
	Start     string
	End       string
	Inclusive bool
}

// FindRange 在索引中执行范围查询
func (idx *Index) FindRange(r IndexRange) []int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var result []int
	for value, rows := range idx.values {
		if (r.Start == "" || value >= r.Start) &&
			(r.End == "" || (r.Inclusive && value <= r.End) || (!r.Inclusive && value < r.End)) {
			result = append(result, rows...)
		}
	}
	return result
}

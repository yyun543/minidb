package storage

import (
	"bytes"
	"sync"

	"github.com/google/btree"
)

// IndexItem 表示B树中的索引项
type IndexItem struct {
	Key   []byte
	Value []byte
}

// Less 实现 btree.Item 接口，用于B树中的项比较
func (i *IndexItem) Less(than btree.Item) bool {
	return bytes.Compare(i.Key, than.(*IndexItem).Key) < 0
}

// Index 实现基于B树的索引结构
type Index struct {
	tree  *btree.BTree
	mutex sync.RWMutex
}

// NewIndex 创建新的索引实例
func NewIndex() *Index {
	return &Index{
		tree: btree.New(32), // 默认度数为32，可根据实际需求调整
	}
}

// Put 插入或更新索引项
func (idx *Index) Put(key, value []byte) {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	item := &IndexItem{
		Key:   append([]byte(nil), key...),   // 复制key
		Value: append([]byte(nil), value...), // 复制value
	}
	idx.tree.ReplaceOrInsert(item)
}

// Get 获取指定key的索引项
func (idx *Index) Get(key []byte) ([]byte, bool) {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	item := idx.tree.Get(&IndexItem{Key: key})
	if item == nil {
		return nil, false
	}
	return item.(*IndexItem).Value, true
}

// Delete 删除指定key的索引项
func (idx *Index) Delete(key []byte) bool {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	item := idx.tree.Delete(&IndexItem{Key: key})
	return item != nil
}

// Range 范围查询，返回指定范围内的所有索引项
func (idx *Index) Range(start, end []byte) []*IndexItem {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	var result []*IndexItem
	idx.tree.AscendRange(&IndexItem{Key: start}, &IndexItem{Key: end}, func(i btree.Item) bool {
		item := i.(*IndexItem)
		result = append(result, &IndexItem{
			Key:   append([]byte(nil), item.Key...),
			Value: append([]byte(nil), item.Value...),
		})
		return true
	})
	return result
}

// Clear 清空索引
func (idx *Index) Clear() {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	idx.tree.Clear(false)
}

// Len 返回索引项数量
func (idx *Index) Len() int {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	return idx.tree.Len()
}

// Iterator 返回一个索引迭代器
type Iterator struct {
	items []*IndexItem
	pos   int
}

// NewIterator 创建新的迭代器
func (idx *Index) NewIterator(start, end []byte) *Iterator {
	return &Iterator{
		items: idx.Range(start, end),
		pos:   -1,
	}
}

// Next 移动到下一个位置
func (it *Iterator) Next() bool {
	if it.pos+1 >= len(it.items) {
		return false
	}
	it.pos++
	return true
}

// Key 返回当前位置的key
func (it *Iterator) Key() []byte {
	if it.pos < 0 || it.pos >= len(it.items) {
		return nil
	}
	return it.items[it.pos].Key
}

// Value 返回当前位置的value
func (it *Iterator) Value() []byte {
	if it.pos < 0 || it.pos >= len(it.items) {
		return nil
	}
	return it.items[it.pos].Value
}

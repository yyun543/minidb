package storage

import (
	"bytes"
	"sync"

	"github.com/google/btree"
)

// TODO 内存表

// MemTable 内存表
type MemTable struct {
	tree    *btree.BTree // B树存储
	wal     *WAL         // 预写日志
	rwMutex sync.RWMutex // 读写锁
}

// memTableItem B树节点项
type memTableItem struct {
	key   []byte
	value []byte
}

// Less 实现btree.Item接口
func (i *memTableItem) Less(than btree.Item) bool {
	return bytes.Compare(i.key, than.(*memTableItem).key) < 0
}

// NewMemTable 创建内存表
func NewMemTable(walPath string) (*MemTable, error) {
	// 创建WAL
	wal, err := NewWAL(walPath)
	if err != nil {
		return nil, err
	}

	mt := &MemTable{
		tree: btree.New(32), // 32阶B树
		wal:  wal,
	}

	// 从WAL恢复数据
	if err := mt.recover(); err != nil {
		return nil, err
	}

	return mt, nil
}

// Get 获取key对应的值
func (mt *MemTable) Get(key []byte) ([]byte, error) {
	mt.rwMutex.RLock()
	defer mt.rwMutex.RUnlock()

	item := mt.tree.Get(&memTableItem{key: key})
	if item == nil {
		return nil, nil
	}
	return item.(*memTableItem).value, nil
}

// Put 写入key-value对
func (mt *MemTable) Put(key []byte, value []byte) error {
	// 先写WAL
	if err := mt.wal.Write(WAL_PUT, key, value); err != nil {
		return err
	}

	// 再更新内存
	mt.rwMutex.Lock()
	mt.tree.ReplaceOrInsert(&memTableItem{
		key:   key,
		value: value,
	})
	mt.rwMutex.Unlock()

	return nil
}

// Delete 删除key对应的数据
func (mt *MemTable) Delete(key []byte) error {
	// 先写WAL
	if err := mt.wal.Write(WAL_DELETE, key, nil); err != nil {
		return err
	}

	// 再更新内存
	mt.rwMutex.Lock()
	mt.tree.Delete(&memTableItem{key: key})
	mt.rwMutex.Unlock()

	return nil
}

// Scan 范围扫描
func (mt *MemTable) Scan(start []byte, end []byte) (Iterator, error) {
	mt.rwMutex.RLock()
	defer mt.rwMutex.RUnlock()

	var items []*memTableItem
	mt.tree.AscendRange(&memTableItem{key: start}, &memTableItem{key: end}, func(i btree.Item) bool {
		items = append(items, i.(*memTableItem))
		return true
	})

	return &memTableIterator{
		items:    items,
		position: -1,
	}, nil
}

// Close 关闭内存表
func (mt *MemTable) Close() error {
	return mt.wal.Close()
}

// recover 从WAL恢复数据
func (mt *MemTable) recover() error {
	records, err := mt.wal.Read()
	if err != nil {
		return err
	}

	for _, record := range records {
		switch record.Type {
		case WAL_PUT:
			mt.tree.ReplaceOrInsert(&memTableItem{
				key:   record.Key,
				value: record.Value,
			})
		case WAL_DELETE:
			mt.tree.Delete(&memTableItem{key: record.Key})
		}
	}

	return nil
}

// memTableIterator 内存表迭代器
type memTableIterator struct {
	items    []*memTableItem
	position int
}

func (it *memTableIterator) Next() bool {
	it.position++
	return it.position < len(it.items)
}

func (it *memTableIterator) Key() []byte {
	if it.position >= len(it.items) {
		return nil
	}
	return it.items[it.position].key
}

func (it *memTableIterator) Value() []byte {
	if it.position >= len(it.items) {
		return nil
	}
	return it.items[it.position].value
}

func (it *memTableIterator) Close() error {
	return nil
}

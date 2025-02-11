package storage

import (
	"bytes"
	"encoding/binary"
	"sync"

	"github.com/google/btree"
)

// TODO 索引 (BTree)

// Index B+树索引实现
type Index struct {
	tree    *btree.BTree // B树存储
	rwMutex sync.RWMutex // 读写锁
}

// indexItem B树节点项
type indexItem struct {
	key   []byte // 索引键
	value []byte // 指向数据的指针
}

// Less 实现btree.Item接口
func (i *indexItem) Less(than btree.Item) bool {
	return bytes.Compare(i.key, than.(*indexItem).key) < 0
}

// NewIndex 创建索引实例
func NewIndex() *Index {
	return &Index{
		tree: btree.New(32), // 32阶B树
	}
}

// Put 写入索引项
func (idx *Index) Put(key []byte, value []byte) error {
	idx.rwMutex.Lock()
	defer idx.rwMutex.Unlock()

	idx.tree.ReplaceOrInsert(&indexItem{
		key:   key,
		value: value,
	})
	return nil
}

// Get 获取索引项
func (idx *Index) Get(key []byte) ([]byte, error) {
	idx.rwMutex.RLock()
	defer idx.rwMutex.RUnlock()

	item := idx.tree.Get(&indexItem{key: key})
	if item == nil {
		return nil, nil
	}
	return item.(*indexItem).value, nil
}

// Delete 删除索引项
func (idx *Index) Delete(key []byte) error {
	idx.rwMutex.Lock()
	defer idx.rwMutex.Unlock()

	idx.tree.Delete(&indexItem{key: key})
	return nil
}

// Scan 范围扫描
func (idx *Index) Scan(start []byte, end []byte) (Iterator, error) {
	idx.rwMutex.RLock()
	defer idx.rwMutex.RUnlock()

	var items []*indexItem
	idx.tree.AscendRange(&indexItem{key: start}, &indexItem{key: end}, func(i btree.Item) bool {
		items = append(items, i.(*indexItem))
		return true
	})

	return &indexIterator{
		items:    items,
		position: -1,
	}, nil
}

// indexIterator 索引迭代器
type indexIterator struct {
	items    []*indexItem
	position int
}

func (it *indexIterator) Next() bool {
	it.position++
	return it.position < len(it.items)
}

func (it *indexIterator) Key() []byte {
	if it.position >= len(it.items) {
		return nil
	}
	return it.items[it.position].key
}

func (it *indexIterator) Value() []byte {
	if it.position >= len(it.items) {
		return nil
	}
	return it.items[it.position].value
}

func (it *indexIterator) Close() error {
	return nil
}

// EncodeIndexKey 编码索引键
func EncodeIndexKey(values ...interface{}) []byte {
	var buf bytes.Buffer
	for _, v := range values {
		switch val := v.(type) {
		case int64:
			binary.Write(&buf, binary.BigEndian, val)
		case string:
			buf.WriteString(val)
		case []byte:
			buf.Write(val)
		}
	}
	return buf.Bytes()
}

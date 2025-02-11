package storage

import (
	"github.com/apache/arrow/go/v18/arrow"
)

// Engine 存储引擎接口
type Engine interface {
	// Open 打开存储引擎
	Open() error

	// Close 关闭存储引擎
	Close() error

	// Get 获取指定key的值
	Get(key []byte) ([]byte, error)

	// Put 写入key-value对
	Put(key []byte, value []byte) error

	// Delete 删除指定key的数据
	Delete(key []byte) error

	// Scan 范围扫描
	Scan(start []byte, end []byte) (Iterator, error)

	// Flush 将内存表刷新到磁盘
	Flush() error
}

// Iterator 迭代器接口
type Iterator interface {
	// Next 移动到下一个位置
	Next() bool

	// Key 获取当前位置的key
	Key() []byte

	// Value 获取当前位置的value
	Value() []byte

	// Close 关闭迭代器
	Close() error
}

// Record 表示一条记录
type Record struct {
	Key   []byte
	Value []byte
}

// Table 表示一个数据表
type Table struct {
	Name    string
	Schema  *arrow.Schema
	Records []*Record
}

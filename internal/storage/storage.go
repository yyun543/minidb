package storage

import (
	"github.com/apache/arrow/go/v18/arrow"
)

/*
Package storage 定义了存储引擎接口及相关类型，专门用于管理 Apache Arrow 格式的记录。

Apache Arrow 最佳实践：

1. 内存管理：Arrow 相关结构（如 Record、Array 等）背后可能涉及大量内存分配，
   因此调用者在数据使用完毕后应调用 Record.Release() 释放资源，避免内存泄露。

2. 不可变性：箭头数据结构（schema 和 record）通常是不可变的，一旦创建后应保持不变，
   如需修改请新建一个 Record。

3. 列式访问：充分利用 Arrow 的列式格式优势，结合批处理和矢量化计算提升性能。

所有实现该接口的存储引擎需遵循以上原则，以确保数据合理管理和高效处理。
*/

// Range 定义了扫描范围
type Range struct {
	Start []byte // 扫描起始key（包含）
	End   []byte // 扫描结束key（不包含）
}

// Engine 定义了存储引擎接口，支持按照 Apache Arrow 格式读取和写入数据
type Engine interface {
	// Open 打开存储引擎
	Open() error

	// Close 关闭存储引擎
	Close() error

	// Get 获取指定 key 对应的 Arrow 记录。
	// IMPORTANT: 调用者应在处理完返回的 Record 后调用 Record.Release() 释放资源！
	Get(key []byte) (*arrow.Record, error)

	// Put 写入 Arrow 记录到指定 key。
	// 注意：通常建议对传入的 record 转移所有权，存储引擎在持久化后可能调用 Release() 释放内存。
	Put(key []byte, record *arrow.Record) error

	// Delete 删除指定 key 的数据
	Delete(key []byte) error

	// Scan 范围扫描并返回 Arrow 记录迭代器。
	// IMPORTANT: 调用完毕后请务必调用迭代器的 Close() 释放资源。
	Scan(start []byte, end []byte) (RecordIterator, error)

	// Flush 将内存数据刷新到磁盘
	Flush() error
}

// RecordIterator 定义了 Arrow 记录迭代器接口
type RecordIterator interface {
	// Next 将迭代器移动到下一个位置
	Next() bool

	// Record 获取当前位置的 Arrow 记录。
	// IMPORTANT: 调用者需在使用完 Record 后调用 Record.Release() 释放内存。
	Record() *arrow.Record

	// Close 关闭迭代器并释放相关资源
	Close() error
}

// Table 表示一个数据表，存储 Arrow 格式的记录
type Table struct {
	// Name 为数据表名称
	Name string

	// Schema 表示数据表的 Arrow Schema，定义各列的类型和名称。
	Schema *arrow.Schema

	// Records 存储 Arrow 格式的记录集合
	// 注意：每个 record 所占用的内存需要在不再使用时进行 Release() 释放。
	Records []*arrow.Record
}

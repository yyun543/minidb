package storage

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/ipc"
	"github.com/apache/arrow/go/v18/arrow/memory"
)

// MemTable 实现基于内存的存储引擎，支持 WAL 日志和内存索引
type MemTable struct {
	wal   *WAL
	index *Index
	mutex sync.RWMutex
}

// NewMemTable 创建新的 MemTable 实例
func NewMemTable(walPath string) (*MemTable, error) {
	wal, err := NewWAL(walPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create WAL: %v", err)
	}

	return &MemTable{
		wal:   wal,
		index: NewIndex(),
	}, nil
}

// Open 实现 Engine 接口
func (mt *MemTable) Open() error {
	// WAL recovery: 从WAL中恢复数据到内存索引
	return mt.recoverFromWAL()
}

// recoverFromWAL 从WAL恢复数据
func (mt *MemTable) recoverFromWAL() error {
	// 扫描所有WAL条目 (从0到当前时间)
	entries, err := mt.wal.Scan(0, 9223372036854775807) // int64最大值
	if err != nil {
		return fmt.Errorf("failed to scan WAL for recovery: %w", err)
	}

	// 按时间戳顺序重放WAL条目
	for _, entry := range entries {
		switch entry.OpType {
		case OpPut:
			// 恢复PUT操作：直接写入索引，不写WAL (避免重复)
			mt.index.Put(entry.Key, entry.Value)
		case OpDelete:
			// 恢复DELETE操作：从索引删除，不写WAL (避免重复)
			mt.index.Delete(entry.Key)
		default:
			// 未知操作类型，跳过
			continue
		}
	}

	return nil
}

// Close 实现 Engine 接口
func (mt *MemTable) Close() error {
	mt.mutex.Lock()
	defer mt.mutex.Unlock()

	// 清理索引
	mt.index.Clear()

	// 关闭 WAL
	return mt.wal.Close()
}

// Get 实现 Engine 接口
func (mt *MemTable) Get(key []byte) (arrow.Record, error) {
	mt.mutex.RLock()
	defer mt.mutex.RUnlock()

	// 从索引中查找
	value, exists := mt.index.Get(key)
	if !exists {
		return nil, nil
	}

	// TODO 将字节数组转换为 Arrow Record
	// 注意：这里假设 value 中存储的是序列化的 Arrow Record
	// 实际实现中需要添加序列化/反序列化逻辑
	record, err := deserializeRecord(value)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize record: %v", err)
	}

	return *record, nil
}

// Put 实现 Engine 接口
func (mt *MemTable) Put(key []byte, record *arrow.Record) error {
	mt.mutex.Lock()
	defer mt.mutex.Unlock()

	// 序列化 Arrow Record
	value, err := serializeRecord(*record)
	if err != nil {
		return fmt.Errorf("failed to serialize record: %v", err)
	}

	// 写入 WAL
	if err := mt.wal.AppendPut(key, value); err != nil {
		return fmt.Errorf("failed to append to WAL: %v", err)
	}

	// 更新内存索引
	mt.index.Put(key, value)

	return nil
}

// Delete 实现 Engine 接口
func (mt *MemTable) Delete(key []byte) error {
	mt.mutex.Lock()
	defer mt.mutex.Unlock()

	// 写入 WAL
	if err := mt.wal.AppendDelete(key); err != nil {
		return fmt.Errorf("failed to append delete to WAL: %v", err)
	}

	// 从索引中删除
	mt.index.Delete(key)

	return nil
}

// Scan 实现 Engine 接口
func (mt *MemTable) Scan(start []byte, end []byte) (RecordIterator, error) {
	mt.mutex.RLock()
	defer mt.mutex.RUnlock()

	// 创建迭代器
	it := &memTableIterator{
		table:    mt,
		iterator: mt.index.NewIterator(start, end),
	}

	return it, nil
}

// Flush 实现 Engine 接口
func (mt *MemTable) Flush() error {
	// TODO MemTable 目前是纯内存实现，需要实现将数据持久化到其他存储介质的逻辑
	return nil
}

// memTableIterator 实现 RecordIterator 接口
type memTableIterator struct {
	table    *MemTable
	iterator *Iterator
	current  arrow.Record
}

// Next 实现 RecordIterator 接口
func (it *memTableIterator) Next() bool {
	// 释放之前的 Record
	if it.current != nil {
		it.current.Release()
		it.current = nil
	}

	if !it.iterator.Next() {
		return false
	}

	// 获取当前值并反序列化为 Record
	value := it.iterator.Value()
	record, err := deserializeRecord(value)
	if err != nil {
		// 处理错误，这里简单返回 false
		return false
	}

	it.current = *record
	return true
}

// Record 实现 RecordIterator 接口
func (it *memTableIterator) Record() arrow.Record {
	return it.current
}

// Close 实现 RecordIterator 接口
func (it *memTableIterator) Close() error {
	if it.current != nil {
		it.current.Release()
		it.current = nil
	}
	return nil
}

// serializeRecord 将 Arrow Record 序列化为字节数组。
// 实现采用 Arrow IPC Writer，将 record 写入到缓冲区。
// 调用者在写入 WAL 时，只需要持久化生成的字节数组。
func serializeRecord(record arrow.Record) ([]byte, error) {
	var buf bytes.Buffer

	// 使用 record 的 schema 进行 IPC writer 初始化，同时指定内存分配器
	allocator := memory.NewGoAllocator()
	writer := ipc.NewWriter(&buf, ipc.WithSchema(record.Schema()), ipc.WithAllocator(allocator))

	// 写入 record（数据批次），必须检查错误
	if err := writer.Write(record); err != nil {
		writer.Close() // 尽量释放 writer 资源
		return nil, fmt.Errorf("failed to write record: %w", err)
	}

	// 关闭 writer，将缓冲区内容刷新
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close IPC writer: %w", err)
	}

	return buf.Bytes(), nil
}

// deserializeRecord 将字节数组反序列化为 Arrow Record。
// 使用 Arrow IPC Reader 从数据中恢复 record，注意仅读取第一个 record 批次。
// 调用者在处理完返回的 record 后应调用 record.Release() 释放资源。
func deserializeRecord(data []byte) (*arrow.Record, error) {
	buf := bytes.NewReader(data)
	reader, err := ipc.NewReader(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create IPC reader: %w", err)
	}

	// 读取第一个 record 批次
	if !reader.Next() {
		reader.Release()
		return nil, fmt.Errorf("no record found in data")
	}
	record := reader.Record()
	// 增加 record 的引用计数，确保在释放 reader 后 record 的数据依然有效
	record.Retain()
	// 释放 reader
	reader.Release()

	return &record, nil
}

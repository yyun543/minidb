package storage

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/ipc"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/logger"
	"go.uber.org/zap"
)

// MemTable 实现基于内存的存储引擎，支持 WAL 日志和内存索引
type MemTable struct {
	wal   *WAL
	index *Index
	mutex sync.RWMutex
}

// NewMemTable 创建新的 MemTable 实例
func NewMemTable(walPath string) (*MemTable, error) {
	logger.WithComponent("memtable").Info("Creating new MemTable instance",
		zap.String("wal_path", walPath))

	start := time.Now()
	wal, err := NewWAL(walPath)
	if err != nil {
		logger.WithComponent("memtable").Error("Failed to create WAL for MemTable",
			zap.String("wal_path", walPath),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create WAL: %v", err)
	}

	memTable := &MemTable{
		wal:   wal,
		index: NewIndex(),
	}

	logger.WithComponent("memtable").Info("MemTable instance created successfully",
		zap.String("wal_path", walPath),
		zap.Duration("creation_time", time.Since(start)))

	return memTable, nil
}

// Open 实现 Engine 接口
func (mt *MemTable) Open() error {
	logger.WithComponent("memtable").Info("Opening MemTable and starting WAL recovery")

	start := time.Now()
	err := mt.recoverFromWAL()
	if err != nil {
		logger.WithComponent("memtable").Error("Failed to open MemTable",
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))
	} else {
		logger.WithComponent("memtable").Info("MemTable opened successfully",
			zap.Duration("recovery_time", time.Since(start)))
	}
	return err
}

// recoverFromWAL 从WAL恢复数据
func (mt *MemTable) recoverFromWAL() error {
	logger.WithComponent("memtable").Info("Starting WAL recovery")

	start := time.Now()
	// 扫描所有WAL条目 (从0到当前时间)
	entries, err := mt.wal.Scan(0, 9223372036854775807) // int64最大值
	if err != nil {
		logger.WithComponent("memtable").Error("Failed to scan WAL entries for recovery",
			zap.Error(err))
		return fmt.Errorf("failed to scan WAL for recovery: %w", err)
	}

	logger.WithComponent("memtable").Info("WAL entries scanned, starting recovery replay",
		zap.Int("total_entries", len(entries)))

	putCount := 0
	deleteCount := 0
	skippedCount := 0

	// 按时间戳顺序重放WAL条目
	for _, entry := range entries {
		switch entry.OpType {
		case OpPut:
			// 恢复PUT操作：直接写入索引，不写WAL (避免重复)
			mt.index.Put(entry.Key, entry.Value)
			putCount++
		case OpDelete:
			// 恢复DELETE操作：从索引删除，不写WAL (避免重复)
			mt.index.Delete(entry.Key)
			deleteCount++
		default:
			// 未知操作类型，跳过
			logger.WithComponent("memtable").Warn("Unknown WAL operation type during recovery",
				zap.Uint8("op_type", uint8(entry.OpType)),
				zap.Int64("timestamp", entry.Timestamp))
			skippedCount++
			continue
		}
	}

	recoveryDuration := time.Since(start)
	logger.WithComponent("memtable").Info("WAL recovery completed",
		zap.Int("total_entries", len(entries)),
		zap.Int("put_operations", putCount),
		zap.Int("delete_operations", deleteCount),
		zap.Int("skipped_operations", skippedCount),
		zap.Duration("recovery_duration", recoveryDuration))

	return nil
}

// Close 实现 Engine 接口
func (mt *MemTable) Close() error {
	logger.WithComponent("memtable").Info("Closing MemTable")

	mt.mutex.Lock()
	defer mt.mutex.Unlock()

	// 清理索引
	mt.index.Clear()
	logger.WithComponent("memtable").Debug("MemTable index cleared")

	// 关闭 WAL
	err := mt.wal.Close()
	if err != nil {
		logger.WithComponent("memtable").Error("Failed to close MemTable WAL",
			zap.Error(err))
	} else {
		logger.WithComponent("memtable").Info("MemTable closed successfully")
	}
	return err
}

// Get 实现 Engine 接口
func (mt *MemTable) Get(key []byte) (arrow.Record, error) {
	logger.WithComponent("memtable").Debug("Getting record from MemTable",
		zap.String("key", string(key)))

	start := time.Now()
	mt.mutex.RLock()
	defer mt.mutex.RUnlock()

	// 从索引中查找
	value, exists := mt.index.Get(key)
	if !exists {
		logger.WithComponent("memtable").Debug("Record not found in MemTable",
			zap.String("key", string(key)),
			zap.Duration("lookup_duration", time.Since(start)))
		return nil, nil
	}

	// 将字节数组转换为 Arrow Record
	record, err := deserializeRecord(value)
	if err != nil {
		logger.WithComponent("memtable").Error("Failed to deserialize record from MemTable",
			zap.String("key", string(key)),
			zap.Int("value_size", len(value)),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))
		return nil, fmt.Errorf("failed to deserialize record: %v", err)
	}

	logger.WithComponent("memtable").Debug("Record retrieved successfully from MemTable",
		zap.String("key", string(key)),
		zap.Int("value_size", len(value)),
		zap.Duration("get_duration", time.Since(start)))

	return *record, nil
}

// Put 实现 Engine 接口
func (mt *MemTable) Put(key []byte, record *arrow.Record) error {
	logger.WithComponent("memtable").Debug("Putting record into MemTable",
		zap.String("key", string(key)),
		zap.Int64("record_rows", (*record).NumRows()))

	start := time.Now()
	mt.mutex.Lock()
	defer mt.mutex.Unlock()

	// 序列化 Arrow Record
	serializeStart := time.Now()
	value, err := serializeRecord(*record)
	if err != nil {
		logger.WithComponent("memtable").Error("Failed to serialize record for MemTable",
			zap.String("key", string(key)),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))
		return fmt.Errorf("failed to serialize record: %v", err)
	}
	serializeDuration := time.Since(serializeStart)

	// 写入 WAL
	walStart := time.Now()
	if err := mt.wal.AppendPut(key, value); err != nil {
		logger.WithComponent("memtable").Error("Failed to append PUT to WAL",
			zap.String("key", string(key)),
			zap.Int("value_size", len(value)),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))
		return fmt.Errorf("failed to append to WAL: %v", err)
	}
	walDuration := time.Since(walStart)

	// 更新内存索引
	mt.index.Put(key, value)

	totalDuration := time.Since(start)
	logger.WithComponent("memtable").Info("Record put into MemTable successfully",
		zap.String("key", string(key)),
		zap.Int("serialized_size", len(value)),
		zap.Int64("record_rows", (*record).NumRows()),
		zap.Duration("serialize_duration", serializeDuration),
		zap.Duration("wal_duration", walDuration),
		zap.Duration("total_duration", totalDuration))

	return nil
}

// Delete 实现 Engine 接口
func (mt *MemTable) Delete(key []byte) error {
	logger.WithComponent("memtable").Debug("Deleting record from MemTable",
		zap.String("key", string(key)))

	start := time.Now()
	mt.mutex.Lock()
	defer mt.mutex.Unlock()

	// 写入 WAL
	walStart := time.Now()
	if err := mt.wal.AppendDelete(key); err != nil {
		logger.WithComponent("memtable").Error("Failed to append DELETE to WAL",
			zap.String("key", string(key)),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))
		return fmt.Errorf("failed to append delete to WAL: %v", err)
	}
	walDuration := time.Since(walStart)

	// 从索引中删除
	mt.index.Delete(key)

	totalDuration := time.Since(start)
	logger.WithComponent("memtable").Info("Record deleted from MemTable successfully",
		zap.String("key", string(key)),
		zap.Duration("wal_duration", walDuration),
		zap.Duration("total_duration", totalDuration))

	return nil
}

// Scan 实现 Engine 接口
func (mt *MemTable) Scan(start []byte, end []byte) (RecordIterator, error) {
	logger.WithComponent("memtable").Debug("Starting MemTable scan",
		zap.String("start_key", string(start)),
		zap.String("end_key", string(end)))

	mt.mutex.RLock()
	defer mt.mutex.RUnlock()

	// 创建迭代器
	it := &memTableIterator{
		table:    mt,
		iterator: mt.index.NewIterator(start, end),
	}

	logger.WithComponent("memtable").Debug("MemTable scan iterator created",
		zap.String("start_key", string(start)),
		zap.String("end_key", string(end)))

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

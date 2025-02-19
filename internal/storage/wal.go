package storage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

// WAL (Write-Ahead Log) 实现基于 bbolt 的预写式日志
type WAL struct {
	db     *bolt.DB
	path   string
	bucket []byte
}

// WALEntry 表示一个WAL条目
type WALEntry struct {
	Timestamp int64  // 操作时间戳
	OpType    OpType // 操作类型
	Key       []byte // 操作的键
	Value     []byte // 操作的值
}

// OpType 定义WAL操作类型
type OpType byte

const (
	OpPut    OpType = 1
	OpDelete OpType = 2
)

// NewWAL 创建新的WAL实例
func NewWAL(path string) (*WAL, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open WAL: %v", err)
	}

	w := &WAL{
		db:     db,
		path:   path,
		bucket: []byte("wal"),
	}

	// 初始化bucket
	err = w.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(w.bucket)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create WAL bucket: %v", err)
	}

	return w, nil
}

// Close 关闭WAL
func (w *WAL) Close() error {
	return w.db.Close()
}

// AppendPut 追加Put操作到WAL
func (w *WAL) AppendPut(key, value []byte) error {
	entry := WALEntry{
		Timestamp: time.Now().UnixNano(),
		OpType:    OpPut,
		Key:       key,
		Value:     value,
	}
	return w.append(entry)
}

// AppendDelete 追加Delete操作到WAL
func (w *WAL) AppendDelete(key []byte) error {
	entry := WALEntry{
		Timestamp: time.Now().UnixNano(),
		OpType:    OpDelete,
		Key:       key,
	}
	return w.append(entry)
}

// append 将WAL条目写入存储
func (w *WAL) append(entry WALEntry) error {
	return w.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(w.bucket)

		// 使用时间戳作为key
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, uint64(entry.Timestamp))

		// 序列化entry
		value := w.encodeEntry(entry)

		return b.Put(key, value)
	})
}

// encodeEntry 序列化WAL条目
func (w *WAL) encodeEntry(entry WALEntry) []byte {
	// 计算总长度：timestamp(8) + opType(1) + keyLen(4) + key + valueLen(4) + value
	keyLen := len(entry.Key)
	valueLen := len(entry.Value)
	totalLen := 8 + 1 + 4 + keyLen + 4 + valueLen

	buf := make([]byte, totalLen)
	offset := 0

	// 写入timestamp
	binary.BigEndian.PutUint64(buf[offset:], uint64(entry.Timestamp))
	offset += 8

	// 写入操作类型
	buf[offset] = byte(entry.OpType)
	offset += 1

	// 写入key长度和key
	binary.BigEndian.PutUint32(buf[offset:], uint32(keyLen))
	offset += 4
	copy(buf[offset:], entry.Key)
	offset += keyLen

	// 写入value长度和value
	binary.BigEndian.PutUint32(buf[offset:], uint32(valueLen))
	offset += 4
	copy(buf[offset:], entry.Value)

	return buf
}

// Scan 扫描指定时间范围内的WAL条目
func (w *WAL) Scan(startTime, endTime int64) ([]WALEntry, error) {
	var entries []WALEntry

	err := w.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(w.bucket)
		c := b.Cursor()

		// 构造范围查询的key
		startKey := make([]byte, 8)
		endKey := make([]byte, 8)
		binary.BigEndian.PutUint64(startKey, uint64(startTime))
		binary.BigEndian.PutUint64(endKey, uint64(endTime))

		for k, v := c.Seek(startKey); k != nil && bytes.Compare(k, endKey) <= 0; k, v = c.Next() {
			entry, err := w.decodeEntry(v)
			if err != nil {
				return err
			}
			entries = append(entries, entry)
		}

		return nil
	})

	return entries, err
}

// decodeEntry 反序列化WAL条目
func (w *WAL) decodeEntry(data []byte) (WALEntry, error) {
	var entry WALEntry

	if len(data) < 13 { // 最小长度：timestamp(8) + opType(1) + keyLen(4)
		return entry, fmt.Errorf("invalid WAL entry data")
	}

	offset := 0

	// 读取timestamp
	entry.Timestamp = int64(binary.BigEndian.Uint64(data[offset:]))
	offset += 8

	// 读取操作类型
	entry.OpType = OpType(data[offset])
	offset += 1

	// 读取key
	keyLen := binary.BigEndian.Uint32(data[offset:])
	offset += 4
	if offset+int(keyLen) > len(data) {
		return entry, fmt.Errorf("invalid key length")
	}
	entry.Key = make([]byte, keyLen)
	copy(entry.Key, data[offset:offset+int(keyLen)])
	offset += int(keyLen)

	// 读取value
	if offset+4 <= len(data) {
		valueLen := binary.BigEndian.Uint32(data[offset:])
		offset += 4
		if offset+int(valueLen) <= len(data) {
			entry.Value = make([]byte, valueLen)
			copy(entry.Value, data[offset:offset+int(valueLen)])
		}
	}

	return entry, nil
}

// Truncate 清除指定时间之前的WAL条目
func (w *WAL) Truncate(beforeTime int64) error {
	return w.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(w.bucket)
		c := b.Cursor()

		endKey := make([]byte, 8)
		binary.BigEndian.PutUint64(endKey, uint64(beforeTime))

		for k, _ := c.First(); k != nil && bytes.Compare(k, endKey) <= 0; k, _ = c.First() {
			if err := c.Delete(); err != nil {
				return err
			}
		}

		return nil
	})
}

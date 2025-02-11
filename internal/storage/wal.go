package storage

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"
)

// TODO WAL (Write-Ahead Log)

// WAL 预写日志
type WAL struct {
	db     *bbolt.DB
	bucket []byte
}

// WALRecord WAL记录
type WALRecord struct {
	Timestamp int64  // 时间戳
	Type      byte   // 操作类型(PUT/DELETE)
	Key       []byte // 键
	Value     []byte // 值
}

const (
	// 操作类型常量
	WAL_PUT    byte = 1
	WAL_DELETE byte = 2
)

// NewWAL 创建WAL实例
func NewWAL(path string) (*WAL, error) {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	// 打开bbolt数据库
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	// 创建bucket
	bucket := []byte("wal")
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &WAL{
		db:     db,
		bucket: bucket,
	}, nil
}

// Write 写入WAL记录
func (w *WAL) Write(recordType byte, key, value []byte) error {
	// 构造WAL记录
	record := &WALRecord{
		Timestamp: time.Now().UnixNano(),
		Type:      recordType,
		Key:       key,
		Value:     value,
	}

	// 序列化记录
	data, err := w.encodeRecord(record)
	if err != nil {
		return err
	}

	// 写入bbolt
	return w.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(w.bucket)
		// 使用时间戳作为key
		k := make([]byte, 8)
		binary.BigEndian.PutUint64(k, uint64(record.Timestamp))
		return b.Put(k, data)
	})
}

// Read 读取WAL记录
func (w *WAL) Read() ([]*WALRecord, error) {
	var records []*WALRecord

	err := w.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(w.bucket)
		return b.ForEach(func(k, v []byte) error {
			record, err := w.decodeRecord(v)
			if err != nil {
				return err
			}
			records = append(records, record)
			return nil
		})
	})

	return records, err
}

// Close 关闭WAL
func (w *WAL) Close() error {
	return w.db.Close()
}

// encodeRecord 序列化WAL记录
func (w *WAL) encodeRecord(record *WALRecord) ([]byte, error) {
	// 8字节时间戳 + 1字节类型 + 4字节key长度 + key + 4字节value长度 + value
	keyLen := len(record.Key)
	valueLen := len(record.Value)
	data := make([]byte, 8+1+4+keyLen+4+valueLen)

	// 写入时间戳
	binary.BigEndian.PutUint64(data[0:], uint64(record.Timestamp))
	// 写入类型
	data[8] = record.Type
	// 写入key长度
	binary.BigEndian.PutUint32(data[9:], uint32(keyLen))
	// 写入key
	copy(data[13:], record.Key)
	// 写入value长度
	binary.BigEndian.PutUint32(data[13+keyLen:], uint32(valueLen))
	// 写入value
	copy(data[17+keyLen:], record.Value)

	return data, nil
}

// decodeRecord 反序列化WAL记录
func (w *WAL) decodeRecord(data []byte) (*WALRecord, error) {
	if len(data) < 13 { // 最小长度检查
		return nil, fmt.Errorf("invalid record data")
	}

	// 读取时间戳
	timestamp := int64(binary.BigEndian.Uint64(data[0:]))
	// 读取类型
	recordType := data[8]
	// 读取key长度
	keyLen := binary.BigEndian.Uint32(data[9:])
	if len(data) < 13+int(keyLen)+4 {
		return nil, fmt.Errorf("invalid record data")
	}
	// 读取key
	key := make([]byte, keyLen)
	copy(key, data[13:13+keyLen])
	// 读取value长度
	valueLen := binary.BigEndian.Uint32(data[13+keyLen:])
	if len(data) < 17+int(keyLen)+int(valueLen) {
		return nil, fmt.Errorf("invalid record data")
	}
	// 读取value
	value := make([]byte, valueLen)
	copy(value, data[17+keyLen:])

	return &WALRecord{
		Timestamp: timestamp,
		Type:      recordType,
		Key:       key,
		Value:     value,
	}, nil
}

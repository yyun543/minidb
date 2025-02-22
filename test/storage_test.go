package test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/yyun543/minidb/internal/storage"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestRecord 创建用于测试的 Arrow Record
func createTestRecord(t *testing.T) arrow.Record {
	pool := memory.NewGoAllocator()

	// 创建 schema
	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "name", Type: arrow.BinaryTypes.String},
		},
		nil,
	)

	// 创建 builder
	builder1 := array.NewInt64Builder(pool)
	defer builder1.Release()
	builder2 := array.NewStringBuilder(pool)
	defer builder2.Release()

	// 添加数据
	builder1.AppendValues([]int64{1, 2, 3}, nil)
	builder2.AppendValues([]string{"a", "b", "c"}, nil)

	// 创建列
	col1 := builder1.NewArray()
	defer col1.Release()
	col2 := builder2.NewArray()
	defer col2.Release()

	// 创建 record
	record := array.NewRecord(schema, []arrow.Array{col1, col2}, 3)
	return record
}

func TestMemTable(t *testing.T) {
	// 创建临时目录用于WAL文件
	tmpDir, err := os.MkdirTemp("", "memtable_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	walPath := filepath.Join(tmpDir, "test.wal")

	// 创建 MemTable
	mt, err := storage.NewMemTable(walPath)
	require.NoError(t, err)

	// 测试 Open
	err = mt.Open()
	require.NoError(t, err)

	// 创建测试数据
	record := createTestRecord(t)
	defer record.Release()

	// 测试 Put
	key := []byte("test-key")
	err = mt.Put(key, &record)
	require.NoError(t, err)

	// 测试 Get
	got, err := mt.Get(key)
	require.NoError(t, err)
	require.NotNil(t, got)
	defer got.Release()

	// 验证获取的数据是否正确
	assert.Equal(t, record.NumRows(), got.NumRows())
	assert.Equal(t, record.NumCols(), got.NumCols())

	// 测试 Scan
	it, err := mt.Scan([]byte("test"), []byte("test-key-z"))
	require.NoError(t, err)

	foundRecord := false
	for it.Next() {
		rec := it.Record()
		assert.Equal(t, record.NumRows(), rec.NumRows())
		assert.Equal(t, record.NumCols(), rec.NumCols())
		foundRecord = true
	}
	assert.True(t, foundRecord)
	it.Close()

	// 测试 Delete
	err = mt.Delete(key)
	require.NoError(t, err)

	// 验证删除是否成功
	got, err = mt.Get(key)
	require.NoError(t, err)
	assert.Nil(t, got)

	// 测试 Close
	err = mt.Close()
	require.NoError(t, err)
}

func TestIndex(t *testing.T) {
	idx := storage.NewIndex()

	// 测试 Put 和 Get
	key := []byte("test-key")
	value := []byte("test-value")
	idx.Put(key, value)

	got, exists := idx.Get(key)
	assert.True(t, exists)
	assert.True(t, bytes.Equal(value, got))

	// 测试范围查询
	idx.Put([]byte("key1"), []byte("value1"))
	idx.Put([]byte("key2"), []byte("value2"))
	idx.Put([]byte("key3"), []byte("value3"))

	items := idx.Range([]byte("key1"), []byte("key3"))
	assert.Equal(t, 2, len(items))
	assert.True(t, bytes.Equal([]byte("value1"), items[0].Value))
	assert.True(t, bytes.Equal([]byte("value2"), items[1].Value))

	// 测试迭代器
	it := idx.NewIterator([]byte("key1"), []byte("key3"))
	count := 0
	for it.Next() {
		count++
		assert.NotNil(t, it.Key())
		assert.NotNil(t, it.Value())
	}
	assert.Equal(t, 2, count)

	// 测试删除
	assert.True(t, idx.Delete(key))
	_, exists = idx.Get(key)
	assert.False(t, exists)

	// 测试清空
	idx.Clear()
	assert.Equal(t, 0, idx.Len())
}

func TestWAL(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "wal_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	walPath := filepath.Join(tmpDir, "test.wal")

	// 创建 WAL
	wal, err := storage.NewWAL(walPath)
	require.NoError(t, err)

	// 测试 AppendPut
	key := []byte("test-key")
	value := []byte("test-value")
	err = wal.AppendPut(key, value)
	require.NoError(t, err)

	// 测试 AppendDelete
	err = wal.AppendDelete(key)
	require.NoError(t, err)

	// 测试 Scan
	entries, err := wal.Scan(0, 1<<63-1) // 扫描所有条目
	require.NoError(t, err)
	assert.Equal(t, 2, len(entries))

	// 验证Put条目
	assert.Equal(t, storage.OpPut, entries[0].OpType)
	assert.True(t, bytes.Equal(key, entries[0].Key))
	assert.True(t, bytes.Equal(value, entries[0].Value))

	// 验证Delete条目
	assert.Equal(t, storage.OpDelete, entries[1].OpType)
	assert.True(t, bytes.Equal(key, entries[1].Key))

	// 测试 Truncate
	err = wal.Truncate(entries[1].Timestamp)
	require.NoError(t, err)

	entries, err = wal.Scan(0, 1<<63-1)
	require.NoError(t, err)
	assert.Equal(t, 0, len(entries))

	// 测试关闭
	err = wal.Close()
	require.NoError(t, err)
}

func TestKeyManager(t *testing.T) {
	km := storage.NewKeyManager()

	// 测试数据库key
	dbKey := km.DatabaseKey("testdb")
	assert.Equal(t, "user:db:testdb", string(dbKey))

	// 测试表key
	tableKey := km.TableChunkKey("testdb", "testtable", 0)
	assert.Equal(t, "user:chunk:testdb:testtable:0", string(tableKey))

	// 测试系统表key
	sysTableKey := km.TableChunkKey(storage.SYS_DATABASE, "sys_tables", 0)
	assert.Equal(t, "sys:chunk:system:sys_tables:0", string(sysTableKey))

	// 测试key解析
	parsed := km.ParseKey("sys:chunk:system:sys_tables:0")
	assert.Equal(t, "system", parsed["database"])
	assert.Equal(t, "sys", parsed["prefix"])
	assert.Equal(t, "sys_tables", parsed["table"])
	assert.Equal(t, "chunk", parsed["type"])
	assert.Equal(t, int64(0), parsed["chunk_id"])
}

package test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/storage"
)

// TODO Storage 单元测试

func TestStorage(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "minidb-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// 测试WAL
	t.Run("WAL", func(t *testing.T) {
		wal, err := storage.NewWAL(tmpDir + "/wal")
		assert.NoError(t, err)
		defer wal.Close()

		// 写入测试数据
		err = wal.Write(storage.WAL_PUT, []byte("key1"), []byte("value1"))
		assert.NoError(t, err)

		err = wal.Write(storage.WAL_PUT, []byte("key2"), []byte("value2"))
		assert.NoError(t, err)

		err = wal.Write(storage.WAL_DELETE, []byte("key1"), nil)
		assert.NoError(t, err)

		// 读取并验证WAL记录
		records, err := wal.Read()
		assert.NoError(t, err)
		assert.Len(t, records, 3)
	})

	// 测试MemTable
	t.Run("MemTable", func(t *testing.T) {
		mt, err := storage.NewMemTable(tmpDir + "/memtable")
		assert.NoError(t, err)
		defer mt.Close()

		// 写入测试数据
		err = mt.Put([]byte("key1"), []byte("value1"))
		assert.NoError(t, err)

		// 读取并验证数据
		value, err := mt.Get([]byte("key1"))
		assert.NoError(t, err)
		assert.Equal(t, []byte("value1"), value)

		// 删除数据
		err = mt.Delete([]byte("key1"))
		assert.NoError(t, err)

		// 验证删除
		value, err = mt.Get([]byte("key1"))
		assert.NoError(t, err)
		assert.Nil(t, value)
	})

	// 测试Index
	t.Run("Index", func(t *testing.T) {
		idx := storage.NewIndex()

		// 写入测试数据
		err := idx.Put([]byte("key1"), []byte("value1"))
		assert.NoError(t, err)

		err = idx.Put([]byte("key2"), []byte("value2"))
		assert.NoError(t, err)

		// 读取并验证数据
		value, err := idx.Get([]byte("key1"))
		assert.NoError(t, err)
		assert.Equal(t, []byte("value1"), value)

		// 范围扫描
		iter, err := idx.Scan([]byte("key1"), []byte("key3"))
		assert.NoError(t, err)

		var count int
		for iter.Next() {
			count++
			assert.True(t, bytes.HasPrefix(iter.Key(), []byte("key")))
		}
		assert.Equal(t, 2, count)
		iter.Close()

		// 删除数据
		err = idx.Delete([]byte("key1"))
		assert.NoError(t, err)

		// 验证删除
		value, err = idx.Get([]byte("key1"))
		assert.NoError(t, err)
		assert.Nil(t, value)
	})
}

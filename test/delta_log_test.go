package test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/storage"
)

// TestDeltaLogOperations 测试 Delta Log 基本操作
func TestDeltaLogOperations(t *testing.T) {
	ctx := context.Background()
	tempDir := setupDeltaLogTempDir(t)
	defer os.RemoveAll(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	require.NoError(t, engine.CreateDatabase("testdb"))

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "name", Type: arrow.BinaryTypes.String},
		}, nil,
	)

	require.NoError(t, engine.CreateTable("testdb", "users", schema))

	deltaLog := engine.GetDeltaLog()

	t.Run("AppendAdd", func(t *testing.T) {
		initialVersion := deltaLog.GetLatestVersion()

		// 写入数据会触发 AppendAdd
		record := createSimpleRecord(t, schema, 1, 100)
		err := engine.Write(ctx, "testdb", "users", record)
		record.Release()
		require.NoError(t, err)

		// 验证版本增加
		newVersion := deltaLog.GetLatestVersion()
		assert.Greater(t, newVersion, initialVersion, "Version should increase after ADD")
	})

	t.Run("GetSnapshot", func(t *testing.T) {
		// 添加更多文件
		for i := 0; i < 3; i++ {
			record := createSimpleRecord(t, schema, (i+2)*100, 50)
			err := engine.Write(ctx, "testdb", "users", record)
			record.Release()
			require.NoError(t, err)
		}

		// 获取快照
		snapshot, err := deltaLog.GetSnapshot("testdb.users", -1)
		require.NoError(t, err)

		// 应该有 4 个文件（1个初始 + 3个新增）
		assert.GreaterOrEqual(t, len(snapshot.Files), 4, "Should have at least 4 files")

		// 验证 schema
		assert.NotNil(t, snapshot.Schema, "Snapshot should have schema")
	})

	t.Run("AppendRemove", func(t *testing.T) {
		// 获取当前快照
		snapshot, err := deltaLog.GetSnapshot("testdb.users", -1)
		require.NoError(t, err)

		if len(snapshot.Files) > 0 {
			// 删除第一个文件
			initialFileCount := len(snapshot.Files)
			initialVersion := deltaLog.GetLatestVersion()

			err = deltaLog.AppendRemove("testdb.users", snapshot.Files[0].Path)
			require.NoError(t, err)

			// 验证版本增加
			newVersion := deltaLog.GetLatestVersion()
			assert.Greater(t, newVersion, initialVersion, "Version should increase after REMOVE")

			// 获取新快照
			newSnapshot, err := deltaLog.GetSnapshot("testdb.users", -1)
			require.NoError(t, err)

			// 文件数应该减少1
			assert.Equal(t, initialFileCount-1, len(newSnapshot.Files), "File count should decrease by 1")
		}
	})
}

// TestDeltaLogVersionControl 测试版本控制功能
func TestDeltaLogVersionControl(t *testing.T) {
	ctx := context.Background()
	tempDir := setupDeltaLogTempDir(t)
	defer os.RemoveAll(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	require.NoError(t, engine.CreateDatabase("testdb"))

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "value", Type: arrow.BinaryTypes.String},
		}, nil,
	)

	require.NoError(t, engine.CreateTable("testdb", "version_test", schema))

	deltaLog := engine.GetDeltaLog()

	// 记录每个版本的快照
	versions := make([]int64, 0)

	// 添加数据并记录版本
	for i := 0; i < 5; i++ {
		record := createSimpleRecord(t, schema, i*100, 50)
		err := engine.Write(ctx, "testdb", "version_test", record)
		record.Release()
		require.NoError(t, err)

		version := deltaLog.GetLatestVersion()
		versions = append(versions, version)
		t.Logf("Batch %d: version %d", i+1, version)
	}

	// 验证可以获取历史版本快照
	for i, version := range versions {
		snapshot, err := deltaLog.GetSnapshot("testdb.version_test", version)
		require.NoError(t, err)

		expectedFiles := i + 1 // 前i+1个文件
		assert.Equal(t, expectedFiles, len(snapshot.Files),
			"Version %d should have %d files", version, expectedFiles)
	}
}

// TestDeltaLogPersistence 测试持久化功能
func TestDeltaLogPersistence(t *testing.T) {
	ctx := context.Background()
	tempDir := setupDeltaLogTempDir(t)
	defer os.RemoveAll(tempDir)

	var lastVersion int64
	var fileCount int

	// 第一阶段：创建数据
	{
		engine, err := storage.NewParquetEngine(tempDir)
		require.NoError(t, err)
		require.NoError(t, engine.Open())

		require.NoError(t, engine.CreateDatabase("testdb"))

		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "id", Type: arrow.PrimitiveTypes.Int64},
				{Name: "name", Type: arrow.BinaryTypes.String},
			}, nil,
		)

		require.NoError(t, engine.CreateTable("testdb", "persist_test", schema))

		// 添加数据
		for i := 0; i < 5; i++ {
			record := createSimpleRecord(t, schema, i*100, 100)
			err := engine.Write(ctx, "testdb", "persist_test", record)
			record.Release()
			require.NoError(t, err)
		}

		deltaLog := engine.GetDeltaLog()
		lastVersion = deltaLog.GetLatestVersion()

		snapshot, err := deltaLog.GetSnapshot("testdb.persist_test", -1)
		require.NoError(t, err)
		fileCount = len(snapshot.Files)

		t.Logf("Created data: version=%d, files=%d", lastVersion, fileCount)

		engine.Close()
	}

	// 第二阶段：重启并验证数据恢复
	{
		engine, err := storage.NewParquetEngine(tempDir)
		require.NoError(t, err)
		require.NoError(t, engine.Open())
		defer engine.Close()

		deltaLog := engine.GetDeltaLog()

		// 验证版本持久化
		recoveredVersion := deltaLog.GetLatestVersion()
		assert.Equal(t, lastVersion, recoveredVersion, "Version should be restored")

		// 验证快照持久化
		snapshot, err := deltaLog.GetSnapshot("testdb.persist_test", -1)
		require.NoError(t, err)
		assert.Equal(t, fileCount, len(snapshot.Files), "File count should be restored")

		// 验证可以查询数据
		iterator, err := engine.Scan(ctx, "testdb", "persist_test", nil)
		require.NoError(t, err)
		defer iterator.Close()

		totalRows := int64(0)
		for iterator.Next() {
			record := iterator.Record()
			totalRows += record.NumRows()
		}
		require.NoError(t, iterator.Err())

		expectedRows := int64(500) // 5 batches * 100 rows
		assert.Equal(t, expectedRows, totalRows, "Row count should be restored")
		t.Logf("Recovered %d rows successfully", totalRows)
	}
}

// TestDeltaLogConcurrency 测试并发安全性
func TestDeltaLogConcurrency(t *testing.T) {
	ctx := context.Background()
	tempDir := setupDeltaLogTempDir(t)
	defer os.RemoveAll(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	require.NoError(t, engine.CreateDatabase("testdb"))

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "data", Type: arrow.BinaryTypes.String},
		}, nil,
	)

	require.NoError(t, engine.CreateTable("testdb", "concurrent_test", schema))

	// 并发写入
	numGoroutines := 5
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			record := createSimpleRecord(t, schema, index*1000, 100)
			err := engine.Write(ctx, "testdb", "concurrent_test", record)
			record.Release()
			errors <- err
		}(i)
	}

	// 收集所有错误
	for i := 0; i < numGoroutines; i++ {
		err := <-errors
		assert.NoError(t, err, "Concurrent write should succeed")
	}

	// 验证所有文件都被添加
	deltaLog := engine.GetDeltaLog()
	snapshot, err := deltaLog.GetSnapshot("testdb.concurrent_test", -1)
	require.NoError(t, err)

	assert.Equal(t, numGoroutines, len(snapshot.Files), "Should have files from all goroutines")

	// 验证数据完整性
	iterator, err := engine.Scan(ctx, "testdb", "concurrent_test", nil)
	require.NoError(t, err)
	defer iterator.Close()

	totalRows := int64(0)
	for iterator.Next() {
		record := iterator.Record()
		totalRows += record.NumRows()
	}
	require.NoError(t, iterator.Err())

	expectedRows := int64(numGoroutines * 100)
	assert.Equal(t, expectedRows, totalRows, "Should have all rows from concurrent writes")
}

// TestDeltaLogTimeTravel 测试时间旅行查询
func TestDeltaLogTimeTravel(t *testing.T) {
	ctx := context.Background()
	tempDir := setupDeltaLogTempDir(t)
	defer os.RemoveAll(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	require.NoError(t, engine.CreateDatabase("testdb"))

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "value", Type: arrow.BinaryTypes.String},
		}, nil,
	)

	require.NoError(t, engine.CreateTable("testdb", "timetravel", schema))

	deltaLog := engine.GetDeltaLog()

	// 添加数据并记录时间戳
	timestamps := make([]int64, 0)
	versions := make([]int64, 0)

	for i := 0; i < 3; i++ {
		record := createSimpleRecord(t, schema, i*100, 50)
		err := engine.Write(ctx, "testdb", "timetravel", record)
		record.Release()
		require.NoError(t, err)

		timestamps = append(timestamps, time.Now().UnixMilli())
		versions = append(versions, deltaLog.GetLatestVersion())

		time.Sleep(10 * time.Millisecond) // 确保时间戳不同
	}

	// 测试根据时间戳查找版本
	for i, ts := range timestamps {
		version, err := deltaLog.GetVersionByTimestamp("testdb.timetravel", ts)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, version, versions[i],
			"Version at timestamp should be >= recorded version")
		t.Logf("Timestamp %d: version %d", ts, version)
	}
}

// TestDeltaLogMetadata 测试 Metadata 操作
func TestDeltaLogMetadata(t *testing.T) {
	tempDir := setupDeltaLogTempDir(t)
	defer os.RemoveAll(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	require.NoError(t, engine.CreateDatabase("testdb"))

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "name", Type: arrow.BinaryTypes.String},
		}, nil,
	)

	// CreateTable 会自动追加 METADATA 操作
	require.NoError(t, engine.CreateTable("testdb", "metadata_test", schema))

	deltaLog := engine.GetDeltaLog()

	// 获取快照并验证 schema
	snapshot, err := deltaLog.GetSnapshot("testdb.metadata_test", -1)
	require.NoError(t, err)

	assert.NotNil(t, snapshot.Schema, "Snapshot should have schema")
	assert.Equal(t, 2, len(snapshot.Schema.Fields()), "Schema should have 2 fields")
	assert.Equal(t, "id", snapshot.Schema.Field(0).Name, "First field should be 'id'")
	assert.Equal(t, "name", snapshot.Schema.Field(1).Name, "Second field should be 'name'")
}

// Helper functions

func setupDeltaLogTempDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "minidb_deltalog_test_*")
	require.NoError(t, err)
	return tempDir
}

func createSimpleRecord(t *testing.T, schema *arrow.Schema, startID int, count int) arrow.Record {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	for i := 0; i < count; i++ {
		id := int64(startID + i)
		value := "value_" + string(rune('A'+(i%26)))

		builder.Field(0).(*array.Int64Builder).Append(id)
		builder.Field(1).(*array.StringBuilder).Append(value)
	}

	return builder.NewRecord()
}

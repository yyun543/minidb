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

// TestCheckpointCreation 测试 Checkpoint 自动创建功能
func TestCheckpointCreation(t *testing.T) {
	ctx := context.Background()
	tempDir := setupCheckpointTempDir(t)
	defer os.RemoveAll(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	// 创建测试数据库和表
	require.NoError(t, engine.CreateDatabase("testdb"))

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "value", Type: arrow.BinaryTypes.String},
		}, nil,
	)

	require.NoError(t, engine.CreateTable("testdb", "checkpoint_table", schema))

	// 插入多批数据以触发 checkpoint 创建（每10个版本创建一次）
	// 版本1: METADATA, 版本2-13: 12个 ADD 操作
	for i := 0; i < 12; i++ {
		record := createTestRecord(t, schema, i*100, 100)
		err := engine.Write(ctx, "testdb", "checkpoint_table", record)
		record.Release()
		require.NoError(t, err)
		t.Logf("Inserted batch %d", i+1)
	}

	// 等待异步 checkpoint 创建完成
	time.Sleep(200 * time.Millisecond)

	// 验证最终版本号（1 METADATA + 12 ADD = 13）
	deltaLog := engine.GetDeltaLog()
	finalVersion := deltaLog.GetLatestVersion()
	t.Logf("Final version: %d", finalVersion)
	assert.Equal(t, int64(13), finalVersion, "Expected version 13 (1 METADATA + 12 ADD)")

	// 验证快照数据完整性
	snapshot, err := deltaLog.GetSnapshot("testdb.checkpoint_table", -1)
	require.NoError(t, err)

	assert.Equal(t, 12, len(snapshot.Files), "Should have 12 data files")
	t.Logf("Final snapshot contains %d files", len(snapshot.Files))

	// 验证可以查询数据
	iterator, err := engine.Scan(ctx, "testdb", "checkpoint_table", nil)
	require.NoError(t, err)
	defer iterator.Close()

	totalRows := int64(0)
	for iterator.Next() {
		record := iterator.Record()
		totalRows += record.NumRows()
	}
	require.NoError(t, iterator.Err())

	assert.Equal(t, int64(1200), totalRows, "Should have 1200 rows (12 batches * 100 rows)")
	t.Logf("Total rows after checkpoint: %d", totalRows)
}

// TestCheckpointRecovery 测试从 Checkpoint 恢复
func TestCheckpointRecovery(t *testing.T) {
	ctx := context.Background()
	tempDir := setupCheckpointTempDir(t)
	defer os.RemoveAll(tempDir)

	// 第一阶段：创建数据并触发 checkpoint
	{
		engine, err := storage.NewParquetEngine(tempDir)
		require.NoError(t, err)
		require.NoError(t, engine.Open())

		require.NoError(t, engine.CreateDatabase("testdb"))

		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "id", Type: arrow.PrimitiveTypes.Int64},
				{Name: "value", Type: arrow.BinaryTypes.String},
			}, nil,
		)

		require.NoError(t, engine.CreateTable("testdb", "recovery_table", schema))

		// 插入15批数据以触发多个 checkpoint（版本10、20）
		for i := 0; i < 15; i++ {
			record := createTestRecord(t, schema, i*100, 50)
			err := engine.Write(ctx, "testdb", "recovery_table", record)
			record.Release()
			require.NoError(t, err)
		}

		// 等待 checkpoint 创建
		time.Sleep(200 * time.Millisecond)

		version := engine.GetDeltaLog().GetLatestVersion()
		t.Logf("Created data with version: %d", version)

		engine.Close()
	}

	// 第二阶段：重启并验证数据恢复
	{
		engine, err := storage.NewParquetEngine(tempDir)
		require.NoError(t, err)
		require.NoError(t, engine.Open())
		defer engine.Close()

		// 验证表存在
		exists, err := engine.TableExists("testdb", "recovery_table")
		require.NoError(t, err)
		assert.True(t, exists, "Table should exist after recovery")

		// 验证快照
		snapshot, err := engine.GetDeltaLog().GetSnapshot("testdb.recovery_table", -1)
		require.NoError(t, err)
		assert.Equal(t, 15, len(snapshot.Files), "Should have 15 files after recovery")

		// 验证数据完整性
		iterator, err := engine.Scan(ctx, "testdb", "recovery_table", nil)
		require.NoError(t, err)
		defer iterator.Close()

		totalRows := int64(0)
		for iterator.Next() {
			record := iterator.Record()
			totalRows += record.NumRows()
		}
		require.NoError(t, iterator.Err())

		assert.Equal(t, int64(750), totalRows, "Should have 750 rows (15 batches * 50 rows)")
		t.Logf("Recovered %d rows successfully", totalRows)
	}
}

// TestCheckpointVersionControl 测试 Checkpoint 版本控制
func TestCheckpointVersionControl(t *testing.T) {
	ctx := context.Background()
	tempDir := setupCheckpointTempDir(t)
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

	require.NoError(t, engine.CreateTable("testdb", "version_table", schema))

	deltaLog := engine.GetDeltaLog()

	// 插入数据并跟踪版本
	versions := make([]int64, 0)
	for i := 0; i < 5; i++ {
		record := createTestRecord(t, schema, i*100, 100)
		err := engine.Write(ctx, "testdb", "version_table", record)
		record.Release()
		require.NoError(t, err)

		version := deltaLog.GetLatestVersion()
		versions = append(versions, version)
		t.Logf("Batch %d: version %d", i+1, version)
	}

	// 验证可以获取历史版本快照
	for i, version := range versions {
		snapshot, err := deltaLog.GetSnapshot("testdb.version_table", version)
		require.NoError(t, err)

		expectedFiles := i + 1 // 前i+1个文件
		assert.Equal(t, expectedFiles, len(snapshot.Files),
			"Version %d should have %d files", version, expectedFiles)
		t.Logf("Version %d: %d files", version, len(snapshot.Files))
	}
}

// Helper functions

func setupCheckpointTempDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "minidb_checkpoint_test_*")
	require.NoError(t, err)
	return tempDir
}

func createTestRecord(t *testing.T, schema *arrow.Schema, startID int, count int) arrow.Record {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	for i := 0; i < count; i++ {
		id := int64(startID + i)
		value := "test_" + string(rune('A'+i%26))

		builder.Field(0).(*array.Int64Builder).Append(id)
		builder.Field(1).(*array.StringBuilder).Append(value)
	}

	return builder.NewRecord()
}

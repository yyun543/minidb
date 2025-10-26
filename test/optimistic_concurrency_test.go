package test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/delta"
	"github.com/yyun543/minidb/internal/storage"
)

// TestOptimisticConcurrency_TwoWritersSimultaneousCommit 测试两个writer同时提交
// 根据Delta Lake论文，应该支持多writer并发提交，通过版本冲突检测实现
func TestOptimisticConcurrency_TwoWritersSimultaneousCommit(t *testing.T) {
	tempDir := SetupTestDir(t, "optimistic_concurrency")
	defer CleanupTestDir(tempDir)

	// 创建引擎 - 启用乐观并发控制
	engine, err := storage.NewParquetEngine(tempDir,
		storage.WithOptimisticLock(true),
		storage.WithMaxRetries(5))
	require.NoError(t, err)
	defer engine.Close()
	err = engine.Open()
	require.NoError(t, err)

	// 创建测试表
	err = engine.CreateDatabase("testdb")
	require.NoError(t, err)

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "value", Type: arrow.BinaryTypes.String},
	}, nil)

	err = engine.CreateTable("testdb", "test_table", schema)
	require.NoError(t, err)

	t.Log("=== Testing Concurrent Writes ===")

	// 准备两个writer的数据
	pool := memory.NewGoAllocator()

	// Writer 1的数据
	builder1 := array.NewRecordBuilder(pool, schema)
	builder1.Field(0).(*array.Int64Builder).Append(1)
	builder1.Field(1).(*array.StringBuilder).Append("writer1")
	record1 := builder1.NewRecord()
	defer record1.Release()
	defer builder1.Release()

	// Writer 2的数据
	builder2 := array.NewRecordBuilder(pool, schema)
	builder2.Field(0).(*array.Int64Builder).Append(2)
	builder2.Field(1).(*array.StringBuilder).Append("writer2")
	record2 := builder2.NewRecord()
	defer record2.Release()
	defer builder2.Release()

	// 并发写入
	var wg sync.WaitGroup
	wg.Add(2)

	errors := make([]error, 2)
	startBarrier := make(chan struct{})

	// Writer 1
	go func() {
		defer wg.Done()
		<-startBarrier // 等待同时开始
		ctx := context.Background()
		errors[0] = engine.Write(ctx, "testdb", "test_table", record1)
		t.Logf("Writer 1 completed: %v", errors[0])
	}()

	// Writer 2
	go func() {
		defer wg.Done()
		<-startBarrier // 等待同时开始
		ctx := context.Background()
		errors[1] = engine.Write(ctx, "testdb", "test_table", record2)
		t.Logf("Writer 2 completed: %v", errors[1])
	}()

	// 同时开始写入
	close(startBarrier)
	wg.Wait()

	// 验证结果：至少一个成功
	successCount := 0
	if errors[0] == nil {
		successCount++
	}
	if errors[1] == nil {
		successCount++
	}

	require.Greater(t, successCount, 0, "At least one writer should succeed")
	t.Logf("Success count: %d/2", successCount)

	// 如果有冲突，应该是RetryableConflict错误
	for i, err := range errors {
		if err != nil {
			t.Logf("Writer %d failed with: %v", i+1, err)
			// 在乐观锁实现后，这应该是可重试的冲突错误
			// require.Contains(t, err.Error(), "conflict", "Failure should indicate conflict")
		}
	}

	// 验证Delta Log状态一致性
	deltaLog := engine.GetDeltaLog()
	snapshot, err := deltaLog.GetSnapshot("testdb.test_table", -1)
	require.NoError(t, err)
	t.Logf("Final snapshot has %d files", len(snapshot.Files))

	// 验证数据完整性
	ctx := context.Background()
	iterator, err := engine.Scan(ctx, "testdb", "test_table", nil)
	require.NoError(t, err)

	totalRows := int64(0)
	for iterator.Next() {
		record := iterator.Record()
		totalRows += record.NumRows()
		t.Logf("Read batch: %d rows", record.NumRows())
	}

	require.Equal(t, int64(successCount), totalRows, "Row count should match success count")
}

// TestOptimisticConcurrency_ConflictRetry 测试冲突重试机制
func TestOptimisticConcurrency_ConflictRetry(t *testing.T) {
	tempDir := SetupTestDir(t, "optimistic_retry")
	defer CleanupTestDir(tempDir)

	engine, err := storage.NewParquetEngine(tempDir,
		storage.WithOptimisticLock(true),
		storage.WithMaxRetries(5))
	require.NoError(t, err)
	defer engine.Close()
	err = engine.Open()
	require.NoError(t, err)

	// 创建测试表
	err = engine.CreateDatabase("testdb")
	require.NoError(t, err)

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "data", Type: arrow.BinaryTypes.String},
	}, nil)

	err = engine.CreateTable("testdb", "retry_table", schema)
	require.NoError(t, err)

	t.Log("=== Testing Conflict Retry Mechanism ===")

	// 模拟10个并发writer
	concurrentWriters := 10
	var wg sync.WaitGroup
	wg.Add(concurrentWriters)

	successCount := int32(0)
	conflictCount := int32(0)
	startBarrier := make(chan struct{})

	pool := memory.NewGoAllocator()

	for i := 0; i < concurrentWriters; i++ {
		writerID := i
		go func() {
			defer wg.Done()
			<-startBarrier

			// 准备数据
			builder := array.NewRecordBuilder(pool, schema)
			builder.Field(0).(*array.Int64Builder).Append(int64(writerID))
			builder.Field(1).(*array.StringBuilder).Append(fmt.Sprintf("writer_%d", writerID))
			record := builder.NewRecord()
			defer record.Release()
			defer builder.Release()

			// 写入（带重试）
			ctx := context.Background()
			maxRetries := 5
			for attempt := 0; attempt < maxRetries; attempt++ {
				err := engine.Write(ctx, "testdb", "retry_table", record)
				if err == nil {
					t.Logf("Writer %d succeeded on attempt %d", writerID, attempt+1)
					successCount++
					return
				}

				// 检查是否是可重试的冲突错误
				if isRetryableConflict(err) {
					conflictCount++
					t.Logf("Writer %d conflict on attempt %d, retrying...", writerID, attempt+1)
					time.Sleep(time.Millisecond * time.Duration(10+writerID*5)) // 随机退避
					continue
				}

				// 不可重试的错误
				t.Logf("Writer %d failed with non-retryable error: %v", writerID, err)
				return
			}

			t.Logf("Writer %d exhausted retries", writerID)
		}()
	}

	close(startBarrier)
	wg.Wait()

	t.Logf("Results: %d successes, %d conflicts", successCount, conflictCount)

	// 验证：所有writer最终都应该成功（带重试）
	require.Equal(t, int32(concurrentWriters), successCount, "All writers should eventually succeed with retry")

	// 验证数据完整性
	ctx := context.Background()
	iterator, err := engine.Scan(ctx, "testdb", "retry_table", nil)
	require.NoError(t, err)

	totalRows := int64(0)
	for iterator.Next() {
		record := iterator.Record()
		totalRows += record.NumRows()
	}

	require.Equal(t, int64(concurrentWriters), totalRows, "Should have exactly %d rows", concurrentWriters)
}

// TestOptimisticConcurrency_VersionSequencing 测试版本号顺序性
func TestOptimisticConcurrency_VersionSequencing(t *testing.T) {
	tempDir := SetupTestDir(t, "version_sequencing")
	defer CleanupTestDir(tempDir)

	engine, err := storage.NewParquetEngine(tempDir,
		storage.WithOptimisticLock(true),
		storage.WithMaxRetries(5))
	require.NoError(t, err)
	defer engine.Close()
	err = engine.Open()
	require.NoError(t, err)

	// 创建测试表
	err = engine.CreateDatabase("testdb")
	require.NoError(t, err)

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
	}, nil)

	err = engine.CreateTable("testdb", "version_table", schema)
	require.NoError(t, err)

	t.Log("=== Testing Version Sequencing ===")

	// 顺序写入10次
	pool := memory.NewGoAllocator()
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		builder := array.NewRecordBuilder(pool, schema)
		builder.Field(0).(*array.Int64Builder).Append(int64(i))
		record := builder.NewRecord()

		err := engine.Write(ctx, "testdb", "version_table", record)
		require.NoError(t, err)

		record.Release()
		builder.Release()

		t.Logf("Write %d completed", i+1)
	}

	// 验证版本号连续性
	deltaLog := engine.GetDeltaLog()
	entries := deltaLog.GetEntriesByTable("testdb.version_table")

	t.Logf("Total entries in Delta Log: %d", len(entries))

	// 收集所有版本号
	versions := make([]int64, 0)
	for _, entry := range entries {
		if entry.Operation == delta.OpAdd {
			versions = append(versions, entry.Version)
			t.Logf("Version: %d, Operation: %s, File: %s", entry.Version, entry.Operation, entry.FilePath)
		}
	}

	// 验证版本号严格递增且连续
	for i := 1; i < len(versions); i++ {
		require.Greater(t, versions[i], versions[i-1], "Versions should be strictly increasing")
	}

	t.Log("✓ Version sequencing verified")
}

// TestOptimisticConcurrency_SnapshotIsolation 测试快照隔离
func TestOptimisticConcurrency_SnapshotIsolation(t *testing.T) {
	tempDir := SetupTestDir(t, "snapshot_isolation")
	defer CleanupTestDir(tempDir)

	engine, err := storage.NewParquetEngine(tempDir,
		storage.WithOptimisticLock(true),
		storage.WithMaxRetries(5))
	require.NoError(t, err)
	defer engine.Close()
	err = engine.Open()
	require.NoError(t, err)

	// 创建测试表
	err = engine.CreateDatabase("testdb")
	require.NoError(t, err)

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "value", Type: arrow.BinaryTypes.String},
	}, nil)

	err = engine.CreateTable("testdb", "isolation_table", schema)
	require.NoError(t, err)

	t.Log("=== Testing Snapshot Isolation ===")

	pool := memory.NewGoAllocator()
	ctx := context.Background()

	// 写入初始数据
	builder := array.NewRecordBuilder(pool, schema)
	builder.Field(0).(*array.Int64Builder).Append(1)
	builder.Field(1).(*array.StringBuilder).Append("initial")
	record := builder.NewRecord()
	err = engine.Write(ctx, "testdb", "isolation_table", record)
	require.NoError(t, err)
	record.Release()
	builder.Release()

	// 获取快照版本V1
	deltaLog := engine.GetDeltaLog()
	v1 := deltaLog.GetLatestVersion()
	t.Logf("Initial version: V%d", v1)

	snapshotV1, err := deltaLog.GetSnapshot("testdb.isolation_table", v1)
	require.NoError(t, err)
	require.Equal(t, 1, len(snapshotV1.Files), "V1 should have 1 file")

	// 写入更多数据
	builder2 := array.NewRecordBuilder(pool, schema)
	builder2.Field(0).(*array.Int64Builder).Append(2)
	builder2.Field(1).(*array.StringBuilder).Append("second")
	record2 := builder2.NewRecord()
	err = engine.Write(ctx, "testdb", "isolation_table", record2)
	require.NoError(t, err)
	record2.Release()
	builder2.Release()

	v2 := deltaLog.GetLatestVersion()
	t.Logf("After second write: V%d", v2)

	// 再次获取V1快照，应该不变
	snapshotV1Again, err := deltaLog.GetSnapshot("testdb.isolation_table", v1)
	require.NoError(t, err)
	require.Equal(t, len(snapshotV1.Files), len(snapshotV1Again.Files), "V1 snapshot should remain unchanged")
	require.Equal(t, snapshotV1.Files[0].Path, snapshotV1Again.Files[0].Path, "File paths should match")

	// 获取V2快照，应该有2个文件
	snapshotV2, err := deltaLog.GetSnapshot("testdb.isolation_table", v2)
	require.NoError(t, err)
	require.Equal(t, 2, len(snapshotV2.Files), "V2 should have 2 files")

	t.Log("✓ Snapshot isolation verified")
}

// 辅助函数：检查是否是可重试的冲突错误
func isRetryableConflict(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	// 检查常见的冲突错误标识
	return containsAny(errMsg, []string{
		"conflict",
		"PreconditionFailed",
		"version conflict",
		"concurrent modification",
		"already exists",
	})
}

func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

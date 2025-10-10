package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/yyun543/minidb/internal/delta"
	"github.com/yyun543/minidb/internal/delta/persistent"
	"github.com/yyun543/minidb/internal/storage"
)

// TestPersistentDeltaLog 测试持久化 Delta Log 功能
func TestPersistentDeltaLog(t *testing.T) {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "minidb-test-delta-log", time.Now().Format("20060102150405"))
	defer os.RemoveAll(tempDir)

	// 创建存储引擎
	engine, err := storage.NewParquetEngine(tempDir)
	if err != nil {
		t.Fatalf("Failed to create ParquetEngine: %v", err)
	}
	defer engine.Close()

	err = engine.Open()
	if err != nil {
		t.Fatalf("Failed to open ParquetEngine: %v", err)
	}

	// 创建系统数据库和 delta_log 表
	err = engine.CreateDatabase("sys")
	if err != nil {
		t.Fatalf("Failed to create sys database: %v", err)
	}

	// 创建 Delta Log schema
	deltaLogSchema := createDeltaLogSchema()
	err = engine.CreateTable("sys", "delta_log", deltaLogSchema)
	if err != nil {
		t.Fatalf("Failed to create delta_log table: %v", err)
	}

	// 创建持久化 Delta Log
	deltaLog := persistent.NewPersistentDeltaLog(engine)

	t.Run("AppendAdd", func(t *testing.T) {
		tableID := "test.users"
		parquetFile := &delta.ParquetFile{
			Path:     "/data/users/part-001.parquet",
			Size:     1024000,
			RowCount: 1000,
			Stats: &delta.FileStats{
				RowCount:   1000,
				FileSize:   1024000,
				MinValues:  map[string]interface{}{"id": int64(1), "name": "alice"},
				MaxValues:  map[string]interface{}{"id": int64(1000), "name": "zoe"},
				NullCounts: map[string]int64{"id": 0, "name": 0},
			},
		}

		err := deltaLog.AppendAdd(tableID, parquetFile)
		if err != nil {
			t.Fatalf("AppendAdd failed: %v", err)
		}

		// 验证 Delta Log 版本增加
		version := deltaLog.GetLatestVersion()
		if version != 1 {
			t.Fatalf("Expected version 1, got %d", version)
		}
	})

	t.Run("AppendRemove", func(t *testing.T) {
		tableID := "test.users"
		filePath := "/data/users/part-001.parquet"

		err := deltaLog.AppendRemove(tableID, filePath)
		if err != nil {
			t.Fatalf("AppendRemove failed: %v", err)
		}

		// 验证版本增加
		version := deltaLog.GetLatestVersion()
		if version != 2 {
			t.Fatalf("Expected version 2, got %d", version)
		}
	})

	t.Run("AppendMetadata", func(t *testing.T) {
		tableID := "test.users"
		schema := arrow.NewSchema([]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "name", Type: arrow.BinaryTypes.String},
		}, nil)

		err := deltaLog.AppendMetadata(tableID, schema)
		if err != nil {
			t.Fatalf("AppendMetadata failed: %v", err)
		}

		// 验证版本增加
		version := deltaLog.GetLatestVersion()
		if version != 3 {
			t.Fatalf("Expected version 3, got %d", version)
		}
	})

	t.Run("GetSnapshot", func(t *testing.T) {
		tableID := "test.inventory"

		// 添加一些日志条目
		file1 := &delta.ParquetFile{
			Path:     "/data/inventory/part-001.parquet",
			Size:     512000,
			RowCount: 500,
			Stats: &delta.FileStats{
				RowCount:   500,
				FileSize:   512000,
				MinValues:  map[string]interface{}{"product_id": int64(1)},
				MaxValues:  map[string]interface{}{"product_id": int64(500)},
				NullCounts: map[string]int64{"product_id": 0},
			},
		}

		file2 := &delta.ParquetFile{
			Path:     "/data/inventory/part-002.parquet",
			Size:     256000,
			RowCount: 250,
			Stats: &delta.FileStats{
				RowCount:   250,
				FileSize:   256000,
				MinValues:  map[string]interface{}{"product_id": int64(501)},
				MaxValues:  map[string]interface{}{"product_id": int64(750)},
				NullCounts: map[string]int64{"product_id": 0},
			},
		}

		// 添加文件
		err := deltaLog.AppendAdd(tableID, file1)
		if err != nil {
			t.Fatalf("Failed to add file1: %v", err)
		}

		err = deltaLog.AppendAdd(tableID, file2)
		if err != nil {
			t.Fatalf("Failed to add file2: %v", err)
		}

		// 移除 file1
		err = deltaLog.AppendRemove(tableID, file1.Path)
		if err != nil {
			t.Fatalf("Failed to remove file1: %v", err)
		}

		// 获取快照
		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		if err != nil {
			t.Fatalf("GetSnapshot failed: %v", err)
		}

		// 验证快照只包含 file2
		if len(snapshot.Files) != 1 {
			t.Fatalf("Expected 1 file in snapshot, got %d", len(snapshot.Files))
		}

		if snapshot.Files[0].Path != file2.Path {
			t.Fatalf("Expected file path %s, got %s", file2.Path, snapshot.Files[0].Path)
		}

		if snapshot.Files[0].RowCount != file2.RowCount {
			t.Fatalf("Expected row count %d, got %d", file2.RowCount, snapshot.Files[0].RowCount)
		}
	})

	t.Run("GetVersionByTimestamp", func(t *testing.T) {
		tableID := "test.orders"

		// 记录开始时间
		startTime := time.Now().UnixMilli()

		file := &delta.ParquetFile{
			Path:     "/data/orders/part-001.parquet",
			Size:     128000,
			RowCount: 100,
		}

		err := deltaLog.AppendAdd(tableID, file)
		if err != nil {
			t.Fatalf("Failed to add file: %v", err)
		}

		currentVersion := deltaLog.GetLatestVersion()

		// 等一小段时间
		time.Sleep(10 * time.Millisecond)

		// 添加另一个条目
		err = deltaLog.AppendRemove(tableID, file.Path)
		if err != nil {
			t.Fatalf("Failed to remove file: %v", err)
		}

		// 获取之前时间点的版本
		version, err := deltaLog.GetVersionByTimestamp(tableID, startTime+5)
		if err != nil {
			t.Fatalf("GetVersionByTimestamp failed: %v", err)
		}

		if version < currentVersion {
			t.Fatalf("Expected version >= %d, got %d", currentVersion, version)
		}
	})

	t.Run("Persistence", func(t *testing.T) {
		// 关闭当前的 Delta Log
		engine.Close()

		// 重新打开存储引擎
		engine2, err := storage.NewParquetEngine(tempDir)
		if err != nil {
			t.Fatalf("Failed to create new ParquetEngine: %v", err)
		}
		defer engine2.Close()

		err = engine2.Open()
		if err != nil {
			t.Fatalf("Failed to open new ParquetEngine: %v", err)
		}

		// 注意：在生产环境中，数据库和表的元数据应该被持久化
		// 这里为了测试，我们需要确保sys数据库存在（通常会自动恢复）
		engine2.CreateDatabase("sys")

		// 检查是否表已存在，如果不存在则需要重新创建表schema
		exists, _ := engine2.TableExists("sys", "delta_log")
		if !exists {
			// 重新创建表
			deltaLogSchema := createDeltaLogSchema()
			err = engine2.CreateTable("sys", "delta_log", deltaLogSchema)
			if err != nil {
				t.Logf("Warning: Failed to recreate delta_log table: %v", err)
			}
		}

		// 创建新的 Delta Log 实例
		deltaLog2 := persistent.NewPersistentDeltaLog(engine2)

		// 验证版本持久化
		version := deltaLog2.GetLatestVersion()
		if version <= 0 {
			t.Fatalf("Expected version > 0 after restart, got %d", version)
		}

		// 验证可以获取之前的快照
		snapshot, err := deltaLog2.GetSnapshot("test.inventory", -1)
		if err != nil {
			t.Fatalf("Failed to get snapshot after restart: %v", err)
		}

		if len(snapshot.Files) != 1 {
			t.Fatalf("Expected 1 file in snapshot after restart, got %d", len(snapshot.Files))
		}
	})

	t.Run("ConcurrentWrites", func(t *testing.T) {
		tableID := "test.concurrent"

		// 并发写入测试
		numGoroutines := 5
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				file := &delta.ParquetFile{
					Path:     filepath.Join("/data/concurrent", "part-"+string(rune('0'+index))+".parquet"),
					Size:     int64(100000 + index*1000),
					RowCount: int64(100 + index*10),
				}

				err := deltaLog.AppendAdd(tableID, file)
				errors <- err
			}(i)
		}

		// 收集所有错误
		for i := 0; i < numGoroutines; i++ {
			if err := <-errors; err != nil {
				t.Fatalf("Concurrent write failed: %v", err)
			}
		}

		// 验证所有文件都被添加
		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		if err != nil {
			t.Fatalf("Failed to get snapshot: %v", err)
		}

		if len(snapshot.Files) != numGoroutines {
			t.Fatalf("Expected %d files, got %d", numGoroutines, len(snapshot.Files))
		}
	})
}

// createDeltaLogSchema 创建 Delta Log 表的 Schema
func createDeltaLogSchema() *arrow.Schema {
	return arrow.NewSchema([]arrow.Field{
		{Name: "version", Type: arrow.PrimitiveTypes.Int64},
		{Name: "timestamp", Type: arrow.PrimitiveTypes.Int64},
		{Name: "table_id", Type: arrow.BinaryTypes.String},
		{Name: "operation", Type: arrow.BinaryTypes.String}, // ADD/REMOVE/METADATA

		// ADD 操作字段
		{Name: "file_path", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "file_size", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
		{Name: "row_count", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
		{Name: "min_values", Type: arrow.BinaryTypes.String, Nullable: true},  // JSON
		{Name: "max_values", Type: arrow.BinaryTypes.String, Nullable: true},  // JSON
		{Name: "null_counts", Type: arrow.BinaryTypes.String, Nullable: true}, // JSON
		{Name: "data_change", Type: arrow.FixedWidthTypes.Boolean, Nullable: true},

		// REMOVE 操作字段
		{Name: "deletion_timestamp", Type: arrow.PrimitiveTypes.Int64, Nullable: true},

		// METADATA 操作字段
		{Name: "schema_json", Type: arrow.BinaryTypes.String, Nullable: true},

		// 审计字段
		{Name: "user_id", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "session_id", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "query_id", Type: arrow.BinaryTypes.String, Nullable: true},
	}, nil)
}

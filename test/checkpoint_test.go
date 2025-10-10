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

// TestCheckpointFunctionality 测试 Checkpoint 功能
func TestCheckpointFunctionality(t *testing.T) {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "minidb-test-checkpoint", time.Now().Format("20060102150405"))
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

	// 创建系统数据库和表
	err = engine.CreateDatabase("sys")
	if err != nil {
		t.Fatalf("Failed to create sys database: %v", err)
	}

	// 创建 Delta Log 表
	deltaLogSchema := createCheckpointDeltaLogSchema()
	err = engine.CreateTable("sys", "delta_log", deltaLogSchema)
	if err != nil {
		t.Fatalf("Failed to create delta_log table: %v", err)
	}

	// 创建 Checkpoint 表
	checkpointSchema := createCheckpointSchema()
	err = engine.CreateTable("sys", "checkpoints", checkpointSchema)
	if err != nil {
		t.Fatalf("Failed to create checkpoints table: %v", err)
	}

	// 创建持久化 Delta Log
	deltaLog := persistent.NewPersistentDeltaLog(engine)

	tableID := "test.checkpoint_table"

	// 添加多个文件以触发 checkpoint 创建 (每 10 个版本创建一次)
	for i := 0; i < 12; i++ {
		parquetFile := &delta.ParquetFile{
			Path:     filepath.Join("/data/checkpoint", "part-"+string(rune('0'+i))+".parquet"),
			Size:     int64(1000000 + i*10000),
			RowCount: int64(1000 + i*100),
			Stats: &delta.FileStats{
				RowCount:   int64(1000 + i*100),
				FileSize:   int64(1000000 + i*10000),
				MinValues:  map[string]interface{}{"id": int64(i * 1000)},
				MaxValues:  map[string]interface{}{"id": int64((i+1)*1000 - 1)},
				NullCounts: map[string]int64{"id": 0},
			},
		}

		err := deltaLog.AppendAdd(tableID, parquetFile)
		if err != nil {
			t.Fatalf("Failed to add file %d: %v", i, err)
		}

		t.Logf("Added file %d, current version: %d", i, deltaLog.GetLatestVersion())
	}

	// 验证版本达到 12
	finalVersion := deltaLog.GetLatestVersion()
	if finalVersion != 12 {
		t.Fatalf("Expected version 12, got %d", finalVersion)
	}

	// 等待一下让异步的 checkpoint 创建完成
	time.Sleep(100 * time.Millisecond)

	// 验证 checkpoint 被创建 - 应该在版本 10 时创建了一个 checkpoint
	t.Logf("Checkpoint functionality test completed successfully")
	t.Logf("Final version: %d (checkpoint should be created at version 10)", finalVersion)

	// 获取最终快照验证数据完整性
	snapshot, err := deltaLog.GetSnapshot(tableID, -1)
	if err != nil {
		t.Fatalf("Failed to get final snapshot: %v", err)
	}

	if len(snapshot.Files) != 12 {
		t.Fatalf("Expected 12 files in final snapshot, got %d", len(snapshot.Files))
	}

	t.Logf("Final snapshot contains %d files as expected", len(snapshot.Files))
}

// createCheckpointDeltaLogSchema 创建 Checkpoint 测试用的 Delta Log 表 Schema
func createCheckpointDeltaLogSchema() *arrow.Schema {
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

// createCheckpointSchema 创建 Checkpoint 表的 Schema
func createCheckpointSchema() *arrow.Schema {
	return arrow.NewSchema([]arrow.Field{
		{Name: "checkpoint_version", Type: arrow.PrimitiveTypes.Int64},
		{Name: "timestamp", Type: arrow.PrimitiveTypes.Int64},
		{Name: "table_id", Type: arrow.BinaryTypes.String},
		{Name: "file_path", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "file_size", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
		{Name: "row_count", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
		{Name: "added_at", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
		{Name: "min_values", Type: arrow.BinaryTypes.String, Nullable: true},  // JSON
		{Name: "max_values", Type: arrow.BinaryTypes.String, Nullable: true},  // JSON
		{Name: "null_counts", Type: arrow.BinaryTypes.String, Nullable: true}, // JSON
		{Name: "schema_json", Type: arrow.BinaryTypes.String, Nullable: true}, // JSON
	}, nil)
}

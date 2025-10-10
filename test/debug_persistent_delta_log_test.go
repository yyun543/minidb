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

// TestPersistentDeltaLogDebug 调试持久化 Delta Log 查询问题
func TestPersistentDeltaLogDebug(t *testing.T) {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "minidb-debug-delta-log", time.Now().Format("20060102150405"))
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
	deltaLogSchema := createDebugDeltaLogSchema()
	err = engine.CreateTable("sys", "delta_log", deltaLogSchema)
	if err != nil {
		t.Fatalf("Failed to create delta_log table: %v", err)
	}

	// 创建持久化 Delta Log
	deltaLog := persistent.NewPersistentDeltaLog(engine)

	tableID := "test.debug"

	// 添加一个文件
	parquetFile := &delta.ParquetFile{
		Path:     "/data/debug/part-001.parquet",
		Size:     1024000,
		RowCount: 1000,
		Stats: &delta.FileStats{
			RowCount:   1000,
			FileSize:   1024000,
			MinValues:  map[string]interface{}{"id": int64(1)},
			MaxValues:  map[string]interface{}{"id": int64(1000)},
			NullCounts: map[string]int64{"id": 0},
		},
	}

	err = deltaLog.AppendAdd(tableID, parquetFile)
	if err != nil {
		t.Fatalf("AppendAdd failed: %v", err)
	}

	t.Logf("Successfully added file to table %s", tableID)

	// 获取当前版本号
	currentVersion := deltaLog.GetLatestVersion()
	t.Logf("Current version: %d", currentVersion)

	// 尝试获取快照，但先不使用过滤器，查看所有数据
	snapshot, err := deltaLog.GetSnapshot(tableID, -1)
	if err != nil {
		t.Fatalf("GetSnapshot failed: %v", err)
	}

	t.Logf("Snapshot retrieved: table=%s, version=%d, files=%d",
		snapshot.TableID, snapshot.Version, len(snapshot.Files))

	if len(snapshot.Files) == 0 {
		t.Fatal("Expected at least 1 file in snapshot, got 0")
	}

	// 验证文件信息
	if snapshot.Files[0].Path != parquetFile.Path {
		t.Fatalf("Expected file path %s, got %s", parquetFile.Path, snapshot.Files[0].Path)
	}
}

// createDebugDeltaLogSchema 创建调试测试用的 Delta Log 表 Schema
func createDebugDeltaLogSchema() *arrow.Schema {
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

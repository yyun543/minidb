package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/delta"
	"github.com/yyun543/minidb/internal/logger"
	"github.com/yyun543/minidb/internal/parquet"
	"go.uber.org/zap"
)

// CreateCheckpoint 创建checkpoint并序列化到Parquet文件
// 实现架构文档建议 (lines 748-764):
// 1. 序列化snapshot到Parquet文件
// 2. 写入_last_checkpoint标记文件
// 3. 可选：清理旧的Delta Log文件
func (pe *ParquetEngine) CreateCheckpoint(tableID string, version int64) error {
	logger.Info("Creating checkpoint for table",
		zap.String("table", tableID),
		zap.Int64("version", version))

	// 获取快照
	snapshot, err := pe.deltaLog.GetSnapshot(tableID, version)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}

	// 创建checkpoints目录
	checkpointDir := filepath.Join(pe.basePath, "sys", "delta_log", "checkpoints")
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		return fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	// 生成checkpoint文件路径
	checkpointPath := pe.getCheckpointPath(tableID, version)

	// 序列化snapshot到Parquet
	if err := pe.serializeCheckpoint(snapshot, checkpointPath); err != nil {
		return fmt.Errorf("failed to serialize checkpoint: %w", err)
	}

	// 写入_last_checkpoint标记文件
	if err := pe.writeCheckpointMarker(tableID, version); err != nil {
		logger.Warn("Failed to write checkpoint marker",
			zap.String("table", tableID),
			zap.Int64("version", version),
			zap.Error(err))
	}

	logger.Info("Checkpoint created successfully",
		zap.String("table", tableID),
		zap.Int64("version", version),
		zap.String("path", checkpointPath),
		zap.Int("file_count", len(snapshot.Files)))

	return nil
}

// serializeCheckpoint 序列化snapshot到Parquet文件
func (pe *ParquetEngine) serializeCheckpoint(snapshot *delta.Snapshot, path string) error {
	// 创建Arrow Schema用于checkpoint
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "file_path", Type: arrow.BinaryTypes.String},
		{Name: "size", Type: arrow.PrimitiveTypes.Int64},
		{Name: "row_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "added_at", Type: arrow.PrimitiveTypes.Int64},
		{Name: "is_delta", Type: arrow.FixedWidthTypes.Boolean},
		{Name: "delta_type", Type: arrow.BinaryTypes.String},
	}, nil)

	// 构建Record
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	for _, file := range snapshot.Files {
		builder.Field(0).(*array.StringBuilder).Append(file.Path)
		builder.Field(1).(*array.Int64Builder).Append(file.Size)
		builder.Field(2).(*array.Int64Builder).Append(file.RowCount)
		builder.Field(3).(*array.Int64Builder).Append(file.AddedAt)
		builder.Field(4).(*array.BooleanBuilder).Append(file.IsDelta)
		builder.Field(5).(*array.StringBuilder).Append(file.DeltaType)
	}

	record := builder.NewRecord()
	defer record.Release()

	// 写入Parquet文件（使用带fsync的writer）
	_, err := parquet.WriteArrowBatch(path, record)
	if err != nil {
		return fmt.Errorf("failed to write checkpoint parquet: %w", err)
	}

	logger.Info("Checkpoint parquet file written",
		zap.String("path", path),
		zap.Int("file_count", len(snapshot.Files)))

	return nil
}

// writeCheckpointMarker 写入_last_checkpoint标记文件
func (pe *ParquetEngine) writeCheckpointMarker(tableID string, version int64) error {
	checkpointDir := filepath.Join(pe.basePath, "sys", "delta_log", "checkpoints")
	markerPath := filepath.Join(checkpointDir, fmt.Sprintf("_last_checkpoint.%s", tableID))

	// 写入版本号
	content := fmt.Sprintf("%d", version)
	if err := os.WriteFile(markerPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write marker file: %w", err)
	}

	// Sync目录确保元数据持久化
	dir, err := os.Open(checkpointDir)
	if err == nil {
		dir.Sync()
		dir.Close()
	}

	return nil
}

// LoadLatestCheckpoint 加载最新的checkpoint
func (pe *ParquetEngine) LoadLatestCheckpoint(tableID string) (*delta.Snapshot, error) {
	checkpointDir := filepath.Join(pe.basePath, "sys", "delta_log", "checkpoints")
	markerPath := filepath.Join(checkpointDir, fmt.Sprintf("_last_checkpoint.%s", tableID))

	// 读取_last_checkpoint标记文件
	data, err := os.ReadFile(markerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 没有checkpoint
		}
		return nil, fmt.Errorf("failed to read checkpoint marker: %w", err)
	}

	var version int64
	if _, err := fmt.Sscanf(string(data), "%d", &version); err != nil {
		return nil, fmt.Errorf("failed to parse checkpoint version: %w", err)
	}

	// 加载checkpoint文件
	checkpointPath := pe.getCheckpointPath(tableID, version)
	return pe.loadCheckpointFile(tableID, version, checkpointPath)
}

// loadCheckpointFile 从Parquet文件加载checkpoint
func (pe *ParquetEngine) loadCheckpointFile(tableID string, version int64, path string) (*delta.Snapshot, error) {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil // checkpoint文件不存在
	}

	// 读取Parquet文件
	record, err := parquet.ReadParquetFile(path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %w", err)
	}
	defer record.Release()

	// 解析文件信息
	files := make([]delta.FileInfo, 0, int(record.NumRows()))

	for i := 0; i < int(record.NumRows()); i++ {
		file := delta.FileInfo{
			Path:       record.Column(0).(*array.String).Value(i),
			Size:       record.Column(1).(*array.Int64).Value(i),
			RowCount:   record.Column(2).(*array.Int64).Value(i),
			AddedAt:    record.Column(3).(*array.Int64).Value(i),
			IsDelta:    record.Column(4).(*array.Boolean).Value(i),
			DeltaType:  record.Column(5).(*array.String).Value(i),
			MinValues:  make(map[string]interface{}),
			MaxValues:  make(map[string]interface{}),
			NullCounts: make(map[string]int64),
		}

		files = append(files, file)
	}

	snapshot := &delta.Snapshot{
		Version: version,
		TableID: tableID,
		Files:   files,
	}

	logger.Info("Checkpoint loaded successfully",
		zap.String("table", tableID),
		zap.Int64("version", version),
		zap.Int("file_count", len(files)))

	return snapshot, nil
}

// getCheckpointPath 获取checkpoint文件路径
func (pe *ParquetEngine) getCheckpointPath(tableID string, version int64) string {
	checkpointDir := filepath.Join(pe.basePath, "sys", "delta_log", "checkpoints")
	filename := fmt.Sprintf("_checkpoint.%s.%020d.parquet", tableID, version)
	return filepath.Join(checkpointDir, filename)
}

package persistent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/delta"
	"github.com/yyun543/minidb/internal/logger"
	"github.com/yyun543/minidb/internal/storage"
	"go.uber.org/zap"
)

// PersistentDeltaLog 持久化的 Delta Log 实现
// 使用存储引擎将 Delta Log 存储为结构化表
type PersistentDeltaLog struct {
	storageEngine storage.StorageEngine
	systemDB      string // "sys"
	deltaTable    string // "delta_log"
	mu            sync.RWMutex
	currentVer    atomic.Int64
}

// NewPersistentDeltaLog 创建持久化 Delta Log 管理器
func NewPersistentDeltaLog(engine storage.StorageEngine) *PersistentDeltaLog {
	dl := &PersistentDeltaLog{
		storageEngine: engine,
		systemDB:      "sys",
		deltaTable:    "delta_log",
	}

	// 初始化时从存储中恢复最新版本号
	dl.loadLatestVersion()

	return dl
}

// loadLatestVersion 从存储中恢复最新版本号
func (dl *PersistentDeltaLog) loadLatestVersion() {
	// 查询最新版本
	filters := []storage.Filter{}

	// 尝试扫描 delta_log 表
	iterator, err := dl.storageEngine.Scan(context.Background(), dl.systemDB, dl.deltaTable, filters)
	if err != nil {
		// 表可能不存在，从版本 0 开始
		dl.currentVer.Store(0)
		return
	}
	defer iterator.Close()

	var maxVersion int64 = 0

	// 遍历所有记录找到最大版本号
	for iterator.Next() {
		record := iterator.Record()
		versionCol := record.Column(0) // version 字段在第 0 列

		if versionArray, ok := versionCol.(*array.Int64); ok {
			for i := 0; i < versionArray.Len(); i++ {
				if !versionArray.IsNull(i) {
					version := versionArray.Value(i)
					if version > maxVersion {
						maxVersion = version
					}
				}
			}
		}
	}

	dl.currentVer.Store(maxVersion)

	logger.Info("Loaded latest version from persistent storage",
		zap.Int64("version", maxVersion))
}

// AppendAdd 追加 ADD 操作
func (dl *PersistentDeltaLog) AppendAdd(tableID string, file *delta.ParquetFile) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	version := dl.currentVer.Add(1)
	timestamp := time.Now().UnixMilli()

	// 构造 Delta Log 记录
	builder := array.NewRecordBuilder(memory.DefaultAllocator, createDeltaLogSchema())
	defer builder.Release()

	// 填充基本字段
	builder.Field(0).(*array.Int64Builder).Append(version)
	builder.Field(1).(*array.Int64Builder).Append(timestamp)
	builder.Field(2).(*array.StringBuilder).Append(tableID)
	builder.Field(3).(*array.StringBuilder).Append(string(delta.OpAdd))

	// 填充 ADD 操作相关字段
	builder.Field(4).(*array.StringBuilder).Append(file.Path)
	builder.Field(5).(*array.Int64Builder).Append(file.Size)
	builder.Field(6).(*array.Int64Builder).Append(file.RowCount)

	// 序列化统计信息为 JSON
	var minValuesJSON, maxValuesJSON, nullCountsJSON string
	if file.Stats != nil {
		if minData, err := json.Marshal(file.Stats.MinValues); err == nil {
			minValuesJSON = string(minData)
		}
		if maxData, err := json.Marshal(file.Stats.MaxValues); err == nil {
			maxValuesJSON = string(maxData)
		}
		if nullData, err := json.Marshal(file.Stats.NullCounts); err == nil {
			nullCountsJSON = string(nullData)
		}
	}

	builder.Field(7).(*array.StringBuilder).Append(minValuesJSON)
	builder.Field(8).(*array.StringBuilder).Append(maxValuesJSON)
	builder.Field(9).(*array.StringBuilder).Append(nullCountsJSON)
	builder.Field(10).(*array.BooleanBuilder).Append(true) // data_change

	// 其他字段设为 null
	for i := 11; i < 16; i++ {
		builder.Field(i).AppendNull()
	}

	record := builder.NewRecord()
	defer record.Release()

	// 写入到存储引擎
	if err := dl.storageEngine.Write(context.Background(), dl.systemDB, dl.deltaTable, record); err != nil {
		// 回滚版本号
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to write ADD log entry: %w", err)
	}

	logger.Info("Delta Log entry appended",
		zap.Int64("version", version),
		zap.String("table", tableID),
		zap.String("operation", string(delta.OpAdd)),
		zap.String("file", file.Path))

	// 检查是否需要创建 Checkpoint
	if version%10 == 0 {
		go dl.createCheckpoint(tableID, version)
	}

	return nil
}

// AppendRemove 追加 REMOVE 操作
func (dl *PersistentDeltaLog) AppendRemove(tableID, filePath string) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	version := dl.currentVer.Add(1)
	timestamp := time.Now().UnixMilli()

	// 构造 Delta Log 记录
	builder := array.NewRecordBuilder(memory.DefaultAllocator, createDeltaLogSchema())
	defer builder.Release()

	// 填充基本字段
	builder.Field(0).(*array.Int64Builder).Append(version)
	builder.Field(1).(*array.Int64Builder).Append(timestamp)
	builder.Field(2).(*array.StringBuilder).Append(tableID)
	builder.Field(3).(*array.StringBuilder).Append(string(delta.OpRemove))

	// REMOVE 操作：file_path 和 deletion_timestamp
	builder.Field(4).(*array.StringBuilder).Append(filePath)
	for i := 5; i < 11; i++ {
		builder.Field(i).AppendNull() // ADD 相关字段为 null
	}
	builder.Field(11).(*array.Int64Builder).Append(timestamp) // deletion_timestamp

	// 其他字段设为 null
	for i := 12; i < 16; i++ {
		builder.Field(i).AppendNull()
	}

	record := builder.NewRecord()
	defer record.Release()

	// 写入到存储引擎
	if err := dl.storageEngine.Write(context.Background(), dl.systemDB, dl.deltaTable, record); err != nil {
		// 回滚版本号
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to write REMOVE log entry: %w", err)
	}

	logger.Info("Delta Log entry appended",
		zap.Int64("version", version),
		zap.String("table", tableID),
		zap.String("operation", string(delta.OpRemove)),
		zap.String("file", filePath))

	return nil
}

// AppendMetadata 追加 METADATA 操作
func (dl *PersistentDeltaLog) AppendMetadata(tableID string, schema *arrow.Schema) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	version := dl.currentVer.Add(1)
	timestamp := time.Now().UnixMilli()

	schemaJSON, err := delta.SchemaToJSON(schema)
	if err != nil {
		// 回滚版本号
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to serialize schema: %w", err)
	}

	// 构造 Delta Log 记录
	builder := array.NewRecordBuilder(memory.DefaultAllocator, createDeltaLogSchema())
	defer builder.Release()

	// 填充基本字段
	builder.Field(0).(*array.Int64Builder).Append(version)
	builder.Field(1).(*array.Int64Builder).Append(timestamp)
	builder.Field(2).(*array.StringBuilder).Append(tableID)
	builder.Field(3).(*array.StringBuilder).Append(string(delta.OpMetadata))

	// METADATA 操作：只有 schema_json
	for i := 4; i < 12; i++ {
		builder.Field(i).AppendNull() // ADD/REMOVE 相关字段为 null
	}
	builder.Field(12).(*array.StringBuilder).Append(schemaJSON) // schema_json

	// 其他字段设为 null
	for i := 13; i < 16; i++ {
		builder.Field(i).AppendNull()
	}

	record := builder.NewRecord()
	defer record.Release()

	// 写入到存储引擎
	if err := dl.storageEngine.Write(context.Background(), dl.systemDB, dl.deltaTable, record); err != nil {
		// 回滚版本号
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to write METADATA log entry: %w", err)
	}

	logger.Info("Delta Log entry appended",
		zap.Int64("version", version),
		zap.String("table", tableID),
		zap.String("operation", string(delta.OpMetadata)))

	return nil
}

// GetSnapshot 获取表快照
func (dl *PersistentDeltaLog) GetSnapshot(tableID string, version int64) (*delta.Snapshot, error) {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	// 如果版本为 -1，使用最新版本
	if version == -1 {
		version = dl.currentVer.Load()
	}

	// 构建查询过滤器 - 临时禁用过滤器用于调试
	filters := []storage.Filter{
		// {Column: "table_id", Operator: "=", Value: tableID},
		// {Column: "version", Operator: "<=", Value: version},
	}

	// 扫描 Delta Log 表
	iterator, err := dl.storageEngine.Scan(context.Background(), dl.systemDB, dl.deltaTable, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to scan delta log: %w", err)
	}
	defer iterator.Close()

	snapshot := &delta.Snapshot{
		Version:   version,
		Timestamp: time.Now(),
		TableID:   tableID,
		Files:     make([]delta.FileInfo, 0),
	}

	// 构建快照：所有 ADD 文件 - REMOVE 文件
	addedFiles := make(map[string]delta.FileInfo)
	removedFiles := make(map[string]bool)

	for iterator.Next() {
		record := iterator.Record()

		// 解析每条记录
		for rowIdx := 0; rowIdx < int(record.NumRows()); rowIdx++ {
			entry, err := dl.parseLogEntry(record, rowIdx)
			if err != nil {
				logger.Warn("Failed to parse log entry", zap.Error(err))
				continue
			}

			// 应用过滤条件 (因为存储层过滤器暂时禁用)
			if entry.TableID != tableID || entry.Version > version {
				continue
			}

			switch entry.Operation {
			case delta.OpAdd:
				fileInfo := delta.FileInfo{
					Path:     entry.FilePath,
					Size:     entry.FileSize,
					RowCount: entry.RowCount,
					AddedAt:  entry.Timestamp,
				}

				// 反序列化统计信息
				if entry.MinValues != nil {
					fileInfo.MinValues = entry.MinValues
				}
				if entry.MaxValues != nil {
					fileInfo.MaxValues = entry.MaxValues
				}
				if entry.NullCounts != nil {
					fileInfo.NullCounts = entry.NullCounts
				}

				addedFiles[entry.FilePath] = fileInfo

			case delta.OpRemove:
				removedFiles[entry.FilePath] = true

			case delta.OpMetadata:
				if entry.SchemaJSON != "" {
					if schema, err := delta.SchemaFromJSON(entry.SchemaJSON); err == nil {
						snapshot.Schema = schema
					}
				}
			}
		}
	}

	// 过滤掉已删除的文件
	for path, file := range addedFiles {
		if !removedFiles[path] {
			snapshot.Files = append(snapshot.Files, file)
		}
	}

	logger.Info("Snapshot retrieved from persistent storage",
		zap.String("table", tableID),
		zap.Int64("version", version),
		zap.Int("file_count", len(snapshot.Files)))

	return snapshot, nil
}

// parseLogEntry 解析日志记录
func (dl *PersistentDeltaLog) parseLogEntry(record arrow.Record, rowIdx int) (*delta.LogEntry, error) {
	entry := &delta.LogEntry{}

	// 基本字段
	if versionCol, ok := record.Column(0).(*array.Int64); ok && !versionCol.IsNull(rowIdx) {
		entry.Version = versionCol.Value(rowIdx)
	}
	if timestampCol, ok := record.Column(1).(*array.Int64); ok && !timestampCol.IsNull(rowIdx) {
		entry.Timestamp = timestampCol.Value(rowIdx)
	}
	if tableCol, ok := record.Column(2).(*array.String); ok && !tableCol.IsNull(rowIdx) {
		entry.TableID = tableCol.Value(rowIdx)
	}
	if opCol, ok := record.Column(3).(*array.String); ok && !opCol.IsNull(rowIdx) {
		entry.Operation = delta.Operation(opCol.Value(rowIdx))
	}

	// ADD/REMOVE 字段
	if pathCol, ok := record.Column(4).(*array.String); ok && !pathCol.IsNull(rowIdx) {
		entry.FilePath = pathCol.Value(rowIdx)
	}
	if sizeCol, ok := record.Column(5).(*array.Int64); ok && !sizeCol.IsNull(rowIdx) {
		entry.FileSize = sizeCol.Value(rowIdx)
	}
	if rowCountCol, ok := record.Column(6).(*array.Int64); ok && !rowCountCol.IsNull(rowIdx) {
		entry.RowCount = rowCountCol.Value(rowIdx)
	}

	// 统计信息
	if minCol, ok := record.Column(7).(*array.String); ok && !minCol.IsNull(rowIdx) {
		var minValues map[string]interface{}
		if err := json.Unmarshal([]byte(minCol.Value(rowIdx)), &minValues); err == nil {
			entry.MinValues = minValues
		}
	}
	if maxCol, ok := record.Column(8).(*array.String); ok && !maxCol.IsNull(rowIdx) {
		var maxValues map[string]interface{}
		if err := json.Unmarshal([]byte(maxCol.Value(rowIdx)), &maxValues); err == nil {
			entry.MaxValues = maxValues
		}
	}
	if nullCol, ok := record.Column(9).(*array.String); ok && !nullCol.IsNull(rowIdx) {
		var nullCounts map[string]int64
		if err := json.Unmarshal([]byte(nullCol.Value(rowIdx)), &nullCounts); err == nil {
			entry.NullCounts = nullCounts
		}
	}

	// REMOVE 字段
	if delCol, ok := record.Column(11).(*array.Int64); ok && !delCol.IsNull(rowIdx) {
		entry.DeletionTimestamp = delCol.Value(rowIdx)
	}

	// METADATA 字段
	if schemaCol, ok := record.Column(12).(*array.String); ok && !schemaCol.IsNull(rowIdx) {
		entry.SchemaJSON = schemaCol.Value(rowIdx)
	}

	return entry, nil
}

// GetLatestVersion 获取最新版本号
func (dl *PersistentDeltaLog) GetLatestVersion() int64 {
	return dl.currentVer.Load()
}

// GetVersionByTimestamp 根据时间戳查找版本号
func (dl *PersistentDeltaLog) GetVersionByTimestamp(tableID string, ts int64) (int64, error) {
	// 构建查询过滤器 - 临时禁用过滤器用于调试
	filters := []storage.Filter{
		// {Column: "table_id", Operator: "=", Value: tableID},
		// {Column: "timestamp", Operator: "<=", Value: ts},
	}

	// 扫描 Delta Log 表
	iterator, err := dl.storageEngine.Scan(context.Background(), dl.systemDB, dl.deltaTable, filters)
	if err != nil {
		return 0, fmt.Errorf("failed to scan delta log: %w", err)
	}
	defer iterator.Close()

	var maxVersion int64 = 0

	for iterator.Next() {
		record := iterator.Record()

		for rowIdx := 0; rowIdx < int(record.NumRows()); rowIdx++ {
			entry, err := dl.parseLogEntry(record, rowIdx)
			if err != nil {
				continue
			}

			// 应用过滤条件 (因为存储层过滤器暂时禁用)
			if entry.TableID == tableID && entry.Timestamp <= ts && entry.Version > maxVersion {
				maxVersion = entry.Version
			}
		}
	}

	if maxVersion == 0 {
		return 0, fmt.Errorf("no version found before timestamp %d", ts)
	}

	return maxVersion, nil
}

// createCheckpoint 创建检查点 (完整实现)
func (dl *PersistentDeltaLog) createCheckpoint(tableID string, version int64) {
	logger.Info("Creating checkpoint",
		zap.String("table", tableID),
		zap.Int64("version", version))

	// 获取指定版本的快照
	snapshot, err := dl.GetSnapshot(tableID, version)
	if err != nil {
		logger.Error("Failed to get snapshot for checkpoint",
			zap.String("table", tableID),
			zap.Int64("version", version),
			zap.Error(err))
		return
	}

	// 创建 Checkpoint Parquet 文件
	err = dl.writeCheckpointFile(tableID, version, snapshot)
	if err != nil {
		logger.Error("Failed to write checkpoint file",
			zap.String("table", tableID),
			zap.Int64("version", version),
			zap.Error(err))
		return
	}

	logger.Info("Checkpoint created successfully",
		zap.String("table", tableID),
		zap.Int64("version", version),
		zap.Int("file_count", len(snapshot.Files)))
}

// writeCheckpointFile 将快照写入 Checkpoint Parquet 文件
func (dl *PersistentDeltaLog) writeCheckpointFile(tableID string, version int64, snapshot *delta.Snapshot) error {
	// 构建 Checkpoint 记录的 Schema
	checkpointSchema := createCheckpointSchema()

	// 创建 Arrow Record Builder
	builder := array.NewRecordBuilder(memory.DefaultAllocator, checkpointSchema)
	defer builder.Release()

	// 为每个文件创建一条记录
	for _, file := range snapshot.Files {
		// 基本信息
		builder.Field(0).(*array.Int64Builder).Append(version)                // checkpoint_version
		builder.Field(1).(*array.Int64Builder).Append(time.Now().UnixMilli()) // timestamp
		builder.Field(2).(*array.StringBuilder).Append(tableID)               // table_id
		builder.Field(3).(*array.StringBuilder).Append(file.Path)             // file_path
		builder.Field(4).(*array.Int64Builder).Append(file.Size)              // file_size
		builder.Field(5).(*array.Int64Builder).Append(file.RowCount)          // row_count
		builder.Field(6).(*array.Int64Builder).Append(file.AddedAt)           // added_at

		// 序列化统计信息
		var minValuesJSON, maxValuesJSON, nullCountsJSON string
		if file.MinValues != nil {
			if minData, err := json.Marshal(file.MinValues); err == nil {
				minValuesJSON = string(minData)
			}
		}
		if file.MaxValues != nil {
			if maxData, err := json.Marshal(file.MaxValues); err == nil {
				maxValuesJSON = string(maxData)
			}
		}
		if file.NullCounts != nil {
			if nullData, err := json.Marshal(file.NullCounts); err == nil {
				nullCountsJSON = string(nullData)
			}
		}

		builder.Field(7).(*array.StringBuilder).Append(minValuesJSON)  // min_values
		builder.Field(8).(*array.StringBuilder).Append(maxValuesJSON)  // max_values
		builder.Field(9).(*array.StringBuilder).Append(nullCountsJSON) // null_counts

		// Schema 信息
		var schemaJSON string
		if snapshot.Schema != nil {
			if schemaData, err := delta.SchemaToJSON(snapshot.Schema); err == nil {
				schemaJSON = schemaData
			}
		}
		builder.Field(10).(*array.StringBuilder).Append(schemaJSON) // schema_json
	}

	// 如果没有文件，创建一个空的检查点记录
	if len(snapshot.Files) == 0 {
		builder.Field(0).(*array.Int64Builder).Append(version)
		builder.Field(1).(*array.Int64Builder).Append(time.Now().UnixMilli())
		builder.Field(2).(*array.StringBuilder).Append(tableID)
		for i := 3; i < 11; i++ {
			builder.Field(i).AppendNull()
		}
	}

	record := builder.NewRecord()
	defer record.Release()

	// 写入到 sys.checkpoints 表
	checkpointTable := "checkpoints"
	return dl.storageEngine.Write(context.Background(), dl.systemDB, checkpointTable, record)
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

// ListTables 列出所有表
func (dl *PersistentDeltaLog) ListTables() []string {
	// 查询所有唯一的 table_id
	iterator, err := dl.storageEngine.Scan(context.Background(), dl.systemDB, dl.deltaTable, []storage.Filter{})
	if err != nil {
		logger.Error("Failed to scan delta log for table list", zap.Error(err))
		return []string{}
	}
	defer iterator.Close()

	tableSet := make(map[string]bool)

	for iterator.Next() {
		record := iterator.Record()

		for rowIdx := 0; rowIdx < int(record.NumRows()); rowIdx++ {
			entry, err := dl.parseLogEntry(record, rowIdx)
			if err != nil {
				continue
			}

			if entry.TableID != "" {
				tableSet[entry.TableID] = true
			}
		}
	}

	tables := make([]string, 0, len(tableSet))
	for table := range tableSet {
		tables = append(tables, table)
	}

	return tables
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

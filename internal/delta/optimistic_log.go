package delta

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/yyun543/minidb/internal/logger"
	"github.com/yyun543/minidb/internal/objectstore"
	"go.uber.org/zap"
)

// OptimisticDeltaLog 实现乐观并发控制的Delta Log
// 核心思想:
// 1. 每个commit写入独立的版本文件 (000001.json, 000002.json, ...)
// 2. 使用对象存储的PutIfNotExists确保版本唯一性
// 3. 发生冲突时返回RetryableConflictError
// 4. 客户端自动重试
type OptimisticDeltaLog struct {
	objectStore objectstore.ConditionalObjectStore
	basePath    string
	currentVer  atomic.Int64

	checkpointCallback CheckpointCallback
}

// ConflictError 表示版本冲突错误（可重试）
type ConflictError struct {
	Version int64
	Message string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("version conflict at V%d: %s", e.Version, e.Message)
}

// NewOptimisticDeltaLog 创建乐观并发控制的Delta Log
func NewOptimisticDeltaLog(objectStore objectstore.ConditionalObjectStore, basePath string) *OptimisticDeltaLog {
	dl := &OptimisticDeltaLog{
		objectStore: objectStore,
		basePath:    basePath,
	}
	dl.currentVer.Store(0)
	return dl
}

// SetCheckpointCallback 设置checkpoint回调
func (dl *OptimisticDeltaLog) SetCheckpointCallback(callback CheckpointCallback) {
	dl.checkpointCallback = callback
}

// Bootstrap 初始化Delta Log（扫描现有版本文件）
func (dl *OptimisticDeltaLog) Bootstrap() error {
	logger.Info("Bootstrapping OptimisticDeltaLog")

	// 扫描_delta_log目录找到最新版本
	// 格式: {basePath}/sys/_delta_log/000001.json
	deltaLogDir := fmt.Sprintf("%s/sys/_delta_log", dl.basePath)

	files, err := dl.objectStore.List(deltaLogDir)
	if err != nil {
		logger.Warn("Failed to list delta log files, starting fresh", zap.Error(err))
		return nil
	}

	maxVersion := int64(0)
	for _, file := range files {
		var version int64
		// 尝试解析文件名格式: 000001.json
		if _, err := fmt.Sscanf(file, "%d.json", &version); err == nil {
			if version > maxVersion {
				maxVersion = version
			}
		}
	}

	dl.currentVer.Store(maxVersion)
	logger.Info("OptimisticDeltaLog bootstrapped",
		zap.Int64("latest_version", maxVersion),
		zap.Int("file_count", len(files)))

	return nil
}

// RestoreFromEntries 从已加载的entries恢复状态（兼容性方法）
func (dl *OptimisticDeltaLog) RestoreFromEntries(entries []LogEntry) error {
	// 乐观锁实现中，entries存储在对象存储中
	// 这个方法主要用于兼容性
	var maxVersion int64 = 0
	for _, entry := range entries {
		if entry.Version > maxVersion {
			maxVersion = entry.Version
		}
	}
	dl.currentVer.Store(maxVersion)
	return nil
}

// AppendAdd 追加ADD操作（乐观并发控制）
func (dl *OptimisticDeltaLog) AppendAdd(tableID string, file *ParquetFile) error {
	// 1. 生成新版本号（无锁）
	version := dl.currentVer.Add(1)
	timestamp := time.Now().UnixMilli()

	entry := LogEntry{
		Version:    version,
		Timestamp:  timestamp,
		TableID:    tableID,
		Operation:  OpAdd,
		FilePath:   file.Path,
		FileSize:   file.Size,
		RowCount:   file.RowCount,
		DataChange: true,
		IsDelta:    file.IsDelta,
		DeltaType:  file.DeltaType,
	}

	if file.Stats != nil {
		entry.MinValues = file.Stats.MinValues
		entry.MaxValues = file.Stats.MaxValues
		entry.NullCounts = file.Stats.NullCounts
	}

	// 2. 序列化entry为JSON
	data, err := json.Marshal(entry)
	if err != nil {
		dl.currentVer.Add(-1) // 回滚版本号
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// 3. 使用PutIfNotExists写入版本文件
	versionFilePath := dl.getVersionFilePath(tableID, version)
	err = dl.objectStore.PutIfNotExists(versionFilePath, data)
	if err != nil {
		// 检查是否是冲突错误
		if isConflictError(err) {
			dl.currentVer.Add(-1) // 回滚版本号
			logger.Warn("Version conflict detected, retry needed",
				zap.Int64("version", version),
				zap.String("table", tableID),
				zap.Error(err))
			return &ConflictError{
				Version: version,
				Message: "another writer committed this version first",
			}
		}
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to write version file: %w", err)
	}

	logger.Info("Delta Log entry committed (optimistic)",
		zap.Int64("version", version),
		zap.String("table", tableID),
		zap.String("operation", string(OpAdd)),
		zap.String("file", file.Path))

	// 4. 检查是否需要创建checkpoint
	if version%10 == 0 {
		if dl.checkpointCallback != nil {
			go dl.checkpointCallback(tableID, version)
		}
	}

	return nil
}

// AppendRemove 追加REMOVE操作
func (dl *OptimisticDeltaLog) AppendRemove(tableID, filePath string) error {
	version := dl.currentVer.Add(1)
	timestamp := time.Now().UnixMilli()

	entry := LogEntry{
		Version:           version,
		Timestamp:         timestamp,
		TableID:           tableID,
		Operation:         OpRemove,
		FilePath:          filePath,
		DeletionTimestamp: timestamp,
		DataChange:        true,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	versionFilePath := dl.getVersionFilePath(tableID, version)
	err = dl.objectStore.PutIfNotExists(versionFilePath, data)
	if err != nil {
		if isConflictError(err) {
			dl.currentVer.Add(-1)
			return &ConflictError{
				Version: version,
				Message: "version conflict on REMOVE operation",
			}
		}
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to write version file: %w", err)
	}

	logger.Info("Delta Log entry committed (optimistic)",
		zap.Int64("version", version),
		zap.String("table", tableID),
		zap.String("operation", string(OpRemove)),
		zap.String("file", filePath))

	return nil
}

// AppendMetadata 追加METADATA操作
func (dl *OptimisticDeltaLog) AppendMetadata(tableID string, schema *arrow.Schema) error {
	version := dl.currentVer.Add(1)
	timestamp := time.Now().UnixMilli()

	schemaJSON, err := SchemaToJSON(schema)
	if err != nil {
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to serialize schema: %w", err)
	}

	entry := LogEntry{
		Version:    version,
		Timestamp:  timestamp,
		TableID:    tableID,
		Operation:  OpMetadata,
		SchemaJSON: schemaJSON,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	versionFilePath := dl.getVersionFilePath(tableID, version)
	err = dl.objectStore.PutIfNotExists(versionFilePath, data)
	if err != nil {
		if isConflictError(err) {
			dl.currentVer.Add(-1)
			return &ConflictError{
				Version: version,
				Message: "version conflict on METADATA operation",
			}
		}
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to write version file: %w", err)
	}

	logger.Info("Delta Log entry committed (optimistic)",
		zap.Int64("version", version),
		zap.String("table", tableID),
		zap.String("operation", string(OpMetadata)))

	return nil
}

// AppendIndexMetadata 追加索引元数据操作
func (dl *OptimisticDeltaLog) AppendIndexMetadata(tableID, indexName string, indexMeta map[string]interface{}) error {
	version := dl.currentVer.Add(1)
	timestamp := time.Now().UnixMilli()

	indexJSON, err := json.Marshal(map[string]interface{}{
		"index_name": indexName,
		"table_id":   tableID,
		"metadata":   indexMeta,
	})
	if err != nil {
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to marshal index metadata: %w", err)
	}

	entry := LogEntry{
		Version:   version,
		Timestamp: timestamp,
		TableID:   tableID,
		Operation: OpMetadata,
		IndexJSON: string(indexJSON),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	versionFilePath := dl.getVersionFilePath(tableID, version)
	err = dl.objectStore.PutIfNotExists(versionFilePath, data)
	if err != nil {
		if isConflictError(err) {
			dl.currentVer.Add(-1)
			return &ConflictError{
				Version: version,
				Message: "version conflict on INDEX METADATA operation",
			}
		}
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to write version file: %w", err)
	}

	logger.Info("Delta Log INDEX METADATA entry committed (optimistic)",
		zap.Int64("version", version),
		zap.String("table", tableID),
		zap.String("index", indexName))

	return nil
}

// RemoveIndexMetadata 删除索引元数据操作
func (dl *OptimisticDeltaLog) RemoveIndexMetadata(tableID, indexName string) error {
	version := dl.currentVer.Add(1)
	timestamp := time.Now().UnixMilli()

	indexJSON, err := json.Marshal(map[string]interface{}{
		"index_name": indexName,
		"table_id":   tableID,
	})
	if err != nil {
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to marshal index deletion metadata: %w", err)
	}

	entry := LogEntry{
		Version:        version,
		Timestamp:      timestamp,
		TableID:        tableID,
		Operation:      OpMetadata,
		IndexJSON:      string(indexJSON),
		IndexOperation: "DROP",
	}

	data, err := json.Marshal(entry)
	if err != nil {
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	versionFilePath := dl.getVersionFilePath(tableID, version)
	err = dl.objectStore.PutIfNotExists(versionFilePath, data)
	if err != nil {
		if isConflictError(err) {
			dl.currentVer.Add(-1)
			return &ConflictError{
				Version: version,
				Message: "version conflict on INDEX DROP operation",
			}
		}
		dl.currentVer.Add(-1)
		return fmt.Errorf("failed to write version file: %w", err)
	}

	logger.Info("Delta Log INDEX DROP entry committed (optimistic)",
		zap.Int64("version", version),
		zap.String("table", tableID),
		zap.String("index", indexName))

	return nil
}

// GetSnapshot 获取表快照（读取所有版本文件）
func (dl *OptimisticDeltaLog) GetSnapshot(tableID string, version int64) (*Snapshot, error) {
	if version == -1 {
		version = dl.currentVer.Load()
	}

	snapshot := &Snapshot{
		Version:   version,
		Timestamp: time.Now(),
		TableID:   tableID,
		Files:     make([]FileInfo, 0),
	}

	// 读取所有版本文件 (从1到version)
	addedFiles := make(map[string]FileInfo)
	removedFiles := make(map[string]bool)

	for v := int64(1); v <= version; v++ {
		versionFilePath := dl.getVersionFilePath(tableID, v)
		data, err := dl.objectStore.Get(versionFilePath)
		if err != nil {
			// 版本文件不存在，跳过（可能是其他表的版本）
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			logger.Warn("Failed to unmarshal log entry",
				zap.String("file", versionFilePath),
				zap.Error(err))
			continue
		}

		// 只处理指定表的条目
		if entry.TableID != tableID {
			continue
		}

		switch entry.Operation {
		case OpAdd:
			addedFiles[entry.FilePath] = FileInfo{
				Path:       entry.FilePath,
				Size:       entry.FileSize,
				RowCount:   entry.RowCount,
				MinValues:  entry.MinValues,
				MaxValues:  entry.MaxValues,
				NullCounts: entry.NullCounts,
				AddedAt:    entry.Timestamp,
				IsDelta:    entry.IsDelta,
				DeltaType:  entry.DeltaType,
			}

		case OpRemove:
			removedFiles[entry.FilePath] = true

		case OpMetadata:
			if entry.SchemaJSON != "" {
				schema, err := SchemaFromJSON(entry.SchemaJSON)
				if err == nil {
					snapshot.Schema = schema
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

	logger.Info("Snapshot retrieved (optimistic)",
		zap.String("table", tableID),
		zap.Int64("version", version),
		zap.Int("file_count", len(snapshot.Files)))

	return snapshot, nil
}

// GetLatestVersion 获取最新版本号
func (dl *OptimisticDeltaLog) GetLatestVersion() int64 {
	return dl.currentVer.Load()
}

// GetVersionByTimestamp 根据时间戳查找版本号
func (dl *OptimisticDeltaLog) GetVersionByTimestamp(tableID string, ts int64) (int64, error) {
	// 二分查找版本号
	// 简化实现：线性扫描
	latestVersion := dl.currentVer.Load()

	for v := latestVersion; v >= 1; v-- {
		versionFilePath := dl.getVersionFilePath(tableID, v)
		data, err := dl.objectStore.Get(versionFilePath)
		if err != nil {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		if entry.TableID == tableID && entry.Timestamp <= ts {
			return v, nil
		}
	}

	return 0, fmt.Errorf("no version found before timestamp %d", ts)
}

// ListTables 列出所有表
func (dl *OptimisticDeltaLog) ListTables() []string {
	// 扫描所有版本文件，提取唯一的table ID
	latestVersion := dl.currentVer.Load()
	tableSet := make(map[string]bool)

	for v := int64(1); v <= latestVersion; v++ {
		// 构建版本文件路径模式
		versionFilePath := fmt.Sprintf("%s/sys/_delta_log/%020d.json", dl.basePath, v)
		data, err := dl.objectStore.Get(versionFilePath)
		if err != nil {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		if entry.TableID != "" {
			tableSet[entry.TableID] = true
		}
	}

	tables := make([]string, 0, len(tableSet))
	for table := range tableSet {
		tables = append(tables, table)
	}

	return tables
}

// GetAllEntries 获取所有日志条目
func (dl *OptimisticDeltaLog) GetAllEntries() []LogEntry {
	latestVersion := dl.currentVer.Load()
	entries := make([]LogEntry, 0)

	for v := int64(1); v <= latestVersion; v++ {
		versionFilePath := fmt.Sprintf("%s/sys/_delta_log/%020d.json", dl.basePath, v)
		data, err := dl.objectStore.Get(versionFilePath)
		if err != nil {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		entries = append(entries, entry)
	}

	return entries
}

// GetEntriesByTable 获取指定表的日志条目
func (dl *OptimisticDeltaLog) GetEntriesByTable(tableID string) []LogEntry {
	allEntries := dl.GetAllEntries()
	entries := make([]LogEntry, 0)

	for _, entry := range allEntries {
		if entry.TableID == tableID {
			entries = append(entries, entry)
		}
	}

	return entries
}

// 辅助方法

// getVersionFilePath 获取版本文件路径
func (dl *OptimisticDeltaLog) getVersionFilePath(tableID string, version int64) string {
	// 格式: {basePath}/sys/_delta_log/{table_id}/000001.json
	// 为简化实现，先使用全局版本号
	return fmt.Sprintf("%s/sys/_delta_log/%020d.json", dl.basePath, version)
}

// isConflictError 判断是否是冲突错误
func isConflictError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return containsAnySubstring(errMsg, []string{
		"PreconditionFailed",
		"already exists",
		"file already exists",
	})
}

func containsAnySubstring(s string, substrs []string) bool {
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

package delta

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/ipc"
	"github.com/yyun543/minidb/internal/logger"
	"go.uber.org/zap"
)

// DeltaLog Delta Log 管理器 (SQL 自举版)
type DeltaLog struct {
	// 使用内存存储模拟 SQL 表 (实际应该用数据库)
	entries    []LogEntry
	mu         sync.RWMutex
	currentVer atomic.Int64
	tableName  string
}

// NewDeltaLog 创建 Delta Log 管理器
func NewDeltaLog() *DeltaLog {
	dl := &DeltaLog{
		entries:   make([]LogEntry, 0),
		tableName: "_system.delta_log",
	}
	dl.currentVer.Store(0)
	return dl
}

// Bootstrap 初始化系统表
func (dl *DeltaLog) Bootstrap() error {
	logger.Info("Bootstrapping Delta Log")
	// 在实际实现中，这里会创建 _system.delta_log 等系统表
	// 现在使用内存存储模拟
	return nil
}

// AppendAdd 追加 ADD 操作
func (dl *DeltaLog) AppendAdd(tableID string, file *ParquetFile) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

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
	}

	if file.Stats != nil {
		entry.MinValues = file.Stats.MinValues
		entry.MaxValues = file.Stats.MaxValues
		entry.NullCounts = file.Stats.NullCounts
	}

	dl.entries = append(dl.entries, entry)

	logger.Info("Delta Log entry appended",
		zap.Int64("version", version),
		zap.String("table", tableID),
		zap.String("operation", string(OpAdd)),
		zap.String("file", file.Path))

	// 检查是否需要创建 Checkpoint
	if version%10 == 0 {
		go dl.createCheckpoint(tableID, version)
	}

	return nil
}

// AppendRemove 追加 REMOVE 操作
func (dl *DeltaLog) AppendRemove(tableID, filePath string) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

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

	dl.entries = append(dl.entries, entry)

	logger.Info("Delta Log entry appended",
		zap.Int64("version", version),
		zap.String("table", tableID),
		zap.String("operation", string(OpRemove)),
		zap.String("file", filePath))

	return nil
}

// AppendMetadata 追加 METADATA 操作
func (dl *DeltaLog) AppendMetadata(tableID string, schema *arrow.Schema) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	version := dl.currentVer.Add(1)
	timestamp := time.Now().UnixMilli()

	schemaJSON, err := schemaToJSON(schema)
	if err != nil {
		return fmt.Errorf("failed to serialize schema: %w", err)
	}

	entry := LogEntry{
		Version:    version,
		Timestamp:  timestamp,
		TableID:    tableID,
		Operation:  OpMetadata,
		SchemaJSON: schemaJSON,
	}

	dl.entries = append(dl.entries, entry)

	logger.Info("Delta Log entry appended",
		zap.Int64("version", version),
		zap.String("table", tableID),
		zap.String("operation", string(OpMetadata)))

	return nil
}

// GetSnapshot 获取表快照
func (dl *DeltaLog) GetSnapshot(tableID string, version int64) (*Snapshot, error) {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	// 如果版本为 -1，使用最新版本
	if version == -1 {
		version = dl.currentVer.Load()
	}

	snapshot := &Snapshot{
		Version:   version,
		Timestamp: time.Now(),
		TableID:   tableID,
		Files:     make([]FileInfo, 0),
	}

	// 构建快照：所有 ADD 文件 - REMOVE 文件
	addedFiles := make(map[string]FileInfo)
	removedFiles := make(map[string]bool)

	for _, entry := range dl.entries {
		if entry.TableID != tableID || entry.Version > version {
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
			}

		case OpRemove:
			removedFiles[entry.FilePath] = true

		case OpMetadata:
			if entry.SchemaJSON != "" {
				schema, err := schemaFromJSON(entry.SchemaJSON)
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

	logger.Info("Snapshot retrieved",
		zap.String("table", tableID),
		zap.Int64("version", version),
		zap.Int("file_count", len(snapshot.Files)))

	return snapshot, nil
}

// GetLatestVersion 获取最新版本号
func (dl *DeltaLog) GetLatestVersion() int64 {
	return dl.currentVer.Load()
}

// GetVersionByTimestamp 根据时间戳查找版本号
func (dl *DeltaLog) GetVersionByTimestamp(tableID string, ts int64) (int64, error) {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	var maxVersion int64 = 0

	for _, entry := range dl.entries {
		if entry.TableID == tableID && entry.Timestamp <= ts && entry.Version > maxVersion {
			maxVersion = entry.Version
		}
	}

	if maxVersion == 0 {
		return 0, fmt.Errorf("no version found before timestamp %d", ts)
	}

	return maxVersion, nil
}

// createCheckpoint 创建检查点
func (dl *DeltaLog) createCheckpoint(tableID string, version int64) {
	logger.Info("Creating checkpoint",
		zap.String("table", tableID),
		zap.Int64("version", version))

	// 在实际实现中，这里会将快照序列化到 Parquet 文件
	// 并插入到 _system.checkpoints 表
	snapshot, err := dl.GetSnapshot(tableID, version)
	if err != nil {
		logger.Error("Failed to create checkpoint",
			zap.String("table", tableID),
			zap.Int64("version", version),
			zap.Error(err))
		return
	}

	logger.Info("Checkpoint created",
		zap.String("table", tableID),
		zap.Int64("version", version),
		zap.Int("file_count", len(snapshot.Files)))
}

// ListTables 列出所有表
func (dl *DeltaLog) ListTables() []string {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	tableSet := make(map[string]bool)
	for _, entry := range dl.entries {
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

// GetAllEntries 获取所有日志条目 (用于系统表查询)
func (dl *DeltaLog) GetAllEntries() []LogEntry {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	// 返回副本以避免并发修改
	entriesCopy := make([]LogEntry, len(dl.entries))
	copy(entriesCopy, dl.entries)

	return entriesCopy
}

// GetEntriesByTable 获取指定表的日志条目
func (dl *DeltaLog) GetEntriesByTable(tableID string) []LogEntry {
	dl.mu.RLock()
	defer dl.mu.RUnlock()

	entries := make([]LogEntry, 0)
	for _, entry := range dl.entries {
		if entry.TableID == tableID {
			entries = append(entries, entry)
		}
	}

	return entries
}

// Helper functions

// schemaToJSON 使用 Arrow IPC 序列化 Schema
// 将 Arrow Schema 序列化为 Base64 编码的 IPC 格式
func schemaToJSON(schema *arrow.Schema) (string, error) {
	if schema == nil {
		return "", fmt.Errorf("schema is nil")
	}

	// 使用 Arrow IPC Writer 序列化 Schema
	var buf bytes.Buffer
	writer := ipc.NewWriter(&buf, ipc.WithSchema(schema))

	// Close writer to finalize the IPC stream (写入 Schema)
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to write schema with IPC: %w", err)
	}

	// 将二进制数据编码为 Base64 字符串（便于存储在文本字段中）
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	logger.Debug("Schema serialized with Arrow IPC",
		zap.Int("ipc_bytes", buf.Len()),
		zap.Int("base64_bytes", len(encoded)))

	return encoded, nil
}

// schemaFromJSON 从 Base64 编码的 IPC 格式反序列化 Arrow Schema
// 使用 Arrow IPC Reader 完整还原 Schema
func schemaFromJSON(encoded string) (*arrow.Schema, error) {
	if encoded == "" {
		return nil, fmt.Errorf("encoded string is empty")
	}

	// 从 Base64 解码
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// 使用 Arrow IPC Reader 读取 Schema
	reader := bytes.NewReader(data)
	ipcReader, err := ipc.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create IPC reader: %w", err)
	}
	defer ipcReader.Release()

	// 获取 Schema
	schema := ipcReader.Schema()
	if schema == nil {
		return nil, fmt.Errorf("failed to read schema from IPC stream")
	}

	logger.Debug("Schema deserialized from Arrow IPC",
		zap.Int("field_count", len(schema.Fields())),
		zap.Int("ipc_bytes", len(data)))

	return schema, nil
}

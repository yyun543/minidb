package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/google/uuid"
	"github.com/yyun543/minidb/internal/delta"
	"github.com/yyun543/minidb/internal/logger"
	"github.com/yyun543/minidb/internal/objectstore"
	"github.com/yyun543/minidb/internal/parquet"
	"go.uber.org/zap"
)

// ParquetEngine Parquet 存储引擎实现
type ParquetEngine struct {
	basePath          string
	objectStore       ObjectStore
	deltaLog          delta.LogInterface
	schemas           map[string]*arrow.Schema // 表 schema 缓存
	mu                sync.RWMutex
	useOptimisticLock bool // 是否使用乐观并发控制
	maxRetries        int  // 冲突重试次数
}

// EngineOption 引擎配置选项
type EngineOption func(*ParquetEngine)

// WithOptimisticLock 启用乐观并发控制
func WithOptimisticLock(enabled bool) EngineOption {
	return func(pe *ParquetEngine) {
		pe.useOptimisticLock = enabled
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) EngineOption {
	return func(pe *ParquetEngine) {
		pe.maxRetries = maxRetries
	}
}

// NewParquetEngine 创建 Parquet 存储引擎
func NewParquetEngine(basePath string, opts ...EngineOption) (*ParquetEngine, error) {
	// 创建本地对象存储
	objStore, err := objectstore.NewLocalStore(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create object store: %w", err)
	}

	engine := &ParquetEngine{
		basePath:          basePath,
		objectStore:       objStore,
		schemas:           make(map[string]*arrow.Schema),
		useOptimisticLock: false, // 默认使用悲观锁（向后兼容）
		maxRetries:        5,     // 默认重试5次
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(engine)
	}

	// 根据配置选择Delta Log实现
	if engine.useOptimisticLock {
		// 使用乐观并发控制的Delta Log
		// objStore实际是*objectstore.LocalStore，它实现了ConditionalObjectStore接口
		engine.deltaLog = delta.NewOptimisticDeltaLog(objStore, basePath)
	} else {
		// 使用传统的悲观锁Delta Log
		engine.deltaLog = delta.NewDeltaLog()
	}

	return engine, nil
}

// Open 打开存储引擎并从磁盘恢复所有库表
func (pe *ParquetEngine) Open() error {
	logger.Info("Opening Parquet engine", zap.String("path", pe.basePath))

	// 1. 创建系统数据库和表
	if err := pe.createSystemTables(); err != nil {
		logger.Warn("Failed to create system tables", zap.Error(err))
	}

	// 2. 从 sys.delta_log 表恢复 Delta Log 状态到内存
	if err := pe.recoverDeltaLogFromDisk(); err != nil {
		logger.Warn("Failed to recover Delta Log, starting fresh", zap.Error(err))
	}

	// 3. 设置持久化回调（将新 entries 写入 sys.delta_log 表）
	if inMemoryLog, ok := pe.deltaLog.(*delta.DeltaLog); ok {
		inMemoryLog.SetPersistenceCallback(pe.persistDeltaLogEntry)
		// 设置checkpoint回调（将snapshot序列化到Parquet文件）
		inMemoryLog.SetCheckpointCallback(func(tableID string, version int64) error {
			return pe.CreateCheckpoint(tableID, version)
		})
	}

	// 4. 从 Delta Log 恢复表的 schema
	if err := pe.rebuildSchemasFromDeltaLog(); err != nil {
		logger.Warn("Failed to rebuild schemas", zap.Error(err))
	}

	logger.Info("Parquet engine opened successfully")
	return nil
}

// createSystemTables 创建系统数据库和表
func (pe *ParquetEngine) createSystemTables() error {
	// 创建 sys 数据库
	sysDBPath := filepath.Join(pe.basePath, "sys", ".db")
	if exists, _ := pe.objectStore.Exists(sysDBPath); !exists {
		if err := pe.objectStore.Put(sysDBPath, []byte{}); err != nil {
			return fmt.Errorf("failed to create sys database: %w", err)
		}
		logger.Info("Created sys database")
	}

	// 创建 sys.delta_log 表的目录标记
	deltaLogMarker := filepath.Join(pe.basePath, "sys", "delta_log", ".table")
	if exists, _ := pe.objectStore.Exists(deltaLogMarker); !exists {
		if err := pe.objectStore.Put(deltaLogMarker, []byte{}); err != nil {
			return fmt.Errorf("failed to create delta_log table marker: %w", err)
		}
		logger.Info("Created sys.delta_log table marker")
	}

	return nil
}

// persistDeltaLogEntry 持久化单个 Delta Log entry 到 sys.delta_log 表
func (pe *ParquetEngine) persistDeltaLogEntry(entry *delta.LogEntry) error {
	// 将 LogEntry 转换为 Arrow Record
	schema := createDeltaLogSchema()
	builder := array.NewRecordBuilder(memory.DefaultAllocator, schema)
	defer builder.Release()

	// 填充字段
	builder.Field(0).(*array.Int64Builder).Append(entry.Version)
	builder.Field(1).(*array.Int64Builder).Append(entry.Timestamp)
	builder.Field(2).(*array.StringBuilder).Append(entry.TableID)
	builder.Field(3).(*array.StringBuilder).Append(string(entry.Operation))

	// 根据操作类型填充不同字段
	switch entry.Operation {
	case delta.OpAdd:
		builder.Field(4).(*array.StringBuilder).Append(entry.FilePath)
		builder.Field(5).(*array.Int64Builder).Append(entry.FileSize)
		builder.Field(6).(*array.Int64Builder).Append(entry.RowCount)
		// TODO: Serialize MinValues, MaxValues, NullCounts as JSON
		builder.Field(7).(*array.StringBuilder).Append("") // min_values
		builder.Field(8).(*array.StringBuilder).Append("") // max_values
		builder.Field(9).(*array.StringBuilder).Append("") // null_counts
		builder.Field(10).(*array.BooleanBuilder).Append(entry.DataChange)
		builder.Field(11).AppendNull()                                   // deletion_timestamp
		builder.Field(12).AppendNull()                                   // schema_json
		builder.Field(13).AppendNull()                                   // index_json
		builder.Field(14).AppendNull()                                   // index_operation
		builder.Field(15).(*array.BooleanBuilder).Append(entry.IsDelta)  // is_delta
		builder.Field(16).(*array.StringBuilder).Append(entry.DeltaType) // delta_type

	case delta.OpRemove:
		builder.Field(4).(*array.StringBuilder).Append(entry.FilePath)
		for i := 5; i < 11; i++ {
			builder.Field(i).AppendNull()
		}
		builder.Field(11).(*array.Int64Builder).Append(entry.DeletionTimestamp)
		builder.Field(12).AppendNull() // schema_json
		builder.Field(13).AppendNull() // index_json
		builder.Field(14).AppendNull() // index_operation
		builder.Field(15).AppendNull() // is_delta
		builder.Field(16).AppendNull() // delta_type

	case delta.OpMetadata:
		for i := 4; i < 12; i++ {
			builder.Field(i).AppendNull()
		}
		// Log what we're about to persist
		logger.Info("Persisting METADATA entry to sys.delta_log",
			zap.String("table", entry.TableID),
			zap.Int64("version", entry.Version),
			zap.Int("schema_json_length", len(entry.SchemaJSON)),
			zap.Bool("has_schema", entry.SchemaJSON != ""),
			zap.Int("index_json_length", len(entry.IndexJSON)),
			zap.Bool("has_index", entry.IndexJSON != ""),
			zap.String("index_operation", entry.IndexOperation))
		builder.Field(12).(*array.StringBuilder).Append(entry.SchemaJSON)
		builder.Field(13).(*array.StringBuilder).Append(entry.IndexJSON)
		builder.Field(14).(*array.StringBuilder).Append(entry.IndexOperation)
		builder.Field(15).AppendNull() // is_delta
		builder.Field(16).AppendNull() // delta_type
	}

	record := builder.NewRecord()
	defer record.Release()

	// 调用 Write 方法，但注意不要递归
	// sys.delta_log 表写入时会被 Write() 跳过 Delta Log 跟踪
	return pe.Write(context.Background(), "sys", "delta_log", record)
}

// createDeltaLogSchema 创建 Delta Log 表的 Schema
func createDeltaLogSchema() *arrow.Schema {
	return arrow.NewSchema([]arrow.Field{
		{Name: "version", Type: arrow.PrimitiveTypes.Int64},
		{Name: "timestamp", Type: arrow.PrimitiveTypes.Int64},
		{Name: "table_id", Type: arrow.BinaryTypes.String},
		{Name: "operation", Type: arrow.BinaryTypes.String},
		{Name: "file_path", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "file_size", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
		{Name: "row_count", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
		{Name: "min_values", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "max_values", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "null_counts", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "data_change", Type: arrow.FixedWidthTypes.Boolean, Nullable: true},
		{Name: "deletion_timestamp", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
		{Name: "schema_json", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "index_json", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "index_operation", Type: arrow.BinaryTypes.String, Nullable: true},
		{Name: "is_delta", Type: arrow.FixedWidthTypes.Boolean, Nullable: true},
		{Name: "delta_type", Type: arrow.BinaryTypes.String, Nullable: true},
	}, nil)
}

// recoverDeltaLogFromDisk 从 sys.delta_log 表恢复 Delta Log 状态
// 直接扫描 Parquet 文件，不使用 Delta Log API (因为 sys.delta_log 不跟踪自己)
func (pe *ParquetEngine) recoverDeltaLogFromDisk() error {
	logger.Info("Recovering Delta Log from sys.delta_log Parquet files")

	// 直接扫描 sys/delta_log/data 目录中的 Parquet 文件
	deltaLogDir := filepath.Join(pe.basePath, "sys", "delta_log", "data")

	// 扫描目录中的所有 Parquet 文件 (filepath.Glob 会自动处理目录不存在的情况)
	files, err := pe.scanParquetFiles(deltaLogDir)
	if err != nil {
		logger.Info("Failed to scan Delta Log directory", zap.Error(err))
		return nil
	}

	if len(files) == 0 {
		logger.Info("No Parquet files found in Delta Log directory, starting fresh")
		return nil
	}

	allEntries := make([]delta.LogEntry, 0)

	// 读取所有 Parquet 文件
	for _, filePath := range files {
		entries, err := pe.readDeltaLogEntriesFromFile(filePath)
		if err != nil {
			logger.Warn("Failed to read Delta Log file",
				zap.String("file", filePath),
				zap.Error(err))
			continue
		}
		allEntries = append(allEntries, entries...)
	}

	// 使用 RestoreFromEntries 恢复状态
	if inMemoryLog, ok := pe.deltaLog.(*delta.DeltaLog); ok {
		if err := inMemoryLog.RestoreFromEntries(allEntries); err != nil {
			return fmt.Errorf("failed to restore entries: %w", err)
		}
	}

	logger.Info("Delta Log recovered from Parquet files",
		zap.Int("file_count", len(files)),
		zap.Int("entry_count", len(allEntries)))

	return nil
}

// loadDeltaLogFromDisk 从磁盘加载 Delta Log 表数据
func (pe *ParquetEngine) loadDeltaLogFromDisk(deltaLogDir string) error {
	// 检查目录是否存在
	if exists, err := pe.objectStore.Exists(deltaLogDir); err != nil || !exists {
		return fmt.Errorf("delta log directory does not exist: %s", deltaLogDir)
	}

	// 扫描目录中的所有 Parquet 文件
	files, err := pe.scanParquetFiles(deltaLogDir)
	if err != nil {
		return fmt.Errorf("failed to scan delta log files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no parquet files found in delta log directory")
	}

	// 读取所有 Parquet 文件并恢复 Delta Log entries
	allEntries := make([]delta.LogEntry, 0)

	for _, filePath := range files {
		entries, err := pe.readDeltaLogEntriesFromFile(filePath)
		if err != nil {
			logger.Warn("Failed to read delta log file",
				zap.String("file", filePath),
				zap.Error(err))
			continue
		}

		allEntries = append(allEntries, entries...)
	}

	// 恢复 Delta Log 状态
	if err := pe.deltaLog.RestoreFromEntries(allEntries); err != nil {
		return fmt.Errorf("failed to restore delta log: %w", err)
	}

	logger.Info("Delta Log restored from disk",
		zap.Int("file_count", len(files)),
		zap.Int("entry_count", len(allEntries)))

	return nil
}

// scanParquetFiles 扫描目录中的所有 Parquet 文件
func (pe *ParquetEngine) scanParquetFiles(dir string) ([]string, error) {
	logger.Debug("Scanning directory for Parquet files",
		zap.String("directory", dir))

	// 使用 map 进行去重
	fileSet := make(map[string]bool)

	// 尝试常见的文件名模式
	patterns := []string{
		filepath.Join(dir, "*.parquet"),
		filepath.Join(dir, "part-*.parquet"),
		filepath.Join(dir, "delta_log_*.parquet"),
	}

	for _, pattern := range patterns {
		logger.Debug("Trying glob pattern", zap.String("pattern", pattern))
		matches, err := filepath.Glob(pattern)
		if err != nil {
			logger.Debug("Glob pattern failed", zap.String("pattern", pattern), zap.Error(err))
			continue
		}
		logger.Debug("Glob pattern matched",
			zap.String("pattern", pattern),
			zap.Int("match_count", len(matches)),
			zap.Strings("matches", matches))

		// 去重：只添加不存在的文件
		for _, match := range matches {
			fileSet[match] = true
		}
	}

	// 转换为切片
	files := make([]string, 0, len(fileSet))
	for file := range fileSet {
		files = append(files, file)
	}

	logger.Debug("Scan completed",
		zap.String("directory", dir),
		zap.Int("total_files", len(files)),
		zap.Strings("files", files))

	return files, nil
}

// readDeltaLogEntriesFromFile 从 Parquet 文件读取 Delta Log entries
func (pe *ParquetEngine) readDeltaLogEntriesFromFile(filePath string) ([]delta.LogEntry, error) {
	// 读取 Parquet 文件
	record, err := parquet.ReadParquetFile(filePath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read parquet file: %w", err)
	}
	defer record.Release()

	entries := make([]delta.LogEntry, 0)

	// 逐行解析 Delta Log entries
	numRows := int(record.NumRows())
	for rowIdx := 0; rowIdx < numRows; rowIdx++ {
		entry := pe.parseDeltaLogEntry(record, rowIdx)
		entries = append(entries, entry)
	}

	return entries, nil
}

// parseDeltaLogEntry 从 Arrow Record 解析单个 Delta Log entry
func (pe *ParquetEngine) parseDeltaLogEntry(record arrow.Record, rowIdx int) delta.LogEntry {
	entry := delta.LogEntry{}

	// 根据列名解析字段
	schema := record.Schema()
	for colIdx := 0; colIdx < int(record.NumCols()); colIdx++ {
		field := schema.Field(colIdx)
		col := record.Column(colIdx)

		if col.IsNull(rowIdx) {
			continue
		}

		switch field.Name {
		case "version":
			if arr, ok := col.(*array.Int64); ok {
				entry.Version = arr.Value(rowIdx)
			}
		case "timestamp":
			if arr, ok := col.(*array.Int64); ok {
				entry.Timestamp = arr.Value(rowIdx)
			}
		case "table_id":
			if arr, ok := col.(*array.String); ok {
				entry.TableID = arr.Value(rowIdx)
			}
		case "operation":
			if arr, ok := col.(*array.String); ok {
				entry.Operation = delta.Operation(arr.Value(rowIdx))
			}
		case "file_path":
			if arr, ok := col.(*array.String); ok {
				entry.FilePath = arr.Value(rowIdx)
			}
		case "file_size":
			if arr, ok := col.(*array.Int64); ok {
				entry.FileSize = arr.Value(rowIdx)
			}
		case "row_count":
			if arr, ok := col.(*array.Int64); ok {
				entry.RowCount = arr.Value(rowIdx)
			}
		case "data_change":
			if arr, ok := col.(*array.Boolean); ok {
				entry.DataChange = arr.Value(rowIdx)
			}
		case "deletion_timestamp":
			if arr, ok := col.(*array.Int64); ok {
				entry.DeletionTimestamp = arr.Value(rowIdx)
			}
		case "schema_json":
			if arr, ok := col.(*array.String); ok {
				entry.SchemaJSON = arr.Value(rowIdx)
				// Log when we find a SchemaJSON value
				if entry.SchemaJSON != "" {
					logger.Debug("Parsed SchemaJSON from Delta Log entry",
						zap.String("table_id", entry.TableID),
						zap.Int("schema_json_length", len(entry.SchemaJSON)),
						zap.Int64("version", entry.Version))
				}
			}
		case "index_json":
			if arr, ok := col.(*array.String); ok {
				entry.IndexJSON = arr.Value(rowIdx)
				// Log when we find an IndexJSON value
				if entry.IndexJSON != "" {
					logger.Debug("Parsed IndexJSON from Delta Log entry",
						zap.String("table_id", entry.TableID),
						zap.Int("index_json_length", len(entry.IndexJSON)),
						zap.Int64("version", entry.Version))
				}
			}
		case "index_operation":
			if arr, ok := col.(*array.String); ok {
				entry.IndexOperation = arr.Value(rowIdx)
			}
		case "is_delta":
			if arr, ok := col.(*array.Boolean); ok {
				entry.IsDelta = arr.Value(rowIdx)
			}
		case "delta_type":
			if arr, ok := col.(*array.String); ok {
				entry.DeltaType = arr.Value(rowIdx)
			}
		}
	}

	// Log parsed entry for METADATA operations
	if entry.Operation == delta.OpMetadata {
		logger.Info("Parsed METADATA entry from Parquet",
			zap.String("table_id", entry.TableID),
			zap.Int64("version", entry.Version),
			zap.Int("schema_json_length", len(entry.SchemaJSON)),
			zap.Bool("has_schema", entry.SchemaJSON != ""),
			zap.Int("index_json_length", len(entry.IndexJSON)),
			zap.Bool("has_index", entry.IndexJSON != ""),
			zap.String("index_operation", entry.IndexOperation))
	}

	return entry
}

// rebuildSchemasFromDeltaLog 从 Delta Log 重建所有表的 schema
func (pe *ParquetEngine) rebuildSchemasFromDeltaLog() error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	// 获取所有表的列表
	tables := pe.deltaLog.ListTables()

	for _, tableID := range tables {
		// 跳过系统表
		if tableID == "sys.delta_log" {
			continue
		}

		// 从 Delta Log 获取 METADATA 条目
		snapshot, err := pe.deltaLog.GetSnapshot(tableID, -1)
		if err != nil {
			logger.Warn("Failed to get snapshot for table",
				zap.String("table", tableID),
				zap.Error(err))
			continue
		}

		// 如果 snapshot 包含 schema，恢复到内存
		if snapshot.Schema != nil {
			pe.schemas[tableID] = snapshot.Schema
			logger.Info("Restored table schema",
				zap.String("table", tableID),
				zap.Int("field_count", len(snapshot.Schema.Fields())))
		} else {
			// Schema is nil - this is a problem that needs investigation
			logger.Warn("Snapshot has nil schema - cannot restore table schema",
				zap.String("table", tableID),
				zap.Int("file_count", len(snapshot.Files)))
		}
	}

	return nil
}

// Close 关闭存储引擎
func (pe *ParquetEngine) Close() error {
	logger.Info("Closing Parquet engine")
	return nil
}

// CreateDatabase 创建数据库
func (pe *ParquetEngine) CreateDatabase(name string) error {
	logger.Info("Creating database", zap.String("name", name))

	// 创建数据库目录
	dbPath := filepath.Join(pe.basePath, name)

	// 先创建目录
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Sync目录确保元数据持久化（P0改进：目录fsync）
	if dir, err := os.Open(dbPath); err == nil {
		dir.Sync()
		dir.Close()
	}

	// 创建.db标记文件
	if err := pe.objectStore.Put(filepath.Join(dbPath, ".db"), []byte{}); err != nil {
		return fmt.Errorf("failed to create database marker: %w", err)
	}

	return nil
}

// DropDatabase 删除数据库
func (pe *ParquetEngine) DropDatabase(name string) error {
	logger.Info("Dropping database", zap.String("name", name))
	// 简化实现：实际应该删除所有表和数据
	return nil
}

// ListDatabases 列出所有数据库
func (pe *ParquetEngine) ListDatabases() ([]string, error) {
	// 简化实现：返回硬编码的数据库列表
	return []string{"default"}, nil
}

// DatabaseExists 检查数据库是否存在
func (pe *ParquetEngine) DatabaseExists(name string) (bool, error) {
	dbPath := filepath.Join(pe.basePath, name, ".db")
	return pe.objectStore.Exists(dbPath)
}

// CreateTable 创建表
func (pe *ParquetEngine) CreateTable(db, table string, schema *arrow.Schema) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	tableID := fmt.Sprintf("%s.%s", db, table)
	logger.Info("Creating table", zap.String("table", tableID))

	// 保存 schema
	pe.schemas[tableID] = schema

	// 追加 METADATA 操作到 Delta Log
	if err := pe.deltaLog.AppendMetadata(tableID, schema); err != nil {
		return fmt.Errorf("failed to append metadata: %w", err)
	}

	logger.Info("Table created", zap.String("table", tableID))
	return nil
}

// DropTable 删除表
func (pe *ParquetEngine) DropTable(db, table string) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	tableID := fmt.Sprintf("%s.%s", db, table)
	logger.Info("Dropping table", zap.String("table", tableID))

	// 删除 schema 缓存
	delete(pe.schemas, tableID)

	// 获取所有文件并标记为删除
	snapshot, err := pe.deltaLog.GetSnapshot(tableID, -1)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}

	// Mark all files as REMOVE
	for _, file := range snapshot.Files {
		if err := pe.deltaLog.AppendRemove(tableID, file.Path); err != nil {
			logger.Warn("Failed to remove file", zap.String("file", file.Path), zap.Error(err))
		}
	}

	// IMPORTANT: If table has no files (empty table), we still need to mark it as deleted
	// Write a special REMOVE entry to mark the table as dropped
	if len(snapshot.Files) == 0 {
		// Use a special marker path to indicate table deletion
		if err := pe.deltaLog.AppendRemove(tableID, fmt.Sprintf("_table_dropped_marker_%s", tableID)); err != nil {
			logger.Warn("Failed to write table deletion marker", zap.String("table", tableID), zap.Error(err))
		}
		logger.Info("Empty table deletion marker written", zap.String("table", tableID))
	}

	logger.Info("Table dropped", zap.String("table", tableID))
	return nil
}

// GetTableSchema 获取表 schema
func (pe *ParquetEngine) GetTableSchema(db, table string) (*arrow.Schema, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	tableID := fmt.Sprintf("%s.%s", db, table)
	schema, ok := pe.schemas[tableID]
	if !ok {
		return nil, fmt.Errorf("table not found: %s", tableID)
	}

	return schema, nil
}

// ListTables 列出数据库中的所有表
func (pe *ParquetEngine) ListTables(db string) ([]string, error) {
	tables := pe.deltaLog.ListTables()

	// 过滤出指定数据库的表
	result := make([]string, 0)
	prefix := db + "."
	for _, table := range tables {
		if len(table) > len(prefix) && table[:len(prefix)] == prefix {
			result = append(result, table[len(prefix):])
		}
	}

	return result, nil
}

// TableExists 检查表是否存在
func (pe *ParquetEngine) TableExists(db, table string) (bool, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	tableID := fmt.Sprintf("%s.%s", db, table)
	_, ok := pe.schemas[tableID]
	return ok, nil
}

// Scan 扫描表数据
func (pe *ParquetEngine) Scan(ctx context.Context, db, table string, filters []Filter) (RecordIterator, error) {
	tableID := fmt.Sprintf("%s.%s", db, table)
	logger.Info("Scanning table", zap.String("table", tableID))

	// 获取最新快照
	snapshot, err := pe.deltaLog.GetSnapshot(tableID, -1)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	// 文件级过滤 (Zone Maps)
	selectedFiles := pe.filterFilesByStats(snapshot.Files, filters)

	// Separate base files and delta files
	baseFiles := make([]delta.FileInfo, 0)
	deltaFiles := make([]delta.FileInfo, 0)
	for _, file := range selectedFiles {
		// Debug: log each file's IsDelta status
		logger.Debug("File classification",
			zap.String("path", file.Path),
			zap.Bool("is_delta", file.IsDelta),
			zap.String("delta_type", file.DeltaType))

		if file.IsDelta {
			deltaFiles = append(deltaFiles, file)
		} else {
			baseFiles = append(baseFiles, file)
		}
	}

	logger.Info("Files selected for scan",
		zap.Int("total", len(snapshot.Files)),
		zap.Int("selected", len(selectedFiles)),
		zap.Int("base_files", len(baseFiles)),
		zap.Int("delta_files", len(deltaFiles)))

	// Use Merge-on-Read iterator if there are delta files
	if len(deltaFiles) > 0 {
		return NewMergeOnReadIterator(baseFiles, deltaFiles, filters)
	}

	// Standard iterator for base files only
	return NewParquetIterator(baseFiles, filters)
}

// Write 写入数据
func (pe *ParquetEngine) Write(ctx context.Context, db, table string, batch arrow.Record) error {
	tableID := fmt.Sprintf("%s.%s", db, table)
	logger.Info("Writing to table",
		zap.String("table", tableID),
		zap.Int64("rows", batch.NumRows()))

	// 生成 Parquet 文件路径
	filePath := pe.generateFilePath(db, table)

	// 写入 Parquet 文件
	stats, err := parquet.WriteArrowBatch(filePath, batch)
	if err != nil {
		return fmt.Errorf("failed to write parquet: %w", err)
	}

	// 追加到 Delta Log
	// 特殊处理：sys.delta_log 表不跟踪自己，避免无限递归
	if tableID != "sys.delta_log" {
		parquetFile := &delta.ParquetFile{
			Path:     filePath,
			Size:     stats.FileSize,
			RowCount: stats.RowCount,
			Stats:    stats,
		}

		if err := pe.deltaLog.AppendAdd(tableID, parquetFile); err != nil {
			return fmt.Errorf("failed to append to delta log: %w", err)
		}
	}

	logger.Info("Write completed",
		zap.String("table", tableID),
		zap.String("file", filePath),
		zap.Int64("rows", stats.RowCount))

	return nil
}

// matchesFilters checks if a row matches all filters
func (pe *ParquetEngine) matchesFilters(record arrow.Record, rowIdx int, filters []Filter) bool {
	if len(filters) == 0 {
		return true // No filters means all rows match
	}

	for _, filter := range filters {
		// Find column index
		colIdx := -1
		for i := 0; i < int(record.NumCols()); i++ {
			if record.Schema().Field(i).Name == filter.Column {
				colIdx = i
				break
			}
		}

		if colIdx == -1 {
			return false // Column not found
		}

		col := record.Column(colIdx)
		value := pe.getValueFromColumn(col, rowIdx)

		if !pe.compareValue(value, filter.Operator, filter.Value) {
			return false
		}
	}

	return true
}

// getValueFromColumn extracts value from a column at given row index
func (pe *ParquetEngine) getValueFromColumn(col arrow.Array, rowIdx int) interface{} {
	if col.IsNull(rowIdx) {
		return nil
	}

	switch arr := col.(type) {
	case *array.Int64:
		return arr.Value(rowIdx)
	case *array.Int32:
		return int64(arr.Value(rowIdx))
	case *array.Int16:
		return int64(arr.Value(rowIdx))
	case *array.Int8:
		return int64(arr.Value(rowIdx))
	case *array.Float64:
		return arr.Value(rowIdx)
	case *array.Float32:
		return float64(arr.Value(rowIdx))
	case *array.String:
		return arr.Value(rowIdx)
	case *array.Boolean:
		return arr.Value(rowIdx)
	default:
		return nil
	}
}

// appendValueToBuilder appends a value to the appropriate builder
func (pe *ParquetEngine) appendValueToBuilder(builder array.Builder, value interface{}, dataType arrow.DataType) {
	if value == nil {
		builder.AppendNull()
		return
	}

	switch b := builder.(type) {
	case *array.Int64Builder:
		if v, ok := value.(int64); ok {
			b.Append(v)
		} else if v, ok := value.(int); ok {
			b.Append(int64(v))
		} else {
			b.AppendNull()
		}
	case *array.Int32Builder:
		if v, ok := value.(int32); ok {
			b.Append(v)
		} else if v, ok := value.(int64); ok {
			b.Append(int32(v))
		} else if v, ok := value.(int); ok {
			b.Append(int32(v))
		} else {
			b.AppendNull()
		}
	case *array.Float64Builder:
		if v, ok := value.(float64); ok {
			b.Append(v)
		} else if v, ok := value.(float32); ok {
			b.Append(float64(v))
		} else {
			b.AppendNull()
		}
	case *array.StringBuilder:
		if v, ok := value.(string); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	case *array.BooleanBuilder:
		if v, ok := value.(bool); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	default:
		builder.AppendNull()
	}
}

// compareValue compares a value with a filter value using the given operator
func (pe *ParquetEngine) compareValue(colValue interface{}, operator string, filterValue interface{}) bool {
	if colValue == nil || filterValue == nil {
		return false
	}

	switch operator {
	case "=", "==":
		return fmt.Sprintf("%v", colValue) == fmt.Sprintf("%v", filterValue)
	case "!=", "<>":
		return fmt.Sprintf("%v", colValue) != fmt.Sprintf("%v", filterValue)
	case ">":
		return pe.compareNumeric(colValue, filterValue) > 0
	case ">=":
		return pe.compareNumeric(colValue, filterValue) >= 0
	case "<":
		return pe.compareNumeric(colValue, filterValue) < 0
	case "<=":
		return pe.compareNumeric(colValue, filterValue) <= 0
	default:
		return false
	}
}

// compareNumeric compares two numeric values
func (pe *ParquetEngine) compareNumeric(a, b interface{}) int {
	aVal := pe.toFloat64(a)
	bVal := pe.toFloat64(b)

	if aVal < bVal {
		return -1
	} else if aVal > bVal {
		return 1
	}
	return 0
}

// toFloat64 converts various numeric types to float64
func (pe *ParquetEngine) toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		return 0
	}
}

// Update 更新数据 (Copy-on-Write)
// Update 更新数据 (使用 Merge-on-Read)
func (pe *ParquetEngine) Update(ctx context.Context, db, table string, filters []Filter, updates map[string]interface{}) (int64, error) {
	// 直接使用 Merge-on-Read 实现
	return pe.UpdateMergeOnRead(ctx, db, table, filters, updates)
}

// Delete 删除数据 (使用 Merge-on-Read)
func (pe *ParquetEngine) Delete(ctx context.Context, db, table string, filters []Filter) (int64, error) {
	// 直接使用 Merge-on-Read 实现
	return pe.DeleteMergeOnRead(ctx, db, table, filters)
}

// BeginTransaction 开始事务
func (pe *ParquetEngine) BeginTransaction() (Transaction, error) {
	return &ParquetTransaction{
		id:      uuid.New().String(),
		version: pe.deltaLog.GetLatestVersion(),
	}, nil
}

// GetTableStats 获取表统计信息
func (pe *ParquetEngine) GetTableStats(db, table string) (*TableStats, error) {
	tableID := fmt.Sprintf("%s.%s", db, table)

	snapshot, err := pe.deltaLog.GetSnapshot(tableID, -1)
	if err != nil {
		return nil, err
	}

	stats := &TableStats{
		TableName:    table,
		FileCount:    len(snapshot.Files),
		LastModified: snapshot.Timestamp.Unix(),
	}

	// 计算总行数和大小
	for _, file := range snapshot.Files {
		stats.RowCount += file.RowCount
		stats.TotalSizeGB += float64(file.Size) / (1024 * 1024 * 1024)
	}

	return stats, nil
}

// GetDeltaLog 获取 Delta Log 实例 (用于系统表查询和compaction)
func (pe *ParquetEngine) GetDeltaLog() delta.LogInterface {
	return pe.deltaLog
}

// ScanVersion 时间旅行查询
func (pe *ParquetEngine) ScanVersion(ctx context.Context, db, table string, version int64, filters []Filter) (RecordIterator, error) {
	tableID := fmt.Sprintf("%s.%s", db, table)
	logger.Info("Scanning table at version",
		zap.String("table", tableID),
		zap.Int64("version", version))

	// 获取指定版本的快照
	snapshot, err := pe.deltaLog.GetSnapshot(tableID, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	// 创建迭代器
	return NewParquetIterator(snapshot.Files, filters)
}

// Helper methods

func (pe *ParquetEngine) generateFilePath(db, table string) string {
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s_%s.parquet", table, timestamp, uuid.New().String()[:8])
	return filepath.Join(pe.basePath, db, table, "data", filename)
}

// filterFilesByStats 使用文件级统计信息进行数据跳过 (Zone Maps)
// 支持所有比较操作符: =, >, <, >=, <=, IN
func (pe *ParquetEngine) filterFilesByStats(files []delta.FileInfo, filters []Filter) []delta.FileInfo {
	if len(filters) == 0 {
		return files
	}

	selected := make([]delta.FileInfo, 0, len(files))

	for _, file := range files {
		skip := false

		for _, filter := range filters {
			min, hasMin := file.MinValues[filter.Column]
			max, hasMax := file.MaxValues[filter.Column]

			// 如果没有统计信息，不能跳过该文件
			if !hasMin || !hasMax {
				continue
			}

			// 根据操作符类型进行文件级过滤
			switch filter.Operator {
			case "=":
				// 等值过滤: 值必须在 [min, max] 范围内
				if !valueInRange(filter.Value, min, max) {
					skip = true
				}

			case ">":
				// 大于过滤: 如果 filterValue >= max, 跳过该文件
				if compareValues(filter.Value, max) >= 0 {
					skip = true
				}

			case "<":
				// 小于过滤: 如果 filterValue <= min, 跳过该文件
				if compareValues(filter.Value, min) <= 0 {
					skip = true
				}

			case ">=":
				// 大于等于过滤: 如果 filterValue > max, 跳过该文件
				if compareValues(filter.Value, max) > 0 {
					skip = true
				}

			case "<=":
				// 小于等于过滤: 如果 filterValue < min, 跳过该文件
				if compareValues(filter.Value, min) < 0 {
					skip = true
				}

			case "IN":
				// IN 过滤: 如果所有值都不在 [min, max] 范围内, 跳过该文件
				if len(filter.Values) > 0 {
					allOutOfRange := true
					for _, val := range filter.Values {
						if valueInRange(val, min, max) {
							allOutOfRange = false
							break
						}
					}
					if allOutOfRange {
						skip = true
					}
				}
			}

			if skip {
				break
			}
		}

		if !skip {
			selected = append(selected, file)
		}
	}

	return selected
}

// valueInRange 检查值是否在范围 [min, max] 内
func valueInRange(val, min, max interface{}) bool {
	// 值必须 >= min 且 <= max
	return compareValues(val, min) >= 0 && compareValues(val, max) <= 0
}

// compareValues 比较两个值
// 返回: -1 (val < other), 0 (val == other), 1 (val > other)
func compareValues(val, other interface{}) int {
	// 处理nil
	if val == nil && other == nil {
		return 0
	}
	if val == nil {
		return -1
	}
	if other == nil {
		return 1
	}

	// Int64 比较
	if v1, ok1 := toInt64(val); ok1 {
		if v2, ok2 := toInt64(other); ok2 {
			if v1 < v2 {
				return -1
			} else if v1 > v2 {
				return 1
			}
			return 0
		}
	}

	// Float64 比较
	if v1, ok1 := toFloat64(val); ok1 {
		if v2, ok2 := toFloat64(other); ok2 {
			if v1 < v2 {
				return -1
			} else if v1 > v2 {
				return 1
			}
			return 0
		}
	}

	// String 比较
	if v1, ok1 := val.(string); ok1 {
		if v2, ok2 := other.(string); ok2 {
			if v1 < v2 {
				return -1
			} else if v1 > v2 {
				return 1
			}
			return 0
		}
	}

	// 不支持的类型，默认返回0
	return 0
}

// toInt64 尝试转换为int64
func toInt64(val interface{}) (int64, bool) {
	switch v := val.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case int16:
		return int64(v), true
	case int8:
		return int64(v), true
	case uint64:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint8:
		return int64(v), true
	default:
		return 0, false
	}
}

// toFloat64 尝试转换为float64
func toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	default:
		return 0, false
	}
}

// ParquetTransaction Parquet 事务实现
type ParquetTransaction struct {
	id      string
	version int64
}

func (pt *ParquetTransaction) GetVersion() int64 {
	return pt.version
}

func (pt *ParquetTransaction) GetID() string {
	return pt.id
}

func (pt *ParquetTransaction) Commit() error {
	// 简化实现
	return nil
}

func (pt *ParquetTransaction) Rollback() error {
	// 简化实现
	return nil
}

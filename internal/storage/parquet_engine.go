package storage

import (
	"context"
	"fmt"
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
	basePath    string
	objectStore ObjectStore
	deltaLog    *delta.DeltaLog
	schemas     map[string]*arrow.Schema // 表 schema 缓存
	mu          sync.RWMutex
}

// NewParquetEngine 创建 Parquet 存储引擎
func NewParquetEngine(basePath string) (*ParquetEngine, error) {
	// 创建本地对象存储
	objStore, err := objectstore.NewLocalStore(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create object store: %w", err)
	}

	engine := &ParquetEngine{
		basePath:    basePath,
		objectStore: objStore,
		deltaLog:    delta.NewDeltaLog(),
		schemas:     make(map[string]*arrow.Schema),
	}

	return engine, nil
}

// Open 打开存储引擎
func (pe *ParquetEngine) Open() error {
	logger.Info("Opening Parquet engine", zap.String("path", pe.basePath))

	// Bootstrap Delta Log
	if err := pe.deltaLog.Bootstrap(); err != nil {
		return fmt.Errorf("failed to bootstrap delta log: %w", err)
	}

	logger.Info("Parquet engine opened successfully")
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
	if err := pe.objectStore.Put(filepath.Join(dbPath, ".db"), []byte{}); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
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

	for _, file := range snapshot.Files {
		if err := pe.deltaLog.AppendRemove(tableID, file.Path); err != nil {
			logger.Warn("Failed to remove file", zap.String("file", file.Path), zap.Error(err))
		}
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

	logger.Info("Files selected for scan",
		zap.Int("total", len(snapshot.Files)),
		zap.Int("selected", len(selectedFiles)))

	// 创建迭代器
	return NewParquetIterator(selectedFiles, filters)
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
	parquetFile := &delta.ParquetFile{
		Path:     filePath,
		Size:     stats.FileSize,
		RowCount: stats.RowCount,
		Stats:    stats,
	}

	if err := pe.deltaLog.AppendAdd(tableID, parquetFile); err != nil {
		return fmt.Errorf("failed to append to delta log: %w", err)
	}

	logger.Info("Write completed",
		zap.String("table", tableID),
		zap.String("file", filePath),
		zap.Int64("rows", stats.RowCount))

	return nil
}

// updateRecord applies updates to a record based on filters
func (pe *ParquetEngine) updateRecord(record arrow.Record, filters []Filter, updates map[string]interface{}, schema *arrow.Schema) (arrow.Record, int64) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	updatedCount := int64(0)
	numRows := int(record.NumRows())

	// Process each row
	for rowIdx := 0; rowIdx < numRows; rowIdx++ {
		// Check if row matches filters
		matches := pe.matchesFilters(record, rowIdx, filters)

		// Track if this row was updated
		rowUpdated := false

		// Copy/update row
		for colIdx := 0; colIdx < int(record.NumCols()); colIdx++ {
			field := schema.Field(colIdx)
			col := record.Column(colIdx)

			var value interface{}

			// If row matches filters and column has update, use updated value
			if matches {
				if updatedVal, hasUpdate := updates[field.Name]; hasUpdate {
					value = updatedVal
					rowUpdated = true
				} else {
					value = pe.getValueFromColumn(col, rowIdx)
				}
			} else {
				value = pe.getValueFromColumn(col, rowIdx)
			}

			// Append value to builder
			pe.appendValueToBuilder(builder.Field(colIdx), value, field.Type)
		}

		// Count this row if it was updated
		if rowUpdated {
			updatedCount++
		}
	}

	// Always return the record (with or without updates)
	return builder.NewRecord(), updatedCount
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
func (pe *ParquetEngine) Update(ctx context.Context, db, table string, filters []Filter, updates map[string]interface{}) (int64, error) {
	tableID := fmt.Sprintf("%s.%s", db, table)
	logger.Info("Updating table with Copy-on-Write", zap.String("table", tableID))

	// 1. Read all current data
	iterator, err := pe.Scan(ctx, db, table, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to scan table: %w", err)
	}
	defer iterator.Close()

	// 2. Get table schema from first record or from schemas map
	var schema *arrow.Schema
	if s, err := pe.GetTableSchema(db, table); err == nil {
		schema = s
	}

	// 3. Collect all records and apply updates (Copy-on-Write)
	updatedRecords := make([]arrow.Record, 0)
	updatedCount := int64(0)

	for iterator.Next() {
		record := iterator.Record()

		// Infer schema from first record if not already set
		if schema == nil {
			schema = record.Schema()
		}

		// Apply updates to matching rows
		updatedRecord, count := pe.updateRecord(record, filters, updates, schema)
		if updatedRecord != nil {
			updatedRecords = append(updatedRecords, updatedRecord)
			updatedCount += count
		}
	}

	if err := iterator.Err(); err != nil {
		// Clean up
		for _, rec := range updatedRecords {
			rec.Release()
		}
		return 0, fmt.Errorf("iteration error: %w", err)
	}

	// 4. Mark old files as REMOVE in Delta Log
	snapshot, err := pe.deltaLog.GetSnapshot(tableID, -1)
	if err != nil {
		for _, rec := range updatedRecords {
			rec.Release()
		}
		return 0, fmt.Errorf("failed to get snapshot: %w", err)
	}

	for _, file := range snapshot.Files {
		if err := pe.deltaLog.AppendRemove(tableID, file.Path); err != nil {
			for _, rec := range updatedRecords {
				rec.Release()
			}
			return 0, fmt.Errorf("failed to mark file for removal: %w", err)
		}
	}

	// 5. Write updated records as new Parquet files
	for i, record := range updatedRecords {
		// Generate new file name
		fileName := fmt.Sprintf("part-%d-%d.parquet", time.Now().UnixNano(), i)
		filePath := filepath.Join(pe.basePath, db, table, fileName)

		// Write to Parquet
		stats, err := parquet.WriteArrowBatch(filePath, record)
		if err != nil {
			record.Release()
			return 0, fmt.Errorf("failed to write updated parquet file: %w", err)
		}

		// Add to Delta Log
		parquetFile := &delta.ParquetFile{
			Path:     filePath,
			Size:     stats.FileSize, // 使用从 parquet writer 返回的实际文件大小
			RowCount: stats.RowCount,
			Stats: &delta.FileStats{
				RowCount:   stats.RowCount,
				FileSize:   stats.FileSize, // 使用实际文件大小
				MinValues:  stats.MinValues,
				MaxValues:  stats.MaxValues,
				NullCounts: stats.NullCounts,
			},
		}

		if err := pe.deltaLog.AppendAdd(tableID, parquetFile); err != nil {
			record.Release()
			return 0, fmt.Errorf("failed to append to delta log: %w", err)
		}

		record.Release()
	}

	logger.Info("Update completed with Copy-on-Write",
		zap.String("table", tableID),
		zap.Int64("updated_rows", updatedCount))

	return updatedCount, nil
}

// Delete 删除数据 (Delta Log integration)
func (pe *ParquetEngine) Delete(ctx context.Context, db, table string, filters []Filter) (int64, error) {
	tableID := fmt.Sprintf("%s.%s", db, table)
	logger.Info("Deleting from table with Delta Log integration", zap.String("table", tableID))

	// 1. Read all current data
	iterator, err := pe.Scan(ctx, db, table, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to scan table: %w", err)
	}
	defer iterator.Close()

	// 2. Get table schema from first record or from schemas map
	var schema *arrow.Schema
	if s, err := pe.GetTableSchema(db, table); err == nil {
		schema = s
	}

	// 3. Collect records excluding deleted rows
	filteredRecords := make([]arrow.Record, 0)
	deletedCount := int64(0)

	for iterator.Next() {
		record := iterator.Record()

		// Infer schema from first record if not already set
		if schema == nil {
			schema = record.Schema()
		}

		// Filter out rows that match deletion criteria
		filteredRecord, count := pe.filterRecord(record, filters, schema)
		if filteredRecord != nil && filteredRecord.NumRows() > 0 {
			filteredRecords = append(filteredRecords, filteredRecord)
		}
		deletedCount += count
	}

	if err := iterator.Err(); err != nil {
		// Clean up
		for _, rec := range filteredRecords {
			rec.Release()
		}
		return 0, fmt.Errorf("iteration error: %w", err)
	}

	// If no rows were deleted, return early
	if deletedCount == 0 {
		for _, rec := range filteredRecords {
			rec.Release()
		}
		logger.Info("DELETE completed - no matching rows",
			zap.String("table", tableID),
			zap.Int64("deleted_rows", 0))
		return 0, nil
	}

	// 4. Mark old files as REMOVE in Delta Log
	snapshot, err := pe.deltaLog.GetSnapshot(tableID, -1)
	if err != nil {
		for _, rec := range filteredRecords {
			rec.Release()
		}
		return 0, fmt.Errorf("failed to get snapshot: %w", err)
	}

	for _, file := range snapshot.Files {
		if err := pe.deltaLog.AppendRemove(tableID, file.Path); err != nil {
			for _, rec := range filteredRecords {
				rec.Release()
			}
			return 0, fmt.Errorf("failed to mark file for removal: %w", err)
		}
	}

	// 5. Write filtered records as new Parquet files
	for i, record := range filteredRecords {
		// Generate new file name
		fileName := fmt.Sprintf("part-%d-%d.parquet", time.Now().UnixNano(), i)
		filePath := filepath.Join(pe.basePath, db, table, fileName)

		// Write to Parquet
		stats, err := parquet.WriteArrowBatch(filePath, record)
		if err != nil {
			record.Release()
			return 0, fmt.Errorf("failed to write filtered parquet file: %w", err)
		}

		// Add to Delta Log
		parquetFile := &delta.ParquetFile{
			Path:     filePath,
			Size:     stats.FileSize, // 使用从 parquet writer 返回的实际文件大小
			RowCount: stats.RowCount,
			Stats: &delta.FileStats{
				RowCount:   stats.RowCount,
				FileSize:   stats.FileSize, // 使用实际文件大小
				MinValues:  stats.MinValues,
				MaxValues:  stats.MaxValues,
				NullCounts: stats.NullCounts,
			},
		}

		if err := pe.deltaLog.AppendAdd(tableID, parquetFile); err != nil {
			record.Release()
			return 0, fmt.Errorf("failed to append to delta log: %w", err)
		}

		record.Release()
	}

	logger.Info("DELETE completed with Delta Log integration",
		zap.String("table", tableID),
		zap.Int64("deleted_rows", deletedCount))

	return deletedCount, nil
}

// filterRecord removes rows that match the filters
func (pe *ParquetEngine) filterRecord(record arrow.Record, filters []Filter, schema *arrow.Schema) (arrow.Record, int64) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	deletedCount := int64(0)
	numRows := int(record.NumRows())

	// Process each row
	for rowIdx := 0; rowIdx < numRows; rowIdx++ {
		// Check if row matches filters (should be deleted)
		matches := pe.matchesFilters(record, rowIdx, filters)

		if matches {
			// Skip this row (delete it)
			deletedCount++
			continue
		}

		// Keep this row - copy to builder
		for colIdx := 0; colIdx < int(record.NumCols()); colIdx++ {
			col := record.Column(colIdx)
			field := schema.Field(colIdx)
			value := pe.getValueFromColumn(col, rowIdx)
			pe.appendValueToBuilder(builder.Field(colIdx), value, field.Type)
		}
	}

	// Return the filtered record
	return builder.NewRecord(), deletedCount
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

// GetDeltaLog 获取 Delta Log 实例 (用于系统表查询)
func (pe *ParquetEngine) GetDeltaLog() *delta.DeltaLog {
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

func (pe *ParquetEngine) filterFilesByStats(files []delta.FileInfo, filters []Filter) []delta.FileInfo {
	if len(filters) == 0 {
		return files
	}

	selected := make([]delta.FileInfo, 0, len(files))

	for _, file := range files {
		skip := false

		for _, filter := range filters {
			min, ok := file.MinValues[filter.Column]
			if !ok {
				continue
			}

			max, ok := file.MaxValues[filter.Column]
			if !ok {
				continue
			}

			// Zone Map 过滤
			if filter.Operator == "=" {
				if !valueInRange(filter.Value, min, max) {
					skip = true
					break
				}
			}
		}

		if !skip {
			selected = append(selected, file)
		}
	}

	return selected
}

func valueInRange(val, min, max interface{}) bool {
	// 简化实现
	return true
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

package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/google/uuid"
	"github.com/yyun543/minidb/internal/delta"
	"github.com/yyun543/minidb/internal/logger"
	"github.com/yyun543/minidb/internal/parquet"
	"go.uber.org/zap"
)

// DeltaFile represents a Merge-on-Read delta file
type DeltaFile struct {
	Type      string                 // "update", "delete", "insert"
	Filters   []Filter               // WHERE conditions
	Updates   map[string]interface{} // SET clauses for updates
	Timestamp int64
	Version   int64
}

// UpdateMergeOnRead performs UPDATE using Merge-on-Read architecture
func (pe *ParquetEngine) UpdateMergeOnRead(ctx context.Context, db, table string, filters []Filter, updates map[string]interface{}) (int64, error) {
	tableID := fmt.Sprintf("%s.%s", db, table)
	logger.Info("Updating table with Merge-on-Read",
		zap.String("table", tableID),
		zap.Int("filter_count", len(filters)))

	// Estimate affected rows for return value
	affectedRows := pe.estimateAffectedRows(ctx, tableID, filters)

	// Create delta file with update information
	deltaFile := &DeltaFile{
		Type:      "update",
		Filters:   filters,
		Updates:   updates,
		Timestamp: time.Now().UnixMilli(),
		Version:   pe.deltaLog.GetLatestVersion() + 1,
	}

	// Serialize delta file to Arrow Record
	schema, err := pe.GetTableSchema(db, table)
	if err != nil {
		return 0, fmt.Errorf("failed to get table schema: %w", err)
	}

	record := pe.serializeDeltaFile(deltaFile, schema)
	defer record.Release()

	// Write small delta file
	fileName := fmt.Sprintf("delta-update-%d-%s.parquet", time.Now().UnixNano(), uuid.New().String()[:8])
	deltaPath := filepath.Join(pe.basePath, db, table, "deltas", fileName)

	stats, err := parquet.WriteArrowBatch(deltaPath, record)
	if err != nil {
		return 0, fmt.Errorf("failed to write delta file: %w", err)
	}

	// Append to Delta Log with IsDelta flag
	parquetFile := &delta.ParquetFile{
		Path:      deltaPath,
		Size:      stats.FileSize,
		RowCount:  stats.RowCount,
		Stats:     stats,
		IsDelta:   true,
		DeltaType: "update",
	}

	if err := pe.deltaLog.AppendAdd(tableID, parquetFile); err != nil {
		return 0, fmt.Errorf("failed to append to delta log: %w", err)
	}

	logger.Info("Update completed with Merge-on-Read",
		zap.String("table", tableID),
		zap.Int64("estimated_affected_rows", affectedRows),
		zap.String("delta_file", deltaPath))

	return affectedRows, nil
}

// DeleteMergeOnRead performs DELETE using Merge-on-Read architecture
func (pe *ParquetEngine) DeleteMergeOnRead(ctx context.Context, db, table string, filters []Filter) (int64, error) {
	tableID := fmt.Sprintf("%s.%s", db, table)
	logger.Info("Deleting from table with Merge-on-Read",
		zap.String("table", tableID))

	// Estimate affected rows
	affectedRows := pe.estimateAffectedRows(ctx, tableID, filters)

	// Create delta file with delete information
	deltaFile := &DeltaFile{
		Type:      "delete",
		Filters:   filters,
		Timestamp: time.Now().UnixMilli(),
		Version:   pe.deltaLog.GetLatestVersion() + 1,
	}

	// Serialize delta file
	schema, err := pe.GetTableSchema(db, table)
	if err != nil {
		return 0, fmt.Errorf("failed to get table schema: %w", err)
	}

	record := pe.serializeDeltaFile(deltaFile, schema)
	defer record.Release()

	// Write delta file
	fileName := fmt.Sprintf("delta-delete-%d-%s.parquet", time.Now().UnixNano(), uuid.New().String()[:8])
	deltaPath := filepath.Join(pe.basePath, db, table, "deltas", fileName)

	stats, err := parquet.WriteArrowBatch(deltaPath, record)
	if err != nil {
		return 0, fmt.Errorf("failed to write delta file: %w", err)
	}

	// Append to Delta Log with IsDelta flag
	parquetFile := &delta.ParquetFile{
		Path:      deltaPath,
		Size:      stats.FileSize,
		RowCount:  stats.RowCount,
		Stats:     stats,
		IsDelta:   true,
		DeltaType: "delete",
	}

	if err := pe.deltaLog.AppendAdd(tableID, parquetFile); err != nil {
		return 0, fmt.Errorf("failed to append to delta log: %w", err)
	}

	logger.Info("Delete completed with Merge-on-Read",
		zap.String("table", tableID),
		zap.Int64("estimated_affected_rows", affectedRows),
		zap.String("delta_file", deltaPath))

	return affectedRows, nil
}

// serializeDeltaFile serializes a delta file to Arrow Record
func (pe *ParquetEngine) serializeDeltaFile(deltaFile *DeltaFile, schema *arrow.Schema) arrow.Record {
	pool := memory.NewGoAllocator()

	// Create schema for delta file metadata
	// Each UPDATE column-value pair gets its own row to support multiple column updates
	deltaSchema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "delta_type", Type: arrow.BinaryTypes.String},
			{Name: "timestamp", Type: arrow.PrimitiveTypes.Int64},
			{Name: "version", Type: arrow.PrimitiveTypes.Int64},
			{Name: "filter_column", Type: arrow.BinaryTypes.String},
			{Name: "filter_operator", Type: arrow.BinaryTypes.String},
			{Name: "filter_value", Type: arrow.BinaryTypes.String},
			{Name: "update_column", Type: arrow.BinaryTypes.String},
			{Name: "update_value", Type: arrow.BinaryTypes.String},
		}, nil,
	)

	builder := array.NewRecordBuilder(pool, deltaSchema)
	defer builder.Release()

	// Serialize filters - use first filter for simplicity
	filterCol := ""
	filterOp := ""
	filterVal := ""
	if len(deltaFile.Filters) > 0 {
		filterCol = deltaFile.Filters[0].Column
		filterOp = deltaFile.Filters[0].Operator
		filterVal = fmt.Sprintf("%v", deltaFile.Filters[0].Value)
	}

	// Create one row per UPDATE column to support multiple column updates
	if len(deltaFile.Updates) == 0 {
		// No updates (e.g., DELETE operation) - create single row with empty update fields
		builder.Field(0).(*array.StringBuilder).Append(deltaFile.Type)
		builder.Field(1).(*array.Int64Builder).Append(deltaFile.Timestamp)
		builder.Field(2).(*array.Int64Builder).Append(deltaFile.Version)
		builder.Field(3).(*array.StringBuilder).Append(filterCol)
		builder.Field(4).(*array.StringBuilder).Append(filterOp)
		builder.Field(5).(*array.StringBuilder).Append(filterVal)
		builder.Field(6).(*array.StringBuilder).Append("")
		builder.Field(7).(*array.StringBuilder).Append("")
	} else {
		// Create one row per column-value pair in Updates map
		for col, val := range deltaFile.Updates {
			builder.Field(0).(*array.StringBuilder).Append(deltaFile.Type)
			builder.Field(1).(*array.Int64Builder).Append(deltaFile.Timestamp)
			builder.Field(2).(*array.Int64Builder).Append(deltaFile.Version)
			builder.Field(3).(*array.StringBuilder).Append(filterCol)
			builder.Field(4).(*array.StringBuilder).Append(filterOp)
			builder.Field(5).(*array.StringBuilder).Append(filterVal)
			builder.Field(6).(*array.StringBuilder).Append(col)
			builder.Field(7).(*array.StringBuilder).Append(fmt.Sprintf("%v", val))
		}
	}

	return builder.NewRecord()
}

// estimateAffectedRows estimates the number of rows affected by filters
func (pe *ParquetEngine) estimateAffectedRows(ctx context.Context, tableID string, filters []Filter) int64 {
	// Get snapshot
	snapshot, err := pe.deltaLog.GetSnapshot(tableID, -1)
	if err != nil {
		return 0
	}

	if len(filters) == 0 {
		totalRows := int64(0)
		for _, file := range snapshot.Files {
			totalRows += file.RowCount
		}
		return totalRows
	}

	// Scan and count actual matching rows for accurate count
	// Extract db and table from tableID
	db, table := "", ""
	for i := 0; i < len(tableID); i++ {
		if tableID[i] == '.' {
			db = tableID[:i]
			table = tableID[i+1:]
			break
		}
	}

	iter, err := pe.Scan(ctx, db, table, filters)
	if err != nil {
		// Fall back to rough estimate if scan fails
		totalRows := int64(0)
		for _, file := range snapshot.Files {
			totalRows += file.RowCount
		}
		return totalRows / 10
	}
	defer iter.Close()

	count := int64(0)
	for iter.Next() {
		record := iter.Record()
		if record != nil {
			count += record.NumRows()
		}
	}

	return count
}

// MergeOnReadIterator implements iterator with delta file merging
type MergeOnReadIterator struct {
	baseIterator    RecordIterator   // Iterator for old base files (affected by deltas)
	newBaseIterator RecordIterator   // Iterator for new base files (immune to deltas)
	baseFiles       []delta.FileInfo // Track old base files for timestamp filtering
	deltaFiles      []delta.FileInfo
	currentRecord   arrow.Record
	processingOld   bool // true if processing old files, false if processing new files
	err             error
}

// NewMergeOnReadIterator creates a new merge-on-read iterator
func NewMergeOnReadIterator(baseFiles, deltaFiles []delta.FileInfo, filters []Filter) (RecordIterator, error) {
	// Find the MINIMUM delta timestamp - any file added after ANY delta should be immune
	// Delta files should ONLY apply to files that existed BEFORE the delta was created
	// This ensures correct Merge-on-Read semantics
	var minDeltaTimestamp int64 = int64(^uint64(0) >> 1) // Max int64
	hasDelta := false
	for _, deltaFile := range deltaFiles {
		hasDelta = true
		if deltaFile.AddedAt < minDeltaTimestamp {
			minDeltaTimestamp = deltaFile.AddedAt
		}
	}

	// If no deltas, all files are "new" (no merging needed)
	if !hasDelta {
		minDeltaTimestamp = 0
	}

	// Separate base files into "old" (affected by deltas) and "new" (immune to all deltas)
	var oldBaseFiles, newBaseFiles []delta.FileInfo
	for _, baseFile := range baseFiles {
		if hasDelta && baseFile.AddedAt >= minDeltaTimestamp {
			// This file was added AT OR AFTER any delta file was created
			// No deltas should apply to it - it contains fresh data
			newBaseFiles = append(newBaseFiles, baseFile)
		} else {
			// This file was added BEFORE all deltas
			// All deltas should be applied to it
			oldBaseFiles = append(oldBaseFiles, baseFile)
		}
	}

	// Create base iterator for old files (will have deltas applied)
	var oldIterator RecordIterator
	var err error
	if len(oldBaseFiles) > 0 {
		oldIterator, err = NewParquetIterator(oldBaseFiles, filters)
		if err != nil {
			return nil, fmt.Errorf("failed to create old base iterator: %w", err)
		}
	}

	// Create base iterator for new files (will NOT have deltas applied)
	var newIterator RecordIterator
	if len(newBaseFiles) > 0 {
		newIterator, err = NewParquetIterator(newBaseFiles, filters)
		if err != nil {
			if oldIterator != nil {
				oldIterator.Close()
			}
			return nil, fmt.Errorf("failed to create new base iterator: %w", err)
		}
	}

	logger.Info("Split base files for merge-on-read",
		zap.Int("total_base_files", len(baseFiles)),
		zap.Int("old_base_files", len(oldBaseFiles)),
		zap.Int("new_base_files", len(newBaseFiles)),
		zap.Int("delta_files", len(deltaFiles)))

	// Create merge-on-read iterator
	return &MergeOnReadIterator{
		baseIterator:    oldIterator,
		newBaseIterator: newIterator,
		baseFiles:       oldBaseFiles,
		deltaFiles:      deltaFiles,
		processingOld:   oldIterator != nil,
	}, nil
}

// Next advances to the next record
func (m *MergeOnReadIterator) Next() bool {
	// First process old files with ALL deltas applied
	if m.processingOld {
		if m.baseIterator != nil && m.baseIterator.Next() {
			baseRecord := m.baseIterator.Record()
			// Apply ALL delta files to merge changes (both UPDATE and DELETE)
			mergedRecord := m.applyDeltas(baseRecord, m.deltaFiles)
			m.currentRecord = mergedRecord
			return true
		}
		// Done with old files, switch to new files
		m.processingOld = false
	}

	// Then process new files WITHOUT any deltas (fresh data)
	if m.newBaseIterator != nil && m.newBaseIterator.Next() {
		newRecord := m.newBaseIterator.Record()
		// New files are added AFTER deltas, so NO deltas should apply
		// Return the record as-is without any merging
		newRecord.Retain()
		m.currentRecord = newRecord
		return true
	}

	return false
}

// filterUpdateDeltas returns only UPDATE deltas from the delta files
func (m *MergeOnReadIterator) filterUpdateDeltas(deltaFiles []delta.FileInfo) []delta.FileInfo {
	var updateDeltas []delta.FileInfo
	for _, delta := range deltaFiles {
		if delta.DeltaType == "update" {
			updateDeltas = append(updateDeltas, delta)
		}
	}
	return updateDeltas
}

// Record returns the current record
func (m *MergeOnReadIterator) Record() arrow.Record {
	return m.currentRecord
}

// Err returns any error
func (m *MergeOnReadIterator) Err() error {
	if m.err != nil {
		return m.err
	}
	if m.baseIterator != nil {
		if err := m.baseIterator.Err(); err != nil {
			return err
		}
	}
	if m.newBaseIterator != nil {
		if err := m.newBaseIterator.Err(); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the iterator
func (m *MergeOnReadIterator) Close() error {
	if m.currentRecord != nil {
		m.currentRecord.Release()
	}
	var err1, err2 error
	if m.baseIterator != nil {
		err1 = m.baseIterator.Close()
	}
	if m.newBaseIterator != nil {
		err2 = m.newBaseIterator.Close()
	}
	if err1 != nil {
		return err1
	}
	return err2
}

// applyDeltas applies delta files to base record
func (m *MergeOnReadIterator) applyDeltas(baseRecord arrow.Record, deltaFiles []delta.FileInfo) arrow.Record {
	if len(deltaFiles) == 0 || baseRecord.NumRows() == 0 {
		baseRecord.Retain()
		return baseRecord
	}

	logger.Info("Applying deltas to base record",
		zap.Int("base_rows", int(baseRecord.NumRows())),
		zap.Int("delta_count", len(deltaFiles)))

	// Read all delta files and build update/delete maps
	updateMap := make(map[int64]map[string]interface{}) // rowID -> column -> value
	deleteSet := make(map[int64]bool)                   // rowID -> deleted

	for _, deltaFile := range deltaFiles {
		// Read delta file
		deltaRecord, err := m.readDeltaFile(deltaFile.Path)
		if err != nil {
			logger.Error("Failed to read delta file", zap.Error(err), zap.String("path", deltaFile.Path))
			continue
		}
		defer deltaRecord.Release()

		if deltaRecord.NumRows() == 0 {
			continue
		}

		// Parse delta metadata
		deltaType := m.getDeltaType(deltaRecord)
		filterColumn, filterOperator, filterValue := m.getFilterInfo(deltaRecord)

		logger.Info("Processing delta",
			zap.String("type", deltaType),
			zap.String("filter_column", filterColumn),
			zap.String("filter_operator", filterOperator),
			zap.String("filter_value", filterValue),
			zap.Int("delta_rows", int(deltaRecord.NumRows())))

		// Apply delta based on type
		switch deltaType {
		case "update":
			// Each row in delta record represents one column update
			for rowIdx := int64(0); rowIdx < deltaRecord.NumRows(); rowIdx++ {
				updateColumn, updateValue := m.getUpdateInfoAtRow(deltaRecord, rowIdx)
				logger.Info("Applying UPDATE delta row",
					zap.Int64("row_idx", rowIdx),
					zap.String("update_column", updateColumn),
					zap.String("update_value", updateValue))
				m.applyUpdateDelta(baseRecord, filterColumn, filterOperator, filterValue, updateColumn, updateValue, updateMap)
			}

		case "delete":
			m.applyDeleteDelta(baseRecord, filterColumn, filterOperator, filterValue, deleteSet)
		}
	}

	// Apply collected deltas to base record
	return m.buildMergedRecord(baseRecord, updateMap, deleteSet)
}

// readDeltaFile reads a delta file and returns its record
func (m *MergeOnReadIterator) readDeltaFile(path string) (arrow.Record, error) {
	// Read delta file without filters (we need the raw metadata)
	return parquet.ReadParquetFile(path, nil)
}

// getDeltaType extracts delta type from delta record
func (m *MergeOnReadIterator) getDeltaType(record arrow.Record) string {
	// Delta record schema: [delta_type, timestamp, version, filter_column, filter_operator, filter_value, update_column, update_value]
	if record.Schema().Field(0).Name != "delta_type" {
		return ""
	}
	col := record.Column(0).(*array.String)
	if col.Len() > 0 {
		return col.Value(0)
	}
	return ""
}

// getFilterInfo extracts filter information from delta record
func (m *MergeOnReadIterator) getFilterInfo(record arrow.Record) (column, operator, value string) {
	// Extract filter_column (field 3)
	if record.NumCols() > 3 && record.Schema().Field(3).Name == "filter_column" {
		col := record.Column(3).(*array.String)
		if col.Len() > 0 {
			column = col.Value(0)
		}
	}

	// Extract filter_operator (field 4)
	if record.NumCols() > 4 && record.Schema().Field(4).Name == "filter_operator" {
		col := record.Column(4).(*array.String)
		if col.Len() > 0 {
			operator = col.Value(0)
		}
	}

	// Extract filter_value (field 5)
	if record.NumCols() > 5 && record.Schema().Field(5).Name == "filter_value" {
		col := record.Column(5).(*array.String)
		if col.Len() > 0 {
			value = col.Value(0)
		}
	}

	return
}

// getUpdateInfo extracts update column and value from delta record (deprecated, use getUpdateInfoAtRow)
func (m *MergeOnReadIterator) getUpdateInfo(record arrow.Record) (column, value string) {
	return m.getUpdateInfoAtRow(record, 0)
}

// getUpdateInfoAtRow extracts update column and value from a specific row in delta record
func (m *MergeOnReadIterator) getUpdateInfoAtRow(record arrow.Record, rowIdx int64) (column, value string) {
	// Extract update_column (field 6)
	if record.NumCols() > 6 && record.Schema().Field(6).Name == "update_column" {
		col := record.Column(6).(*array.String)
		if col.Len() > int(rowIdx) {
			column = col.Value(int(rowIdx))
		}
	}

	// Extract update_value (field 7)
	if record.NumCols() > 7 && record.Schema().Field(7).Name == "update_value" {
		col := record.Column(7).(*array.String)
		if col.Len() > int(rowIdx) {
			value = col.Value(int(rowIdx))
		}
	}

	return
}

// applyUpdateDelta marks rows for update based on filter
func (m *MergeOnReadIterator) applyUpdateDelta(baseRecord arrow.Record, filterColumn, filterOperator, filterValue, updateColumn, updateValue string, updateMap map[int64]map[string]interface{}) {
	// updateColumn must not be empty
	if updateColumn == "" {
		logger.Warn("Skipping UPDATE delta: empty update column",
			zap.String("update_column", updateColumn))
		return
	}

	// Find filter column index (only if filter is specified)
	filterColIdx := -1
	if filterColumn != "" {
		for i := 0; i < int(baseRecord.NumCols()); i++ {
			if baseRecord.Schema().Field(i).Name == filterColumn {
				filterColIdx = i
				break
			}
		}
		if filterColIdx == -1 {
			logger.Warn("Filter column not found in base record",
				zap.String("filter_column", filterColumn))
			return
		}
	}

	// Find update column index to get correct type
	updateColIdx := -1
	for i := 0; i < int(baseRecord.NumCols()); i++ {
		if baseRecord.Schema().Field(i).Name == updateColumn {
			updateColIdx = i
			break
		}
	}
	if updateColIdx == -1 {
		logger.Warn("Update column not found in base record",
			zap.String("update_column", updateColumn))
		return // Update column not found in schema
	}

	logger.Info("Evaluating UPDATE delta against base record",
		zap.Int("filter_col_idx", filterColIdx),
		zap.Int("update_col_idx", updateColIdx),
		zap.Int("base_rows", int(baseRecord.NumRows())))

	// If no filter specified, update ALL rows
	if filterColumn == "" {
		for rowIdx := 0; rowIdx < int(baseRecord.NumRows()); rowIdx++ {
			if updateMap[int64(rowIdx)] == nil {
				updateMap[int64(rowIdx)] = make(map[string]interface{})
			}
			parsedValue := m.parseValue(updateValue, baseRecord.Schema().Field(updateColIdx).Type)
			updateMap[int64(rowIdx)][updateColumn] = parsedValue
			logger.Info("Row updated (no WHERE clause)",
				zap.Int("row_idx", rowIdx),
				zap.String("update_column", updateColumn),
				zap.Any("update_value", parsedValue))
		}
		logger.Info("UPDATE delta evaluation complete (no filter)",
			zap.Int("updated_rows", int(baseRecord.NumRows())),
			zap.Int("total_updates_in_map", len(updateMap)))
		return
	}

	// Evaluate filter and mark rows for update
	matchCount := 0
	for rowIdx := 0; rowIdx < int(baseRecord.NumRows()); rowIdx++ {
		if m.evaluateFilter(baseRecord, rowIdx, filterColIdx, filterOperator, filterValue) {
			matchCount++
			if updateMap[int64(rowIdx)] == nil {
				updateMap[int64(rowIdx)] = make(map[string]interface{})
			}
			// Parse update value based on UPDATE column type (not filter column type!)
			parsedValue := m.parseValue(updateValue, baseRecord.Schema().Field(updateColIdx).Type)
			updateMap[int64(rowIdx)][updateColumn] = parsedValue
			logger.Info("Row matched UPDATE filter",
				zap.Int("row_idx", rowIdx),
				zap.String("update_column", updateColumn),
				zap.Any("update_value", parsedValue))
		}
	}

	logger.Info("UPDATE delta evaluation complete",
		zap.Int("matched_rows", matchCount),
		zap.Int("total_updates_in_map", len(updateMap)))
}

// applyDeleteDelta marks rows for deletion based on filter
func (m *MergeOnReadIterator) applyDeleteDelta(baseRecord arrow.Record, filterColumn, filterOperator, filterValue string, deleteSet map[int64]bool) {
	// If no filter specified, delete ALL rows
	if filterColumn == "" {
		for rowIdx := 0; rowIdx < int(baseRecord.NumRows()); rowIdx++ {
			deleteSet[int64(rowIdx)] = true
		}
		return
	}

	// Find filter column index
	filterColIdx := -1
	for i := 0; i < int(baseRecord.NumCols()); i++ {
		if baseRecord.Schema().Field(i).Name == filterColumn {
			filterColIdx = i
			break
		}
	}
	if filterColIdx == -1 {
		return
	}

	// Evaluate filter and mark rows for deletion
	for rowIdx := 0; rowIdx < int(baseRecord.NumRows()); rowIdx++ {
		if m.evaluateFilter(baseRecord, rowIdx, filterColIdx, filterOperator, filterValue) {
			deleteSet[int64(rowIdx)] = true
		}
	}
}

// evaluateFilter evaluates filter condition for a row
func (m *MergeOnReadIterator) evaluateFilter(record arrow.Record, rowIdx, colIdx int, operator, value string) bool {
	col := record.Column(colIdx)

	switch col.DataType().ID() {
	case arrow.INT64:
		intCol := col.(*array.Int64)
		rowValue := intCol.Value(rowIdx)
		filterValue := m.parseInt64(value)
		return m.compareInt64(rowValue, filterValue, operator)

	case arrow.STRING:
		strCol := col.(*array.String)
		rowValue := strCol.Value(rowIdx)
		return m.compareString(rowValue, value, operator)

	case arrow.BOOL:
		boolCol := col.(*array.Boolean)
		rowValue := boolCol.Value(rowIdx)
		filterValue := m.parseBool(value)
		return m.compareBool(rowValue, filterValue, operator)

	default:
		return false
	}
}

// parseBool parses a string to boolean
// Handles: "true", "1", "t", "T", "TRUE" -> true
//
//	"false", "0", "f", "F", "FALSE" -> false
func (m *MergeOnReadIterator) parseBool(s string) bool {
	switch s {
	case "true", "1", "t", "T", "TRUE":
		return true
	case "false", "0", "f", "F", "FALSE":
		return false
	default:
		return false
	}
}

// compareBool compares two boolean values based on operator
func (m *MergeOnReadIterator) compareBool(a, b bool, operator string) bool {
	switch operator {
	case "=":
		return a == b
	case "!=":
		return a != b
	default:
		return false
	}
}

// compareInt64 compares two int64 values based on operator
func (m *MergeOnReadIterator) compareInt64(a, b int64, operator string) bool {
	switch operator {
	case "=":
		return a == b
	case "<":
		return a < b
	case ">":
		return a > b
	case "<=":
		return a <= b
	case ">=":
		return a >= b
	case "!=":
		return a != b
	default:
		return false
	}
}

// compareString compares two strings based on operator
func (m *MergeOnReadIterator) compareString(a, b, operator string) bool {
	switch operator {
	case "=":
		return a == b
	case "!=":
		return a != b
	default:
		return false
	}
}

// parseInt64 parses a string to int64
func (m *MergeOnReadIterator) parseInt64(s string) int64 {
	var val int64
	fmt.Sscanf(s, "%d", &val)
	return val
}

// parseValue parses a string value based on Arrow type
func (m *MergeOnReadIterator) parseValue(s string, dataType arrow.DataType) interface{} {
	switch dataType.ID() {
	case arrow.INT64:
		var val int64
		fmt.Sscanf(s, "%d", &val)
		return val
	case arrow.INT32:
		var val int32
		fmt.Sscanf(s, "%d", &val)
		return val
	case arrow.FLOAT64:
		var val float64
		// Try parsing as float first
		if _, err := fmt.Sscanf(s, "%f", &val); err == nil {
			return val
		}
		// Fallback: try parsing as int and convert to float
		var intVal int64
		if _, err := fmt.Sscanf(s, "%d", &intVal); err == nil {
			return float64(intVal)
		}
		return 0.0
	case arrow.FLOAT32:
		var val float32
		// Try parsing as float first
		if _, err := fmt.Sscanf(s, "%f", &val); err == nil {
			return val
		}
		// Fallback: try parsing as int and convert to float
		var intVal int64
		if _, err := fmt.Sscanf(s, "%d", &intVal); err == nil {
			return float32(intVal)
		}
		return float32(0.0)
	case arrow.BOOL:
		// Parse boolean values: "true", "1", "t", "T", "TRUE" -> true
		// "false", "0", "f", "F", "FALSE" -> false
		switch s {
		case "true", "1", "t", "T", "TRUE":
			return true
		case "false", "0", "f", "F", "FALSE":
			return false
		default:
			return false
		}
	case arrow.STRING:
		return s
	default:
		return s
	}
}

// buildMergedRecord builds final merged record with updates and deletes applied
func (m *MergeOnReadIterator) buildMergedRecord(baseRecord arrow.Record, updateMap map[int64]map[string]interface{}, deleteSet map[int64]bool) arrow.Record {
	if len(updateMap) == 0 && len(deleteSet) == 0 {
		baseRecord.Retain()
		return baseRecord
	}

	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, baseRecord.Schema())
	defer builder.Release()

	// Process each row
	for rowIdx := 0; rowIdx < int(baseRecord.NumRows()); rowIdx++ {
		// Skip deleted rows
		if deleteSet[int64(rowIdx)] {
			continue
		}

		// Copy or update each column
		for colIdx := 0; colIdx < int(baseRecord.NumCols()); colIdx++ {
			colName := baseRecord.Schema().Field(colIdx).Name
			col := baseRecord.Column(colIdx)

			// Check if this column has an update for this row
			if updateMap[int64(rowIdx)] != nil {
				if updateVal, hasUpdate := updateMap[int64(rowIdx)][colName]; hasUpdate {
					m.appendValue(builder.Field(colIdx), updateVal, col.DataType())
					continue
				}
			}

			// No update, copy original value
			m.copyValue(builder.Field(colIdx), col, rowIdx)
		}
	}

	return builder.NewRecord()
}

// appendValue appends a value to a builder
func (m *MergeOnReadIterator) appendValue(builder array.Builder, value interface{}, dataType arrow.DataType) {
	if value == nil {
		builder.AppendNull()
		return
	}

	switch dataType.ID() {
	case arrow.INT64:
		// Handle different integer types
		switch v := value.(type) {
		case int64:
			builder.(*array.Int64Builder).Append(v)
		case int:
			builder.(*array.Int64Builder).Append(int64(v))
		case int32:
			builder.(*array.Int64Builder).Append(int64(v))
		default:
			builder.AppendNull()
		}
	case arrow.INT32:
		switch v := value.(type) {
		case int32:
			builder.(*array.Int32Builder).Append(v)
		case int:
			builder.(*array.Int32Builder).Append(int32(v))
		case int64:
			builder.(*array.Int32Builder).Append(int32(v))
		default:
			builder.AppendNull()
		}
	case arrow.FLOAT64:
		switch v := value.(type) {
		case float64:
			builder.(*array.Float64Builder).Append(v)
		case float32:
			builder.(*array.Float64Builder).Append(float64(v))
		case int64:
			// Allow int64 to float64 conversion (e.g., UPDATE price = 1099)
			builder.(*array.Float64Builder).Append(float64(v))
		case int:
			builder.(*array.Float64Builder).Append(float64(v))
		case int32:
			builder.(*array.Float64Builder).Append(float64(v))
		default:
			builder.AppendNull()
		}
	case arrow.STRING:
		if v, ok := value.(string); ok {
			builder.(*array.StringBuilder).Append(v)
		} else {
			builder.AppendNull()
		}
	case arrow.BOOL:
		if v, ok := value.(bool); ok {
			builder.(*array.BooleanBuilder).Append(v)
		} else {
			builder.AppendNull()
		}
	default:
		builder.AppendNull()
	}
}

// copyValue copies a value from source column to builder
func (m *MergeOnReadIterator) copyValue(builder array.Builder, col arrow.Array, rowIdx int) {
	if col.IsNull(rowIdx) {
		builder.AppendNull()
		return
	}

	switch col.DataType().ID() {
	case arrow.INT64:
		intCol := col.(*array.Int64)
		builder.(*array.Int64Builder).Append(intCol.Value(rowIdx))
	case arrow.INT32:
		intCol := col.(*array.Int32)
		builder.(*array.Int32Builder).Append(intCol.Value(rowIdx))
	case arrow.INT16:
		intCol := col.(*array.Int16)
		builder.(*array.Int16Builder).Append(intCol.Value(rowIdx))
	case arrow.INT8:
		intCol := col.(*array.Int8)
		builder.(*array.Int8Builder).Append(intCol.Value(rowIdx))
	case arrow.FLOAT64:
		floatCol := col.(*array.Float64)
		builder.(*array.Float64Builder).Append(floatCol.Value(rowIdx))
	case arrow.FLOAT32:
		floatCol := col.(*array.Float32)
		builder.(*array.Float32Builder).Append(floatCol.Value(rowIdx))
	case arrow.STRING:
		strCol := col.(*array.String)
		builder.(*array.StringBuilder).Append(strCol.Value(rowIdx))
	case arrow.BOOL:
		boolCol := col.(*array.Boolean)
		builder.(*array.BooleanBuilder).Append(boolCol.Value(rowIdx))
	default:
		builder.AppendNull()
	}
}

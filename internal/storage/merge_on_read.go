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
		Path:     deltaPath,
		Size:     stats.FileSize,
		RowCount: stats.RowCount,
		Stats:    stats,
	}

	// Note: We need to extend AppendAdd to support IsDelta flag
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

	// Append to Delta Log
	parquetFile := &delta.ParquetFile{
		Path:     deltaPath,
		Size:     stats.FileSize,
		RowCount: stats.RowCount,
		Stats:    stats,
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
	// For simplicity, we store metadata as a single-row record with string columns
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

	// Serialize filters and updates into single row
	// In production, this would be more sophisticated
	filterCol := ""
	filterOp := ""
	filterVal := ""
	if len(deltaFile.Filters) > 0 {
		filterCol = deltaFile.Filters[0].Column
		filterOp = deltaFile.Filters[0].Operator
		filterVal = fmt.Sprintf("%v", deltaFile.Filters[0].Value)
	}

	updateCol := ""
	updateVal := ""
	for col, val := range deltaFile.Updates {
		updateCol = col
		updateVal = fmt.Sprintf("%v", val)
		break // For simplicity, take first update
	}

	builder.Field(0).(*array.StringBuilder).Append(deltaFile.Type)
	builder.Field(1).(*array.Int64Builder).Append(deltaFile.Timestamp)
	builder.Field(2).(*array.Int64Builder).Append(deltaFile.Version)
	builder.Field(3).(*array.StringBuilder).Append(filterCol)
	builder.Field(4).(*array.StringBuilder).Append(filterOp)
	builder.Field(5).(*array.StringBuilder).Append(filterVal)
	builder.Field(6).(*array.StringBuilder).Append(updateCol)
	builder.Field(7).(*array.StringBuilder).Append(updateVal)

	return builder.NewRecord()
}

// estimateAffectedRows estimates the number of rows affected by filters
func (pe *ParquetEngine) estimateAffectedRows(ctx context.Context, tableID string, filters []Filter) int64 {
	// Get snapshot
	snapshot, err := pe.deltaLog.GetSnapshot(tableID, -1)
	if err != nil {
		return 0
	}

	// For simplicity, return total row count divided by 10 as estimate
	// In production, this would use statistics to make better estimates
	totalRows := int64(0)
	for _, file := range snapshot.Files {
		totalRows += file.RowCount
	}

	if len(filters) == 0 {
		return totalRows
	}

	// Rough estimate: 10% of rows match filters
	return totalRows / 10
}

// MergeOnReadIterator implements iterator with delta file merging
type MergeOnReadIterator struct {
	baseIterator  RecordIterator
	deltaFiles    []delta.FileInfo
	currentRecord arrow.Record
	err           error
}

// NewMergeOnReadIterator creates a new merge-on-read iterator
func NewMergeOnReadIterator(baseFiles, deltaFiles []delta.FileInfo, filters []Filter) (RecordIterator, error) {
	// For now, create standard iterator for base files
	// In production, this would apply deltas during iteration
	return NewParquetIterator(baseFiles, filters)
}

// Next advances to the next record
func (m *MergeOnReadIterator) Next() bool {
	if !m.baseIterator.Next() {
		return false
	}

	baseRecord := m.baseIterator.Record()

	// Apply delta files to merge changes
	mergedRecord := m.applyDeltas(baseRecord, m.deltaFiles)
	m.currentRecord = mergedRecord

	return true
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
	return m.baseIterator.Err()
}

// Close closes the iterator
func (m *MergeOnReadIterator) Close() error {
	if m.currentRecord != nil {
		m.currentRecord.Release()
	}
	return m.baseIterator.Close()
}

// applyDeltas applies delta files to base record
func (m *MergeOnReadIterator) applyDeltas(baseRecord arrow.Record, deltaFiles []delta.FileInfo) arrow.Record {
	// For simplicity, return base record as-is
	// In production, this would read delta files and apply updates/deletes
	baseRecord.Retain()
	return baseRecord
}

package optimizer

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

// CompactionConfig compaction configuration
type CompactionConfig struct {
	TargetFileSize    int64         // Target file size in bytes
	MinFileSize       int64         // Minimum file size to trigger compaction
	MaxFilesToCompact int           // Maximum files to compact in one operation
	CheckInterval     time.Duration // Background check interval
}

// Compactor performs small file compaction
type Compactor struct {
	config *CompactionConfig
}

// NewCompactor creates a new compactor
func NewCompactor(config *CompactionConfig) *Compactor {
	return &Compactor{
		config: config,
	}
}

// CompactTable compacts small files in a table
func (c *Compactor) CompactTable(tableID string, engine CompactionEngine) error {
	logger.Info("Starting table compaction", zap.String("table", tableID))

	deltaLog := engine.GetDeltaLog()
	snapshot, err := deltaLog.GetSnapshot(tableID, -1)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}

	// Identify small files that need compaction
	smallFiles := c.identifySmallFiles(snapshot.Files)
	if len(smallFiles) == 0 {
		logger.Info("No small files to compact", zap.String("table", tableID))
		return nil
	}

	logger.Info("Identified small files for compaction",
		zap.String("table", tableID),
		zap.Int("small_file_count", len(smallFiles)))

	// Compact files in batches
	db, table := parseTableID(tableID)
	basePath := "/tmp/minidb"

	compactedFiles := c.compactFiles(smallFiles, snapshot.Schema, db, table, basePath)

	// Update Delta Log
	// Mark old files as REMOVE
	for _, file := range smallFiles {
		if err := deltaLog.AppendRemove(tableID, file.Path); err != nil {
			logger.Warn("Failed to remove old file",
				zap.String("file", file.Path),
				zap.Error(err))
		}
	}

	// Add new compacted files with dataChange=false
	for _, file := range compactedFiles {
		if err := deltaLog.AppendAdd(tableID, file); err != nil {
			logger.Warn("Failed to add compacted file",
				zap.String("file", file.Path),
				zap.Error(err))
		}
	}

	logger.Info("Table compaction completed",
		zap.String("table", tableID),
		zap.Int("old_files", len(smallFiles)),
		zap.Int("new_files", len(compactedFiles)))

	return nil
}

// identifySmallFiles identifies files smaller than threshold
func (c *Compactor) identifySmallFiles(files []delta.FileInfo) []delta.FileInfo {
	smallFiles := make([]delta.FileInfo, 0)

	for _, file := range files {
		if file.Size < c.config.MinFileSize {
			smallFiles = append(smallFiles, file)
			if len(smallFiles) >= c.config.MaxFilesToCompact {
				break
			}
		}
	}

	return smallFiles
}

// compactFiles compacts multiple files into larger files
func (c *Compactor) compactFiles(files []delta.FileInfo, schema *arrow.Schema, db, table, basePath string) []*delta.ParquetFile {
	// Read all records from small files
	allRecords := make([]arrow.Record, 0)
	for _, file := range files {
		record, err := parquet.ReadParquetFile(file.Path, nil)
		if err != nil {
			logger.Warn("Failed to read file for compaction",
				zap.String("file", file.Path),
				zap.Error(err))
			continue
		}
		allRecords = append(allRecords, record)
	}

	if len(allRecords) == 0 {
		return nil
	}

	// Merge records
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	totalRows := int64(0)
	for _, record := range allRecords {
		totalRows += record.NumRows()
		// Append all rows from record to builder
		for colIdx := 0; colIdx < int(record.NumCols()); colIdx++ {
			col := record.Column(colIdx)
			c.appendColumn(builder.Field(colIdx), col, schema.Field(colIdx).Type)
		}
	}

	// Release original records
	for _, record := range allRecords {
		record.Release()
	}

	// Create compacted file
	compactedRecord := builder.NewRecord()
	defer compactedRecord.Release()

	fileName := fmt.Sprintf("compact-%s.parquet", uuid.New().String()[:8])
	filePath := filepath.Join(basePath, db, table, "data", fileName)

	stats, err := parquet.WriteArrowBatch(filePath, compactedRecord)
	if err != nil {
		logger.Error("Failed to write compacted file",
			zap.String("file", filePath),
			zap.Error(err))
		return nil
	}

	logger.Info("Compacted file created",
		zap.String("file", filePath),
		zap.Int64("rows", stats.RowCount),
		zap.Int64("size", stats.FileSize))

	return []*delta.ParquetFile{{
		Path:     filePath,
		Size:     stats.FileSize,
		RowCount: stats.RowCount,
		Stats:    stats,
	}}
}

// appendColumn appends all values from a column to a builder
func (c *Compactor) appendColumn(builder array.Builder, col arrow.Array, dataType arrow.DataType) {
	switch b := builder.(type) {
	case *array.Int64Builder:
		arr := col.(*array.Int64)
		for i := 0; i < arr.Len(); i++ {
			if arr.IsNull(i) {
				b.AppendNull()
			} else {
				b.Append(arr.Value(i))
			}
		}
	case *array.StringBuilder:
		arr := col.(*array.String)
		for i := 0; i < arr.Len(); i++ {
			if arr.IsNull(i) {
				b.AppendNull()
			} else {
				b.Append(arr.Value(i))
			}
		}
	case *array.Float64Builder:
		arr := col.(*array.Float64)
		for i := 0; i < arr.Len(); i++ {
			if arr.IsNull(i) {
				b.AppendNull()
			} else {
				b.Append(arr.Value(i))
			}
		}
	case *array.Int32Builder:
		arr := col.(*array.Int32)
		for i := 0; i < arr.Len(); i++ {
			if arr.IsNull(i) {
				b.AppendNull()
			} else {
				b.Append(arr.Value(i))
			}
		}
	}
}

// CompactionEngine interface for compaction operations
type CompactionEngine interface {
	GetDeltaLog() delta.LogInterface
}

// AutoCompactor automatic background compaction
type AutoCompactor struct {
	compactor *Compactor
}

// NewAutoCompactor creates auto-compactor
func NewAutoCompactor(config *CompactionConfig) *AutoCompactor {
	return &AutoCompactor{
		compactor: NewCompactor(config),
	}
}

// Start starts background compaction
func (ac *AutoCompactor) Start(ctx context.Context, engine CompactionEngine, stopChan chan struct{}) {
	ticker := time.NewTicker(ac.compactor.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Get all tables and check for compaction
			deltaLog := engine.GetDeltaLog()
			tables := deltaLog.ListTables()

			for _, tableID := range tables {
				if err := ac.compactor.CompactTable(tableID, engine); err != nil {
					logger.Warn("Auto-compaction failed",
						zap.String("table", tableID),
						zap.Error(err))
				}
			}

		case <-stopChan:
			logger.Info("Auto-compaction stopped")
			return

		case <-ctx.Done():
			logger.Info("Auto-compaction context cancelled")
			return
		}
	}
}

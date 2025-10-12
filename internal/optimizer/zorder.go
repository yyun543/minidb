package optimizer

import (
	"fmt"
	"hash/fnv"
	"path/filepath"
	"sort"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/google/uuid"
	"github.com/yyun543/minidb/internal/delta"
	"github.com/yyun543/minidb/internal/logger"
	"github.com/yyun543/minidb/internal/parquet"
	"go.uber.org/zap"
)

// ZOrderOptimizer implements Z-Order multi-dimensional clustering
type ZOrderOptimizer struct {
	columns    []string
	bitsPerDim int
}

// ZOrderedRow represents a row with its Z-Order value
type ZOrderedRow struct {
	ZValue uint64
	Record arrow.Record
	RowIdx int
}

// StorageEngine interface for Z-Order operations
type StorageEngine interface {
	GetDeltaLog() delta.LogInterface
	GetTableSchema(db, table string) (*arrow.Schema, error)
}

// NewZOrderOptimizer creates a new Z-Order optimizer
func NewZOrderOptimizer(columns []string) *ZOrderOptimizer {
	return &ZOrderOptimizer{
		columns:    columns,
		bitsPerDim: 21, // 21 bits per dimension, supports 2^21 = 2M distinct values
	}
}

// OptimizeTable applies Z-Order clustering to a table
func (z *ZOrderOptimizer) OptimizeTable(tableID string, files []delta.FileInfo, engine StorageEngine) error {
	logger.Info("Starting Z-Order optimization",
		zap.String("table", tableID),
		zap.Strings("columns", z.columns),
		zap.Int("file_count", len(files)))

	if len(files) == 0 {
		return fmt.Errorf("no files to optimize")
	}

	// Extract db and table names from tableID
	db, table := parseTableID(tableID)

	// 1. Read all data from existing files
	allRecords, err := z.readAllFiles(files)
	if err != nil {
		return fmt.Errorf("failed to read files: %w", err)
	}
	defer z.releaseRecords(allRecords)

	if len(allRecords) == 0 {
		return fmt.Errorf("no records found in files")
	}

	// 2. Compute Z-Order values and sort
	zOrderedRows := z.computeZOrder(allRecords)
	logger.Info("Z-Order values computed",
		zap.Int("row_count", len(zOrderedRows)))

	// 3. Repartition and write new files
	basePath := "/tmp/minidb"
	targetFileSize := int64(1024 * 1024 * 1024) // 1GB
	newFiles := z.partitionAndWrite(tableID, db, table, zOrderedRows, allRecords[0].Schema(), basePath, targetFileSize)

	// 4. Update Delta Log
	deltaLog := engine.GetDeltaLog()

	// Mark old files as REMOVE
	for _, file := range files {
		if err := deltaLog.AppendRemove(tableID, file.Path); err != nil {
			return fmt.Errorf("failed to mark file for removal: %w", err)
		}
	}

	// Add new Z-Ordered files
	for _, newFile := range newFiles {
		if err := deltaLog.AppendAdd(tableID, newFile); err != nil {
			return fmt.Errorf("failed to add new file: %w", err)
		}
	}

	logger.Info("Z-Order optimization completed",
		zap.String("table", tableID),
		zap.Int("old_files", len(files)),
		zap.Int("new_files", len(newFiles)))

	return nil
}

// readAllFiles reads all Arrow records from Parquet files
func (z *ZOrderOptimizer) readAllFiles(files []delta.FileInfo) ([]arrow.Record, error) {
	records := make([]arrow.Record, 0, len(files))

	for _, file := range files {
		// Use parquet.ReadParquetFile with no filters to read entire file
		record, err := parquet.ReadParquetFile(file.Path, nil)
		if err != nil {
			logger.Warn("Failed to read file, skipping",
				zap.String("file", file.Path),
				zap.Error(err))
			continue
		}
		records = append(records, record)
	}

	return records, nil
}

// computeZOrder computes Z-Order values for all rows
func (z *ZOrderOptimizer) computeZOrder(records []arrow.Record) []ZOrderedRow {
	var allRows []ZOrderedRow

	for _, record := range records {
		for rowIdx := 0; rowIdx < int(record.NumRows()); rowIdx++ {
			zValue := z.computeZValue(record, rowIdx)

			row := ZOrderedRow{
				ZValue: zValue,
				Record: record,
				RowIdx: rowIdx,
			}
			allRows = append(allRows, row)
		}
	}

	// Sort by Z-Value
	sort.Slice(allRows, func(i, j int) bool {
		return allRows[i].ZValue < allRows[j].ZValue
	})

	return allRows
}

// computeZValue computes the Z-Order value for a row
func (z *ZOrderOptimizer) computeZValue(record arrow.Record, rowIdx int) uint64 {
	var zValue uint64

	// Get dimension values
	dimValues := make([]uint64, len(z.columns))
	for i, colName := range z.columns {
		col := z.findColumn(record, colName)
		if col == nil {
			continue
		}
		rawValue := z.getValueFromColumn(col, rowIdx)
		dimValues[i] = z.normalizeValue(rawValue, col.DataType())
	}

	// Bit interleaving
	for bitPos := 0; bitPos < z.bitsPerDim; bitPos++ {
		for dimIdx, dimValue := range dimValues {
			bit := (dimValue >> bitPos) & 1
			zValue |= bit << (bitPos*len(z.columns) + dimIdx)
		}
	}

	return zValue
}

// ComputeZValueFromDimensions computes Z-Value from pre-normalized dimension values (for testing)
func (z *ZOrderOptimizer) ComputeZValueFromDimensions(dimValues []uint64) uint64 {
	var zValue uint64

	// Bit interleaving
	for bitPos := 0; bitPos < z.bitsPerDim; bitPos++ {
		for dimIdx, dimValue := range dimValues {
			bit := (dimValue >> bitPos) & 1
			zValue |= bit << (bitPos*len(dimValues) + dimIdx)
		}
	}

	return zValue
}

// findColumn finds a column by name in the record
func (z *ZOrderOptimizer) findColumn(record arrow.Record, colName string) arrow.Array {
	for i := 0; i < int(record.NumCols()); i++ {
		if record.Schema().Field(i).Name == colName {
			return record.Column(i)
		}
	}
	return nil
}

// getValueFromColumn extracts value from a column at given row index
func (z *ZOrderOptimizer) getValueFromColumn(col arrow.Array, rowIdx int) interface{} {
	if col.IsNull(rowIdx) {
		return int64(0)
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
		if arr.Value(rowIdx) {
			return int64(1)
		}
		return int64(0)
	default:
		return int64(0)
	}
}

// normalizeValue normalizes a value to the bit space
func (z *ZOrderOptimizer) normalizeValue(value interface{}, dataType arrow.DataType) uint64 {
	switch dataType.ID() {
	case arrow.INT64, arrow.INT32, arrow.INT16, arrow.INT8:
		if v, ok := value.(int64); ok {
			// Map signed int to unsigned space and scale to bitsPerDim
			return uint64(v) >> (64 - z.bitsPerDim)
		}
		return 0

	case arrow.FLOAT64, arrow.FLOAT32:
		if v, ok := value.(float64); ok {
			// Convert float to fixed-point representation
			scale := float64(uint64(1) << uint(z.bitsPerDim))
			return uint64(v * scale)
		}
		return 0

	case arrow.TIMESTAMP:
		if v, ok := value.(int64); ok {
			// Normalize timestamp to relative time
			yearInSeconds := int64(365 * 24 * 3600)
			normalized := v % yearInSeconds
			return uint64(normalized) >> (64 - z.bitsPerDim)
		}
		return 0

	case arrow.STRING:
		if v, ok := value.(string); ok {
			// Hash string to numeric space
			hash := fnv.New64a()
			hash.Write([]byte(v))
			return hash.Sum64() >> (64 - z.bitsPerDim)
		}
		return 0

	default:
		return 0
	}
}

// partitionAndWrite partitions Z-Ordered data into files
func (z *ZOrderOptimizer) partitionAndWrite(tableID, db, table string, zOrderedRows []ZOrderedRow, schema *arrow.Schema, basePath string, targetFileSize int64) []*delta.ParquetFile {
	pool := memory.NewGoAllocator()
	var newFiles []*delta.ParquetFile

	currentBuilder := array.NewRecordBuilder(pool, schema)
	currentSize := int64(0)
	fileIdx := 0

	for _, zRow := range zOrderedRows {
		// Append row to current builder
		for colIdx := 0; colIdx < int(schema.NumFields()); colIdx++ {
			col := zRow.Record.Column(colIdx)
			value := z.getValueFromColumn(col, zRow.RowIdx)
			z.appendValueToBuilder(currentBuilder.Field(colIdx), value, schema.Field(colIdx).Type)
		}

		currentSize += 100 // Approximate row size

		// Write file if target size reached
		if currentSize >= targetFileSize {
			file := z.writePartitionFile(tableID, db, table, currentBuilder, basePath, fileIdx)
			if file != nil {
				newFiles = append(newFiles, file)
				fileIdx++
			}
			currentBuilder.Release()
			currentBuilder = array.NewRecordBuilder(pool, schema)
			currentSize = 0
		}
	}

	// Write remaining data
	if currentSize > 0 {
		file := z.writePartitionFile(tableID, db, table, currentBuilder, basePath, fileIdx)
		if file != nil {
			newFiles = append(newFiles, file)
		}
	}
	currentBuilder.Release()

	return newFiles
}

// writePartitionFile writes a single partition file
func (z *ZOrderOptimizer) writePartitionFile(tableID, db, table string, builder *array.RecordBuilder, basePath string, fileIdx int) *delta.ParquetFile {
	record := builder.NewRecord()
	defer record.Release()

	if record.NumRows() == 0 {
		return nil
	}

	// Generate file path
	fileName := fmt.Sprintf("zorder-%s-%d.parquet", uuid.New().String()[:8], fileIdx)
	filePath := filepath.Join(basePath, db, table, "data", fileName)

	// Write Parquet file
	stats, err := parquet.WriteArrowBatch(filePath, record)
	if err != nil {
		logger.Error("Failed to write Z-Order partition file",
			zap.String("file", filePath),
			zap.Error(err))
		return nil
	}

	logger.Debug("Z-Order partition written",
		zap.String("file", filePath),
		zap.Int64("rows", stats.RowCount),
		zap.Int64("size", stats.FileSize))

	return &delta.ParquetFile{
		Path:     filePath,
		Size:     stats.FileSize,
		RowCount: stats.RowCount,
		Stats:    stats,
	}
}

// appendValueToBuilder appends a value to the appropriate builder
func (z *ZOrderOptimizer) appendValueToBuilder(builder array.Builder, value interface{}, dataType arrow.DataType) {
	if value == nil {
		builder.AppendNull()
		return
	}

	switch b := builder.(type) {
	case *array.Int64Builder:
		if v, ok := value.(int64); ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	case *array.Int32Builder:
		if v, ok := value.(int64); ok {
			b.Append(int32(v))
		} else {
			b.AppendNull()
		}
	case *array.Float64Builder:
		if v, ok := value.(float64); ok {
			b.Append(v)
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
		if v, ok := value.(int64); ok {
			b.Append(v != 0)
		} else {
			b.AppendNull()
		}
	default:
		builder.AppendNull()
	}
}

// releaseRecords releases all Arrow records
func (z *ZOrderOptimizer) releaseRecords(records []arrow.Record) {
	for _, record := range records {
		record.Release()
	}
}

// Helper functions

func parseTableID(tableID string) (string, string) {
	for i, ch := range tableID {
		if ch == '.' {
			return tableID[:i], tableID[i+1:]
		}
	}
	return "default", tableID
}

package parquet

import (
	"context"
	"fmt"
	"os"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/apache/arrow/go/v18/parquet/file"
	"github.com/apache/arrow/go/v18/parquet/pqarrow"
	"github.com/yyun543/minidb/internal/logger"
	"go.uber.org/zap"
)

// Filter 查询过滤条件 (避免循环依赖)
type Filter struct {
	Column   string
	Operator string
	Value    interface{}
}

// ReadParquetFile 读取 Parquet 文件并返回 Arrow Record (使用 Arrow 原生 reader)
func ReadParquetFile(path string, filters []Filter) (arrow.Record, error) {
	logger.Info("Reading Parquet file",
		zap.String("path", path),
		zap.Int("filters", len(filters)))

	// 打开文件
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open parquet file: %w", err)
	}
	defer f.Close()

	// 使用 Arrow 的 Parquet reader
	reader, err := file.NewParquetReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet reader: %w", err)
	}
	defer reader.Close()

	// 创建 Arrow file reader
	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	if err != nil {
		return nil, fmt.Errorf("failed to create arrow file reader: %w", err)
	}

	// 读取整个表为 Arrow Table
	ctx := context.Background()
	table, err := arrowReader.ReadTable(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read table: %w", err)
	}
	defer table.Release()

	// 将 Table 转换为单个 Record
	// Note: 这里简化处理，假设数据量不大
	// 在生产环境中应该使用分批读取或 TableReader
	if table.NumRows() == 0 {
		// 返回空 Record
		schema := table.Schema()
		pool := memory.NewGoAllocator()
		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()
		record := builder.NewRecord()

		logger.Info("Parquet file read (empty)",
			zap.String("path", path),
			zap.Int64("rows", 0))

		return record, nil
	}

	// 将 Table 的所有 chunk 合并为一个 Record
	// 这里使用 TableReader 来获取 Record
	tr := array.NewTableReader(table, table.NumRows())
	defer tr.Release()

	var records []arrow.Record
	for tr.Next() {
		rec := tr.Record()
		rec.Retain() // 保留引用
		records = append(records, rec)
	}

	if err := tr.Err(); err != nil {
		for _, rec := range records {
			rec.Release()
		}
		return nil, fmt.Errorf("failed to read records from table: %w", err)
	}

	// 如果只有一个 record，应用过滤后返回
	if len(records) == 1 {
		record := records[0]
		logger.Info("Parquet file read",
			zap.String("path", path),
			zap.Int64("rows", record.NumRows()))

		// 应用过滤条件
		if len(filters) > 0 {
			filteredRecord, err := applyFilters(record, filters)
			if err != nil {
				record.Release()
				return nil, fmt.Errorf("failed to apply filters: %w", err)
			}
			record.Release()
			return filteredRecord, nil
		}

		return record, nil
	}

	// 合并多个 records 为一个 record
	logger.Info("Merging multiple records",
		zap.String("path", path),
		zap.Int("record_count", len(records)))

	mergedRecord, err := mergeRecords(records)
	if err != nil {
		for _, rec := range records {
			rec.Release()
		}
		return nil, fmt.Errorf("failed to merge records: %w", err)
	}

	// 释放原始 records
	for _, rec := range records {
		rec.Release()
	}

	logger.Info("Parquet file read (merged)",
		zap.String("path", path),
		zap.Int64("rows", mergedRecord.NumRows()),
		zap.Int("merged_from", len(records)))

	// 应用过滤条件 (predicate pushdown)
	if len(filters) > 0 {
		filteredRecord, err := applyFilters(mergedRecord, filters)
		if err != nil {
			mergedRecord.Release()
			return nil, fmt.Errorf("failed to apply filters: %w", err)
		}
		mergedRecord.Release()
		return filteredRecord, nil
	}

	return mergedRecord, nil
}

// mergeRecords 合并多个 Arrow Records 为一个 Record
// 使用 RecordBuilder 逐列合并数据
func mergeRecords(records []arrow.Record) (arrow.Record, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no records to merge")
	}

	schema := records[0].Schema()
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 遍历所有 records 并追加数据
	for _, record := range records {
		// 确保 schema 一致
		if !schema.Equal(record.Schema()) {
			return nil, fmt.Errorf("schema mismatch during record merge")
		}

		// 逐列追加数据
		for colIdx := 0; colIdx < int(record.NumCols()); colIdx++ {
			col := record.Column(colIdx)
			fieldBuilder := builder.Field(colIdx)

			// 根据列类型追加数据
			if err := appendColumn(fieldBuilder, col); err != nil {
				return nil, fmt.Errorf("failed to append column %d: %w", colIdx, err)
			}
		}
	}

	// 构建合并后的 record
	mergedRecord := builder.NewRecord()
	return mergedRecord, nil
}

// appendColumn 将源列的数据追加到目标 builder
func appendColumn(builder array.Builder, sourceCol arrow.Array) error {
	// 根据数据类型选择合适的追加方法
	switch builder := builder.(type) {
	case *array.Int64Builder:
		srcArray := sourceCol.(*array.Int64)
		for i := 0; i < srcArray.Len(); i++ {
			if srcArray.IsNull(i) {
				builder.AppendNull()
			} else {
				builder.Append(srcArray.Value(i))
			}
		}

	case *array.Float64Builder:
		srcArray := sourceCol.(*array.Float64)
		for i := 0; i < srcArray.Len(); i++ {
			if srcArray.IsNull(i) {
				builder.AppendNull()
			} else {
				builder.Append(srcArray.Value(i))
			}
		}

	case *array.StringBuilder:
		srcArray := sourceCol.(*array.String)
		for i := 0; i < srcArray.Len(); i++ {
			if srcArray.IsNull(i) {
				builder.AppendNull()
			} else {
				builder.Append(srcArray.Value(i))
			}
		}

	case *array.BooleanBuilder:
		srcArray := sourceCol.(*array.Boolean)
		for i := 0; i < srcArray.Len(); i++ {
			if srcArray.IsNull(i) {
				builder.AppendNull()
			} else {
				builder.Append(srcArray.Value(i))
			}
		}

	case *array.Int32Builder:
		srcArray := sourceCol.(*array.Int32)
		for i := 0; i < srcArray.Len(); i++ {
			if srcArray.IsNull(i) {
				builder.AppendNull()
			} else {
				builder.Append(srcArray.Value(i))
			}
		}

	case *array.Float32Builder:
		srcArray := sourceCol.(*array.Float32)
		for i := 0; i < srcArray.Len(); i++ {
			if srcArray.IsNull(i) {
				builder.AppendNull()
			} else {
				builder.Append(srcArray.Value(i))
			}
		}

	default:
		return fmt.Errorf("unsupported column type for merging: %T", builder)
	}

	return nil
}

// applyFilters 应用过滤条件到 Arrow Record (使用 Arrow Compute)
// 注意：这是一个基础实现，支持简单的比较操作
func applyFilters(record arrow.Record, filters []Filter) (arrow.Record, error) {
	if len(filters) == 0 {
		return record, nil
	}

	logger.Info("Applying filters to record",
		zap.Int("filter_count", len(filters)),
		zap.Int64("input_rows", record.NumRows()))

	// 构建过滤掩码
	mask := make([]bool, record.NumRows())
	for i := range mask {
		mask[i] = true
	}

	schema := record.Schema()

	// 应用每个过滤条件
	for _, filter := range filters {
		// 查找列索引
		colIdx := -1
		for i, field := range schema.Fields() {
			if field.Name == filter.Column {
				colIdx = i
				break
			}
		}

		if colIdx == -1 {
			logger.Warn("Filter column not found in schema",
				zap.String("column", filter.Column))
			continue
		}

		col := record.Column(colIdx)

		// 根据操作符和列类型应用过滤
		if err := applyColumnFilter(col, filter.Operator, filter.Value, mask); err != nil {
			return nil, fmt.Errorf("failed to apply filter on column %s: %w", filter.Column, err)
		}
	}

	// 根据掩码构建过滤后的 record
	filteredRecord, err := buildFilteredRecord(record, mask)
	if err != nil {
		return nil, fmt.Errorf("failed to build filtered record: %w", err)
	}

	logger.Info("Filters applied successfully",
		zap.Int64("output_rows", filteredRecord.NumRows()),
		zap.Int64("filtered_out", record.NumRows()-filteredRecord.NumRows()))

	return filteredRecord, nil
}

// applyColumnFilter 对单列应用过滤条件
func applyColumnFilter(col arrow.Array, operator string, value interface{}, mask []bool) error {
	switch col.DataType().ID() {
	case arrow.INT64:
		arr := col.(*array.Int64)
		targetVal, ok := value.(int64)
		if !ok {
			// 尝试类型转换
			if v, ok := value.(int); ok {
				targetVal = int64(v)
			} else if v, ok := value.(float64); ok {
				targetVal = int64(v)
			} else {
				return fmt.Errorf("cannot convert value %v to int64", value)
			}
		}
		return applyInt64Filter(arr, operator, targetVal, mask)

	case arrow.STRING:
		arr := col.(*array.String)
		targetVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("cannot convert value %v to string", value)
		}
		return applyStringFilter(arr, operator, targetVal, mask)

	case arrow.FLOAT64:
		arr := col.(*array.Float64)
		targetVal, ok := value.(float64)
		if !ok {
			if v, ok := value.(int); ok {
				targetVal = float64(v)
			} else if v, ok := value.(int64); ok {
				targetVal = float64(v)
			} else {
				return fmt.Errorf("cannot convert value %v to float64", value)
			}
		}
		return applyFloat64Filter(arr, operator, targetVal, mask)

	default:
		logger.Warn("Unsupported column type for filtering",
			zap.String("type", col.DataType().String()))
		return nil // 不应用过滤，保持所有行
	}
}

// applyInt64Filter 对 INT64 列应用过滤
func applyInt64Filter(arr *array.Int64, operator string, value int64, mask []bool) error {
	for i := 0; i < arr.Len(); i++ {
		if !mask[i] {
			continue // 已被过滤
		}

		if arr.IsNull(i) {
			mask[i] = false // NULL 值不匹配任何条件
			continue
		}

		val := arr.Value(i)
		match := false

		switch operator {
		case "=", "==":
			match = val == value
		case "!=", "<>":
			match = val != value
		case ">":
			match = val > value
		case ">=":
			match = val >= value
		case "<":
			match = val < value
		case "<=":
			match = val <= value
		default:
			return fmt.Errorf("unsupported operator: %s", operator)
		}

		mask[i] = mask[i] && match
	}

	return nil
}

// applyStringFilter 对 STRING 列应用过滤
func applyStringFilter(arr *array.String, operator string, value string, mask []bool) error {
	for i := 0; i < arr.Len(); i++ {
		if !mask[i] {
			continue
		}

		if arr.IsNull(i) {
			mask[i] = false
			continue
		}

		val := arr.Value(i)
		match := false

		switch operator {
		case "=", "==":
			match = val == value
		case "!=", "<>":
			match = val != value
		case ">":
			match = val > value
		case ">=":
			match = val >= value
		case "<":
			match = val < value
		case "<=":
			match = val <= value
		default:
			return fmt.Errorf("unsupported operator: %s", operator)
		}

		mask[i] = mask[i] && match
	}

	return nil
}

// applyFloat64Filter 对 FLOAT64 列应用过滤
func applyFloat64Filter(arr *array.Float64, operator string, value float64, mask []bool) error {
	for i := 0; i < arr.Len(); i++ {
		if !mask[i] {
			continue
		}

		if arr.IsNull(i) {
			mask[i] = false
			continue
		}

		val := arr.Value(i)
		match := false

		switch operator {
		case "=", "==":
			match = val == value
		case "!=", "<>":
			match = val != value
		case ">":
			match = val > value
		case ">=":
			match = val >= value
		case "<":
			match = val < value
		case "<=":
			match = val <= value
		default:
			return fmt.Errorf("unsupported operator: %s", operator)
		}

		mask[i] = mask[i] && match
	}

	return nil
}

// buildFilteredRecord 根据掩码构建过滤后的 record
func buildFilteredRecord(record arrow.Record, mask []bool) (arrow.Record, error) {
	// 计算过滤后的行数
	filteredRowCount := 0
	for _, keep := range mask {
		if keep {
			filteredRowCount++
		}
	}

	if filteredRowCount == 0 {
		// 返回空 record
		schema := record.Schema()
		pool := memory.NewGoAllocator()
		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()
		return builder.NewRecord(), nil
	}

	schema := record.Schema()
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 逐列复制过滤后的数据
	for colIdx := 0; colIdx < int(record.NumCols()); colIdx++ {
		col := record.Column(colIdx)
		fieldBuilder := builder.Field(colIdx)

		// 根据掩码复制数据
		if err := copyFilteredColumn(fieldBuilder, col, mask); err != nil {
			return nil, fmt.Errorf("failed to copy filtered column %d: %w", colIdx, err)
		}
	}

	return builder.NewRecord(), nil
}

// copyFilteredColumn 根据掩码复制列数据
func copyFilteredColumn(builder array.Builder, sourceCol arrow.Array, mask []bool) error {
	switch builder := builder.(type) {
	case *array.Int64Builder:
		srcArray := sourceCol.(*array.Int64)
		for i := 0; i < srcArray.Len(); i++ {
			if mask[i] {
				if srcArray.IsNull(i) {
					builder.AppendNull()
				} else {
					builder.Append(srcArray.Value(i))
				}
			}
		}

	case *array.Float64Builder:
		srcArray := sourceCol.(*array.Float64)
		for i := 0; i < srcArray.Len(); i++ {
			if mask[i] {
				if srcArray.IsNull(i) {
					builder.AppendNull()
				} else {
					builder.Append(srcArray.Value(i))
				}
			}
		}

	case *array.StringBuilder:
		srcArray := sourceCol.(*array.String)
		for i := 0; i < srcArray.Len(); i++ {
			if mask[i] {
				if srcArray.IsNull(i) {
					builder.AppendNull()
				} else {
					builder.Append(srcArray.Value(i))
				}
			}
		}

	case *array.Int32Builder:
		srcArray := sourceCol.(*array.Int32)
		for i := 0; i < srcArray.Len(); i++ {
			if mask[i] {
				if srcArray.IsNull(i) {
					builder.AppendNull()
				} else {
					builder.Append(srcArray.Value(i))
				}
			}
		}

	case *array.Float32Builder:
		srcArray := sourceCol.(*array.Float32)
		for i := 0; i < srcArray.Len(); i++ {
			if mask[i] {
				if srcArray.IsNull(i) {
					builder.AppendNull()
				} else {
					builder.Append(srcArray.Value(i))
				}
			}
		}

	default:
		return fmt.Errorf("unsupported column type for filtering: %T", builder)
	}

	return nil
}

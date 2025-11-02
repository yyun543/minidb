package parquet

import (
	"context"
	"fmt"
	"os"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/apache/arrow/go/v18/parquet"
	"github.com/apache/arrow/go/v18/parquet/file"
	"github.com/apache/arrow/go/v18/parquet/metadata"
	"github.com/apache/arrow/go/v18/parquet/pqarrow"
	"github.com/yyun543/minidb/internal/logger"
	"go.uber.org/zap"
)

// Filter 查询过滤条件 (避免循环依赖)
type Filter struct {
	Column   string        // 列名
	Operator string        // =, >, <, >=, <=, !=, LIKE, IN, BETWEEN
	Value    interface{}   // 单值操作符使用
	Values   []interface{} // IN 操作符使用多个值
}

// RowGroupStats Row Group 统计信息
type RowGroupStats struct {
	RowGroupIndex int
	NumRows       int64
	MinValues     map[string]interface{}
	MaxValues     map[string]interface{}
}

// ReadParquetFile 读取 Parquet 文件并返回 Arrow Record (使用 Arrow 原生 reader)
// P1.4: 增强为支持 Row Group 级别的数据跳过 (Data Skipping)
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

	// 创建 Arrow file reader (需要在 Row Group 检查之前创建)
	arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	if err != nil {
		return nil, fmt.Errorf("failed to create arrow file reader: %w", err)
	}

	// P1.4: Row Group 过滤 (基于 Min/Max 统计信息)
	numRowGroups := reader.NumRowGroups()
	var rowGroupsToRead []int

	logger.Debug("Parquet file Row Group information",
		zap.String("path", path),
		zap.Int("num_row_groups", numRowGroups),
		zap.Int("filter_count", len(filters)))

	if len(filters) > 0 && numRowGroups > 0 {
		// 检查每个 Row Group 的统计信息
		for rgIdx := 0; rgIdx < numRowGroups; rgIdx++ {
			rgReader := reader.RowGroup(rgIdx)
			stats := extractRowGroupStats(rgIdx, rgReader)

			// 使用统计信息判断是否可以跳过此 Row Group
			if !canSkipRowGroup(stats, filters) {
				rowGroupsToRead = append(rowGroupsToRead, rgIdx)
			} else {
				logger.Debug("Skipping Row Group based on statistics",
					zap.String("path", path),
					zap.Int("row_group", rgIdx),
					zap.Int64("rows", stats.NumRows))
			}
		}

		// 如果所有 Row Group 都被跳过，返回空结果
		if len(rowGroupsToRead) == 0 {
			logger.Info("All Row Groups skipped by predicate pushdown",
				zap.String("path", path),
				zap.Int("total_row_groups", numRowGroups))

			// 创建空 Record
			arrowReader, err := pqarrow.NewFileReader(reader, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
			if err != nil {
				return nil, fmt.Errorf("failed to create arrow file reader: %w", err)
			}
			schema, err := arrowReader.Schema()
			if err != nil {
				return nil, fmt.Errorf("failed to get schema: %w", err)
			}

			pool := memory.NewGoAllocator()
			builder := array.NewRecordBuilder(pool, schema)
			defer builder.Release()
			return builder.NewRecord(), nil
		}

		logger.Info("Row Group filtering results",
			zap.String("path", path),
			zap.Int("total_row_groups", numRowGroups),
			zap.Int("row_groups_to_read", len(rowGroupsToRead)),
			zap.Int("row_groups_skipped", numRowGroups-len(rowGroupsToRead)))
	} else if numRowGroups > 0 {
		// 无过滤条件，读取所有 Row Groups
		for i := 0; i < numRowGroups; i++ {
			rowGroupsToRead = append(rowGroupsToRead, i)
		}
		logger.Debug("No filters, reading all Row Groups",
			zap.String("path", path),
			zap.Int("num_row_groups", numRowGroups))
	} else {
		// 文件没有 Row Groups，可能是空文件或者单个小文件
		logger.Debug("No Row Groups found in file, returning empty result",
			zap.String("path", path))

		schema, err := arrowReader.Schema()
		if err != nil {
			return nil, fmt.Errorf("failed to get schema: %w", err)
		}
		pool := memory.NewGoAllocator()
		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()
		return builder.NewRecord(), nil
	}

	// 如果没有 Row Groups 要读取，返回空结果
	if len(rowGroupsToRead) == 0 {
		logger.Info("No Row Groups selected after filtering",
			zap.String("path", path))

		schema, err := arrowReader.Schema()
		if err != nil {
			return nil, fmt.Errorf("failed to get schema: %w", err)
		}
		pool := memory.NewGoAllocator()
		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()
		return builder.NewRecord(), nil
	}

	// P1.4: Row Group 过滤
	// 如果启用了过滤且能跳过部分 Row Groups，使用 GetRecordReader
	// 否则使用原始的 ReadTable 方法（更稳定）
	ctx := context.Background()
	var records []arrow.Record

	useRowGroupFiltering := len(filters) > 0 && len(rowGroupsToRead) > 0 && len(rowGroupsToRead) < numRowGroups

	if useRowGroupFiltering {
		// 使用 Row Group 过滤
		logger.Info("Using Row Group filtering",
			zap.String("path", path),
			zap.Int("total_row_groups", numRowGroups),
			zap.Int("row_groups_to_read", len(rowGroupsToRead)),
			zap.Int("row_groups_skipped", numRowGroups-len(rowGroupsToRead)))

		rgReader, err := arrowReader.GetRecordReader(ctx, nil, rowGroupsToRead)
		if err != nil {
			return nil, fmt.Errorf("failed to get record reader for row groups: %w", err)
		}
		defer rgReader.Release()

		for rgReader.Next() {
			rec := rgReader.Record()
			rec.Retain()
			records = append(records, rec)
		}

		if err := rgReader.Err(); err != nil {
			for _, rec := range records {
				rec.Release()
			}
			return nil, fmt.Errorf("failed to read row groups: %w", err)
		}
	} else {
		// 使用原始的 ReadTable 方法（兼容性更好）
		table, err := arrowReader.ReadTable(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to read table: %w", err)
		}
		defer table.Release()

		if table.NumRows() == 0 {
			// 返回空 Record
			schema := table.Schema()
			pool := memory.NewGoAllocator()
			builder := array.NewRecordBuilder(pool, schema)
			defer builder.Release()
			return builder.NewRecord(), nil
		}

		// 使用 TableReader 来获取 Records
		tr := array.NewTableReader(table, table.NumRows())
		defer tr.Release()

		for tr.Next() {
			rec := tr.Record()
			rec.Retain()
			records = append(records, rec)
		}

		if err := tr.Err(); err != nil {
			for _, rec := range records {
				rec.Release()
			}
			return nil, fmt.Errorf("failed to read records from table: %w", err)
		}
	}

	// 如果没有读取到任何记录，返回空结果
	if len(records) == 0 {
		schema, err := arrowReader.Schema()
		if err != nil {
			return nil, fmt.Errorf("failed to get schema: %w", err)
		}
		pool := memory.NewGoAllocator()
		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()
		record := builder.NewRecord()

		logger.Info("Parquet file read (empty)",
			zap.String("path", path),
			zap.Int64("rows", 0))

		return record, nil
	}

	// 合并多个 Row Groups 的记录
	var mergedRecord arrow.Record
	if len(records) == 1 {
		mergedRecord = records[0]
	} else {
		logger.Info("Merging multiple Row Group records",
			zap.String("path", path),
			zap.Int("record_count", len(records)))

		mergedRecord, err = mergeRecords(records)
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
	}

	logger.Info("Parquet file read with Row Group filtering",
		zap.String("path", path),
		zap.Int64("rows", mergedRecord.NumRows()),
		zap.Int("row_groups_read", len(rowGroupsToRead)),
		zap.Int("row_groups_skipped", numRowGroups-len(rowGroupsToRead)))

	// P1.4: Row Group filtering 已经大幅减少了数据读取
	// 但仍需要应用 record 级别的过滤以处理 Row Group 内的不匹配行
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

// extractRowGroupStats 从 Row Group 元数据中提取统计信息
func extractRowGroupStats(rgIdx int, rgReader *file.RowGroupReader) *RowGroupStats {
	stats := &RowGroupStats{
		RowGroupIndex: rgIdx,
		NumRows:       rgReader.NumRows(),
		MinValues:     make(map[string]interface{}),
		MaxValues:     make(map[string]interface{}),
	}

	// 获取 Row Group 元数据
	rgMeta := rgReader.MetaData()
	if rgMeta == nil {
		return stats
	}

	// 遍历所有列，提取 Min/Max 统计信息
	for colIdx := 0; colIdx < rgMeta.NumColumns(); colIdx++ {
		colChunkMeta, err := rgMeta.ColumnChunk(colIdx)
		if err != nil {
			continue
		}

		// 检查是否有统计信息
		statsSet, err := colChunkMeta.StatsSet()
		if err != nil || !statsSet {
			continue
		}

		colStats, err := colChunkMeta.Statistics()
		if err != nil {
			continue
		}

		// 获取列路径（列名）
		colPath := colChunkMeta.PathInSchema()
		colName := colPath.String()

		// 根据列类型提取 Min/Max
		switch colChunkMeta.Type() {
		case parquet.Types.Int64:
			if typedStats, ok := colStats.(*metadata.Int64Statistics); ok && typedStats.HasMinMax() {
				stats.MinValues[colName] = typedStats.Min()
				stats.MaxValues[colName] = typedStats.Max()
			}
		case parquet.Types.Int32:
			if typedStats, ok := colStats.(*metadata.Int32Statistics); ok && typedStats.HasMinMax() {
				stats.MinValues[colName] = typedStats.Min()
				stats.MaxValues[colName] = typedStats.Max()
			}
		case parquet.Types.Float:
			if typedStats, ok := colStats.(*metadata.Float32Statistics); ok && typedStats.HasMinMax() {
				stats.MinValues[colName] = typedStats.Min()
				stats.MaxValues[colName] = typedStats.Max()
			}
		case parquet.Types.Double:
			if typedStats, ok := colStats.(*metadata.Float64Statistics); ok && typedStats.HasMinMax() {
				stats.MinValues[colName] = typedStats.Min()
				stats.MaxValues[colName] = typedStats.Max()
			}
		case parquet.Types.ByteArray:
			if typedStats, ok := colStats.(*metadata.ByteArrayStatistics); ok && typedStats.HasMinMax() {
				// 字符串类型
				stats.MinValues[colName] = string(typedStats.Min().Bytes())
				stats.MaxValues[colName] = string(typedStats.Max().Bytes())
			}
		}
	}

	return stats
}

// canSkipRowGroup 判断是否可以跳过某个 Row Group
// 返回 true 表示可以跳过（Row Group 不包含任何匹配数据）
// 返回 false 表示不能跳过（Row Group 可能包含匹配数据）
func canSkipRowGroup(stats *RowGroupStats, filters []Filter) bool {
	for _, filter := range filters {
		// 如果该列没有统计信息，保守起见不跳过
		minVal, hasMin := stats.MinValues[filter.Column]
		maxVal, hasMax := stats.MaxValues[filter.Column]

		if !hasMin || !hasMax {
			return false
		}

		// 根据操作符判断是否可以跳过
		switch filter.Operator {
		case "=", "==":
			// 等值查询：如果 value 不在 [min, max] 范围内，可以跳过
			if canSkipEquality(minVal, maxVal, filter.Value) {
				return true
			}

		case ">":
			// 大于查询：如果 max <= value，可以跳过
			if canSkipGreaterThan(maxVal, filter.Value) {
				return true
			}

		case ">=":
			// 大于等于查询：如果 max < value，可以跳过
			if canSkipGreaterThanOrEqual(maxVal, filter.Value) {
				return true
			}

		case "<":
			// 小于查询：如果 min >= value，可以跳过
			if canSkipLessThan(minVal, filter.Value) {
				return true
			}

		case "<=":
			// 小于等于查询：如果 min > value，可以跳过
			if canSkipLessThanOrEqual(minVal, filter.Value) {
				return true
			}

		case "!=", "<>":
			// 不等于查询：如果 min == max == value，可以跳过
			if canSkipNotEqual(minVal, maxVal, filter.Value) {
				return true
			}

		case "IN":
			// IN 查询：如果所有 values 都不在 [min, max] 范围内，可以跳过
			if canSkipIN(minVal, maxVal, filter.Values) {
				return true
			}
		}
	}

	// 没有找到可以跳过的理由，不跳过
	return false
}

// 辅助函数：比较函数

func canSkipEquality(min, max, value interface{}) bool {
	switch min := min.(type) {
	case int64:
		maxInt := max.(int64)
		valInt := convertToInt64(value)
		return valInt < min || valInt > maxInt

	case int32:
		maxInt := max.(int32)
		valInt := int32(convertToInt64(value))
		return valInt < min || valInt > maxInt

	case float64:
		maxFloat := max.(float64)
		valFloat := convertToFloat64(value)
		return valFloat < min || valFloat > maxFloat

	case float32:
		maxFloat := max.(float32)
		valFloat := float32(convertToFloat64(value))
		return valFloat < min || valFloat > maxFloat

	case string:
		maxStr := max.(string)
		valStr := value.(string)
		return valStr < min || valStr > maxStr
	}
	return false
}

func canSkipGreaterThan(max, value interface{}) bool {
	switch max := max.(type) {
	case int64:
		return max <= convertToInt64(value)
	case int32:
		return max <= int32(convertToInt64(value))
	case float64:
		return max <= convertToFloat64(value)
	case float32:
		return max <= float32(convertToFloat64(value))
	case string:
		return max <= value.(string)
	}
	return false
}

func canSkipGreaterThanOrEqual(max, value interface{}) bool {
	switch max := max.(type) {
	case int64:
		return max < convertToInt64(value)
	case int32:
		return max < int32(convertToInt64(value))
	case float64:
		return max < convertToFloat64(value)
	case float32:
		return max < float32(convertToFloat64(value))
	case string:
		return max < value.(string)
	}
	return false
}

func canSkipLessThan(min, value interface{}) bool {
	switch min := min.(type) {
	case int64:
		return min >= convertToInt64(value)
	case int32:
		return min >= int32(convertToInt64(value))
	case float64:
		return min >= convertToFloat64(value)
	case float32:
		return min >= float32(convertToFloat64(value))
	case string:
		return min >= value.(string)
	}
	return false
}

func canSkipLessThanOrEqual(min, value interface{}) bool {
	switch min := min.(type) {
	case int64:
		return min > convertToInt64(value)
	case int32:
		return min > int32(convertToInt64(value))
	case float64:
		return min > convertToFloat64(value)
	case float32:
		return min > float32(convertToFloat64(value))
	case string:
		return min > value.(string)
	}
	return false
}

func canSkipNotEqual(min, max, value interface{}) bool {
	// 只有当 min == max == value 时才能跳过
	switch min := min.(type) {
	case int64:
		maxInt := max.(int64)
		valInt := convertToInt64(value)
		return min == maxInt && min == valInt

	case int32:
		maxInt := max.(int32)
		valInt := int32(convertToInt64(value))
		return min == maxInt && min == valInt

	case float64:
		maxFloat := max.(float64)
		valFloat := convertToFloat64(value)
		return min == maxFloat && min == valFloat

	case float32:
		maxFloat := max.(float32)
		valFloat := float32(convertToFloat64(value))
		return min == maxFloat && min == valFloat

	case string:
		maxStr := max.(string)
		valStr := value.(string)
		return min == maxStr && min == valStr
	}
	return false
}

func canSkipIN(min, max interface{}, values []interface{}) bool {
	// 如果所有 IN 值都不在 [min, max] 范围内，可以跳过
	for _, value := range values {
		if !canSkipEquality(min, max, value) {
			// 至少有一个值在范围内，不能跳过
			return false
		}
	}
	// 所有值都不在范围内，可以跳过
	return true
}

func convertToInt64(value interface{}) int64 {
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	default:
		return 0
	}
}

func convertToFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int64:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	default:
		return 0.0
	}
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
		if err := applyColumnFilter(col, filter.Operator, filter.Value, filter.Values, mask); err != nil {
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
func applyColumnFilter(col arrow.Array, operator string, value interface{}, values []interface{}, mask []bool) error {
	switch col.DataType().ID() {
	case arrow.INT64:
		arr := col.(*array.Int64)

		// Handle IN operator with multiple values
		if operator == "IN" {
			targetVals := make([]int64, 0, len(values))
			for _, v := range values {
				if val, ok := v.(int64); ok {
					targetVals = append(targetVals, val)
				} else if val, ok := v.(int); ok {
					targetVals = append(targetVals, int64(val))
				} else if val, ok := v.(float64); ok {
					targetVals = append(targetVals, int64(val))
				} else {
					return fmt.Errorf("cannot convert IN value %v to int64", v)
				}
			}
			return applyInt64FilterIN(arr, targetVals, mask)
		}

		// Handle single value operators
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

		// Handle IN operator with multiple values
		if operator == "IN" {
			targetVals := make([]string, 0, len(values))
			for _, v := range values {
				if val, ok := v.(string); ok {
					targetVals = append(targetVals, val)
				} else {
					return fmt.Errorf("cannot convert IN value %v to string", v)
				}
			}
			return applyStringFilterIN(arr, targetVals, mask)
		}

		// Handle single value operators
		targetVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("cannot convert value %v to string", value)
		}
		return applyStringFilter(arr, operator, targetVal, mask)

	case arrow.FLOAT64:
		arr := col.(*array.Float64)

		// Handle IN operator with multiple values
		if operator == "IN" {
			targetVals := make([]float64, 0, len(values))
			for _, v := range values {
				if val, ok := v.(float64); ok {
					targetVals = append(targetVals, val)
				} else if val, ok := v.(int); ok {
					targetVals = append(targetVals, float64(val))
				} else if val, ok := v.(int64); ok {
					targetVals = append(targetVals, float64(val))
				} else {
					return fmt.Errorf("cannot convert IN value %v to float64", v)
				}
			}
			return applyFloat64FilterIN(arr, targetVals, mask)
		}

		// Handle single value operators
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

	case arrow.BOOL:
		arr := col.(*array.Boolean)

		// Convert value to bool
		var targetVal bool
		if v, ok := value.(bool); ok {
			targetVal = v
		} else if v, ok := value.(int64); ok {
			targetVal = (v != 0) // 0 = false, non-zero = true
		} else if v, ok := value.(int); ok {
			targetVal = (v != 0)
		} else {
			return fmt.Errorf("cannot convert value %v to bool", value)
		}

		return applyBoolFilter(arr, operator, targetVal, mask)

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

// applyBoolFilter applies filter to BOOLEAN column
func applyBoolFilter(arr *array.Boolean, operator string, value bool, mask []bool) error {
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
		default:
			return fmt.Errorf("unsupported operator for bool: %s", operator)
		}

		mask[i] = mask[i] && match
	}

	return nil
}

// applyInt64FilterIN applies IN filter to INT64 column with multiple values
func applyInt64FilterIN(arr *array.Int64, values []int64, mask []bool) error {
	// Create a map for O(1) lookup
	valuesSet := make(map[int64]bool)
	for _, v := range values {
		valuesSet[v] = true
	}

	for i := 0; i < arr.Len(); i++ {
		if !mask[i] {
			continue
		}

		if arr.IsNull(i) {
			mask[i] = false
			continue
		}

		val := arr.Value(i)
		mask[i] = mask[i] && valuesSet[val]
	}

	return nil
}

// applyStringFilterIN applies IN filter to STRING column with multiple values
func applyStringFilterIN(arr *array.String, values []string, mask []bool) error {
	// Create a map for O(1) lookup
	valuesSet := make(map[string]bool)
	for _, v := range values {
		valuesSet[v] = true
	}

	for i := 0; i < arr.Len(); i++ {
		if !mask[i] {
			continue
		}

		if arr.IsNull(i) {
			mask[i] = false
			continue
		}

		val := arr.Value(i)
		mask[i] = mask[i] && valuesSet[val]
	}

	return nil
}

// applyFloat64FilterIN applies IN filter to FLOAT64 column with multiple values
func applyFloat64FilterIN(arr *array.Float64, values []float64, mask []bool) error {
	// Create a map for O(1) lookup
	valuesSet := make(map[float64]bool)
	for _, v := range values {
		valuesSet[v] = true
	}

	for i := 0; i < arr.Len(); i++ {
		if !mask[i] {
			continue
		}

		if arr.IsNull(i) {
			mask[i] = false
			continue
		}

		val := arr.Value(i)
		mask[i] = mask[i] && valuesSet[val]
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

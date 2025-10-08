package parquet

import (
	"fmt"
	"os"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/parquet/pqarrow"
	"github.com/yyun543/minidb/internal/delta"
	"github.com/yyun543/minidb/internal/logger"
	"go.uber.org/zap"
)

// WriteArrowBatch 将 Arrow Batch 写入 Parquet 文件 (使用 Arrow 原生 Parquet writer)
func WriteArrowBatch(path string, batch arrow.Record) (*delta.FileStats, error) {
	logger.Info("Writing Arrow batch to Parquet",
		zap.String("path", path),
		zap.Int64("rows", batch.NumRows()))

	// 确保父目录存在
	if err := ensureDir(path); err != nil {
		return nil, err
	}

	// 收集统计信息
	stats := collectStats(batch)

	// 创建文件
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet file: %w", err)
	}

	// 使用 Arrow 的原生 Parquet writer
	writer, err := pqarrow.NewFileWriter(
		batch.Schema(),
		file,
		nil, // 使用默认的 Parquet 属性
		pqarrow.DefaultWriterProps(),
	)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to create arrow parquet writer: %w", err)
	}

	// 写入 Arrow Record
	if err := writer.Write(batch); err != nil {
		writer.Close()
		file.Close()
		return nil, fmt.Errorf("failed to write arrow record to parquet: %w", err)
	}

	// 关闭 writer (会自动写入 footer 并关闭底层文件)
	if err := writer.Close(); err != nil {
		// writer.Close() 可能已经关闭了文件，所以不再调用 file.Close()
		return nil, fmt.Errorf("failed to close parquet writer: %w", err)
	}

	// 获取文件大小（writer 已经关闭文件，使用 os.Stat）
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat parquet file: %w", err)
	}

	stats.FileSize = fileInfo.Size()

	logger.Info("Parquet file written successfully",
		zap.String("path", path),
		zap.Int64("size", stats.FileSize),
		zap.Int64("rows", stats.RowCount))

	return stats, nil
}

// collectStats 收集列统计信息
func collectStats(batch arrow.Record) *delta.FileStats {
	stats := &delta.FileStats{
		RowCount:   batch.NumRows(),
		MinValues:  make(map[string]interface{}),
		MaxValues:  make(map[string]interface{}),
		NullCounts: make(map[string]int64),
	}

	schema := batch.Schema()

	for i := 0; i < int(batch.NumCols()); i++ {
		col := batch.Column(i)
		field := schema.Field(i)

		// 计算 null count
		stats.NullCounts[field.Name] = int64(col.NullN())

		// 计算 min/max (支持所有常用数据类型)
		switch col.DataType().ID() {
		case arrow.INT64:
			arr := col.(*array.Int64)
			if arr.Len() > 0 && arr.Len() > arr.NullN() {
				min, max := computeInt64MinMax(arr)
				stats.MinValues[field.Name] = min
				stats.MaxValues[field.Name] = max
			}

		case arrow.INT32:
			arr := col.(*array.Int32)
			if arr.Len() > 0 && arr.Len() > arr.NullN() {
				min, max := computeInt32MinMax(arr)
				stats.MinValues[field.Name] = min
				stats.MaxValues[field.Name] = max
			}

		case arrow.INT16:
			arr := col.(*array.Int16)
			if arr.Len() > 0 && arr.Len() > arr.NullN() {
				min, max := computeInt16MinMax(arr)
				stats.MinValues[field.Name] = min
				stats.MaxValues[field.Name] = max
			}

		case arrow.INT8:
			arr := col.(*array.Int8)
			if arr.Len() > 0 && arr.Len() > arr.NullN() {
				min, max := computeInt8MinMax(arr)
				stats.MinValues[field.Name] = min
				stats.MaxValues[field.Name] = max
			}

		case arrow.FLOAT64:
			arr := col.(*array.Float64)
			if arr.Len() > 0 && arr.Len() > arr.NullN() {
				min, max := computeFloat64MinMax(arr)
				stats.MinValues[field.Name] = min
				stats.MaxValues[field.Name] = max
			}

		case arrow.FLOAT32:
			arr := col.(*array.Float32)
			if arr.Len() > 0 && arr.Len() > arr.NullN() {
				min, max := computeFloat32MinMax(arr)
				stats.MinValues[field.Name] = min
				stats.MaxValues[field.Name] = max
			}

		case arrow.STRING:
			arr := col.(*array.String)
			if arr.Len() > 0 && arr.Len() > arr.NullN() {
				min, max := computeStringMinMax(arr)
				stats.MinValues[field.Name] = min
				stats.MaxValues[field.Name] = max
			}

		case arrow.BOOL:
			arr := col.(*array.Boolean)
			if arr.Len() > 0 && arr.Len() > arr.NullN() {
				min, max := computeBooleanMinMax(arr)
				stats.MinValues[field.Name] = min
				stats.MaxValues[field.Name] = max
			}

		case arrow.DATE32, arrow.DATE64:
			// 日期类型统计
			if col.Len() > 0 && col.Len() > col.NullN() {
				stats.MinValues[field.Name] = "date_min"
				stats.MaxValues[field.Name] = "date_max"
			}

		case arrow.TIMESTAMP:
			// 时间戳类型统计
			if col.Len() > 0 && col.Len() > col.NullN() {
				stats.MinValues[field.Name] = "timestamp_min"
				stats.MaxValues[field.Name] = "timestamp_max"
			}

		default:
			// 对于不支持的类型，记录类型信息
			logger.Debug("Skipping min/max stats for unsupported type",
				zap.String("column", field.Name),
				zap.String("type", col.DataType().String()))
		}
	}

	return stats
}

// Helper functions

func computeInt64MinMax(arr *array.Int64) (int64, int64) {
	if arr.Len() == 0 {
		return 0, 0
	}

	min := arr.Value(0)
	max := arr.Value(0)

	for i := 1; i < arr.Len(); i++ {
		if arr.IsNull(i) {
			continue
		}

		val := arr.Value(i)
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
	}

	return min, max
}

func computeInt32MinMax(arr *array.Int32) (int32, int32) {
	if arr.Len() == 0 {
		return 0, 0
	}

	var min, max int32
	initialized := false

	for i := 0; i < arr.Len(); i++ {
		if arr.IsNull(i) {
			continue
		}

		val := arr.Value(i)
		if !initialized {
			min = val
			max = val
			initialized = true
		} else {
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
		}
	}

	return min, max
}

func computeInt16MinMax(arr *array.Int16) (int16, int16) {
	if arr.Len() == 0 {
		return 0, 0
	}

	var min, max int16
	initialized := false

	for i := 0; i < arr.Len(); i++ {
		if arr.IsNull(i) {
			continue
		}

		val := arr.Value(i)
		if !initialized {
			min = val
			max = val
			initialized = true
		} else {
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
		}
	}

	return min, max
}

func computeInt8MinMax(arr *array.Int8) (int8, int8) {
	if arr.Len() == 0 {
		return 0, 0
	}

	var min, max int8
	initialized := false

	for i := 0; i < arr.Len(); i++ {
		if arr.IsNull(i) {
			continue
		}

		val := arr.Value(i)
		if !initialized {
			min = val
			max = val
			initialized = true
		} else {
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
		}
	}

	return min, max
}

func computeFloat64MinMax(arr *array.Float64) (float64, float64) {
	if arr.Len() == 0 {
		return 0, 0
	}

	var min, max float64
	initialized := false

	for i := 0; i < arr.Len(); i++ {
		if arr.IsNull(i) {
			continue
		}

		val := arr.Value(i)
		if !initialized {
			min = val
			max = val
			initialized = true
		} else {
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
		}
	}

	return min, max
}

func computeFloat32MinMax(arr *array.Float32) (float32, float32) {
	if arr.Len() == 0 {
		return 0, 0
	}

	var min, max float32
	initialized := false

	for i := 0; i < arr.Len(); i++ {
		if arr.IsNull(i) {
			continue
		}

		val := arr.Value(i)
		if !initialized {
			min = val
			max = val
			initialized = true
		} else {
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
		}
	}

	return min, max
}

func computeStringMinMax(arr *array.String) (string, string) {
	if arr.Len() == 0 {
		return "", ""
	}

	var min, max string
	initialized := false

	for i := 0; i < arr.Len(); i++ {
		if arr.IsNull(i) {
			continue
		}

		val := arr.Value(i)
		if !initialized {
			min = val
			max = val
			initialized = true
		} else {
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
		}
	}

	return min, max
}

func computeBooleanMinMax(arr *array.Boolean) (bool, bool) {
	if arr.Len() == 0 {
		return false, false
	}

	hasTrue := false
	hasFalse := false

	for i := 0; i < arr.Len(); i++ {
		if arr.IsNull(i) {
			continue
		}

		if arr.Value(i) {
			hasTrue = true
		} else {
			hasFalse = true
		}

		// 一旦找到true和false，就可以确定min/max了
		if hasTrue && hasFalse {
			break
		}
	}

	// Boolean min/max: false < true
	min := hasFalse
	max := hasTrue

	return min, max
}

func ensureDir(path string) error {
	// 获取文件所在的目录路径
	dir := path
	lastSlash := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			lastSlash = i
			break
		}
	}

	if lastSlash > 0 {
		dir = path[:lastSlash]
	} else {
		// 如果没有目录分隔符，说明文件在当前目录
		return nil
	}

	// 创建目录（如果不存在）
	return os.MkdirAll(dir, 0755)
}

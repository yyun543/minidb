package statistics

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/yyun543/minidb/internal/types"
)

// TableStatistics 表级统计信息
type TableStatistics struct {
	TableName   string                       // 表名
	RowCount    int64                        // 总行数
	DataSize    int64                        // 数据大小（字节）
	LastUpdated time.Time                    // 最后更新时间
	ColumnStats map[string]*ColumnStatistics // 列统计信息
	IndexStats  map[string]*IndexStatistics  // 索引统计信息
}

// ColumnStatistics 列级统计信息
type ColumnStatistics struct {
	ColumnName       string         // 列名
	DataType         types.DataType // 数据类型
	NullCount        int64          // NULL值数量
	DistinctCount    int64          // 唯一值数量
	MinValue         interface{}    // 最小值
	MaxValue         interface{}    // 最大值
	MostCommonValues []MCV          // 最常见值
	Histogram        *Histogram     // 直方图
	LastUpdated      time.Time      // 最后更新时间
}

// MCV 最常见值 (Most Common Value)
type MCV struct {
	Value     interface{} // 值
	Frequency float64     // 频率
}

// IndexStatistics 索引统计信息
type IndexStatistics struct {
	IndexName    string    // 索引名
	KeyColumns   []string  // 索引键列
	UniqueValues int64     // 唯一值数量
	IndexSize    int64     // 索引大小
	LastUpdated  time.Time // 最后更新时间
}

// Histogram 直方图
type Histogram struct {
	Buckets    []HistogramBucket // 桶
	TotalCount int64             // 总数
}

// HistogramBucket 直方图桶
type HistogramBucket struct {
	LowerBound interface{} // 下界
	UpperBound interface{} // 上界
	Count      int64       // 计数
	Frequency  float64     // 频率
}

// StatisticsManager 统计信息管理器
type StatisticsManager struct {
	tableStats     map[string]*TableStatistics // 表统计信息
	mu             sync.RWMutex                // 读写锁
	autoUpdate     bool                        // 是否自动更新
	updateInterval time.Duration               // 更新间隔
}

// NewStatisticsManager 创建统计信息管理器
func NewStatisticsManager() *StatisticsManager {
	return &StatisticsManager{
		tableStats:     make(map[string]*TableStatistics),
		autoUpdate:     true,
		updateInterval: 5 * time.Minute, // 默认5分钟更新一次
	}
}

// GetTableStatistics 获取表统计信息
func (sm *StatisticsManager) GetTableStatistics(tableName string) (*TableStatistics, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if stats, exists := sm.tableStats[tableName]; exists {
		return stats, nil
	}
	return nil, fmt.Errorf("statistics not found for table %s", tableName)
}

// UpdateTableStatistics 更新表统计信息
func (sm *StatisticsManager) UpdateTableStatistics(tableName string, schema *types.TableSchema, batches []*types.Batch) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	stats := &TableStatistics{
		TableName:   tableName,
		LastUpdated: time.Now(),
		ColumnStats: make(map[string]*ColumnStatistics),
		IndexStats:  make(map[string]*IndexStatistics),
	}

	// 收集基本统计信息
	totalRows := int64(0)
	totalSize := int64(0)

	// 为每列创建统计收集器
	columnCollectors := make(map[string]*ColumnStatsCollector)
	for _, col := range schema.Columns {
		collector := NewColumnStatsCollector(col.Name, col.Type)
		columnCollectors[col.Name] = collector
	}

	// 遍历所有批次收集统计信息
	for _, batch := range batches {
		if batch == nil {
			continue
		}

		record := batch.Record()
		if record == nil {
			continue
		}

		batchRows := record.NumRows()
		totalRows += batchRows
		totalSize += estimateRecordSize(record)

		// 收集列级统计信息
		for i := int64(0); i < record.NumCols(); i++ {
			column := record.Column(int(i))
			fieldName := record.Schema().Field(int(i)).Name

			if collector, exists := columnCollectors[fieldName]; exists {
				collector.ProcessArray(column)
			}
		}
	}

	// 完成统计信息收集
	stats.RowCount = totalRows
	stats.DataSize = totalSize

	for name, collector := range columnCollectors {
		colStats := collector.Finalize()
		stats.ColumnStats[name] = colStats
	}

	sm.tableStats[tableName] = stats
	return nil
}

// EstimateSelectivity 估算选择性（为优化器提供基础）
func (sm *StatisticsManager) EstimateSelectivity(tableName, columnName, operator string, value interface{}) (float64, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	tableStats, exists := sm.tableStats[tableName]
	if !exists {
		return 0.5, nil // 默认选择性为50%
	}

	colStats, exists := tableStats.ColumnStats[columnName]
	if !exists {
		return 0.5, nil
	}

	switch operator {
	case "=":
		return sm.estimateEqualitySelectivity(colStats, value), nil
	case "<", "<=":
		return sm.estimateRangeSelectivity(colStats, value, true), nil
	case ">", ">=":
		return sm.estimateRangeSelectivity(colStats, value, false), nil
	case "!=", "<>":
		eq := sm.estimateEqualitySelectivity(colStats, value)
		return 1.0 - eq, nil
	default:
		return 0.5, nil
	}
}

// EstimateJoinSelectivity 估算连接选择性
func (sm *StatisticsManager) EstimateJoinSelectivity(leftTable, leftColumn, rightTable, rightColumn string) (float64, error) {
	leftStats, err := sm.GetTableStatistics(leftTable)
	if err != nil {
		return 0.1, nil // 默认连接选择性为10%
	}

	rightStats, err := sm.GetTableStatistics(rightTable)
	if err != nil {
		return 0.1, nil
	}

	leftColStats := leftStats.ColumnStats[leftColumn]
	rightColStats := rightStats.ColumnStats[rightColumn]

	if leftColStats == nil || rightColStats == nil {
		return 0.1, nil
	}

	// 简化的连接选择性估算
	// 选择性 ≈ 1 / max(distinctCount1, distinctCount2)
	maxDistinct := math.Max(float64(leftColStats.DistinctCount), float64(rightColStats.DistinctCount))
	if maxDistinct <= 0 {
		return 0.1, nil
	}

	selectivity := 1.0 / maxDistinct

	// 限制在合理范围内
	if selectivity > 0.5 {
		selectivity = 0.5
	}
	if selectivity < 0.001 {
		selectivity = 0.001
	}

	return selectivity, nil
}

// 私有方法
func (sm *StatisticsManager) estimateEqualitySelectivity(colStats *ColumnStatistics, value interface{}) float64 {
	if colStats.DistinctCount <= 0 {
		return 0.5
	}

	// 检查是否是最常见值
	for _, mcv := range colStats.MostCommonValues {
		if compareValues(mcv.Value, value) == 0 {
			return mcv.Frequency
		}
	}

	// 如果不是最常见值，假设均匀分布
	return 1.0 / float64(colStats.DistinctCount)
}

func (sm *StatisticsManager) estimateRangeSelectivity(colStats *ColumnStatistics, value interface{}, lessThan bool) float64 {
	if colStats.Histogram == nil {
		// 没有直方图时使用简单估算
		if colStats.MinValue == nil || colStats.MaxValue == nil {
			return 0.5
		}

		// 简化的线性估算
		minVal := convertToFloat64(colStats.MinValue)
		maxVal := convertToFloat64(colStats.MaxValue)
		val := convertToFloat64(value)

		if maxVal <= minVal {
			return 0.5
		}

		ratio := (val - minVal) / (maxVal - minVal)
		if lessThan {
			return math.Max(0.0, math.Min(1.0, ratio))
		} else {
			return math.Max(0.0, math.Min(1.0, 1.0-ratio))
		}
	}

	// 使用直方图进行精确估算
	return sm.estimateWithHistogram(colStats.Histogram, value, lessThan)
}

func (sm *StatisticsManager) estimateWithHistogram(histogram *Histogram, value interface{}, lessThan bool) float64 {
	totalCount := float64(histogram.TotalCount)
	if totalCount <= 0 {
		return 0.5
	}

	matchingCount := int64(0)

	for _, bucket := range histogram.Buckets {
		if lessThan {
			if compareValues(bucket.UpperBound, value) <= 0 {
				matchingCount += bucket.Count
			} else if compareValues(bucket.LowerBound, value) < 0 {
				// 部分匹配，线性插值
				lowerVal := convertToFloat64(bucket.LowerBound)
				upperVal := convertToFloat64(bucket.UpperBound)
				val := convertToFloat64(value)

				if upperVal > lowerVal {
					ratio := (val - lowerVal) / (upperVal - lowerVal)
					matchingCount += int64(float64(bucket.Count) * ratio)
				}
			}
		} else {
			if compareValues(bucket.LowerBound, value) >= 0 {
				matchingCount += bucket.Count
			} else if compareValues(bucket.UpperBound, value) > 0 {
				// 部分匹配，线性插值
				lowerVal := convertToFloat64(bucket.LowerBound)
				upperVal := convertToFloat64(bucket.UpperBound)
				val := convertToFloat64(value)

				if upperVal > lowerVal {
					ratio := (upperVal - val) / (upperVal - lowerVal)
					matchingCount += int64(float64(bucket.Count) * ratio)
				}
			}
		}
	}

	return math.Max(0.0, math.Min(1.0, float64(matchingCount)/totalCount))
}

// ColumnStatsCollector 列统计信息收集器
type ColumnStatsCollector struct {
	columnName string
	dataType   types.DataType
	nullCount  int64
	totalCount int64
	minValue   interface{}
	maxValue   interface{}
	valueFreqs map[interface{}]int64 // 值频率统计
	samples    []interface{}         // 采样值（用于直方图）
}

// NewColumnStatsCollector 创建列统计收集器
func NewColumnStatsCollector(columnName string, dataType types.DataType) *ColumnStatsCollector {
	return &ColumnStatsCollector{
		columnName: columnName,
		dataType:   dataType,
		valueFreqs: make(map[interface{}]int64),
		samples:    make([]interface{}, 0),
	}
}

// ProcessArray 处理数组数据
func (csc *ColumnStatsCollector) ProcessArray(arr arrow.Array) {
	if arr == nil {
		return
	}

	for i := 0; i < arr.Len(); i++ {
		csc.totalCount++

		if arr.IsNull(i) {
			csc.nullCount++
			continue
		}

		var value interface{}
		switch a := arr.(type) {
		case *array.Int64:
			value = a.Value(i)
		case *array.Float64:
			value = a.Value(i)
		case *array.String:
			value = a.Value(i)
		case *array.Boolean:
			value = a.Value(i)
		default:
			continue
		}

		// 更新最值
		if csc.minValue == nil || compareValues(value, csc.minValue) < 0 {
			csc.minValue = value
		}
		if csc.maxValue == nil || compareValues(value, csc.maxValue) > 0 {
			csc.maxValue = value
		}

		// 统计频率
		csc.valueFreqs[value]++

		// 采样（限制样本数量）
		if len(csc.samples) < 1000 {
			csc.samples = append(csc.samples, value)
		}
	}
}

// Finalize 完成统计信息收集
func (csc *ColumnStatsCollector) Finalize() *ColumnStatistics {
	stats := &ColumnStatistics{
		ColumnName:    csc.columnName,
		DataType:      csc.dataType,
		NullCount:     csc.nullCount,
		DistinctCount: int64(len(csc.valueFreqs)),
		MinValue:      csc.minValue,
		MaxValue:      csc.maxValue,
		LastUpdated:   time.Now(),
	}

	// 计算最常见值
	stats.MostCommonValues = csc.calculateMCV()

	// 构建直方图
	stats.Histogram = csc.buildHistogram()

	return stats
}

// calculateMCV 计算最常见值
func (csc *ColumnStatsCollector) calculateMCV() []MCV {
	type valueCount struct {
		value interface{}
		count int64
	}

	// 转换为切片并排序
	values := make([]valueCount, 0, len(csc.valueFreqs))
	for value, count := range csc.valueFreqs {
		values = append(values, valueCount{value, count})
	}

	// 简单排序（找前10个最常见值）
	// TODO: 使用更高效的排序算法
	mcvs := make([]MCV, 0, 10)
	totalCount := float64(csc.totalCount)

	for i := 0; i < len(values) && i < 10; i++ {
		maxIdx := i
		for j := i + 1; j < len(values); j++ {
			if values[j].count > values[maxIdx].count {
				maxIdx = j
			}
		}
		if maxIdx != i {
			values[i], values[maxIdx] = values[maxIdx], values[i]
		}

		mcvs = append(mcvs, MCV{
			Value:     values[i].value,
			Frequency: float64(values[i].count) / totalCount,
		})
	}

	return mcvs
}

// buildHistogram 构建直方图
func (csc *ColumnStatsCollector) buildHistogram() *Histogram {
	if len(csc.samples) < 10 {
		return nil // 样本太少，不构建直方图
	}

	// TODO: 实现更sophisticated的直方图构建算法
	// 这里使用简化的等宽直方图

	const bucketCount = 10
	histogram := &Histogram{
		Buckets:    make([]HistogramBucket, bucketCount),
		TotalCount: csc.totalCount,
	}

	// 简化实现，仅支持数值类型
	if csc.dataType == types.Int64Type || csc.dataType == types.Float64Type {
		// 构建数值直方图
		csc.buildNumericHistogram(histogram, bucketCount)
	}

	return histogram
}

func (csc *ColumnStatsCollector) buildNumericHistogram(histogram *Histogram, bucketCount int) {
	minVal := convertToFloat64(csc.minValue)
	maxVal := convertToFloat64(csc.maxValue)

	if maxVal <= minVal {
		return
	}

	bucketWidth := (maxVal - minVal) / float64(bucketCount)

	// 初始化桶
	for i := 0; i < bucketCount; i++ {
		lower := minVal + float64(i)*bucketWidth
		upper := minVal + float64(i+1)*bucketWidth
		histogram.Buckets[i] = HistogramBucket{
			LowerBound: lower,
			UpperBound: upper,
			Count:      0,
		}
	}

	// 分配样本到桶中
	for _, sample := range csc.samples {
		val := convertToFloat64(sample)
		bucketIdx := int((val - minVal) / bucketWidth)
		if bucketIdx >= bucketCount {
			bucketIdx = bucketCount - 1
		}
		if bucketIdx < 0 {
			bucketIdx = 0
		}
		histogram.Buckets[bucketIdx].Count++
	}

	// 计算频率
	for i := range histogram.Buckets {
		histogram.Buckets[i].Frequency = float64(histogram.Buckets[i].Count) / float64(csc.totalCount)
	}
}

// 工具函数
func estimateRecordSize(record arrow.Record) int64 {
	// 简化的记录大小估算
	size := int64(0)
	for i := int64(0); i < record.NumCols(); i++ {
		column := record.Column(int(i))
		size += int64(column.Len()) * 8 // 假设平均每个值8字节
	}
	return size
}

func compareValues(a, b interface{}) int {
	switch va := a.(type) {
	case int64:
		if vb, ok := b.(int64); ok {
			if va < vb {
				return -1
			} else if va > vb {
				return 1
			} else {
				return 0
			}
		}
	case float64:
		if vb, ok := b.(float64); ok {
			if va < vb {
				return -1
			} else if va > vb {
				return 1
			} else {
				return 0
			}
		}
	case string:
		if vb, ok := b.(string); ok {
			if va < vb {
				return -1
			} else if va > vb {
				return 1
			} else {
				return 0
			}
		}
	}
	return 0
}

func convertToFloat64(value interface{}) float64 {
	switch v := value.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	case int:
		return float64(v)
	case float32:
		return float64(v)
	default:
		return 0.0
	}
}

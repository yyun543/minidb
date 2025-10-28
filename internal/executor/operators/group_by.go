package operators

import (
	"fmt"
	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/types"
	"strconv"
)

// GroupBy GROUP BY算子
type GroupBy struct {
	groupKeys     []optimizer.ColumnRef     // 分组键
	aggregations  []optimizer.AggregateExpr // 聚合表达式
	selectColumns []optimizer.ColumnRef     // SELECT列信息（包含别名）
	child         Operator                  // 子算子
	ctx           interface{}
	resultSent    bool                  // 是否已发送结果
	initialized   bool                  // 是否已初始化
	groupedData   map[string]*GroupData // 分组数据
}

// GroupData 存储每个分组的数据
type GroupData struct {
	keys       []interface{}          // 分组键值
	count      int64                  // 行数
	rows       [][]interface{}        // 该分组的所有行数据
	aggregates map[string]interface{} // 聚合计算结果
	sums       map[string]float64     // SUM计算累计值
	avgCounts  map[string]int64       // AVG计算行数
}

// NewGroupBy 创建GROUP BY算子
func NewGroupBy(groupKeys []optimizer.ColumnRef, aggregations []optimizer.AggregateExpr, selectColumns []optimizer.ColumnRef, child Operator, ctx interface{}) *GroupBy {
	return &GroupBy{
		groupKeys:     groupKeys,
		aggregations:  aggregations,
		selectColumns: selectColumns,
		child:         child,
		ctx:           ctx,
		resultSent:    false,
		initialized:   false,
		groupedData:   make(map[string]*GroupData),
	}
}

// Init 初始化算子
func (op *GroupBy) Init(ctx interface{}) error {
	return op.child.Init(ctx)
}

// Next 获取下一批数据
func (op *GroupBy) Next() (*types.Batch, error) {
	// 第一次调用时，处理所有数据
	if !op.initialized {
		if err := op.processAllData(); err != nil {
			return nil, err
		}
		op.initialized = true
	}

	// GROUP BY算子只返回一次结果
	if op.resultSent {
		return nil, nil
	}

	op.resultSent = true

	// 构建分组结果
	return op.buildGroupResult()
}

// processAllData 处理所有子算子数据进行分组
func (op *GroupBy) processAllData() error {
	for {
		batch, err := op.child.Next()
		if err != nil {
			return err
		}
		if batch == nil {
			break
		}

		if err := op.processGroupBatch(batch); err != nil {
			return err
		}
	}
	return nil
}

// processGroupBatch 处理单个批次的分组
func (op *GroupBy) processGroupBatch(batch *types.Batch) error {
	record := batch.Record()
	schema := record.Schema()

	// 找到分组列的索引
	groupKeyIndices := make([]int, len(op.groupKeys))
	for i, key := range op.groupKeys {
		found := false
		for j, field := range schema.Fields() {
			if field.Name == key.Column {
				groupKeyIndices[i] = j
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("group key column %s not found", key.Column)
		}
	}

	// 遍历每一行进行分组
	for rowIdx := int64(0); rowIdx < record.NumRows(); rowIdx++ {
		// 提取分组键值
		groupKeyValues := make([]interface{}, len(op.groupKeys))
		for i, colIdx := range groupKeyIndices {
			column := record.Column(colIdx)
			switch col := column.(type) {
			case *array.Int64:
				groupKeyValues[i] = col.Value(int(rowIdx))
			case *array.String:
				groupKeyValues[i] = col.Value(int(rowIdx))
			case *array.Float64:
				groupKeyValues[i] = col.Value(int(rowIdx))
			case *array.Boolean:
				groupKeyValues[i] = col.Value(int(rowIdx))
			default:
				groupKeyValues[i] = nil
			}
		}

		// 创建分组键字符串
		groupKey := op.makeGroupKeyString(groupKeyValues)

		// 提取整行数据
		rowData := make([]interface{}, record.NumCols())
		for colIdx := int64(0); colIdx < record.NumCols(); colIdx++ {
			column := record.Column(int(colIdx))
			switch col := column.(type) {
			case *array.Int64:
				rowData[colIdx] = col.Value(int(rowIdx))
			case *array.String:
				rowData[colIdx] = col.Value(int(rowIdx))
			case *array.Float64:
				rowData[colIdx] = col.Value(int(rowIdx))
			case *array.Boolean:
				rowData[colIdx] = col.Value(int(rowIdx))
			default:
				rowData[colIdx] = nil
			}
		}

		// 加入分组
		if group, exists := op.groupedData[groupKey]; exists {
			group.count++
			group.rows = append(group.rows, rowData)
			op.updateAggregates(group, record, int(rowIdx))
		} else {
			newGroup := &GroupData{
				keys:       groupKeyValues,
				count:      1,
				rows:       [][]interface{}{rowData},
				aggregates: make(map[string]interface{}),
				sums:       make(map[string]float64),
				avgCounts:  make(map[string]int64),
			}
			op.updateAggregates(newGroup, record, int(rowIdx))
			op.groupedData[groupKey] = newGroup
		}
	}

	return nil
}

// makeGroupKeyString 生成分组键字符串
func (op *GroupBy) makeGroupKeyString(values []interface{}) string {
	result := ""
	for i, value := range values {
		if i > 0 {
			result += "|"
		}
		result += fmt.Sprintf("%v", value)
	}
	return result
}

// updateAggregates 更新聚合计算
func (op *GroupBy) updateAggregates(group *GroupData, record arrow.Record, rowIdx int) {
	schema := record.Schema()

	for _, agg := range op.aggregations {
		aggKey := fmt.Sprintf("%s_%s", agg.Function, agg.Column)

		switch agg.Function {
		case "COUNT":
			if agg.Column == "*" {
				// COUNT(*) 计算所有行
				group.aggregates[aggKey] = group.count
			} else {
				// COUNT(column) 计算非空值
				colIdx := op.findColumnIndex(schema, agg.Column)
				if colIdx >= 0 && !record.Column(colIdx).IsNull(rowIdx) {
					if currentCount, exists := group.aggregates[aggKey]; exists {
						group.aggregates[aggKey] = currentCount.(int64) + 1
					} else {
						group.aggregates[aggKey] = int64(1)
					}
				}
			}

		case "SUM":
			colIdx := op.findColumnIndex(schema, agg.Column)
			if colIdx >= 0 && !record.Column(colIdx).IsNull(rowIdx) {
				value := op.getNumericValue(record.Column(colIdx), rowIdx)
				if value != nil {
					group.sums[aggKey] += *value
					group.aggregates[aggKey] = group.sums[aggKey]
				}
			}

		case "AVG":
			colIdx := op.findColumnIndex(schema, agg.Column)
			if colIdx >= 0 && !record.Column(colIdx).IsNull(rowIdx) {
				value := op.getNumericValue(record.Column(colIdx), rowIdx)
				if value != nil {
					group.sums[aggKey] += *value
					group.avgCounts[aggKey]++
					if group.avgCounts[aggKey] > 0 {
						group.aggregates[aggKey] = group.sums[aggKey] / float64(group.avgCounts[aggKey])
					}
				}
			}

		case "MIN":
			colIdx := op.findColumnIndex(schema, agg.Column)
			if colIdx >= 0 && !record.Column(colIdx).IsNull(rowIdx) {
				value := op.getNumericValue(record.Column(colIdx), rowIdx)
				if value != nil {
					if currentMin, exists := group.aggregates[aggKey]; exists {
						if *value < currentMin.(float64) {
							group.aggregates[aggKey] = *value
						}
					} else {
						group.aggregates[aggKey] = *value
					}
				}
			}

		case "MAX":
			colIdx := op.findColumnIndex(schema, agg.Column)
			if colIdx >= 0 && !record.Column(colIdx).IsNull(rowIdx) {
				value := op.getNumericValue(record.Column(colIdx), rowIdx)
				if value != nil {
					if currentMax, exists := group.aggregates[aggKey]; exists {
						if *value > currentMax.(float64) {
							group.aggregates[aggKey] = *value
						}
					} else {
						group.aggregates[aggKey] = *value
					}
				}
			}
		}
	}
}

// findColumnIndex 在schema中找到列的索引
func (op *GroupBy) findColumnIndex(schema *arrow.Schema, columnName string) int {
	for i, field := range schema.Fields() {
		if field.Name == columnName {
			return i
		}
	}
	return -1
}

// getNumericValue 从Arrow列中获取数值
func (op *GroupBy) getNumericValue(column arrow.Array, rowIdx int) *float64 {
	if column.IsNull(rowIdx) {
		return nil
	}

	switch col := column.(type) {
	case *array.Int64:
		val := float64(col.Value(rowIdx))
		return &val
	case *array.Float64:
		val := col.Value(rowIdx)
		return &val
	case *array.String:
		// 尝试将字符串转换为数字
		strVal := col.Value(rowIdx)
		if floatVal, err := strconv.ParseFloat(strVal, 64); err == nil {
			return &floatVal
		}
	}
	return nil
}

// inferGroupKeyType 从分组数据中推断分组键的类型
func (op *GroupBy) inferGroupKeyType(columnName string) arrow.DataType {
	// 找到列在groupKeys中的索引
	keyIndex := -1
	for i, key := range op.groupKeys {
		if key.Column == columnName {
			keyIndex = i
			break
		}
	}

	if keyIndex < 0 {
		// 默认为字符串
		return arrow.BinaryTypes.String
	}

	// 从任意一个分组中获取该键的值类型
	for _, group := range op.groupedData {
		if keyIndex < len(group.keys) {
			value := group.keys[keyIndex]
			switch value.(type) {
			case int64, int, int32:
				return arrow.PrimitiveTypes.Int64
			case float64, float32:
				return arrow.PrimitiveTypes.Float64
			case bool:
				return arrow.FixedWidthTypes.Boolean
			case string:
				return arrow.BinaryTypes.String
			}
		}
		break // 只需要检查第一个分组
	}

	// 默认为字符串
	return arrow.BinaryTypes.String
}

// buildGroupResult 构建分组结果
func (op *GroupBy) buildGroupResult() (*types.Batch, error) {
	// Special case: aggregation without GROUP BY (global aggregation)
	// According to SQL standard, aggregations without GROUP BY must return exactly one row
	// even when the input is empty
	if len(op.groupKeys) == 0 && len(op.aggregations) > 0 {
		// Global aggregation case - must return one row
		if len(op.groupedData) == 0 {
			// Create a synthetic empty group for global aggregation
			op.groupedData[""] = &GroupData{
				keys:       []interface{}{},
				count:      0,
				rows:       [][]interface{}{},
				aggregates: make(map[string]interface{}),
				sums:       make(map[string]float64),
				avgCounts:  make(map[string]int64),
			}
		}
	} else if len(op.groupedData) == 0 {
		// Regular GROUP BY with no data - return empty result
		return nil, nil
	}

	// 根据selectColumns构建结果schema
	var fields []arrow.Field
	for _, col := range op.selectColumns {
		var fieldName string
		var fieldType arrow.DataType

		// 使用别名作为列名，如果没有别名则使用原列名
		if col.Alias != "" {
			fieldName = col.Alias
		} else if col.Type == optimizer.ColumnRefTypeFunction {
			fieldName = fmt.Sprintf("%s(%s)", col.FunctionName, col.Column)
		} else {
			fieldName = col.Column
		}

		// 根据列类型确定Arrow数据类型
		if col.Type == optimizer.ColumnRefTypeFunction {
			switch col.FunctionName {
			case "COUNT":
				fieldType = arrow.PrimitiveTypes.Int64
			case "SUM", "AVG", "MIN", "MAX":
				fieldType = arrow.PrimitiveTypes.Float64
			default:
				fieldType = arrow.BinaryTypes.String
			}
		} else {
			// 分组键应保留原始类型，从groupedData中推断
			fieldType = op.inferGroupKeyType(col.Column)
		}

		fields = append(fields, arrow.Field{
			Name: fieldName,
			Type: fieldType,
		})
	}

	schema := arrow.NewSchema(fields, nil)

	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 填充数据
	for _, group := range op.groupedData {
		for colIdx, col := range op.selectColumns {
			field := builder.Field(colIdx)

			if col.Type == optimizer.ColumnRefTypeFunction {
				// 处理聚合函数
				// 修复聚合函数列名丢失问题
				column := col.Column
				if column == "" {
					// 如果Column为空，从aggregations中查找对应的列名
					for _, agg := range op.aggregations {
						if agg.Function == col.FunctionName && agg.Alias == col.Alias {
							column = agg.Column
							break
						}
					}
					// 特殊处理COUNT(*)
					if col.FunctionName == "COUNT" && column == "" {
						column = "*"
					}
				}
				aggKey := fmt.Sprintf("%s_%s", col.FunctionName, column)
				value := group.aggregates[aggKey]

				switch col.FunctionName {
				case "COUNT":
					if intBuilder, ok := field.(*array.Int64Builder); ok {
						if value != nil {
							intBuilder.Append(value.(int64))
						} else {
							// COUNT on empty set returns 0
							intBuilder.Append(0)
						}
					}
				case "SUM", "AVG", "MIN", "MAX":
					if floatBuilder, ok := field.(*array.Float64Builder); ok {
						if value != nil {
							floatBuilder.Append(value.(float64))
						} else {
							// SUM/AVG/MIN/MAX on empty set returns NULL per SQL standard
							floatBuilder.AppendNull()
						}
					}
				}
			} else {
				// 处理分组键 - 保留原始类型
				// 在分组键中找到对应的值
				keyValue := interface{}(nil)
				for i, key := range op.groupKeys {
					if key.Column == col.Column {
						keyValue = group.keys[i]
						break
					}
				}

				if keyValue == nil {
					field.AppendNull()
					continue
				}

				// 根据字段类型追加值
				switch builder := field.(type) {
				case *array.Int64Builder:
					switch v := keyValue.(type) {
					case int64:
						builder.Append(v)
					case int:
						builder.Append(int64(v))
					case int32:
						builder.Append(int64(v))
					default:
						builder.AppendNull()
					}
				case *array.Float64Builder:
					switch v := keyValue.(type) {
					case float64:
						builder.Append(v)
					case float32:
						builder.Append(float64(v))
					case int64:
						builder.Append(float64(v))
					case int:
						builder.Append(float64(v))
					default:
						builder.AppendNull()
					}
				case *array.BooleanBuilder:
					if v, ok := keyValue.(bool); ok {
						builder.Append(v)
					} else {
						builder.AppendNull()
					}
				case *array.StringBuilder:
					builder.Append(fmt.Sprintf("%v", keyValue))
				default:
					// Fallback: append null for unsupported types
					field.AppendNull()
				}
			}
		}
	}

	groupedRecord := builder.NewRecord()
	if groupedRecord.NumRows() == 0 {
		groupedRecord.Release()
		return nil, nil
	}

	return types.NewBatch(groupedRecord), nil
}

// Close 关闭算子
func (op *GroupBy) Close() error {
	return op.child.Close()
}

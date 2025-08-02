package operators

import (
	"fmt"
	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/types"
	"sort"
)

// OrderBy ORDER BY算子
type OrderBy struct {
	orderKeys   []optimizer.OrderKey // 排序键
	child       Operator             // 子算子
	ctx         interface{}
	resultSent  bool           // 是否已发送结果
	initialized bool           // 是否已初始化
	sortedData  []*types.Batch // 排序后的数据
}

// sortableRow 可排序的行数据
type sortableRow struct {
	data      []interface{} // 行数据
	keyValues []interface{} // 排序键值
}

// sortableRows 可排序的行集合
type sortableRows struct {
	rows      []sortableRow
	orderKeys []optimizer.OrderKey
}

// NewOrderBy 创建ORDER BY算子
func NewOrderBy(orderKeys []optimizer.OrderKey, child Operator, ctx interface{}) *OrderBy {
	return &OrderBy{
		orderKeys:   orderKeys,
		child:       child,
		ctx:         ctx,
		resultSent:  false,
		initialized: false,
		sortedData:  make([]*types.Batch, 0),
	}
}

// Init 初始化算子
func (op *OrderBy) Init(ctx interface{}) error {
	return op.child.Init(ctx)
}

// Next 获取下一批数据
func (op *OrderBy) Next() (*types.Batch, error) {
	// 第一次调用时，处理所有数据并排序
	if !op.initialized {
		if err := op.processAndSortAllData(); err != nil {
			return nil, err
		}
		op.initialized = true
	}

	// ORDER BY算子只返回一次结果
	if op.resultSent {
		return nil, nil
	}

	op.resultSent = true

	// 返回排序后的结果
	if len(op.sortedData) > 0 {
		return op.sortedData[0], nil
	}

	return nil, nil
}

// processAndSortAllData 处理所有子算子数据并排序
func (op *OrderBy) processAndSortAllData() error {
	// 收集所有数据
	var allRows []sortableRow
	var schema *arrow.Schema
	var orderKeyIndices []int

	for {
		batch, err := op.child.Next()
		if err != nil {
			return err
		}
		if batch == nil {
			break
		}

		record := batch.Record()

		// 第一次处理时，保存schema并查找排序列索引
		if schema == nil {
			schema = record.Schema()

			// 找到排序列的索引
			orderKeyIndices = make([]int, len(op.orderKeys))
			for i, key := range op.orderKeys {
				found := false
				for j, field := range schema.Fields() {
					if field.Name == key.Column {
						orderKeyIndices[i] = j
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("order key column %s not found in schema", key.Column)
				}
			}
		}

		// 提取每一行数据
		for rowIdx := int64(0); rowIdx < record.NumRows(); rowIdx++ {
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
				default:
					rowData[colIdx] = nil
				}
			}

			// 提取排序键值
			keyValues := make([]interface{}, len(op.orderKeys))
			for i, colIdx := range orderKeyIndices {
				keyValues[i] = rowData[colIdx]
			}

			allRows = append(allRows, sortableRow{
				data:      rowData,
				keyValues: keyValues,
			})
		}
	}

	// 排序数据
	if len(allRows) > 0 {
		sortableData := &sortableRows{
			rows:      allRows,
			orderKeys: op.orderKeys,
		}
		sort.Sort(sortableData)

		// 重建Arrow记录
		sortedBatch, err := op.buildSortedResult(allRows, schema)
		if err != nil {
			return err
		}

		if sortedBatch != nil {
			op.sortedData = append(op.sortedData, sortedBatch)
		}
	}

	return nil
}

// buildSortedResult 构建排序后的结果
func (op *OrderBy) buildSortedResult(rows []sortableRow, schema *arrow.Schema) (*types.Batch, error) {
	if len(rows) == 0 {
		return nil, nil
	}

	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 填充排序后的数据
	for _, row := range rows {
		for colIdx, value := range row.data {
			field := builder.Field(colIdx)
			switch b := field.(type) {
			case *array.Int64Builder:
				if intVal, ok := value.(int64); ok {
					b.Append(intVal)
				} else {
					b.AppendNull()
				}
			case *array.StringBuilder:
				if strVal, ok := value.(string); ok {
					b.Append(strVal)
				} else {
					b.AppendNull()
				}
			case *array.Float64Builder:
				if floatVal, ok := value.(float64); ok {
					b.Append(floatVal)
				} else {
					b.AppendNull()
				}
			default:
				// 对于其他类型，尝试转换为字符串
				if b, ok := field.(*array.StringBuilder); ok {
					b.Append(fmt.Sprintf("%v", value))
				}
			}
		}
	}

	sortedRecord := builder.NewRecord()
	if sortedRecord.NumRows() == 0 {
		sortedRecord.Release()
		return nil, nil
	}

	return types.NewBatch(sortedRecord), nil
}

// Close 关闭算子
func (op *OrderBy) Close() error {
	return op.child.Close()
}

// 实现sort.Interface接口
func (sr *sortableRows) Len() int {
	return len(sr.rows)
}

func (sr *sortableRows) Swap(i, j int) {
	sr.rows[i], sr.rows[j] = sr.rows[j], sr.rows[i]
}

func (sr *sortableRows) Less(i, j int) bool {
	row1 := sr.rows[i]
	row2 := sr.rows[j]

	// 按多个排序键比较
	for keyIdx, orderKey := range sr.orderKeys {
		val1 := row1.keyValues[keyIdx]
		val2 := row2.keyValues[keyIdx]

		cmp := compareValues(val1, val2)

		// 处理DESC排序
		if orderKey.Direction == "DESC" {
			cmp = -cmp
		}

		if cmp != 0 {
			return cmp < 0
		}
	}

	return false // 相等
}

// compareValues 比较两个值
func compareValues(val1, val2 interface{}) int {
	// 处理nil值
	if val1 == nil && val2 == nil {
		return 0
	}
	if val1 == nil {
		return -1
	}
	if val2 == nil {
		return 1
	}

	// 类型转换和比较
	switch v1 := val1.(type) {
	case int64:
		if v2, ok := val2.(int64); ok {
			if v1 < v2 {
				return -1
			} else if v1 > v2 {
				return 1
			}
			return 0
		}
	case string:
		if v2, ok := val2.(string); ok {
			if v1 < v2 {
				return -1
			} else if v1 > v2 {
				return 1
			}
			return 0
		}
	case float64:
		if v2, ok := val2.(float64); ok {
			if v1 < v2 {
				return -1
			} else if v1 > v2 {
				return 1
			}
			return 0
		}
	}

	// 默认字符串比较
	str1 := fmt.Sprintf("%v", val1)
	str2 := fmt.Sprintf("%v", val2)
	if str1 < str2 {
		return -1
	} else if str1 > str2 {
		return 1
	}
	return 0
}

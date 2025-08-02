package operators

import (
	"fmt"
	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/types"
)

// GroupBy GROUP BY算子
type GroupBy struct {
	groupKeys   []optimizer.ColumnRef // 分组键
	child       Operator              // 子算子
	ctx         interface{}
	resultSent  bool                  // 是否已发送结果
	initialized bool                  // 是否已初始化
	groupedData map[string]*GroupData // 分组数据
}

// GroupData 存储每个分组的数据
type GroupData struct {
	keys  []interface{}   // 分组键值
	count int64           // 行数
	rows  [][]interface{} // 该分组的所有行数据
}

// NewGroupBy 创建GROUP BY算子
func NewGroupBy(groupKeys []optimizer.ColumnRef, child Operator, ctx interface{}) *GroupBy {
	return &GroupBy{
		groupKeys:   groupKeys,
		child:       child,
		ctx:         ctx,
		resultSent:  false,
		initialized: false,
		groupedData: make(map[string]*GroupData),
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
			default:
				rowData[colIdx] = nil
			}
		}

		// 加入分组
		if group, exists := op.groupedData[groupKey]; exists {
			group.count++
			group.rows = append(group.rows, rowData)
		} else {
			op.groupedData[groupKey] = &GroupData{
				keys:  groupKeyValues,
				count: 1,
				rows:  [][]interface{}{rowData},
			}
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

// buildGroupResult 构建分组结果
func (op *GroupBy) buildGroupResult() (*types.Batch, error) {
	if len(op.groupedData) == 0 {
		return nil, nil
	}

	// 简化实现：创建包含分组键和COUNT的结果
	// 构建结果schema：分组键列 + COUNT列
	var fields []arrow.Field
	for _, key := range op.groupKeys {
		// 简化：假设所有分组键都是字符串类型
		fields = append(fields, arrow.Field{
			Name: key.Column,
			Type: arrow.BinaryTypes.String,
		})
	}
	fields = append(fields, arrow.Field{
		Name: "COUNT(*)",
		Type: arrow.PrimitiveTypes.Int64,
	})

	schema := arrow.NewSchema(fields, nil)

	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 填充数据
	for _, group := range op.groupedData {
		// 添加分组键值
		for i, keyValue := range group.keys {
			field := builder.Field(i)
			if strBuilder, ok := field.(*array.StringBuilder); ok {
				strBuilder.Append(fmt.Sprintf("%v", keyValue))
			}
		}

		// 添加COUNT值
		field := builder.Field(len(group.keys))
		if intBuilder, ok := field.(*array.Int64Builder); ok {
			intBuilder.Append(group.count)
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

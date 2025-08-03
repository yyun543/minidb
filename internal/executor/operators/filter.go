package operators

import (
	"strings"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/types"
)

// Filter 过滤算子
type Filter struct {
	condition optimizer.Expression // 过滤条件
	child     Operator             // Use local Operator interface
	ctx       interface{}          // Use interface{} instead of *executor.Context
}

// NewFilter 创建过滤算子
func NewFilter(condition optimizer.Expression, child Operator, ctx interface{}) *Filter {
	return &Filter{
		condition: condition,
		child:     child,
		ctx:       ctx,
	}
}

// Init 初始化算子
func (op *Filter) Init(ctx interface{}) error {
	return op.child.Init(ctx)
}

// Next 获取下一批数据
func (op *Filter) Next() (*types.Batch, error) {
	// 获取子算子数据
	batch, err := op.child.Next()
	if err != nil {
		return nil, err
	}
	if batch == nil {
		return nil, nil
	}

	// 应用过滤条件
	filteredRecord, err := op.applyFilter(batch.Record())
	if err != nil {
		return nil, err
	}

	if filteredRecord == nil || filteredRecord.NumRows() == 0 {
		return nil, nil
	}

	return types.NewBatch(filteredRecord), nil
}

// Close 关闭算子
func (op *Filter) Close() error {
	return op.child.Close()
}

// applyFilter 应用过滤条件
func (op *Filter) applyFilter(record arrow.Record) (arrow.Record, error) {
	// 支持二元比较表达式 (column = value, column LIKE pattern, etc.)
	if binExpr, ok := op.condition.(*optimizer.BinaryExpression); ok {
		return op.applyBinaryFilter(record, binExpr)
	}

	// TODO: Support IN expressions when optimizer.InExpression is implemented

	// 如果不支持的条件类型，返回原记录
	return record, nil
}

// applyBinaryFilter 应用二元表达式过滤
func (op *Filter) applyBinaryFilter(record arrow.Record, binExpr *optimizer.BinaryExpression) (arrow.Record, error) {
	// 获取列引用和比较值
	var colName string
	var compareValue interface{}

	if colRef, ok := binExpr.Left.(*optimizer.ColumnReference); ok {
		colName = colRef.Column
	} else {
		return record, nil // 不支持的左表达式类型
	}

	// 使用类型断言处理不同的表达式类型
	switch rightExpr := binExpr.Right.(type) {
	case *optimizer.LiteralValue:
		compareValue = rightExpr.Value
	case interface{ GetValue() interface{} }:
		compareValue = rightExpr.GetValue()
	default:
		return record, nil // 不支持的右表达式类型
	}

	// 找到列索引
	schema := record.Schema()
	colIndex := -1
	for i, field := range schema.Fields() {
		if field.Name == colName {
			colIndex = i
			break
		}
	}

	if colIndex == -1 {
		return record, nil // 列不存在，返回原记录
	}

	// 根据操作符类型过滤记录
	switch binExpr.Operator {
	case "=":
		return op.filterRecordByColumn(record, colIndex, compareValue, "=")
	case "LIKE":
		return op.filterRecordByColumn(record, colIndex, compareValue, "LIKE")
	case "NOT LIKE":
		return op.filterRecordByColumn(record, colIndex, compareValue, "NOT LIKE")
	case ">", "<", ">=", "<=", "!=", "<>":
		return op.filterRecordByColumn(record, colIndex, compareValue, binExpr.Operator)
	default:
		// 不支持的操作符，返回原记录
		return record, nil
	}
}

// filterRecordByColumn 按列值过滤记录
func (op *Filter) filterRecordByColumn(record arrow.Record, colIndex int, compareValue interface{}, operator string) (arrow.Record, error) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, record.Schema())
	defer builder.Release()

	column := record.Column(colIndex)

	// 检查每一行
	for rowIdx := int64(0); rowIdx < record.NumRows(); rowIdx++ {
		var matches bool

		// 根据列类型和操作符比较值
		switch col := column.(type) {
		case *array.Int64:
			if intVal, ok := compareValue.(int64); ok {
				columnValue := col.Value(int(rowIdx))
				matches = op.compareValues(columnValue, intVal, operator)
			} else if intVal, ok := compareValue.(int); ok {
				columnValue := col.Value(int(rowIdx))
				matches = op.compareValues(columnValue, int64(intVal), operator)
			}
		case *array.String:
			if strVal, ok := compareValue.(string); ok {
				columnValue := col.Value(int(rowIdx))
				matches = op.compareStrings(columnValue, strVal, operator)
			}
		case *array.Float64:
			if floatVal, ok := compareValue.(float64); ok {
				columnValue := col.Value(int(rowIdx))
				matches = op.compareValues(columnValue, floatVal, operator)
			}
		}

		// 如果匹配，复制这一行到结果
		if matches {
			for colIdx := int64(0); colIdx < record.NumCols(); colIdx++ {
				field := builder.Field(int(colIdx))
				srcCol := record.Column(int(colIdx))

				switch srcCol := srcCol.(type) {
				case *array.Int64:
					if intBuilder, ok := field.(*array.Int64Builder); ok {
						intBuilder.Append(srcCol.Value(int(rowIdx)))
					}
				case *array.String:
					if strBuilder, ok := field.(*array.StringBuilder); ok {
						strBuilder.Append(srcCol.Value(int(rowIdx)))
					}
				case *array.Float64:
					if floatBuilder, ok := field.(*array.Float64Builder); ok {
						floatBuilder.Append(srcCol.Value(int(rowIdx)))
					}
				}
			}
		}
	}

	return builder.NewRecord(), nil
}

// compareValues 比较数值类型 (int64, float64)
func (op *Filter) compareValues(columnValue, compareValue interface{}, operator string) bool {
	switch operator {
	case "=":
		return columnValue == compareValue
	case "!=", "<>":
		return columnValue != compareValue
	case ">":
		switch cv := columnValue.(type) {
		case int64:
			if cmp, ok := compareValue.(int64); ok {
				return cv > cmp
			}
		case float64:
			if cmp, ok := compareValue.(float64); ok {
				return cv > cmp
			}
		}
	case "<":
		switch cv := columnValue.(type) {
		case int64:
			if cmp, ok := compareValue.(int64); ok {
				return cv < cmp
			}
		case float64:
			if cmp, ok := compareValue.(float64); ok {
				return cv < cmp
			}
		}
	case ">=":
		switch cv := columnValue.(type) {
		case int64:
			if cmp, ok := compareValue.(int64); ok {
				return cv >= cmp
			}
		case float64:
			if cmp, ok := compareValue.(float64); ok {
				return cv >= cmp
			}
		}
	case "<=":
		switch cv := columnValue.(type) {
		case int64:
			if cmp, ok := compareValue.(int64); ok {
				return cv <= cmp
			}
		case float64:
			if cmp, ok := compareValue.(float64); ok {
				return cv <= cmp
			}
		}
	}
	return false
}

// compareStrings 比较字符串类型 (包括LIKE操作)
func (op *Filter) compareStrings(columnValue, compareValue string, operator string) bool {
	switch operator {
	case "=":
		return columnValue == compareValue
	case "!=", "<>":
		return columnValue != compareValue
	case ">":
		return columnValue > compareValue
	case "<":
		return columnValue < compareValue
	case ">=":
		return columnValue >= compareValue
	case "<=":
		return columnValue <= compareValue
	case "LIKE":
		return op.matchLikePattern(columnValue, compareValue)
	case "NOT LIKE":
		return !op.matchLikePattern(columnValue, compareValue)
	}
	return false
}

// matchLikePattern 简单的LIKE模式匹配
func (op *Filter) matchLikePattern(value, pattern string) bool {
	// 简化的LIKE实现：只支持%通配符
	if !strings.Contains(pattern, "%") {
		// 没有通配符，等于比较
		return value == pattern
	}

	// 处理开头的%
	if strings.HasPrefix(pattern, "%") && strings.HasSuffix(pattern, "%") {
		// %text% - 包含
		middle := pattern[1 : len(pattern)-1]
		return strings.Contains(value, middle)
	} else if strings.HasPrefix(pattern, "%") {
		// %text - 以text结尾
		suffix := pattern[1:]
		return strings.HasSuffix(value, suffix)
	} else if strings.HasSuffix(pattern, "%") {
		// text% - 以text开头
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(value, prefix)
	}

	// 其他情况暂不支持，返回false
	return false
}

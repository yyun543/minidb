package operators

import (
	"fmt"
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
	println("DEBUG Filter.Init called, condition=", fmt.Sprintf("%v", op.condition))
	return op.child.Init(ctx)
}

// Next 获取下一批数据
func (op *Filter) Next() (*types.Batch, error) {
	println("DEBUG Filter.Next called")
	// 循环直到找到匹配的行或所有数据处理完
	for {
		// 获取子算子数据
		batch, err := op.child.Next()
		if err != nil {
			println("DEBUG Filter.Next child.Next returned error:", err.Error())
			return nil, err
		}
		if batch == nil {
			println("DEBUG Filter.Next child.Next returned nil")
			return nil, nil
		}
		println("DEBUG Filter.Next received batch with", batch.Record().NumRows(), "rows")

		// 应用过滤条件
		filteredRecord, err := op.applyFilter(batch.Record())
		if err != nil {
			return nil, err
		}

		// 如果过滤结果为空，继续处理下一个 batch
		if filteredRecord == nil || filteredRecord.NumRows() == 0 {
			continue
		}

		return types.NewBatch(filteredRecord), nil
	}
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
	println("DEBUG applyBinaryFilter called, operator=", binExpr.Operator, "record.NumRows=", record.NumRows())

	// Check if this is a logical operator (AND/OR)
	if binExpr.Operator == "AND" || binExpr.Operator == "OR" {
		return op.applyLogicalFilter(record, binExpr)
	}

	// 获取列引用和比较值
	var colName string
	var compareValue interface{}

	if colRef, ok := binExpr.Left.(*optimizer.ColumnReference); ok {
		colName = colRef.Column
		println("DEBUG colName=", colName)
	} else {
		println("DEBUG left is not ColumnReference, type=", fmt.Sprintf("%T", binExpr.Left))
		return record, nil // 不支持的左表达式类型
	}

	// 使用类型断言处理不同的表达式类型
	switch rightExpr := binExpr.Right.(type) {
	case *optimizer.LiteralValue:
		compareValue = rightExpr.Value
		println("DEBUG compareValue from LiteralValue=", compareValue, "type=", fmt.Sprintf("%T", compareValue))
	case interface{ GetValue() interface{} }:
		compareValue = rightExpr.GetValue()
		println("DEBUG compareValue from GetValue()=", compareValue, "type=", fmt.Sprintf("%T", compareValue))
	default:
		println("DEBUG right is not LiteralValue, type=", fmt.Sprintf("%T", binExpr.Right))
		return record, nil // 不支持的右表达式类型
	}

	// 找到列索引
	schema := record.Schema()
	colIndex := -1
	for i, field := range schema.Fields() {
		if field.Name == colName {
			colIndex = i
			println("DEBUG found column", colName, "at index", i, "type=", field.Type.String())
			break
		}
	}

	if colIndex == -1 {
		println("DEBUG column", colName, "not found in schema")
		return record, nil // 列不存在，返回原记录
	}

	// 根据操作符类型过滤记录
	println("DEBUG calling filterRecordByColumn, colIndex=", colIndex, "operator=", binExpr.Operator)
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

// applyLogicalFilter 应用逻辑表达式过滤 (AND/OR)
func (op *Filter) applyLogicalFilter(record arrow.Record, binExpr *optimizer.BinaryExpression) (arrow.Record, error) {
	if binExpr.Operator == "AND" {
		// Apply left filter first
		leftBinExpr, ok := binExpr.Left.(*optimizer.BinaryExpression)
		if !ok {
			return record, nil
		}
		tempFilterLeft := &Filter{condition: leftBinExpr, child: nil, ctx: op.ctx}
		leftResult, err := tempFilterLeft.applyBinaryFilter(record, leftBinExpr)
		if err != nil {
			return nil, err
		}
		if leftResult == nil || leftResult.NumRows() == 0 {
			return leftResult, nil
		}

		// Apply right filter on the result
		rightBinExpr, ok := binExpr.Right.(*optimizer.BinaryExpression)
		if !ok {
			return leftResult, nil
		}
		tempFilterRight := &Filter{condition: rightBinExpr, child: nil, ctx: op.ctx}
		return tempFilterRight.applyBinaryFilter(leftResult, rightBinExpr)
	} else if binExpr.Operator == "OR" {
		// Apply left and right filters separately
		leftBinExpr, ok := binExpr.Left.(*optimizer.BinaryExpression)
		if !ok {
			return record, nil
		}
		tempFilterLeft := &Filter{condition: leftBinExpr, child: nil, ctx: op.ctx}
		leftResult, err := tempFilterLeft.applyBinaryFilter(record, leftBinExpr)
		if err != nil {
			return nil, err
		}

		rightBinExpr, ok := binExpr.Right.(*optimizer.BinaryExpression)
		if !ok {
			return leftResult, nil
		}
		tempFilterRight := &Filter{condition: rightBinExpr, child: nil, ctx: op.ctx}
		rightResult, err := tempFilterRight.applyBinaryFilter(record, rightBinExpr)
		if err != nil {
			return nil, err
		}

		// Merge results (union without duplicates)
		return op.mergeRecords(leftResult, rightResult)
	}

	return record, nil
}

// mergeRecords 合并两个记录（用于OR操作）
func (op *Filter) mergeRecords(left, right arrow.Record) (arrow.Record, error) {
	if left == nil || left.NumRows() == 0 {
		return right, nil
	}
	if right == nil || right.NumRows() == 0 {
		return left, nil
	}

	// Simple implementation: just concatenate the records
	// TODO: Remove duplicates if needed
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, left.Schema())
	defer builder.Release()

	// Copy rows from left
	for rowIdx := int64(0); rowIdx < left.NumRows(); rowIdx++ {
		for colIdx := int64(0); colIdx < left.NumCols(); colIdx++ {
			field := builder.Field(int(colIdx))
			srcCol := left.Column(int(colIdx))

			switch srcCol := srcCol.(type) {
			case *array.Int64:
				if intBuilder, ok := field.(*array.Int64Builder); ok {
					intBuilder.Append(srcCol.Value(int(rowIdx)))
				}
			case *array.Int32:
				if intBuilder, ok := field.(*array.Int32Builder); ok {
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
			case *array.Float32:
				if floatBuilder, ok := field.(*array.Float32Builder); ok {
					floatBuilder.Append(srcCol.Value(int(rowIdx)))
				}
			case *array.Boolean:
				if boolBuilder, ok := field.(*array.BooleanBuilder); ok {
					boolBuilder.Append(srcCol.Value(int(rowIdx)))
				}
			case *array.Timestamp:
				if tsBuilder, ok := field.(*array.TimestampBuilder); ok {
					tsBuilder.Append(srcCol.Value(int(rowIdx)))
				}
			}
		}
	}

	// Copy rows from right
	for rowIdx := int64(0); rowIdx < right.NumRows(); rowIdx++ {
		for colIdx := int64(0); colIdx < right.NumCols(); colIdx++ {
			field := builder.Field(int(colIdx))
			srcCol := right.Column(int(colIdx))

			switch srcCol := srcCol.(type) {
			case *array.Int64:
				if intBuilder, ok := field.(*array.Int64Builder); ok {
					intBuilder.Append(srcCol.Value(int(rowIdx)))
				}
			case *array.Int32:
				if intBuilder, ok := field.(*array.Int32Builder); ok {
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
			case *array.Float32:
				if floatBuilder, ok := field.(*array.Float32Builder); ok {
					floatBuilder.Append(srcCol.Value(int(rowIdx)))
				}
			case *array.Boolean:
				if boolBuilder, ok := field.(*array.BooleanBuilder); ok {
					boolBuilder.Append(srcCol.Value(int(rowIdx)))
				}
			case *array.Timestamp:
				if tsBuilder, ok := field.(*array.TimestampBuilder); ok {
					tsBuilder.Append(srcCol.Value(int(rowIdx)))
				}
			}
		}
	}

	return builder.NewRecord(), nil
}

// filterRecordByColumn 按列值过滤记录
func (op *Filter) filterRecordByColumn(record arrow.Record, colIndex int, compareValue interface{}, operator string) (arrow.Record, error) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, record.Schema())
	defer builder.Release()

	column := record.Column(colIndex)
	schema := record.Schema()
	field := schema.Field(colIndex)

	// Check if this is a boolean column based on schema
	isBooleanColumn := (field.Type.ID() == arrow.BOOL)

	// 检查每一行
	for rowIdx := int64(0); rowIdx < record.NumRows(); rowIdx++ {
		var matches bool

		// Special handling for boolean columns
		if isBooleanColumn {
			boolCol, ok := column.(*array.Boolean)
			if !ok {
				// Column schema says BOOL but array type is not Boolean - data corruption?
				continue
			}

			columnValue := boolCol.Value(int(rowIdx))

			// Convert compareValue to boolean
			var boolVal bool
			switch v := compareValue.(type) {
			case bool:
				boolVal = v
			case int64:
				boolVal = (v != 0)
			case int:
				boolVal = (v != 0)
			case int32:
				boolVal = (v != 0)
			case string:
				// Handle "true"/"false"/"1"/"0"
				switch v {
				case "true", "1", "t", "T", "TRUE":
					boolVal = true
				case "false", "0", "f", "F", "FALSE":
					boolVal = false
				default:
					continue // Invalid boolean value, skip this row
				}
			default:
				// Unknown type, skip
				continue
			}

			// Compare booleans
			switch operator {
			case "=", "==":
				matches = (columnValue == boolVal)
			case "!=", "<>":
				matches = (columnValue != boolVal)
			default:
				matches = false
			}
		} else {
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
			case *array.Int32:
				if intVal, ok := compareValue.(int32); ok {
					columnValue := col.Value(int(rowIdx))
					matches = op.compareValues(int64(columnValue), int64(intVal), operator)
				} else if intVal, ok := compareValue.(int); ok {
					columnValue := col.Value(int(rowIdx))
					matches = op.compareValues(int64(columnValue), int64(intVal), operator)
				} else if intVal, ok := compareValue.(int64); ok {
					columnValue := col.Value(int(rowIdx))
					matches = op.compareValues(int64(columnValue), intVal, operator)
				}
			case *array.String:
				if strVal, ok := compareValue.(string); ok {
					columnValue := col.Value(int(rowIdx))
					matches = op.compareStrings(columnValue, strVal, operator)
				}
			case *array.Float64:
				columnValue := col.Value(int(rowIdx))
				if floatVal, ok := compareValue.(float64); ok {
					matches = op.compareValues(columnValue, floatVal, operator)
				} else if intVal, ok := compareValue.(int64); ok {
					// 支持int64与float64的比较（常见于HAVING子句中的整数字面量）
					matches = op.compareValues(columnValue, float64(intVal), operator)
				} else if intVal, ok := compareValue.(int); ok {
					// 支持int与float64的比较
					matches = op.compareValues(columnValue, float64(intVal), operator)
				}
			case *array.Float32:
				if floatVal, ok := compareValue.(float32); ok {
					columnValue := col.Value(int(rowIdx))
					matches = op.compareValues(float64(columnValue), float64(floatVal), operator)
				} else if floatVal, ok := compareValue.(float64); ok {
					columnValue := col.Value(int(rowIdx))
					matches = op.compareValues(float64(columnValue), floatVal, operator)
				}
			case *array.Timestamp:
				if tsVal, ok := compareValue.(arrow.Timestamp); ok {
					columnValue := col.Value(int(rowIdx))
					matches = op.compareValues(int64(columnValue), int64(tsVal), operator)
				}
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
				case *array.Int32:
					if intBuilder, ok := field.(*array.Int32Builder); ok {
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
				case *array.Float32:
					if floatBuilder, ok := field.(*array.Float32Builder); ok {
						floatBuilder.Append(srcCol.Value(int(rowIdx)))
					}
				case *array.Boolean:
					if boolBuilder, ok := field.(*array.BooleanBuilder); ok {
						boolBuilder.Append(srcCol.Value(int(rowIdx)))
					}
				case *array.Timestamp:
					if tsBuilder, ok := field.(*array.TimestampBuilder); ok {
						tsBuilder.Append(srcCol.Value(int(rowIdx)))
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

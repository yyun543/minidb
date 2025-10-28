package operators

import (
	"fmt"
	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/types"
	"strings"
)

// Projection 投影算子 - 用于选择特定的列
type Projection struct {
	columns []optimizer.ColumnRef // 要投影的列
	child   Operator              // 子算子
	ctx     interface{}
}

// NewProjection 创建投影算子
func NewProjection(columns []optimizer.ColumnRef, child Operator, ctx interface{}) *Projection {
	return &Projection{
		columns: columns,
		child:   child,
		ctx:     ctx,
	}
}

// Init 初始化算子
func (op *Projection) Init(ctx interface{}) error {
	return op.child.Init(ctx)
}

// Next 获取下一批数据并应用投影
func (op *Projection) Next() (*types.Batch, error) {
	// 从子算子获取数据
	childBatch, err := op.child.Next()
	if err != nil {
		return nil, err
	}
	if childBatch == nil {
		return nil, nil
	}

	// 应用投影
	return op.applyProjection(childBatch)
}

// Close 关闭算子
func (op *Projection) Close() error {
	return op.child.Close()
}

// applyProjection 应用投影转换
func (op *Projection) applyProjection(batch *types.Batch) (*types.Batch, error) {
	record := batch.Record()

	// 构建投影后的schema和数据源信息
	var projectedFields []arrow.Field
	type columnSource struct {
		isExpression bool
		columnIndex  int                  // 用于直接列引用
		expression   optimizer.Expression // 用于表达式计算
		columnRef    optimizer.ColumnRef  // 原始列引用信息
	}
	var sources []columnSource

	for _, projCol := range op.columns {
		if projCol.Type == optimizer.ColumnRefTypeFunction {
			// 处理函数调用类型 (如 UPPER(name), LOWER(name), etc.)
			fieldName := projCol.Column
			if fieldName == "" && projCol.Alias != "" {
				fieldName = projCol.Alias
			}
			if fieldName == "" {
				fieldName = projCol.FunctionName // 使用函数名作为默认名称
			}

			// 函数结果类型根据函数类型推断 (字符串函数返回String)
			var fieldType arrow.DataType = arrow.BinaryTypes.String
			projectedFields = append(projectedFields, arrow.Field{
				Name: fieldName,
				Type: fieldType,
			})
			sources = append(sources, columnSource{
				isExpression: true,
				expression:   nil, // 函数不使用expression字段
				columnRef:    projCol,
			})
		} else if projCol.Type == optimizer.ColumnRefTypeExpression && projCol.Expression != nil {
			// 处理表达式类型 - 计算表达式并添加为新列
			fieldName := projCol.Column
			if fieldName == "" && projCol.Alias != "" {
				fieldName = projCol.Alias
			}
			if fieldName == "" {
				fieldName = "expr" // 默认名称
			}

			// 表达式结果类型推断为 Float64（算术运算结果）
			projectedFields = append(projectedFields, arrow.Field{
				Name: fieldName,
				Type: arrow.PrimitiveTypes.Float64,
			})
			sources = append(sources, columnSource{
				isExpression: true,
				expression:   projCol.Expression,
				columnRef:    projCol,
			})
		} else {
			// 处理普通列引用
			found := false
			for i, field := range record.Schema().Fields() {
				if op.matchesColumn(field.Name, projCol) {
					// 如果有别名，使用别名作为字段名
					if projCol.Alias != "" {
						projectedFields = append(projectedFields, arrow.Field{
							Name: projCol.Alias,
							Type: field.Type,
						})
					} else {
						projectedFields = append(projectedFields, field)
					}
					sources = append(sources, columnSource{
						isExpression: false,
						columnIndex:  i,
						columnRef:    projCol,
					})
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("column not found: %s", op.formatColumnName(projCol))
			}
		}
	}

	// 如果没有任何列，返回空结果
	if len(projectedFields) == 0 {
		return nil, nil
	}

	// 创建新的schema
	projectedSchema := arrow.NewSchema(projectedFields, nil)

	// 创建投影后的记录
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, projectedSchema)
	defer builder.Release()

	// 处理每一行数据
	for rowIdx := int64(0); rowIdx < record.NumRows(); rowIdx++ {
		for fieldIdx, source := range sources {
			field := builder.Field(fieldIdx)

			if source.isExpression {
				// 检查是否是函数调用
				if source.columnRef.Type == optimizer.ColumnRefTypeFunction {
					// 执行函数调用
					result, err := op.evaluateFunction(source.columnRef, record, int(rowIdx))
					if err != nil {
						return nil, fmt.Errorf("failed to evaluate function %s: %w", source.columnRef.FunctionName, err)
					}

					// 函数结果作为String
					if strBuilder, ok := field.(*array.StringBuilder); ok {
						if result == nil {
							strBuilder.AppendNull()
						} else if strVal, ok := result.(string); ok {
							strBuilder.Append(strVal)
						} else {
							strBuilder.Append(fmt.Sprintf("%v", result))
						}
					}
				} else {
					// 计算表达式的值
					value, err := op.evaluateExpression(source.expression, record, int(rowIdx))
					if err != nil {
						return nil, fmt.Errorf("failed to evaluate expression: %w", err)
					}

					// 表达式结果作为 Float64
					if floatBuilder, ok := field.(*array.Float64Builder); ok {
						floatBuilder.Append(value)
					}
				}
			} else {
				// 直接复制列数据
				column := record.Column(source.columnIndex)

				switch col := column.(type) {
				case *array.Int64:
					if intBuilder, ok := field.(*array.Int64Builder); ok {
						if col.IsNull(int(rowIdx)) {
							intBuilder.AppendNull()
						} else {
							intBuilder.Append(col.Value(int(rowIdx)))
						}
					}
				case *array.String:
					if strBuilder, ok := field.(*array.StringBuilder); ok {
						if col.IsNull(int(rowIdx)) {
							strBuilder.AppendNull()
						} else {
							strBuilder.Append(col.Value(int(rowIdx)))
						}
					}
				case *array.Float64:
					if floatBuilder, ok := field.(*array.Float64Builder); ok {
						if col.IsNull(int(rowIdx)) {
							floatBuilder.AppendNull()
						} else {
							floatBuilder.Append(col.Value(int(rowIdx)))
						}
					}
				case *array.Boolean:
					if boolBuilder, ok := field.(*array.BooleanBuilder); ok {
						if col.IsNull(int(rowIdx)) {
							boolBuilder.AppendNull()
						} else {
							boolBuilder.Append(col.Value(int(rowIdx)))
						}
					}
				default:
					// 对于不支持的类型，跳过该行
					continue
				}
			}
		}
	}

	projectedRecord := builder.NewRecord()
	return types.NewBatch(projectedRecord), nil
}

// matchesColumn 检查字段名是否匹配投影列
func (op *Projection) matchesColumn(fieldName string, projCol optimizer.ColumnRef) bool {
	// 直接匹配列名
	if fieldName == projCol.Column {
		return true
	}

	// 如果有表别名，检查表别名.列名格式
	if projCol.Table != "" {
		expectedName := fmt.Sprintf("%s.%s", projCol.Table, projCol.Column)
		if fieldName == expectedName {
			return true
		}
	}

	return false
}

// formatColumnName 格式化列名用于错误信息
func (op *Projection) formatColumnName(col optimizer.ColumnRef) string {
	if col.Table != "" {
		return fmt.Sprintf("%s.%s", col.Table, col.Column)
	}
	return col.Column
}

// evaluateExpression 计算表达式的值
func (op *Projection) evaluateExpression(expr optimizer.Expression, record arrow.Record, rowIdx int) (float64, error) {
	switch e := expr.(type) {
	case *optimizer.BinaryExpression:
		// 递归计算左右操作数
		leftVal, err := op.evaluateExpression(e.Left, record, rowIdx)
		if err != nil {
			return 0, err
		}
		rightVal, err := op.evaluateExpression(e.Right, record, rowIdx)
		if err != nil {
			return 0, err
		}

		// 根据操作符计算结果
		switch e.Operator {
		case "+":
			return leftVal + rightVal, nil
		case "-":
			return leftVal - rightVal, nil
		case "*":
			return leftVal * rightVal, nil
		case "/":
			if rightVal == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return leftVal / rightVal, nil
		default:
			return 0, fmt.Errorf("unsupported operator in expression: %s", e.Operator)
		}

	case *optimizer.ColumnReference:
		// 从record中获取列的值
		colIdx := -1
		for i, field := range record.Schema().Fields() {
			if field.Name == e.Column || (e.Table != "" && field.Name == fmt.Sprintf("%s.%s", e.Table, e.Column)) {
				colIdx = i
				break
			}
		}
		if colIdx == -1 {
			return 0, fmt.Errorf("column not found in expression: %s", e.Column)
		}

		column := record.Column(colIdx)
		switch col := column.(type) {
		case *array.Int64:
			if col.IsNull(rowIdx) {
				return 0, fmt.Errorf("null value in expression")
			}
			return float64(col.Value(rowIdx)), nil
		case *array.Float64:
			if col.IsNull(rowIdx) {
				return 0, fmt.Errorf("null value in expression")
			}
			return col.Value(rowIdx), nil
		case *array.Float32:
			if col.IsNull(rowIdx) {
				return 0, fmt.Errorf("null value in expression")
			}
			return float64(col.Value(rowIdx)), nil
		default:
			return 0, fmt.Errorf("unsupported column type in expression: %T", col)
		}

	case *optimizer.LiteralValue:
		// 字面量值
		switch e.Type {
		case optimizer.LiteralTypeInteger:
			if intVal, ok := e.Value.(int64); ok {
				return float64(intVal), nil
			}
			if intVal, ok := e.Value.(int); ok {
				return float64(intVal), nil
			}
		case optimizer.LiteralTypeFloat:
			if floatVal, ok := e.Value.(float64); ok {
				return floatVal, nil
			}
		}
		return 0, fmt.Errorf("unsupported literal type in expression: %v", e.Type)

	default:
		return 0, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// evaluateFunction 执行函数调用并返回结果
func (op *Projection) evaluateFunction(colRef optimizer.ColumnRef, record arrow.Record, rowIdx int) (interface{}, error) {
	// 解析函数参数（通常是列引用）
	if len(colRef.FunctionArgs) == 0 {
		return nil, fmt.Errorf("function %s requires arguments", colRef.FunctionName)
	}

	// 获取第一个参数的值
	var argValue interface{}
	firstArg := colRef.FunctionArgs[0]

	if colRef, ok := firstArg.(*optimizer.ColumnReference); ok {
		// 从record中获取列的值
		colIdx := -1
		for i, field := range record.Schema().Fields() {
			if field.Name == colRef.Column || (colRef.Table != "" && field.Name == fmt.Sprintf("%s.%s", colRef.Table, colRef.Column)) {
				colIdx = i
				break
			}
		}
		if colIdx == -1 {
			return nil, fmt.Errorf("column not found in function argument: %s", colRef.Column)
		}

		column := record.Column(colIdx)
		switch col := column.(type) {
		case *array.String:
			if col.IsNull(rowIdx) {
				return nil, nil
			}
			argValue = col.Value(rowIdx)
		case *array.Int64:
			if col.IsNull(rowIdx) {
				return nil, nil
			}
			argValue = col.Value(rowIdx)
		case *array.Float64:
			if col.IsNull(rowIdx) {
				return nil, nil
			}
			argValue = col.Value(rowIdx)
		default:
			return nil, fmt.Errorf("unsupported column type for function argument: %T", col)
		}
	} else {
		return nil, fmt.Errorf("unsupported function argument type: %T", firstArg)
	}

	// 执行函数
	switch colRef.FunctionName {
	case "UPPER":
		if strVal, ok := argValue.(string); ok {
			return strings.ToUpper(strVal), nil
		}
		return nil, fmt.Errorf("UPPER requires string argument")
	case "LOWER":
		if strVal, ok := argValue.(string); ok {
			return strings.ToLower(strVal), nil
		}
		return nil, fmt.Errorf("LOWER requires string argument")
	case "LENGTH", "LEN":
		if strVal, ok := argValue.(string); ok {
			return int64(len(strVal)), nil
		}
		return nil, fmt.Errorf("LENGTH requires string argument")
	default:
		return nil, fmt.Errorf("unsupported function: %s", colRef.FunctionName)
	}
}

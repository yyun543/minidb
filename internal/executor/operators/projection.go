package operators

import (
	"fmt"
	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/types"
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

	// 构建投影后的schema
	var projectedFields []arrow.Field
	var columnIndices []int

	for _, projCol := range op.columns {
		// 查找列在原始schema中的位置
		found := false
		for i, field := range record.Schema().Fields() {
			// 匹配列名 (支持表别名.列名格式)
			if op.matchesColumn(field.Name, projCol) {
				projectedFields = append(projectedFields, field)
				columnIndices = append(columnIndices, i)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("column not found: %s", op.formatColumnName(projCol))
		}
	}

	// 如果没有找到任何列，返回空结果
	if len(projectedFields) == 0 {
		return nil, nil
	}

	// 创建新的schema
	projectedSchema := arrow.NewSchema(projectedFields, nil)

	// 创建投影后的记录
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, projectedSchema)
	defer builder.Release()

	// 复制选中列的数据
	for rowIdx := int64(0); rowIdx < record.NumRows(); rowIdx++ {
		for fieldIdx, colIdx := range columnIndices {
			field := builder.Field(fieldIdx)
			column := record.Column(colIdx)

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

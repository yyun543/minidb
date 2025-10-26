package operators

import (
	"fmt"

	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/types"
)

// extractRows 从batch中提取指定范围的行 [startRow, startRow + numRows)
func extractRows(batch *types.Batch, startRow int, numRows int) (*types.Batch, error) {
	if batch == nil {
		return nil, fmt.Errorf("batch is nil")
	}

	record := batch.Record()
	schema := record.Schema()

	// 边界检查
	if startRow < 0 || startRow >= int(record.NumRows()) {
		return nil, fmt.Errorf("startRow %d out of range [0, %d)", startRow, record.NumRows())
	}

	endRow := startRow + numRows
	if endRow > int(record.NumRows()) {
		endRow = int(record.NumRows())
	}

	actualNumRows := endRow - startRow
	if actualNumRows <= 0 {
		// Return empty batch
		return types.NewEmptyBatch(schema, memory.DefaultAllocator), nil
	}

	// 创建新的record builder
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 复制指定范围的行
	for colIdx := 0; colIdx < int(record.NumCols()); colIdx++ {
		column := record.Column(colIdx)
		fieldBuilder := builder.Field(colIdx)

		// 根据列类型复制数据
		switch col := column.(type) {
		case *array.Int64:
			int64Builder := fieldBuilder.(*array.Int64Builder)
			for i := startRow; i < endRow; i++ {
				if col.IsNull(i) {
					int64Builder.AppendNull()
				} else {
					int64Builder.Append(col.Value(i))
				}
			}
		case *array.Float64:
			float64Builder := fieldBuilder.(*array.Float64Builder)
			for i := startRow; i < endRow; i++ {
				if col.IsNull(i) {
					float64Builder.AppendNull()
				} else {
					float64Builder.Append(col.Value(i))
				}
			}
		case *array.String:
			stringBuilder := fieldBuilder.(*array.StringBuilder)
			for i := startRow; i < endRow; i++ {
				if col.IsNull(i) {
					stringBuilder.AppendNull()
				} else {
					stringBuilder.Append(col.Value(i))
				}
			}
		case *array.Boolean:
			boolBuilder := fieldBuilder.(*array.BooleanBuilder)
			for i := startRow; i < endRow; i++ {
				if col.IsNull(i) {
					boolBuilder.AppendNull()
				} else {
					boolBuilder.Append(col.Value(i))
				}
			}
		default:
			return nil, fmt.Errorf("unsupported column type: %T", col)
		}
	}

	newRecord := builder.NewRecord()
	return types.NewBatch(newRecord), nil
}

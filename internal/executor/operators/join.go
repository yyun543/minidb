package operators

import (
	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/types"
)

// Join 连接算子
type Join struct {
	joinType     string               // 连接类型
	condition    optimizer.Expression // 连接条件
	left         Operator             // 左子算子
	right        Operator             // 右子算子
	ctx          interface{}
	leftBatches  []*types.Batch // 缓存左表所有批次
	rightBatches []*types.Batch // 缓存右表所有批次
	initialized  bool           // 是否已初始化
	resultSent   bool           // 是否已发送结果
}

// NewJoin 创建连接算子
func NewJoin(joinType string, condition optimizer.Expression, left, right Operator, ctx interface{}) *Join {
	return &Join{
		joinType:     joinType,
		condition:    condition,
		left:         left,
		right:        right,
		ctx:          ctx,
		leftBatches:  []*types.Batch{},
		rightBatches: []*types.Batch{},
		initialized:  false,
		resultSent:   false,
	}
}

// Init 初始化算子
func (op *Join) Init(ctx interface{}) error {
	if err := op.left.Init(ctx); err != nil {
		return err
	}
	return op.right.Init(ctx)
}

// Next 获取下一批数据
func (op *Join) Next() (*types.Batch, error) {
	// 第一次调用时，缓存所有左右表数据
	if !op.initialized {
		if err := op.cacheAllData(); err != nil {
			return nil, err
		}
		op.initialized = true
	}

	// JOIN算子只返回一次结果
	if op.resultSent {
		return nil, nil
	}

	op.resultSent = true

	// 执行JOIN并构建结果
	return op.buildJoinResult(op.leftBatches, op.rightBatches)
}

// cacheAllData 缓存左右表所有数据
func (op *Join) cacheAllData() error {
	// 缓存左表数据
	for {
		batch, err := op.left.Next()
		if err != nil {
			return err
		}
		if batch == nil {
			break
		}
		op.leftBatches = append(op.leftBatches, batch)
	}

	// 缓存右表数据
	for {
		batch, err := op.right.Next()
		if err != nil {
			return err
		}
		if batch == nil {
			break
		}
		op.rightBatches = append(op.rightBatches, batch)
	}

	return nil
}

// buildJoinResult 构建JOIN结果
func (op *Join) buildJoinResult(leftBatches, rightBatches []*types.Batch) (*types.Batch, error) {
	// 构建结果schema（左表字段 + 右表字段）
	var leftRecord, rightRecord arrow.Record
	if len(leftBatches) > 0 {
		leftRecord = leftBatches[0].Record()
	}
	if len(rightBatches) > 0 {
		rightRecord = rightBatches[0].Record()
	}

	if leftRecord == nil || rightRecord == nil {
		return nil, nil
	}

	// 合并schema
	joinSchema := op.mergeSchemas(leftRecord.Schema(), rightRecord.Schema())

	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, joinSchema)
	defer builder.Release()

	// 嵌套循环JOIN
	for _, leftBatch := range leftBatches {
		leftRec := leftBatch.Record()
		for leftRowIdx := int64(0); leftRowIdx < leftRec.NumRows(); leftRowIdx++ {
			for _, rightBatch := range rightBatches {
				rightRec := rightBatch.Record()
				for rightRowIdx := int64(0); rightRowIdx < rightRec.NumRows(); rightRowIdx++ {
					// 检查JOIN条件
					if op.evaluateJoinCondition(leftRec, leftRowIdx, rightRec, rightRowIdx) {
						// 合并行数据
						op.appendJoinedRow(builder, leftRec, leftRowIdx, rightRec, rightRowIdx)
					}
				}
			}
		}
	}

	joinedRecord := builder.NewRecord()
	if joinedRecord.NumRows() == 0 {
		joinedRecord.Release()
		return nil, nil
	}

	return types.NewBatch(joinedRecord), nil
}

// mergeSchemas 合并左右表的schema
func (op *Join) mergeSchemas(leftSchema, rightSchema *arrow.Schema) *arrow.Schema {
	var fields []arrow.Field

	// 添加左表字段
	for _, field := range leftSchema.Fields() {
		fields = append(fields, field)
	}

	// 添加右表字段
	for _, field := range rightSchema.Fields() {
		fields = append(fields, field)
	}

	return arrow.NewSchema(fields, nil)
}

// evaluateJoinCondition 评估JOIN条件
func (op *Join) evaluateJoinCondition(leftRec arrow.Record, leftRowIdx int64, rightRec arrow.Record, rightRowIdx int64) bool {
	// 简化实现：只支持等值连接 (left.col = right.col)
	if binExpr, ok := op.condition.(*optimizer.BinaryExpression); ok && binExpr.Operator == "=" {
		leftValue := op.getColumnValue(leftRec, leftRowIdx, binExpr.Left)
		rightValue := op.getColumnValue(rightRec, rightRowIdx, binExpr.Right)
		return op.compareValues(leftValue, rightValue)
	}
	return false
}

// getColumnValue 根据表达式获取列值
func (op *Join) getColumnValue(record arrow.Record, rowIdx int64, expr optimizer.Expression) interface{} {
	if colRef, ok := expr.(*optimizer.ColumnReference); ok {
		// 查找列索引
		schema := record.Schema()
		for i, field := range schema.Fields() {
			if field.Name == colRef.Column {
				column := record.Column(i)
				switch col := column.(type) {
				case *array.Int64:
					return col.Value(int(rowIdx))
				case *array.String:
					return col.Value(int(rowIdx))
				case *array.Float64:
					return col.Value(int(rowIdx))
				}
			}
		}
	}
	return nil
}

// compareValues 比较两个值是否相等
func (op *Join) compareValues(left, right interface{}) bool {
	if left == nil || right == nil {
		return false
	}
	return left == right
}

// appendJoinedRow 添加连接后的行到builder
func (op *Join) appendJoinedRow(builder *array.RecordBuilder, leftRec arrow.Record, leftRowIdx int64, rightRec arrow.Record, rightRowIdx int64) {
	fieldIdx := 0

	// 添加左表字段值
	for colIdx := 0; colIdx < int(leftRec.NumCols()); colIdx++ {
		field := builder.Field(fieldIdx)
		column := leftRec.Column(colIdx)

		switch col := column.(type) {
		case *array.Int64:
			if intBuilder, ok := field.(*array.Int64Builder); ok {
				intBuilder.Append(col.Value(int(leftRowIdx)))
			}
		case *array.String:
			if strBuilder, ok := field.(*array.StringBuilder); ok {
				strBuilder.Append(col.Value(int(leftRowIdx)))
			}
		case *array.Float64:
			if floatBuilder, ok := field.(*array.Float64Builder); ok {
				floatBuilder.Append(col.Value(int(leftRowIdx)))
			}
		}
		fieldIdx++
	}

	// 添加右表字段值
	for colIdx := 0; colIdx < int(rightRec.NumCols()); colIdx++ {
		field := builder.Field(fieldIdx)
		column := rightRec.Column(colIdx)

		switch col := column.(type) {
		case *array.Int64:
			if intBuilder, ok := field.(*array.Int64Builder); ok {
				intBuilder.Append(col.Value(int(rightRowIdx)))
			}
		case *array.String:
			if strBuilder, ok := field.(*array.StringBuilder); ok {
				strBuilder.Append(col.Value(int(rightRowIdx)))
			}
		case *array.Float64:
			if floatBuilder, ok := field.(*array.Float64Builder); ok {
				floatBuilder.Append(col.Value(int(rightRowIdx)))
			}
		}
		fieldIdx++
	}
}

// Close 关闭算子
func (op *Join) Close() error {
	if err := op.left.Close(); err != nil {
		return err
	}
	return op.right.Close()
}

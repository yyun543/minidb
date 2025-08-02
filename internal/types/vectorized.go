package types

import (
	"context"
	"fmt"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
)

// VectorizedBatch 向量化批处理，充分利用Arrow的向量化能力
type VectorizedBatch struct {
	schema  *arrow.Schema
	columns []arrow.Array
	length  int64
	pool    memory.Allocator
}

// NewVectorizedBatch 创建向量化批处理
func NewVectorizedBatch(schema *arrow.Schema, pool memory.Allocator) *VectorizedBatch {
	if pool == nil {
		pool = memory.DefaultAllocator
	}
	return &VectorizedBatch{
		schema:  schema,
		columns: make([]arrow.Array, schema.NumFields()),
		pool:    pool,
	}
}

// Schema 获取模式
func (vb *VectorizedBatch) Schema() *arrow.Schema {
	return vb.schema
}

// NumRows 获取行数
func (vb *VectorizedBatch) NumRows() int64 {
	return vb.length
}

// NumColumns 获取列数
func (vb *VectorizedBatch) NumColumns() int64 {
	return int64(len(vb.columns))
}

// Column 获取指定列
func (vb *VectorizedBatch) Column(i int) arrow.Array {
	if i < 0 || i >= len(vb.columns) {
		return nil
	}
	return vb.columns[i]
}

// SetColumn 设置指定列
func (vb *VectorizedBatch) SetColumn(i int, column arrow.Array) error {
	if i < 0 || i >= len(vb.columns) {
		return fmt.Errorf("column index out of range: %d", i)
	}

	// 更新长度（使用第一个非空列的长度）
	if vb.length == 0 && column != nil {
		vb.length = int64(column.Len())
	}

	vb.columns[i] = column
	return nil
}

// ToRecord 转换为Arrow Record
func (vb *VectorizedBatch) ToRecord() arrow.Record {
	arrays := make([]arrow.Array, len(vb.columns))
	for i, col := range vb.columns {
		if col != nil {
			arrays[i] = col.(arrow.Array)
		} else {
			// 如果列为nil，创建一个空列
			field := vb.schema.Field(i)
			arrays[i] = vb.createEmptyColumn(field, int(vb.length))
		}
	}
	return array.NewRecord(vb.schema, arrays, vb.length)
}

// createEmptyColumn 创建一个空列
func (vb *VectorizedBatch) createEmptyColumn(field arrow.Field, length int) arrow.Array {
	pool := memory.DefaultAllocator
	switch field.Type {
	case arrow.PrimitiveTypes.Int64:
		builder := array.NewInt64Builder(pool)
		defer builder.Release()
		for i := 0; i < length; i++ {
			builder.AppendNull()
		}
		return builder.NewArray()
	case arrow.BinaryTypes.String:
		builder := array.NewStringBuilder(pool)
		defer builder.Release()
		for i := 0; i < length; i++ {
			builder.AppendNull()
		}
		return builder.NewArray()
	case arrow.PrimitiveTypes.Float64:
		builder := array.NewFloat64Builder(pool)
		defer builder.Release()
		for i := 0; i < length; i++ {
			builder.AppendNull()
		}
		return builder.NewArray()
	default:
		// 默认创建string类型的空列
		builder := array.NewStringBuilder(pool)
		defer builder.Release()
		for i := 0; i < length; i++ {
			builder.AppendNull()
		}
		return builder.NewArray()
	}
}

// VectorizedOperation 向量化操作接口
type VectorizedOperation interface {
	Execute(batch *VectorizedBatch) (*VectorizedBatch, error)
	Name() string
}

// FilterOperation 向量化过滤操作
type FilterOperation struct {
	predicate *VectorizedPredicate
}

// NewFilterOperation 创建过滤操作
func NewFilterOperation(predicate *VectorizedPredicate) *FilterOperation {
	return &FilterOperation{predicate: predicate}
}

// Execute 执行过滤操作
func (op *FilterOperation) Execute(batch *VectorizedBatch) (*VectorizedBatch, error) {
	// 简化实现：使用基本的行级过滤
	// 在生产环境中应该使用Arrow Compute API进行真正的向量化过滤

	if batch.NumRows() == 0 {
		return batch, nil
	}

	// 创建过滤后的批次
	builder := array.NewRecordBuilder(batch.pool, batch.schema)
	defer builder.Release()

	record := batch.ToRecord()
	// Note: We don't release the record here because it shares arrays with the batch
	// The batch still owns the arrays and will handle their lifecycle

	// 逐行检查过滤条件
	for rowIdx := int64(0); rowIdx < record.NumRows(); rowIdx++ {
		matches, err := op.predicate.EvaluateRow(record, int(rowIdx))
		if err != nil {
			return nil, err
		}

		if matches {
			// 复制这一行到结果中
			for colIdx := 0; colIdx < int(record.NumCols()); colIdx++ {
				field := builder.Field(colIdx)
				column := record.Column(colIdx)

				if err := op.copyValue(field, column, int(rowIdx)); err != nil {
					return nil, fmt.Errorf("failed to copy value: %w", err)
				}
			}
		}
	}

	// 构建结果记录
	filteredRecord := builder.NewRecord()
	// Note: We don't release the record here because the VectorizedBatch will own the columns

	result := NewVectorizedBatch(batch.schema, batch.pool)
	for i := 0; i < int(filteredRecord.NumCols()); i++ {
		if err := result.SetColumn(i, filteredRecord.Column(i)); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Name 返回操作名称
func (op *FilterOperation) Name() string {
	return "VectorizedFilter"
}

// ProjectOperation 向量化投影操作
type ProjectOperation struct {
	columnIndices []int
	newSchema     *arrow.Schema
}

// NewProjectOperation 创建投影操作
func NewProjectOperation(columnIndices []int, newSchema *arrow.Schema) *ProjectOperation {
	return &ProjectOperation{
		columnIndices: columnIndices,
		newSchema:     newSchema,
	}
}

// Execute 执行投影操作
func (op *ProjectOperation) Execute(batch *VectorizedBatch) (*VectorizedBatch, error) {
	result := NewVectorizedBatch(op.newSchema, batch.pool)
	result.length = batch.length

	for i, colIdx := range op.columnIndices {
		if colIdx >= 0 && colIdx < int(batch.NumColumns()) {
			column := batch.Column(colIdx)
			if err := result.SetColumn(i, column); err != nil {
				return nil, fmt.Errorf("failed to set column %d: %w", i, err)
			}
		} else {
			// 如果列索引无效，创建一个空列
			field := op.newSchema.Field(i)
			emptyColumn := op.createEmptyColumn(field, int(batch.length))
			if err := result.SetColumn(i, emptyColumn); err != nil {
				return nil, fmt.Errorf("failed to set empty column %d: %w", i, err)
			}
		}
	}

	return result, nil
}

// Name 返回操作名称
func (op *ProjectOperation) Name() string {
	return "VectorizedProject"
}

// createEmptyColumn 创建一个空列
func (op *ProjectOperation) createEmptyColumn(field arrow.Field, length int) arrow.Array {
	pool := memory.DefaultAllocator
	switch field.Type {
	case arrow.PrimitiveTypes.Int64:
		builder := array.NewInt64Builder(pool)
		defer builder.Release()
		for i := 0; i < length; i++ {
			builder.AppendNull()
		}
		return builder.NewArray()
	case arrow.BinaryTypes.String:
		builder := array.NewStringBuilder(pool)
		defer builder.Release()
		for i := 0; i < length; i++ {
			builder.AppendNull()
		}
		return builder.NewArray()
	case arrow.PrimitiveTypes.Float64:
		builder := array.NewFloat64Builder(pool)
		defer builder.Release()
		for i := 0; i < length; i++ {
			builder.AppendNull()
		}
		return builder.NewArray()
	default:
		// 默认创建string类型的空列
		builder := array.NewStringBuilder(pool)
		defer builder.Release()
		for i := 0; i < length; i++ {
			builder.AppendNull()
		}
		return builder.NewArray()
	}
}

// VectorizedPredicate 向量化谓词
type VectorizedPredicate struct {
	// 简单谓词字段
	columnIndex int
	operator    string
	value       interface{}
	dataType    arrow.DataType

	// 复合谓词字段
	isCompound bool
	logicalOp  string // "AND" 或 "OR"
	leftPred   *VectorizedPredicate
	rightPred  *VectorizedPredicate
}

// NewVectorizedPredicate 创建简单向量化谓词
func NewVectorizedPredicate(columnIndex int, operator string, value interface{}, dataType arrow.DataType) *VectorizedPredicate {
	return &VectorizedPredicate{
		columnIndex: columnIndex,
		operator:    operator,
		value:       value,
		dataType:    dataType,
		isCompound:  false,
	}
}

// NewCompoundVectorizedPredicate 创建复合向量化谓词
func NewCompoundVectorizedPredicate(logicalOp string, left, right *VectorizedPredicate) *VectorizedPredicate {
	return &VectorizedPredicate{
		isCompound: true,
		logicalOp:  logicalOp,
		leftPred:   left,
		rightPred:  right,
	}
}

// EvaluateRow 评估单行谓词
func (vp *VectorizedPredicate) EvaluateRow(record arrow.Record, rowIdx int) (bool, error) {
	// 如果是复合谓词，递归评估左右子谓词
	if vp.isCompound {
		leftResult, err := vp.leftPred.EvaluateRow(record, rowIdx)
		if err != nil {
			return false, err
		}

		rightResult, err := vp.rightPred.EvaluateRow(record, rowIdx)
		if err != nil {
			return false, err
		}

		switch vp.logicalOp {
		case "AND":
			return leftResult && rightResult, nil
		case "OR":
			return leftResult || rightResult, nil
		default:
			return false, fmt.Errorf("unsupported logical operator: %s", vp.logicalOp)
		}
	}

	// 简单谓词的原有逻辑
	if vp.columnIndex < 0 || vp.columnIndex >= int(record.NumCols()) {
		return false, fmt.Errorf("column index %d out of range", vp.columnIndex)
	}

	column := record.Column(vp.columnIndex)
	if column == nil {
		return false, fmt.Errorf("column %d is nil", vp.columnIndex)
	}

	if rowIdx < 0 || rowIdx >= column.Len() {
		return false, fmt.Errorf("row index %d out of range", rowIdx)
	}

	// 获取该行的值
	var cellValue interface{}
	switch col := column.(type) {
	case *array.Int64:
		cellValue = col.Value(rowIdx)
	case *array.Float64:
		cellValue = col.Value(rowIdx)
	case *array.String:
		cellValue = col.Value(rowIdx)
	case *array.Boolean:
		cellValue = col.Value(rowIdx)
	default:
		return false, fmt.Errorf("unsupported column type")
	}

	// 执行比较
	return vp.compareValues(cellValue, vp.value, vp.operator), nil
}

// compareValues 比较两个值
func (vp *VectorizedPredicate) compareValues(a, b interface{}, operator string) bool {
	switch operator {
	case "=", "==":
		return vp.isEqual(a, b)
	case "!=", "<>":
		return !vp.isEqual(a, b)
	case "<":
		return vp.isLess(a, b)
	case "<=":
		return vp.isLess(a, b) || vp.isEqual(a, b)
	case ">":
		return vp.isGreater(a, b)
	case ">=":
		return vp.isGreater(a, b) || vp.isEqual(a, b)
	default:
		return false
	}
}

// isEqual 判断是否相等
func (vp *VectorizedPredicate) isEqual(a, b interface{}) bool {
	switch va := a.(type) {
	case int64:
		if vb, ok := b.(int64); ok {
			return va == vb
		}
		if vb, ok := b.(int); ok {
			return va == int64(vb)
		}
	case float64:
		if vb, ok := b.(float64); ok {
			return va == vb
		}
	case string:
		if vb, ok := b.(string); ok {
			return va == vb
		}
	case bool:
		if vb, ok := b.(bool); ok {
			return va == vb
		}
	}
	return false
}

// isLess 判断是否小于
func (vp *VectorizedPredicate) isLess(a, b interface{}) bool {
	switch va := a.(type) {
	case int64:
		if vb, ok := b.(int64); ok {
			return va < vb
		}
		if vb, ok := b.(int); ok {
			return va < int64(vb)
		}
	case float64:
		if vb, ok := b.(float64); ok {
			return va < vb
		}
	case string:
		if vb, ok := b.(string); ok {
			return va < vb
		}
	}
	return false
}

// isGreater 判断是否大于
func (vp *VectorizedPredicate) isGreater(a, b interface{}) bool {
	switch va := a.(type) {
	case int64:
		if vb, ok := b.(int64); ok {
			return va > vb
		}
		if vb, ok := b.(int); ok {
			return va > int64(vb)
		}
	case float64:
		if vb, ok := b.(float64); ok {
			return va > vb
		}
	case string:
		if vb, ok := b.(string); ok {
			return va > vb
		}
	}
	return false
}

// copyValue 复制值到字段构建器
func (op *FilterOperation) copyValue(field array.Builder, column arrow.Array, rowIdx int) error {
	switch col := column.(type) {
	case *array.Int64:
		if builder, ok := field.(*array.Int64Builder); ok {
			builder.Append(col.Value(rowIdx))
		} else {
			return fmt.Errorf("builder type mismatch for Int64: expected *array.Int64Builder, got %T", field)
		}
	case *array.Float64:
		if builder, ok := field.(*array.Float64Builder); ok {
			builder.Append(col.Value(rowIdx))
		} else {
			return fmt.Errorf("builder type mismatch for Float64: expected *array.Float64Builder, got %T", field)
		}
	case *array.String:
		if builder, ok := field.(*array.StringBuilder); ok {
			builder.Append(col.Value(rowIdx))
		} else {
			return fmt.Errorf("builder type mismatch for String: expected *array.StringBuilder, got %T", field)
		}
	case *array.Boolean:
		if builder, ok := field.(*array.BooleanBuilder); ok {
			builder.Append(col.Value(rowIdx))
		} else {
			return fmt.Errorf("builder type mismatch for Boolean: expected *array.BooleanBuilder, got %T", field)
		}
	default:
		return fmt.Errorf("unsupported column type: %T", column)
	}
	return nil
}

// VectorizedAggregation 向量化聚合操作（简化版）
type VectorizedAggregation struct {
	function    string // COUNT, SUM, AVG, MIN, MAX
	columnIndex int
	dataType    arrow.DataType
}

// NewVectorizedAggregation 创建向量化聚合
func NewVectorizedAggregation(function string, columnIndex int, dataType arrow.DataType) *VectorizedAggregation {
	return &VectorizedAggregation{
		function:    function,
		columnIndex: columnIndex,
		dataType:    dataType,
	}
}

// Execute 执行聚合操作（简化实现）
func (va *VectorizedAggregation) Execute(ctx context.Context, batch *VectorizedBatch) (interface{}, error) {
	column := batch.Column(va.columnIndex)
	if column == nil {
		return nil, fmt.Errorf("column %d is nil", va.columnIndex)
	}

	// 简化的聚合实现
	switch va.function {
	case "COUNT":
		return int64(column.Len()), nil
	case "SUM":
		return va.calculateSum(column)
	case "MIN":
		return va.calculateMin(column)
	case "MAX":
		return va.calculateMax(column)
	default:
		return nil, fmt.Errorf("unsupported aggregation function: %s", va.function)
	}
}

// calculateSum 计算求和
func (va *VectorizedAggregation) calculateSum(column arrow.Array) (interface{}, error) {
	switch col := column.(type) {
	case *array.Int64:
		var sum int64 = 0
		for i := 0; i < col.Len(); i++ {
			if !col.IsNull(i) {
				sum += col.Value(i)
			}
		}
		return sum, nil
	case *array.Float64:
		var sum float64 = 0
		for i := 0; i < col.Len(); i++ {
			if !col.IsNull(i) {
				sum += col.Value(i)
			}
		}
		return sum, nil
	default:
		return nil, fmt.Errorf("SUM not supported for column type")
	}
}

// calculateMin 计算最小值
func (va *VectorizedAggregation) calculateMin(column arrow.Array) (interface{}, error) {
	switch col := column.(type) {
	case *array.Int64:
		if col.Len() == 0 {
			return nil, nil
		}
		var min int64 = col.Value(0)
		for i := 1; i < col.Len(); i++ {
			if !col.IsNull(i) && col.Value(i) < min {
				min = col.Value(i)
			}
		}
		return min, nil
	case *array.Float64:
		if col.Len() == 0 {
			return nil, nil
		}
		var min float64 = col.Value(0)
		for i := 1; i < col.Len(); i++ {
			if !col.IsNull(i) && col.Value(i) < min {
				min = col.Value(i)
			}
		}
		return min, nil
	default:
		return nil, fmt.Errorf("MIN not supported for column type")
	}
}

// calculateMax 计算最大值
func (va *VectorizedAggregation) calculateMax(column arrow.Array) (interface{}, error) {
	switch col := column.(type) {
	case *array.Int64:
		if col.Len() == 0 {
			return nil, nil
		}
		var max int64 = col.Value(0)
		for i := 1; i < col.Len(); i++ {
			if !col.IsNull(i) && col.Value(i) > max {
				max = col.Value(i)
			}
		}
		return max, nil
	case *array.Float64:
		if col.Len() == 0 {
			return nil, nil
		}
		var max float64 = col.Value(0)
		for i := 1; i < col.Len(); i++ {
			if !col.IsNull(i) && col.Value(i) > max {
				max = col.Value(i)
			}
		}
		return max, nil
	default:
		return nil, fmt.Errorf("MAX not supported for column type")
	}
}

// Name 返回聚合函数名称
func (va *VectorizedAggregation) Name() string {
	return fmt.Sprintf("VectorizedAggregation_%s", va.function)
}

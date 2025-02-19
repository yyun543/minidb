package types

import (
	"fmt"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
)

// Batch 表示一批数据
type Batch struct {
	record arrow.Record
	pool   memory.Allocator
	schema *arrow.Schema
	cols   []arrow.Array
}

// NewBatch 创建新的数据批次
func NewBatch(schema *arrow.Schema, capacity int, pool memory.Allocator) *Batch {
	if pool == nil {
		pool = memory.DefaultAllocator
	}
	return &Batch{
		pool:   pool,
		schema: schema,
		cols:   make([]arrow.Array, schema.NumFields()),
	}
}

// AddColumn 添加列数据
func (b *Batch) AddColumn(name string, data interface{}) error {
	if b.schema == nil {
		return fmt.Errorf("schema is not initialized")
	}

	fieldIdx := b.schema.FieldIndices(name)
	if len(fieldIdx) == 0 {
		return fmt.Errorf("column %s not found in schema", name)
	}

	idx := fieldIdx[0]
	field := b.schema.Field(idx)

	var arr arrow.Array
	switch v := data.(type) {
	case []int64:
		builder := array.NewInt64Builder(b.pool)
		defer builder.Release()
		builder.AppendValues(v, nil)
		arr = builder.NewArray()
	case []float64:
		builder := array.NewFloat64Builder(b.pool)
		defer builder.Release()
		builder.AppendValues(v, nil)
		arr = builder.NewArray()
	case []string:
		builder := array.NewStringBuilder(b.pool)
		defer builder.Release()
		builder.AppendValues(v, nil)
		arr = builder.NewArray()
	case []bool:
		builder := array.NewBooleanBuilder(b.pool)
		defer builder.Release()
		builder.AppendValues(v, nil)
		arr = builder.NewArray()
	default:
		return fmt.Errorf("unsupported data type %T for column %s", data, name)
	}

	if !arrow.TypeEqual(arr.DataType(), field.Type) {
		arr.Release()
		return fmt.Errorf("data type mismatch for column %s: expected %s, got %s",
			name, field.Type, arr.DataType())
	}

	if b.cols[idx] != nil {
		b.cols[idx].Release()
	}
	b.cols[idx] = arr

	// 检查是否所有列都已添加
	complete := true
	for _, col := range b.cols {
		if col == nil {
			complete = false
			break
		}
	}

	// 如果之前的 record 存在，释放它
	if b.record != nil {
		b.record.Release()
	}

	// 创建新的 record
	if complete {
		b.record = array.NewRecord(b.schema, b.cols, int64(arr.Len()))
	}

	return nil
}

// Values 获取指定行的所有列值
func (b *Batch) Values(row int) ([]interface{}, error) {
	if b.record == nil {
		return nil, fmt.Errorf("record is not initialized")
	}
	if row < 0 || row >= int(b.record.NumRows()) {
		return nil, fmt.Errorf("row index out of range")
	}

	result := make([]interface{}, b.record.NumCols())
	for i := 0; i < int(b.record.NumCols()); i++ {
		col := b.record.Column(i)
		switch col.DataType().ID() {
		case arrow.STRING:
			result[i] = col.(*array.String).Value(row)
		case arrow.INT64:
			result[i] = col.(*array.Int64).Value(row)
		case arrow.FLOAT64:
			result[i] = col.(*array.Float64).Value(row)
		case arrow.BOOL:
			result[i] = col.(*array.Boolean).Value(row)
		default:
			return nil, fmt.Errorf("unsupported data type for column %d", i)
		}
	}
	return result, nil
}

// GetString returns the string value at the given column and row index
func (b *Batch) GetString(col, row int) (string, error) {
	if b.record == nil {
		return "", fmt.Errorf("record is not initialized")
	}
	if col < 0 || col >= int(b.record.NumCols()) {
		return "", fmt.Errorf("column index out of range")
	}
	if row < 0 || row >= int(b.record.NumRows()) {
		return "", fmt.Errorf("row index out of range")
	}

	column := b.record.Column(col)
	if column.DataType().ID() != arrow.STRING {
		return "", fmt.Errorf("column %d is not of string type", col)
	}
	return column.(*array.String).Value(row), nil
}

// Release 释放资源
func (b *Batch) Release() {
	if b.record != nil {
		b.record.Release()
	}
	for _, col := range b.cols {
		if col != nil {
			col.Release()
		}
	}
}

// MapSQLTypeToArrowType converts SQL type to arrow.DataType
func MapSQLTypeToArrowType(sqlType string) arrow.DataType {
	switch sqlType {
	case "INT":
		return arrow.PrimitiveTypes.Int64
	case "FLOAT":
		return arrow.PrimitiveTypes.Float64
	case "VARCHAR", "TEXT":
		return arrow.BinaryTypes.String
	case "BOOL":
		return arrow.FixedWidthTypes.Boolean
	default:
		return arrow.BinaryTypes.String
	}
}

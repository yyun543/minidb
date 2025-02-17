package types

import (
	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
)

// Batch 表示一批数据
type Batch struct {
	record arrow.Record
	pool   *memory.GoAllocator
}

// NewBatch 创建新的数据批次
func NewBatch(schema *arrow.Schema, capacity int) *Batch {
	return &Batch{
		pool: memory.NewGoAllocator(),
	}
}

// AddColumn 添加列数据
func (b *Batch) AddColumn(name string, data interface{}) error {
	switch v := data.(type) {
	case []int64:
		builder := array.NewInt64Builder(b.pool)
		defer builder.Release()
		for _, val := range v {
			builder.Append(val)
		}
	case []string:
		builder := array.NewStringBuilder(b.pool)
		defer builder.Release()
		for _, val := range v {
			builder.Append(val)
		}
	}
	return nil
}

// Values 获取当前行的所有列值
func (b *Batch) Values() []interface{} {
	if b.record == nil {
		return nil
	}

	result := make([]interface{}, b.record.NumCols())
	for i := int64(0); i < b.record.NumCols(); i++ {
		col := b.record.Column(int(i))
		result[i] = col.(*array.String).Value(0)
	}
	return result
}

// Release 释放资源
func (b *Batch) Release() {
	if b.record != nil {
		b.record.Release()
	}
}

// GetString returns the string value at the given column and row index
func (b *Batch) GetString(col, row int) string {
	column := b.record.Column(col)
	return column.(*array.String).Value(row)
}

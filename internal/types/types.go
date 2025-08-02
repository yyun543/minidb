package types

import (
	"bytes"
	"fmt"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/ipc"
	"github.com/apache/arrow/go/v18/arrow/memory"
)

/**
* 核心设计思路：
*
* Chunk（数据块）- 存储层面的概念：
* - 表示存储引擎中的数据分片，对应一个Key-Value存储项
* - 有固定的ID和大小限制（默认4KB），用于数据持久化
* - 一个表由多个Chunk组成，每个Chunk包含部分行数据
*
* Batch（数据批次）- 执行层面的概念：
* - 表示查询执行过程中的数据批次，用于内存中的数据处理
* - 可能包含多个Chunk的数据或单个Chunk的部分数据
* - 用于算子之间的数据传递
 */

// Chunk 表示存储层面的数据块，对应存储引擎中的一个数据分片
type Chunk struct {
	record arrow.Record // 底层Arrow记录
	id     int64        // 数据块ID，用于存储定位
	size   int64        // 数据块大小，单位为字节
}

// Batch 表示执行层面的数据批次，用于算子间的数据传递
type Batch struct {
	record arrow.Record // 底层Arrow记录
}

// NewChunk 创建新的数据块
func NewChunk(record arrow.Record, id int64, size int64) *Chunk {
	record.Retain() // 增加引用计数
	return &Chunk{record: record, id: id, size: size}
}

// NewEmptyChunk 创建空的数据块
func NewEmptyChunk(schema *arrow.Schema, pool memory.Allocator) *Chunk {
	if pool == nil {
		pool = memory.DefaultAllocator
	}
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()
	record := builder.NewRecord()
	return &Chunk{record: record, id: 0, size: 0}
}

// NewBatch 创建新的数据批次
func NewBatch(record arrow.Record) *Batch {
	record.Retain() // 增加引用计数
	return &Batch{record: record}
}

// NewEmptyBatch 创建空的数据批次
func NewEmptyBatch(schema *arrow.Schema, pool memory.Allocator) *Batch {
	if pool == nil {
		pool = memory.DefaultAllocator
	}
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()
	record := builder.NewRecord()
	return &Batch{record: record}
}

// Chunk 方法
func (c *Chunk) Schema() *arrow.Schema {
	return c.record.Schema()
}

func (c *Chunk) NumRows() int64 {
	return c.record.NumRows()
}

func (c *Chunk) NumCols() int64 {
	return c.record.NumCols()
}

func (c *Chunk) Column(i int) arrow.Array {
	return c.record.Column(i)
}

func (c *Chunk) ColumnByName(name string) (arrow.Array, error) {
	idx := c.Schema().FieldIndices(name)
	if len(idx) == 0 {
		return nil, fmt.Errorf("column %s not found", name)
	}
	return c.Column(idx[0]), nil
}

func (c *Chunk) Release() {
	if c.record != nil {
		c.record.Release()
		c.record = nil
	}
}

func (c *Chunk) Record() arrow.Record {
	return c.record
}

func (c *Chunk) ID() int64 {
	return c.id
}

func (c *Chunk) Size() int64 {
	return c.size
}

// Batch 方法
func (b *Batch) Schema() *arrow.Schema {
	return b.record.Schema()
}

func (b *Batch) NumRows() int64 {
	return b.record.NumRows()
}

func (b *Batch) NumCols() int64 {
	return b.record.NumCols()
}

func (b *Batch) Column(i int) arrow.Array {
	return b.record.Column(i)
}

func (b *Batch) ColumnByName(name string) (arrow.Array, error) {
	idx := b.Schema().FieldIndices(name)
	if len(idx) == 0 {
		return nil, fmt.Errorf("column %s not found", name)
	}
	return b.Column(idx[0]), nil
}

func (b *Batch) Release() {
	if b.record != nil {
		b.record.Release()
		b.record = nil
	}
}

func (b *Batch) Record() arrow.Record {
	return b.record
}

// Values 返回第一行的所有列值，用于测试
func (b *Batch) Values() []interface{} {
	if b.record.NumRows() == 0 {
		return nil
	}

	values := make([]interface{}, b.record.NumCols())
	for i := int64(0); i < b.record.NumCols(); i++ {
		col := b.record.Column(int(i))
		switch col := col.(type) {
		case *array.Int64:
			values[i] = col.Value(0)
		case *array.Float64:
			values[i] = col.Value(0)
		case *array.String:
			values[i] = col.Value(0)
		case *array.Boolean:
			values[i] = col.Value(0)
		default:
			values[i] = nil
		}
	}
	return values
}

// GetString 获取指定位置的字符串值，用于测试
func (b *Batch) GetString(col, row int) string {
	if row >= int(b.record.NumRows()) || col >= int(b.record.NumCols()) {
		return ""
	}
	column := b.record.Column(col)
	if strCol, ok := column.(*array.String); ok {
		return strCol.Value(row)
	}
	return ""
}

// ChunkBuilder 用于构建数据块
type ChunkBuilder struct {
	schema  *arrow.Schema
	pool    memory.Allocator
	builder *array.RecordBuilder
}

// NewChunkBuilder 创建新的数据块构建器
func NewChunkBuilder(schema *arrow.Schema, pool memory.Allocator) *ChunkBuilder {
	if pool == nil {
		pool = memory.DefaultAllocator
	}
	return &ChunkBuilder{
		schema:  schema,
		pool:    pool,
		builder: array.NewRecordBuilder(pool, schema),
	}
}

// BatchBuilder 用于构建数据批次
type BatchBuilder struct {
	schema  *arrow.Schema
	pool    memory.Allocator
	builder *array.RecordBuilder
}

// NewBatchBuilder 创建新的数据批次构建器
func NewBatchBuilder(schema *arrow.Schema, pool memory.Allocator) *BatchBuilder {
	if pool == nil {
		pool = memory.DefaultAllocator
	}
	return &BatchBuilder{
		schema:  schema,
		pool:    pool,
		builder: array.NewRecordBuilder(pool, schema),
	}
}

// AppendValue 添加值到指定列（ChunkBuilder）
func (cb *ChunkBuilder) AppendValue(colIdx int, value interface{}) error {
	return appendValue(cb.builder, colIdx, value)
}

// AppendValue 添加值到指定列（BatchBuilder）
func (bb *BatchBuilder) AppendValue(colIdx int, value interface{}) error {
	return appendValue(bb.builder, colIdx, value)
}

// appendValue 通用的添加值方法
func appendValue(builder *array.RecordBuilder, colIdx int, value interface{}) error {
	field := builder.Field(colIdx)
	switch field := field.(type) {
	case *array.Int64Builder:
		if v, ok := value.(int64); ok {
			field.Append(v)
			return nil
		}
	case *array.Float64Builder:
		if v, ok := value.(float64); ok {
			field.Append(v)
			return nil
		}
	case *array.StringBuilder:
		if v, ok := value.(string); ok {
			field.Append(v)
			return nil
		}
	case *array.BooleanBuilder:
		if v, ok := value.(bool); ok {
			field.Append(v)
			return nil
		}
	}
	return fmt.Errorf("unsupported value type for column %d", colIdx)
}

// Build 从构建器创建新的数据块
func (cb *ChunkBuilder) Build(id int64, size int64) *Chunk {
	record := cb.builder.NewRecord()
	return NewChunk(record, id, size)
}

// Build 从构建器创建新的数据批次
func (bb *BatchBuilder) Build() *Batch {
	record := bb.builder.NewRecord()
	return NewBatch(record)
}

// Release 释放构建器使用的资源
func (cb *ChunkBuilder) Release() {
	if cb.builder != nil {
		cb.builder.Release()
		cb.builder = nil
	}
}

// Release 释放构建器使用的资源
func (bb *BatchBuilder) Release() {
	if bb.builder != nil {
		bb.builder.Release()
		bb.builder = nil
	}
}

// SerializeSchema 序列化 schema
func SerializeSchema(schema *arrow.Schema) ([]byte, error) {
	var buf bytes.Buffer
	// 创建 writer 并立即关闭以写入 schema
	writer := ipc.NewWriter(&buf, ipc.WithSchema(schema))
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to serialize schema: %w", err)
	}
	return buf.Bytes(), nil
}

// DeserializeSchema 反序列化 schema
func DeserializeSchema(data []byte) (*arrow.Schema, error) {
	reader, err := ipc.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	schema := reader.Schema()
	return schema, nil
}

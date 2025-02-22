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
* Arrow Record和存储是通过动态的Chunk分割和元数据更新的数据读写的。
* 一个arrow.Record对应一个chunk，多个chunk构成的chunks  []*types.Chunk来表示一个表，
* 一个Chunk包含小于4KB的数据行（需要用storage.KeyManager.CalculateChunkSize根据表的Schema来计算出当前表的一个Chunk最大行数），
* 一个Chunk对应storage存储引擎一个Key-Value。
* 一个表的元数据包括：表的Schema，表的Chunk个数，每个Chunk的ID。
* 一个Chunk的元数据包括：Chunk的ID，Chunk的行数，Chunk的Schema。
* 一个Chunk的数据包括：Chunk的行数据。
* 一个表的元数据和Chunk的元数据存储在storage存储引擎中，Chunk的数据存储在storage存储引擎中。
* 一个表的元数据和Chunk的元数据通过Key-Value的形式存储在storage存储引擎中。
* 一个表的Chunk的数据通过Key-Value的形式存储在storage存储引擎中。
 */

// Chunk 表示从一个Chunks表分片出的一批数据，是对 arrow.Record 的封装
type Chunk struct {
	chunk     arrow.Record
	chunkId   int64 // Chunk 的序号，从 0 开始
	chunkSize int64 // Chunk 的大小，单位为字节
}

// NewChunk 从 arrow.Record 创建新的数据批次
func NewChunk(chunk arrow.Record, chunkId int64, chunkSize int64) *Chunk {
	// 增加引用计数，确保数据不会被过早释放
	chunk.Retain()
	return &Chunk{chunk: chunk, chunkId: chunkId, chunkSize: chunkSize}
}

// NewEmptyChunk 创建空的数据批次
func NewEmptyChunk(schema *arrow.Schema, pool memory.Allocator) *Chunk {
	if pool == nil {
		pool = memory.DefaultAllocator
	}

	// 创建空记录
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	chunk := builder.NewRecord()
	return &Chunk{chunk: chunk, chunkId: 0, chunkSize: 0}
}

// Schema 返回批次的 schema
func (b *Chunk) Schema() *arrow.Schema {
	return b.chunk.Schema()
}

// NumRows 返回行数
func (b *Chunk) NumRows() int64 {
	return b.chunk.NumRows()
}

// NumCols 返回列数
func (b *Chunk) NumCols() int64 {
	return b.chunk.NumCols()
}

// Column 返回指定索引的列
func (b *Chunk) Column(i int) arrow.Array {
	return b.chunk.Column(i)
}

// ColumnByName 返回指定名称的列
func (b *Chunk) ColumnByName(name string) (arrow.Array, error) {
	idx := b.Schema().FieldIndices(name)
	if len(idx) == 0 {
		return nil, fmt.Errorf("column %s not found", name)
	}
	return b.Column(idx[0]), nil
}

// Release 释放底层 arrow.Record 的内存
func (b *Chunk) Release() {
	if b.chunk != nil {
		b.chunk.Release()
		b.chunk = nil
	}
}

// Record 返回底层的 arrow.Record
// 注意：调用者不应该修改返回的 Record
func (b *Chunk) Record() arrow.Record {
	return b.chunk
}

// Slice 返回指定范围的数据切片
func (b *Chunk) Slice(offset, length int64) *Chunk {
	return NewChunk(b.chunk.NewSlice(offset, length), b.chunkId, b.chunkSize)
}

// ChunkBuilder 创建用于构建新块的 Builder
type ChunkBuilder struct {
	schema  *arrow.Schema
	pool    memory.Allocator
	builder *array.RecordBuilder
}

// NewChunkBuilder 创建新的批次构建器
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

// AppendValue 添加值到指定列
func (cb *ChunkBuilder) AppendValue(colIdx int, value interface{}) error {
	field := cb.builder.Field(colIdx)
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

// NewChunk 从构建器创建新的批次
func (cb *ChunkBuilder) NewChunk(chunkId int64, chunkSize int64) *Chunk {
	record := cb.builder.NewRecord()
	return NewChunk(record, chunkId, chunkSize)
}

// Release 释放构建器使用的资源
func (cb *ChunkBuilder) Release() {
	if cb.builder != nil {
		cb.builder.Release()
		cb.builder = nil
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

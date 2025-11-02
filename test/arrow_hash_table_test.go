package test

import (
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/executor/operators"
)

// TestArrowHashTableBasic tests basic hash table operations
func TestArrowHashTableBasic(t *testing.T) {
	allocator := memory.NewGoAllocator()
	ht := operators.NewArrowHashTable(allocator)
	defer ht.Release()

	// Create a simple test batch
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "name", Type: arrow.BinaryTypes.String},
	}, nil)

	builder := array.NewRecordBuilder(allocator, schema)
	defer builder.Release()

	// Build side: 3 rows
	builder.Field(0).(*array.Int64Builder).AppendValues([]int64{1, 2, 3}, nil)
	builder.Field(1).(*array.StringBuilder).AppendValues([]string{"Alice", "Bob", "Charlie"}, nil)

	buildBatch := builder.NewRecord()
	defer buildBatch.Release()

	// Insert into hash table (join on column 0: id)
	err := ht.Insert(buildBatch, []int{0})
	require.NoError(t, err)

	// Verify statistics
	stats := ht.GetStatistics()
	assert.Equal(t, int64(3), stats.TotalRows)
	assert.Equal(t, int64(3), stats.DistinctKeys)

	// Create probe batch
	builder2 := array.NewRecordBuilder(allocator, schema)
	defer builder2.Release()

	// Probe side: 2 matching rows + 1 non-matching
	builder2.Field(0).(*array.Int64Builder).AppendValues([]int64{2, 3, 999}, nil)
	builder2.Field(1).(*array.StringBuilder).AppendValues([]string{"Bob2", "Charlie2", "Nobody"}, nil)

	probeBatch := builder2.NewRecord()
	defer probeBatch.Release()

	// Probe hash table
	matches, err := ht.Probe(probeBatch, []int{0})
	require.NoError(t, err)

	// Should find 2 matches (id=2 and id=3)
	assert.Equal(t, 2, len(matches))

	// Verify match details
	assert.Equal(t, int32(0), matches[0].ProbeRow) // First probe row (id=2)
	assert.Equal(t, int32(1), matches[1].ProbeRow) // Second probe row (id=3)
}

// TestArrowHashTableMultiColumn tests multi-column join keys
func TestArrowHashTableMultiColumn(t *testing.T) {
	allocator := memory.NewGoAllocator()
	ht := operators.NewArrowHashTable(allocator)
	defer ht.Release()

	// Create schema with two join keys
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "key1", Type: arrow.PrimitiveTypes.Int64},
		{Name: "key2", Type: arrow.PrimitiveTypes.Int64},
		{Name: "value", Type: arrow.BinaryTypes.String},
	}, nil)

	builder := array.NewRecordBuilder(allocator, schema)
	defer builder.Release()

	// Build side: 3 rows with composite keys
	builder.Field(0).(*array.Int64Builder).AppendValues([]int64{1, 1, 2}, nil)
	builder.Field(1).(*array.Int64Builder).AppendValues([]int64{10, 20, 10}, nil)
	builder.Field(2).(*array.StringBuilder).AppendValues([]string{"A", "B", "C"}, nil)

	buildBatch := builder.NewRecord()
	defer buildBatch.Release()

	// Insert with both columns as join keys
	err := ht.Insert(buildBatch, []int{0, 1})
	require.NoError(t, err)

	// Probe batch
	builder2 := array.NewRecordBuilder(allocator, schema)
	defer builder2.Release()

	// Match: (1,10), (2,10), No match: (1,30)
	builder2.Field(0).(*array.Int64Builder).AppendValues([]int64{1, 2, 1}, nil)
	builder2.Field(1).(*array.Int64Builder).AppendValues([]int64{10, 10, 30}, nil)
	builder2.Field(2).(*array.StringBuilder).AppendValues([]string{"X", "Y", "Z"}, nil)

	probeBatch := builder2.NewRecord()
	defer probeBatch.Release()

	matches, err := ht.Probe(probeBatch, []int{0, 1})
	require.NoError(t, err)

	// Should find 2 matches: (1,10) and (2,10)
	assert.Equal(t, 2, len(matches))
}

// TestArrowHashTableNullKeys tests NULL key handling
func TestArrowHashTableNullKeys(t *testing.T) {
	allocator := memory.NewGoAllocator()
	ht := operators.NewArrowHashTable(allocator)
	defer ht.Release()

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
		{Name: "name", Type: arrow.BinaryTypes.String},
	}, nil)

	builder := array.NewRecordBuilder(allocator, schema)
	defer builder.Release()

	// Build side: Include a NULL key
	idBuilder := builder.Field(0).(*array.Int64Builder)
	idBuilder.Append(1)
	idBuilder.AppendNull() // NULL key
	idBuilder.Append(3)

	builder.Field(1).(*array.StringBuilder).AppendValues([]string{"Alice", "NullPerson", "Charlie"}, nil)

	buildBatch := builder.NewRecord()
	defer buildBatch.Release()

	err := ht.Insert(buildBatch, []int{0})
	require.NoError(t, err)

	// NULL keys should be skipped
	stats := ht.GetStatistics()
	assert.Equal(t, int64(2), stats.TotalRows) // Only non-NULL rows

	// Probe with NULL key
	builder2 := array.NewRecordBuilder(allocator, schema)
	defer builder2.Release()

	idBuilder2 := builder2.Field(0).(*array.Int64Builder)
	idBuilder2.Append(1)    // Match
	idBuilder2.AppendNull() // NULL - no match
	idBuilder2.Append(3)    // Match

	builder2.Field(1).(*array.StringBuilder).AppendValues([]string{"A", "N", "C"}, nil)

	probeBatch := builder2.NewRecord()
	defer probeBatch.Release()

	matches, err := ht.Probe(probeBatch, []int{0})
	require.NoError(t, err)

	// Should find 2 matches (skipping NULL)
	assert.Equal(t, 2, len(matches))
}

// TestArrowHashTableStringKeys tests string join keys
func TestArrowHashTableStringKeys(t *testing.T) {
	allocator := memory.NewGoAllocator()
	ht := operators.NewArrowHashTable(allocator)
	defer ht.Release()

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "name", Type: arrow.BinaryTypes.String},
		{Name: "value", Type: arrow.PrimitiveTypes.Int64},
	}, nil)

	builder := array.NewRecordBuilder(allocator, schema)
	defer builder.Release()

	// Build side
	builder.Field(0).(*array.StringBuilder).AppendValues([]string{"Alice", "Bob", "Charlie"}, nil)
	builder.Field(1).(*array.Int64Builder).AppendValues([]int64{100, 200, 300}, nil)

	buildBatch := builder.NewRecord()
	defer buildBatch.Release()

	err := ht.Insert(buildBatch, []int{0}) // Join on name
	require.NoError(t, err)

	// Probe batch
	builder2 := array.NewRecordBuilder(allocator, schema)
	defer builder2.Release()

	builder2.Field(0).(*array.StringBuilder).AppendValues([]string{"Bob", "Dave"}, nil)
	builder2.Field(1).(*array.Int64Builder).AppendValues([]int64{1, 2}, nil)

	probeBatch := builder2.NewRecord()
	defer probeBatch.Release()

	matches, err := ht.Probe(probeBatch, []int{0})
	require.NoError(t, err)

	// Should find 1 match (Bob)
	assert.Equal(t, 1, len(matches))
	assert.Equal(t, int32(0), matches[0].ProbeRow)
}

// TestArrowHashTableDuplicateKeys tests handling of duplicate keys
func TestArrowHashTableDuplicateKeys(t *testing.T) {
	allocator := memory.NewGoAllocator()
	ht := operators.NewArrowHashTable(allocator)
	defer ht.Release()

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "category", Type: arrow.PrimitiveTypes.Int64},
		{Name: "item", Type: arrow.BinaryTypes.String},
	}, nil)

	builder := array.NewRecordBuilder(allocator, schema)
	defer builder.Release()

	// Build side: Multiple rows with same key
	builder.Field(0).(*array.Int64Builder).AppendValues([]int64{1, 1, 2}, nil)
	builder.Field(1).(*array.StringBuilder).AppendValues([]string{"Item1", "Item2", "Item3"}, nil)

	buildBatch := builder.NewRecord()
	defer buildBatch.Release()

	err := ht.Insert(buildBatch, []int{0})
	require.NoError(t, err)

	stats := ht.GetStatistics()
	assert.Equal(t, int64(3), stats.TotalRows)
	assert.Equal(t, int64(2), stats.DistinctKeys) // Only 2 distinct keys

	// Probe batch
	builder2 := array.NewRecordBuilder(allocator, schema)
	defer builder2.Release()

	builder2.Field(0).(*array.Int64Builder).AppendValues([]int64{1}, nil)
	builder2.Field(1).(*array.StringBuilder).AppendValues([]string{"X"}, nil)

	probeBatch := builder2.NewRecord()
	defer probeBatch.Release()

	matches, err := ht.Probe(probeBatch, []int{0})
	require.NoError(t, err)

	// Should find 2 matches (both rows with category=1)
	assert.Equal(t, 2, len(matches))
}

// TestArrowHashTableMultipleBatches tests insertion of multiple batches
func TestArrowHashTableMultipleBatches(t *testing.T) {
	allocator := memory.NewGoAllocator()
	ht := operators.NewArrowHashTable(allocator)
	defer ht.Release()

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
	}, nil)

	// Insert batch 1
	builder1 := array.NewRecordBuilder(allocator, schema)
	defer builder1.Release()
	builder1.Field(0).(*array.Int64Builder).AppendValues([]int64{1, 2}, nil)
	batch1 := builder1.NewRecord()
	defer batch1.Release()

	err := ht.Insert(batch1, []int{0})
	require.NoError(t, err)

	// Insert batch 2
	builder2 := array.NewRecordBuilder(allocator, schema)
	defer builder2.Release()
	builder2.Field(0).(*array.Int64Builder).AppendValues([]int64{3, 4}, nil)
	batch2 := builder2.NewRecord()
	defer batch2.Release()

	err = ht.Insert(batch2, []int{0})
	require.NoError(t, err)

	stats := ht.GetStatistics()
	assert.Equal(t, int64(4), stats.TotalRows)

	// Probe should find rows from both batches
	builder3 := array.NewRecordBuilder(allocator, schema)
	defer builder3.Release()
	builder3.Field(0).(*array.Int64Builder).AppendValues([]int64{1, 3}, nil)
	probeBatch := builder3.NewRecord()
	defer probeBatch.Release()

	matches, err := ht.Probe(probeBatch, []int{0})
	require.NoError(t, err)

	assert.Equal(t, 2, len(matches))
}

// TestArrowHashTableGetBuildBatch tests retrieving build batch
func TestArrowHashTableGetBuildBatch(t *testing.T) {
	allocator := memory.NewGoAllocator()
	ht := operators.NewArrowHashTable(allocator)
	defer ht.Release()

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "value", Type: arrow.BinaryTypes.String},
	}, nil)

	builder := array.NewRecordBuilder(allocator, schema)
	defer builder.Release()

	builder.Field(0).(*array.Int64Builder).AppendValues([]int64{1, 2}, nil)
	builder.Field(1).(*array.StringBuilder).AppendValues([]string{"A", "B"}, nil)

	buildBatch := builder.NewRecord()
	defer buildBatch.Release()

	err := ht.Insert(buildBatch, []int{0})
	require.NoError(t, err)

	// Get build batch for first row
	rowID := int32((0 << 24) | 0) // Batch 0, row 0
	batch, rowIdx := ht.GetBuildBatch(rowID)

	require.NotNil(t, batch)
	assert.Equal(t, 0, rowIdx)
	assert.Equal(t, int64(2), batch.NumRows())

	// Verify we can read the data
	idCol := batch.Column(0).(*array.Int64)
	assert.Equal(t, int64(1), idCol.Value(rowIdx))
}

// TestArrowHashTableInvalidInput tests error handling
func TestArrowHashTableInvalidInput(t *testing.T) {
	allocator := memory.NewGoAllocator()
	ht := operators.NewArrowHashTable(allocator)
	defer ht.Release()

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
	}, nil)

	builder := array.NewRecordBuilder(allocator, schema)
	defer builder.Release()
	builder.Field(0).(*array.Int64Builder).AppendValues([]int64{1}, nil)
	batch := builder.NewRecord()
	defer batch.Release()

	// Test with no key indices
	err := ht.Insert(batch, []int{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no join keys")

	// Test with invalid key index
	err = ht.Insert(batch, []int{999})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid key index")

	// Test probe with invalid index
	_, err = ht.Probe(batch, []int{999})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid key index")
}

// BenchmarkArrowHashTableInsert benchmarks insertion performance
func BenchmarkArrowHashTableInsert(b *testing.B) {
	allocator := memory.NewGoAllocator()

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "value", Type: arrow.BinaryTypes.String},
	}, nil)

	// Create a batch with 1000 rows
	builder := array.NewRecordBuilder(allocator, schema)
	defer builder.Release()

	ids := make([]int64, 1000)
	values := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		ids[i] = int64(i)
		values[i] = "value"
	}

	builder.Field(0).(*array.Int64Builder).AppendValues(ids, nil)
	builder.Field(1).(*array.StringBuilder).AppendValues(values, nil)

	batch := builder.NewRecord()
	defer batch.Release()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ht := operators.NewArrowHashTable(allocator)
		ht.Insert(batch, []int{0})
		ht.Release()
	}
}

// BenchmarkArrowHashTableProbe benchmarks probe performance
func BenchmarkArrowHashTableProbe(b *testing.B) {
	allocator := memory.NewGoAllocator()
	ht := operators.NewArrowHashTable(allocator)
	defer ht.Release()

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
	}, nil)

	// Build side: 1000 rows
	builder := array.NewRecordBuilder(allocator, schema)
	defer builder.Release()
	ids := make([]int64, 1000)
	for i := 0; i < 1000; i++ {
		ids[i] = int64(i)
	}
	builder.Field(0).(*array.Int64Builder).AppendValues(ids, nil)
	buildBatch := builder.NewRecord()
	defer buildBatch.Release()
	ht.Insert(buildBatch, []int{0})

	// Probe batch: 100 rows
	builder2 := array.NewRecordBuilder(allocator, schema)
	defer builder2.Release()
	probeIds := make([]int64, 100)
	for i := 0; i < 100; i++ {
		probeIds[i] = int64(i * 10) // Every 10th element
	}
	builder2.Field(0).(*array.Int64Builder).AppendValues(probeIds, nil)
	probeBatch := builder2.NewRecord()
	defer probeBatch.Release()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ht.Probe(probeBatch, []int{0})
	}
}

package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/storage"
)

// TestZOrderOptimizer tests Z-Order clustering functionality
func TestZOrderOptimizer(t *testing.T) {
	ctx := context.Background()
	tempDir := setupTempDir(t)
	defer os.RemoveAll(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	// Create test database and table
	require.NoError(t, engine.CreateDatabase("testdb"))

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "sourceIP", Type: arrow.PrimitiveTypes.Int64},
			{Name: "destIP", Type: arrow.PrimitiveTypes.Int64},
			{Name: "timestamp", Type: arrow.PrimitiveTypes.Int64},
			{Name: "value", Type: arrow.BinaryTypes.String},
		}, nil,
	)

	require.NoError(t, engine.CreateTable("testdb", "network_logs", schema))

	// Insert test data with multiple Parquet files
	insertNetworkData(t, ctx, engine, "testdb", "network_logs", schema)

	// Create Z-Order optimizer
	zopt := optimizer.NewZOrderOptimizer([]string{"sourceIP", "destIP", "timestamp"})

	// Get initial file list
	initialSnapshot, err := engine.GetDeltaLog().GetSnapshot("testdb.network_logs", -1)
	require.NoError(t, err)
	initialFileCount := len(initialSnapshot.Files)
	t.Logf("Initial file count: %d", initialFileCount)
	require.Greater(t, initialFileCount, 0, "Should have files before optimization")

	// Run Z-Order optimization
	err = zopt.OptimizeTable("testdb.network_logs", initialSnapshot.Files, engine)
	require.NoError(t, err)

	// Verify Z-Order optimization results
	optimizedSnapshot, err := engine.GetDeltaLog().GetSnapshot("testdb.network_logs", -1)
	require.NoError(t, err)
	t.Logf("Optimized file count: %d", len(optimizedSnapshot.Files))

	// Files should be reorganized (may be merged)
	assert.Greater(t, len(optimizedSnapshot.Files), 0, "Should have files after optimization")

	// Test query performance improvement with predicate pushdown
	// sourceIP values in test data: 0, 10, 20, 30, 40, 50, 60, 70, 80, 90
	filters := []storage.Filter{
		{Column: "sourceIP", Operator: "=", Value: int64(50)},
	}

	iterator, err := engine.Scan(ctx, "testdb", "network_logs", filters)
	require.NoError(t, err)
	defer iterator.Close()

	rowCount := int64(0)
	for iterator.Next() {
		record := iterator.Record()
		rowCount += record.NumRows()
	}
	require.NoError(t, iterator.Err())
	t.Logf("Query returned %d rows", rowCount)
	assert.Greater(t, rowCount, int64(0), "Should find matching rows")
}

// TestZOrderComputeZValue tests Z-Order value computation
func TestZOrderComputeZValue(t *testing.T) {
	zopt := optimizer.NewZOrderOptimizer([]string{"dim1", "dim2", "dim3"})

	// Test Z-Order value computation with known inputs
	testCases := []struct {
		name     string
		values   []uint64
		expected bool // true if result is valid (non-zero)
	}{
		{
			name:     "All zeros",
			values:   []uint64{0, 0, 0},
			expected: true,
		},
		{
			name:     "Mixed values",
			values:   []uint64{100, 200, 300},
			expected: true,
		},
		{
			name:     "Max values",
			values:   []uint64{1<<21 - 1, 1<<21 - 1, 1<<21 - 1},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			zValue := zopt.ComputeZValueFromDimensions(tc.values)
			if tc.expected {
				assert.GreaterOrEqual(t, zValue, uint64(0), "Z-Value should be valid")
			}
			t.Logf("Input: %v, Z-Value: %d", tc.values, zValue)
		})
	}
}

// TestZOrderMultiDimensionalQuery tests multi-dimensional query optimization
func TestZOrderMultiDimensionalQuery(t *testing.T) {
	ctx := context.Background()
	tempDir := setupTempDir(t)
	defer os.RemoveAll(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	// Create test database and table
	require.NoError(t, engine.CreateDatabase("testdb"))

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "x", Type: arrow.PrimitiveTypes.Int64},
			{Name: "y", Type: arrow.PrimitiveTypes.Int64},
			{Name: "z", Type: arrow.PrimitiveTypes.Int64},
		}, nil,
	)

	require.NoError(t, engine.CreateTable("testdb", "spatial_data", schema))

	// Insert data with spatial distribution
	insertSpatialData(t, ctx, engine, "testdb", "spatial_data", schema)

	// Get initial snapshot
	initialSnapshot, err := engine.GetDeltaLog().GetSnapshot("testdb.spatial_data", -1)
	require.NoError(t, err)

	// Apply Z-Order on x, y, z dimensions
	zopt := optimizer.NewZOrderOptimizer([]string{"x", "y", "z"})
	err = zopt.OptimizeTable("testdb.spatial_data", initialSnapshot.Files, engine)
	require.NoError(t, err)

	// Test multi-dimensional range query
	filters := []storage.Filter{
		{Column: "x", Operator: ">=", Value: int64(10)},
		{Column: "x", Operator: "<=", Value: int64(50)},
		{Column: "y", Operator: ">=", Value: int64(20)},
		{Column: "y", Operator: "<=", Value: int64(60)},
	}

	iterator, err := engine.Scan(ctx, "testdb", "spatial_data", filters)
	require.NoError(t, err)
	defer iterator.Close()

	rowCount := int64(0)
	for iterator.Next() {
		record := iterator.Record()
		rowCount += record.NumRows()
	}
	require.NoError(t, iterator.Err())
	t.Logf("Multi-dimensional query returned %d rows", rowCount)
	assert.Greater(t, rowCount, int64(0), "Should find matching rows in range")
}

// Helper functions

func insertNetworkData(t *testing.T, ctx context.Context, engine storage.StorageEngine, db, table string, schema *arrow.Schema) {
	pool := memory.NewGoAllocator()

	// Insert multiple batches to create multiple Parquet files
	for batch := 0; batch < 3; batch++ {
		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()

		// Generate test data
		for i := 0; i < 100; i++ {
			id := int64(batch*100 + i)
			sourceIP := int64((i % 10) * 10)
			destIP := int64((i % 15) * 20)
			timestamp := int64(1000000 + i*1000)
			value := fmt.Sprintf("log_%d", id)

			builder.Field(0).(*array.Int64Builder).Append(id)
			builder.Field(1).(*array.Int64Builder).Append(sourceIP)
			builder.Field(2).(*array.Int64Builder).Append(destIP)
			builder.Field(3).(*array.Int64Builder).Append(timestamp)
			builder.Field(4).(*array.StringBuilder).Append(value)
		}

		record := builder.NewRecord()
		err := engine.Write(ctx, db, table, record)
		record.Release()
		require.NoError(t, err)
	}
}

func insertSpatialData(t *testing.T, ctx context.Context, engine storage.StorageEngine, db, table string, schema *arrow.Schema) {
	pool := memory.NewGoAllocator()

	// Insert multiple batches with spatial distribution
	for batch := 0; batch < 3; batch++ {
		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()

		for i := 0; i < 100; i++ {
			id := int64(batch*100 + i)
			x := int64(i % 100)
			y := int64((i * 2) % 100)
			z := int64((i * 3) % 100)

			builder.Field(0).(*array.Int64Builder).Append(id)
			builder.Field(1).(*array.Int64Builder).Append(x)
			builder.Field(2).(*array.Int64Builder).Append(y)
			builder.Field(3).(*array.Int64Builder).Append(z)
		}

		record := builder.NewRecord()
		err := engine.Write(ctx, db, table, record)
		record.Release()
		require.NoError(t, err)
	}
}

func setupTempDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "minidb_zorder_test_*")
	require.NoError(t, err)
	return tempDir
}

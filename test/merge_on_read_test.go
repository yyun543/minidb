package test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/delta"
	"github.com/yyun543/minidb/internal/storage"
)

// TestMergeOnReadUpdate tests Merge-on-Read architecture for UPDATE operations
func TestMergeOnReadUpdate(t *testing.T) {
	ctx := context.Background()
	tempDir := setupTempDirMOR(t)
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
			{Name: "name", Type: arrow.BinaryTypes.String},
			{Name: "value", Type: arrow.PrimitiveTypes.Int64},
		}, nil,
	)

	require.NoError(t, engine.CreateTable("testdb", "test_mor", schema))

	// Insert initial data (creates base files)
	insertTestDataMOR(t, ctx, engine, "testdb", "test_mor", schema, 1000)

	// Get initial snapshot
	initialSnapshot, err := engine.GetDeltaLog().GetSnapshot("testdb.test_mor", -1)
	require.NoError(t, err)
	initialFileCount := len(initialSnapshot.Files)
	initialVersion := engine.GetDeltaLog().GetLatestVersion()

	t.Logf("Initial: files=%d, version=%d", initialFileCount, initialVersion)

	// Perform UPDATE using Merge-on-Read (should create delta file, not rewrite all data)
	filters := []storage.Filter{
		{Column: "id", Operator: "<", Value: int64(10)},
	}
	updates := map[string]interface{}{
		"value": int64(9999),
	}

	startTime := time.Now()
	updatedCount, err := engine.UpdateMergeOnRead(ctx, "testdb", "test_mor", filters, updates)
	updateDuration := time.Since(startTime)
	require.NoError(t, err)
	assert.Greater(t, updatedCount, int64(0), "Should update some rows")

	t.Logf("Updated %d rows in %v using Merge-on-Read", updatedCount, updateDuration)

	// Verify delta file was created (not full rewrite)
	updatedSnapshot, err := engine.GetDeltaLog().GetSnapshot("testdb.test_mor", -1)
	require.NoError(t, err)

	deltaFileCount := 0
	baseFileCount := 0
	for _, file := range updatedSnapshot.Files {
		if file.IsDelta {
			deltaFileCount++
		} else {
			baseFileCount++
		}
	}

	t.Logf("After update: base_files=%d, delta_files=%d", baseFileCount, deltaFileCount)
	assert.Greater(t, deltaFileCount, 0, "Should have created delta files")
	assert.Equal(t, initialFileCount, baseFileCount, "Base files should not be rewritten")

	// Verify data correctness by reading with Merge-on-Read
	iterator, err := engine.Scan(ctx, "testdb", "test_mor", filters)
	require.NoError(t, err)
	defer iterator.Close()

	verifiedCount := int64(0)
	for iterator.Next() {
		record := iterator.Record()
		// Verify updated values
		valueCol := record.Column(2).(*array.Int64)
		for i := 0; i < int(record.NumRows()); i++ {
			assert.Equal(t, int64(9999), valueCol.Value(i), "Value should be updated")
			verifiedCount++
		}
	}
	require.NoError(t, iterator.Err())
	assert.Equal(t, updatedCount, verifiedCount, "Should read correct number of updated rows")
}

// TestMergeOnReadDelete tests Merge-on-Read architecture for DELETE operations
func TestMergeOnReadDelete(t *testing.T) {
	ctx := context.Background()
	tempDir := setupTempDirMOR(t)
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
			{Name: "status", Type: arrow.BinaryTypes.String},
		}, nil,
	)

	require.NoError(t, engine.CreateTable("testdb", "test_delete", schema))

	// Insert initial data
	insertDeleteTestData(t, ctx, engine, "testdb", "test_delete", schema, 500)

	// Get initial count
	initialCount := countRows(t, ctx, engine, "testdb", "test_delete")
	t.Logf("Initial row count: %d", initialCount)

	// Perform DELETE using Merge-on-Read
	filters := []storage.Filter{
		{Column: "id", Operator: "<", Value: int64(50)},
	}

	deletedCount, err := engine.DeleteMergeOnRead(ctx, "testdb", "test_delete", filters)
	require.NoError(t, err)
	assert.Greater(t, deletedCount, int64(0), "Should delete some rows")
	t.Logf("Deleted %d rows using Merge-on-Read", deletedCount)

	// Verify deletion by reading
	finalCount := countRows(t, ctx, engine, "testdb", "test_delete")
	t.Logf("Final row count: %d", finalCount)
	assert.Equal(t, initialCount-deletedCount, finalCount, "Row count should reflect deletions")

	// Verify deleted rows are not returned
	iterator, err := engine.Scan(ctx, "testdb", "test_delete", filters)
	require.NoError(t, err)
	defer iterator.Close()

	deletedRowsFound := int64(0)
	for iterator.Next() {
		record := iterator.Record()
		deletedRowsFound += record.NumRows()
	}
	require.NoError(t, iterator.Err())
	assert.Equal(t, int64(0), deletedRowsFound, "Deleted rows should not be returned")
}

// TestMergeOnReadWriteAmplification 验证 Merge-on-Read 的写放大特性
func TestMergeOnReadWriteAmplification(t *testing.T) {
	ctx := context.Background()

	// Test Merge-on-Read
	morDir := setupTempDirMOR(t)
	defer os.RemoveAll(morDir)

	morEngine, err := storage.NewParquetEngine(morDir)
	require.NoError(t, err)
	require.NoError(t, morEngine.Open())
	defer morEngine.Close()

	require.NoError(t, morEngine.CreateDatabase("testdb"))
	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "data", Type: arrow.BinaryTypes.String},
		}, nil,
	)
	require.NoError(t, morEngine.CreateTable("testdb", "mor_table", schema))
	insertTestDataMOR(t, ctx, morEngine, "testdb", "mor_table", schema, 10000)

	morInitialSnapshot, _ := morEngine.GetDeltaLog().GetSnapshot("testdb.mor_table", -1)
	morInitialSize := calculateTotalSize(morInitialSnapshot.Files)

	// Update 1 row with MoR (现在 Update() 默认使用 MoR)
	morStartTime := time.Now()
	_, err = morEngine.Update(ctx, "testdb", "mor_table", []storage.Filter{{Column: "id", Operator: "=", Value: int64(1)}}, map[string]interface{}{"data": "updated"})
	morDuration := time.Since(morStartTime)
	require.NoError(t, err)

	morFinalSnapshot, _ := morEngine.GetDeltaLog().GetSnapshot("testdb.mor_table", -1)
	morFinalSize := calculateTotalSize(morFinalSnapshot.Files)
	morWriteAmplification := float64(morFinalSize) / float64(morInitialSize)

	t.Logf("Merge-on-Read: duration=%v, write_amplification=%.2fx", morDuration, morWriteAmplification)

	// MoR 写放大应该很小 (只添加一个小的 delta 文件)
	assert.Less(t, morWriteAmplification, 1.1, "MoR write amplification should be minimal (< 1.1x)")
	assert.Less(t, morDuration.Milliseconds(), int64(100), "MoR update should be fast (< 100ms)")
}

// Helper functions

func insertTestDataMOR(t *testing.T, ctx context.Context, engine storage.StorageEngine, db, table string, schema *arrow.Schema, rowCount int) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	for i := 0; i < rowCount; i++ {
		builder.Field(0).(*array.Int64Builder).Append(int64(i))
		if schema.NumFields() > 1 {
			builder.Field(1).(*array.StringBuilder).Append(fmt.Sprintf("data_%d", i))
		}
		if schema.NumFields() > 2 {
			builder.Field(2).(*array.Int64Builder).Append(int64(i * 10))
		}
	}

	record := builder.NewRecord()
	err := engine.Write(ctx, db, table, record)
	record.Release()
	require.NoError(t, err)
}

func insertDeleteTestData(t *testing.T, ctx context.Context, engine storage.StorageEngine, db, table string, schema *arrow.Schema, rowCount int) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	for i := 0; i < rowCount; i++ {
		builder.Field(0).(*array.Int64Builder).Append(int64(i))
		builder.Field(1).(*array.StringBuilder).Append("active")
	}

	record := builder.NewRecord()
	err := engine.Write(ctx, db, table, record)
	record.Release()
	require.NoError(t, err)
}

func countRows(t *testing.T, ctx context.Context, engine storage.StorageEngine, db, table string) int64 {
	iterator, err := engine.Scan(ctx, db, table, nil)
	require.NoError(t, err)
	defer iterator.Close()

	count := int64(0)
	for iterator.Next() {
		record := iterator.Record()
		count += record.NumRows()
	}
	require.NoError(t, iterator.Err())
	return count
}

func calculateTotalSize(files []delta.FileInfo) int64 {
	total := int64(0)
	for _, file := range files {
		total += file.Size
	}
	return total
}

func setupTempDirMOR(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "minidb_mor_test_*")
	require.NoError(t, err)
	return tempDir
}

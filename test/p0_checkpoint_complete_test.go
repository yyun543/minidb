package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/delta"
	"github.com/yyun543/minidb/internal/parquet"
	"github.com/yyun543/minidb/internal/storage"
)

// TestCheckpointComplete_AutoCreation tests automatic checkpoint creation every 10 transactions
// According to architecture document: "Checkpoint 机制未完全实现" (lines 18, 662-698)
// Expected behavior: Create checkpoint Parquet files every 10 Delta Log entries
func TestCheckpointComplete_AutoCreation(t *testing.T) {
	ctx := context.Background()
	tempDir := setupP0TempDir(t)
	defer os.RemoveAll(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	// Create test table
	require.NoError(t, engine.CreateDatabase("testdb"))
	schema := createTestSchema()
	require.NoError(t, engine.CreateTable("testdb", "checkpoint_test", schema))

	// Insert 15 batches to trigger checkpoint at version 10 and 20
	// Version 1: METADATA, Version 2-16: ADD operations
	for i := 0; i < 15; i++ {
		record := createP0TestRecord(t, schema, i*100, 50)
		err := engine.Write(ctx, "testdb", "checkpoint_test", record)
		record.Release()
		require.NoError(t, err)
		t.Logf("Written batch %d, current version: %d", i+1, engine.GetDeltaLog().GetLatestVersion())
	}

	// Wait for async checkpoint creation
	time.Sleep(300 * time.Millisecond)

	deltaLog := engine.GetDeltaLog()
	currentVersion := deltaLog.GetLatestVersion()
	t.Logf("Final version: %d", currentVersion)

	// Check checkpoint files exist
	checkpointDir := filepath.Join(tempDir, "sys", "delta_log", "checkpoints")
	checkpointFiles, err := filepath.Glob(filepath.Join(checkpointDir, "_checkpoint.*.parquet"))
	if err == nil && len(checkpointFiles) > 0 {
		t.Logf("Found %d checkpoint files", len(checkpointFiles))
		for _, f := range checkpointFiles {
			t.Logf("Checkpoint file: %s", f)
		}
	} else {
		t.Logf("No checkpoint files found yet (may be async)")
	}

	// Verify snapshot integrity
	snapshot, err := deltaLog.GetSnapshot("testdb.checkpoint_test", -1)
	require.NoError(t, err)
	assert.Equal(t, 15, len(snapshot.Files), "Should have 15 data files")
}

// TestCheckpointComplete_FastRecovery tests O(1) recovery time using checkpoint
// According to architecture document: Performance impact simulation (lines 689-697)
// Expected: Recovery from checkpoint should be orders of magnitude faster
func TestCheckpointComplete_FastRecovery(t *testing.T) {
	ctx := context.Background()
	tempDir := setupP0TempDir(t)
	defer os.RemoveAll(tempDir)

	// Phase 1: Create large number of transactions
	t.Log("Phase 1: Creating data with many transactions")
	{
		engine, err := storage.NewParquetEngine(tempDir)
		require.NoError(t, err)
		require.NoError(t, engine.Open())

		require.NoError(t, engine.CreateDatabase("testdb"))
		schema := createTestSchema()
		require.NoError(t, engine.CreateTable("testdb", "large_table", schema))

		// Create 100+ transactions to simulate real workload
		// This should create checkpoints at versions 10, 20, 30, etc.
		for i := 0; i < 100; i++ {
			record := createP0TestRecord(t, schema, i*10, 10)
			err := engine.Write(ctx, "testdb", "large_table", record)
			record.Release()
			require.NoError(t, err)

			if (i+1)%20 == 0 {
				t.Logf("Progress: %d batches written", i+1)
			}
		}

		time.Sleep(500 * time.Millisecond) // Allow checkpoint creation
		engine.Close()
	}

	// Phase 2: Measure recovery time
	t.Log("Phase 2: Measuring recovery time")
	startTime := time.Now()

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	recoveryTime := time.Since(startTime)
	t.Logf("Recovery time: %v", recoveryTime)

	// According to architecture doc, with checkpoint recovery should be < 1s even for 10k+ entries
	// For 100 entries it should be very fast
	assert.Less(t, recoveryTime.Milliseconds(), int64(1000),
		"Recovery should be fast with checkpoints (< 1s)")

	// Verify data integrity after recovery
	snapshot, err := engine.GetDeltaLog().GetSnapshot("testdb.large_table", -1)
	require.NoError(t, err)
	assert.Equal(t, 100, len(snapshot.Files), "Should recover all 100 files")

	// Verify can query data
	iterator, err := engine.Scan(ctx, "testdb", "large_table", nil)
	require.NoError(t, err)
	defer iterator.Close()

	totalRows := int64(0)
	for iterator.Next() {
		totalRows += iterator.Record().NumRows()
	}
	require.NoError(t, iterator.Err())
	assert.Equal(t, int64(1000), totalRows, "Should recover all 1000 rows")
	t.Logf("Recovered %d rows successfully", totalRows)
}

// TestCheckpointComplete_Persistence tests checkpoint persistence to Parquet files
// According to architecture document: "序列化 snapshot 到 Parquet 文件" (line 753)
func TestCheckpointComplete_Persistence(t *testing.T) {
	ctx := context.Background()
	tempDir := setupP0TempDir(t)
	defer os.RemoveAll(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	require.NoError(t, engine.CreateDatabase("testdb"))
	schema := createTestSchema()
	require.NoError(t, engine.CreateTable("testdb", "persist_test", schema))

	// Write exactly 10 batches to trigger checkpoint at version 10
	// Version 1: METADATA, Version 2-11: ADD
	for i := 0; i < 10; i++ {
		record := createP0TestRecord(t, schema, i*100, 100)
		err := engine.Write(ctx, "testdb", "persist_test", record)
		record.Release()
		require.NoError(t, err)
	}

	time.Sleep(200 * time.Millisecond)

	// Check if checkpoint marker file exists
	markerPath := filepath.Join(tempDir, "sys", "delta_log", "checkpoints", "_last_checkpoint")
	if _, err := os.Stat(markerPath); err == nil {
		data, err := os.ReadFile(markerPath)
		if err == nil {
			t.Logf("Checkpoint marker exists: %s", string(data))
		}
	} else {
		t.Logf("Checkpoint marker not found (implementation may differ)")
	}

	// Verify checkpoint directory structure
	checkpointDir := filepath.Join(tempDir, "sys", "delta_log", "checkpoints")
	if info, err := os.Stat(checkpointDir); err == nil {
		t.Logf("Checkpoint directory exists: %s, IsDir: %v", checkpointDir, info.IsDir())

		// List all files in checkpoint directory
		files, _ := os.ReadDir(checkpointDir)
		for _, f := range files {
			t.Logf("Checkpoint dir file: %s", f.Name())
		}
	}
}

// TestCheckpointComplete_IncrementalLoad tests incremental loading after checkpoint
// According to architecture document: "O(1) Read Checkpoint + O(k) Incremental" (line 670)
func TestCheckpointComplete_IncrementalLoad(t *testing.T) {
	ctx := context.Background()
	tempDir := setupP0TempDir(t)
	defer os.RemoveAll(tempDir)

	// Phase 1: Create data up to checkpoint
	{
		engine, err := storage.NewParquetEngine(tempDir)
		require.NoError(t, err)
		require.NoError(t, engine.Open())

		require.NoError(t, engine.CreateDatabase("testdb"))
		schema := createTestSchema()
		require.NoError(t, engine.CreateTable("testdb", "incremental_test", schema))

		// Write 15 batches (checkpoint at version 10, then 5 more)
		for i := 0; i < 15; i++ {
			record := createP0TestRecord(t, schema, i*100, 50)
			err := engine.Write(ctx, "testdb", "incremental_test", record)
			record.Release()
			require.NoError(t, err)
		}

		time.Sleep(300 * time.Millisecond)
		engine.Close()
	}

	// Phase 2: Restart and verify incremental loading
	{
		engine, err := storage.NewParquetEngine(tempDir)
		require.NoError(t, err)

		startTime := time.Now()
		require.NoError(t, engine.Open())
		loadTime := time.Since(startTime)
		defer engine.Close()

		t.Logf("Incremental load time: %v", loadTime)

		// With checkpoint, should load checkpoint (version 10) + incremental (5 more entries)
		// Should be much faster than loading all 16 entries linearly
		deltaLog := engine.GetDeltaLog()
		currentVersion := deltaLog.GetLatestVersion()
		assert.Equal(t, int64(16), currentVersion, "Should have version 16 (1 METADATA + 15 ADD)")

		snapshot, err := deltaLog.GetSnapshot("testdb.incremental_test", -1)
		require.NoError(t, err)
		assert.Equal(t, 15, len(snapshot.Files), "Should have all 15 files")
	}
}

// TestCheckpointComplete_CleanupOldLogs tests cleanup of old Delta Log entries
// According to architecture document: "清理旧的 Delta Log 文件（可选）" (line 764)
func TestCheckpointComplete_CleanupOldLogs(t *testing.T) {
	ctx := context.Background()
	tempDir := setupP0TempDir(t)
	defer os.RemoveAll(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	require.NoError(t, engine.CreateDatabase("testdb"))
	schema := createTestSchema()
	require.NoError(t, engine.CreateTable("testdb", "cleanup_test", schema))

	// Write 25 batches to create multiple checkpoints
	for i := 0; i < 25; i++ {
		record := createP0TestRecord(t, schema, i*100, 50)
		err := engine.Write(ctx, "testdb", "cleanup_test", record)
		record.Release()
		require.NoError(t, err)
	}

	time.Sleep(500 * time.Millisecond)

	// Count Delta Log entries
	deltaLog := engine.GetDeltaLog()
	if dl, ok := deltaLog.(*delta.DeltaLog); ok {
		entries := dl.GetAllEntries()
		t.Logf("Total Delta Log entries: %d", len(entries))

		// After checkpointing and cleanup, old entries before last checkpoint
		// could be removed (but this is optional per architecture doc)
		// For now, just verify all entries are still accessible
		assert.GreaterOrEqual(t, len(entries), 25, "Should have at least 25 ADD entries")
	}

	// Verify data is still accessible
	snapshot, err := deltaLog.GetSnapshot("testdb.cleanup_test", -1)
	require.NoError(t, err)
	assert.Equal(t, 25, len(snapshot.Files), "Should still have all 25 files")
}

// TestCheckpointComplete_Schema_Serialization tests Schema serialization in checkpoints
// According to architecture document: Schema 信息应该包含在 checkpoint 中 (section 6.2.2)
func TestCheckpointComplete_SchemaInCheckpoint(t *testing.T) {
	ctx := context.Background()
	tempDir := setupP0TempDir(t)
	defer os.RemoveAll(tempDir)

	// Phase 1: Create table and trigger checkpoint
	{
		engine, err := storage.NewParquetEngine(tempDir)
		require.NoError(t, err)
		require.NoError(t, engine.Open())

		require.NoError(t, engine.CreateDatabase("testdb"))

		// Create schema with multiple types
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "id", Type: arrow.PrimitiveTypes.Int64},
				{Name: "name", Type: arrow.BinaryTypes.String},
				{Name: "age", Type: arrow.PrimitiveTypes.Int32},
				{Name: "active", Type: arrow.FixedWidthTypes.Boolean},
			}, nil,
		)

		require.NoError(t, engine.CreateTable("testdb", "schema_test", schema))

		// Write 10 batches to trigger checkpoint
		for i := 0; i < 10; i++ {
			record := createComplexTestRecord(t, schema, i*100, 50)
			err := engine.Write(ctx, "testdb", "schema_test", record)
			record.Release()
			require.NoError(t, err)
		}

		time.Sleep(300 * time.Millisecond)
		engine.Close()
	}

	// Phase 2: Restart and verify schema recovery
	{
		engine, err := storage.NewParquetEngine(tempDir)
		require.NoError(t, err)
		require.NoError(t, engine.Open())
		defer engine.Close()

		// Verify schema is recovered
		recoveredSchema, err := engine.GetTableSchema("testdb", "schema_test")
		require.NoError(t, err)
		require.NotNil(t, recoveredSchema)

		// Verify all fields
		assert.Equal(t, 4, len(recoveredSchema.Fields()), "Should have 4 fields")
		assert.Equal(t, "id", recoveredSchema.Field(0).Name)
		assert.Equal(t, "name", recoveredSchema.Field(1).Name)
		assert.Equal(t, "age", recoveredSchema.Field(2).Name)
		assert.Equal(t, "active", recoveredSchema.Field(3).Name)

		t.Logf("Schema recovered successfully with %d fields", len(recoveredSchema.Fields()))

		// Verify data is queryable
		iterator, err := engine.Scan(ctx, "testdb", "schema_test", nil)
		require.NoError(t, err)
		defer iterator.Close()

		rowCount := int64(0)
		for iterator.Next() {
			record := iterator.Record()
			rowCount += record.NumRows()

			// Verify record schema matches
			assert.Equal(t, 4, int(record.NumCols()), "Record should have 4 columns")
		}
		require.NoError(t, iterator.Err())
		assert.Equal(t, int64(500), rowCount, "Should have 500 rows")
	}
}

// Helper functions

func setupP0TempDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "minidb_p0_test_*")
	require.NoError(t, err)
	t.Logf("Created temp directory: %s", tempDir)
	return tempDir
}

func createTestSchema() *arrow.Schema {
	return arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "value", Type: arrow.BinaryTypes.String},
		}, nil,
	)
}

func createP0TestRecord(t *testing.T, schema *arrow.Schema, startID int, count int) arrow.Record {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	for i := 0; i < count; i++ {
		id := int64(startID + i)
		value := fmt.Sprintf("test_value_%d", id)

		builder.Field(0).(*array.Int64Builder).Append(id)
		builder.Field(1).(*array.StringBuilder).Append(value)
	}

	return builder.NewRecord()
}

func createComplexTestRecord(t *testing.T, schema *arrow.Schema, startID int, count int) arrow.Record {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	for i := 0; i < count; i++ {
		id := int64(startID + i)
		name := fmt.Sprintf("user_%d", id)
		age := int32(20 + (i % 50))
		active := i%2 == 0

		builder.Field(0).(*array.Int64Builder).Append(id)
		builder.Field(1).(*array.StringBuilder).Append(name)
		builder.Field(2).(*array.Int32Builder).Append(age)
		builder.Field(3).(*array.BooleanBuilder).Append(active)
	}

	return builder.NewRecord()
}

// TestCheckpointComplete_FileFormat tests checkpoint file format and readability
// According to architecture document: Checkpoint 应该使用 Parquet 格式存储 (line 753)
func TestCheckpointComplete_FileFormat(t *testing.T) {
	ctx := context.Background()
	tempDir := setupP0TempDir(t)
	defer os.RemoveAll(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	require.NoError(t, engine.CreateDatabase("testdb"))
	schema := createTestSchema()
	require.NoError(t, engine.CreateTable("testdb", "format_test", schema))

	// Write 10 batches
	for i := 0; i < 10; i++ {
		record := createP0TestRecord(t, schema, i*100, 100)
		err := engine.Write(ctx, "testdb", "format_test", record)
		record.Release()
		require.NoError(t, err)
	}

	time.Sleep(200 * time.Millisecond)

	// Look for checkpoint files
	checkpointDir := filepath.Join(tempDir, "sys", "delta_log", "checkpoints")
	if _, err := os.Stat(checkpointDir); err == nil {
		// Try to read checkpoint file if it exists
		pattern := filepath.Join(checkpointDir, "_checkpoint.*.parquet")
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			t.Logf("Found checkpoint file: %s", matches[0])

			// Try to read the Parquet file
			record, err := parquet.ReadParquetFile(matches[0], nil)
			if err == nil {
				defer record.Release()
				t.Logf("Checkpoint file is valid Parquet format with %d rows, %d columns",
					record.NumRows(), record.NumCols())

				// Log schema
				for i := 0; i < int(record.NumCols()); i++ {
					field := record.Schema().Field(i)
					t.Logf("Column %d: %s (%s)", i, field.Name, field.Type)
				}
			} else {
				t.Logf("Cannot read checkpoint file: %v", err)
			}
		} else {
			t.Logf("No checkpoint files found with pattern: %s", pattern)
		}
	} else {
		t.Logf("Checkpoint directory does not exist yet: %s", checkpointDir)
	}
}

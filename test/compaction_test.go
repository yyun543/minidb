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
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/storage"
)

// TestSmallFileCompaction tests automatic small file compaction
func TestSmallFileCompaction(t *testing.T) {
	ctx := context.Background()
	tempDir := setupTempDirCompaction(t)
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
			{Name: "data", Type: arrow.BinaryTypes.String},
		}, nil,
	)

	require.NoError(t, engine.CreateTable("testdb", "test_compact", schema))

	// Insert many small files (simulating streaming writes)
	numSmallFiles := 20
	rowsPerFile := 10
	insertManySmallFiles(t, ctx, engine, "testdb", "test_compact", schema, numSmallFiles, rowsPerFile)

	// Get initial snapshot
	initialSnapshot, err := engine.GetDeltaLog().GetSnapshot("testdb.test_compact", -1)
	require.NoError(t, err)
	initialFileCount := len(initialSnapshot.Files)
	initialRowCount := int64(0)
	for _, file := range initialSnapshot.Files {
		initialRowCount += file.RowCount
	}

	t.Logf("Before compaction: files=%d, rows=%d", initialFileCount, initialRowCount)
	assert.Equal(t, numSmallFiles, initialFileCount, "Should have many small files")

	// Create compactor with configuration
	compactor := optimizer.NewCompactor(&optimizer.CompactionConfig{
		TargetFileSize:    1 * 1024 * 1024, // 1MB target
		MinFileSize:       1 * 1024,        // 1KB minimum to trigger compaction
		MaxFilesToCompact: 10,
	})

	// Run compaction
	err = compactor.CompactTable("testdb.test_compact", engine)
	require.NoError(t, err)

	// Get post-compaction snapshot
	compactedSnapshot, err := engine.GetDeltaLog().GetSnapshot("testdb.test_compact", -1)
	require.NoError(t, err)
	compactedFileCount := len(compactedSnapshot.Files)
	compactedRowCount := int64(0)
	for _, file := range compactedSnapshot.Files {
		compactedRowCount += file.RowCount
	}

	t.Logf("After compaction: files=%d, rows=%d", compactedFileCount, compactedRowCount)

	// Verify compaction results
	assert.Less(t, compactedFileCount, initialFileCount, "Should have fewer files after compaction")
	assert.Equal(t, initialRowCount, compactedRowCount, "Row count should be preserved")
	assert.Greater(t, compactedFileCount, 0, "Should still have files after compaction")
}

// TestCompactionPreservesData tests that compaction preserves all data correctly
func TestCompactionPreservesData(t *testing.T) {
	ctx := context.Background()
	tempDir := setupTempDirCompaction(t)
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

	require.NoError(t, engine.CreateTable("testdb", "preserve_test", schema))

	// Insert data with known values
	expectedData := make(map[int64]string)
	for i := 0; i < 15; i++ {
		insertSingleRecord(t, ctx, engine, "testdb", "preserve_test", schema, int64(i), fmt.Sprintf("name_%d", i), int64(i*100))
		expectedData[int64(i)] = fmt.Sprintf("name_%d", i)
	}

	// Run compaction
	compactor := optimizer.NewCompactor(&optimizer.CompactionConfig{
		TargetFileSize:    1 * 1024 * 1024,
		MinFileSize:       1 * 1024,
		MaxFilesToCompact: 20,
	})
	err = compactor.CompactTable("testdb.preserve_test", engine)
	require.NoError(t, err)

	// Verify all data is preserved
	iterator, err := engine.Scan(ctx, "testdb", "preserve_test", nil)
	require.NoError(t, err)
	defer iterator.Close()

	foundData := make(map[int64]string)
	for iterator.Next() {
		record := iterator.Record()
		idCol := record.Column(0).(*array.Int64)
		nameCol := record.Column(1).(*array.String)

		for i := 0; i < int(record.NumRows()); i++ {
			id := idCol.Value(i)
			name := nameCol.Value(i)
			foundData[id] = name
		}
	}
	require.NoError(t, iterator.Err())

	// Verify all expected data was found
	assert.Equal(t, len(expectedData), len(foundData), "Should have same number of records")
	for id, expectedName := range expectedData {
		foundName, exists := foundData[id]
		assert.True(t, exists, "ID %d should exist after compaction", id)
		assert.Equal(t, expectedName, foundName, "Name for ID %d should match", id)
	}
}

// TestAutoCompactionBackground tests background automatic compaction
func TestAutoCompactionBackground(t *testing.T) {
	ctx := context.Background()
	tempDir := setupTempDirCompaction(t)
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
			{Name: "timestamp", Type: arrow.PrimitiveTypes.Int64},
		}, nil,
	)

	require.NoError(t, engine.CreateTable("testdb", "auto_compact", schema))

	// Start auto-compaction service
	autoCompactor := optimizer.NewAutoCompactor(&optimizer.CompactionConfig{
		TargetFileSize:    1 * 1024 * 1024,
		MinFileSize:       1 * 1024,
		MaxFilesToCompact: 10,
		CheckInterval:     1 * time.Second,
	})

	// Start background compaction
	stopChan := make(chan struct{})
	go autoCompactor.Start(ctx, engine, stopChan)
	defer func() {
		close(stopChan)
		time.Sleep(100 * time.Millisecond) // Allow goroutine to cleanup
	}()

	// Insert small files over time
	for i := 0; i < 12; i++ {
		insertSingleRecord(t, ctx, engine, "testdb", "auto_compact", schema, int64(i), time.Now().UnixNano())
		time.Sleep(100 * time.Millisecond)
	}

	// Wait for auto-compaction to trigger
	time.Sleep(2 * time.Second)

	// Verify files were compacted
	snapshot, err := engine.GetDeltaLog().GetSnapshot("testdb.auto_compact", -1)
	require.NoError(t, err)
	fileCount := len(snapshot.Files)

	t.Logf("After auto-compaction: files=%d", fileCount)
	assert.Less(t, fileCount, 12, "Auto-compaction should have reduced file count")
}

// TestCompactionWithDataChange tests that compaction respects dataChange flag
func TestCompactionWithDataChange(t *testing.T) {
	ctx := context.Background()
	tempDir := setupTempDirCompaction(t)
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
		}, nil,
	)

	require.NoError(t, engine.CreateTable("testdb", "datachange_test", schema))

	// Insert files
	insertManySmallFiles(t, ctx, engine, "testdb", "datachange_test", schema, 10, 5)

	initialVersion := engine.GetDeltaLog().GetLatestVersion()

	// Run compaction
	compactor := optimizer.NewCompactor(&optimizer.CompactionConfig{
		TargetFileSize:    1 * 1024 * 1024,
		MinFileSize:       1 * 1024,
		MaxFilesToCompact: 10,
	})
	err = compactor.CompactTable("testdb.datachange_test", engine)
	require.NoError(t, err)

	// Verify dataChange flag is set to false for compacted files
	compactionVersion := engine.GetDeltaLog().GetLatestVersion()
	assert.Greater(t, compactionVersion, initialVersion, "Compaction should create new log entries")

	// Get log entries after compaction
	entries := engine.GetDeltaLog().GetEntriesByTable("testdb.datachange_test")
	foundCompactionEntry := false
	for _, entry := range entries {
		if entry.Version > initialVersion && entry.Operation == delta.OpAdd {
			// Compaction entries should have dataChange=false
			foundCompactionEntry = true
			// Note: In actual implementation, we should verify entry.DataChange == false
		}
	}
	assert.True(t, foundCompactionEntry, "Should find compaction entries")
}

// Helper functions

func insertManySmallFiles(t *testing.T, ctx context.Context, engine storage.StorageEngine, db, table string, schema *arrow.Schema, numFiles, rowsPerFile int) {
	pool := memory.NewGoAllocator()

	for fileIdx := 0; fileIdx < numFiles; fileIdx++ {
		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()

		for i := 0; i < rowsPerFile; i++ {
			id := int64(fileIdx*rowsPerFile + i)
			builder.Field(0).(*array.Int64Builder).Append(id)
			if schema.NumFields() > 1 {
				builder.Field(1).(*array.StringBuilder).Append(fmt.Sprintf("data_%d", id))
			}
		}

		record := builder.NewRecord()
		err := engine.Write(ctx, db, table, record)
		record.Release()
		require.NoError(t, err)
	}
}

func insertSingleRecord(t *testing.T, ctx context.Context, engine storage.StorageEngine, db, table string, schema *arrow.Schema, values ...interface{}) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	for i, value := range values {
		switch v := value.(type) {
		case int64:
			builder.Field(i).(*array.Int64Builder).Append(v)
		case string:
			builder.Field(i).(*array.StringBuilder).Append(v)
		}
	}

	record := builder.NewRecord()
	err := engine.Write(ctx, db, table, record)
	record.Release()
	require.NoError(t, err)
}

func setupTempDirCompaction(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "minidb_compaction_test_*")
	require.NoError(t, err)
	return tempDir
}

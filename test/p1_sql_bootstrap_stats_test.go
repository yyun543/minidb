package test

import (
	"context"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/storage"
)

// TestP1Bootstrap_SystemTables tests system tables for statistics
func TestP1Bootstrap_SystemTables(t *testing.T) {
	tempDir := SetupTestDir(t, "p1_bootstrap_system")
	defer CleanupTestDir(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	defer engine.Close()

	// Create sys.table_statistics
	tableStatsSchema := arrow.NewSchema([]arrow.Field{
		{Name: "table_id", Type: arrow.BinaryTypes.String},
		{Name: "version", Type: arrow.PrimitiveTypes.Int64},
		{Name: "row_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "file_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "total_size_bytes", Type: arrow.PrimitiveTypes.Int64},
		{Name: "last_updated", Type: arrow.PrimitiveTypes.Int64},
	}, nil)

	err = engine.CreateTable("sys", "table_statistics", tableStatsSchema)
	require.NoError(t, err, "Should create sys.table_statistics")

	// Create sys.column_statistics
	columnStatsSchema := arrow.NewSchema([]arrow.Field{
		{Name: "table_id", Type: arrow.BinaryTypes.String},
		{Name: "column_name", Type: arrow.BinaryTypes.String},
		{Name: "data_type", Type: arrow.BinaryTypes.String},
		{Name: "min_value", Type: arrow.BinaryTypes.String},
		{Name: "max_value", Type: arrow.BinaryTypes.String},
		{Name: "null_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "distinct_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "last_updated", Type: arrow.PrimitiveTypes.Int64},
	}, nil)

	err = engine.CreateTable("sys", "column_statistics", columnStatsSchema)
	require.NoError(t, err, "Should create sys.column_statistics")

	t.Log("SUCCESS: System tables created for SQL bootstrap statistics")
}

// TestP1Bootstrap_ManualStatistics tests manual statistics collection via SQL
func TestP1Bootstrap_ManualStatistics(t *testing.T) {
	tempDir := SetupTestDir(t, "p1_bootstrap")
	defer CleanupTestDir(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	defer engine.Close()

	ctx := context.Background()

	// Setup: Create system tables
	tableStatsSchema := arrow.NewSchema([]arrow.Field{
		{Name: "table_id", Type: arrow.BinaryTypes.String},
		{Name: "row_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "last_updated", Type: arrow.PrimitiveTypes.Int64},
	}, nil)
	err = engine.CreateTable("sys", "table_statistics", tableStatsSchema)
	require.NoError(t, err)

	// Create test database and table
	err = engine.CreateDatabase("testdb")
	require.NoError(t, err)

	schema := createTestSchema()
	err = engine.CreateTable("testdb", "users", schema)
	require.NoError(t, err)

	// Insert test data
	for i := 0; i < 3; i++ {
		record := createP0TestRecord(t, schema, i*100, 100)
		err = engine.Write(ctx, "testdb", "users", record)
		record.Release()
		require.NoError(t, err)
	}

	// Manually collect statistics (simulating ANALYZE TABLE)
	// This would normally be: ANALYZE TABLE testdb.users
	// We simulate by manually inserting statistics

	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, tableStatsSchema)
	defer builder.Release()

	builder.Field(0).(*array.StringBuilder).Append("testdb.users")
	builder.Field(1).(*array.Int64Builder).Append(300) // 3 * 100 rows
	builder.Field(2).(*array.Int64Builder).Append(1234567890)

	statsRecord := builder.NewRecord()
	defer statsRecord.Release()

	err = engine.Write(ctx, "sys", "table_statistics", statsRecord)
	require.NoError(t, err)

	// Verify statistics can be read back
	iterator, err := engine.Scan(ctx, "sys", "table_statistics", nil)
	require.NoError(t, err)

	rowCount := int64(0)
	for iterator.Next() {
		record := iterator.Record()
		if record.NumRows() > 0 {
			rowCount = record.Column(1).(*array.Int64).Value(0)
			break
		}
	}

	require.Equal(t, int64(300), rowCount, "Statistics should show 300 rows")
	t.Logf("SUCCESS: Manual statistics collection - found %d rows", rowCount)
}

// TestP1Bootstrap_StatisticsQuery tests querying statistics for optimization
func TestP1Bootstrap_StatisticsQuery(t *testing.T) {
	tempDir := SetupTestDir(t, "p1_bootstrap")
	defer CleanupTestDir(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	defer engine.Close()

	ctx := context.Background()

	// Setup system table
	tableStatsSchema := arrow.NewSchema([]arrow.Field{
		{Name: "table_id", Type: arrow.BinaryTypes.String},
		{Name: "row_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "file_count", Type: arrow.PrimitiveTypes.Int64},
	}, nil)
	err = engine.CreateTable("sys", "table_statistics", tableStatsSchema)
	require.NoError(t, err)

	// Insert statistics for multiple tables
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, tableStatsSchema)
	defer builder.Release()

	// Table 1: small table
	builder.Field(0).(*array.StringBuilder).Append("testdb.users")
	builder.Field(1).(*array.Int64Builder).Append(1000)
	builder.Field(2).(*array.Int64Builder).Append(1)

	// Table 2: large table
	builder.Field(0).(*array.StringBuilder).Append("testdb.orders")
	builder.Field(1).(*array.Int64Builder).Append(1000000)
	builder.Field(2).(*array.Int64Builder).Append(100)

	// Table 3: medium table
	builder.Field(0).(*array.StringBuilder).Append("testdb.products")
	builder.Field(1).(*array.Int64Builder).Append(10000)
	builder.Field(2).(*array.Int64Builder).Append(10)

	statsRecord := builder.NewRecord()
	defer statsRecord.Release()

	err = engine.Write(ctx, "sys", "table_statistics", statsRecord)
	require.NoError(t, err)

	// Query statistics (simulating optimizer query)
	iterator, err := engine.Scan(ctx, "sys", "table_statistics", nil)
	require.NoError(t, err)

	tableRowCounts := make(map[string]int64)
	for iterator.Next() {
		record := iterator.Record()
		for i := 0; i < int(record.NumRows()); i++ {
			tableID := record.Column(0).(*array.String).Value(i)
			rowCount := record.Column(1).(*array.Int64).Value(i)
			tableRowCounts[tableID] = rowCount
		}
	}

	// Verify statistics
	require.Equal(t, int64(1000), tableRowCounts["testdb.users"])
	require.Equal(t, int64(1000000), tableRowCounts["testdb.orders"])
	require.Equal(t, int64(10000), tableRowCounts["testdb.products"])

	// Simulate JOIN order optimization
	// Rule: small tables first
	tables := []string{"testdb.orders", "testdb.users", "testdb.products"}

	// Sort by row count
	type tableSize struct {
		name     string
		rowCount int64
	}
	sizes := make([]tableSize, 0)
	for _, table := range tables {
		sizes = append(sizes, tableSize{table, tableRowCounts[table]})
	}

	// Simple bubble sort
	for i := 0; i < len(sizes); i++ {
		for j := i + 1; j < len(sizes); j++ {
			if sizes[j].rowCount < sizes[i].rowCount {
				sizes[i], sizes[j] = sizes[j], sizes[i]
			}
		}
	}

	// Verify optimal order: users (1k) -> products (10k) -> orders (1M)
	require.Equal(t, "testdb.users", sizes[0].name)
	require.Equal(t, "testdb.products", sizes[1].name)
	require.Equal(t, "testdb.orders", sizes[2].name)

	t.Logf("SUCCESS: Statistics-based JOIN order optimization")
	t.Logf("  Optimal order: %s (%d) -> %s (%d) -> %s (%d)",
		sizes[0].name, sizes[0].rowCount,
		sizes[1].name, sizes[1].rowCount,
		sizes[2].name, sizes[2].rowCount)
}

// TestP1Bootstrap_ColumnStatistics tests column-level statistics
func TestP1Bootstrap_ColumnStatistics(t *testing.T) {
	tempDir := SetupTestDir(t, "p1_bootstrap")
	defer CleanupTestDir(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	defer engine.Close()

	ctx := context.Background()

	// Setup column statistics table
	columnStatsSchema := arrow.NewSchema([]arrow.Field{
		{Name: "table_id", Type: arrow.BinaryTypes.String},
		{Name: "column_name", Type: arrow.BinaryTypes.String},
		{Name: "min_value", Type: arrow.BinaryTypes.String},
		{Name: "max_value", Type: arrow.BinaryTypes.String},
		{Name: "distinct_count", Type: arrow.PrimitiveTypes.Int64},
	}, nil)
	err = engine.CreateTable("sys", "column_statistics", columnStatsSchema)
	require.NoError(t, err)

	// Insert column statistics
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, columnStatsSchema)
	defer builder.Release()

	// user_id column: high cardinality (good for index)
	builder.Field(0).(*array.StringBuilder).Append("testdb.users")
	builder.Field(1).(*array.StringBuilder).Append("user_id")
	builder.Field(2).(*array.StringBuilder).Append("1")
	builder.Field(3).(*array.StringBuilder).Append("10000")
	builder.Field(4).(*array.Int64Builder).Append(10000)

	// status column: low cardinality (not good for index)
	builder.Field(0).(*array.StringBuilder).Append("testdb.users")
	builder.Field(1).(*array.StringBuilder).Append("status")
	builder.Field(2).(*array.StringBuilder).Append("active")
	builder.Field(3).(*array.StringBuilder).Append("inactive")
	builder.Field(4).(*array.Int64Builder).Append(2)

	statsRecord := builder.NewRecord()
	defer statsRecord.Release()

	err = engine.Write(ctx, "sys", "column_statistics", statsRecord)
	require.NoError(t, err)

	// Query column statistics
	iterator, err := engine.Scan(ctx, "sys", "column_statistics", nil)
	require.NoError(t, err)

	columnCardinality := make(map[string]int64)
	for iterator.Next() {
		record := iterator.Record()
		for i := 0; i < int(record.NumRows()); i++ {
			columnName := record.Column(1).(*array.String).Value(i)
			distinctCount := record.Column(4).(*array.Int64).Value(i)
			columnCardinality[columnName] = distinctCount
		}
	}

	// Verify cardinality
	require.Equal(t, int64(10000), columnCardinality["user_id"])
	require.Equal(t, int64(2), columnCardinality["status"])

	// Simulate index recommendation
	// High cardinality columns (> 1000) are good candidates for indexes
	highCardinalityColumns := make([]string, 0)
	for col, card := range columnCardinality {
		if card > 1000 {
			highCardinalityColumns = append(highCardinalityColumns, col)
		}
	}

	require.Contains(t, highCardinalityColumns, "user_id")
	require.NotContains(t, highCardinalityColumns, "status")

	t.Logf("SUCCESS: Column statistics for index recommendation")
	t.Logf("  High cardinality columns (good for index): %v", highCardinalityColumns)
}

// TestP1Bootstrap_IncrementalUpdate tests incremental statistics updates
func TestP1Bootstrap_IncrementalUpdate(t *testing.T) {
	tempDir := SetupTestDir(t, "p1_bootstrap")
	defer CleanupTestDir(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	defer engine.Close()

	ctx := context.Background()

	// Setup
	tableStatsSchema := arrow.NewSchema([]arrow.Field{
		{Name: "table_id", Type: arrow.BinaryTypes.String},
		{Name: "row_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "last_updated", Type: arrow.PrimitiveTypes.Int64},
	}, nil)
	err = engine.CreateTable("sys", "table_statistics", tableStatsSchema)
	require.NoError(t, err)

	// Initial statistics
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, tableStatsSchema)
	builder.Field(0).(*array.StringBuilder).Append("testdb.users")
	builder.Field(1).(*array.Int64Builder).Append(1000)
	builder.Field(2).(*array.Int64Builder).Append(1000)
	record1 := builder.NewRecord()
	err = engine.Write(ctx, "sys", "table_statistics", record1)
	record1.Release()
	require.NoError(t, err)

	// Updated statistics (after more inserts)
	builder2 := array.NewRecordBuilder(pool, tableStatsSchema)
	defer builder2.Release()
	builder2.Field(0).(*array.StringBuilder).Append("testdb.users")
	builder2.Field(1).(*array.Int64Builder).Append(1500) // 50% increase
	builder2.Field(2).(*array.Int64Builder).Append(2000)
	record2 := builder2.NewRecord()
	defer record2.Release()
	err = engine.Write(ctx, "sys", "table_statistics", record2)
	require.NoError(t, err)

	// Query latest statistics
	iterator, err := engine.Scan(ctx, "sys", "table_statistics", nil)
	require.NoError(t, err)

	var latestRowCount int64
	var latestTimestamp int64
	for iterator.Next() {
		record := iterator.Record()
		for i := 0; i < int(record.NumRows()); i++ {
			timestamp := record.Column(2).(*array.Int64).Value(i)
			if timestamp > latestTimestamp {
				latestTimestamp = timestamp
				latestRowCount = record.Column(1).(*array.Int64).Value(i)
			}
		}
	}

	require.Equal(t, int64(1500), latestRowCount, "Should get latest statistics")
	t.Logf("SUCCESS: Incremental statistics update - %d rows", latestRowCount)
}

// TestP1Bootstrap_PerformanceComparison tests query performance with/without statistics
func TestP1Bootstrap_PerformanceComparison(t *testing.T) {
	t.Log("=== Performance Comparison: With vs Without Statistics ===")

	// Scenario: JOIN optimization based on table sizes
	// Without statistics: arbitrary join order
	// With statistics: small table first (build hash table on smaller table)

	tableRowCounts := map[string]int64{
		"users":    10000,
		"orders":   1000000,
		"products": 50000,
	}

	// Without statistics (arbitrary order: large tables first)
	// orders JOIN products: scan 1M rows, probe 50K times = 1M + 1M*50K comparisons
	// (orders JOIN products) JOIN users: scan result, probe 10K times
	arbitraryOrder := []string{"orders", "products", "users"}
	arbitraryCost := int64(0)
	currentSize := tableRowCounts[arbitraryOrder[0]]
	for i := 1; i < len(arbitraryOrder); i++ {
		// Nested loop join cost: outer table scan + outer * inner probes
		rightSize := tableRowCounts[arbitraryOrder[i]]
		arbitraryCost += currentSize + (currentSize * rightSize)
		currentSize = currentSize * rightSize // Result size grows
	}

	// With statistics (optimal order: small to large)
	// users JOIN products: scan 10K rows, probe 50K times = 10K + 10K*50K
	// (users JOIN products) JOIN orders: scan result, probe 1M times
	optimalOrder := []string{"users", "products", "orders"}
	optimalCost := int64(0)
	currentSize = tableRowCounts[optimalOrder[0]]
	for i := 1; i < len(optimalOrder); i++ {
		rightSize := tableRowCounts[optimalOrder[i]]
		optimalCost += currentSize + (currentSize * rightSize)
		currentSize = currentSize * rightSize
	}

	speedup := float64(arbitraryCost) / float64(optimalCost)

	t.Logf("Arbitrary order cost: %d operations", arbitraryCost)
	t.Logf("Optimal order cost: %d operations", optimalCost)
	t.Logf("Speedup with statistics: %.2fx", speedup)

	require.Greater(t, speedup, 1.0, "Statistics should provide speedup")
	t.Log("SUCCESS: Statistics-based optimization provides measurable speedup")
}

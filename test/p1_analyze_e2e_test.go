package test

import (
	"context"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

// TestAnalyzeTableE2E 端到端测试ANALYZE TABLE命令
func TestAnalyzeTableE2E(t *testing.T) {
	tempDir := SetupTestDir(t, "analyze_e2e")
	defer CleanupTestDir(tempDir)

	// 创建引擎和catalog
	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	defer engine.Close()
	err = engine.Open()
	require.NoError(t, err)

	cat := catalog.NewCatalog()
	cat.SetStorageEngine(engine)
	err = cat.Init()
	require.NoError(t, err)

	// 创建executor和optimizer
	exec := executor.NewExecutor(cat)
	opt := optimizer.NewOptimizer()

	// 创建session
	sess := &session.Session{
		ID:        1,
		CurrentDB: "",
	}

	// 1. 创建测试数据库和表
	t.Log("=== Step 1: Creating test database and table ===")

	createDBSQL := "CREATE DATABASE testdb"
	ast, err := parser.Parse(createDBSQL)
	require.NoError(t, err, "Failed to parse CREATE DATABASE")

	plan, err := opt.Optimize(ast)
	require.NoError(t, err, "Failed to optimize CREATE DATABASE")

	_, err = exec.Execute(plan, sess)
	require.NoError(t, err, "Failed to execute CREATE DATABASE")

	// 设置当前数据库
	sess.CurrentDB = "testdb"

	// 创建表
	err = engine.CreateDatabase("testdb")
	require.NoError(t, err)

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "name", Type: arrow.BinaryTypes.String},
		{Name: "score", Type: arrow.PrimitiveTypes.Int64},
	}, nil)

	err = engine.CreateTable("testdb", "students", schema)
	require.NoError(t, err)

	err = cat.CreateTable("testdb", catalog.TableMeta{
		Database: "testdb",
		Table:    "students",
		Schema:   schema,
	})
	require.NoError(t, err)

	t.Log("✓ Database and table created")

	// 2. 插入测试数据
	t.Log("=== Step 2: Inserting test data ===")

	pool := memory.NewGoAllocator()

	// 插入100条记录，分3批
	for batch := 0; batch < 3; batch++ {
		builder := array.NewRecordBuilder(pool, schema)

		for i := 0; i < 100; i++ {
			rowNum := batch*100 + i
			builder.Field(0).(*array.Int64Builder).Append(int64(rowNum + 1))
			builder.Field(1).(*array.StringBuilder).Append("Student" + string(rune('A'+rowNum%26)))
			builder.Field(2).(*array.Int64Builder).Append(int64(60 + rowNum%40)) // scores 60-99
		}

		record := builder.NewRecord()
		ctx := context.Background()
		err = engine.Write(ctx, "testdb", "students", record)
		record.Release()
		builder.Release()
		require.NoError(t, err)
	}

	t.Log("✓ Inserted 300 rows")

	// 3. 执行ANALYZE TABLE命令
	t.Log("=== Step 3: Executing ANALYZE TABLE ===")

	analyzeSQL := "ANALYZE TABLE testdb.students"
	ast, err = parser.Parse(analyzeSQL)
	require.NoError(t, err, "Failed to parse ANALYZE TABLE")

	plan, err = opt.Optimize(ast)
	require.NoError(t, err, "Failed to optimize ANALYZE TABLE")
	require.Equal(t, optimizer.AnalyzePlan, plan.Type, "Plan type should be AnalyzePlan")

	result, err := exec.Execute(plan, sess)
	require.NoError(t, err, "Failed to execute ANALYZE TABLE")
	require.NotNil(t, result)

	t.Log("✓ ANALYZE TABLE executed successfully")

	// 4. 验证表级统计信息
	t.Log("=== Step 4: Verifying table statistics ===")

	ctx := context.Background()
	iterator, err := engine.Scan(ctx, "sys", "table_statistics", nil)
	require.NoError(t, err)

	foundTableStats := false
	var rowCount int64

	for iterator.Next() {
		record := iterator.Record()
		for i := 0; i < int(record.NumRows()); i++ {
			tableID := record.Column(0).(*array.String).Value(i)
			if tableID == "testdb.students" {
				foundTableStats = true
				rowCount = record.Column(2).(*array.Int64).Value(i)
				break
			}
		}
	}

	require.True(t, foundTableStats, "Should find table statistics")
	require.Equal(t, int64(300), rowCount, "Should have 300 rows")

	t.Logf("✓ Table statistics verified: %d rows", rowCount)

	// 5. 验证列级统计信息
	t.Log("=== Step 5: Verifying column statistics ===")

	iterator, err = engine.Scan(ctx, "sys", "column_statistics", nil)
	require.NoError(t, err)

	columnStats := make(map[string]struct {
		minValue      string
		maxValue      string
		distinctCount int64
	})

	for iterator.Next() {
		record := iterator.Record()
		for i := 0; i < int(record.NumRows()); i++ {
			tableID := record.Column(0).(*array.String).Value(i)
			if tableID == "testdb.students" {
				colName := record.Column(1).(*array.String).Value(i)
				minValue := record.Column(3).(*array.String).Value(i)
				maxValue := record.Column(4).(*array.String).Value(i)
				distinctCount := record.Column(6).(*array.Int64).Value(i)

				columnStats[colName] = struct {
					minValue      string
					maxValue      string
					distinctCount int64
				}{minValue, maxValue, distinctCount}
			}
		}
	}

	// 验证id列统计
	require.Contains(t, columnStats, "id")
	idStats := columnStats["id"]
	require.Equal(t, "1", idStats.minValue, "ID min should be 1")
	require.Equal(t, "300", idStats.maxValue, "ID max should be 300")
	require.Equal(t, int64(300), idStats.distinctCount, "ID should have 300 distinct values")

	// 验证score列统计
	require.Contains(t, columnStats, "score")
	scoreStats := columnStats["score"]
	require.NotEmpty(t, scoreStats.minValue, "Score should have min value")
	require.NotEmpty(t, scoreStats.maxValue, "Score should have max value")
	require.True(t, scoreStats.distinctCount <= 40, "Score should have at most 40 distinct values")

	t.Logf("✓ Column statistics verified:")
	t.Logf("  - id: min=%s, max=%s, distinct=%d", idStats.minValue, idStats.maxValue, idStats.distinctCount)
	t.Logf("  - score: min=%s, max=%s, distinct=%d", scoreStats.minValue, scoreStats.maxValue, scoreStats.distinctCount)
	t.Logf("  - name: distinct=%d", columnStats["name"].distinctCount)

	// 6. 测试带列列表的ANALYZE
	t.Log("=== Step 6: Testing ANALYZE with column list ===")

	analyzeSQL = "ANALYZE TABLE testdb.students (id, score)"
	ast, err = parser.Parse(analyzeSQL)
	require.NoError(t, err, "Failed to parse ANALYZE TABLE with columns")

	plan, err = opt.Optimize(ast)
	require.NoError(t, err, "Failed to optimize ANALYZE TABLE with columns")

	props, ok := plan.Properties.(*optimizer.AnalyzeProperties)
	require.True(t, ok)
	require.Equal(t, "testdb.students", props.Table)
	require.Equal(t, 2, len(props.Columns))
	require.Contains(t, props.Columns, "id")
	require.Contains(t, props.Columns, "score")

	t.Log("✓ ANALYZE with column list parsed correctly")

	t.Log("=== All ANALYZE TABLE E2E tests PASSED ===")
}

// TestAnalyzeTableIncrementalE2E 测试增量ANALYZE
func TestAnalyzeTableIncrementalE2E(t *testing.T) {
	tempDir := SetupTestDir(t, "analyze_incremental")
	defer CleanupTestDir(tempDir)

	engine, err := storage.NewParquetEngine(tempDir)
	require.NoError(t, err)
	defer engine.Close()
	err = engine.Open()
	require.NoError(t, err)

	cat := catalog.NewCatalog()
	cat.SetStorageEngine(engine)
	err = cat.Init()
	require.NoError(t, err)

	exec := executor.NewExecutor(cat)
	opt := optimizer.NewOptimizer()

	sess := &session.Session{
		ID:        1,
		CurrentDB: "testdb",
	}

	// 创建表
	err = engine.CreateDatabase("testdb")
	require.NoError(t, err)
	err = cat.CreateDatabase("testdb")
	require.NoError(t, err)

	schema := arrow.NewSchema([]arrow.Field{
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "value", Type: arrow.BinaryTypes.String},
	}, nil)

	err = engine.CreateTable("testdb", "data", schema)
	require.NoError(t, err)

	err = cat.CreateTable("testdb", catalog.TableMeta{
		Database: "testdb",
		Table:    "data",
		Schema:   schema,
	})
	require.NoError(t, err)

	// 第一次插入100行
	t.Log("=== First insert: 100 rows ===")
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	for i := 0; i < 100; i++ {
		builder.Field(0).(*array.Int64Builder).Append(int64(i + 1))
		builder.Field(1).(*array.StringBuilder).Append("value" + string(rune('A'+i%10)))
	}
	record := builder.NewRecord()
	ctx := context.Background()
	err = engine.Write(ctx, "testdb", "data", record)
	record.Release()
	builder.Release()
	require.NoError(t, err)

	// 第一次ANALYZE
	analyzeSQL := "ANALYZE TABLE testdb.data"
	ast, err := parser.Parse(analyzeSQL)
	require.NoError(t, err)
	plan, err := opt.Optimize(ast)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.NoError(t, err)

	// 查询第一次统计
	iterator, err := engine.Scan(ctx, "sys", "table_statistics", nil)
	require.NoError(t, err)

	var firstRowCount int64
	for iterator.Next() {
		record := iterator.Record()
		for i := 0; i < int(record.NumRows()); i++ {
			tableID := record.Column(0).(*array.String).Value(i)
			if tableID == "testdb.data" {
				firstRowCount = record.Column(2).(*array.Int64).Value(i)
			}
		}
	}
	require.Equal(t, int64(100), firstRowCount, "First analyze should show 100 rows")
	t.Logf("✓ First ANALYZE: %d rows", firstRowCount)

	// 第二次插入150行
	t.Log("=== Second insert: 150 rows ===")
	builder2 := array.NewRecordBuilder(pool, schema)
	for i := 0; i < 150; i++ {
		builder2.Field(0).(*array.Int64Builder).Append(int64(100 + i + 1))
		builder2.Field(1).(*array.StringBuilder).Append("value" + string(rune('A'+i%10)))
	}
	record2 := builder2.NewRecord()
	err = engine.Write(ctx, "testdb", "data", record2)
	record2.Release()
	builder2.Release()
	require.NoError(t, err)

	// 第二次ANALYZE
	_, err = exec.Execute(plan, sess)
	require.NoError(t, err)

	// 查询第二次统计 - 应该看到多条记录（增量）
	iterator2, err := engine.Scan(ctx, "sys", "table_statistics", nil)
	require.NoError(t, err)

	var latestRowCount int64
	var latestTimestamp int64
	recordCount := 0

	for iterator2.Next() {
		record := iterator2.Record()
		for i := 0; i < int(record.NumRows()); i++ {
			tableID := record.Column(0).(*array.String).Value(i)
			if tableID == "testdb.data" {
				recordCount++
				timestamp := record.Column(5).(*array.Int64).Value(i)
				rowCount := record.Column(2).(*array.Int64).Value(i)
				t.Logf("Found record: timestamp=%d, row_count=%d", timestamp, rowCount)
				if timestamp >= latestTimestamp {
					latestTimestamp = timestamp
					latestRowCount = rowCount
				}
			}
		}
	}

	require.True(t, recordCount >= 2, "Should have at least 2 statistics records (incremental)")
	require.Equal(t, int64(250), latestRowCount, "Latest analyze should show 250 rows")
	t.Logf("✓ Second ANALYZE: %d rows (total %d statistics records)", latestRowCount, recordCount)

	t.Log("=== Incremental ANALYZE test PASSED ===")
}

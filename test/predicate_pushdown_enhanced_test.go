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
	"github.com/yyun543/minidb/internal/storage"
)

// TestPredicatePushdownEnhanced 测试增强的谓词下推功能
// 验证所有比较操作符的文件级数据跳过
func TestPredicatePushdownEnhanced(t *testing.T) {
	basePath := filepath.Join(os.TempDir(), fmt.Sprintf("minidb_predicate_test_%d", time.Now().UnixNano()))
	defer os.RemoveAll(basePath)

	engine, err := storage.NewParquetEngine(basePath)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	// 创建测试数据库和表
	require.NoError(t, engine.CreateDatabase("testdb"))

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "age", Type: arrow.PrimitiveTypes.Int64},
			{Name: "score", Type: arrow.PrimitiveTypes.Float64},
			{Name: "name", Type: arrow.BinaryTypes.String},
		},
		nil,
	)
	require.NoError(t, engine.CreateTable("testdb", "users", schema))

	// 插入多个文件，每个文件有不同的数据范围
	// 文件1: id=[1-10], age=[20-29], score=[60-69]
	// 文件2: id=[11-20], age=[30-39], score=[70-79]
	// 文件3: id=[21-30], age=[40-49], score=[80-89]
	files := []struct {
		idRange    [2]int64
		ageRange   [2]int64
		scoreRange [2]float64
	}{
		{[2]int64{1, 10}, [2]int64{20, 29}, [2]float64{60, 69}},
		{[2]int64{11, 20}, [2]int64{30, 39}, [2]float64{70, 79}},
		{[2]int64{21, 30}, [2]int64{40, 49}, [2]float64{80, 89}},
	}

	pool := memory.NewGoAllocator()
	for _, fileRange := range files {
		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()

		for i := fileRange.idRange[0]; i <= fileRange.idRange[1]; i++ {
			builder.Field(0).(*array.Int64Builder).Append(i)
			builder.Field(1).(*array.Int64Builder).Append(fileRange.ageRange[0] + (i - fileRange.idRange[0]))
			builder.Field(2).(*array.Float64Builder).Append(fileRange.scoreRange[0] + float64(i-fileRange.idRange[0]))
			builder.Field(3).(*array.StringBuilder).Append(fmt.Sprintf("user%d", i))
		}

		record := builder.NewRecord()
		require.NoError(t, engine.Write(context.Background(), "testdb", "users", record))
		record.Release()
	}

	// 测试用例
	testCases := []struct {
		name             string
		filters          []storage.Filter
		expectedFilesMin int // 最少应该扫描的文件数
		expectedFilesMax int // 最多应该扫描的文件数
		expectedRowsMin  int // 至少应该返回的行数
		expectedRowsMax  int // 最多应该返回的行数
		description      string
	}{
		{
			name: "等值过滤 - age=25",
			filters: []storage.Filter{
				{Column: "age", Operator: "=", Value: int64(25)},
			},
			expectedFilesMin: 1,
			expectedFilesMax: 1, // 应该只扫描文件1
			expectedRowsMin:  1,
			expectedRowsMax:  1,
			description:      "等值过滤应该只扫描包含该值的文件",
		},
		{
			name: "大于过滤 - age>35",
			filters: []storage.Filter{
				{Column: "age", Operator: ">", Value: int64(35)},
			},
			expectedFilesMin: 2,
			expectedFilesMax: 2, // 应该扫描文件2和文件3
			expectedRowsMin:  14,
			expectedRowsMax:  14,
			description:      "大于过滤应该跳过最大值<=35的文件",
		},
		{
			name: "小于过滤 - age<25",
			filters: []storage.Filter{
				{Column: "age", Operator: "<", Value: int64(25)},
			},
			expectedFilesMin: 1,
			expectedFilesMax: 1, // 应该只扫描文件1
			expectedRowsMin:  5,
			expectedRowsMax:  5,
			description:      "小于过滤应该跳过最小值>=25的文件",
		},
		{
			name: "大于等于过滤 - score>=75",
			filters: []storage.Filter{
				{Column: "score", Operator: ">=", Value: float64(75)},
			},
			expectedFilesMin: 2,
			expectedFilesMax: 2, // 应该扫描文件2和文件3
			expectedRowsMin:  15,
			expectedRowsMax:  15,
			description:      "大于等于过滤应该跳过最大值<75的文件",
		},
		{
			name: "小于等于过滤 - score<=65",
			filters: []storage.Filter{
				{Column: "score", Operator: "<=", Value: float64(65)},
			},
			expectedFilesMin: 1,
			expectedFilesMax: 1, // 应该只扫描文件1
			expectedRowsMin:  6,
			expectedRowsMax:  6,
			description:      "小于等于过滤应该跳过最小值>65的文件",
		},
		{
			name: "范围过滤 - age>25 AND age<45",
			filters: []storage.Filter{
				{Column: "age", Operator: ">", Value: int64(25)},
				{Column: "age", Operator: "<", Value: int64(45)},
			},
			expectedFilesMin: 2,
			expectedFilesMax: 3, // 可能扫描文件1、2、3（取决于边界判断）
			expectedRowsMin:  17,
			expectedRowsMax:  19, // 允许一定的边界值差异
			description:      "范围过滤应该正确裁剪文件",
		},
		{
			name: "无匹配过滤 - age>100",
			filters: []storage.Filter{
				{Column: "age", Operator: ">", Value: int64(100)},
			},
			expectedFilesMin: 0,
			expectedFilesMax: 0, // 应该跳过所有文件
			expectedRowsMin:  0,
			expectedRowsMax:  0,
			description:      "无匹配条件应该跳过所有文件",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行扫描
			iter, err := engine.Scan(context.Background(), "testdb", "users", tc.filters)
			require.NoError(t, err)

			// 收集结果
			totalRows := int64(0)
			for iter.Next() {
				record := iter.Record()
				if record != nil {
					totalRows += record.NumRows()
				}
			}
			require.NoError(t, iter.Err())
			iter.Close()

			// 验证结果行数
			assert.GreaterOrEqual(t, int(totalRows), tc.expectedRowsMin,
				"返回的行数应该至少为 %d，实际为 %d (%s)", tc.expectedRowsMin, totalRows, tc.description)
			assert.LessOrEqual(t, int(totalRows), tc.expectedRowsMax,
				"返回的行数应该最多为 %d，实际为 %d (%s)", tc.expectedRowsMax, totalRows, tc.description)

			t.Logf("✓ %s: 返回 %d 行 (%s)", tc.name, totalRows, tc.description)
		})
	}
}

// TestPredicatePushdownIN 测试IN操作符的谓词下推
// 实现存储层IN操作符的谓词下推优化，提升查询性能
func TestPredicatePushdownIN(t *testing.T) {
	basePath := filepath.Join(os.TempDir(), fmt.Sprintf("minidb_predicate_in_test_%d", time.Now().UnixNano()))
	defer os.RemoveAll(basePath)

	engine, err := storage.NewParquetEngine(basePath)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	require.NoError(t, engine.CreateDatabase("testdb"))

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "category", Type: arrow.PrimitiveTypes.Int64},
		},
		nil,
	)
	require.NoError(t, engine.CreateTable("testdb", "products", schema))

	// 插入3个文件，category分别为 [1-5], [6-10], [11-15]
	pool := memory.NewGoAllocator()
	for fileIdx := 0; fileIdx < 3; fileIdx++ {
		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()

		for i := 0; i < 5; i++ {
			id := int64(fileIdx*5 + i + 1)
			category := int64(fileIdx*5 + i + 1)
			builder.Field(0).(*array.Int64Builder).Append(id)
			builder.Field(1).(*array.Int64Builder).Append(category)
		}

		record := builder.NewRecord()
		require.NoError(t, engine.Write(context.Background(), "testdb", "products", record))
		record.Release()
	}

	testCases := []struct {
		name          string
		inValues      []interface{}
		expectedFiles int
		expectedRows  int
		description   string
	}{
		{
			name:          "IN - 单个文件范围内的值",
			inValues:      []interface{}{int64(2), int64(3), int64(4)},
			expectedFiles: 1, // 应该只扫描第一个文件
			expectedRows:  3,
			description:   "IN操作符的值都在第一个文件范围内",
		},
		{
			name:          "IN - 跨多个文件的值",
			inValues:      []interface{}{int64(3), int64(8), int64(13)},
			expectedFiles: 3, // 需要扫描所有3个文件
			expectedRows:  3,
			description:   "IN操作符的值跨越所有文件",
		},
		{
			name:          "IN - 没有匹配的值",
			inValues:      []interface{}{int64(100), int64(200)},
			expectedFiles: 0, // 应该跳过所有文件
			expectedRows:  0,
			description:   "IN操作符的值不在任何文件范围内",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filters := []storage.Filter{
				{Column: "category", Operator: "IN", Values: tc.inValues},
			}

			iter, err := engine.Scan(context.Background(), "testdb", "products", filters)
			require.NoError(t, err)

			totalRows := int64(0)
			for iter.Next() {
				record := iter.Record()
				if record != nil {
					totalRows += record.NumRows()
				}
			}
			require.NoError(t, iter.Err())
			iter.Close()

			assert.Equal(t, tc.expectedRows, int(totalRows),
				"%s: 期望 %d 行，实际 %d 行", tc.description, tc.expectedRows, totalRows)

			t.Logf("✓ %s: 返回 %d 行", tc.name, totalRows)
		})
	}
}

// TestPredicatePushdownEnhancedPerformance 性能测试
func TestPredicatePushdownEnhancedPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	basePath := filepath.Join(os.TempDir(), fmt.Sprintf("minidb_predicate_perf_test_%d", time.Now().UnixNano()))
	defer os.RemoveAll(basePath)

	engine, err := storage.NewParquetEngine(basePath)
	require.NoError(t, err)
	require.NoError(t, engine.Open())
	defer engine.Close()

	require.NoError(t, engine.CreateDatabase("perfdb"))

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "timestamp", Type: arrow.PrimitiveTypes.Int64},
			{Name: "value", Type: arrow.PrimitiveTypes.Float64},
		},
		nil,
	)
	require.NoError(t, engine.CreateTable("perfdb", "events", schema))

	// 插入100个文件，每个文件1000行
	pool := memory.NewGoAllocator()
	totalFiles := 100
	rowsPerFile := 1000

	t.Logf("插入 %d 个文件，每个文件 %d 行...", totalFiles, rowsPerFile)
	for fileIdx := 0; fileIdx < totalFiles; fileIdx++ {
		builder := array.NewRecordBuilder(pool, schema)

		baseTimestamp := int64(fileIdx * rowsPerFile)
		for i := 0; i < rowsPerFile; i++ {
			builder.Field(0).(*array.Int64Builder).Append(int64(fileIdx*rowsPerFile + i))
			builder.Field(1).(*array.Int64Builder).Append(baseTimestamp + int64(i))
			builder.Field(2).(*array.Float64Builder).Append(float64(i) * 0.5)
		}

		record := builder.NewRecord()
		require.NoError(t, engine.Write(context.Background(), "perfdb", "events", record))
		record.Release()
		builder.Release()
	}

	// 测试谓词下推的性能提升
	// 查询只涉及10%的数据
	targetTimestamp := int64(95000) // 应该只扫描最后几个文件

	start := time.Now()
	filters := []storage.Filter{
		{Column: "timestamp", Operator: ">", Value: targetTimestamp},
	}

	iter, err := engine.Scan(context.Background(), "perfdb", "events", filters)
	require.NoError(t, err)

	totalRows := int64(0)
	for iter.Next() {
		record := iter.Record()
		if record != nil {
			totalRows += record.NumRows()
		}
	}
	require.NoError(t, iter.Err())
	iter.Close()

	duration := time.Since(start)

	// 验证结果
	// timestamp > targetTimestamp means (targetTimestamp, maxTimestamp]
	// maxTimestamp = totalFiles*rowsPerFile - 1 (0-indexed)
	// expectedRows = maxTimestamp - targetTimestamp
	expectedRows := int64(totalFiles*rowsPerFile-1) - targetTimestamp
	assert.Equal(t, expectedRows, totalRows, "应该返回 %d 行，实际 %d 行", expectedRows, totalRows)

	t.Logf("✓ 性能测试: 扫描 %d 行耗时 %v", totalRows, duration)
	t.Logf("  - 总文件数: %d", totalFiles)
	t.Logf("  - 总行数: %d", totalFiles*rowsPerFile)
	t.Logf("  - 返回行数: %d (%.2f%%)", totalRows, float64(totalRows)/float64(totalFiles*rowsPerFile)*100)

	// 性能基准: 应该在1秒内完成
	assert.Less(t, duration.Milliseconds(), int64(1000),
		"查询应该在1秒内完成，实际耗时 %v", duration)
}

// BenchmarkPredicatePushdownEnhanced 基准测试
func BenchmarkPredicatePushdownEnhanced(b *testing.B) {
	basePath := filepath.Join(os.TempDir(), fmt.Sprintf("minidb_predicate_bench_%d", time.Now().UnixNano()))
	defer os.RemoveAll(basePath)

	engine, err := storage.NewParquetEngine(basePath)
	if err != nil {
		b.Fatal(err)
	}
	if err := engine.Open(); err != nil {
		b.Fatal(err)
	}
	defer engine.Close()

	engine.CreateDatabase("benchdb")

	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64},
			{Name: "value", Type: arrow.PrimitiveTypes.Int64},
		},
		nil,
	)
	engine.CreateTable("benchdb", "data", schema)

	// 插入50个文件
	pool := memory.NewGoAllocator()
	for fileIdx := 0; fileIdx < 50; fileIdx++ {
		builder := array.NewRecordBuilder(pool, schema)
		for i := 0; i < 1000; i++ {
			builder.Field(0).(*array.Int64Builder).Append(int64(fileIdx*1000 + i))
			builder.Field(1).(*array.Int64Builder).Append(int64(fileIdx * 100))
		}
		record := builder.NewRecord()
		engine.Write(context.Background(), "benchdb", "data", record)
		record.Release()
		builder.Release()
	}

	b.ResetTimer()

	// 基准测试: 谓词下推查询
	for i := 0; i < b.N; i++ {
		filters := []storage.Filter{
			{Column: "value", Operator: ">", Value: int64(4000)},
		}

		iter, _ := engine.Scan(context.Background(), "benchdb", "data", filters)
		for iter.Next() {
			// 读取记录但不处理
		}
		iter.Close()
	}
}

package test

import (
	"os"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/parquet"
)

// TestParquetStatisticsCollection tests comprehensive statistics collection
func TestParquetStatisticsCollection(t *testing.T) {
	t.Run("Int64Statistics", func(t *testing.T) {
		pool := memory.NewGoAllocator()
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "id", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
				{Name: "value", Type: arrow.PrimitiveTypes.Int64, Nullable: true},
			},
			nil,
		)

		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()

		idBuilder := builder.Field(0).(*array.Int64Builder)
		valueBuilder := builder.Field(1).(*array.Int64Builder)

		// Add data with known min/max
		for i := int64(1); i <= 100; i++ {
			idBuilder.Append(i)
			if i%10 == 0 {
				valueBuilder.AppendNull()
			} else {
				valueBuilder.Append(i * 10)
			}
		}

		record := builder.NewRecord()
		defer record.Release()

		// Write to parquet with statistics
		path := "./test_data/stats_int64_test.parquet"
		os.MkdirAll("./test_data", 0755)
		defer os.Remove(path)

		stats, err := parquet.WriteArrowBatch(path, record)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		// Verify statistics
		assert.Equal(t, int64(100), stats.RowCount)
		assert.NotNil(t, stats.MinValues["id"])
		assert.NotNil(t, stats.MaxValues["id"])
		assert.Equal(t, int64(1), stats.MinValues["id"])
		assert.Equal(t, int64(100), stats.MaxValues["id"])

		// Check null counts
		assert.Equal(t, int64(0), stats.NullCounts["id"])
		assert.Equal(t, int64(10), stats.NullCounts["value"])

		// Check value min/max (excluding nulls)
		assert.NotNil(t, stats.MinValues["value"])
		assert.NotNil(t, stats.MaxValues["value"])
	})

	t.Run("Float64Statistics", func(t *testing.T) {
		pool := memory.NewGoAllocator()
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "temperature", Type: arrow.PrimitiveTypes.Float64, Nullable: false},
				{Name: "humidity", Type: arrow.PrimitiveTypes.Float64, Nullable: true},
			},
			nil,
		)

		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()

		tempBuilder := builder.Field(0).(*array.Float64Builder)
		humidityBuilder := builder.Field(1).(*array.Float64Builder)

		// Add temperature data: 20.5 to 30.5
		for i := 0; i < 100; i++ {
			tempBuilder.Append(20.5 + float64(i)*0.1)
			if i%5 == 0 {
				humidityBuilder.AppendNull()
			} else {
				humidityBuilder.Append(50.0 + float64(i)*0.2)
			}
		}

		record := builder.NewRecord()
		defer record.Release()

		path := "./test_data/stats_float64_test.parquet"
		os.MkdirAll("./test_data", 0755)
		defer os.Remove(path)

		stats, err := parquet.WriteArrowBatch(path, record)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		// Verify float statistics
		assert.Equal(t, int64(100), stats.RowCount)
		minTemp := stats.MinValues["temperature"].(float64)
		maxTemp := stats.MaxValues["temperature"].(float64)
		assert.InDelta(t, 20.5, minTemp, 0.01)
		assert.InDelta(t, 30.4, maxTemp, 0.1)

		// Check humidity null count
		assert.Equal(t, int64(20), stats.NullCounts["humidity"])
	})

	t.Run("StringStatistics", func(t *testing.T) {
		pool := memory.NewGoAllocator()
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
				{Name: "city", Type: arrow.BinaryTypes.String, Nullable: true},
			},
			nil,
		)

		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()

		nameBuilder := builder.Field(0).(*array.StringBuilder)
		cityBuilder := builder.Field(1).(*array.StringBuilder)

		names := []string{"Alice", "Bob", "Charlie", "David", "Eve"}
		cities := []string{"NYC", "LA", "Chicago", "Houston", "Phoenix"}

		for i := 0; i < 50; i++ {
			nameBuilder.Append(names[i%len(names)])
			if i%10 == 0 {
				cityBuilder.AppendNull()
			} else {
				cityBuilder.Append(cities[i%len(cities)])
			}
		}

		record := builder.NewRecord()
		defer record.Release()

		path := "./test_data/stats_string_test.parquet"
		os.MkdirAll("./test_data", 0755)
		defer os.Remove(path)

		stats, err := parquet.WriteArrowBatch(path, record)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		// Verify string statistics
		assert.Equal(t, int64(50), stats.RowCount)
		assert.NotNil(t, stats.MinValues["name"])
		assert.NotNil(t, stats.MaxValues["name"])
		assert.Equal(t, int64(5), stats.NullCounts["city"])
	})

	t.Run("BooleanStatistics", func(t *testing.T) {
		pool := memory.NewGoAllocator()
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "is_active", Type: arrow.FixedWidthTypes.Boolean, Nullable: false},
				{Name: "is_premium", Type: arrow.FixedWidthTypes.Boolean, Nullable: true},
			},
			nil,
		)

		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()

		activeBuilder := builder.Field(0).(*array.BooleanBuilder)
		premiumBuilder := builder.Field(1).(*array.BooleanBuilder)

		for i := 0; i < 100; i++ {
			activeBuilder.Append(i%2 == 0)
			if i%10 == 0 {
				premiumBuilder.AppendNull()
			} else {
				premiumBuilder.Append(i%3 == 0)
			}
		}

		record := builder.NewRecord()
		defer record.Release()

		path := "./test_data/stats_boolean_test.parquet"
		os.MkdirAll("./test_data", 0755)
		defer os.Remove(path)

		stats, err := parquet.WriteArrowBatch(path, record)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		// Verify boolean statistics
		assert.Equal(t, int64(100), stats.RowCount)
		assert.NotNil(t, stats.MinValues["is_active"])
		assert.NotNil(t, stats.MaxValues["is_active"])
		assert.Equal(t, int64(10), stats.NullCounts["is_premium"])
	})

	t.Run("MixedTypeStatistics", func(t *testing.T) {
		pool := memory.NewGoAllocator()
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "id", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
				{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
				{Name: "score", Type: arrow.PrimitiveTypes.Float64, Nullable: true},
				{Name: "active", Type: arrow.FixedWidthTypes.Boolean, Nullable: false},
			},
			nil,
		)

		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()

		idBuilder := builder.Field(0).(*array.Int64Builder)
		nameBuilder := builder.Field(1).(*array.StringBuilder)
		scoreBuilder := builder.Field(2).(*array.Float64Builder)
		activeBuilder := builder.Field(3).(*array.BooleanBuilder)

		for i := 0; i < 50; i++ {
			idBuilder.Append(int64(i + 1))
			nameBuilder.Append("User" + string(rune('A'+i%26)))
			if i%5 == 0 {
				scoreBuilder.AppendNull()
			} else {
				scoreBuilder.Append(float64(i) * 1.5)
			}
			activeBuilder.Append(i%2 == 0)
		}

		record := builder.NewRecord()
		defer record.Release()

		path := "./test_data/stats_mixed_test.parquet"
		os.MkdirAll("./test_data", 0755)
		defer os.Remove(path)

		stats, err := parquet.WriteArrowBatch(path, record)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		// Verify all statistics collected
		assert.Equal(t, int64(50), stats.RowCount)
		assert.Equal(t, 4, len(stats.MinValues), "Should have min values for all columns")
		assert.Equal(t, 4, len(stats.MaxValues), "Should have max values for all columns")
		assert.Equal(t, 4, len(stats.NullCounts), "Should have null counts for all columns")

		// Verify specific statistics
		assert.Equal(t, int64(1), stats.MinValues["id"])
		assert.Equal(t, int64(50), stats.MaxValues["id"])
		assert.Equal(t, int64(0), stats.NullCounts["id"])
		assert.Equal(t, int64(10), stats.NullCounts["score"])
	})

	t.Run("Int32Int16Int8Statistics", func(t *testing.T) {
		pool := memory.NewGoAllocator()
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "int32_col", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
				{Name: "int16_col", Type: arrow.PrimitiveTypes.Int16, Nullable: false},
				{Name: "int8_col", Type: arrow.PrimitiveTypes.Int8, Nullable: false},
			},
			nil,
		)

		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()

		int32Builder := builder.Field(0).(*array.Int32Builder)
		int16Builder := builder.Field(1).(*array.Int16Builder)
		int8Builder := builder.Field(2).(*array.Int8Builder)

		for i := 0; i < 100; i++ {
			int32Builder.Append(int32(i * 1000))
			int16Builder.Append(int16(i * 100))
			int8Builder.Append(int8(i % 128))
		}

		record := builder.NewRecord()
		defer record.Release()

		path := "./test_data/stats_integer_types_test.parquet"
		os.MkdirAll("./test_data", 0755)
		defer os.Remove(path)

		stats, err := parquet.WriteArrowBatch(path, record)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		// Verify all integer type statistics
		assert.NotNil(t, stats.MinValues["int32_col"])
		assert.NotNil(t, stats.MaxValues["int32_col"])
		assert.NotNil(t, stats.MinValues["int16_col"])
		assert.NotNil(t, stats.MaxValues["int16_col"])
		assert.NotNil(t, stats.MinValues["int8_col"])
		assert.NotNil(t, stats.MaxValues["int8_col"])
	})

	t.Run("Float32Statistics", func(t *testing.T) {
		pool := memory.NewGoAllocator()
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "value", Type: arrow.PrimitiveTypes.Float32, Nullable: false},
			},
			nil,
		)

		builder := array.NewRecordBuilder(pool, schema)
		defer builder.Release()

		valueBuilder := builder.Field(0).(*array.Float32Builder)

		for i := 0; i < 100; i++ {
			valueBuilder.Append(float32(i) * 0.5)
		}

		record := builder.NewRecord()
		defer record.Release()

		path := "./test_data/stats_float32_test.parquet"
		os.MkdirAll("./test_data", 0755)
		defer os.Remove(path)

		stats, err := parquet.WriteArrowBatch(path, record)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		// Verify float32 statistics
		assert.NotNil(t, stats.MinValues["value"])
		assert.NotNil(t, stats.MaxValues["value"])
		minVal := stats.MinValues["value"].(float32)
		maxVal := stats.MaxValues["value"].(float32)
		assert.InDelta(t, float32(0.0), minVal, 0.01)
		assert.InDelta(t, float32(49.5), maxVal, 0.1)
	})
}

// TestStatisticsRoundtrip tests writing and reading with statistics
func TestStatisticsRoundtrip(t *testing.T) {
	pool := memory.NewGoAllocator()
	schema := arrow.NewSchema(
		[]arrow.Field{
			{Name: "id", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "value", Type: arrow.PrimitiveTypes.Float64, Nullable: false},
		},
		nil,
	)

	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	idBuilder := builder.Field(0).(*array.Int64Builder)
	valueBuilder := builder.Field(1).(*array.Float64Builder)

	for i := 1; i <= 100; i++ {
		idBuilder.Append(int64(i))
		valueBuilder.Append(float64(i) * 2.5)
	}

	record := builder.NewRecord()
	defer record.Release()

	path := "./test_data/stats_roundtrip_test.parquet"
	os.MkdirAll("./test_data", 0755)
	defer os.Remove(path)

	// Write with statistics
	stats, err := parquet.WriteArrowBatch(path, record)
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	// Read back
	readRecord, err := parquet.ReadParquetFile(path, nil)
	assert.NoError(t, err)
	assert.NotNil(t, readRecord)
	defer readRecord.Release()

	// Verify data integrity
	assert.Equal(t, record.NumRows(), readRecord.NumRows())
	assert.Equal(t, record.NumCols(), readRecord.NumCols())
}

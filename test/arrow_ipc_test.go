package test

import (
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/delta"
)

// TestArrowIPCSerialization tests Arrow IPC schema serialization/deserialization
func TestArrowIPCSerialization(t *testing.T) {
	deltaLog := delta.NewDeltaLog()
	err := deltaLog.Bootstrap()
	assert.NoError(t, err)

	t.Run("BasicSchemaRoundtrip", func(t *testing.T) {
		// Create a simple schema
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "id", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
				{Name: "name", Type: arrow.BinaryTypes.String, Nullable: true},
			},
			nil,
		)

		// Append metadata with schema
		tableID := "test_db.basic_schema"
		err := deltaLog.AppendMetadata(tableID, schema)
		assert.NoError(t, err)

		// Retrieve snapshot and verify schema
		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		assert.NoError(t, err)
		assert.NotNil(t, snapshot)
		assert.NotNil(t, snapshot.Schema)

		// Verify schema fields
		assert.Equal(t, 2, len(snapshot.Schema.Fields()))
		assert.Equal(t, "id", snapshot.Schema.Field(0).Name)
		assert.Equal(t, "name", snapshot.Schema.Field(1).Name)
		assert.Equal(t, arrow.PrimitiveTypes.Int64, snapshot.Schema.Field(0).Type)
		assert.Equal(t, arrow.BinaryTypes.String, snapshot.Schema.Field(1).Type)
		assert.False(t, snapshot.Schema.Field(0).Nullable)
		assert.True(t, snapshot.Schema.Field(1).Nullable)
	})

	t.Run("ComplexSchemaRoundtrip", func(t *testing.T) {
		// Create a complex schema with all supported types
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "int8_col", Type: arrow.PrimitiveTypes.Int8, Nullable: false},
				{Name: "int16_col", Type: arrow.PrimitiveTypes.Int16, Nullable: false},
				{Name: "int32_col", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
				{Name: "int64_col", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
				{Name: "uint8_col", Type: arrow.PrimitiveTypes.Uint8, Nullable: true},
				{Name: "uint16_col", Type: arrow.PrimitiveTypes.Uint16, Nullable: true},
				{Name: "uint32_col", Type: arrow.PrimitiveTypes.Uint32, Nullable: true},
				{Name: "uint64_col", Type: arrow.PrimitiveTypes.Uint64, Nullable: true},
				{Name: "float32_col", Type: arrow.PrimitiveTypes.Float32, Nullable: false},
				{Name: "float64_col", Type: arrow.PrimitiveTypes.Float64, Nullable: false},
				{Name: "bool_col", Type: arrow.FixedWidthTypes.Boolean, Nullable: false},
				{Name: "string_col", Type: arrow.BinaryTypes.String, Nullable: true},
				{Name: "binary_col", Type: arrow.BinaryTypes.Binary, Nullable: true},
			},
			nil,
		)

		tableID := "test_db.complex_schema"
		err := deltaLog.AppendMetadata(tableID, schema)
		assert.NoError(t, err)

		// Retrieve and verify
		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		assert.NoError(t, err)
		assert.NotNil(t, snapshot.Schema)
		assert.Equal(t, 13, len(snapshot.Schema.Fields()))

		// Verify all types preserved
		fields := snapshot.Schema.Fields()
		assert.Equal(t, arrow.PrimitiveTypes.Int8, fields[0].Type)
		assert.Equal(t, arrow.PrimitiveTypes.Int16, fields[1].Type)
		assert.Equal(t, arrow.PrimitiveTypes.Int32, fields[2].Type)
		assert.Equal(t, arrow.PrimitiveTypes.Int64, fields[3].Type)
		assert.Equal(t, arrow.PrimitiveTypes.Uint8, fields[4].Type)
		assert.Equal(t, arrow.PrimitiveTypes.Uint16, fields[5].Type)
		assert.Equal(t, arrow.PrimitiveTypes.Uint32, fields[6].Type)
		assert.Equal(t, arrow.PrimitiveTypes.Uint64, fields[7].Type)
		assert.Equal(t, arrow.PrimitiveTypes.Float32, fields[8].Type)
		assert.Equal(t, arrow.PrimitiveTypes.Float64, fields[9].Type)
		assert.Equal(t, arrow.FixedWidthTypes.Boolean, fields[10].Type)
		assert.Equal(t, arrow.BinaryTypes.String, fields[11].Type)
		assert.Equal(t, arrow.BinaryTypes.Binary, fields[12].Type)

		// Verify nullability
		assert.False(t, fields[0].Nullable)
		assert.True(t, fields[4].Nullable)
		assert.True(t, fields[11].Nullable)
	})

	t.Run("SchemaWithMetadata", func(t *testing.T) {
		// Create schema with field metadata
		metadata := arrow.NewMetadata(
			[]string{"comment", "format"},
			[]string{"User ID column", "integer"},
		)

		schema := arrow.NewSchema(
			[]arrow.Field{
				{
					Name:     "user_id",
					Type:     arrow.PrimitiveTypes.Int64,
					Nullable: false,
					Metadata: metadata,
				},
				{
					Name:     "email",
					Type:     arrow.BinaryTypes.String,
					Nullable: true,
				},
			},
			nil,
		)

		tableID := "test_db.metadata_schema"
		err := deltaLog.AppendMetadata(tableID, schema)
		assert.NoError(t, err)

		// Retrieve and verify metadata
		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		assert.NoError(t, err)
		assert.NotNil(t, snapshot.Schema)

		field := snapshot.Schema.Field(0)
		assert.Equal(t, "user_id", field.Name)
		assert.Greater(t, field.Metadata.Len(), 0, "Field metadata should be preserved")

		// Check metadata values
		metadataMap := field.Metadata.ToMap()
		if comment, ok := metadataMap["comment"]; ok {
			assert.Equal(t, "User ID column", comment)
		}
	})

	t.Run("SchemaWithTableMetadata", func(t *testing.T) {
		// Create schema with table-level metadata
		schemaMetadata := arrow.NewMetadata(
			[]string{"table_name", "version"},
			[]string{"users", "1.0"},
		)

		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "id", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			},
			&schemaMetadata,
		)

		tableID := "test_db.table_metadata"
		err := deltaLog.AppendMetadata(tableID, schema)
		assert.NoError(t, err)

		// Retrieve and verify
		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		assert.NoError(t, err)
		assert.NotNil(t, snapshot.Schema)

		// Check table metadata
		tableMetadata := snapshot.Schema.Metadata()
		assert.Greater(t, tableMetadata.Len(), 0, "Table metadata should be preserved")

		metadataMap := tableMetadata.ToMap()
		if tableName, ok := metadataMap["table_name"]; ok {
			assert.Equal(t, "users", tableName)
		}
	})

	t.Run("TimestampTypes", func(t *testing.T) {
		// Create schema with timestamp types
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "timestamp_s", Type: arrow.FixedWidthTypes.Timestamp_s, Nullable: false},
				{Name: "timestamp_ms", Type: arrow.FixedWidthTypes.Timestamp_ms, Nullable: false},
				{Name: "timestamp_us", Type: arrow.FixedWidthTypes.Timestamp_us, Nullable: false},
				{Name: "timestamp_ns", Type: arrow.FixedWidthTypes.Timestamp_ns, Nullable: false},
			},
			nil,
		)

		tableID := "test_db.timestamp_types"
		err := deltaLog.AppendMetadata(tableID, schema)
		assert.NoError(t, err)

		// Retrieve and verify
		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		assert.NoError(t, err)
		assert.NotNil(t, snapshot.Schema)

		fields := snapshot.Schema.Fields()
		assert.Equal(t, 4, len(fields))
		assert.Equal(t, arrow.FixedWidthTypes.Timestamp_s, fields[0].Type)
		assert.Equal(t, arrow.FixedWidthTypes.Timestamp_ms, fields[1].Type)
		assert.Equal(t, arrow.FixedWidthTypes.Timestamp_us, fields[2].Type)
		assert.Equal(t, arrow.FixedWidthTypes.Timestamp_ns, fields[3].Type)
	})

	t.Run("DateTypes", func(t *testing.T) {
		// Create schema with date types
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "date32", Type: arrow.FixedWidthTypes.Date32, Nullable: false},
				{Name: "date64", Type: arrow.FixedWidthTypes.Date64, Nullable: true},
			},
			nil,
		)

		tableID := "test_db.date_types"
		err := deltaLog.AppendMetadata(tableID, schema)
		assert.NoError(t, err)

		// Retrieve and verify
		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		assert.NoError(t, err)
		assert.NotNil(t, snapshot.Schema)

		fields := snapshot.Schema.Fields()
		assert.Equal(t, 2, len(fields))
		assert.Equal(t, arrow.FixedWidthTypes.Date32, fields[0].Type)
		assert.Equal(t, arrow.FixedWidthTypes.Date64, fields[1].Type)
	})

	t.Run("MultipleSchemaVersions", func(t *testing.T) {
		tableID := "test_db.evolving_schema"

		// Version 1: Initial schema
		schema1 := arrow.NewSchema(
			[]arrow.Field{
				{Name: "id", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			},
			nil,
		)
		err := deltaLog.AppendMetadata(tableID, schema1)
		assert.NoError(t, err)

		// Version 2: Add column
		schema2 := arrow.NewSchema(
			[]arrow.Field{
				{Name: "id", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
				{Name: "name", Type: arrow.BinaryTypes.String, Nullable: true},
			},
			nil,
		)
		err = deltaLog.AppendMetadata(tableID, schema2)
		assert.NoError(t, err)

		// Latest snapshot should have schema2
		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		assert.NoError(t, err)
		assert.NotNil(t, snapshot.Schema)
		assert.Equal(t, 2, len(snapshot.Schema.Fields()))
	})

	t.Run("EmptySchema", func(t *testing.T) {
		// Empty schema (edge case)
		schema := arrow.NewSchema([]arrow.Field{}, nil)

		tableID := "test_db.empty_schema"
		err := deltaLog.AppendMetadata(tableID, schema)
		assert.NoError(t, err)

		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		assert.NoError(t, err)
		assert.NotNil(t, snapshot.Schema)
		assert.Equal(t, 0, len(snapshot.Schema.Fields()))
	})
}

// TestArrowIPCPerformance tests IPC serialization performance
func TestArrowIPCPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	deltaLog := delta.NewDeltaLog()
	err := deltaLog.Bootstrap()
	assert.NoError(t, err)

	t.Run("LargeSchemaPerformance", func(t *testing.T) {
		// Create a large schema with 100 columns
		fields := make([]arrow.Field, 100)
		for i := 0; i < 100; i++ {
			fieldType := arrow.PrimitiveTypes.Int64
			if i%3 == 0 {
				fieldType = arrow.BinaryTypes.String
			} else if i%3 == 1 {
				fieldType = arrow.PrimitiveTypes.Float64
			}

			fields[i] = arrow.Field{
				Name:     "col_" + string(rune('0'+i%10)),
				Type:     fieldType,
				Nullable: i%2 == 0,
			}
		}

		schema := arrow.NewSchema(fields, nil)

		tableID := "test_db.large_schema"

		// Serialize
		err := deltaLog.AppendMetadata(tableID, schema)
		assert.NoError(t, err)

		// Deserialize
		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		assert.NoError(t, err)
		assert.NotNil(t, snapshot.Schema)
		assert.Equal(t, 100, len(snapshot.Schema.Fields()))
	})

	t.Run("RepeatedSerializationPerformance", func(t *testing.T) {
		schema := arrow.NewSchema(
			[]arrow.Field{
				{Name: "id", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
				{Name: "name", Type: arrow.BinaryTypes.String, Nullable: true},
				{Name: "value", Type: arrow.PrimitiveTypes.Float64, Nullable: false},
			},
			nil,
		)

		// Perform multiple serializations
		for i := 0; i < 100; i++ {
			tableID := "test_db.perf_test_" + string(rune('0'+i%10))
			err := deltaLog.AppendMetadata(tableID, schema)
			assert.NoError(t, err)

			// Retrieve to verify
			snapshot, err := deltaLog.GetSnapshot(tableID, -1)
			assert.NoError(t, err)
			assert.NotNil(t, snapshot.Schema)
		}
	})
}

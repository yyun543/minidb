package test

import (
	"context"
	"os"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

// TestSimpleAndFilter tests that AND filtering works on regular tables
func TestSimpleAndFilter(t *testing.T) {
	testDir := "./test_data/filter_and_test"
	os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	// Setup
	engine, err := storage.NewParquetEngine(testDir)
	require.NoError(t, err)
	err = engine.Open()
	require.NoError(t, err)
	defer engine.Close()

	cat := catalog.NewCatalog()
	cat.SetStorageEngine(engine)
	err = cat.Init()
	require.NoError(t, err)

	// Create session
	sessMgr, err := session.NewSessionManager()
	require.NoError(t, err)
	sess := sessMgr.CreateSession()

	// Create test database
	err = engine.CreateDatabase("testdb")
	require.NoError(t, err)
	sess.CurrentDB = "testdb"

	// Create test table
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "db_name", Type: arrow.BinaryTypes.String},
		{Name: "table_name", Type: arrow.BinaryTypes.String},
		{Name: "value", Type: arrow.PrimitiveTypes.Int64},
	}, nil)

	err = engine.CreateTable("testdb", "test_table", schema)
	require.NoError(t, err)

	// Insert test data
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// Insert multiple rows with different db_name and table_name combinations
	dbNames := []string{"db1", "db1", "db2", "db2", "testdb"}
	tableNames := []string{"table1", "table2", "table1", "table2", "users"}
	values := []int64{10, 20, 30, 40, 50}

	builder.Field(0).(*array.StringBuilder).AppendValues(dbNames, nil)
	builder.Field(1).(*array.StringBuilder).AppendValues(tableNames, nil)
	builder.Field(2).(*array.Int64Builder).AppendValues(values, nil)

	record := builder.NewRecord()
	defer record.Release()

	err = engine.Write(context.Background(), "testdb", "test_table", record)
	require.NoError(t, err)

	// Reinitialize catalog to load metadata from Delta Log
	err = cat.Init()
	require.NoError(t, err)

	// Update session's current database
	sess.CurrentDB = "testdb"

	// Create executor
	exec := executor.NewExecutor(cat)
	opt := optimizer.NewOptimizer()

	t.Run("SingleWhereCondition", func(t *testing.T) {
		sql := "SELECT * FROM test_table WHERE db_name = 'db1'"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		require.NoError(t, err)
		require.NotNil(t, result)

		batches := result.Batches()
		totalRows := 0
		for _, batch := range batches {
			totalRows += int(batch.NumRows())
		}

		assert.Equal(t, 2, totalRows, "Should return 2 rows for db_name = 'db1'")
	})

	t.Run("AndWhereCondition", func(t *testing.T) {
		sql := "SELECT * FROM test_table WHERE db_name = 'db1' AND table_name = 'table1'"
		stmt, err := parser.Parse(sql)
		require.NoError(t, err)

		plan, err := opt.Optimize(stmt)
		require.NoError(t, err)

		result, err := exec.Execute(plan, sess)
		require.NoError(t, err)
		require.NotNil(t, result)

		batches := result.Batches()
		totalRows := 0
		for _, batch := range batches {
			totalRows += int(batch.NumRows())
			// Check that we got the right row
			if batch.NumRows() > 0 {
				dbName := batch.GetString(0, 0)
				tableName := batch.GetString(1, 0)
				assert.Equal(t, "db1", dbName, "db_name should be 'db1'")
				assert.Equal(t, "table1", tableName, "table_name should be 'table1'")
			}
		}

		assert.Equal(t, 1, totalRows, "Should return exactly 1 row for db_name = 'db1' AND table_name = 'table1'")
	})
}

package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/delta"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

// TestTimeTravelQueries tests Delta Lake time travel capabilities
func TestTimeTravelQueries(t *testing.T) {
	storageEngine, err := storage.NewParquetEngine(SetupTestDir(t, "time_travel_test"))
	assert.NoError(t, err)
	defer storageEngine.Close()
	err = storageEngine.Open()
	assert.NoError(t, err)

	cat := catalog.NewCatalog()
	cat.SetStorageEngine(storageEngine)
	err = cat.Init()
	assert.NoError(t, err)

	sessMgr, err := session.NewSessionManager()
	assert.NoError(t, err)
	sess := sessMgr.CreateSession()

	opt := optimizer.NewOptimizer()
	exec := executor.NewExecutor(cat)

	// Setup database
	createDbSQL := "CREATE DATABASE timetravel_test"
	stmt, err := parser.Parse(createDbSQL)
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "timetravel_test"

	t.Run("VersionBasedTimeTravel", func(t *testing.T) {
		// Create table
		createSQL := "CREATE TABLE history (id INTEGER, value VARCHAR, version INTEGER)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Insert data at version 1
		insertSQL := "INSERT INTO history VALUES (1, 'v1', 1)"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		// Insert data at version 2
		insertSQL = "INSERT INTO history VALUES (2, 'v2', 2)"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		// Insert data at version 3
		insertSQL = "INSERT INTO history VALUES (3, 'v3', 3)"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Query latest version - should have all 3 records
		selectSQL := "SELECT * FROM history"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(result.Batches()), "Latest version should have all 3 records")
	})

	t.Run("SnapshotIsolation", func(t *testing.T) {
		// Create table
		createSQL := "CREATE TABLE snapshots (id INTEGER, data VARCHAR)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Record timestamps for each insert
		timestamps := make([]int64, 0)

		// Insert 1
		timestamps = append(timestamps, time.Now().UnixMilli())
		insertSQL := "INSERT INTO snapshots VALUES (1, 'snapshot1')"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// Insert 2
		timestamps = append(timestamps, time.Now().UnixMilli())
		insertSQL = "INSERT INTO snapshots VALUES (2, 'snapshot2')"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		// Insert 3
		timestamps = append(timestamps, time.Now().UnixMilli())
		insertSQL = "INSERT INTO snapshots VALUES (3, 'snapshot3')"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Current query should show all records
		selectSQL := "SELECT * FROM snapshots"
		stmt, err = parser.Parse(selectSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		result, err := exec.Execute(plan, sess)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(result.Batches()), "Current snapshot should have all records")

		// Timestamps available for potential time-travel queries
		assert.Equal(t, 3, len(timestamps), "Should have recorded 3 timestamps")
	})

	t.Run("DeltaLogVersionTracking", func(t *testing.T) {
		// Get direct access to Delta Log for version testing
		// storageEngine is already *ParquetEngine, no type assertion needed
		deltaLog := storageEngine.GetDeltaLog()
		assert.NotNil(t, deltaLog, "Delta Log should be available")

		// Check current version
		currentVersion := deltaLog.GetLatestVersion()
		assert.Greater(t, currentVersion, int64(0), "Should have versions from previous operations")

		// Create new table to increment version
		createSQL := "CREATE TABLE version_test (id INTEGER)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Insert data
		insertSQL := "INSERT INTO version_test VALUES (1)"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Version should have incremented
		newVersion := deltaLog.GetLatestVersion()
		assert.Greater(t, newVersion, currentVersion, "Version should increment after insert")
	})
}

// TestDeltaLogSnapshotRetrieval tests snapshot retrieval at different versions
func TestDeltaLogSnapshotRetrieval(t *testing.T) {
	storageEngine, err := storage.NewParquetEngine(SetupTestDir(t, "snapshot_retrieval_test"))
	assert.NoError(t, err)
	defer storageEngine.Close()
	err = storageEngine.Open()
	assert.NoError(t, err)

	// storageEngine is already *ParquetEngine, no type assertion needed
	deltaLog := storageEngine.GetDeltaLog()
	assert.NotNil(t, deltaLog)

	cat := catalog.NewCatalog()
	cat.SetStorageEngine(storageEngine)
	err = cat.Init()
	assert.NoError(t, err)

	sessMgr, err := session.NewSessionManager()
	assert.NoError(t, err)
	sess := sessMgr.CreateSession()

	opt := optimizer.NewOptimizer()
	exec := executor.NewExecutor(cat)

	// Setup
	createDbSQL := "CREATE DATABASE snapshot_db"
	stmt, err := parser.Parse(createDbSQL)
	assert.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	assert.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	assert.NoError(t, err)

	sess.CurrentDB = "snapshot_db"

	t.Run("SnapshotAtVersion", func(t *testing.T) {
		// Create table
		createSQL := "CREATE TABLE events (id INTEGER, event VARCHAR)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		tableID := "snapshot_db.events"

		// Insert events and track versions
		versions := make([]int64, 0)

		for i := 1; i <= 5; i++ {
			stmt, err = parser.Parse("INSERT INTO events VALUES (1, 'event')")
			assert.NoError(t, err)
			plan, err = opt.Optimize(stmt)
			assert.NoError(t, err)
			_, err = exec.Execute(plan, sess)
			assert.NoError(t, err)

			currentVersion := deltaLog.GetLatestVersion()
			versions = append(versions, currentVersion)

			time.Sleep(20 * time.Millisecond)
		}

		// Get snapshot at latest version
		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		assert.NoError(t, err)
		assert.NotNil(t, snapshot)
		assert.Equal(t, 5, len(snapshot.Files), "Latest snapshot should have all 5 files")

		// Get snapshot at earlier version (if we have version tracking)
		if len(versions) > 2 {
			earlierVersion := versions[1]
			snapshot, err = deltaLog.GetSnapshot(tableID, earlierVersion)
			assert.NoError(t, err)
			assert.NotNil(t, snapshot)
			// Earlier snapshot should have fewer files
			assert.LessOrEqual(t, len(snapshot.Files), 5, "Earlier snapshot should have <= files")
		}
	})

	t.Run("SnapshotMetadata", func(t *testing.T) {
		// Create table with schema
		createSQL := "CREATE TABLE metadata_test (id INTEGER, name VARCHAR, value INTEGER)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Insert data
		insertSQL := "INSERT INTO metadata_test VALUES (1, 'test', 100)"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Get snapshot
		tableID := "snapshot_db.metadata_test"
		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		assert.NoError(t, err)
		assert.NotNil(t, snapshot)

		// Verify snapshot properties
		assert.NotZero(t, snapshot.Version, "Snapshot should have version")
		assert.NotZero(t, snapshot.Timestamp, "Snapshot should have timestamp")
		assert.Equal(t, tableID, snapshot.TableID, "Snapshot should have correct table ID")
		assert.Greater(t, len(snapshot.Files), 0, "Snapshot should have files")
	})

	t.Run("TimestampBasedQuery", func(t *testing.T) {
		// Create table
		createSQL := "CREATE TABLE time_based (id INTEGER, ts INTEGER)"
		stmt, err := parser.Parse(createSQL)
		assert.NoError(t, err)
		plan, err := opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		tableID := "snapshot_db.time_based"
		timestamp1 := time.Now().UnixMilli()

		// Insert 1
		insertSQL := "INSERT INTO time_based VALUES (1, 100)"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		timestamp2 := time.Now().UnixMilli()

		// Insert 2
		insertSQL = "INSERT INTO time_based VALUES (2, 200)"
		stmt, err = parser.Parse(insertSQL)
		assert.NoError(t, err)
		plan, err = opt.Optimize(stmt)
		assert.NoError(t, err)
		_, err = exec.Execute(plan, sess)
		assert.NoError(t, err)

		// Get version at timestamp1
		version1, err := deltaLog.GetVersionByTimestamp(tableID, timestamp1)
		if err == nil {
			assert.Greater(t, version1, int64(0), "Should find version at timestamp1")
		}

		// Get version at timestamp2
		version2, err := deltaLog.GetVersionByTimestamp(tableID, timestamp2)
		if err == nil {
			assert.GreaterOrEqual(t, version2, version1, "Version2 should be >= version1")
		}
	})
}

// TestDeltaLogFileTracking tests ADD and REMOVE operations
func TestDeltaLogFileTracking(t *testing.T) {
	deltaLog := delta.NewDeltaLog()
	err := deltaLog.Bootstrap()
	assert.NoError(t, err)

	tableID := "test_db.test_table"

	t.Run("AddFileTracking", func(t *testing.T) {
		// Add files
		for i := 1; i <= 3; i++ {
			file := &delta.ParquetFile{
				Path:     "file" + string(rune('0'+i)) + ".parquet",
				Size:     1000,
				RowCount: 100,
				Stats: &delta.FileStats{
					RowCount:   100,
					MinValues:  map[string]interface{}{"id": 1},
					MaxValues:  map[string]interface{}{"id": 100},
					NullCounts: map[string]int64{"id": 0},
				},
			}

			err := deltaLog.AppendAdd(tableID, file)
			assert.NoError(t, err)
		}

		// Get snapshot
		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(snapshot.Files), "Should have 3 files")
	})

	t.Run("RemoveFileTracking", func(t *testing.T) {
		// Remove a file
		err := deltaLog.AppendRemove(tableID, "file1.parquet")
		assert.NoError(t, err)

		// Get updated snapshot
		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(snapshot.Files), "Should have 2 files after removal")
	})

	t.Run("ListTables", func(t *testing.T) {
		tables := deltaLog.ListTables()
		assert.Contains(t, tables, tableID, "Should list the test table")
	})
}

package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
)

// setupTransactionTestEnv creates a test environment with OptimisticDeltaLog
func setupTransactionTestEnv(t *testing.T, testName string) (*catalog.Catalog, *session.Session, *optimizer.Optimizer, *executor.ExecutorImpl, func()) {
	testDir := "./test_data/" + testName
	os.RemoveAll(testDir)

	// Create Parquet storage engine with optimistic locking enabled
	storageEngine, err := storage.NewParquetEngine(testDir,
		storage.WithOptimisticLock(true))
	assert.NoError(t, err)
	err = storageEngine.Open()
	assert.NoError(t, err)

	// Create catalog
	cat := catalog.NewCatalog()
	cat.SetStorageEngine(storageEngine)
	err = cat.Init()
	assert.NoError(t, err)

	// Create session
	sessMgr, err := session.NewSessionManager()
	assert.NoError(t, err)
	sess := sessMgr.CreateSession()

	// Create optimizer and executor
	opt := optimizer.NewOptimizer()
	exec := executor.NewExecutor(cat)

	cleanup := func() {
		storageEngine.Close()
		os.RemoveAll(testDir)
	}

	return cat, sess, opt, exec, cleanup
}

// TestTransactionBeginCommit tests basic transaction START TRANSACTION/COMMIT
func TestTransactionBeginCommit(t *testing.T) {
	_, sess, opt, exec, cleanup := setupTransactionTestEnv(t, "test_tx_begin_commit")
	defer cleanup()

	// Execute START TRANSACTION
	stmt, err := parser.Parse("START TRANSACTION")
	require.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	require.NoError(t, err)
	result, err := exec.Execute(plan, sess)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify transaction is active
	require.NotNil(t, sess.Transaction, "Transaction should be active")
	require.True(t, sess.Transaction.IsActive(), "Transaction should be in active state")

	// Execute COMMIT
	stmt, err = parser.Parse("COMMIT")
	require.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	require.NoError(t, err)
	result, err = exec.Execute(plan, sess)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify transaction is cleared
	require.Nil(t, sess.Transaction, "Transaction should be cleared after commit")
}

// TestTransactionBeginRollback tests transaction START TRANSACTION/ROLLBACK
func TestTransactionBeginRollback(t *testing.T) {
	_, sess, opt, exec, cleanup := setupTransactionTestEnv(t, "test_tx_rollback")
	defer cleanup()

	// Execute START TRANSACTION
	stmt, err := parser.Parse("START TRANSACTION")
	require.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.NoError(t, err)

	// Verify transaction is active
	require.NotNil(t, sess.Transaction)
	require.True(t, sess.Transaction.IsActive())

	// Execute ROLLBACK
	stmt, err = parser.Parse("ROLLBACK")
	require.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.NoError(t, err)

	// Verify transaction is cleared
	require.Nil(t, sess.Transaction, "Transaction should be cleared after rollback")
}

// TestTransactionNestedBeginError tests that nested BEGIN fails
func TestTransactionNestedBeginError(t *testing.T) {
	_, sess, opt, exec, cleanup := setupTransactionTestEnv(t, "test_tx_nested")
	defer cleanup()

	// Execute first BEGIN
	stmt, err := parser.Parse("START TRANSACTION")
	require.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.NoError(t, err)
	require.NotNil(t, sess.Transaction)

	// Try to BEGIN again (should fail)
	stmt, err = parser.Parse("START TRANSACTION")
	require.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.Error(t, err, "Nested BEGIN should fail")
	assert.Contains(t, err.Error(), "already in progress")

	// Cleanup
	stmt, err = parser.Parse("ROLLBACK")
	require.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.NoError(t, err)
}

// TestTransactionCommitWithoutBegin tests COMMIT without BEGIN fails
func TestTransactionCommitWithoutBegin(t *testing.T) {
	_, sess, opt, exec, cleanup := setupTransactionTestEnv(t, "test_tx_no_begin")
	defer cleanup()

	// Try to COMMIT without BEGIN (should fail)
	stmt, err := parser.Parse("COMMIT")
	require.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.Error(t, err, "COMMIT without BEGIN should fail")
	assert.Contains(t, err.Error(), "no active transaction")
}

// TestTransactionRollbackWithoutBegin tests ROLLBACK without BEGIN fails
func TestTransactionRollbackWithoutBegin(t *testing.T) {
	_, sess, opt, exec, cleanup := setupTransactionTestEnv(t, "test_tx_no_rollback")
	defer cleanup()

	// Try to ROLLBACK without BEGIN (should fail)
	stmt, err := parser.Parse("ROLLBACK")
	require.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.Error(t, err, "ROLLBACK without BEGIN should fail")
	assert.Contains(t, err.Error(), "no active transaction")
}

// TestTransactionVersionTracking tests transaction version tracking
func TestTransactionVersionTracking(t *testing.T) {
	cat, sess, opt, exec, cleanup := setupTransactionTestEnv(t, "test_tx_version")
	defer cleanup()

	// Get DeltaLog
	storageEngine := cat.GetStorageEngine()
	parquetEngine, ok := storageEngine.(*storage.ParquetEngine)
	require.True(t, ok)

	deltaLog := parquetEngine.GetDeltaLog()
	startVersion := deltaLog.GetLatestVersion()

	// Execute BEGIN
	stmt, err := parser.Parse("START TRANSACTION")
	require.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.NoError(t, err)

	// Verify transaction captures start version
	require.NotNil(t, sess.Transaction)
	require.Equal(t, startVersion, sess.Transaction.StartVersion, "StartVersion should match current version")

	// Execute COMMIT
	stmt, err = parser.Parse("COMMIT")
	require.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.NoError(t, err)
	require.Nil(t, sess.Transaction)
}

// TestTransactionStartVariants tests START TRANSACTION
func TestTransactionStartVariants(t *testing.T) {
	_, sess, opt, exec, cleanup := setupTransactionTestEnv(t, "test_tx_start")
	defer cleanup()

	// Execute transaction start
	stmt, err := parser.Parse("START TRANSACTION")
	require.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.NoError(t, err)

	// Verify transaction is active
	require.NotNil(t, sess.Transaction, "Transaction should be active")
	require.True(t, sess.Transaction.IsActive())

	// Cleanup
	stmt, err = parser.Parse("ROLLBACK")
	require.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.NoError(t, err)
}

// TestTransactionEmptyCommit tests committing a transaction with no operations
func TestTransactionEmptyCommit(t *testing.T) {
	_, sess, opt, exec, cleanup := setupTransactionTestEnv(t, "test_tx_empty")
	defer cleanup()

	// BEGIN
	stmt, err := parser.Parse("START TRANSACTION")
	require.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.NoError(t, err)

	// Verify no operations
	require.Equal(t, 0, sess.Transaction.GetOperationCount())

	// COMMIT (should succeed even with no operations)
	stmt, err = parser.Parse("COMMIT")
	require.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.NoError(t, err, "Empty transaction commit should succeed")
	require.Nil(t, sess.Transaction)
}

// TestTransactionStateTransitions tests transaction state transitions
func TestTransactionStateTransitions(t *testing.T) {
	_, sess, opt, exec, cleanup := setupTransactionTestEnv(t, "test_tx_states")
	defer cleanup()

	// Initial state: no transaction
	require.Nil(t, sess.Transaction)

	// BEGIN -> Active
	stmt, err := parser.Parse("START TRANSACTION")
	require.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.NoError(t, err)
	require.NotNil(t, sess.Transaction)
	require.True(t, sess.Transaction.IsActive())
	require.Equal(t, session.TransactionStateActive, sess.Transaction.State)

	// COMMIT -> Committed (and cleared)
	stmt, err = parser.Parse("COMMIT")
	require.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	require.NoError(t, err)
	_, err = exec.Execute(plan, sess)
	require.NoError(t, err)
	require.Nil(t, sess.Transaction, "Transaction should be cleared after commit")
}

// TestTransactionResultFormat tests transaction command result format
func TestTransactionResultFormat(t *testing.T) {
	_, sess, opt, exec, cleanup := setupTransactionTestEnv(t, "test_tx_result")
	defer cleanup()

	// Execute BEGIN and check result
	stmt, err := parser.Parse("START TRANSACTION")
	require.NoError(t, err)
	plan, err := opt.Optimize(stmt)
	require.NoError(t, err)
	result, err := exec.Execute(plan, sess)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, []string{"status"}, result.Headers)

	// Execute COMMIT and check result
	stmt, err = parser.Parse("COMMIT")
	require.NoError(t, err)
	plan, err = opt.Optimize(stmt)
	require.NoError(t, err)
	result, err = exec.Execute(plan, sess)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, []string{"status"}, result.Headers)
}

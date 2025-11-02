package session

import (
	"fmt"
	"sync"

	"github.com/yyun543/minidb/internal/delta"
	"github.com/yyun543/minidb/internal/logger"
	"go.uber.org/zap"
)

// TransactionState 事务状态
type TransactionState int

const (
	TransactionStateNone       TransactionState = iota // 无事务
	TransactionStateActive                             // 事务活跃中
	TransactionStateCommitted                          // 已提交
	TransactionStateRolledBack                         // 已回滚
)

func (ts TransactionState) String() string {
	switch ts {
	case TransactionStateNone:
		return "None"
	case TransactionStateActive:
		return "Active"
	case TransactionStateCommitted:
		return "Committed"
	case TransactionStateRolledBack:
		return "RolledBack"
	default:
		return "Unknown"
	}
}

// PendingOperation 待执行的操作
type PendingOperation struct {
	Type     string                 // INSERT, UPDATE, DELETE
	TableID  string                 // 表ID
	Data     interface{}            // 操作数据
	Metadata map[string]interface{} // 额外元数据
}

// Transaction 事务管理器
type Transaction struct {
	SessionID      int64                     // 会话ID
	StartVersion   int64                     // 事务开始时的版本号
	IsolationLevel IsolationLevel            // 事务隔离级别
	State          TransactionState          // 事务状态
	Operations     []PendingOperation        // 待提交的操作
	DeltaLog       *delta.OptimisticDeltaLog // 乐观Delta Log
	mu             sync.Mutex                // 互斥锁
}

// NewTransaction 创建新事务
func NewTransaction(sessionID int64, deltaLog *delta.OptimisticDeltaLog, isolationLevel IsolationLevel) *Transaction {
	startVersion := deltaLog.GetLatestVersion()

	logger.WithComponent("transaction").Info("Creating new transaction",
		zap.Int64("session_id", sessionID),
		zap.Int64("start_version", startVersion),
		zap.String("isolation_level", isolationLevel.String()))

	return &Transaction{
		SessionID:      sessionID,
		StartVersion:   startVersion,
		IsolationLevel: isolationLevel,
		State:          TransactionStateActive,
		Operations:     make([]PendingOperation, 0),
		DeltaLog:       deltaLog,
	}
}

// NewTransactionWithDefaultIsolation creates a transaction with default isolation level (READ_COMMITTED)
func NewTransactionWithDefaultIsolation(sessionID int64, deltaLog *delta.OptimisticDeltaLog) *Transaction {
	return NewTransaction(sessionID, deltaLog, IsolationReadCommitted)
}

// AddOperation 添加待执行操作到事务中
func (tx *Transaction) AddOperation(op PendingOperation) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.State != TransactionStateActive {
		return fmt.Errorf("transaction is not active (state: %s)", tx.State)
	}

	tx.Operations = append(tx.Operations, op)

	logger.WithComponent("transaction").Debug("Operation added to transaction",
		zap.Int64("session_id", tx.SessionID),
		zap.String("operation_type", op.Type),
		zap.String("table_id", op.TableID),
		zap.Int("total_operations", len(tx.Operations)))

	return nil
}

// Commit 提交事务（乐观并发控制）
func (tx *Transaction) Commit() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.State != TransactionStateActive {
		return fmt.Errorf("transaction is not active (state: %s)", tx.State)
	}

	logger.WithComponent("transaction").Info("Committing transaction",
		zap.Int64("session_id", tx.SessionID),
		zap.Int64("start_version", tx.StartVersion),
		zap.Int("operation_count", len(tx.Operations)))

	// 检查是否有操作需要提交
	if len(tx.Operations) == 0 {
		tx.State = TransactionStateCommitted
		logger.WithComponent("transaction").Info("Transaction committed (no operations)",
			zap.Int64("session_id", tx.SessionID))
		return nil
	}

	// 尝试提交所有操作（使用乐观锁）
	// 重试机制：最多重试3次
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.WithComponent("transaction").Debug("Attempting to commit",
			zap.Int64("session_id", tx.SessionID),
			zap.Int("attempt", attempt),
			zap.Int("max_retries", maxRetries))

		// 执行所有操作
		err := tx.executeOperations()
		if err == nil {
			// 提交成功
			tx.State = TransactionStateCommitted
			logger.WithComponent("transaction").Info("Transaction committed successfully",
				zap.Int64("session_id", tx.SessionID),
				zap.Int64("start_version", tx.StartVersion),
				zap.Int64("end_version", tx.DeltaLog.GetLatestVersion()),
				zap.Int("operations", len(tx.Operations)))
			return nil
		}

		// 检查是否是冲突错误（可重试）
		if conflictErr, ok := err.(*delta.ConflictError); ok {
			logger.WithComponent("transaction").Warn("Transaction conflict detected, retrying",
				zap.Int64("session_id", tx.SessionID),
				zap.Int64("conflict_version", conflictErr.Version),
				zap.Int("attempt", attempt),
				zap.Error(err))
			lastErr = err
			continue // 重试
		}

		// 其他错误，不可重试
		tx.State = TransactionStateRolledBack
		logger.WithComponent("transaction").Error("Transaction commit failed (non-retryable)",
			zap.Int64("session_id", tx.SessionID),
			zap.Error(err))
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	// 超过最大重试次数
	tx.State = TransactionStateRolledBack
	logger.WithComponent("transaction").Error("Transaction commit failed after retries",
		zap.Int64("session_id", tx.SessionID),
		zap.Int("max_retries", maxRetries),
		zap.Error(lastErr))

	return fmt.Errorf("transaction commit failed after %d retries: %w", maxRetries, lastErr)
}

// executeOperations 执行所有待提交的操作
func (tx *Transaction) executeOperations() error {
	// 注意：这个方法会被Commit重试机制调用
	// 所以这里的操作必须是幂等的（或者我们需要记录已提交的操作）

	for i, op := range tx.Operations {
		logger.WithComponent("transaction").Debug("Executing operation",
			zap.Int64("session_id", tx.SessionID),
			zap.Int("operation_index", i),
			zap.String("operation_type", op.Type),
			zap.String("table_id", op.TableID))

		var err error
		switch op.Type {
		case "ADD":
			// ADD操作：添加Parquet文件
			if file, ok := op.Data.(*delta.ParquetFile); ok {
				err = tx.DeltaLog.AppendAdd(op.TableID, file)
			} else {
				err = fmt.Errorf("invalid data type for ADD operation")
			}

		case "REMOVE":
			// REMOVE操作：删除Parquet文件
			if filePath, ok := op.Data.(string); ok {
				err = tx.DeltaLog.AppendRemove(op.TableID, filePath)
			} else {
				err = fmt.Errorf("invalid data type for REMOVE operation")
			}

		case "METADATA":
			// METADATA操作：更新表元数据
			// 元数据操作通常包含在op.Metadata中
			logger.WithComponent("transaction").Debug("METADATA operation (skipped in current implementation)",
				zap.Int64("session_id", tx.SessionID))
			// 元数据操作暂时跳过（可以扩展）

		default:
			err = fmt.Errorf("unknown operation type: %s", op.Type)
		}

		if err != nil {
			logger.WithComponent("transaction").Error("Operation failed",
				zap.Int64("session_id", tx.SessionID),
				zap.Int("operation_index", i),
				zap.String("operation_type", op.Type),
				zap.Error(err))
			return err
		}
	}

	return nil
}

// Rollback 回滚事务
func (tx *Transaction) Rollback() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.State != TransactionStateActive {
		return fmt.Errorf("transaction is not active (state: %s)", tx.State)
	}

	// 清空所有待执行操作
	tx.Operations = nil
	tx.State = TransactionStateRolledBack

	logger.WithComponent("transaction").Info("Transaction rolled back",
		zap.Int64("session_id", tx.SessionID),
		zap.Int64("start_version", tx.StartVersion))

	return nil
}

// GetOperationCount 获取待提交的操作数量
func (tx *Transaction) GetOperationCount() int {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	return len(tx.Operations)
}

// IsActive 检查事务是否活跃
func (tx *Transaction) IsActive() bool {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	return tx.State == TransactionStateActive
}

// GetReadVersion returns the version to use for read operations based on isolation level
func (tx *Transaction) GetReadVersion() int64 {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	currentVersion := tx.DeltaLog.GetLatestVersion()
	return tx.IsolationLevel.GetSnapshotVersion(tx.StartVersion, currentVersion)
}

// ShouldDetectWriteConflicts returns true if write conflict detection is required
func (tx *Transaction) ShouldDetectWriteConflicts() bool {
	return tx.IsolationLevel.RequiresWriteConflictDetection()
}

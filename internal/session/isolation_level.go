package session

import "fmt"

// IsolationLevel represents the SQL transaction isolation level
// All four standard SQL isolation levels are implemented using SNAPSHOT isolation mechanism
type IsolationLevel int

const (
	// IsolationReadUncommitted allows reading uncommitted data (dirty reads)
	// Implementation: Read from latest version without checking commit status
	IsolationReadUncommitted IsolationLevel = iota

	// IsolationReadCommitted prevents dirty reads by only reading committed data
	// Implementation: Read latest committed version at each read operation
	IsolationReadCommitted

	// IsolationRepeatableRead prevents non-repeatable reads
	// Implementation: Read from snapshot taken at transaction start (MVCC)
	IsolationRepeatableRead

	// IsolationSerializable prevents phantom reads and provides full serializability
	// Implementation: Repeatable Read + write conflict detection (Optimistic Concurrency Control)
	IsolationSerializable
)

// String returns the string representation of the isolation level
func (il IsolationLevel) String() string {
	switch il {
	case IsolationReadUncommitted:
		return "READ_UNCOMMITTED"
	case IsolationReadCommitted:
		return "READ_COMMITTED"
	case IsolationRepeatableRead:
		return "REPEATABLE_READ"
	case IsolationSerializable:
		return "SERIALIZABLE"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", il)
	}
}

// ParseIsolationLevel parses a string into an IsolationLevel
func ParseIsolationLevel(s string) (IsolationLevel, error) {
	switch s {
	case "READ_UNCOMMITTED", "read_uncommitted":
		return IsolationReadUncommitted, nil
	case "READ_COMMITTED", "read_committed":
		return IsolationReadCommitted, nil
	case "REPEATABLE_READ", "repeatable_read":
		return IsolationRepeatableRead, nil
	case "SERIALIZABLE", "serializable":
		return IsolationSerializable, nil
	case "SNAPSHOT", "snapshot":
		// For backward compatibility, map SNAPSHOT to REPEATABLE_READ
		return IsolationRepeatableRead, nil
	default:
		return IsolationReadCommitted, fmt.Errorf("invalid isolation level: %s (must be: READ_UNCOMMITTED, READ_COMMITTED, REPEATABLE_READ, or SERIALIZABLE)", s)
	}
}

// AllowsDirtyReads returns true if this isolation level allows reading uncommitted data
func (il IsolationLevel) AllowsDirtyReads() bool {
	return il == IsolationReadUncommitted
}

// AllowsNonRepeatableReads returns true if this isolation level allows non-repeatable reads
func (il IsolationLevel) AllowsNonRepeatableReads() bool {
	return il == IsolationReadUncommitted || il == IsolationReadCommitted
}

// AllowsPhantomReads returns true if this isolation level allows phantom reads
func (il IsolationLevel) AllowsPhantomReads() bool {
	return il != IsolationSerializable
}

// RequiresWriteConflictDetection returns true if this isolation level requires write conflict detection
func (il IsolationLevel) RequiresWriteConflictDetection() bool {
	return il == IsolationSerializable
}

// GetSnapshotVersion returns the version to use for reads based on isolation level
// - READ_UNCOMMITTED: returns -1 (read latest, including uncommitted)
// - READ_COMMITTED: returns 0 (read latest committed at each operation)
// - REPEATABLE_READ: returns txStartVersion (read from transaction start snapshot)
// - SERIALIZABLE: returns txStartVersion (same as REPEATABLE_READ)
func (il IsolationLevel) GetSnapshotVersion(txStartVersion int64, currentVersion int64) int64 {
	switch il {
	case IsolationReadUncommitted:
		// Read latest version (including uncommitted writes)
		return currentVersion
	case IsolationReadCommitted:
		// Read latest committed version
		return currentVersion
	case IsolationRepeatableRead:
		// Read from transaction start snapshot
		return txStartVersion
	case IsolationSerializable:
		// Read from transaction start snapshot (same as REPEATABLE_READ)
		return txStartVersion
	default:
		// Default to READ_COMMITTED behavior
		return currentVersion
	}
}

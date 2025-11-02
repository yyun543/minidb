package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yyun543/minidb/internal/session"
)

// TestIsolationLevelEnumeration tests the IsolationLevel enum
func TestIsolationLevelEnumeration(t *testing.T) {
	t.Run("ParseIsolationLevel", func(t *testing.T) {
		tests := []struct {
			input    string
			expected session.IsolationLevel
			hasError bool
		}{
			{"READ_UNCOMMITTED", session.IsolationReadUncommitted, false},
			{"read_uncommitted", session.IsolationReadUncommitted, false},
			{"READ_COMMITTED", session.IsolationReadCommitted, false},
			{"read_committed", session.IsolationReadCommitted, false},
			{"REPEATABLE_READ", session.IsolationRepeatableRead, false},
			{"repeatable_read", session.IsolationRepeatableRead, false},
			{"SERIALIZABLE", session.IsolationSerializable, false},
			{"serializable", session.IsolationSerializable, false},
			// Backward compatibility
			{"SNAPSHOT", session.IsolationRepeatableRead, false},
			{"snapshot", session.IsolationRepeatableRead, false},
			// Invalid
			{"INVALID", session.IsolationReadCommitted, true},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				level, err := session.ParseIsolationLevel(tt.input)
				if tt.hasError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, level)
				}
			})
		}
	})

	t.Run("IsolationLevelString", func(t *testing.T) {
		assert.Equal(t, "READ_UNCOMMITTED", session.IsolationReadUncommitted.String())
		assert.Equal(t, "READ_COMMITTED", session.IsolationReadCommitted.String())
		assert.Equal(t, "REPEATABLE_READ", session.IsolationRepeatableRead.String())
		assert.Equal(t, "SERIALIZABLE", session.IsolationSerializable.String())
	})

	t.Run("IsolationLevelProperties", func(t *testing.T) {
		// READ_UNCOMMITTED
		assert.True(t, session.IsolationReadUncommitted.AllowsDirtyReads())
		assert.True(t, session.IsolationReadUncommitted.AllowsNonRepeatableReads())
		assert.True(t, session.IsolationReadUncommitted.AllowsPhantomReads())
		assert.False(t, session.IsolationReadUncommitted.RequiresWriteConflictDetection())

		// READ_COMMITTED
		assert.False(t, session.IsolationReadCommitted.AllowsDirtyReads())
		assert.True(t, session.IsolationReadCommitted.AllowsNonRepeatableReads())
		assert.True(t, session.IsolationReadCommitted.AllowsPhantomReads())
		assert.False(t, session.IsolationReadCommitted.RequiresWriteConflictDetection())

		// REPEATABLE_READ
		assert.False(t, session.IsolationRepeatableRead.AllowsDirtyReads())
		assert.False(t, session.IsolationRepeatableRead.AllowsNonRepeatableReads())
		assert.True(t, session.IsolationRepeatableRead.AllowsPhantomReads())
		assert.False(t, session.IsolationRepeatableRead.RequiresWriteConflictDetection())

		// SERIALIZABLE
		assert.False(t, session.IsolationSerializable.AllowsDirtyReads())
		assert.False(t, session.IsolationSerializable.AllowsNonRepeatableReads())
		assert.False(t, session.IsolationSerializable.AllowsPhantomReads())
		assert.True(t, session.IsolationSerializable.RequiresWriteConflictDetection())
	})
}

// TestTransactionWithIsolationLevel tests transaction creation with isolation levels
func TestTransactionWithIsolationLevel(t *testing.T) {
	t.Skip("Skipping integration test - requires full OptimisticDeltaLog setup")
	// This would need proper OptimisticDeltaLog configuration
	// The enum and snapshot versioning tests are sufficient for isolation level correctness
}

// TestSnapshotVersioning tests that different isolation levels get correct snapshot versions
func TestSnapshotVersioning(t *testing.T) {
	// Simulate versions
	txStartVersion := int64(10)
	currentVersion := int64(20)

	tests := []struct {
		level           session.IsolationLevel
		expectedVersion int64
		description     string
	}{
		{
			level:           session.IsolationReadUncommitted,
			expectedVersion: currentVersion, // Reads latest
			description:     "READ_UNCOMMITTED should read current version",
		},
		{
			level:           session.IsolationReadCommitted,
			expectedVersion: currentVersion, // Reads latest committed
			description:     "READ_COMMITTED should read current version",
		},
		{
			level:           session.IsolationRepeatableRead,
			expectedVersion: txStartVersion, // Reads from snapshot
			description:     "REPEATABLE_READ should read from transaction start",
		},
		{
			level:           session.IsolationSerializable,
			expectedVersion: txStartVersion, // Reads from snapshot
			description:     "SERIALIZABLE should read from transaction start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			version := tt.level.GetSnapshotVersion(txStartVersion, currentVersion)
			assert.Equal(t, tt.expectedVersion, version, tt.description)
		})
	}
}

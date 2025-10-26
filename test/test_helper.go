package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestDataDir returns the base directory for all test data
func TestDataDir() string {
	return "./test_data"
}

// SetupTestDir creates and returns a clean test directory under ./test/test_data/
// The directory will be automatically cleaned up when the test finishes
func SetupTestDir(t *testing.T, name string) string {
	baseDir := TestDataDir()
	testDir := filepath.Join(baseDir, name)

	// Clean up any existing data
	os.RemoveAll(testDir)

	// Create fresh directory
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory %s: %v", testDir, err)
	}

	// Register cleanup
	t.Cleanup(func() {
		os.RemoveAll(testDir)
	})

	return testDir
}

// SetupTestDirWithoutCleanup creates a test directory but doesn't clean it up automatically
// Use this only for tests that need to verify persistence across restarts
func SetupTestDirWithoutCleanup(t *testing.T, name string) string {
	baseDir := TestDataDir()
	testDir := filepath.Join(baseDir, name)

	// Clean up any existing data
	os.RemoveAll(testDir)

	// Create fresh directory
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory %s: %v", testDir, err)
	}

	return testDir
}

// CleanupTestDir manually cleans up a test directory
func CleanupTestDir(testDir string) {
	os.RemoveAll(testDir)
}

// GetTestDataPath returns a path under the test data directory
func GetTestDataPath(name string) string {
	return filepath.Join(TestDataDir(), name)
}

// EnsureTestDataDirExists ensures the base test_data directory exists
func EnsureTestDataDirExists() error {
	return os.MkdirAll(TestDataDir(), 0755)
}

// CleanAllTestData removes all test data (use with caution!)
func CleanAllTestData() error {
	return os.RemoveAll(TestDataDir())
}

// CreateTempTestDir creates a temporary test directory with a unique name
func CreateTempTestDir(t *testing.T, prefix string) string {
	baseDir := TestDataDir()
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatalf("Failed to create base test directory: %v", err)
	}

	testDir, err := os.MkdirTemp(baseDir, prefix+"_*")
	if err != nil {
		t.Fatalf("Failed to create temp test directory: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		os.RemoveAll(testDir)
	})

	return testDir
}

// InitTestEnvironment should be called at the beginning of test suite
func InitTestEnvironment() error {
	// Ensure test_data directory exists
	if err := EnsureTestDataDirExists(); err != nil {
		return fmt.Errorf("failed to initialize test environment: %w", err)
	}
	return nil
}

package test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
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

// ExpectedResult 定义期望的查询结果
type ExpectedResult struct {
	Headers  []string        // 表头名称
	Types    []string        // 列类型 (INTEGER, DOUBLE, VARCHAR, BOOLEAN, TIMESTAMP)
	Rows     [][]interface{} // 期望的数据行
	RowCount int             // 期望的行数 (-1 表示不检查)
}

// VerifyQueryResult 综合验证查询结果
func VerifyQueryResult(t *testing.T, record arrow.Record, expected ExpectedResult) {
	t.Helper()

	// 1. 验证表头
	if len(expected.Headers) > 0 {
		actualHeaders := make([]string, record.NumCols())
		for i := 0; i < int(record.NumCols()); i++ {
			actualHeaders[i] = record.ColumnName(i)
		}
		if !reflect.DeepEqual(actualHeaders, expected.Headers) {
			t.Errorf("Headers mismatch:\nExpected: %v\nActual:   %v", expected.Headers, actualHeaders)
		}
	}

	// 2. 验证列类型
	if len(expected.Types) > 0 {
		for i, expectedType := range expected.Types {
			actualType := record.Column(i).DataType()
			if !matchesType(actualType, expectedType) {
				t.Errorf("Column %d (%s) type mismatch:\nExpected: %s\nActual:   %s",
					i, record.ColumnName(i), expectedType, actualType)
			}
		}
	}

	// 3. 验证行数
	if expected.RowCount >= 0 {
		actualRowCount := int(record.NumRows())
		if actualRowCount != expected.RowCount {
			t.Errorf("Row count mismatch:\nExpected: %d\nActual:   %d", expected.RowCount, actualRowCount)
		}
	}

	// 4. 验证具体的数据值
	if len(expected.Rows) > 0 {
		for rowIdx, expectedRow := range expected.Rows {
			if rowIdx >= int(record.NumRows()) {
				t.Errorf("Expected row %d but record only has %d rows", rowIdx, record.NumRows())
				continue
			}
			for colIdx, expectedValue := range expectedRow {
				if colIdx >= int(record.NumCols()) {
					t.Errorf("Expected column %d but record only has %d columns", colIdx, record.NumCols())
					continue
				}
				actualValue := getColumnValue(record.Column(colIdx), rowIdx)
				if !valuesEqual(expectedValue, actualValue) {
					t.Errorf("Value mismatch at row %d, col %d (%s):\nExpected: %v (%T)\nActual:   %v (%T)",
						rowIdx, colIdx, record.ColumnName(colIdx), expectedValue, expectedValue, actualValue, actualValue)
				}
			}
		}
	}
}

// matchesType 检查Arrow类型是否匹配SQL类型
func matchesType(arrowType arrow.DataType, sqlType string) bool {
	switch sqlType {
	case "INTEGER", "INT":
		return arrowType.ID() == arrow.INT64 || arrowType.ID() == arrow.INT32
	case "DOUBLE", "FLOAT":
		return arrowType.ID() == arrow.FLOAT64 || arrowType.ID() == arrow.FLOAT32
	case "VARCHAR", "STRING", "TEXT":
		return arrowType.ID() == arrow.STRING
	case "BOOLEAN", "BOOL":
		return arrowType.ID() == arrow.BOOL
	case "TIMESTAMP":
		return arrowType.ID() == arrow.TIMESTAMP
	default:
		return false
	}
}

// getColumnValue 从Arrow列中提取指定行的值
func getColumnValue(column arrow.Array, rowIdx int) interface{} {
	if column.IsNull(rowIdx) {
		return nil
	}

	switch col := column.(type) {
	case *array.Int64:
		return col.Value(rowIdx)
	case *array.Int32:
		return int64(col.Value(rowIdx)) // 统一转换为int64以便比较
	case *array.Float64:
		return col.Value(rowIdx)
	case *array.Float32:
		return float64(col.Value(rowIdx)) // 统一转换为float64以便比较
	case *array.String:
		return col.Value(rowIdx)
	case *array.Boolean:
		return col.Value(rowIdx)
	default:
		return fmt.Sprintf("unsupported type: %T", col)
	}
}

// valuesEqual 比较两个值是否相等(处理类型转换)
func valuesEqual(expected, actual interface{}) bool {
	if expected == nil {
		return actual == nil
	}
	if actual == nil {
		return false
	}

	// 处理数值类型的比较
	switch e := expected.(type) {
	case int:
		if a, ok := actual.(int64); ok {
			return int64(e) == a
		}
	case int64:
		if a, ok := actual.(int64); ok {
			return e == a
		}
	case float64:
		if a, ok := actual.(float64); ok {
			// 浮点数比较使用小精度
			return abs(e-a) < 1e-9
		}
	case float32:
		if a, ok := actual.(float64); ok {
			return abs(float64(e)-a) < 1e-6
		}
	case string:
		if a, ok := actual.(string); ok {
			return e == a
		}
	case bool:
		if a, ok := actual.(bool); ok {
			return e == a
		}
	}

	// 默认使用reflect.DeepEqual
	return reflect.DeepEqual(expected, actual)
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

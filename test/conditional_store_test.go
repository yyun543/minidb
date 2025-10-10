package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yyun543/minidb/internal/objectstore"
)

// TestConditionalObjectStore 测试条件写入功能
func TestConditionalObjectStore(t *testing.T) {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "minidb-test-conditional", time.Now().Format("20060102150405"))
	defer os.RemoveAll(tempDir)

	// 创建 LocalStore
	store, err := objectstore.NewLocalStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create LocalStore: %v", err)
	}

	// LocalStore 直接实现了 ConditionalObjectStore 接口的所有方法
	conditionalStore := store

	t.Run("PutIfNotExists - Success", func(t *testing.T) {
		path := "test/new_file.txt"
		data := []byte("test content")

		// 文件不存在，应该成功写入
		err := conditionalStore.PutIfNotExists(path, data)
		if err != nil {
			t.Fatalf("PutIfNotExists should succeed for non-existent file: %v", err)
		}

		// 验证文件已创建
		exists, err := store.Exists(path)
		if err != nil {
			t.Fatalf("Failed to check file existence: %v", err)
		}
		if !exists {
			t.Fatal("File should exist after PutIfNotExists")
		}

		// 验证文件内容
		readData, err := store.Get(path)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(readData) != string(data) {
			t.Fatalf("File content mismatch: expected %q, got %q", string(data), string(readData))
		}
	})

	t.Run("PutIfNotExists - Failure (File Exists)", func(t *testing.T) {
		path := "test/existing_file.txt"
		data1 := []byte("original content")
		data2 := []byte("new content")

		// 先创建文件
		err := store.Put(path, data1)
		if err != nil {
			t.Fatalf("Failed to create initial file: %v", err)
		}

		// 尝试使用 PutIfNotExists，应该失败
		err = conditionalStore.PutIfNotExists(path, data2)
		if err == nil {
			t.Fatal("PutIfNotExists should fail for existing file")
		}

		// 验证错误类型
		expectedError := "PreconditionFailed"
		if !containsString(err.Error(), expectedError) {
			t.Fatalf("Expected error containing %q, got: %v", expectedError, err)
		}

		// 验证原文件内容未改变
		readData, err := store.Get(path)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(readData) != string(data1) {
			t.Fatalf("File content should not change: expected %q, got %q", string(data1), string(readData))
		}
	})

	t.Run("PutIfMatch - Success", func(t *testing.T) {
		path := "test/etag_file.txt"
		data1 := []byte("original content")
		data2 := []byte("updated content")

		// 创建初始文件
		err := store.Put(path, data1)
		if err != nil {
			t.Fatalf("Failed to create initial file: %v", err)
		}

		// 获取当前 ETag
		info, err := store.Stat(path)
		if err != nil {
			t.Fatalf("Failed to get file info: %v", err)
		}
		currentETag := info.ETag

		// 使用正确的 ETag 更新文件
		err = conditionalStore.PutIfMatch(path, data2, currentETag)
		if err != nil {
			t.Fatalf("PutIfMatch should succeed with correct ETag: %v", err)
		}

		// 验证文件已更新
		readData, err := store.Get(path)
		if err != nil {
			t.Fatalf("Failed to read updated file: %v", err)
		}
		if string(readData) != string(data2) {
			t.Fatalf("File content should be updated: expected %q, got %q", string(data2), string(readData))
		}
	})

	t.Run("PutIfMatch - Failure (Wrong ETag)", func(t *testing.T) {
		path := "test/etag_mismatch_file.txt"
		data1 := []byte("original content")
		data2 := []byte("should not be written")

		// 创建初始文件
		err := store.Put(path, data1)
		if err != nil {
			t.Fatalf("Failed to create initial file: %v", err)
		}

		// 使用错误的 ETag 尝试更新
		wrongETag := "wrong-etag-12345"
		err = conditionalStore.PutIfMatch(path, data2, wrongETag)
		if err == nil {
			t.Fatal("PutIfMatch should fail with wrong ETag")
		}

		// 验证错误类型
		expectedError := "PreconditionFailed"
		if !containsString(err.Error(), expectedError) {
			t.Fatalf("Expected error containing %q, got: %v", expectedError, err)
		}

		// 验证原文件内容未改变
		readData, err := store.Get(path)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(readData) != string(data1) {
			t.Fatalf("File content should not change: expected %q, got %q", string(data1), string(readData))
		}
	})

	t.Run("PutIfMatch - File Not Exists", func(t *testing.T) {
		path := "test/non_existent_for_etag.txt"
		data := []byte("new content")

		// 使用空 ETag 创建新文件
		err := conditionalStore.PutIfMatch(path, data, "")
		if err != nil {
			t.Fatalf("PutIfMatch should succeed for non-existent file with empty ETag: %v", err)
		}

		// 验证文件已创建
		exists, err := store.Exists(path)
		if err != nil {
			t.Fatalf("Failed to check file existence: %v", err)
		}
		if !exists {
			t.Fatal("File should exist after PutIfMatch with empty ETag")
		}
	})

	t.Run("Concurrent PutIfNotExists", func(t *testing.T) {
		path := "test/concurrent_file.txt"
		data1 := []byte("content from goroutine 1")
		data2 := []byte("content from goroutine 2")

		// 模拟并发写入
		results := make(chan error, 2)

		go func() {
			results <- conditionalStore.PutIfNotExists(path, data1)
		}()

		go func() {
			results <- conditionalStore.PutIfNotExists(path, data2)
		}()

		// 收集结果
		var errors []error
		for i := 0; i < 2; i++ {
			if err := <-results; err != nil {
				errors = append(errors, err)
			}
		}

		// 应该有一个成功，一个失败
		if len(errors) != 1 {
			t.Fatalf("Expected exactly one error from concurrent writes, got %d errors: %v", len(errors), errors)
		}

		// 验证文件存在
		exists, err := store.Exists(path)
		if err != nil {
			t.Fatalf("Failed to check file existence: %v", err)
		}
		if !exists {
			t.Fatal("File should exist after concurrent PutIfNotExists")
		}

		// 验证文件内容是其中之一
		readData, err := store.Get(path)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		content := string(readData)
		if content != string(data1) && content != string(data2) {
			t.Fatalf("File content should be one of the concurrent writes: got %q", content)
		}
	})
}

// containsString 检查字符串是否包含子字符串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package objectstore

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ObjectInfo 对象元数据
type ObjectInfo struct {
	Path         string
	Size         int64
	ModifiedTime int64
	ETag         string
}

// LocalStore 本地文件系统对象存储实现
type LocalStore struct {
	basePath string
}

// NewLocalStore 创建本地对象存储
func NewLocalStore(basePath string) (*LocalStore, error) {
	// 确保基础路径存在
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}

	return &LocalStore{
		basePath: basePath,
	}, nil
}

// Get 获取对象内容
func (ls *LocalStore) Get(path string) ([]byte, error) {
	fullPath := ls.getFullPath(path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read object: %w", err)
	}
	return data, nil
}

// Put 写入对象
func (ls *LocalStore) Put(path string, data []byte) error {
	fullPath := ls.getFullPath(path)

	// 确保父目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write object: %w", err)
	}

	return nil
}

// Delete 删除对象
func (ls *LocalStore) Delete(path string) error {
	fullPath := ls.getFullPath(path)
	if err := os.Remove(fullPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete object: %w", err)
		}
	}
	return nil
}

// List 列出指定前缀的所有对象
func (ls *LocalStore) List(prefix string) ([]string, error) {
	fullPrefix := ls.getFullPath(prefix)
	var results []string

	// 如果前缀路径不存在，返回空列表
	if _, err := os.Stat(fullPrefix); os.IsNotExist(err) {
		return results, nil
	}

	err := filepath.Walk(fullPrefix, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 转换为相对路径
		relPath, err := filepath.Rel(ls.basePath, path)
		if err != nil {
			return err
		}

		// 统一路径分隔符为 /
		relPath = filepath.ToSlash(relPath)
		results = append(results, relPath)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	return results, nil
}

// GetReader 获取对象读取器
func (ls *LocalStore) GetReader(path string) (io.ReadCloser, error) {
	fullPath := ls.getFullPath(path)
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object not found: %s", path)
		}
		return nil, fmt.Errorf("failed to open object: %w", err)
	}
	return file, nil
}

// GetWriter 获取对象写入器
func (ls *LocalStore) GetWriter(path string) (io.WriteCloser, error) {
	fullPath := ls.getFullPath(path)

	// 确保父目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create object: %w", err)
	}

	return file, nil
}

// Exists 检查对象是否存在
func (ls *LocalStore) Exists(path string) (bool, error) {
	fullPath := ls.getFullPath(path)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Stat 获取对象元数据
func (ls *LocalStore) Stat(path string) (*ObjectInfo, error) {
	fullPath := ls.getFullPath(path)
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object not found: %s", path)
		}
		return nil, fmt.Errorf("failed to stat object: %w", err)
	}

	return &ObjectInfo{
		Path:         path,
		Size:         info.Size(),
		ModifiedTime: info.ModTime().Unix(),
		ETag:         fmt.Sprintf("%d-%d", info.Size(), info.ModTime().Unix()),
	}, nil
}

// getFullPath 获取完整文件路径
func (ls *LocalStore) getFullPath(path string) string {
	// 移除路径开头的斜杠
	path = strings.TrimPrefix(path, "/")
	return filepath.Join(ls.basePath, path)
}

// Close 关闭对象存储 (本地文件系统无需关闭)
func (ls *LocalStore) Close() error {
	return nil
}

// ensureDir 确保目录存在
func ensureDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}

// fsync 刷盘确保持久化
func fsyncFile(file *os.File) error {
	if err := file.Sync(); err != nil {
		return fmt.Errorf("fsync failed: %w", err)
	}
	return nil
}

// atomicWrite 原子写入文件
func atomicWrite(path string, data []byte) error {
	// 写入临时文件
	tmpPath := path + ".tmp." + fmt.Sprintf("%d", time.Now().UnixNano())

	if err := ensureDir(tmpPath); err != nil {
		return err
	}

	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpPath) // 清理临时文件

	if _, err := file.Write(data); err != nil {
		file.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// fsync 确保数据落盘
	if err := fsyncFile(file); err != nil {
		file.Close()
		return err
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// 原子重命名
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

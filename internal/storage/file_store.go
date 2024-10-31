package storage

import (
	"os"
	"sync"
)

// FileStore 处理文件存储
type FileStore struct {
	path string
	file *os.File
	mu   sync.RWMutex
}

// NewFileStore 创建新的文件存储
func NewFileStore(path string) (*FileStore, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &FileStore{
		path: path,
		file: file,
	}, nil
}

// Put 写入数据
func (fs *FileStore) Put(key []byte, value []byte) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// 简单实现，实际应该使用更复杂的存储格式
	_, err := fs.file.Write(append(key, value...))
	return err
}

package delta

import (
	"time"

	"github.com/apache/arrow/go/v18/arrow"
)

// Operation 操作类型
type Operation string

const (
	OpAdd      Operation = "ADD"
	OpRemove   Operation = "REMOVE"
	OpMetadata Operation = "METADATA"
)

// LogEntry Delta Log 条目
type LogEntry struct {
	Version   int64     `json:"version"`
	Timestamp int64     `json:"timestamp"`
	TableID   string    `json:"table_id"`
	Operation Operation `json:"operation"`

	// ADD 操作字段
	FilePath   string                 `json:"file_path,omitempty"`
	FileSize   int64                  `json:"file_size,omitempty"`
	RowCount   int64                  `json:"row_count,omitempty"`
	MinValues  map[string]interface{} `json:"min_values,omitempty"`
	MaxValues  map[string]interface{} `json:"max_values,omitempty"`
	NullCounts map[string]int64       `json:"null_counts,omitempty"`
	DataChange bool                   `json:"data_change,omitempty"`

	// REMOVE 操作字段
	DeletionTimestamp int64 `json:"deletion_timestamp,omitempty"`

	// METADATA 操作字段
	SchemaJSON string `json:"schema_json,omitempty"`

	// 审计字段
	UserID    string `json:"user_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	QueryID   string `json:"query_id,omitempty"`
}

// Snapshot 表快照
type Snapshot struct {
	Version   int64
	Timestamp time.Time
	TableID   string
	Files     []FileInfo
	Schema    *arrow.Schema
}

// FileInfo Parquet 文件信息
type FileInfo struct {
	Path       string
	Size       int64
	RowCount   int64
	MinValues  map[string]interface{}
	MaxValues  map[string]interface{}
	NullCounts map[string]int64
	AddedAt    int64
}

// ParquetFile Parquet 文件描述
type ParquetFile struct {
	Path     string
	Size     int64
	RowCount int64
	Stats    *FileStats
}

// FileStats 文件统计信息
type FileStats struct {
	RowCount   int64
	FileSize   int64
	MinValues  map[string]interface{}
	MaxValues  map[string]interface{}
	NullCounts map[string]int64
}

// Checkpoint 检查点
type Checkpoint struct {
	TableID   string
	Version   int64
	Timestamp int64
	FileCount int
	TotalSize int64
	Snapshot  *Snapshot
}

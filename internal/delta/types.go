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

	// Merge-on-Read 字段
	IsDelta   bool   `json:"is_delta,omitempty"`   // 是否为 Delta 文件
	DeltaType string `json:"delta_type,omitempty"` // Delta 文件类型: "update", "delete", "insert"

	// METADATA 操作字段
	SchemaJSON     string `json:"schema_json,omitempty"`
	IndexJSON      string `json:"index_json,omitempty"`      // 索引元数据
	IndexOperation string `json:"index_operation,omitempty"` // 索引操作类型: "CREATE", "DROP"

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
	IsDelta    bool   // Merge-on-Read: 是否为 Delta 文件
	DeltaType  string // Delta 文件类型: "update", "delete", "insert"
}

// ParquetFile Parquet 文件描述
type ParquetFile struct {
	Path      string
	Size      int64
	RowCount  int64
	Stats     *FileStats
	IsDelta   bool   // Merge-on-Read: 是否为 Delta 文件
	DeltaType string // Delta 文件类型: "update", "delete", "insert"
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

// LogInterface 定义 Delta Log 的通用接口
// 这个接口被 storage 和 optimizer 包使用，避免循环依赖
type LogInterface interface {
	// Bootstrap 初始化 Delta Log
	Bootstrap() error
	// RestoreFromEntries 从已加载的 entries 恢复状态
	RestoreFromEntries(entries []LogEntry) error
	// AppendAdd 追加 ADD 操作
	AppendAdd(tableID string, file *ParquetFile) error
	// AppendRemove 追加 REMOVE 操作
	AppendRemove(tableID, filePath string) error
	// AppendMetadata 追加 METADATA 操作
	AppendMetadata(tableID string, schema *arrow.Schema) error
	// AppendIndexMetadata 追加索引元数据操作
	AppendIndexMetadata(tableID, indexName string, indexMeta map[string]interface{}) error
	// RemoveIndexMetadata 删除索引元数据操作
	RemoveIndexMetadata(tableID, indexName string) error
	// GetSnapshot 获取表快照
	GetSnapshot(tableID string, version int64) (*Snapshot, error)
	// GetLatestVersion 获取最新版本号
	GetLatestVersion() int64
	// GetVersionByTimestamp 根据时间戳查找版本号
	GetVersionByTimestamp(tableID string, ts int64) (int64, error)
	// ListTables 列出所有表
	ListTables() []string
	// GetAllEntries 获取所有日志条目
	GetAllEntries() []LogEntry
	// GetEntriesByTable 获取指定表的日志条目
	GetEntriesByTable(tableID string) []LogEntry
}

package storage

import (
	"context"
	"io"

	"github.com/apache/arrow/go/v18/arrow"
)

// StorageEngine 统一存储引擎接口
type StorageEngine interface {
	// 生命周期管理
	Open() error
	Close() error

	// 数据库操作
	CreateDatabase(name string) error
	DropDatabase(name string) error
	ListDatabases() ([]string, error)
	DatabaseExists(name string) (bool, error)

	// 表操作
	CreateTable(db, table string, schema *arrow.Schema) error
	DropTable(db, table string) error
	GetTableSchema(db, table string) (*arrow.Schema, error)
	ListTables(db string) ([]string, error)
	TableExists(db, table string) (bool, error)

	// 数据读写
	Scan(ctx context.Context, db, table string, filters []Filter) (RecordIterator, error)
	Write(ctx context.Context, db, table string, batch arrow.Record) error
	Update(ctx context.Context, db, table string, filters []Filter, updates map[string]interface{}) (int64, error)
	Delete(ctx context.Context, db, table string, filters []Filter) (int64, error)

	// 事务支持
	BeginTransaction() (Transaction, error)

	// 统计信息
	GetTableStats(db, table string) (*TableStats, error)

	// 时间旅行查询
	ScanVersion(ctx context.Context, db, table string, version int64, filters []Filter) (RecordIterator, error)
}

// RecordIterator Arrow 记录迭代器
type RecordIterator interface {
	Next() bool
	Record() arrow.Record
	Err() error
	Close() error
}

// Transaction ACID 事务
type Transaction interface {
	GetVersion() int64
	GetID() string
	Commit() error
	Rollback() error
}

// Filter 查询过滤条件
type Filter struct {
	Column   string        // 列名
	Operator string        // =, >, <, >=, <=, !=, LIKE, IN, BETWEEN
	Value    interface{}   // 单值操作符使用
	Values   []interface{} // IN 操作符使用多个值
}

// TableStats 表统计信息
type TableStats struct {
	TableName    string
	RowCount     int64
	FileCount    int
	TotalSizeGB  float64
	MinValues    map[string]interface{}
	MaxValues    map[string]interface{}
	LastModified int64
}

// ObjectStore 对象存储抽象接口
type ObjectStore interface {
	// 文件操作
	Get(path string) ([]byte, error)
	Put(path string, data []byte) error
	Delete(path string) error
	List(prefix string) ([]string, error)

	// 流式操作
	GetReader(path string) (io.ReadCloser, error)
	GetWriter(path string) (io.WriteCloser, error)

	// 元数据
	Exists(path string) (bool, error)
}

package catalog

import (
	"encoding/json"
	"fmt"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/storage"
)

// MetadataManager 封装了 Catalog 元数据的读写操作（例如数据库、表信息），
// 其实现基于存储引擎以及 Apache Arrow 记录格式。
type MetadataManager struct {
	engine storage.Engine
	km     *storage.KeyManager
}

// NewMetadataManager 创建 MetadataManager 实例。
func NewMetadataManager(engine storage.Engine) *MetadataManager {
	return &MetadataManager{
		engine: engine,
		km:     storage.NewKeyManager(),
	}
}

// DatabaseMeta 表示数据库元数据（目前仅包含名称，后续可扩展）。
type DatabaseMeta struct {
	Name string `json:"name"`
}

// TableMeta 表示表元数据，其中 Schema 字段存储 Arrow Schema 的 JSON 表示。
type TableMeta struct {
	Database string `json:"database"`
	Table    string `json:"table"`
	Schema   string `json:"schema"`
}

// CreateDatabase 构造一条 Arrow 记录写入存储引擎，实现“数据库创建”。
//
// 注：这里采用单行单列的 Arrow 记录，列名固定为 "name"。
func (m *MetadataManager) CreateDatabase(name string) error {
	pool := memory.NewGoAllocator()
	// 定义 schema：单个 string 字段 "name"
	field := arrow.Field{Name: "name", Type: arrow.BinaryTypes.String}
	schema := arrow.NewSchema([]arrow.Field{field}, nil)

	builder := array.NewStringBuilder(pool)
	builder.Append(name)
	arr := builder.NewArray()
	builder.Release()

	record := array.NewRecord(schema, []arrow.Array{arr}, 1)

	key := m.km.DatabaseKey(name)
	err := m.engine.Put(key, &record)
	if err != nil {
		record.Release()
		return fmt.Errorf("failed to create database meta: %w", err)
	}
	// 所有权已转移，不需要调用 Release() 由引擎统一管理
	return nil
}

// GetDatabase 通过存储引擎中 key 获取数据库的元数据。
func (m *MetadataManager) GetDatabase(name string) (DatabaseMeta, error) {
	key := m.km.DatabaseKey(name)
	record, err := m.engine.Get(key)
	if err != nil || record == nil {
		return DatabaseMeta{}, fmt.Errorf("database not found")
	}
	defer record.Release()

	// 假设 record 中第一列存储 "name"
	col := record.Column(0)
	stringArr := col.(*array.String)
	if stringArr.Len() < 1 {
		return DatabaseMeta{}, fmt.Errorf("empty database metadata")
	}
	dbName := stringArr.Value(0)
	return DatabaseMeta{Name: dbName}, nil
}

// CreateTable 构造一条表元数据 Arrow 记录写入存储引擎，实现“建表”操作。
//
// 此处将表元数据（包括所属数据库、表名、以及 Schema 信息）序列化为 JSON 存入单列 "meta" 中。
func (m *MetadataManager) CreateTable(dbName string, table TableMeta) error {
	// 将 TableMeta 序列化为 JSON
	b, err := json.Marshal(table)
	if err != nil {
		return fmt.Errorf("failed to marshal table metadata: %w", err)
	}

	pool := memory.NewGoAllocator()
	field := arrow.Field{Name: "meta", Type: arrow.BinaryTypes.String}
	schema := arrow.NewSchema([]arrow.Field{field}, nil)

	builder := array.NewStringBuilder(pool)
	builder.Append(string(b))
	arr := builder.NewArray()
	builder.Release()

	record := array.NewRecord(schema, []arrow.Array{arr}, 1)

	key := m.km.TableKey(dbName, table.Table)
	err = m.engine.Put(key, &record)
	if err != nil {
		record.Release()
		return fmt.Errorf("failed to create table meta: %w", err)
	}
	return nil
}

// GetTable 通过存储引擎获取表元数据，并反序列化成 TableMeta 结构体。
func (m *MetadataManager) GetTable(dbName, tableName string) (TableMeta, error) {
	key := m.km.TableKey(dbName, tableName)
	record, err := m.engine.Get(key)
	if err != nil || record == nil {
		return TableMeta{}, fmt.Errorf("table not found")
	}
	defer record.Release()

	col := record.Column(0)
	stringArr := col.(*array.String)
	if stringArr.Len() < 1 {
		return TableMeta{}, fmt.Errorf("empty table metadata")
	}
	metaJSON := stringArr.Value(0)
	var tableMeta TableMeta
	err = json.Unmarshal([]byte(metaJSON), &tableMeta)
	if err != nil {
		return TableMeta{}, fmt.Errorf("failed to unmarshal table metadata: %w", err)
	}
	return tableMeta, nil
}

// GetAllDatabases 获取所有数据库的元数据。
func (m *MetadataManager) GetAllDatabases() ([]DatabaseMeta, error) {
	// 使用 KeyManager 构造扫描范围
	prefix := m.km.DatabaseKey("")
	startKey, endKey := m.km.GetKeyRange(prefix)

	// 创建迭代器
	it, err := m.engine.Scan(startKey, endKey)
	if err != nil {
		return nil, fmt.Errorf("failed to scan databases: %w", err)
	}
	defer it.Close()

	var databases []DatabaseMeta
	// 遍历所有数据库记录
	for it.Next() {
		record := it.Record()
		// 确保在处理完后释放记录
		defer record.Release()

		if record.NumCols() < 1 {
			continue
		}

		// 获取数据库名称
		col := record.Column(0)
		stringArr := col.(*array.String)
		for i := 0; i < stringArr.Len(); i++ {
			databases = append(databases, DatabaseMeta{
				Name: stringArr.Value(i),
			})
		}
	}

	return databases, nil
}

// DeleteDatabase 删除指定数据库的元数据。
func (m *MetadataManager) DeleteDatabase(name string) error {
	key := m.km.DatabaseKey(name)
	return m.engine.Delete(key)
}

// DeleteTable 删除指定表的元数据。
func (m *MetadataManager) DeleteTable(dbName, tableName string) error {
	key := m.km.TableKey(dbName, tableName)
	return m.engine.Delete(key)
}

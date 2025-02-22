package catalog

import (
	"fmt"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/storage"
	"github.com/yyun543/minidb/internal/types"
)

// MetadataManager 负责管理数据库元数据，包括数据库、表等信息。
type MetadataManager struct {
	engine storage.Engine
	km     *storage.KeyManager
}

// NewMetadataManager 创建一个新的元数据管理器实例。
func NewMetadataManager(engine storage.Engine) *MetadataManager {
	return &MetadataManager{
		engine: engine,
		km:     storage.NewKeyManager(),
	}
}

// DatabaseMeta 表示数据库元数据。
type DatabaseMeta struct {
	Name string
}

// TableMeta 表示表元数据。
type TableMeta struct {
	Database   string
	Table      string
	ChunkCount int64
	Schema     *arrow.Schema
}

// CreateDatabase 创建数据库元数据记录。
func (m *MetadataManager) CreateDatabase(name string) error {
	// 检查数据库是否已存在
	key := m.km.TableChunkKey(storage.SYS_DATABASE, storage.SYS_DATABASES, 0)
	existingRecord, err := m.engine.Get(key)
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	// 获取当前最大ID
	maxID := int64(0)
	if existingRecord != nil {
		defer existingRecord.Release()
		if existingRecord.NumRows() > 0 {
			// 检查数据库是否已存在
			nameCol := existingRecord.Column(1).(*array.String)
			for i := int64(0); i < existingRecord.NumRows(); i++ {
				if nameCol.Value(int(i)) == name {
					return fmt.Errorf("database %s already exists", name)
				}
			}
			// 获取最大ID
			idCol := existingRecord.Column(0).(*array.Int64)
			for i := int64(0); i < existingRecord.NumRows(); i++ {
				if idCol.Value(int(i)) > maxID {
					maxID = idCol.Value(int(i))
				}
			}
		}
	}

	// 创建新的数据库记录
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, SysDatabasesSchema)
	defer builder.Release()

	// 保留现有记录
	if existingRecord != nil && existingRecord.NumRows() > 0 {
		// 复制现有的ID
		idCol := existingRecord.Column(0).(*array.Int64)
		idBuilder := builder.Field(0).(*array.Int64Builder)
		for i := int64(0); i < existingRecord.NumRows(); i++ {
			idBuilder.Append(idCol.Value(int(i)))
		}
		// 复制现有的名称
		nameCol := existingRecord.Column(1).(*array.String)
		nameBuilder := builder.Field(1).(*array.StringBuilder)
		for i := int64(0); i < existingRecord.NumRows(); i++ {
			nameBuilder.Append(nameCol.Value(int(i)))
		}
	}

	// 添加新数据库记录
	builder.Field(0).(*array.Int64Builder).Append(maxID + 1)
	builder.Field(1).(*array.StringBuilder).Append(name)

	// 创建Record
	record := builder.NewRecord()
	defer record.Release()

	// 写入存储
	err = m.engine.Put(key, &record)
	if err != nil {
		return fmt.Errorf("failed to create database meta: %w", err)
	}

	return nil
}

// GetDatabase 获取数据库元数据。
func (m *MetadataManager) GetDatabase(name string) (DatabaseMeta, error) {
	key := m.km.TableChunkKey(storage.SYS_DATABASE, storage.SYS_DATABASES, 0)
	record, err := m.engine.Get(key)
	if err != nil {
		return DatabaseMeta{}, fmt.Errorf("failed to get database meta: %w", err)
	}
	if record == nil {
		return DatabaseMeta{}, fmt.Errorf("database not found")
	}
	defer record.Release()

	// 遍历记录查找匹配的数据库
	for i := int64(0); i < record.NumRows(); i++ {
		nameCol := record.Column(1).(*array.String)
		if nameCol.Value(int(i)) == name {
			return DatabaseMeta{
				Name: name,
			}, nil
		}
	}

	return DatabaseMeta{}, fmt.Errorf("database %s not found", name)
}

// GetAllDatabases 获取所有数据库的元数据。
func (m *MetadataManager) GetAllDatabases() ([]DatabaseMeta, error) {
	key := m.km.TableChunkKey(storage.SYS_DATABASE, storage.SYS_DATABASES, 0)
	record, err := m.engine.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get databases: %w", err)
	}
	if record == nil {
		return []DatabaseMeta{}, nil
	}
	defer record.Release()

	var databases []DatabaseMeta
	nameCol := record.Column(1).(*array.String)
	for i := int64(0); i < record.NumRows(); i++ {
		databases = append(databases, DatabaseMeta{Name: nameCol.Value(int(i))})
	}

	return databases, nil
}

// CreateTable 创建表元数据记录。
func (m *MetadataManager) CreateTable(dbName string, table TableMeta) error {
	// 检查数据库是否存在
	_, err := m.GetDatabase(dbName)
	if err != nil {
		return fmt.Errorf("database %s not found", dbName)
	}

	// 获取现有表记录
	key := m.km.TableChunkKey(storage.SYS_DATABASE, storage.SYS_TABLES, 0)
	existingRecord, err := m.engine.Get(key)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	// 获取当前最大ID
	maxID := int64(0)
	if existingRecord != nil {
		defer existingRecord.Release()
		if existingRecord.NumRows() > 0 {
			// 检查表是否已存在
			dbNameCol := existingRecord.Column(2).(*array.String)
			tableNameCol := existingRecord.Column(3).(*array.String)
			for i := int64(0); i < existingRecord.NumRows(); i++ {
				if dbNameCol.Value(int(i)) == dbName && tableNameCol.Value(int(i)) == table.Table {
					return fmt.Errorf("table %s already exists in database %s", table.Table, dbName)
				}
			}
			// 获取最大ID
			idCol := existingRecord.Column(0).(*array.Int64)
			for i := int64(0); i < existingRecord.NumRows(); i++ {
				if idCol.Value(int(i)) > maxID {
					maxID = idCol.Value(int(i))
				}
			}
		}
	}

	// 创建新的表记录
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, SysTablesSchema)
	defer builder.Release()

	// 保留现有记录
	if existingRecord != nil && existingRecord.NumRows() > 0 {
		// 复制现有的ID
		idCol := existingRecord.Column(0).(*array.Int64)
		idBuilder := builder.Field(0).(*array.Int64Builder)
		for i := int64(0); i < existingRecord.NumRows(); i++ {
			idBuilder.Append(idCol.Value(int(i)))
		}
		// 复制现有的数据库ID
		dbIDCol := existingRecord.Column(1).(*array.Int64)
		dbIDBuilder := builder.Field(1).(*array.Int64Builder)
		for i := int64(0); i < existingRecord.NumRows(); i++ {
			dbIDBuilder.Append(dbIDCol.Value(int(i)))
		}
		// 复制现有的数据库名
		dbNameCol := existingRecord.Column(2).(*array.String)
		dbNameBuilder := builder.Field(2).(*array.StringBuilder)
		for i := int64(0); i < existingRecord.NumRows(); i++ {
			dbNameBuilder.Append(dbNameCol.Value(int(i)))
		}
		// 复制现有的表名
		tableNameCol := existingRecord.Column(3).(*array.String)
		tableNameBuilder := builder.Field(3).(*array.StringBuilder)
		for i := int64(0); i < existingRecord.NumRows(); i++ {
			tableNameBuilder.Append(tableNameCol.Value(int(i)))
		}
		// 复制现有的chunk_count
		chunkCountCol := existingRecord.Column(4).(*array.Int64)
		chunkCountBuilder := builder.Field(4).(*array.Int64Builder)
		for i := int64(0); i < existingRecord.NumRows(); i++ {
			chunkCountBuilder.Append(chunkCountCol.Value(int(i)))
		}
		// 复制现有的schema
		schemaCol := existingRecord.Column(5).(*array.Binary)
		schemaBuilder := builder.Field(5).(*array.BinaryBuilder)
		for i := int64(0); i < existingRecord.NumRows(); i++ {
			schemaBuilder.Append(schemaCol.Value(int(i)))
		}
	}

	// 添加新表记录
	builder.Field(0).(*array.Int64Builder).Append(maxID + 1)    // id
	builder.Field(1).(*array.Int64Builder).Append(1)            // db_id (固定为1，因为是系统表)
	builder.Field(2).(*array.StringBuilder).Append(dbName)      // db_name
	builder.Field(3).(*array.StringBuilder).Append(table.Table) // table_name
	builder.Field(4).(*array.Int64Builder).Append(0)            // chunk_count

	// 序列化schema
	schemaBytes, err := types.SerializeSchema(table.Schema)
	if err != nil {
		return fmt.Errorf("failed to serialize schema: %w", err)
	}
	builder.Field(5).(*array.BinaryBuilder).Append(schemaBytes)

	// 创建Record
	record := builder.NewRecord()
	defer record.Release()

	// 写入存储
	err = m.engine.Put(key, &record)
	if err != nil {
		return fmt.Errorf("failed to create table meta: %w", err)
	}

	return nil
}

// GetTable 获取表元数据。
func (m *MetadataManager) GetTable(dbName, tableName string) (TableMeta, error) {
	key := m.km.TableChunkKey(storage.SYS_DATABASE, storage.SYS_TABLES, 0)
	record, err := m.engine.Get(key)
	if err != nil {
		return TableMeta{}, fmt.Errorf("failed to get table meta: %w", err)
	}
	if record == nil {
		return TableMeta{}, fmt.Errorf("table not found")
	}
	defer record.Release()

	// 遍历记录查找匹配的表
	for i := int64(0); i < record.NumRows(); i++ {
		dbNameCol := record.Column(2).(*array.String)
		tableNameCol := record.Column(3).(*array.String)
		chunkCountCol := record.Column(4).(*array.Int64)
		schemaCol := record.Column(5).(*array.Binary)

		if dbNameCol.Value(int(i)) == dbName && tableNameCol.Value(int(i)) == tableName {
			// 找到匹配的表，解析schema
			schemaData := schemaCol.Value(int(i))
			chunkCountData := chunkCountCol.Value(int(i))
			schema, err := types.DeserializeSchema(schemaData)
			if err != nil {
				return TableMeta{}, fmt.Errorf("failed to deserialize schema: %w", err)
			}

			return TableMeta{
				Database:   dbName,
				Table:      tableName,
				ChunkCount: chunkCountData,
				Schema:     schema,
			}, nil
		}
	}

	return TableMeta{}, fmt.Errorf("table %s.%s not found", dbName, tableName)
}

// DeleteDatabase 删除数据库元数据。
func (m *MetadataManager) DeleteDatabase(name string) error {
	if name == storage.SYS_DATABASE {
		return fmt.Errorf("cannot delete system database")
	}

	key := m.km.TableChunkKey(storage.SYS_DATABASE, storage.SYS_DATABASES, 0)
	record, err := m.engine.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get database meta: %w", err)
	}
	if record == nil {
		return fmt.Errorf("database not found")
	}
	defer record.Release()

	// 检查数据库是否存在
	nameCol := record.Column(1).(*array.String)
	found := false
	for i := int64(0); i < record.NumRows(); i++ {
		if nameCol.Value(int(i)) == name {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("database %s not found", name)
	}

	// 创建新的Record，排除要删除的数据库
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, SysDatabasesSchema)
	defer builder.Release()

	idCol := record.Column(0).(*array.Int64)

	for i := int64(0); i < record.NumRows(); i++ {
		if nameCol.Value(int(i)) != name {
			builder.Field(0).(*array.Int64Builder).Append(idCol.Value(int(i)))
			builder.Field(1).(*array.StringBuilder).Append(nameCol.Value(int(i)))
		}
	}

	newRecord := builder.NewRecord()
	defer newRecord.Release()

	// 更新存储
	err = m.engine.Put(key, &newRecord)
	if err != nil {
		return fmt.Errorf("failed to update database meta: %w", err)
	}

	return nil
}

// DeleteTable 删除表元数据。
func (m *MetadataManager) DeleteTable(dbName, tableName string) error {
	if dbName == storage.SYS_DATABASE {
		return fmt.Errorf("cannot delete system tables")
	}

	key := m.km.TableChunkKey(storage.SYS_DATABASE, storage.SYS_TABLES, 0)
	record, err := m.engine.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get table meta: %w", err)
	}
	if record == nil {
		return fmt.Errorf("table not found")
	}
	defer record.Release()

	// 检查表是否存在
	dbNameCol := record.Column(2).(*array.String)
	tableNameCol := record.Column(3).(*array.String)
	found := false
	for i := int64(0); i < record.NumRows(); i++ {
		if dbNameCol.Value(int(i)) == dbName && tableNameCol.Value(int(i)) == tableName {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("table %s.%s not found", dbName, tableName)
	}

	// 创建新的Record，排除要删除的表
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, SysTablesSchema)
	defer builder.Release()

	for i := int64(0); i < record.NumRows(); i++ {
		if record.Column(2).(*array.String).Value(int(i)) != dbName ||
			record.Column(3).(*array.String).Value(int(i)) != tableName {
			// 复制所有字段
			builder.Field(0).(*array.Int64Builder).Append(record.Column(0).(*array.Int64).Value(int(i)))
			builder.Field(1).(*array.Int64Builder).Append(record.Column(1).(*array.Int64).Value(int(i)))
			builder.Field(2).(*array.StringBuilder).Append(record.Column(2).(*array.String).Value(int(i)))
			builder.Field(3).(*array.StringBuilder).Append(record.Column(3).(*array.String).Value(int(i)))
			builder.Field(4).(*array.Int64Builder).Append(record.Column(4).(*array.Int64).Value(int(i)))
			builder.Field(5).(*array.BinaryBuilder).Append(record.Column(5).(*array.Binary).Value(int(i)))
		}
	}

	newRecord := builder.NewRecord()
	defer newRecord.Release()

	// 更新存储
	err = m.engine.Put(key, &newRecord)
	if err != nil {
		return fmt.Errorf("failed to update table meta: %w", err)
	}

	return nil
}

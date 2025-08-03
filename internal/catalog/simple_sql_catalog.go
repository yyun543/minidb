package catalog

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/ipc"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/storage"
)

// SQLRunner SQL执行接口（简化版）
type SQLRunner interface {
	ExecuteSQL(sql string) (arrow.Record, error)
}

// SimpleSQLCatalog 简化的基于SQL的catalog实现
// 专注于SQL统一管理的核心思想，暂时使用简单的实现
type SimpleSQLCatalog struct {
	engine    storage.Engine
	mutex     sync.RWMutex
	sqlRunner SQLRunner

	// 临时的内存缓存（将来会完全通过SQL查询）
	databases map[string]*DatabaseInfo
	tables    map[string]map[string]*TableInfo
}

// NewSimpleSQLCatalog 创建简化的SQL-based catalog
func NewSimpleSQLCatalog(engine storage.Engine) *SimpleSQLCatalog {
	return &SimpleSQLCatalog{
		engine:    engine,
		databases: make(map[string]*DatabaseInfo),
		tables:    make(map[string]map[string]*TableInfo),
	}
}

// SetSQLRunner 设置SQL执行器
func (c *SimpleSQLCatalog) SetSQLRunner(runner SQLRunner) {
	c.sqlRunner = runner
}

// Init 初始化catalog
func (c *SimpleSQLCatalog) Init() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 如果有SQL执行器，使用SQL初始化
	if c.sqlRunner != nil {
		return c.initWithSQL()
	}

	// 否则使用简单初始化
	return c.simpleInit()
}

// initWithSQL 使用SQL初始化（未来的完整实现）
func (c *SimpleSQLCatalog) initWithSQL() error {
	// TODO: 使用SQL创建系统表
	// 1. CREATE DATABASE IF NOT EXISTS sys;
	// 2. CREATE TABLE IF NOT EXISTS sys.databases (id INT, name STRING);
	// 3. CREATE TABLE IF NOT EXISTS sys.tables (...);
	// 4. INSERT initial data...

	// 暂时先用简单初始化
	return c.simpleInit()
}

// simpleInit 简单初始化（向后兼容）
func (c *SimpleSQLCatalog) simpleInit() error {
	// 创建系统数据库
	c.databases["sys"] = &DatabaseInfo{Name: "sys"}
	c.databases["default"] = &DatabaseInfo{Name: "default"}

	// 初始化表映射
	c.tables["sys"] = make(map[string]*TableInfo)
	c.tables["default"] = make(map[string]*TableInfo)

	// 创建系统表
	if err := c.createSystemTables(); err != nil {
		return fmt.Errorf("failed to create system tables: %w", err)
	}

	// 从存储引擎恢复catalog元数据 (WAL recovery)
	if err := c.recoverCatalogMetadata(); err != nil {
		return fmt.Errorf("failed to recover catalog metadata: %w", err)
	}

	return nil
}

// createSystemTables 创建系统表
func (c *SimpleSQLCatalog) createSystemTables() error {
	// 创建 databases 系统表的 schema
	databasesSchema := arrow.NewSchema([]arrow.Field{
		{Name: "name", Type: arrow.BinaryTypes.String},
	}, nil)

	// 创建 tables 系统表的 schema
	tablesSchema := arrow.NewSchema([]arrow.Field{
		{Name: "database_name", Type: arrow.BinaryTypes.String},
		{Name: "table_name", Type: arrow.BinaryTypes.String},
	}, nil)

	// 添加系统表到内存缓存
	c.tables["sys"]["databases"] = &TableInfo{
		Database: "sys",
		Name:     "databases",
		Schema:   databasesSchema,
	}

	c.tables["sys"]["tables"] = &TableInfo{
		Database: "sys",
		Name:     "tables",
		Schema:   tablesSchema,
	}

	return nil
}

// CreateDatabase 通过SQL创建数据库
func (c *SimpleSQLCatalog) CreateDatabase(name string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查是否已存在
	if _, exists := c.databases[name]; exists {
		return fmt.Errorf("database '%s' already exists", name)
	}

	// 如果有SQL执行器，使用SQL
	if c.sqlRunner != nil {
		sql := fmt.Sprintf("INSERT INTO sys.databases (name) VALUES ('%s')", name)
		_, err := c.sqlRunner.ExecuteSQL(sql)
		if err != nil {
			// SQL执行失败，记录日志但继续使用简单方式
			fmt.Printf("SQL execution failed, falling back to simple mode: %v\n", err)
		}
	}

	// 持久化数据库元数据到存储引擎 (WAL支持)
	err := c.persistDatabaseMetadata(name)
	if err != nil {
		return fmt.Errorf("failed to persist database metadata: %w", err)
	}

	// 更新内存缓存
	c.databases[name] = &DatabaseInfo{Name: name}
	c.tables[name] = make(map[string]*TableInfo)

	return nil
}

// DropDatabase 删除数据库
func (c *SimpleSQLCatalog) DropDatabase(name string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if name == "sys" {
		return fmt.Errorf("cannot drop system database")
	}

	if _, exists := c.databases[name]; !exists {
		return fmt.Errorf("database '%s' does not exist", name)
	}

	// 如果有SQL执行器，使用SQL删除
	if c.sqlRunner != nil {
		// 删除表记录
		sql1 := fmt.Sprintf("DELETE FROM sys.tables WHERE db_name = '%s'", name)
		c.sqlRunner.ExecuteSQL(sql1)

		// 删除数据库记录
		sql2 := fmt.Sprintf("DELETE FROM sys.databases WHERE name = '%s'", name)
		c.sqlRunner.ExecuteSQL(sql2)
	}

	// 更新内存缓存
	delete(c.databases, name)
	delete(c.tables, name)

	return nil
}

// CreateTable 创建表
func (c *SimpleSQLCatalog) CreateTable(database string, tableMeta TableMeta) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查数据库是否存在
	if _, exists := c.databases[database]; !exists {
		return fmt.Errorf("database '%s' does not exist", database)
	}

	// 检查表是否已存在
	if tables, exists := c.tables[database]; exists {
		if _, exists := tables[tableMeta.Table]; exists {
			return fmt.Errorf("table '%s.%s' already exists", database, tableMeta.Table)
		}
	}

	// 如果有SQL执行器，使用SQL创建
	if c.sqlRunner != nil {
		sql := fmt.Sprintf(`
			INSERT INTO sys.tables (db_name, table_name, chunk_count, schema_info) 
			VALUES ('%s', '%s', %d, 'schema_placeholder')`,
			database, tableMeta.Table, tableMeta.ChunkCount)
		c.sqlRunner.ExecuteSQL(sql)
	}

	// 持久化表元数据到存储引擎 (WAL支持)
	err := c.persistTableMetadata(database, tableMeta.Table, tableMeta)
	if err != nil {
		return fmt.Errorf("failed to persist table metadata: %w", err)
	}

	// 更新内存缓存
	if c.tables[database] == nil {
		c.tables[database] = make(map[string]*TableInfo)
	}
	c.tables[database][tableMeta.Table] = &TableInfo{
		Database: database,
		Name:     tableMeta.Table,
		Schema:   tableMeta.Schema,
	}

	return nil
}

// GetTable 获取表信息
func (c *SimpleSQLCatalog) GetTable(database, table string) (TableMeta, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// 优先尝试SQL查询（如果有执行器）
	if c.sqlRunner != nil {
		sql := fmt.Sprintf(`
			SELECT db_name, table_name, chunk_count 
			FROM sys.tables 
			WHERE db_name = '%s' AND table_name = '%s'`,
			database, table)

		result, err := c.sqlRunner.ExecuteSQL(sql)
		if err == nil && result != nil {
			// 解析SQL结果（简化处理）
			// 在实际实现中，这里需要正确解析arrow.Record
		}
	}

	// 回退到内存缓存查询
	tables, exists := c.tables[database]
	if !exists {
		return TableMeta{}, fmt.Errorf("database '%s' does not exist", database)
	}

	tableInfo, exists := tables[table]
	if !exists {
		return TableMeta{}, fmt.Errorf("table '%s.%s' does not exist", database, table)
	}

	return TableMeta{
		Database:   tableInfo.Database,
		Table:      tableInfo.Name,
		ChunkCount: 0, // 简化
		Schema:     tableInfo.Schema,
	}, nil
}

// GetAllDatabases 获取所有数据库
func (c *SimpleSQLCatalog) GetAllDatabases() ([]string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// 优先尝试SQL查询
	if c.sqlRunner != nil {
		result, err := c.sqlRunner.ExecuteSQL("SELECT name FROM sys.databases ORDER BY name")
		if err == nil && result != nil {
			// 解析结果（简化处理）
			// return parseStringArray(result)
		}
	}

	// 回退到内存缓存
	databases := make([]string, 0, len(c.databases))
	for name := range c.databases {
		databases = append(databases, name)
	}
	return databases, nil
}

// GetAllTables 获取指定数据库的所有表
func (c *SimpleSQLCatalog) GetAllTables(database string) ([]string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// 优先尝试SQL查询
	if c.sqlRunner != nil {
		sql := fmt.Sprintf("SELECT table_name FROM sys.tables WHERE db_name = '%s' ORDER BY table_name", database)
		result, err := c.sqlRunner.ExecuteSQL(sql)
		if err == nil && result != nil {
			// 解析结果（简化处理）
		}
	}

	// 回退到内存缓存
	tables, exists := c.tables[database]
	if !exists {
		return nil, fmt.Errorf("database '%s' does not exist", database)
	}

	tableNames := make([]string, 0, len(tables))
	for name := range tables {
		tableNames = append(tableNames, name)
	}
	return tableNames, nil
}

// persistDatabaseMetadata 将数据库元数据持久化到存储引擎
func (c *SimpleSQLCatalog) persistDatabaseMetadata(dbName string) error {
	if c.engine == nil {
		return nil // 没有存储引擎，跳过持久化
	}

	// 创建元数据key
	key := []byte("catalog.database." + dbName)

	// 创建简单的数据库元数据记录
	metadata := map[string]interface{}{
		"name": dbName,
		"type": "database",
	}

	// 将元数据序列化为Arrow记录并存储
	return c.storeMetadataRecord(key, metadata)
}

// persistTableMetadata 将表元数据持久化到存储引擎
func (c *SimpleSQLCatalog) persistTableMetadata(dbName, tableName string, tableMeta TableMeta) error {
	if c.engine == nil {
		return nil // 没有存储引擎，跳过持久化
	}

	// 创建元数据key
	key := []byte("catalog.table." + dbName + "." + tableName)

	// 序列化schema信息为二进制数据
	schemaBytes := c.serializeSchema(tableMeta.Schema)

	// 直接创建Arrow记录来存储表元数据
	return c.storeTableMetadataRecord(key, dbName, tableName, tableMeta.ChunkCount, schemaBytes)
}

// storeTableMetadataRecord 存储表元数据记录（包含二进制schema）
func (c *SimpleSQLCatalog) storeTableMetadataRecord(key []byte, dbName, tableName string, chunkCount int64, schemaBytes []byte) error {
	// 创建专门的表元数据schema
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "key", Type: arrow.BinaryTypes.String},
		{Name: "database", Type: arrow.BinaryTypes.String},
		{Name: "table", Type: arrow.BinaryTypes.String},
		{Name: "type", Type: arrow.BinaryTypes.String},
		{Name: "chunk_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "schema_data", Type: arrow.BinaryTypes.Binary}, // 存储二进制schema数据
	}, nil)

	// 创建Arrow记录
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 添加数据
	keyBuilder := builder.Field(0).(*array.StringBuilder)
	dbBuilder := builder.Field(1).(*array.StringBuilder)
	tableBuilder := builder.Field(2).(*array.StringBuilder)
	typeBuilder := builder.Field(3).(*array.StringBuilder)
	chunkCountBuilder := builder.Field(4).(*array.Int64Builder)
	schemaDataBuilder := builder.Field(5).(*array.BinaryBuilder)

	keyBuilder.Append(string(key))
	dbBuilder.Append(dbName)
	tableBuilder.Append(tableName)
	typeBuilder.Append("table")
	chunkCountBuilder.Append(chunkCount)

	// 存储二进制schema数据
	if schemaBytes != nil {
		schemaDataBuilder.Append(schemaBytes)
	} else {
		schemaDataBuilder.AppendNull()
	}

	record := builder.NewRecord()
	defer record.Release()

	// 存储到引擎
	return c.engine.Put(key, &record)
}

// serializeSchema 使用Arrow IPC将schema序列化为二进制数据
func (c *SimpleSQLCatalog) serializeSchema(schema *arrow.Schema) []byte {
	if schema == nil {
		return nil
	}

	// 使用Arrow IPC序列化schema
	var buf bytes.Buffer
	writer := ipc.NewWriter(&buf, ipc.WithSchema(schema))

	// 创建一个空的记录批次来序列化schema
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 创建空记录（只为了序列化schema）
	record := builder.NewRecord()
	defer record.Release()

	// 写入记录（这会包含schema信息）
	if err := writer.Write(record); err != nil {
		return nil
	}

	writer.Close()
	return buf.Bytes()
}

// deserializeSchema 使用Arrow IPC从二进制数据反序列化schema
func (c *SimpleSQLCatalog) deserializeSchema(data []byte) *arrow.Schema {
	if len(data) == 0 {
		return arrow.NewSchema([]arrow.Field{}, nil)
	}

	// 使用Arrow IPC反序列化
	buf := bytes.NewReader(data)
	reader, err := ipc.NewReader(buf)
	if err != nil {
		return arrow.NewSchema([]arrow.Field{}, nil)
	}
	defer reader.Release()

	// 获取schema
	return reader.Schema()
}

// parseSchemaFromJSON 从JSON数据中解析schema（向后兼容旧格式）
func (c *SimpleSQLCatalog) parseSchemaFromJSON(dataJSON string) *arrow.Schema {
	// 简化的JSON解析，寻找schema字段
	schemaStart := strings.Index(dataJSON, `"schema":`)
	if schemaStart == -1 {
		return arrow.NewSchema([]arrow.Field{}, nil)
	}

	// 找到schema值的开始位置
	schemaStart = strings.Index(dataJSON[schemaStart:], "[")
	if schemaStart == -1 {
		return arrow.NewSchema([]arrow.Field{}, nil)
	}
	schemaStart += strings.Index(dataJSON, `"schema":`)

	// 找到schema值的结束位置
	schemaEnd := strings.Index(dataJSON[schemaStart:], "]")
	if schemaEnd == -1 {
		return arrow.NewSchema([]arrow.Field{}, nil)
	}
	schemaEnd += schemaStart + 1 // 包含]

	schemaJSON := dataJSON[schemaStart:schemaEnd]
	return c.deserializeSchemaFromJSON(schemaJSON)
}

// deserializeSchemaFromJSON 从JSON字符串反序列化schema（向后兼容）
func (c *SimpleSQLCatalog) deserializeSchemaFromJSON(schemaJSON string) *arrow.Schema {
	if schemaJSON == "" || schemaJSON == "[]" {
		return arrow.NewSchema([]arrow.Field{}, nil)
	}

	// 简化的JSON解析（生产环境应该使用标准JSON解析）
	// 这里只处理基本的字段名和类型
	var fields []arrow.Field

	// 移除方括号
	schemaJSON = strings.Trim(schemaJSON, "[]")
	if schemaJSON == "" {
		return arrow.NewSchema([]arrow.Field{}, nil)
	}

	// 分割字段
	fieldStrs := strings.Split(schemaJSON, "},{")
	for _, fieldStr := range fieldStrs {
		fieldStr = strings.Trim(fieldStr, "{}")

		// 解析name和type
		parts := strings.Split(fieldStr, ",")
		var name, typeStr string

		for _, part := range parts {
			if strings.Contains(part, "\"name\":") {
				name = strings.Trim(strings.Split(part, ":")[1], "\"")
			} else if strings.Contains(part, "\"type\":") {
				typeStr = strings.Trim(strings.Split(part, ":")[1], "\"")
			}
		}

		// 根据类型字符串创建Arrow类型
		var dataType arrow.DataType
		switch {
		case strings.Contains(typeStr, "int64"):
			dataType = arrow.PrimitiveTypes.Int64
		case strings.Contains(typeStr, "string"):
			dataType = arrow.BinaryTypes.String
		case strings.Contains(typeStr, "float64"):
			dataType = arrow.PrimitiveTypes.Float64
		case strings.Contains(typeStr, "bool"):
			dataType = arrow.FixedWidthTypes.Boolean
		default:
			// 默认字符串类型
			dataType = arrow.BinaryTypes.String
		}

		if name != "" {
			fields = append(fields, arrow.Field{Name: name, Type: dataType})
		}
	}

	return arrow.NewSchema(fields, nil)
}

// storeMetadataRecord 将元数据存储为Arrow记录
func (c *SimpleSQLCatalog) storeMetadataRecord(key []byte, metadata map[string]interface{}) error {
	// 创建简单的Arrow schema用于元数据
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "key", Type: arrow.BinaryTypes.String},
		{Name: "data", Type: arrow.BinaryTypes.String},
	}, nil)

	// 创建Arrow记录
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 添加数据
	keyBuilder := builder.Field(0).(*array.StringBuilder)
	dataBuilder := builder.Field(1).(*array.StringBuilder)

	keyBuilder.Append(string(key))

	// 将metadata转换为完整的JSON字符串
	var metadataJSON string
	if schema, exists := metadata["schema"]; exists {
		metadataJSON = fmt.Sprintf(`{"type":"%s","name":"%s","database":"%s","schema":%s}`,
			metadata["type"], metadata["name"], metadata["database"], schema)
	} else {
		metadataJSON = fmt.Sprintf(`{"type":"%s","name":"%s"}`,
			metadata["type"], metadata["name"])
	}
	dataBuilder.Append(metadataJSON)

	record := builder.NewRecord()
	defer record.Release()

	// 存储到引擎 (会自动写入WAL)
	return c.engine.Put(key, &record)
}

// recoverCatalogMetadata 从存储引擎恢复catalog元数据
func (c *SimpleSQLCatalog) recoverCatalogMetadata() error {
	if c.engine == nil {
		return nil // 没有存储引擎，跳过恢复
	}

	// 扫描所有catalog相关的key
	startKey := []byte("catalog.")
	endKey := []byte("catalog.z") // 确保覆盖所有catalog.开头的key

	iterator, err := c.engine.Scan(startKey, endKey)
	if err != nil {
		return fmt.Errorf("failed to scan catalog metadata: %w", err)
	}
	defer iterator.Close()

	// 遍历所有catalog元数据记录
	for iterator.Next() {
		record := iterator.Record()

		// 解析记录（简化处理）
		if record.NumRows() > 0 {
			// 获取key列
			keyArray := record.Column(0).(*array.String)
			if keyArray.Len() > 0 {
				key := keyArray.Value(0)

				// 根据key类型进行恢复
				if strings.HasPrefix(key, "catalog.database.") {
					// 恢复数据库
					dbName := strings.TrimPrefix(key, "catalog.database.")
					c.databases[dbName] = &DatabaseInfo{Name: dbName}
					if c.tables[dbName] == nil {
						c.tables[dbName] = make(map[string]*TableInfo)
					}
				} else if strings.HasPrefix(key, "catalog.table.") {
					// 恢复表 - 从新的二进制格式中恢复
					tableKey := strings.TrimPrefix(key, "catalog.table.")
					parts := strings.SplitN(tableKey, ".", 2)
					if len(parts) == 2 {
						dbName := parts[0]
						tableName := parts[1]

						// 确保数据库存在
						if _, exists := c.databases[dbName]; !exists {
							c.databases[dbName] = &DatabaseInfo{Name: dbName}
							c.tables[dbName] = make(map[string]*TableInfo)
						}

						// 从新格式的记录中恢复schema
						var schema *arrow.Schema
						if record.NumCols() >= 6 {
							// 新格式：检查是否有schema_data列（第5列，索引为5）
							if schemaDataArray, ok := record.Column(5).(*array.Binary); ok && schemaDataArray.Len() > 0 {
								if !schemaDataArray.IsNull(0) {
									schemaBytes := schemaDataArray.Value(0)
									schema = c.deserializeSchema(schemaBytes)
								}
							}
						} else if record.NumCols() > 1 {
							// 兼容旧格式：尝试从JSON解析
							if dataArray, ok := record.Column(1).(*array.String); ok && dataArray.Len() > 0 {
								dataJSON := dataArray.Value(0)
								schema = c.parseSchemaFromJSON(dataJSON)
							}
						}

						// 如果没有找到schema信息，使用空schema
						if schema == nil {
							schema = arrow.NewSchema([]arrow.Field{}, nil)
						}

						// 恢复表信息
						c.tables[dbName][tableName] = &TableInfo{
							Database: dbName,
							Name:     tableName,
							Schema:   schema,
						}
					}
				}
			}
		}
	}

	return nil
}

// DropTable 删除表
func (c *SimpleSQLCatalog) DropTable(database, table string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查表是否存在
	if tables, exists := c.tables[database]; !exists || tables[table] == nil {
		return fmt.Errorf("table '%s.%s' does not exist", database, table)
	}

	// 使用SQL删除
	if c.sqlRunner != nil {
		sql := fmt.Sprintf("DELETE FROM sys.tables WHERE db_name = '%s' AND table_name = '%s'", database, table)
		c.sqlRunner.ExecuteSQL(sql)
	}

	// 更新内存缓存
	delete(c.tables[database], table)
	return nil
}

// UpdateTable 更新表元数据
func (c *SimpleSQLCatalog) UpdateTable(dbName string, table TableMeta) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查表是否存在
	if tables, exists := c.tables[dbName]; !exists || tables[table.Table] == nil {
		return fmt.Errorf("table '%s.%s' does not exist", dbName, table.Table)
	}

	// 使用SQL更新
	if c.sqlRunner != nil {
		sql := fmt.Sprintf(`
			UPDATE sys.tables 
			SET chunk_count = %d 
			WHERE db_name = '%s' AND table_name = '%s'`,
			table.ChunkCount, dbName, table.Table)
		c.sqlRunner.ExecuteSQL(sql)
	}

	// 更新内存缓存
	c.tables[dbName][table.Table] = &TableInfo{
		Database: table.Database,
		Name:     table.Table,
		Schema:   table.Schema,
	}

	return nil
}

// GetEngine 获取存储引擎
func (c *SimpleSQLCatalog) GetEngine() storage.Engine {
	return c.engine
}

// 兼容性方法
func (c *SimpleSQLCatalog) GetDatabase(name string) (DatabaseMeta, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if _, exists := c.databases[name]; !exists {
		return DatabaseMeta{}, fmt.Errorf("database '%s' does not exist", name)
	}
	return DatabaseMeta{Name: name}, nil
}

func (c *SimpleSQLCatalog) DeleteDatabase(name string) error {
	return c.DropDatabase(name)
}

func (c *SimpleSQLCatalog) DeleteTable(dbName, tableName string) error {
	return c.DropTable(dbName, tableName)
}

// DatabaseInfo 数据库信息
type DatabaseInfo struct {
	Name string `json:"name"`
}

// TableInfo 表信息
type TableInfo struct {
	Database string        `json:"database"`
	Name     string        `json:"name"`
	Schema   *arrow.Schema `json:"schema"`
}

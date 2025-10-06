package catalog

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/ipc"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/logger"
	"github.com/yyun543/minidb/internal/storage"
	"go.uber.org/zap"
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
	indexes   map[string]map[string]map[string]*IndexInfo // database -> table -> index_name -> IndexInfo
}

// NewSimpleSQLCatalog 创建简化的SQL-based catalog
func NewSimpleSQLCatalog(engine storage.Engine) *SimpleSQLCatalog {
	logger.WithComponent("catalog").Info("Creating SimpleSQLCatalog instance",
		zap.String("engine_type", fmt.Sprintf("%T", engine)))

	catalog := &SimpleSQLCatalog{
		engine:    engine,
		databases: make(map[string]*DatabaseInfo),
		tables:    make(map[string]map[string]*TableInfo),
		indexes:   make(map[string]map[string]map[string]*IndexInfo),
	}

	logger.WithComponent("catalog").Info("SimpleSQLCatalog instance created successfully")
	return catalog
}

// SetSQLRunner 设置SQL执行器
func (c *SimpleSQLCatalog) SetSQLRunner(runner SQLRunner) {
	logger.WithComponent("catalog").Info("Setting SQL runner for SimpleSQLCatalog",
		zap.String("runner_type", fmt.Sprintf("%T", runner)))

	c.sqlRunner = runner

	logger.WithComponent("catalog").Info("SQL runner set successfully for SimpleSQLCatalog")
}

// Init 初始化catalog
func (c *SimpleSQLCatalog) Init() error {
	logger.WithComponent("catalog").Info("Initializing SimpleSQLCatalog")

	start := time.Now()
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 如果有SQL执行器，使用SQL初始化
	if c.sqlRunner != nil {
		logger.WithComponent("catalog").Info("SQL runner available, using SQL-based initialization")
		err := c.initWithSQL()
		if err != nil {
			logger.WithComponent("catalog").Error("SQL-based initialization failed",
				zap.Duration("duration", time.Since(start)),
				zap.Error(err))
		} else {
			logger.WithComponent("catalog").Info("SimpleSQLCatalog initialized successfully with SQL",
				zap.Duration("initialization_time", time.Since(start)))
		}
		return err
	}

	// 否则使用简单初始化
	logger.WithComponent("catalog").Info("No SQL runner available, using simple initialization")
	err := c.simpleInit()
	if err != nil {
		logger.WithComponent("catalog").Error("Simple initialization failed",
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))
	} else {
		logger.WithComponent("catalog").Info("SimpleSQLCatalog initialized successfully with simple mode",
			zap.Duration("initialization_time", time.Since(start)))
	}
	return err
}

// initWithSQL 使用SQL初始化（未来的完整实现）
func (c *SimpleSQLCatalog) initWithSQL() error {
	logger.WithComponent("catalog").Info("Starting SQL-based catalog initialization")

	// TODO: 使用SQL创建系统表
	// 1. CREATE DATABASE IF NOT EXISTS sys;
	// 2. CREATE TABLE IF NOT EXISTS sys.databases (id INT, name STRING);
	// 3. CREATE TABLE IF NOT EXISTS sys.tables (...);
	// 4. INSERT initial data...

	logger.WithComponent("catalog").Warn("SQL-based initialization not fully implemented yet, falling back to simple initialization")
	// 暂时先用简单初始化
	return c.simpleInit()
}

// simpleInit 简单初始化（向后兼容）
func (c *SimpleSQLCatalog) simpleInit() error {
	logger.WithComponent("catalog").Info("Starting simple catalog initialization")

	start := time.Now()

	// 创建系统数据库
	logger.WithComponent("catalog").Debug("Creating system databases")
	c.databases["sys"] = &DatabaseInfo{Name: "sys"}
	c.databases["default"] = &DatabaseInfo{Name: "default"}

	// 初始化表映射
	c.tables["sys"] = make(map[string]*TableInfo)
	c.tables["default"] = make(map[string]*TableInfo)
	logger.WithComponent("catalog").Debug("System databases and table mappings created")

	// 创建系统表
	systemTableStart := time.Now()
	if err := c.createSystemTables(); err != nil {
		logger.WithComponent("catalog").Error("Failed to create system tables",
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))
		return fmt.Errorf("failed to create system tables: %w", err)
	}
	logger.WithComponent("catalog").Debug("System tables created successfully",
		zap.Duration("system_tables_duration", time.Since(systemTableStart)))

	// 从存储引擎恢复catalog元数据 (WAL recovery)
	recoveryStart := time.Now()
	if err := c.recoverCatalogMetadata(); err != nil {
		logger.WithComponent("catalog").Error("Failed to recover catalog metadata",
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))
		return fmt.Errorf("failed to recover catalog metadata: %w", err)
	}
	logger.WithComponent("catalog").Info("Catalog metadata recovered successfully",
		zap.Duration("recovery_duration", time.Since(recoveryStart)))

	// 确保系统数据库和表在恢复后仍然存在（防御性编程）
	c.ensureSystemEntitiesExist()

	totalDuration := time.Since(start)
	logger.WithComponent("catalog").Info("Simple catalog initialization completed",
		zap.Duration("total_duration", totalDuration),
		zap.Int("databases_count", len(c.databases)),
		zap.Int("sys_tables_count", len(c.tables["sys"])),
		zap.Int("default_tables_count", len(c.tables["default"])))

	return nil
}

// ensureSystemEntitiesExist 确保系统数据库和表存在
func (c *SimpleSQLCatalog) ensureSystemEntitiesExist() {
	// 确保系统数据库存在
	if c.databases["sys"] == nil {
		c.databases["sys"] = &DatabaseInfo{Name: "sys"}
	}
	if c.databases["default"] == nil {
		c.databases["default"] = &DatabaseInfo{Name: "default"}
	}

	// 确保系统数据库的表映射存在
	if c.tables["sys"] == nil {
		c.tables["sys"] = make(map[string]*TableInfo)
	}
	if c.tables["default"] == nil {
		c.tables["default"] = make(map[string]*TableInfo)
	}

	// 确保系统表存在
	if c.tables["sys"]["schemata"] == nil {
		schemataSchema := arrow.NewSchema([]arrow.Field{
			{Name: "schema_name", Type: arrow.BinaryTypes.String},
		}, nil)
		c.tables["sys"]["schemata"] = &TableInfo{
			Database: "sys",
			Name:     "schemata",
			Schema:   schemataSchema,
		}
	}

	if c.tables["sys"]["table_catalog"] == nil {
		tableCatalogSchema := arrow.NewSchema([]arrow.Field{
			{Name: "table_schema", Type: arrow.BinaryTypes.String},
			{Name: "table_name", Type: arrow.BinaryTypes.String},
		}, nil)
		c.tables["sys"]["table_catalog"] = &TableInfo{
			Database: "sys",
			Name:     "table_catalog",
			Schema:   tableCatalogSchema,
		}
	}
}

// createSystemTables 创建系统表
func (c *SimpleSQLCatalog) createSystemTables() error {
	// 创建 schemata 系统表的 schema (替代 databases)
	schemataSchema := arrow.NewSchema([]arrow.Field{
		{Name: "schema_name", Type: arrow.BinaryTypes.String},
	}, nil)

	// 创建 table_catalog 系统表的 schema (替代 tables)
	tableCatalogSchema := arrow.NewSchema([]arrow.Field{
		{Name: "table_schema", Type: arrow.BinaryTypes.String},
		{Name: "table_name", Type: arrow.BinaryTypes.String},
	}, nil)

	// 添加系统表到内存缓存
	c.tables["sys"]["schemata"] = &TableInfo{
		Database: "sys",
		Name:     "schemata",
		Schema:   schemataSchema,
	}

	c.tables["sys"]["table_catalog"] = &TableInfo{
		Database: "sys",
		Name:     "table_catalog",
		Schema:   tableCatalogSchema,
	}

	return nil
}

// CreateDatabase 通过SQL创建数据库
func (c *SimpleSQLCatalog) CreateDatabase(name string) error {
	logger.WithComponent("catalog").Info("Creating database",
		zap.String("database", name))

	start := time.Now()
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查是否已存在
	if _, exists := c.databases[name]; exists {
		logger.WithComponent("catalog").Warn("Database creation failed - database already exists",
			zap.String("database", name))
		return fmt.Errorf("database '%s' already exists", name)
	}

	// 如果有SQL执行器，使用SQL
	if c.sqlRunner != nil {
		logger.WithComponent("catalog").Debug("Using SQL runner to create database",
			zap.String("database", name))
		sql := fmt.Sprintf("INSERT INTO sys.schemata (schema_name) VALUES ('%s')", name)
		_, err := c.sqlRunner.ExecuteSQL(sql)
		if err != nil {
			// SQL执行失败，记录日志但继续使用简单方式
			logger.WithComponent("catalog").Warn("SQL execution failed for database creation, falling back to simple mode",
				zap.String("database", name),
				zap.String("sql", sql),
				zap.Error(err))
			// Note: Also keeping fmt.Printf for backward compatibility
			fmt.Printf("SQL execution failed, falling back to simple mode: %v\n", err)
		} else {
			logger.WithComponent("catalog").Debug("Database created successfully via SQL",
				zap.String("database", name))
		}
	}

	// 持久化数据库元数据到存储引擎 (WAL支持)
	persistStart := time.Now()
	err := c.persistDatabaseMetadata(name)
	if err != nil {
		logger.WithComponent("catalog").Error("Failed to persist database metadata",
			zap.String("database", name),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))
		return fmt.Errorf("failed to persist database metadata: %w", err)
	}
	logger.WithComponent("catalog").Debug("Database metadata persisted successfully",
		zap.String("database", name),
		zap.Duration("persist_duration", time.Since(persistStart)))

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
		sql1 := fmt.Sprintf("DELETE FROM sys.table_catalog WHERE table_schema = '%s'", name)
		c.sqlRunner.ExecuteSQL(sql1)

		// 删除数据库记录
		sql2 := fmt.Sprintf("DELETE FROM sys.schemata WHERE schema_name = '%s'", name)
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
			INSERT INTO sys.table_catalog (table_schema, table_name, chunk_count, schema_info) 
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
			SELECT table_schema, table_name, chunk_count 
			FROM sys.table_catalog 
			WHERE table_schema = '%s' AND table_name = '%s'`,
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

	// 直接从内存缓存获取，这是权威数据源
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
		sql := fmt.Sprintf("SELECT table_name FROM sys.table_catalog WHERE table_schema = '%s' ORDER BY table_name", database)
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
		{Name: "table_schema", Type: arrow.BinaryTypes.String}, // 重命名 database -> table_schema
		{Name: "table_name", Type: arrow.BinaryTypes.String},   // 重命名 table -> table_name
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
		sql := fmt.Sprintf("DELETE FROM sys.table_catalog WHERE table_schema = '%s' AND table_name = '%s'", database, table)
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
			UPDATE sys.table_catalog 
			SET chunk_count = %d 
			WHERE table_schema = '%s' AND table_name = '%s'`,
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

// IndexInfo 索引信息
type IndexInfo struct {
	Database  string   `json:"database"`
	Table     string   `json:"table"`
	Name      string   `json:"name"`
	Columns   []string `json:"columns"`
	IsUnique  bool     `json:"is_unique"`
	IndexType string   `json:"index_type"`
}

// CreateIndex 创建索引
func (c *SimpleSQLCatalog) CreateIndex(indexMeta IndexMeta) error {
	logger.WithComponent("catalog").Info("Creating index",
		zap.String("database", indexMeta.Database),
		zap.String("table", indexMeta.Table),
		zap.String("index", indexMeta.Name))

	start := time.Now()
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查数据库是否存在
	if _, exists := c.databases[indexMeta.Database]; !exists {
		return fmt.Errorf("database '%s' does not exist", indexMeta.Database)
	}

	// 检查表是否存在
	if tables, exists := c.tables[indexMeta.Database]; !exists || tables[indexMeta.Table] == nil {
		return fmt.Errorf("table '%s.%s' does not exist", indexMeta.Database, indexMeta.Table)
	}

	// 初始化索引映射
	if c.indexes[indexMeta.Database] == nil {
		c.indexes[indexMeta.Database] = make(map[string]map[string]*IndexInfo)
	}
	if c.indexes[indexMeta.Database][indexMeta.Table] == nil {
		c.indexes[indexMeta.Database][indexMeta.Table] = make(map[string]*IndexInfo)
	}

	// 检查索引是否已存在
	if _, exists := c.indexes[indexMeta.Database][indexMeta.Table][indexMeta.Name]; exists {
		return fmt.Errorf("index '%s' already exists on table '%s.%s'", indexMeta.Name, indexMeta.Database, indexMeta.Table)
	}

	// 验证列是否存在于表中
	tableInfo := c.tables[indexMeta.Database][indexMeta.Table]
	for _, col := range indexMeta.Columns {
		found := false
		for _, field := range tableInfo.Schema.Fields() {
			if field.Name == col {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("column '%s' does not exist in table '%s.%s'", col, indexMeta.Database, indexMeta.Table)
		}
	}

	// 存储索引信息
	c.indexes[indexMeta.Database][indexMeta.Table][indexMeta.Name] = &IndexInfo{
		Database:  indexMeta.Database,
		Table:     indexMeta.Table,
		Name:      indexMeta.Name,
		Columns:   indexMeta.Columns,
		IsUnique:  indexMeta.IsUnique,
		IndexType: indexMeta.IndexType,
	}

	// 持久化索引元数据
	err := c.persistIndexMetadata(indexMeta)
	if err != nil {
		logger.WithComponent("catalog").Error("Failed to persist index metadata",
			zap.String("index", indexMeta.Name),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))
		return fmt.Errorf("failed to persist index metadata: %w", err)
	}

	logger.WithComponent("catalog").Info("Index created successfully",
		zap.String("index", indexMeta.Name),
		zap.Duration("duration", time.Since(start)))

	return nil
}

// DropIndex 删除索引
func (c *SimpleSQLCatalog) DropIndex(database, table, indexName string) error {
	logger.WithComponent("catalog").Info("Dropping index",
		zap.String("database", database),
		zap.String("table", table),
		zap.String("index", indexName))

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查索引是否存在
	if c.indexes[database] == nil || c.indexes[database][table] == nil || c.indexes[database][table][indexName] == nil {
		return fmt.Errorf("index '%s' does not exist on table '%s.%s'", indexName, database, table)
	}

	// 删除索引
	delete(c.indexes[database][table], indexName)

	logger.WithComponent("catalog").Info("Index dropped successfully",
		zap.String("index", indexName))

	return nil
}

// GetIndex 获取索引信息
func (c *SimpleSQLCatalog) GetIndex(database, table, indexName string) (IndexMeta, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.indexes[database] == nil || c.indexes[database][table] == nil {
		return IndexMeta{}, fmt.Errorf("no indexes found for table '%s.%s'", database, table)
	}

	indexInfo, exists := c.indexes[database][table][indexName]
	if !exists {
		return IndexMeta{}, fmt.Errorf("index '%s' does not exist on table '%s.%s'", indexName, database, table)
	}

	return IndexMeta{
		Database:  indexInfo.Database,
		Table:     indexInfo.Table,
		Name:      indexInfo.Name,
		Columns:   indexInfo.Columns,
		IsUnique:  indexInfo.IsUnique,
		IndexType: indexInfo.IndexType,
	}, nil
}

// GetAllIndexes 获取表的所有索引
func (c *SimpleSQLCatalog) GetAllIndexes(database, table string) ([]IndexMeta, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.indexes[database] == nil || c.indexes[database][table] == nil {
		return []IndexMeta{}, nil // 返回空列表而不是错误
	}

	indexes := make([]IndexMeta, 0, len(c.indexes[database][table]))
	for _, indexInfo := range c.indexes[database][table] {
		indexes = append(indexes, IndexMeta{
			Database:  indexInfo.Database,
			Table:     indexInfo.Table,
			Name:      indexInfo.Name,
			Columns:   indexInfo.Columns,
			IsUnique:  indexInfo.IsUnique,
			IndexType: indexInfo.IndexType,
		})
	}

	return indexes, nil
}

// persistIndexMetadata 持久化索引元数据
func (c *SimpleSQLCatalog) persistIndexMetadata(indexMeta IndexMeta) error {
	if c.engine == nil {
		return nil // 没有存储引擎，跳过持久化
	}

	// 创建元数据key
	key := []byte(fmt.Sprintf("catalog.index.%s.%s.%s", indexMeta.Database, indexMeta.Table, indexMeta.Name))

	// 创建索引元数据schema
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "key", Type: arrow.BinaryTypes.String},
		{Name: "database", Type: arrow.BinaryTypes.String},
		{Name: "table", Type: arrow.BinaryTypes.String},
		{Name: "name", Type: arrow.BinaryTypes.String},
		{Name: "columns", Type: arrow.BinaryTypes.String}, // JSON encoded
		{Name: "is_unique", Type: arrow.FixedWidthTypes.Boolean},
		{Name: "index_type", Type: arrow.BinaryTypes.String},
	}, nil)

	// 创建Arrow记录
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 添加数据
	builder.Field(0).(*array.StringBuilder).Append(string(key))
	builder.Field(1).(*array.StringBuilder).Append(indexMeta.Database)
	builder.Field(2).(*array.StringBuilder).Append(indexMeta.Table)
	builder.Field(3).(*array.StringBuilder).Append(indexMeta.Name)

	// Encode columns as JSON
	columnsJSON := strings.Join(indexMeta.Columns, ",")
	builder.Field(4).(*array.StringBuilder).Append(columnsJSON)

	builder.Field(5).(*array.BooleanBuilder).Append(indexMeta.IsUnique)
	builder.Field(6).(*array.StringBuilder).Append(indexMeta.IndexType)

	record := builder.NewRecord()
	defer record.Release()

	// 存储到引擎
	return c.engine.Put(key, &record)
}

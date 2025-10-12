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
	"github.com/yyun543/minidb/internal/delta"
	"github.com/yyun543/minidb/internal/logger"
	"github.com/yyun543/minidb/internal/storage"
	"go.uber.org/zap"
)

// SQLRunner SQL执行接口（简化版）
type SQLRunner interface {
	ExecuteSQL(sql string) (arrow.Record, error)
}

// SimpleSQLCatalog SQL自举的catalog实现
// 所有元数据操作通过SQL完成，元数据存储在sys数据库的系统表中
type SimpleSQLCatalog struct {
	storageEngine storage.StorageEngine // v2.0 storage engine
	mutex         sync.RWMutex
	sqlRunner     SQLRunner

	// 临时的内存缓存（将来会完全通过SQL查询）
	databases map[string]*DatabaseInfo
	tables    map[string]map[string]*TableInfo
	indexes   map[string]map[string]map[string]*IndexInfo // database -> table -> index_name -> IndexInfo
}

// NewSimpleSQLCatalog 创建简化的SQL-based catalog (v2.0)
func NewSimpleSQLCatalog() *SimpleSQLCatalog {
	logger.WithComponent("catalog").Info("Creating SimpleSQLCatalog instance (v2.0)")

	catalog := &SimpleSQLCatalog{
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

// initWithSQL 使用SQL自举初始化
// 按照SQL-first原则，所有元数据操作通过SQL完成
func (c *SimpleSQLCatalog) initWithSQL() error {
	logger.WithComponent("catalog").Info("Starting SQL-based catalog initialization (SQL Bootstrap)")

	// 第一步：确保基本的内存结构存在（引导阶段必需）
	c.databases = make(map[string]*DatabaseInfo)
	c.tables = make(map[string]map[string]*TableInfo)
	c.indexes = make(map[string]map[string]map[string]*IndexInfo)

	// 创建sys和default数据库（引导阶段）
	c.databases["sys"] = &DatabaseInfo{Name: "sys"}
	c.databases["default"] = &DatabaseInfo{Name: "default"}
	c.tables["sys"] = make(map[string]*TableInfo)
	c.tables["default"] = make(map[string]*TableInfo)

	// 第二步：创建系统表结构（这些表的schema必须先定义好）
	if err := c.createSystemTables(); err != nil {
		return fmt.Errorf("failed to create system tables: %w", err)
	}

	// 第三步：使用SQL插入初始元数据到系统表
	// 这是SQL自举的核心：系统表创建完成后，后续所有元数据操作都通过SQL
	if err := c.bootstrapSystemMetadata(); err != nil {
		logger.WithComponent("catalog").Error("Failed to bootstrap system metadata via SQL",
			zap.Error(err))
		return fmt.Errorf("failed to bootstrap system metadata: %w", err)
	}

	// 第四步：从系统表恢复catalog状态（读取已有的元数据）
	if err := c.loadMetadataFromSQL(); err != nil {
		logger.WithComponent("catalog").Error("Failed to load metadata from SQL",
			zap.Error(err))
		return fmt.Errorf("failed to load metadata from SQL: %w", err)
	}

	logger.WithComponent("catalog").Info("SQL-based catalog initialization completed successfully")
	return nil
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

	// v2.0: Catalog metadata is recovered from Delta Log system tables
	// 确保系统数据库和表存在（防御性编程）
	c.ensureSystemEntitiesExist()

	// 从 Delta Log 恢复用户数据库和表的元数据
	if err := c.loadMetadataFromSQL(); err != nil {
		logger.WithComponent("catalog").Error("Failed to load metadata from Delta Log",
			zap.Error(err))
		// 不返回错误，只是警告（向后兼容）
	}

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
	if c.tables["sys"]["db_metadata"] == nil {
		dbMetadataSchema := arrow.NewSchema([]arrow.Field{
			{Name: "db_name", Type: arrow.BinaryTypes.String},
		}, nil)
		c.tables["sys"]["db_metadata"] = &TableInfo{
			Database: "sys",
			Name:     "db_metadata",
			Schema:   dbMetadataSchema,
		}
	}

	if c.tables["sys"]["table_metadata"] == nil {
		tableMetadataSchema := arrow.NewSchema([]arrow.Field{
			{Name: "db_name", Type: arrow.BinaryTypes.String},
			{Name: "table_name", Type: arrow.BinaryTypes.String},
		}, nil)
		c.tables["sys"]["table_metadata"] = &TableInfo{
			Database: "sys",
			Name:     "table_metadata",
			Schema:   tableMetadataSchema,
		}
	}

	if c.tables["sys"]["columns_metadata"] == nil {
		columnsMetadataSchema := arrow.NewSchema([]arrow.Field{
			{Name: "db_name", Type: arrow.BinaryTypes.String},
			{Name: "table_name", Type: arrow.BinaryTypes.String},
			{Name: "column_name", Type: arrow.BinaryTypes.String},
			{Name: "ordinal_position", Type: arrow.PrimitiveTypes.Int64},
			{Name: "data_type", Type: arrow.BinaryTypes.String},
			{Name: "is_nullable", Type: arrow.BinaryTypes.String},
		}, nil)
		c.tables["sys"]["columns_metadata"] = &TableInfo{
			Database: "sys",
			Name:     "columns_metadata",
			Schema:   columnsMetadataSchema,
		}
	}

	if c.tables["sys"]["index_metadata"] == nil {
		indexMetadataSchema := arrow.NewSchema([]arrow.Field{
			{Name: "db_name", Type: arrow.BinaryTypes.String},
			{Name: "table_name", Type: arrow.BinaryTypes.String},
			{Name: "index_name", Type: arrow.BinaryTypes.String},
			{Name: "index_type", Type: arrow.BinaryTypes.String},
			{Name: "column_name", Type: arrow.BinaryTypes.String},
			{Name: "is_unique", Type: arrow.BinaryTypes.String},
		}, nil)
		c.tables["sys"]["index_metadata"] = &TableInfo{
			Database: "sys",
			Name:     "index_metadata",
			Schema:   indexMetadataSchema,
		}
	}

	if c.tables["sys"]["delta_log"] == nil {
		deltaLogSchema := arrow.NewSchema([]arrow.Field{
			{Name: "version", Type: arrow.PrimitiveTypes.Int64},
			{Name: "timestamp", Type: arrow.BinaryTypes.String},
			{Name: "operation", Type: arrow.BinaryTypes.String},
			{Name: "db_name", Type: arrow.BinaryTypes.String},
			{Name: "table_name", Type: arrow.BinaryTypes.String},
			{Name: "file_path", Type: arrow.BinaryTypes.String},
		}, nil)
		c.tables["sys"]["delta_log"] = &TableInfo{
			Database: "sys",
			Name:     "delta_log",
			Schema:   deltaLogSchema,
		}
	}

	if c.tables["sys"]["table_files"] == nil {
		tableFilesSchema := arrow.NewSchema([]arrow.Field{
			{Name: "db_name", Type: arrow.BinaryTypes.String},
			{Name: "table_name", Type: arrow.BinaryTypes.String},
			{Name: "file_path", Type: arrow.BinaryTypes.String},
			{Name: "file_size", Type: arrow.PrimitiveTypes.Int64},
			{Name: "row_count", Type: arrow.PrimitiveTypes.Int64},
			{Name: "status", Type: arrow.BinaryTypes.String},
		}, nil)
		c.tables["sys"]["table_files"] = &TableInfo{
			Database: "sys",
			Name:     "table_files",
			Schema:   tableFilesSchema,
		}
	}
}

// createSystemTables 创建系统表
func (c *SimpleSQLCatalog) createSystemTables() error {
	// 创建 db_metadata 系统表的 schema (替代 schemata)
	dbMetadataSchema := arrow.NewSchema([]arrow.Field{
		{Name: "db_name", Type: arrow.BinaryTypes.String},
	}, nil)

	// 创建 table_metadata 系统表的 schema (替代 table_catalog)
	tableMetadataSchema := arrow.NewSchema([]arrow.Field{
		{Name: "db_name", Type: arrow.BinaryTypes.String},
		{Name: "table_name", Type: arrow.BinaryTypes.String},
	}, nil)

	// 创建 columns_metadata 系统表的 schema
	columnsMetadataSchema := arrow.NewSchema([]arrow.Field{
		{Name: "db_name", Type: arrow.BinaryTypes.String},
		{Name: "table_name", Type: arrow.BinaryTypes.String},
		{Name: "column_name", Type: arrow.BinaryTypes.String},
		{Name: "ordinal_position", Type: arrow.PrimitiveTypes.Int64},
		{Name: "data_type", Type: arrow.BinaryTypes.String},
		{Name: "is_nullable", Type: arrow.BinaryTypes.String},
	}, nil)

	// 创建 index_metadata 系统表的 schema
	indexMetadataSchema := arrow.NewSchema([]arrow.Field{
		{Name: "db_name", Type: arrow.BinaryTypes.String},
		{Name: "table_name", Type: arrow.BinaryTypes.String},
		{Name: "index_name", Type: arrow.BinaryTypes.String},
		{Name: "index_type", Type: arrow.BinaryTypes.String},
		{Name: "column_name", Type: arrow.BinaryTypes.String},
		{Name: "is_unique", Type: arrow.BinaryTypes.String},
	}, nil)

	// 创建 delta_log 系统表的 schema
	deltaLogSchema := arrow.NewSchema([]arrow.Field{
		{Name: "version", Type: arrow.PrimitiveTypes.Int64},
		{Name: "timestamp", Type: arrow.BinaryTypes.String},
		{Name: "operation", Type: arrow.BinaryTypes.String},
		{Name: "db_name", Type: arrow.BinaryTypes.String},
		{Name: "table_name", Type: arrow.BinaryTypes.String},
		{Name: "file_path", Type: arrow.BinaryTypes.String},
	}, nil)

	// 创建 table_files 系统表的 schema
	tableFilesSchema := arrow.NewSchema([]arrow.Field{
		{Name: "db_name", Type: arrow.BinaryTypes.String},
		{Name: "table_name", Type: arrow.BinaryTypes.String},
		{Name: "file_path", Type: arrow.BinaryTypes.String},
		{Name: "file_size", Type: arrow.PrimitiveTypes.Int64},
		{Name: "row_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "status", Type: arrow.BinaryTypes.String},
	}, nil)

	// 添加系统表到内存缓存
	c.tables["sys"]["db_metadata"] = &TableInfo{
		Database: "sys",
		Name:     "db_metadata",
		Schema:   dbMetadataSchema,
	}

	c.tables["sys"]["table_metadata"] = &TableInfo{
		Database: "sys",
		Name:     "table_metadata",
		Schema:   tableMetadataSchema,
	}

	c.tables["sys"]["columns_metadata"] = &TableInfo{
		Database: "sys",
		Name:     "columns_metadata",
		Schema:   columnsMetadataSchema,
	}

	c.tables["sys"]["index_metadata"] = &TableInfo{
		Database: "sys",
		Name:     "index_metadata",
		Schema:   indexMetadataSchema,
	}

	c.tables["sys"]["delta_log"] = &TableInfo{
		Database: "sys",
		Name:     "delta_log",
		Schema:   deltaLogSchema,
	}

	c.tables["sys"]["table_files"] = &TableInfo{
		Database: "sys",
		Name:     "table_files",
		Schema:   tableFilesSchema,
	}

	return nil
}

// CreateDatabase 通过SQL创建数据库
func (c *SimpleSQLCatalog) CreateDatabase(name string) error {
	logger.WithComponent("catalog").Info("Creating database",
		zap.String("database", name))

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
		sql := fmt.Sprintf("INSERT INTO sys.db_metadata (db_name) VALUES ('%s')", name)
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

	// v2.0: Metadata is persisted in Delta Log, no need for explicit persistence
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
		sql1 := fmt.Sprintf("DELETE FROM sys.table_metadata WHERE db_name = '%s'", name)
		c.sqlRunner.ExecuteSQL(sql1)

		// 删除数据库记录
		sql2 := fmt.Sprintf("DELETE FROM sys.db_metadata WHERE db_name = '%s'", name)
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
			INSERT INTO sys.table_metadata (db_name, table_name, chunk_count, schema_info)
			VALUES ('%s', '%s', %d, 'schema_placeholder')`,
			database, tableMeta.Table, tableMeta.ChunkCount)
		c.sqlRunner.ExecuteSQL(sql)
	}

	// v2.0: Persist schema to storage engine (Delta Log)
	// This is CRITICAL for schema recovery after restart
	if c.storageEngine != nil {
		err := c.storageEngine.CreateTable(database, tableMeta.Table, tableMeta.Schema)
		if err != nil {
			logger.WithComponent("catalog").Error("Failed to persist table to storage engine",
				zap.String("database", database),
				zap.String("table", tableMeta.Table),
				zap.Error(err))
			return fmt.Errorf("failed to persist table schema: %w", err)
		}
		logger.WithComponent("catalog").Info("Table schema persisted to storage engine",
			zap.String("database", database),
			zap.String("table", tableMeta.Table))
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
			FROM sys.table_metadata
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
		sql := fmt.Sprintf("SELECT table_name FROM sys.table_metadata WHERE db_name = '%s' ORDER BY table_name", database)
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

// DropTable 删除表
func (c *SimpleSQLCatalog) DropTable(database, table string) error {
	logger.WithComponent("catalog").Info("Dropping table",
		zap.String("database", database),
		zap.String("table", table))

	start := time.Now()
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查表是否存在
	if tables, exists := c.tables[database]; !exists || tables[table] == nil {
		return fmt.Errorf("table '%s.%s' does not exist", database, table)
	}

	// 使用SQL删除
	if c.sqlRunner != nil {
		sql := fmt.Sprintf("DELETE FROM sys.table_metadata WHERE db_name = '%s' AND table_name = '%s'", database, table)
		c.sqlRunner.ExecuteSQL(sql)
	}

	// v2.0: Persist table deletion to storage engine (marks all files as REMOVE in Delta Log)
	if c.storageEngine != nil {
		err := c.storageEngine.DropTable(database, table)
		if err != nil {
			logger.WithComponent("catalog").Error("Failed to persist table deletion to storage engine",
				zap.String("database", database),
				zap.String("table", table),
				zap.Error(err))
			return fmt.Errorf("failed to persist table deletion: %w", err)
		}
		logger.WithComponent("catalog").Info("Table deletion persisted to storage engine",
			zap.String("database", database),
			zap.String("table", table))
	}

	// 更新内存缓存
	delete(c.tables[database], table)

	logger.WithComponent("catalog").Info("Table dropped successfully",
		zap.String("table", table),
		zap.Duration("duration", time.Since(start)))

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
			UPDATE sys.table_metadata
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

	// v2.0: Persist index metadata to Delta Log (CRITICAL for recovery after restart)
	if c.storageEngine != nil {
		// 获取 Delta Log 接口
		type deltaLogProvider interface {
			GetDeltaLog() delta.LogInterface
		}

		provider, ok := c.storageEngine.(deltaLogProvider)
		if ok && provider.GetDeltaLog() != nil {
			tableID := fmt.Sprintf("%s.%s", indexMeta.Database, indexMeta.Table)

			// 构造索引元数据
			indexMetaMap := make(map[string]interface{})
			indexMetaMap["columns"] = strings.Join(indexMeta.Columns, ",")
			indexMetaMap["is_unique"] = fmt.Sprintf("%v", indexMeta.IsUnique)
			indexMetaMap["index_type"] = indexMeta.IndexType

			// 持久化索引元数据到 Delta Log
			err := provider.GetDeltaLog().AppendIndexMetadata(tableID, indexMeta.Name, indexMetaMap)
			if err != nil {
				logger.WithComponent("catalog").Error("Failed to persist index metadata to Delta Log",
					zap.String("database", indexMeta.Database),
					zap.String("table", indexMeta.Table),
					zap.String("index", indexMeta.Name),
					zap.Error(err))
				return fmt.Errorf("failed to persist index metadata: %w", err)
			}
			logger.WithComponent("catalog").Info("Index metadata persisted to Delta Log",
				zap.String("database", indexMeta.Database),
				zap.String("table", indexMeta.Table),
				zap.String("index", indexMeta.Name))
		}
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

	start := time.Now()
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查索引是否存在
	if c.indexes[database] == nil || c.indexes[database][table] == nil || c.indexes[database][table][indexName] == nil {
		return fmt.Errorf("index '%s' does not exist on table '%s.%s'", indexName, database, table)
	}

	// v2.0: Persist index deletion to Delta Log (CRITICAL for preventing deleted indexes from reappearing)
	if c.storageEngine != nil {
		// 获取 Delta Log 接口
		type deltaLogProvider interface {
			GetDeltaLog() delta.LogInterface
		}

		provider, ok := c.storageEngine.(deltaLogProvider)
		if ok && provider.GetDeltaLog() != nil {
			tableID := fmt.Sprintf("%s.%s", database, table)

			// 持久化索引删除操作到 Delta Log
			err := provider.GetDeltaLog().RemoveIndexMetadata(tableID, indexName)
			if err != nil {
				logger.WithComponent("catalog").Error("Failed to persist index deletion to Delta Log",
					zap.String("database", database),
					zap.String("table", table),
					zap.String("index", indexName),
					zap.Error(err))
				return fmt.Errorf("failed to persist index deletion: %w", err)
			}
			logger.WithComponent("catalog").Info("Index deletion persisted to Delta Log",
				zap.String("database", database),
				zap.String("table", table),
				zap.String("index", indexName))
		}
	}

	// 删除索引 (从内存中删除)
	delete(c.indexes[database][table], indexName)

	logger.WithComponent("catalog").Info("Index dropped successfully",
		zap.String("index", indexName),
		zap.Duration("duration", time.Since(start)))

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

// GetStorageEngine returns the v2.0 storage engine
func (c *SimpleSQLCatalog) GetStorageEngine() storage.StorageEngine {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.storageEngine
}

// SetStorageEngine sets the v2.0 storage engine
func (c *SimpleSQLCatalog) SetStorageEngine(engine storage.StorageEngine) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.storageEngine = engine
	logger.WithComponent("catalog").Info("Storage engine set successfully",
		zap.String("engine_type", fmt.Sprintf("%T", engine)))
}

// bootstrapSystemMetadata 使用SQL插入初始系统元数据
// SQL自举设计说明：
// 1. 系统表(sys.*)是特殊的"虚拟表"，其数据由DataManager动态生成
// 2. sys.db_metadata - 从databases map生成
// 3. sys.table_metadata - 从tables map生成
// 4. sys.columns_metadata - 从table schemas生成
// 5. sys.index_metadata - 从indexes map生成
// 6. sys.delta_log - 从Delta Log实时获取
// 7. sys.table_files - 从Delta Log快照获取
//
// 这种设计的优势：
// - 系统表数据始终是最新的（实时查询）
// - 避免了系统表元数据的循环依赖问题
// - 用户可以通过SQL查询所有元数据（SQL-first原则）
func (c *SimpleSQLCatalog) bootstrapSystemMetadata() error {
	logger.WithComponent("catalog").Info("Bootstrapping system metadata via SQL")

	// 系统表是虚拟表，无需初始化数据
	// 所有数据在查询时动态生成

	logger.WithComponent("catalog").Info("System metadata bootstrap completed (virtual system tables)")
	return nil
}

// loadMetadataFromSQL 从Delta Log恢复catalog状态
// SQL自举设计中，元数据的持久化由Delta Log负责：
// 1. CREATE DATABASE -> Delta Log记录METADATA操作
// 2. CREATE TABLE -> Delta Log记录METADATA操作 + schema信息
// 3. CREATE INDEX -> Delta Log记录METADATA操作 + index信息
// 4. 启动时，从Delta Log回放所有METADATA操作，恢复catalog状态
//
// 这样设计的好处：
// - 元数据变更和数据变更使用统一的事务日志（Delta Log）
// - 时间旅行查询可以看到历史的元数据状态
// - 分布式环境下，所有节点通过Delta Log同步元数据
func (c *SimpleSQLCatalog) loadMetadataFromSQL() error {
	logger.WithComponent("catalog").Info("Loading metadata from Delta Log")

	// 检查是否有 storage engine
	if c.storageEngine == nil {
		logger.WithComponent("catalog").Warn("No storage engine available, skipping metadata recovery")
		return nil
	}

	// 从 Storage Engine 获取所有表的列表（通过 Delta Log）
	// 使用类型断言获取 ParquetEngine 的 GetDeltaLog 方法
	type deltaLogProvider interface {
		GetDeltaLog() delta.LogInterface
	}

	provider, ok := c.storageEngine.(deltaLogProvider)
	if !ok {
		logger.WithComponent("catalog").Warn("Storage engine does not support Delta Log, skipping metadata recovery")
		return nil
	}

	deltaLog := provider.GetDeltaLog()
	if deltaLog == nil {
		logger.WithComponent("catalog").Warn("No Delta Log available, skipping metadata recovery")
		return nil
	}

	// 获取所有表
	tables := deltaLog.ListTables()
	logger.WithComponent("catalog").Info("Found tables in Delta Log",
		zap.Int("table_count", len(tables)))

	// 从表名中提取数据库和表
	// 表名格式: "database.table"
	for _, tableID := range tables {
		// 跳过系统表
		if strings.HasPrefix(tableID, "sys.") {
			continue
		}

		// 解析 database.table
		parts := strings.SplitN(tableID, ".", 2)
		if len(parts) != 2 {
			logger.WithComponent("catalog").Warn("Invalid table ID format",
				zap.String("table_id", tableID))
			continue
		}

		dbName := parts[0]
		tableName := parts[1]

		// 检查表是否已被删除（所有文件都被标记为 REMOVE）
		snapshot, err := deltaLog.GetSnapshot(tableID, -1)
		if err != nil {
			logger.WithComponent("catalog").Warn("Failed to get snapshot for table",
				zap.String("table", tableID),
				zap.Error(err))
			continue
		}

		// 如果表没有任何活跃文件且没有 schema，说明表已被删除，跳过恢复
		// 但是如果有 schema 说明表是刚创建的空表，应该恢复
		// 所以正确的逻辑是：如果表至少有一次 REMOVE 操作且当前没有文件，说明表被删除了
		entries := deltaLog.GetEntriesByTable(tableID)
		hasRemoveOp := false
		for _, entry := range entries {
			if entry.Operation == delta.OpRemove {
				hasRemoveOp = true
				break
			}
		}

		// 如果有 REMOVE 操作且没有活跃文件，说明表已被删除
		if hasRemoveOp && len(snapshot.Files) == 0 {
			logger.WithComponent("catalog").Debug("Skipping dropped table during recovery (all files removed)",
				zap.String("database", dbName),
				zap.String("table", tableName))
			continue
		}

		// 恢复数据库
		if _, exists := c.databases[dbName]; !exists {
			c.databases[dbName] = &DatabaseInfo{Name: dbName}
			c.tables[dbName] = make(map[string]*TableInfo)
			logger.WithComponent("catalog").Info("Recovered database from Delta Log",
				zap.String("database", dbName))
		}

		// 从 Storage Engine 获取表的 schema
		schema, err := c.storageEngine.GetTableSchema(dbName, tableName)
		if err != nil {
			logger.WithComponent("catalog").Warn("Failed to get table schema",
				zap.String("table", tableID),
				zap.Error(err))
			continue
		}

		// 恢复表
		c.tables[dbName][tableName] = &TableInfo{
			Database: dbName,
			Name:     tableName,
			Schema:   schema,
		}

		logger.WithComponent("catalog").Info("Recovered table from Delta Log",
			zap.String("database", dbName),
			zap.String("table", tableName),
			zap.Int("field_count", len(schema.Fields())))
	}

	// 恢复索引元数据（处理 CREATE 和 DROP 操作）
	indexCount := 0
	droppedIndexes := make(map[string]bool) // 跟踪已删除的索引: "db.table.index_name" -> true
	entries := deltaLog.GetAllEntries()

	// 第一遍：找出所有被删除的索引
	for _, entry := range entries {
		if entry.Operation == delta.OpMetadata && entry.IndexJSON != "" && entry.IndexOperation == "DROP" {
			parts := strings.SplitN(entry.TableID, ".", 2)
			if len(parts) == 2 {
				dbName := parts[0]
				tableName := parts[1]
				indexName := c.parseIndexName(entry.IndexJSON)
				if indexName != "" {
					key := fmt.Sprintf("%s.%s.%s", dbName, tableName, indexName)
					droppedIndexes[key] = true
					logger.WithComponent("catalog").Debug("Found DROP INDEX operation",
						zap.String("database", dbName),
						zap.String("table", tableName),
						zap.String("index", indexName))
				}
			}
		}
	}

	// 第二遍：恢复未被删除的索引
	for _, entry := range entries {
		// 只处理 METADATA 操作且包含索引信息的条目
		// 跳过 DROP 操作（已经在第一遍处理过）
		if entry.Operation == delta.OpMetadata && entry.IndexJSON != "" && entry.IndexOperation != "DROP" {
			// 解析表 ID
			parts := strings.SplitN(entry.TableID, ".", 2)
			if len(parts) != 2 {
				logger.WithComponent("catalog").Warn("Invalid table ID in index metadata",
					zap.String("table_id", entry.TableID))
				continue
			}

			dbName := parts[0]
			tableName := parts[1]

			// 解析 IndexJSON
			// 格式: {"index_name":"idx_name","table_id":"db.table","columns":"col1,col2","is_unique":"true","index_type":"btree"}
			indexName := c.parseIndexName(entry.IndexJSON)
			columns := c.parseIndexColumns(entry.IndexJSON)
			isUnique := c.parseIndexIsUnique(entry.IndexJSON)
			indexType := c.parseIndexType(entry.IndexJSON)

			if indexName == "" {
				logger.WithComponent("catalog").Warn("Failed to parse index name from IndexJSON",
					zap.String("table_id", entry.TableID),
					zap.String("index_json", entry.IndexJSON))
				continue
			}

			// 检查此索引是否被删除
			key := fmt.Sprintf("%s.%s.%s", dbName, tableName, indexName)
			if droppedIndexes[key] {
				logger.WithComponent("catalog").Debug("Skipping dropped index during recovery",
					zap.String("database", dbName),
					zap.String("table", tableName),
					zap.String("index", indexName))
				continue
			}

			// 初始化索引映射
			if c.indexes[dbName] == nil {
				c.indexes[dbName] = make(map[string]map[string]*IndexInfo)
			}
			if c.indexes[dbName][tableName] == nil {
				c.indexes[dbName][tableName] = make(map[string]*IndexInfo)
			}

			// 恢复索引
			c.indexes[dbName][tableName][indexName] = &IndexInfo{
				Database:  dbName,
				Table:     tableName,
				Name:      indexName,
				Columns:   columns,
				IsUnique:  isUnique,
				IndexType: indexType,
			}

			indexCount++
			logger.WithComponent("catalog").Info("Recovered index from Delta Log",
				zap.String("database", dbName),
				zap.String("table", tableName),
				zap.String("index", indexName),
				zap.Strings("columns", columns))
		}
	}

	logger.WithComponent("catalog").Info("Metadata loaded from Delta Log",
		zap.Int("database_count", len(c.databases)),
		zap.Int("total_tables", len(tables)),
		zap.Int("total_indexes", indexCount))
	return nil
}

// parseIndexName 从 IndexJSON 解析索引名称
func (c *SimpleSQLCatalog) parseIndexName(indexJSON string) string {
	// 格式: {"index_name":"idx_name",...}
	if start := strings.Index(indexJSON, `"index_name":"`); start != -1 {
		start += len(`"index_name":"`)
		if end := strings.Index(indexJSON[start:], `"`); end != -1 {
			return indexJSON[start : start+end]
		}
	}
	return ""
}

// parseIndexColumns 从 IndexJSON 解析索引列
func (c *SimpleSQLCatalog) parseIndexColumns(indexJSON string) []string {
	// 格式: {"columns":"col1,col2",...}
	if start := strings.Index(indexJSON, `"columns":"`); start != -1 {
		start += len(`"columns":"`)
		if end := strings.Index(indexJSON[start:], `"`); end != -1 {
			columnsStr := indexJSON[start : start+end]
			if columnsStr != "" {
				return strings.Split(columnsStr, ",")
			}
		}
	}
	return []string{}
}

// parseIndexIsUnique 从 IndexJSON 解析是否唯一索引
func (c *SimpleSQLCatalog) parseIndexIsUnique(indexJSON string) bool {
	// 格式: {"is_unique":"true",...}
	if start := strings.Index(indexJSON, `"is_unique":"`); start != -1 {
		start += len(`"is_unique":"`)
		if end := strings.Index(indexJSON[start:], `"`); end != -1 {
			valueStr := indexJSON[start : start+end]
			return valueStr == "true"
		}
	}
	return false
}

// parseIndexType 从 IndexJSON 解析索引类型
func (c *SimpleSQLCatalog) parseIndexType(indexJSON string) string {
	// 格式: {"index_type":"btree",...}
	if start := strings.Index(indexJSON, `"index_type":"`); start != -1 {
		start += len(`"index_type":"`)
		if end := strings.Index(indexJSON[start:], `"`); end != -1 {
			return indexJSON[start : start+end]
		}
	}
	return "btree" // 默认值
}

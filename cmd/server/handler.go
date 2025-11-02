package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/yyun543/minidb/internal/cache"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/config"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/logger"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/statistics"
	"github.com/yyun543/minidb/internal/storage"
	"go.uber.org/zap"
)

// catalogAdapter 适配catalog到optimizer的CatalogInterface
// 解决类型不匹配问题: catalog.TableMeta vs optimizer.TableMeta
type catalogAdapter struct {
	cat *catalog.Catalog
}

// GetTable 实现optimizer.CatalogInterface接口
func (a *catalogAdapter) GetTable(database, table string) (optimizer.TableMeta, error) {
	// 调用catalog的GetTable
	catMeta, err := a.cat.GetTable(database, table)
	if err != nil {
		return optimizer.TableMeta{}, err
	}

	// 转换类型
	return optimizer.TableMeta{
		Database: catMeta.Database,
		Table:    catMeta.Table,
		Schema:   catMeta.Schema,
	}, nil
}

// CatalogSQLAdapter 为catalog提供SQL执行能力的适配器
// 这个适配器直接操作系统表存储，避免循环依赖
type CatalogSQLAdapter struct {
	storageEngine storage.StorageEngine
}

// NewCatalogSQLAdapter 创建catalog SQL适配器 (v2.0)
func NewCatalogSQLAdapter(cat *catalog.Catalog, exec *executor.BaseExecutor, vectorizedExec *executor.VectorizedExecutor) *CatalogSQLAdapter {
	return &CatalogSQLAdapter{
		storageEngine: cat.GetStorageEngine(),
	}
}

// ExecuteSQL 执行SQL语句（专门用于系统表操作）
func (adapter *CatalogSQLAdapter) ExecuteSQL(sql string) (arrow.Record, error) {
	// 对于catalog的SQL操作，我们直接操作存储层
	// 这避免了执行器的循环依赖问题

	// 解析SQL以确定操作类型
	upperSQL := strings.ToUpper(strings.TrimSpace(sql))

	if strings.HasPrefix(upperSQL, "INSERT INTO SYS.SCHEMATA") {
		// 插入数据库记录 - 直接操作存储
		return adapter.handleInsertSchemata(sql)
	} else if strings.HasPrefix(upperSQL, "INSERT INTO SYS.TABLE_CATALOG") {
		// 插入表记录 - 直接操作存储
		return adapter.handleInsertTableCatalog(sql)
	} else if strings.HasPrefix(upperSQL, "DELETE FROM SYS.SCHEMATA") ||
		strings.HasPrefix(upperSQL, "DELETE FROM SYS.TABLE_CATALOG") ||
		strings.HasPrefix(upperSQL, "UPDATE SYS.TABLE_CATALOG") {
		// DELETE和UPDATE操作暂时跳过，因为catalog已经通过其他方式管理
		return nil, nil
	} else if strings.HasPrefix(upperSQL, "SELECT") {
		// SELECT操作暂时返回空结果
		return nil, nil
	}

	return nil, nil
}

// handleInsertSchemata 处理插入schemata表的操作
func (adapter *CatalogSQLAdapter) handleInsertSchemata(sql string) (arrow.Record, error) {
	// 简单解析: INSERT INTO sys.schemata (schema_name) VALUES ('database_name')
	// 提取数据库名称
	startIdx := strings.Index(strings.ToUpper(sql), "VALUES")
	if startIdx == -1 {
		return nil, fmt.Errorf("invalid INSERT syntax")
	}

	valuesStr := sql[startIdx+6:] // 跳过 "VALUES"
	valuesStr = strings.TrimSpace(valuesStr)
	valuesStr = strings.Trim(valuesStr, "()'\"")
	dbName := strings.TrimSpace(valuesStr)

	// 直接向系统表存储数据库信息
	return adapter.insertDatabaseRecord(dbName)
}

// handleInsertTableCatalog 处理插入table_catalog表的操作
func (adapter *CatalogSQLAdapter) handleInsertTableCatalog(sql string) (arrow.Record, error) {
	// 简单解析: INSERT INTO sys.table_catalog (table_schema, table_name, ...) VALUES (...)
	// 暂时跳过，因为表记录由catalog直接管理
	return nil, nil
}

// insertDatabaseRecord 向系统表插入数据库记录
// v2.0: Database creation is handled by StorageEngine.CreateDatabase()
func (adapter *CatalogSQLAdapter) insertDatabaseRecord(dbName string) (arrow.Record, error) {
	// In v2.0, database metadata is managed by Delta Log
	// This is a stub for compatibility
	return nil, nil
}

// QueryHandler 处理SQL查询请求 (v2.0)
type QueryHandler struct {
	catalog                *catalog.Catalog
	executor               *executor.BaseExecutor
	vectorizedExecutor     *executor.VectorizedExecutor
	sessionManager         *session.SessionManager
	statisticsManager      *statistics.StatisticsManager
	storageEngine          storage.StorageEngine // v2.0: Using new StorageEngine interface
	queryCache             *cache.QueryCache     // P1.1: Query result cache
	useVectorizedExecution bool
}

// NewQueryHandler 创建新的查询处理器 (v2.0 with ParquetEngine)
func NewQueryHandler() (*QueryHandler, error) {
	// 1. 创建 v2.0 Parquet 存储引擎
	storageEngine, err := storage.NewParquetEngine("./minidb_data")
	if err != nil {
		return nil, fmt.Errorf("Failed to create Parquet storage engine: %v", err)
	}
	if err := storageEngine.Open(); err != nil {
		return nil, fmt.Errorf("Failed to open Parquet storage engine: %v", err)
	}

	// 2. 创建会话管理器
	sessMgr, err := session.NewSessionManager()
	if err != nil {
		return nil, fmt.Errorf("Failure to create session manager: %v", err)
	}

	// 3. 创建 catalog (v2.0 - no engine parameter)
	cat := catalog.NewCatalog()
	cat.SetStorageEngine(storageEngine)

	// 4. 创建统计信息管理器
	statsMgr := statistics.NewStatisticsManager()

	// 5. 创建共享的数据管理器
	dataManager := executor.NewDataManager(cat)

	// 6. 创建执行器（常规和向量化）
	exec := executor.NewExecutorWithDataManager(cat, dataManager)
	vectorizedExec := executor.NewVectorizedExecutorWithDataManager(cat, dataManager, statsMgr)

	// 7. 创建catalog的SQL执行器适配器
	sqlAdapter := NewCatalogSQLAdapter(cat, exec, vectorizedExec)
	cat.SetSQLRunner(sqlAdapter)

	// 8. 初始化catalog
	if err := cat.Init(); err != nil {
		return nil, fmt.Errorf("Failed to initialize catalog: %v", err)
	}

	// 9. 创建查询缓存 (P1.1)
	queryCacheConfig := cache.DefaultCacheConfig()
	queryCache := cache.NewQueryCache(queryCacheConfig)

	// 10. 创建QueryHandler
	handler := &QueryHandler{
		catalog:                cat,
		executor:               exec,
		vectorizedExecutor:     vectorizedExec,
		sessionManager:         sessMgr,
		statisticsManager:      statsMgr,
		storageEngine:          storageEngine,
		queryCache:             queryCache, // P1.1: Initialize query cache
		useVectorizedExecution: true,       // 默认启用向量化执行
	}

	// 11. 启动后台服务
	go handler.startBackgroundServices()

	return handler, nil
}

// NewQueryHandlerWithConfig 创建带配置的查询处理器 (P1.3)
func NewQueryHandlerWithConfig(cfg *config.Config) (*QueryHandler, error) {
	// 1. 创建 Parquet 存储引擎 (使用配置的路径)
	storageEngine, err := storage.NewParquetEngine(cfg.Storage.DataDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to create Parquet storage engine: %v", err)
	}
	if err := storageEngine.Open(); err != nil {
		return nil, fmt.Errorf("Failed to open Parquet storage engine: %v", err)
	}

	// 2. 创建会话管理器
	sessMgr, err := session.NewSessionManager()
	if err != nil {
		return nil, fmt.Errorf("Failure to create session manager: %v", err)
	}

	// 3. 创建 catalog
	cat := catalog.NewCatalog()
	cat.SetStorageEngine(storageEngine)

	// 4. 创建统计信息管理器
	statsMgr := statistics.NewStatisticsManager()

	// 5. 创建共享的数据管理器
	dataManager := executor.NewDataManager(cat)

	// 6. 创建执行器（常规和向量化）
	exec := executor.NewExecutorWithDataManager(cat, dataManager)
	vectorizedExec := executor.NewVectorizedExecutorWithDataManager(cat, dataManager, statsMgr)

	// 7. 创建catalog的SQL执行器适配器
	sqlAdapter := NewCatalogSQLAdapter(cat, exec, vectorizedExec)
	cat.SetSQLRunner(sqlAdapter)

	// 8. 初始化catalog
	if err := cat.Init(); err != nil {
		return nil, fmt.Errorf("Failed to initialize catalog: %v", err)
	}

	// 9. 创建查询缓存 (P1.1) - 使用配置
	var queryCache *cache.QueryCache
	if cfg.QueryCache.Enabled {
		queryCacheConfig := &cache.CacheConfig{
			MaxSizeMB:        int64(cfg.QueryCache.MaxSizeMB),
			TTL:              cfg.QueryCache.GetTTL(),
			MinComputeTimeMs: int64(cfg.QueryCache.MinComputeTimeMS),
		}
		queryCache = cache.NewQueryCache(queryCacheConfig)
		logger.WithComponent("handler").Info("Query cache enabled from config",
			zap.Int("max_size_mb", cfg.QueryCache.MaxSizeMB),
			zap.Duration("ttl", queryCacheConfig.TTL))
	} else {
		logger.WithComponent("handler").Info("Query cache disabled by config")
	}

	// 10. 创建QueryHandler
	handler := &QueryHandler{
		catalog:                cat,
		executor:               exec,
		vectorizedExecutor:     vectorizedExec,
		sessionManager:         sessMgr,
		statisticsManager:      statsMgr,
		storageEngine:          storageEngine,
		queryCache:             queryCache,
		useVectorizedExecution: cfg.Vectorization.Enabled,
	}

	logger.WithComponent("handler").Info("QueryHandler created with configuration",
		zap.String("storage_dir", cfg.Storage.DataDir),
		zap.Bool("vectorization", cfg.Vectorization.Enabled),
		zap.Bool("cache_enabled", cfg.QueryCache.Enabled))

	// 11. 启动后台服务
	go handler.startBackgroundServices()

	return handler, nil
}

// startBackgroundServices 启动后台服务
func (h *QueryHandler) startBackgroundServices() {
	// 定期清理过期会话
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			h.sessionManager.CleanupExpiredSessions(2 * time.Hour)
		}
	}()

	// 定期更新统计信息
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			h.updateStatistics()
		}
	}()
}

// updateStatistics 更新统计信息
func (h *QueryHandler) updateStatistics() {
	// 获取所有数据库
	databases, err := h.catalog.GetAllDatabases()
	if err != nil {
		// 记录错误但不中断服务
		return
	}

	// 遍历每个数据库，更新其中表的统计信息
	for _, dbName := range databases {
		tables, err := h.catalog.GetAllTables(dbName)
		if err != nil {
			continue // 跳过有错误的数据库
		}

		// 为每个表更新统计信息
		for _, tableName := range tables {
			// 简化实现：仅记录统计信息更新请求
			// 在生产环境中，这里应该获取表schema和数据批次
			// 当前跳过以避免复杂的依赖关系
			_ = tableName // 避免未使用变量警告
		}
	}
}

// HandleQuery 处理单个SQL查询
func (h *QueryHandler) HandleQuery(sessionID int64, sql string) (string, error) {
	// 获取或创建会话
	sess, ok := h.sessionManager.GetSession(sessionID)
	if !ok {
		return "", fmt.Errorf("Invalid session ID: %d", sessionID)
	}

	// 1. 解析SQL
	ast, err := parser.Parse(sql)
	if err != nil {
		return "", fmt.Errorf("parsing error: %v", err)
	}

	// 2. 处理特殊命令
	if result, handled := h.handleSpecialCommands(ast, sess); handled {
		return result, nil
	}

	// 3. 优化查询
	opt := optimizer.NewOptimizer()
	// P0.2: 设置catalog和当前数据库,使SELECT *展开优化生效
	opt.SetCatalog(&catalogAdapter{cat: h.catalog})
	opt.SetCurrentDatabase(sess.CurrentDB)
	plan, err := opt.Optimize(ast)
	if err != nil {
		return "", fmt.Errorf("optimization error: %v", err)
	}

	// 检查plan是否为nil
	if plan == nil {
		return "", fmt.Errorf("optimizer returned nil plan")
	}

	// P1.1: Check if this is a cacheable SELECT query
	isCacheable := h.isCacheableQuery(plan, ast)
	var cacheKey cache.CacheKey
	if isCacheable {
		// Build cache key
		currentDB := sess.CurrentDB
		if currentDB == "" {
			currentDB = "default"
		}
		cacheKey = cache.CacheKey{
			SQL:          sql,
			Database:     currentDB,
			TableVersion: 1, // TODO: Get actual table version from Delta Log
		}

		// Check cache
		if cached, found := h.queryCache.Get(cacheKey); found {
			// Cache hit! Return cached result
			return h.formatExecutionResult(cached), nil
		}
	}

	// 4. 执行查询（选择向量化或常规执行器）
	startTime := time.Now()
	var result interface{}
	if h.useVectorizedExecution && h.isVectorizableQuery(plan) {
		// 使用向量化执行器
		vectorizedResult, err := h.vectorizedExecutor.Execute(plan, sess)
		if err != nil {
			// 检查是否需要回退到常规执行器
			if strings.Contains(err.Error(), "fallback to regular executor") ||
				strings.Contains(err.Error(), "not supported in vectorized execution") ||
				strings.Contains(err.Error(), "unsupported predicate type") {
				// 自动回退到常规执行器
				logger.WithComponent("handler").Debug("Falling back to regular executor",
					zap.String("reason", err.Error()))
				regularResult, err := h.executor.Execute(plan, sess)
				if err != nil {
					return "", fmt.Errorf("execution error: %v", err)
				}
				result = regularResult
			} else {
				return "", fmt.Errorf("vectorized execution error: %v", err)
			}
		} else {
			result = vectorizedResult
		}
	} else {
		// 使用常规执行器
		regularResult, err := h.executor.Execute(plan, sess)
		if err != nil {
			return "", fmt.Errorf("execution error: %v", err)
		}
		result = regularResult
	}
	computeTime := time.Since(startTime)

	// P1.1: Cache the result if this is a cacheable query
	if isCacheable {
		if resultSet, ok := result.(*executor.ResultSet); ok {
			h.queryCache.Put(cacheKey, resultSet, computeTime)
		}
	}

	// P1.1: Invalidate cache for DML operations
	h.invalidateCacheForDML(plan, sess)

	// 5. 格式化结果
	return h.formatExecutionResult(result), nil
}

// handleSpecialCommands 处理特殊命令
func (h *QueryHandler) handleSpecialCommands(ast interface{}, sess *session.Session) (string, bool) {
	// 处理USE语句
	if useStmt, ok := ast.(*parser.UseStmt); ok {
		sess.CurrentDB = useStmt.Database
		return fmt.Sprintf("Switched to database: %s", useStmt.Database), true
	}

	// 处理SHOW DATABASES命令
	if _, ok := ast.(*parser.ShowDatabasesStmt); ok {
		databases, err := h.catalog.GetAllDatabases()
		if err != nil {
			return fmt.Sprintf("Error: %v", err), true
		}

		var sb strings.Builder
		sb.WriteString("+----------------+\n")
		sb.WriteString("| Database       |\n")
		sb.WriteString("+----------------+\n")

		if len(databases) == 0 {
			sb.WriteString("| (no databases) |\n")
		} else {
			for _, db := range databases {
				sb.WriteString(fmt.Sprintf("| %-14s |\n", db))
			}
		}
		sb.WriteString("+----------------+\n")
		return sb.String(), true
	}

	// 处理SHOW TABLES命令
	if _, ok := ast.(*parser.ShowTablesStmt); ok {
		// 获取当前数据库的所有表
		currentDB := sess.CurrentDB
		if currentDB == "" {
			currentDB = "default"
		}

		// 通过catalog获取表列表
		tables, err := h.catalog.GetAllTables(currentDB)
		if err != nil {
			return fmt.Sprintf("Error: %v", err), true
		}

		var sb strings.Builder
		sb.WriteString("+----------------+\n")
		sb.WriteString("| Tables         |\n")
		sb.WriteString("+----------------+\n")

		if len(tables) == 0 {
			sb.WriteString("| (no tables)    |\n")
		} else {
			for _, table := range tables {
				sb.WriteString(fmt.Sprintf("| %-14s |\n", table))
			}
		}
		sb.WriteString("+----------------+\n")
		return sb.String(), true
	}

	// 处理SHOW INDEXES命令
	if showIndexesStmt, ok := ast.(*parser.ShowIndexesStmt); ok {
		currentDB := sess.CurrentDB
		if currentDB == "" {
			currentDB = "default"
		}

		// 通过catalog获取索引列表
		indexes, err := h.catalog.GetAllIndexes(currentDB, showIndexesStmt.Table)
		if err != nil {
			return fmt.Sprintf("Error: %v", err), true
		}

		var sb strings.Builder
		sb.WriteString("+----------------+----------------+----------------+----------------+\n")
		sb.WriteString("| Index Name     | Table          | Columns        | Unique         |\n")
		sb.WriteString("+----------------+----------------+----------------+----------------+\n")

		if len(indexes) == 0 {
			sb.WriteString("| (no indexes)                                                     |\n")
		} else {
			for _, idx := range indexes {
				columnsStr := strings.Join(idx.Columns, ",")
				uniqueStr := "NO"
				if idx.IsUnique {
					uniqueStr = "YES"
				}
				sb.WriteString(fmt.Sprintf("| %-14s | %-14s | %-14s | %-14s |\n",
					idx.Name, idx.Table, columnsStr, uniqueStr))
			}
		}
		sb.WriteString("+----------------+----------------+----------------+----------------+\n")
		return sb.String(), true
	}

	// 处理EXPLAIN命令
	if explainStmt, ok := ast.(*parser.ExplainStmt); ok {
		// 优化被解释的查询
		opt := optimizer.NewOptimizer()
		plan, err := opt.Optimize(explainStmt.Query)
		if err != nil {
			return fmt.Sprintf("Error optimizing query: %v", err), true
		}
		return h.formatQueryPlan(plan), true
	}

	return "", false
}

// isVectorizableQuery 判断查询是否适合向量化执行
func (h *QueryHandler) isVectorizableQuery(plan *optimizer.Plan) bool {
	// 递归检查计划树，确定是否所有操作都支持向量化
	return h.checkPlanVectorizable(plan)
}

// isCacheableQuery determines if a query result should be cached (P1.1)
// Only SELECT queries are cacheable; DML operations should invalidate cache
func (h *QueryHandler) isCacheableQuery(plan *optimizer.Plan, ast interface{}) bool {
	if plan == nil {
		return false
	}

	// Only cache SELECT queries (read-only)
	switch plan.Type {
	case optimizer.SelectPlan, optimizer.TableScanPlan, optimizer.FilterPlan,
		optimizer.ProjectionPlan, optimizer.GroupPlan, optimizer.OrderPlan,
		optimizer.JoinPlan, optimizer.LimitPlan, optimizer.HavingPlan:
		// These are read-only query operations - cacheable
		return true
	case optimizer.InsertPlan, optimizer.UpdatePlan, optimizer.DeletePlan,
		optimizer.CreateTablePlan, optimizer.CreateDatabasePlan,
		optimizer.DropTablePlan, optimizer.DropDatabasePlan:
		// DML/DDL operations - not cacheable
		return false
	case optimizer.ShowPlan:
		// System queries - not cacheable (metadata may change)
		return false
	default:
		// Unknown plan types - don't cache to be safe
		return false
	}
}

// invalidateCacheForDML invalidates cache entries when tables are modified (P1.1)
func (h *QueryHandler) invalidateCacheForDML(plan *optimizer.Plan, sess *session.Session) {
	if plan == nil {
		return
	}

	currentDB := sess.CurrentDB
	if currentDB == "" {
		currentDB = "default"
	}

	// Check if this is a DML or DDL operation that modifies tables
	switch plan.Type {
	case optimizer.InsertPlan:
		props, ok := plan.Properties.(*optimizer.InsertProperties)
		if ok && props.Table != "" {
			h.queryCache.InvalidateTable(currentDB, props.Table)
		}
	case optimizer.UpdatePlan:
		props, ok := plan.Properties.(*optimizer.UpdateProperties)
		if ok && props.Table != "" {
			h.queryCache.InvalidateTable(currentDB, props.Table)
		}
	case optimizer.DeletePlan:
		props, ok := plan.Properties.(*optimizer.DeleteProperties)
		if ok && props.Table != "" {
			h.queryCache.InvalidateTable(currentDB, props.Table)
		}
	case optimizer.DropTablePlan:
		props, ok := plan.Properties.(*optimizer.DropTableProperties)
		if ok && props.Table != "" {
			h.queryCache.InvalidateTable(currentDB, props.Table)
		}
	case optimizer.CreateTablePlan:
		// New table creation - no need to invalidate (no cached queries yet)
	case optimizer.DropDatabasePlan, optimizer.CreateDatabasePlan:
		// Database-level operations - clear entire cache to be safe
		h.queryCache.Clear()
	}
}

// checkPlanVectorizable 递归检查计划节点是否支持向量化
func (h *QueryHandler) checkPlanVectorizable(plan *optimizer.Plan) bool {
	// 检查plan是否为nil
	if plan == nil {
		return false
	}

	// DDL和工具命令不适合向量化
	switch plan.Type {
	case optimizer.CreateTablePlan, optimizer.CreateDatabasePlan,
		optimizer.DropTablePlan, optimizer.DropDatabasePlan,
		optimizer.ShowPlan:
		return false
	case optimizer.JoinPlan:
		// JOIN operations work fine with regular executor
		// Recursively check if children are vectorizable
		// For now, use regular executor for JOINs
		return false
	case optimizer.GroupPlan, optimizer.OrderPlan,
		optimizer.HavingPlan, optimizer.LimitPlan:
		// 复杂操作暂时不支持向量化，回退到常规执行器
		return false
	case optimizer.FilterPlan:
		// 检查过滤条件是否包含不支持的表达式
		props := plan.Properties.(*optimizer.FilterProperties)
		if !h.checkExpressionVectorizable(props.Condition) {
			return false
		}
	case optimizer.SelectPlan, optimizer.TableScanPlan, optimizer.ProjectionPlan:
		// 基本操作支持向量化
		break
	case optimizer.InsertPlan, optimizer.UpdatePlan, optimizer.DeletePlan:
		// DML操作支持向量化
		return true
	default:
		// 未知计划类型不支持向量化
		return false
	}

	// 递归检查子计划
	for _, child := range plan.Children {
		if !h.checkPlanVectorizable(child) {
			return false
		}
	}

	return true
}

// checkExpressionVectorizable 检查表达式是否支持向量化
func (h *QueryHandler) checkExpressionVectorizable(expr optimizer.Expression) bool {
	if expr == nil {
		return true
	}

	switch e := expr.(type) {
	case *optimizer.BinaryExpression:
		// 检查LIKE表达式
		if e.Operator == "LIKE" || e.Operator == "NOT LIKE" {
			return false
		}
		// 检查BETWEEN表达式 (如果operator是 BETWEEN, NOT BETWEEN)
		if e.Operator == "BETWEEN" || e.Operator == "NOT BETWEEN" {
			return false
		}
		// 检查IN表达式 (如果operator是 IN, NOT IN)
		if e.Operator == "IN" || e.Operator == "NOT IN" {
			return false
		}

		// 递归检查左右子表达式
		return h.checkExpressionVectorizable(e.Left) && h.checkExpressionVectorizable(e.Right)
	default:
		// 其他表达式类型（列引用、字面量等）支持向量化
		return true
	}
}

// formatQueryPlan 格式化查询计划
func (h *QueryHandler) formatQueryPlan(plan *optimizer.Plan) string {
	var sb strings.Builder
	sb.WriteString("Query Execution Plan:\n")
	sb.WriteString("--------------------\n")
	h.formatPlanNode(plan, &sb, 0)
	return sb.String()
}

// formatPlanNode 递归格式化计划节点
func (h *QueryHandler) formatPlanNode(plan *optimizer.Plan, sb *strings.Builder, depth int) {
	indent := strings.Repeat("  ", depth)
	sb.WriteString(fmt.Sprintf("%s%s\n", indent, plan.Type.String()))

	for _, child := range plan.Children {
		h.formatPlanNode(child, sb, depth+1)
	}
}

// formatExecutionResult 格式化执行结果
func (h *QueryHandler) formatExecutionResult(result interface{}) string {
	// 处理常规执行器结果
	if regularResult, ok := result.(*executor.ResultSet); ok {
		return h.formatRegularResult(regularResult)
	}

	// 处理向量化执行器结果
	if vectorizedResult, ok := result.(*executor.VectorizedResultSet); ok {
		return h.formatVectorizedResult(vectorizedResult)
	}

	return "OK"
}

// formatRegularResult 格式化常规结果
func (h *QueryHandler) formatRegularResult(result *executor.ResultSet) string {
	if result == nil {
		return "OK"
	}

	var sb strings.Builder
	headers := result.Headers
	rowCount := 0

	// 如果headers只有一个且为"status"，这是DDL/DML操作，直接返回OK
	if len(headers) == 1 && headers[0] == "status" {
		return "OK"
	}

	// 检查是否有数据批次
	batches := result.Batches()
	if len(batches) == 0 {
		return "Empty set"
	}

	// 统计总行数
	totalRows := 0
	for _, batch := range batches {
		if batch != nil {
			totalRows += int(batch.NumRows())
		}
	}

	if totalRows == 0 {
		return "Empty set"
	}

	// 写入列名
	sb.WriteString("|")
	for _, header := range headers {
		sb.WriteString(fmt.Sprintf(" %-15s |", header))
	}
	sb.WriteString("\n")

	// 写入分隔线
	sb.WriteString("+")
	for range headers {
		sb.WriteString("-----------------+")
	}
	sb.WriteString("\n")

	// 写入数据行
	for _, batch := range batches {
		if batch == nil {
			continue
		}
		record := batch.Record()
		for i := int64(0); i < record.NumRows(); i++ {
			sb.WriteString("|")
			for j := int64(0); j < record.NumCols(); j++ {
				column := record.Column(int(j))
				value := h.getColumnValue(column, int(i))
				sb.WriteString(fmt.Sprintf(" %-15v |", value))
			}
			sb.WriteString("\n")
			rowCount++
		}
	}

	if rowCount == 0 {
		return "Empty set"
	}

	sb.WriteString(fmt.Sprintf("%d rows in set\n", rowCount))
	return sb.String()
}

// formatVectorizedResult 格式化向量化结果
func (h *QueryHandler) formatVectorizedResult(result *executor.VectorizedResultSet) string {
	if result == nil {
		return "OK"
	}

	// 如果headers只有一个且为"status"，这是DDL/DML操作，直接返回OK
	if len(result.Headers) == 1 && result.Headers[0] == "status" {
		return "OK"
	}

	if len(result.Batches) == 0 {
		return "Empty set"
	}

	var sb strings.Builder
	headers := result.Headers
	totalRows := int64(0)

	// 写入列名
	sb.WriteString("|")
	for _, header := range headers {
		sb.WriteString(fmt.Sprintf(" %-15s |", header))
	}
	sb.WriteString("\n")

	// 写入分隔线
	sb.WriteString("+")
	for range headers {
		sb.WriteString("-----------------+")
	}
	sb.WriteString("\n")

	// 处理所有批次
	for _, batch := range result.Batches {
		if batch == nil {
			continue
		}

		record := batch.ToRecord()
		defer record.Release()

		// 写入数据行
		for i := int64(0); i < record.NumRows(); i++ {
			sb.WriteString("|")
			for j := int64(0); j < record.NumCols(); j++ {
				column := record.Column(int(j))
				value := h.getColumnValue(column, int(i))
				sb.WriteString(fmt.Sprintf(" %-15v |", value))
			}
			sb.WriteString("\n")
			totalRows++
		}
	}

	if totalRows == 0 {
		return "Empty set"
	}

	sb.WriteString(fmt.Sprintf("%d rows in set\n", totalRows))
	return sb.String()
}

// getColumnValue 从 Arrow 列中获取指定行的值
func (h *QueryHandler) getColumnValue(column arrow.Array, rowIdx int) interface{} {
	logger.WithComponent("handler").Info("getColumnValue called",
		zap.String("column_type", fmt.Sprintf("%T", column)),
		zap.Int("rowIdx", rowIdx),
		zap.Bool("is_null", column.IsNull(rowIdx)))

	if column.IsNull(rowIdx) {
		logger.WithComponent("handler").Info("getColumnValue: returning NULL")
		return "NULL"
	}

	switch col := column.(type) {
	case *array.Int64:
		value := col.Value(rowIdx)
		logger.WithComponent("handler").Info("getColumnValue: Int64", zap.Int64("value", value))
		return value
	case *array.Int32:
		value := col.Value(rowIdx)
		logger.WithComponent("handler").Info("getColumnValue: Int32", zap.Int32("value", value))
		return value
	case *array.Float64:
		value := col.Value(rowIdx)
		logger.WithComponent("handler").Info("getColumnValue: Float64", zap.Float64("value", value))
		return value
	case *array.Float32:
		value := col.Value(rowIdx)
		logger.WithComponent("handler").Info("getColumnValue: Float32", zap.Float32("value", value))
		return value
	case *array.String:
		value := col.Value(rowIdx)
		logger.WithComponent("handler").Info("getColumnValue: String", zap.String("value", value))
		return value
	case *array.Boolean:
		return col.Value(rowIdx)
	default:
		return "?"
	}
}

// Close 关闭查询处理器
func (h *QueryHandler) Close() error {
	if h.storageEngine != nil {
		return h.storageEngine.Close()
	}
	return nil
}

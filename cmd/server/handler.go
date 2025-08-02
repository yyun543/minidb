package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/statistics"
	"github.com/yyun543/minidb/internal/storage"
)

// QueryHandler 处理SQL查询请求
type QueryHandler struct {
	catalog                *catalog.Catalog
	executor               *executor.BaseExecutor
	vectorizedExecutor     *executor.VectorizedExecutor
	sessionManager         *session.SessionManager
	statisticsManager      *statistics.StatisticsManager
	storageEngine          storage.Engine
	useVectorizedExecution bool
}

// NewQueryHandler 创建新的查询处理器
func NewQueryHandler() (*QueryHandler, error) {
	// 1. 创建存储引擎
	storageEngine, err := storage.NewMemTable("minidb.wal")
	if err != nil {
		return nil, fmt.Errorf("Failed to create storage engine: %v", err)
	}
	if err := storageEngine.Open(); err != nil {
		return nil, fmt.Errorf("Failed to open storage engine: %v", err)
	}

	// 2. 创建catalog（使用同一个存储引擎）
	cat := catalog.NewCatalog(storageEngine)

	// 初始化catalog
	if err := cat.Init(); err != nil {
		return nil, fmt.Errorf("Failed to initialize catalog: %v", err)
	}

	// 3. 创建统计信息管理器
	statsMgr := statistics.NewStatisticsManager()

	// 4. 创建共享的数据管理器
	dataManager := executor.NewDataManager(cat)

	// 5. 创建执行器（常规和向量化）
	exec := executor.NewExecutorWithDataManager(cat, dataManager)
	vectorizedExec := executor.NewVectorizedExecutorWithDataManager(cat, dataManager, statsMgr)

	// 5. 创建会话管理器
	sessMgr, err := session.NewSessionManager()
	if err != nil {
		return nil, fmt.Errorf("Failure to create session manager: %v", err)
	}

	// 6. 创建QueryHandler
	handler := &QueryHandler{
		catalog:                cat,
		executor:               exec,
		vectorizedExecutor:     vectorizedExec,
		sessionManager:         sessMgr,
		statisticsManager:      statsMgr,
		storageEngine:          storageEngine,
		useVectorizedExecution: true, // 默认启用向量化执行
	}

	// 7. 启动后台服务
	go handler.startBackgroundServices()

	return handler, nil
}

// getTablesInDatabase 获取指定数据库中的所有表
func (h *QueryHandler) getTablesInDatabase(dbName string) ([]string, error) {
	// 通过存储引擎直接查询sys_tables
	engine := h.catalog.GetEngine()
	keyManager := storage.NewKeyManager()

	key := keyManager.TableChunkKey("system", "sys_tables", 0)
	record, err := engine.Get(key)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return []string{}, nil
	}
	defer record.Release()

	var tables []string
	dbNameCol := record.Column(2).(*array.String)
	tableNameCol := record.Column(3).(*array.String)

	for i := int64(0); i < record.NumRows(); i++ {
		if dbNameCol.Value(int(i)) == dbName {
			tables = append(tables, tableNameCol.Value(int(i)))
		}
	}

	return tables, nil
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
	// TODO: 实现获取所有表的方法
	// 当前简化实现：不进行统计信息更新
	// 在生产环境中应该从catalog获取所有表并更新统计信息
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
	plan, err := opt.Optimize(ast)
	if err != nil {
		return "", fmt.Errorf("optimization error: %v", err)
	}

	// 4. 执行查询（选择向量化或常规执行器）
	var result interface{}
	if h.useVectorizedExecution && h.isVectorizableQuery(plan) {
		// 使用向量化执行器
		// fmt.Printf("DEBUG: Using vectorized executor for plan type: %v\n", plan.Type)
		vectorizedResult, err := h.vectorizedExecutor.Execute(plan, sess)
		if err != nil {
			return "", fmt.Errorf("vectorized execution error: %v", err)
		}
		result = vectorizedResult
	} else {
		// 使用常规执行器
		// fmt.Printf("DEBUG: Using regular executor for plan type: %v\n", plan.Type)
		regularResult, err := h.executor.Execute(plan, sess)
		if err != nil {
			return "", fmt.Errorf("execution error: %v", err)
		}
		result = regularResult
	}

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
				sb.WriteString(fmt.Sprintf("| %-14s |\n", db.Name))
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

		// 通过从系统表查询获取表列表
		tables, err := h.getTablesInDatabase(currentDB)
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

// checkPlanVectorizable 递归检查计划节点是否支持向量化
func (h *QueryHandler) checkPlanVectorizable(plan *optimizer.Plan) bool {
	// DDL和工具命令不适合向量化
	switch plan.Type {
	case optimizer.CreateTablePlan, optimizer.CreateDatabasePlan,
		optimizer.DropTablePlan, optimizer.DropDatabasePlan,
		optimizer.ShowPlan:
		return false
	case optimizer.JoinPlan, optimizer.GroupPlan, optimizer.OrderPlan,
		optimizer.HavingPlan, optimizer.LimitPlan:
		// 复杂操作暂时不支持向量化，回退到常规执行器
		return false
	case optimizer.FilterPlan:
		// 检查过滤条件是否包含不支持的表达式
		props := plan.Properties.(*optimizer.FilterProperties)
		if !h.checkExpressionVectorizable(props.Condition) {
			return false
		}
	case optimizer.SelectPlan, optimizer.TableScanPlan:
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
	if column.IsNull(rowIdx) {
		return "NULL"
	}

	switch col := column.(type) {
	case *array.Int64:
		return col.Value(rowIdx)
	case *array.Float64:
		return col.Value(rowIdx)
	case *array.String:
		return col.Value(rowIdx)
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

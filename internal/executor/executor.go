package executor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor/operators"
	"github.com/yyun543/minidb/internal/logger"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/storage"
	"github.com/yyun543/minidb/internal/types"
	"go.uber.org/zap"
)

// NoOpOperator 空操作符，用于DDL等不需要返回数据的操作
type NoOpOperator struct{}

func (op *NoOpOperator) Init(ctx interface{}) error {
	return nil
}

func (op *NoOpOperator) Next() (*types.Batch, error) {
	return nil, nil
}

func (op *NoOpOperator) Close() error {
	return nil
}

// ExecutorImpl 执行器实现
type ExecutorImpl struct {
	catalog     *catalog.Catalog
	dataManager *DataManager
}

// BaseExecutor 是 ExecutorImpl 的类型别名，用于向后兼容
type BaseExecutor = ExecutorImpl

// NewExecutor 创建执行器实例
func NewExecutor(cat *catalog.Catalog) *ExecutorImpl {
	logger.WithComponent("executor").Info("Creating new executor instance")

	start := time.Now()
	executor := &ExecutorImpl{
		catalog:     cat,
		dataManager: NewDataManager(cat),
	}

	logger.WithComponent("executor").Info("Executor instance created successfully",
		zap.Duration("creation_time", time.Since(start)))

	return executor
}

// logExecutionResult 记录执行结果
func (e *ExecutorImpl) logExecutionResult(operation string, start time.Time, err error) {
	duration := time.Since(start)
	if err != nil {
		logger.WithComponent("executor").Error("Query execution failed",
			zap.String("operation", operation),
			zap.Duration("duration", duration),
			zap.Error(err))
	} else {
		logger.WithComponent("executor").Info("Query executed successfully",
			zap.String("operation", operation),
			zap.Duration("execution_duration", duration))
	}
}

func NewExecutorWithDataManager(cat *catalog.Catalog, dm *DataManager) *ExecutorImpl {
	return &ExecutorImpl{
		catalog:     cat,
		dataManager: dm,
	}
}

// Execute 执行查询计划
func (e *ExecutorImpl) Execute(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	logger.WithComponent("executor").Info("Executing query plan",
		zap.String("plan_type", plan.Type.String()),
		zap.Int64("session_id", sess.ID))

	start := time.Now()

	// 为 DDL/DML 操作特别处理
	switch plan.Type {
	case optimizer.CreateDatabasePlan:
		logger.WithComponent("executor").Debug("Executing CREATE DATABASE plan")
		result, err := e.executeCreateDatabase(plan, sess)
		e.logExecutionResult("CREATE DATABASE", start, err)
		return result, err
	case optimizer.DropDatabasePlan:
		logger.WithComponent("executor").Debug("Executing DROP DATABASE plan")
		result, err := e.executeDropDatabase(plan, sess)
		e.logExecutionResult("DROP DATABASE", start, err)
		return result, err
	case optimizer.CreateTablePlan:
		logger.WithComponent("executor").Debug("Executing CREATE TABLE plan")
		result, err := e.executeCreateTable(plan, sess)
		e.logExecutionResult("CREATE TABLE", start, err)
		return result, err
	case optimizer.CreateIndexPlan:
		logger.WithComponent("executor").Debug("Executing CREATE INDEX plan")
		result, err := e.executeCreateIndex(plan, sess)
		e.logExecutionResult("CREATE INDEX", start, err)
		return result, err
	case optimizer.DropIndexPlan:
		logger.WithComponent("executor").Debug("Executing DROP INDEX plan")
		result, err := e.executeDropIndex(plan, sess)
		e.logExecutionResult("DROP INDEX", start, err)
		return result, err
	case optimizer.ShowPlan:
		logger.WithComponent("executor").Debug("Executing SHOW plan")
		result, err := e.executeShow(plan, sess)
		e.logExecutionResult("SHOW", start, err)
		return result, err
	case optimizer.InsertPlan:
		logger.WithComponent("executor").Debug("Executing INSERT plan")
		result, err := e.executeInsert(plan, sess)
		e.logExecutionResult("INSERT", start, err)
		return result, err
	case optimizer.UpdatePlan:
		logger.WithComponent("executor").Debug("Executing UPDATE plan")
		result, err := e.executeUpdate(plan, sess)
		e.logExecutionResult("UPDATE", start, err)
		return result, err
	case optimizer.DeletePlan:
		logger.WithComponent("executor").Debug("Executing DELETE plan")
		result, err := e.executeDelete(plan, sess)
		e.logExecutionResult("DELETE", start, err)
		return result, err
	case optimizer.AnalyzePlan:
		logger.WithComponent("executor").Debug("Executing ANALYZE TABLE plan")
		result, err := e.executeAnalyze(plan, sess)
		e.logExecutionResult("ANALYZE", start, err)
		return result, err
	}

	logger.WithComponent("executor").Debug("Executing query plan with operator tree",
		zap.String("plan_type", plan.Type.String()))

	// 创建执行上下文
	ctxStart := time.Now()
	ctx := NewContext(e.catalog, sess, e.dataManager)
	logger.WithComponent("executor").Debug("Execution context created",
		zap.Duration("context_creation_time", time.Since(ctxStart)))

	// 构建执行算子树
	buildStart := time.Now()
	op, err := e.buildOperator(plan, ctx)
	if err != nil {
		logger.WithComponent("executor").Error("Failed to build operator tree",
			zap.String("plan_type", plan.Type.String()),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))
		return nil, err
	}
	logger.WithComponent("executor").Debug("Operator tree built successfully",
		zap.Duration("build_duration", time.Since(buildStart)))

	// 初始化算子
	initStart := time.Now()
	if err := op.Init(ctx); err != nil {
		logger.WithComponent("executor").Error("Failed to initialize operator",
			zap.String("plan_type", plan.Type.String()),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))
		return nil, err
	}
	logger.WithComponent("executor").Debug("Operator initialized successfully",
		zap.Duration("init_duration", time.Since(initStart)))

	// 执行查询并收集结果
	execStart := time.Now()
	var batches []*types.Batch
	batchCount := 0
	for {
		batch, err := op.Next()
		if err != nil {
			logger.WithComponent("executor").Error("Error during batch execution",
				zap.String("plan_type", plan.Type.String()),
				zap.Int("batches_processed", batchCount),
				zap.Duration("duration", time.Since(start)),
				zap.Error(err))
			return nil, err
		}
		if batch == nil {
			break
		}
		batches = append(batches, batch)
		batchCount++
	}
	logger.WithComponent("executor").Debug("Query execution completed",
		zap.Int("batches_collected", batchCount),
		zap.Duration("execution_duration", time.Since(execStart)))

	// 关闭算子
	closeStart := time.Now()
	if err := op.Close(); err != nil {
		logger.WithComponent("executor").Error("Failed to close operator",
			zap.String("plan_type", plan.Type.String()),
			zap.Error(err))
		return nil, err
	}
	logger.WithComponent("executor").Debug("Operator closed successfully",
		zap.Duration("close_duration", time.Since(closeStart)))

	// 构建结果集
	resultStart := time.Now()
	headers := e.getResultHeaders(plan, sess)
	result := &ResultSet{
		Headers: headers,
		rows:    batches,
		curRow:  -1,
	}

	totalDuration := time.Since(start)
	logger.WithComponent("executor").Info("Query plan execution completed successfully",
		zap.String("plan_type", plan.Type.String()),
		zap.Int("result_batches", len(batches)),
		zap.Int("result_columns", len(headers)),
		zap.Duration("total_duration", totalDuration),
		zap.Duration("result_building_duration", time.Since(resultStart)))

	return result, nil
}

// parseTableReference 解析表引用，支持 "database.table" 或 "table" 格式
func (e *ExecutorImpl) parseTableReference(tableRef string, currentDB string) (string, string) {
	parts := strings.Split(tableRef, ".")
	if len(parts) == 2 {
		// "database.table" 格式
		logger.WithComponent("executor").Debug("Parsed qualified table reference",
			zap.String("input", tableRef),
			zap.String("database", parts[0]),
			zap.String("table", parts[1]))
		return parts[0], parts[1]
	}
	// "table" 格式，使用当前数据库
	if currentDB == "" {
		currentDB = "default"
	}
	logger.WithComponent("executor").Debug("Parsed unqualified table reference",
		zap.String("input", tableRef),
		zap.String("database", currentDB),
		zap.String("table", tableRef))
	return currentDB, tableRef
}

// buildOperator 根据计划节点构建算子
func (e *ExecutorImpl) buildOperator(plan *optimizer.Plan, ctx *Context) (operators.Operator, error) {
	if plan == nil {
		return nil, fmt.Errorf("plan is nil")
	}

	switch plan.Type {
	case optimizer.SelectPlan:
		// SELECT 计划通常有一个子节点
		if len(plan.Children) > 0 {
			return e.buildOperator(plan.Children[0], ctx)
		}
		return nil, fmt.Errorf("SELECT 计划缺少子节点")

	case optimizer.TableScanPlan:
		props := plan.Properties.(*optimizer.TableScanProperties)
		// 解析表引用：支持 "database.table" 或 "table" 格式
		dbName, tableName := e.parseTableReference(props.Table, ctx.Session.CurrentDB)
		logger.WithComponent("executor").Info("Building TableScan operator",
			zap.String("props.Table", props.Table),
			zap.String("currentDB", ctx.Session.CurrentDB),
			zap.String("resolved_db", dbName),
			zap.String("resolved_table", tableName))
		return operators.NewTableScan(dbName, tableName, e.catalog, e.dataManager), nil

	case optimizer.JoinPlan:
		props := plan.Properties.(*optimizer.JoinProperties)
		left, err := e.buildOperator(plan.Children[0], ctx)
		if err != nil {
			return nil, err
		}
		right, err := e.buildOperator(plan.Children[1], ctx)
		if err != nil {
			return nil, err
		}
		return operators.NewJoin(props.JoinType, props.Condition, left, right, ctx), nil

	case optimizer.ProjectionPlan:
		props := plan.Properties.(*optimizer.ProjectionProperties)
		child, err := e.buildOperator(plan.Children[0], ctx)
		if err != nil {
			return nil, err
		}
		return operators.NewProjection(props.Columns, child, ctx), nil

	case optimizer.FilterPlan:
		props := plan.Properties.(*optimizer.FilterProperties)
		child, err := e.buildOperator(plan.Children[0], ctx)
		if err != nil {
			return nil, err
		}
		return operators.NewFilter(props.Condition, child, ctx), nil

	case optimizer.HavingPlan:
		props := plan.Properties.(*optimizer.HavingProperties)
		child, err := e.buildOperator(plan.Children[0], ctx)
		if err != nil {
			return nil, err
		}
		// HAVING is conceptually similar to filtering, but operates on grouped data
		return operators.NewFilter(props.Condition, child, ctx), nil

	case optimizer.GroupPlan:
		props := plan.Properties.(*optimizer.GroupByProperties)
		child, err := e.buildOperator(plan.Children[0], ctx)
		if err != nil {
			return nil, err
		}
		return operators.NewGroupBy(props.GroupKeys, props.Aggregations, props.SelectColumns, child, ctx), nil

	case optimizer.OrderPlan:
		props := plan.Properties.(*optimizer.OrderByProperties)
		child, err := e.buildOperator(plan.Children[0], ctx)
		if err != nil {
			return nil, err
		}
		return operators.NewOrderBy(props.OrderKeys, child, ctx), nil

	case optimizer.LimitPlan:
		props := plan.Properties.(*optimizer.LimitProperties)
		child, err := e.buildOperator(plan.Children[0], ctx)
		if err != nil {
			return nil, err
		}
		// Create Limit operator (offset=0 for now, can be extended later)
		return operators.NewLimit(int64(props.Limit), 0, child, ctx), nil

	case optimizer.DropTablePlan:
		// For DDL operations, create simple NoOp operator
		return &NoOpOperator{}, nil

	case optimizer.TransactionPlan:
		// For transaction operations, create simple NoOp operator
		return &NoOpOperator{}, nil

	case optimizer.UsePlan:
		// For USE database operations, create simple NoOp operator
		return &NoOpOperator{}, nil

	case optimizer.ExplainPlan:
		// For EXPLAIN operations, create simple NoOp operator
		return &NoOpOperator{}, nil

	case optimizer.CreateTablePlan:
		// 对于DDL操作，我们创建一个简单的空操作符
		return &NoOpOperator{}, nil

	case optimizer.DropDatabasePlan:
		// 对于DROP DATABASE操作，创建简单的空操作符
		return &NoOpOperator{}, nil

	case optimizer.InsertPlan:
		// 对于DML操作，创建简单的空操作符
		return &NoOpOperator{}, nil

	default:
		return nil, fmt.Errorf("不支持的计划节点类型: %v", plan.Type)
	}
}

// getResultHeaders 获取结果集列名
func (e *ExecutorImpl) getResultHeaders(plan *optimizer.Plan, sess *session.Session) []string {
	switch plan.Type {
	case optimizer.SelectPlan:
		// 递归查找GROUP BY计划（可能在多层嵌套中）
		groupPlan := e.findGroupByPlan(plan)
		if groupPlan != nil {
			groupProps := groupPlan.Properties.(*optimizer.GroupByProperties)
			headers := make([]string, len(groupProps.SelectColumns))
			for i, col := range groupProps.SelectColumns {
				if col.Alias != "" {
					headers[i] = col.Alias
				} else if col.Type == optimizer.ColumnRefTypeFunction {
					headers[i] = fmt.Sprintf("%s(%s)", col.FunctionName, col.Column)
				} else {
					if col.Table != "" {
						headers[i] = fmt.Sprintf("%s.%s", col.Table, col.Column)
					} else {
						headers[i] = col.Column
					}
				}
			}
			return headers
		}

		props := plan.Properties.(*optimizer.SelectProperties)

		// 处理SELECT *的情况
		if props.All || len(props.Columns) == 0 {
			// 递归查找表扫描节点获取schema
			headers := e.getSchemaFromPlan(plan.Children[0], sess)
			if headers != nil {
				return headers
			}
			// fallback: 返回通用列名
			return []string{"*"}
		}

		columns := make([]string, len(props.Columns))
		for i, col := range props.Columns {
			if col.Table != "" {
				columns[i] = fmt.Sprintf("%s.%s", col.Table, col.Column)
			} else {
				columns[i] = col.Column
			}
		}
		return columns

	case optimizer.ProjectionPlan:
		props := plan.Properties.(*optimizer.ProjectionProperties)
		headers := make([]string, len(props.Columns))
		for i, col := range props.Columns {
			if col.Alias != "" {
				headers[i] = col.Alias
			} else if col.Table != "" {
				headers[i] = fmt.Sprintf("%s.%s", col.Table, col.Column)
			} else {
				headers[i] = col.Column
			}
		}
		return headers

	case optimizer.GroupPlan:
		props := plan.Properties.(*optimizer.GroupByProperties)
		headers := make([]string, len(props.SelectColumns))
		for i, col := range props.SelectColumns {
			if col.Alias != "" {
				headers[i] = col.Alias
			} else if col.Type == optimizer.ColumnRefTypeFunction {
				headers[i] = fmt.Sprintf("%s(%s)", col.FunctionName, col.Column)
			} else {
				headers[i] = col.Column
			}
		}
		return headers

	default:
		return nil
	}
}

// getSchemaFromPlan 递归从计划树中获取schema信息
func (e *ExecutorImpl) getSchemaFromPlan(plan *optimizer.Plan, sess *session.Session) []string {
	if plan == nil {
		return nil
	}

	switch plan.Type {
	case optimizer.TableScanPlan:
		// 直接从表扫描获取schema
		tableScanProps := plan.Properties.(*optimizer.TableScanProperties)
		currentDB := sess.CurrentDB
		if currentDB == "" {
			currentDB = "default"
		}
		if tableMeta, err := e.catalog.GetTable(currentDB, tableScanProps.Table); err == nil {
			headers := make([]string, len(tableMeta.Schema.Fields()))
			for i, field := range tableMeta.Schema.Fields() {
				headers[i] = field.Name
			}
			return headers
		}
	case optimizer.JoinPlan:
		// 从JOIN获取合并后的schema
		var allHeaders []string
		for _, child := range plan.Children {
			childHeaders := e.getSchemaFromPlan(child, sess)
			if childHeaders != nil {
				allHeaders = append(allHeaders, childHeaders...)
			}
		}
		return allHeaders
	case optimizer.FilterPlan:
		// 过滤不改变schema，递归到子节点
		return e.getSchemaFromPlan(plan.Children[0], sess)
	default:
		// 其他类型，尝试递归到第一个子节点
		if len(plan.Children) > 0 {
			return e.getSchemaFromPlan(plan.Children[0], sess)
		}
	}

	return nil
}

// findGroupByPlan 递归查找计划树中的GroupBy节点
func (e *ExecutorImpl) findGroupByPlan(plan *optimizer.Plan) *optimizer.Plan {
	if plan == nil {
		return nil
	}

	// 如果当前节点是GroupBy，直接返回
	if plan.Type == optimizer.GroupPlan {
		return plan
	}

	// 递归搜索所有子节点
	for _, child := range plan.Children {
		if groupPlan := e.findGroupByPlan(child); groupPlan != nil {
			return groupPlan
		}
	}

	return nil
}

// executeCreateTable 执行创建表操作
func (e *ExecutorImpl) executeCreateTable(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props := plan.Properties.(*optimizer.CreateTableProperties)

	// 创建 Arrow Schema
	fields := make([]arrow.Field, len(props.Columns))
	for i, col := range props.Columns {
		// 从列定义中获取类型，默认根据列名推断
		dataType := e.convertToArrowType(col.Column)
		fields[i] = arrow.Field{
			Name: col.Column,
			Type: dataType,
		}
	}
	schema := arrow.NewSchema(fields, nil)

	// 使用会话中的当前数据库，默认为"default"
	currentDB := sess.CurrentDB
	if currentDB == "" {
		currentDB = "default"
	}

	// 创建表元数据
	tableMeta := catalog.TableMeta{
		Database:   currentDB,
		Table:      props.Table,
		ChunkCount: 0,
		Schema:     schema,
	}

	err := e.catalog.CreateTable(currentDB, tableMeta)
	if err != nil {
		return nil, err
	}

	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// executeInsert 执行插入操作
func (e *ExecutorImpl) executeInsert(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props := plan.Properties.(*optimizer.InsertProperties)

	// 提取插入的值
	values := make([]interface{}, len(props.Values))
	for i, expr := range props.Values {
		if lit, ok := expr.(*optimizer.LiteralValue); ok {
			values[i] = lit.Value
		}
	}

	// 使用会话中的当前数据库，默认为"default"
	currentDB := sess.CurrentDB
	if currentDB == "" {
		currentDB = "default"
	}

	// 如果没有指定列名，使用表的所有列
	columns := props.Columns
	if len(columns) == 0 {
		// 获取表的schema来确定列顺序
		tableMeta, err := e.catalog.GetTable(currentDB, props.Table)
		if err != nil {
			return nil, err
		}

		// 使用schema中的字段名
		columns = make([]string, len(tableMeta.Schema.Fields()))
		for i, field := range tableMeta.Schema.Fields() {
			columns[i] = field.Name
		}
	}

	// 使用 DataManager 插入数据
	err := e.dataManager.InsertData(currentDB, props.Table, columns, values)
	if err != nil {
		return nil, err
	}

	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// executeUpdate 执行更新操作
func (e *ExecutorImpl) executeUpdate(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props := plan.Properties.(*optimizer.UpdateProperties)

	// 解析赋值表达式，将Expression转换为实际值
	assignments := make(map[string]interface{})
	for column, expr := range props.Assignments {
		if litExpr, ok := expr.(*optimizer.LiteralValue); ok {
			assignments[column] = litExpr.Value
		} else if strLit, ok := expr.(*parser.StringLiteral); ok {
			assignments[column] = strLit.Value
		} else if intLit, ok := expr.(*parser.IntegerLiteral); ok {
			assignments[column] = intLit.Value
		}
	}

	// 将WHERE表达式转换为storage.Filter
	filters := e.whereToFilters(props.Where)

	// 使用会话中的当前数据库
	currentDB := sess.CurrentDB
	if currentDB == "" {
		currentDB = "default"
	}

	// 使用 DataManager 更新数据（传递filters而不是whereCondition）
	err := e.dataManager.UpdateDataWithFilters(currentDB, props.Table, assignments, filters)
	if err != nil {
		return nil, err
	}

	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// evaluateWhereCondition 评估WHERE条件
func (e *ExecutorImpl) evaluateWhereCondition(record arrow.Record, rowIdx int, whereExpr interface{}) bool {
	// 如果没有WHERE条件，匹配所有行
	if whereExpr == nil {
		return true
	}

	// 尝试解析为BinaryExpr（如 id = 1）
	if binExpr, ok := whereExpr.(*parser.BinaryExpr); ok {
		return e.evaluateBinaryCondition(record, rowIdx, binExpr)
	}

	// 尝试解析为InExpr（如 amount IN (100, 250)）
	if inExpr, ok := whereExpr.(*parser.InExpr); ok {
		return e.evaluateInCondition(record, rowIdx, inExpr)
	}

	// 对于其他类型的表达式，暂时返回false（保守策略）
	return false
}

// evaluateBinaryCondition 评估二元条件表达式
func (e *ExecutorImpl) evaluateBinaryCondition(record arrow.Record, rowIdx int, expr *parser.BinaryExpr) bool {
	// 支持多种比较操作符
	supportedOps := map[string]bool{
		"=": true, "!=": true, "<>": true,
		"<": true, "<=": true, ">": true, ">=": true,
	}
	if !supportedOps[expr.Operator] {
		return false
	}

	// 左操作数应该是列引用
	leftCol, ok := expr.Left.(*parser.ColumnRef)
	if !ok {
		return false
	}

	// 右操作数应该是字面值
	var rightValue interface{}
	switch rightNode := expr.Right.(type) {
	case *parser.IntegerLiteral:
		rightValue = rightNode.Value
	case *parser.StringLiteral:
		rightValue = rightNode.Value
	default:
		return false
	}

	// 查找列在schema中的位置
	columnIndex := -1
	schema := record.Schema()
	for i, field := range schema.Fields() {
		if field.Name == leftCol.Column {
			columnIndex = i
			break
		}
	}

	if columnIndex == -1 {
		return false // 列不存在
	}

	// 获取该行该列的实际值
	column := record.Column(columnIndex)
	var actualValue interface{}

	switch col := column.(type) {
	case *array.Int64:
		if rowIdx < col.Len() {
			actualValue = col.Value(rowIdx)
		}
	case *array.String:
		if rowIdx < col.Len() {
			actualValue = col.Value(rowIdx)
		}
	default:
		return false // 不支持的列类型
	}

	// 根据操作符进行比较
	return e.compareValues(actualValue, rightValue, expr.Operator)
}

// compareValues 比较两个值
func (e *ExecutorImpl) compareValues(left, right interface{}, operator string) bool {
	switch operator {
	case "=":
		return left == right
	case "!=", "<>":
		return left != right
	case "<":
		return e.compareOrderedValues(left, right) < 0
	case "<=":
		return e.compareOrderedValues(left, right) <= 0
	case ">":
		return e.compareOrderedValues(left, right) > 0
	case ">=":
		return e.compareOrderedValues(left, right) >= 0
	default:
		return false
	}
}

// compareOrderedValues 比较两个有序值，返回比较结果 (-1, 0, 1)
func (e *ExecutorImpl) compareOrderedValues(left, right interface{}) int {
	// 尝试作为int64比较
	if leftInt, ok := left.(int64); ok {
		if rightInt, ok := right.(int64); ok {
			if leftInt < rightInt {
				return -1
			} else if leftInt > rightInt {
				return 1
			}
			return 0
		}
	}

	// 尝试作为字符串比较
	if leftStr, ok := left.(string); ok {
		if rightStr, ok := right.(string); ok {
			if leftStr < rightStr {
				return -1
			} else if leftStr > rightStr {
				return 1
			}
			return 0
		}
	}

	// 不支持的类型比较
	return 0
}

// evaluateInCondition 评估IN条件表达式
func (e *ExecutorImpl) evaluateInCondition(record arrow.Record, rowIdx int, expr *parser.InExpr) bool {
	// 左操作数应该是列引用
	leftCol, ok := expr.Left.(*parser.ColumnRef)
	if !ok {
		return false
	}

	// 查找列在schema中的位置
	columnIndex := -1
	schema := record.Schema()
	for i, field := range schema.Fields() {
		if field.Name == leftCol.Column {
			columnIndex = i
			break
		}
	}

	if columnIndex == -1 {
		return false // 列不存在
	}

	// 获取该行该列的实际值
	column := record.Column(columnIndex)
	var actualValue interface{}

	switch col := column.(type) {
	case *array.Int64:
		if rowIdx < col.Len() {
			actualValue = col.Value(rowIdx)
		}
	case *array.String:
		if rowIdx < col.Len() {
			actualValue = col.Value(rowIdx)
		}
	default:
		return false
	}

	// 检查值是否在IN列表中
	for _, valueNode := range expr.Values {
		var inValue interface{}
		switch val := valueNode.(type) {
		case *parser.IntegerLiteral:
			inValue = val.Value
		case *parser.StringLiteral:
			inValue = val.Value
		default:
			continue
		}

		// 如果找到匹配的值
		if actualValue == inValue {
			// 对于 IN 操作，找到匹配就返回true
			// 对于 NOT IN 操作，找到匹配应该返回false
			return expr.Operator == "IN"
		}
	}

	// 没有找到匹配的值
	// 对于 IN 操作，没找到匹配返回false
	// 对于 NOT IN 操作，没找到匹配返回true
	return expr.Operator == "NOT IN"
}

// executeDelete 执行删除操作
func (e *ExecutorImpl) executeDelete(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props := plan.Properties.(*optimizer.DeleteProperties)

	// 将WHERE表达式转换为storage.Filter
	filters := e.whereToFilters(props.Where)

	// 使用会话中的当前数据库
	currentDB := sess.CurrentDB
	if currentDB == "" {
		currentDB = "default"
	}

	// 使用 DataManager 删除数据（传递filters而不是whereCondition）
	err := e.dataManager.DeleteDataWithFilters(currentDB, props.Table, filters)
	if err != nil {
		return nil, err
	}

	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// executeCreateDatabase 执行创建数据库操作
func (e *ExecutorImpl) executeCreateDatabase(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props := plan.Properties.(*optimizer.CreateDatabaseProperties)

	err := e.catalog.CreateDatabase(props.Database)
	if err != nil {
		return nil, err
	}

	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// executeDropDatabase 执行删除数据库操作
func (e *ExecutorImpl) executeDropDatabase(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props := plan.Properties.(*optimizer.DropDatabaseProperties)

	err := e.catalog.DropDatabase(props.Database)
	if err != nil {
		return nil, err
	}

	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// executeShow 执行SHOW命令
func (e *ExecutorImpl) executeShow(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	// 使用类型断言判断是ShowProperties还是ShowIndexesProperties
	switch props := plan.Properties.(type) {
	case *optimizer.ShowProperties:
		// 处理 SHOW DATABASES 和 SHOW TABLES
		switch props.Type {
		case "DATABASES":
			databaseNames, err := e.catalog.GetAllDatabases()
			if err != nil {
				return nil, err
			}

			// 如果没有数据库，返回空结果集
			if len(databaseNames) == 0 {
				return &ResultSet{
					Headers: []string{"Database"},
					rows:    []*types.Batch{},
					curRow:  -1,
				}, nil
			}

			// 创建包含数据库名的结果集
			batch, err := e.createDatabaseListBatch(databaseNames)
			if err != nil {
				return nil, fmt.Errorf("failed to create database list batch: %w", err)
			}

			return &ResultSet{
				Headers: []string{"Database"},
				rows:    []*types.Batch{batch},
				curRow:  -1,
			}, nil

		case "TABLES":
			// 获取当前数据库名
			currentDB := sess.CurrentDB
			if currentDB == "" {
				currentDB = "default"
			}

			// 从catalog获取表列表
			tableNames, err := e.catalog.GetAllTables(currentDB)
			if err != nil {
				return nil, fmt.Errorf("failed to get table list: %w", err)
			}

			// 如果没有表，返回空结果集但不出错
			if len(tableNames) == 0 {
				return &ResultSet{
					Headers: []string{"Tables_in_" + currentDB},
					rows:    []*types.Batch{},
					curRow:  -1,
				}, nil
			}

			// 创建包含表名的结果集
			batch, err := e.createTableListBatch(tableNames, currentDB)
			if err != nil {
				return nil, fmt.Errorf("failed to create table list batch: %w", err)
			}

			return &ResultSet{
				Headers: []string{"Tables_in_" + currentDB},
				rows:    []*types.Batch{batch},
				curRow:  -1,
			}, nil

		default:
			return nil, fmt.Errorf("unsupported SHOW type: %s", props.Type)
		}

	case *optimizer.ShowIndexesProperties:
		// 处理 SHOW INDEXES
		currentDB := sess.CurrentDB
		if currentDB == "" {
			currentDB = "default"
		}

		// 从catalog获取索引列表
		indexes, err := e.catalog.GetAllIndexes(currentDB, props.Table)
		if err != nil {
			return nil, fmt.Errorf("failed to get indexes: %w", err)
		}

		// 如果没有索引，返回空结果集
		if len(indexes) == 0 {
			return &ResultSet{
				Headers: []string{"Table", "Index_name", "Column_name", "Is_unique"},
				rows:    []*types.Batch{},
				curRow:  -1,
			}, nil
		}

		// 创建包含索引信息的结果集
		batch, err := e.createIndexListBatch(indexes, props.Table)
		if err != nil {
			return nil, fmt.Errorf("failed to create index list batch: %w", err)
		}

		return &ResultSet{
			Headers: []string{"Table", "Index_name", "Column_name", "Is_unique"},
			rows:    []*types.Batch{batch},
			curRow:  -1,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported SHOW properties type: %T", props)
	}
}

// createTableListBatch 创建包含表名列表的批次数据
func (e *ExecutorImpl) createTableListBatch(tableNames []string, dbName string) (*types.Batch, error) {
	// 创建Arrow schema，包含一个字符串列
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "Tables_in_" + dbName, Type: arrow.BinaryTypes.String},
	}, nil)

	// 创建Arrow记录
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 填充表名数据
	stringBuilder := builder.Field(0).(*array.StringBuilder)
	for _, tableName := range tableNames {
		stringBuilder.Append(tableName)
	}

	// 构建记录
	record := builder.NewRecord()
	defer record.Release()

	// 创建批次
	batch := types.NewBatch(record)
	return batch, nil
}

// createDatabaseListBatch 创建包含数据库名列表的批次数据
func (e *ExecutorImpl) createDatabaseListBatch(databaseNames []string) (*types.Batch, error) {
	// 创建Arrow schema，包含一个字符串列
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "Database", Type: arrow.BinaryTypes.String},
	}, nil)

	// 创建Arrow记录
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 填充数据库名数据
	stringBuilder := builder.Field(0).(*array.StringBuilder)
	for _, dbName := range databaseNames {
		stringBuilder.Append(dbName)
	}

	// 构建记录
	record := builder.NewRecord()
	defer record.Release()

	// 创建批次
	batch := types.NewBatch(record)
	return batch, nil
}

// createIndexListBatch 创建包含索引信息的批次数据
func (e *ExecutorImpl) createIndexListBatch(indexes []catalog.IndexMeta, tableName string) (*types.Batch, error) {
	// 创建Arrow schema，包含四个字符串列
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "Table", Type: arrow.BinaryTypes.String},
		{Name: "Index_name", Type: arrow.BinaryTypes.String},
		{Name: "Column_name", Type: arrow.BinaryTypes.String},
		{Name: "Is_unique", Type: arrow.BinaryTypes.String},
	}, nil)

	// 创建Arrow记录
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 获取各列的builder
	tableBuilder := builder.Field(0).(*array.StringBuilder)
	indexNameBuilder := builder.Field(1).(*array.StringBuilder)
	columnNameBuilder := builder.Field(2).(*array.StringBuilder)
	isUniqueBuilder := builder.Field(3).(*array.StringBuilder)

	// 填充索引数据
	for _, index := range indexes {
		for _, column := range index.Columns {
			tableBuilder.Append(tableName)
			indexNameBuilder.Append(index.Name)
			columnNameBuilder.Append(column)
			if index.IsUnique {
				isUniqueBuilder.Append("YES")
			} else {
				isUniqueBuilder.Append("NO")
			}
		}
	}

	// 构建记录
	record := builder.NewRecord()
	defer record.Release()

	// 创建批次
	batch := types.NewBatch(record)
	return batch, nil
}

// convertToArrowType 将字符串类型转换为 Arrow 类型
func (e *ExecutorImpl) convertToArrowType(typeName string) arrow.DataType {
	switch typeName {
	case "INTEGER", "INT", "id", "age", "user_id", "amount":
		return arrow.PrimitiveTypes.Int64
	case "VARCHAR", "name", "email", "created_at", "order_date":
		return arrow.BinaryTypes.String
	default:
		// 默认根据列名推断类型
		lowerName := strings.ToLower(typeName)
		if strings.Contains(lowerName, "id") || lowerName == "age" || lowerName == "amount" {
			return arrow.PrimitiveTypes.Int64
		}
		return arrow.BinaryTypes.String
	}
}

// convertSQLTypeToArrow 将SQL类型转换为Arrow类型
func (e *ExecutorImpl) convertSQLTypeToArrow(sqlType string) arrow.DataType {
	upperType := strings.ToUpper(sqlType)
	switch upperType {
	case "INT", "INTEGER":
		return arrow.PrimitiveTypes.Int64
	case "VARCHAR", "TEXT", "STRING":
		return arrow.BinaryTypes.String
	case "FLOAT", "DOUBLE":
		return arrow.PrimitiveTypes.Float64
	case "BOOLEAN", "BOOL":
		return arrow.FixedWidthTypes.Boolean
	default:
		return arrow.BinaryTypes.String
	}
}

// executeCreateIndex 执行创建索引操作
func (e *ExecutorImpl) executeCreateIndex(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props := plan.Properties.(*optimizer.CreateIndexProperties)

	currentDB := sess.CurrentDB
	if currentDB == "" {
		currentDB = "default"
	}

	// 创建索引元数据
	indexMeta := catalog.IndexMeta{
		Database:  currentDB,
		Table:     props.Table,
		Name:      props.Name,
		Columns:   props.Columns,
		IsUnique:  props.IsUnique,
		IndexType: "BTREE", // 默认使用B树索引
	}

	// 调用catalog创建索引
	err := e.catalog.CreateIndex(indexMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	// 返回成功结果
	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// executeDropIndex 执行删除索引操作
func (e *ExecutorImpl) executeDropIndex(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props := plan.Properties.(*optimizer.DropIndexProperties)

	currentDB := sess.CurrentDB
	if currentDB == "" {
		currentDB = "default"
	}

	// 调用catalog删除索引
	err := e.catalog.DropIndex(currentDB, props.Table, props.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to drop index: %w", err)
	}

	// 返回成功结果
	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// whereToFilters converts WHERE expression to storage.Filter slice
func (e *ExecutorImpl) whereToFilters(whereExpr interface{}) []storage.Filter {
	if whereExpr == nil {
		logger.Debug("whereToFilters: no WHERE clause")
		return []storage.Filter{}
	}

	var filters []storage.Filter

	// Handle BinaryExpr (e.g., "name = 'John'" or "age > 25")
	if binExpr, ok := whereExpr.(*parser.BinaryExpr); ok {
		filter := e.binaryExprToFilter(binExpr)
		if filter != nil {
			filters = append(filters, *filter)
			logger.Debug("whereToFilters: converted BinaryExpr to filter",
				zap.String("column", filter.Column),
				zap.String("operator", filter.Operator),
				zap.Any("value", filter.Value))
		} else {
			logger.Warn("whereToFilters: failed to convert BinaryExpr to filter")
		}
	} else {
		logger.Warn("whereToFilters: WHERE expression is not a BinaryExpr",
			zap.String("type", fmt.Sprintf("%T", whereExpr)))
	}

	// Handle AND expressions (multiple filters)
	// 实现对 OR, NOT, AND 和复杂表达式的支持
	if logicalExpr, ok := whereExpr.(*parser.LogicalExpr); ok {
		switch logicalExpr.Operator {
		case "AND":
			// 处理 AND 表达式：递归处理左右子表达式
			leftFilters := e.whereToFilters(logicalExpr.Left)
			rightFilters := e.whereToFilters(logicalExpr.Right)
			filters = append(filters, leftFilters...)
			filters = append(filters, rightFilters...)
			logger.Debug("whereToFilters: processed AND expression",
				zap.Int("left_filters", len(leftFilters)),
				zap.Int("right_filters", len(rightFilters)))
		case "OR":
			// 处理 OR 表达式：在简化实现中，OR操作较复杂
			// 当前设计中每个storage.Filter表示一个AND条件
			// 对于OR操作，我们需要特殊处理或在更高层实现
			logger.Warn("whereToFilters: OR expressions not fully supported yet",
				zap.String("suggestion", "consider using IN operator or multiple queries"))
			// 暂时处理为分别收集两边的过滤条件（这不是真正的OR语义）
			leftFilters := e.whereToFilters(logicalExpr.Left)
			rightFilters := e.whereToFilters(logicalExpr.Right)
			// 注意：这里的实现不是真正的OR逻辑，需要在storage层改进
			filters = append(filters, leftFilters...)
			filters = append(filters, rightFilters...)
		case "NOT":
			// 处理 NOT 表达式：需要将内部条件取反
			// 注意：在LogicalExpr中，NOT是一元操作，只有Left操作数
			innerFilters := e.whereToFilters(logicalExpr.Left)
			for _, filter := range innerFilters {
				// 将操作符取反
				negatedFilter := e.negateFilter(filter)
				if negatedFilter != nil {
					filters = append(filters, *negatedFilter)
					logger.Debug("whereToFilters: negated filter",
						zap.String("original_op", filter.Operator),
						zap.String("negated_op", negatedFilter.Operator))
				}
			}
		}
	}

	return filters
}

// binaryExprToFilter converts a binary expression to a storage.Filter
func (e *ExecutorImpl) binaryExprToFilter(expr *parser.BinaryExpr) *storage.Filter {
	// Left side should be a column reference
	var column string
	if colRef, ok := expr.Left.(*parser.ColumnRef); ok {
		column = colRef.Column
	} else {
		logger.Warn("binaryExprToFilter: left side is not a column reference",
			zap.String("type", fmt.Sprintf("%T", expr.Left)))
		return nil // Cannot convert non-column reference
	}

	// Operator mapping
	operator := expr.Operator

	// Right side should be a literal value
	var value interface{}
	switch right := expr.Right.(type) {
	case *parser.StringLiteral:
		value = right.Value
	case *parser.IntegerLiteral:
		value = right.Value
	case *parser.FloatLiteral:
		value = right.Value
	case *parser.BooleanLiteral:
		value = right.Value
	default:
		logger.Warn("binaryExprToFilter: right side is not a literal",
			zap.String("type", fmt.Sprintf("%T", expr.Right)))
		return nil // Cannot convert non-literal value
	}

	logger.Debug("binaryExprToFilter: successfully converted",
		zap.String("column", column),
		zap.String("operator", operator),
		zap.Any("value", value))

	return &storage.Filter{
		Column:   column,
		Operator: operator,
		Value:    value,
	}
}

// negateFilter 将过滤条件取反
func (e *ExecutorImpl) negateFilter(filter storage.Filter) *storage.Filter {
	var negatedOp string

	// 将操作符映射到其否定形式
	switch filter.Operator {
	case "=":
		negatedOp = "!="
	case "!=", "<>":
		negatedOp = "="
	case "<":
		negatedOp = ">="
	case "<=":
		negatedOp = ">"
	case ">":
		negatedOp = "<="
	case ">=":
		negatedOp = "<"
	case "LIKE":
		negatedOp = "NOT LIKE"
	case "NOT LIKE":
		negatedOp = "LIKE"
	case "IN":
		negatedOp = "NOT IN"
	case "NOT IN":
		negatedOp = "IN"
	default:
		// 对于不支持的操作符，记录警告并返回nil
		logger.Warn("negateFilter: unsupported operator for negation",
			zap.String("operator", filter.Operator))
		return nil
	}

	return &storage.Filter{
		Column:   filter.Column,
		Operator: negatedOp,
		Value:    filter.Value,
	}
}

// executeAnalyze 执行ANALYZE TABLE语句
func (e *ExecutorImpl) executeAnalyze(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props, ok := plan.Properties.(*optimizer.AnalyzeProperties)
	if !ok {
		return nil, fmt.Errorf("invalid ANALYZE plan properties")
	}

	logger.WithComponent("executor").Info("Executing ANALYZE TABLE",
		zap.String("table", props.Table),
		zap.Strings("columns", props.Columns))

	// 解析表名（可能包含数据库名）
	tableParts := strings.Split(props.Table, ".")
	var dbName, tableName string
	if len(tableParts) == 2 {
		dbName = tableParts[0]
		tableName = tableParts[1]
	} else {
		dbName = sess.CurrentDB
		tableName = props.Table
	}

	if dbName == "" {
		return nil, fmt.Errorf("no database selected")
	}

	// 检查表是否存在
	table, err := e.catalog.GetTable(dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("table %s.%s not found: %w", dbName, tableName, err)
	}

	// 获取ParquetEngine通过类型断言
	storageEngine := e.catalog.GetStorageEngine()
	engine, ok := storageEngine.(*storage.ParquetEngine)
	if !ok {
		return nil, fmt.Errorf("storage engine is not ParquetEngine")
	}

	// 确保sys数据库和系统表存在
	if err := e.ensureSystemTables(engine); err != nil {
		return nil, fmt.Errorf("failed to ensure system tables: %w", err)
	}

	// 收集表级统计信息
	if err := e.collectTableStatistics(engine, dbName, tableName); err != nil {
		return nil, fmt.Errorf("failed to collect table statistics: %w", err)
	}

	// 收集列级统计信息
	columns := props.Columns
	if len(columns) == 0 {
		// 如果未指定列，收集所有列的统计信息
		for _, field := range table.Schema.Fields() {
			columns = append(columns, field.Name)
		}
	}

	if err := e.collectColumnStatistics(engine, dbName, tableName, columns); err != nil {
		return nil, fmt.Errorf("failed to collect column statistics: %w", err)
	}

	logger.WithComponent("executor").Info("ANALYZE TABLE completed successfully",
		zap.String("table", props.Table),
		zap.Int("columns_analyzed", len(columns)))

	// 返回成功消息
	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// ensureSystemTables 确保系统表存在
func (e *ExecutorImpl) ensureSystemTables(engine *storage.ParquetEngine) error {
	// 创建sys数据库（如果不存在）
	if err := engine.CreateDatabase("sys"); err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create sys database: %w", err)
	}

	// 创建sys.table_statistics表（如果不存在）
	tableStatsSchema := arrow.NewSchema([]arrow.Field{
		{Name: "table_id", Type: arrow.BinaryTypes.String},
		{Name: "version", Type: arrow.PrimitiveTypes.Int64},
		{Name: "row_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "file_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "total_size_bytes", Type: arrow.PrimitiveTypes.Int64},
		{Name: "last_updated", Type: arrow.PrimitiveTypes.Int64},
	}, nil)

	if err := engine.CreateTable("sys", "table_statistics", tableStatsSchema); err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create sys.table_statistics: %w", err)
	}

	// 创建sys.column_statistics表（如果不存在）
	columnStatsSchema := arrow.NewSchema([]arrow.Field{
		{Name: "table_id", Type: arrow.BinaryTypes.String},
		{Name: "column_name", Type: arrow.BinaryTypes.String},
		{Name: "data_type", Type: arrow.BinaryTypes.String},
		{Name: "min_value", Type: arrow.BinaryTypes.String},
		{Name: "max_value", Type: arrow.BinaryTypes.String},
		{Name: "null_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "distinct_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "last_updated", Type: arrow.PrimitiveTypes.Int64},
	}, nil)

	if err := engine.CreateTable("sys", "column_statistics", columnStatsSchema); err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("failed to create sys.column_statistics: %w", err)
	}

	return nil
}

// collectTableStatistics 收集表级统计信息
func (e *ExecutorImpl) collectTableStatistics(engine *storage.ParquetEngine, dbName, tableName string) error {
	tableID := fmt.Sprintf("%s.%s", dbName, tableName)

	// 使用COUNT(*)查询获取行数
	rowCount, err := e.getRowCount(engine, dbName, tableName)
	if err != nil {
		return fmt.Errorf("failed to get row count: %w", err)
	}

	// 构建统计记录
	pool := memory.NewGoAllocator()
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "table_id", Type: arrow.BinaryTypes.String},
		{Name: "version", Type: arrow.PrimitiveTypes.Int64},
		{Name: "row_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "file_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "total_size_bytes", Type: arrow.PrimitiveTypes.Int64},
		{Name: "last_updated", Type: arrow.PrimitiveTypes.Int64},
	}, nil)

	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	builder.Field(0).(*array.StringBuilder).Append(tableID)
	builder.Field(1).(*array.Int64Builder).Append(0) // version: 0 for now
	builder.Field(2).(*array.Int64Builder).Append(rowCount)
	builder.Field(3).(*array.Int64Builder).Append(0) // file_count: 0 for now
	builder.Field(4).(*array.Int64Builder).Append(0) // total_size_bytes: 0 for now
	builder.Field(5).(*array.Int64Builder).Append(time.Now().Unix())

	record := builder.NewRecord()
	defer record.Release()

	// 写入sys.table_statistics
	ctx := context.Background()
	if err := engine.Write(ctx, "sys", "table_statistics", record); err != nil {
		return fmt.Errorf("failed to write table statistics: %w", err)
	}

	logger.WithComponent("executor").Info("Table statistics collected",
		zap.String("table_id", tableID),
		zap.Int64("row_count", rowCount))

	return nil
}

// collectColumnStatistics 收集列级统计信息
func (e *ExecutorImpl) collectColumnStatistics(engine *storage.ParquetEngine, dbName, tableName string, columns []string) error {
	tableID := fmt.Sprintf("%s.%s", dbName, tableName)

	// 获取表schema
	table, err := e.catalog.GetTable(dbName, tableName)
	if err != nil {
		return fmt.Errorf("failed to get table: %w", err)
	}

	pool := memory.NewGoAllocator()
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "table_id", Type: arrow.BinaryTypes.String},
		{Name: "column_name", Type: arrow.BinaryTypes.String},
		{Name: "data_type", Type: arrow.BinaryTypes.String},
		{Name: "min_value", Type: arrow.BinaryTypes.String},
		{Name: "max_value", Type: arrow.BinaryTypes.String},
		{Name: "null_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "distinct_count", Type: arrow.PrimitiveTypes.Int64},
		{Name: "last_updated", Type: arrow.PrimitiveTypes.Int64},
	}, nil)

	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 扫描表数据收集统计
	ctx := context.Background()
	iterator, err := engine.Scan(ctx, dbName, tableName, nil)
	if err != nil {
		return fmt.Errorf("failed to scan table: %w", err)
	}

	// 为每个列初始化统计收集器
	columnStats := make(map[string]*ColumnStatsCollector)
	for _, colName := range columns {
		// 查找列的数据类型
		var colType arrow.DataType
		for _, field := range table.Schema.Fields() {
			if field.Name == colName {
				colType = field.Type
				break
			}
		}
		if colType == nil {
			logger.WithComponent("executor").Warn("Column not found in schema, skipping",
				zap.String("column", colName))
			continue
		}
		columnStats[colName] = NewColumnStatsCollector(colType)
	}

	// 遍历数据收集统计
	for iterator.Next() {
		record := iterator.Record()
		for colName, collector := range columnStats {
			// 找到列索引
			colIdx := -1
			for i, field := range record.Schema().Fields() {
				if field.Name == colName {
					colIdx = i
					break
				}
			}
			if colIdx >= 0 {
				collector.Collect(record.Column(colIdx))
			}
		}
	}

	if err := iterator.Err(); err != nil {
		return fmt.Errorf("iterator error: %w", err)
	}

	// 构建统计记录
	timestamp := time.Now().Unix()
	for _, colName := range columns {
		collector, ok := columnStats[colName]
		if !ok {
			continue
		}

		stats := collector.Finalize()
		builder.Field(0).(*array.StringBuilder).Append(tableID)
		builder.Field(1).(*array.StringBuilder).Append(colName)
		builder.Field(2).(*array.StringBuilder).Append(stats.DataType)
		builder.Field(3).(*array.StringBuilder).Append(stats.MinValue)
		builder.Field(4).(*array.StringBuilder).Append(stats.MaxValue)
		builder.Field(5).(*array.Int64Builder).Append(stats.NullCount)
		builder.Field(6).(*array.Int64Builder).Append(stats.DistinctCount)
		builder.Field(7).(*array.Int64Builder).Append(timestamp)
	}

	record := builder.NewRecord()
	defer record.Release()

	// 写入sys.column_statistics
	if err := engine.Write(ctx, "sys", "column_statistics", record); err != nil {
		return fmt.Errorf("failed to write column statistics: %w", err)
	}

	logger.WithComponent("executor").Info("Column statistics collected",
		zap.String("table_id", tableID),
		zap.Int("column_count", len(columns)))

	return nil
}

// ColumnStats 列统计信息
type ColumnStats struct {
	DataType      string
	MinValue      string
	MaxValue      string
	NullCount     int64
	DistinctCount int64
}

// ColumnStatsCollector 列统计收集器
type ColumnStatsCollector struct {
	dataType       arrow.DataType
	minValue       interface{}
	maxValue       interface{}
	nullCount      int64
	distinctValues map[string]bool
}

// NewColumnStatsCollector 创建列统计收集器
func NewColumnStatsCollector(dataType arrow.DataType) *ColumnStatsCollector {
	return &ColumnStatsCollector{
		dataType:       dataType,
		distinctValues: make(map[string]bool),
	}
}

// Collect 收集列的统计信息
func (c *ColumnStatsCollector) Collect(column arrow.Array) {
	switch arr := column.(type) {
	case *array.Int64:
		for i := 0; i < arr.Len(); i++ {
			if arr.IsNull(i) {
				c.nullCount++
				continue
			}
			val := arr.Value(i)
			c.distinctValues[fmt.Sprintf("%d", val)] = true

			if c.minValue == nil || val < c.minValue.(int64) {
				c.minValue = val
			}
			if c.maxValue == nil || val > c.maxValue.(int64) {
				c.maxValue = val
			}
		}
	case *array.String:
		for i := 0; i < arr.Len(); i++ {
			if arr.IsNull(i) {
				c.nullCount++
				continue
			}
			val := arr.Value(i)
			c.distinctValues[val] = true

			if c.minValue == nil || val < c.minValue.(string) {
				c.minValue = val
			}
			if c.maxValue == nil || val > c.maxValue.(string) {
				c.maxValue = val
			}
		}
	}
}

// Finalize 完成统计收集并返回结果
func (c *ColumnStatsCollector) Finalize() ColumnStats {
	minStr := ""
	maxStr := ""
	if c.minValue != nil {
		minStr = fmt.Sprintf("%v", c.minValue)
	}
	if c.maxValue != nil {
		maxStr = fmt.Sprintf("%v", c.maxValue)
	}

	return ColumnStats{
		DataType:      c.dataType.String(),
		MinValue:      minStr,
		MaxValue:      maxStr,
		NullCount:     c.nullCount,
		DistinctCount: int64(len(c.distinctValues)),
	}
}

// getRowCount 获取表的行数
func (e *ExecutorImpl) getRowCount(engine *storage.ParquetEngine, dbName, tableName string) (int64, error) {
	ctx := context.Background()
	iterator, err := engine.Scan(ctx, dbName, tableName, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to scan table: %w", err)
	}

	var rowCount int64
	for iterator.Next() {
		record := iterator.Record()
		rowCount += record.NumRows()
	}

	if err := iterator.Err(); err != nil {
		return 0, fmt.Errorf("iterator error: %w", err)
	}

	return rowCount, nil
}

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
		zap.String("plan_type", string(plan.Type)),
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
	case optimizer.DropTablePlan:
		logger.WithComponent("executor").Debug("Executing DROP TABLE plan")
		result, err := e.executeDropTable(plan, sess)
		e.logExecutionResult("DROP TABLE", start, err)
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
	case optimizer.AnalyzePlan:
		logger.WithComponent("executor").Debug("Executing ANALYZE TABLE plan")
		result, err := e.executeAnalyze(plan, sess)
		e.logExecutionResult("ANALYZE", start, err)
		return result, err
	case optimizer.TransactionPlan:
		logger.WithComponent("executor").Debug("Executing TRANSACTION plan")
		result, err := e.executeTransaction(plan, sess)
		e.logExecutionResult("TRANSACTION", start, err)
		return result, err
	case optimizer.UsePlan:
		logger.WithComponent("executor").Debug("Executing USE DATABASE plan")
		result, err := e.executeUseDatabase(plan, sess)
		e.logExecutionResult("USE", start, err)
		return result, err
	case optimizer.ExplainPlan:
		logger.WithComponent("executor").Debug("Executing EXPLAIN plan")
		result, err := e.executeExplain(plan, sess)
		e.logExecutionResult("EXPLAIN", start, err)
		return result, err
	}

	logger.WithComponent("executor").Debug("Executing query plan with operator tree",
		zap.String("plan_type", string(plan.Type)))

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
			zap.String("plan_type", string(plan.Type)),
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
			zap.String("plan_type", string(plan.Type)),
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
				zap.String("plan_type", string(plan.Type)),
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
			zap.String("plan_type", string(plan.Type)),
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
		zap.String("plan_type", string(plan.Type)),
		zap.Int("result_batches", len(batches)),
		zap.Int("result_columns", len(headers)),
		zap.Duration("total_duration", totalDuration),
		zap.Duration("result_building_duration", time.Since(resultStart)))

	return result, nil
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
		// 从上下文中获取当前数据库
		currentDB := ctx.Session.CurrentDB
		if currentDB == "" {
			currentDB = "default"
		}

		// 检查表名是否已经包含数据库限定符 (如 "sys.table_name")
		tableName := props.Table
		dbName := currentDB
		if strings.Contains(tableName, ".") {
			// 表名已经限定数据库，分割并使用指定的数据库
			parts := strings.SplitN(tableName, ".", 2)
			if len(parts) == 2 {
				dbName = parts[0]
				tableName = parts[1]
			}
		}

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
		// For now, create a simple limit operator (can be implemented later)
		// Return child for basic functionality
		_ = props // avoid unused variable error
		return child, nil

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
			// 优先使用别名
			if col.Alias != "" {
				columns[i] = col.Alias
			} else if col.Table != "" {
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

		// 检查表名是否已经包含数据库限定符 (如 "sys.table_name")
		tableName := tableScanProps.Table
		dbName := currentDB
		if strings.Contains(tableName, ".") {
			// 表名已经限定数据库，分割并使用指定的数据库
			parts := strings.SplitN(tableName, ".", 2)
			if len(parts) == 2 {
				dbName = parts[0]
				tableName = parts[1]
			}
		}

		if tableMeta, err := e.catalog.GetTable(dbName, tableName); err == nil {
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
		// 从列定义中获取类型 - 使用col.Type而不是col.Name
		dataType := e.convertSQLTypeToArrow(col.Type)
		fields[i] = arrow.Field{
			Name: col.Name,
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

	// 处理多行INSERT或单行INSERT
	if len(props.Rows) > 0 {
		// 多行INSERT - 逐行插入
		for _, row := range props.Rows {
			values := make([]interface{}, len(row))
			for i, expr := range row {
				if lit, ok := expr.(*optimizer.LiteralValue); ok {
					values[i] = lit.Value
				}
			}

			// 插入每一行
			err := e.dataManager.InsertData(currentDB, props.Table, columns, values)
			if err != nil {
				return nil, fmt.Errorf("failed to insert row: %w", err)
			}
		}
	} else {
		// 单行INSERT（向后兼容）
		values := make([]interface{}, len(props.Values))
		for i, expr := range props.Values {
			if lit, ok := expr.(*optimizer.LiteralValue); ok {
				values[i] = lit.Value
			}
		}

		err := e.dataManager.InsertData(currentDB, props.Table, columns, values)
		if err != nil {
			return nil, err
		}
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

	// 使用会话中的当前数据库
	currentDB := sess.CurrentDB
	if currentDB == "" {
		currentDB = "default"
	}

	// Convert WHERE condition to storage filters
	var filters []storage.Filter
	if props.Where != nil {
		filters = e.convertWhereToFilters(props.Where)
	}

	// Check if any assignments contain expressions (not just literals)
	hasExpressions := false
	for column, expr := range props.Assignments {
		// Log the type of each assignment
		logger.Info("Checking assignment type",
			zap.String("column", column),
			zap.String("type", fmt.Sprintf("%T", expr)))

		// Check if it's a literal value
		isLiteral := false
		switch expr.(type) {
		case *optimizer.LiteralValue, *parser.StringLiteral, *parser.IntegerLiteral, *parser.FloatLiteral, *parser.BooleanLiteral:
			isLiteral = true
		}
		if !isLiteral {
			hasExpressions = true
			logger.Info("Found expression in UPDATE", zap.String("column", column))
		}
	}

	// For expression-based updates, we need to read data, evaluate expressions, and write back
	if hasExpressions {
		return e.executeUpdateWithExpressions(currentDB, props.Table, props.Assignments, filters, sess)
	}

	// For simple literal updates, use the fast path
	assignments := make(map[string]interface{})
	for column, expr := range props.Assignments {
		if litExpr, ok := expr.(*optimizer.LiteralValue); ok {
			assignments[column] = litExpr.Value
		} else if strLit, ok := expr.(*parser.StringLiteral); ok {
			assignments[column] = strLit.Value
		} else if intLit, ok := expr.(*parser.IntegerLiteral); ok {
			assignments[column] = intLit.Value
		} else if floatLit, ok := expr.(*parser.FloatLiteral); ok {
			assignments[column] = floatLit.Value
		} else if boolLit, ok := expr.(*parser.BooleanLiteral); ok {
			assignments[column] = boolLit.Value
		}
	}

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

// executeUpdateWithExpressions handles UPDATE statements with expressions (e.g., price = price * 1.1)
func (e *ExecutorImpl) executeUpdateWithExpressions(dbName, tableName string, assignments map[string]interface{}, filters []storage.Filter, sess *session.Session) (*ResultSet, error) {
	ctx := context.Background()
	storageEngine := e.catalog.GetStorageEngine()

	logger.Info("Executing UPDATE with expressions",
		zap.String("table", fmt.Sprintf("%s.%s", dbName, tableName)),
		zap.Int("assignments", len(assignments)),
		zap.Int("filters", len(filters)))

	// Read current data WITHOUT filters - we need original values to evaluate expressions
	// We'll apply filters in-memory after reading
	iter, err := storageEngine.Scan(ctx, dbName, tableName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to scan table for update: %w", err)
	}
	defer iter.Close()

	// Get table schema
	schema, err := storageEngine.GetTableSchema(dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get table schema: %w", err)
	}

	// Create column name to index mapping
	colNameToIdx := make(map[string]int)
	for i, field := range schema.Fields() {
		colNameToIdx[field.Name] = i
	}

	updatedCount := int64(0)

	// Process each batch
	for iter.Next() {
		record := iter.Record()
		if record == nil {
			continue
		}

		logger.Info("Processing batch for expression UPDATE",
			zap.Int64("rows", record.NumRows()))

		// For each row, evaluate expressions and update
		for rowIdx := int64(0); rowIdx < record.NumRows(); rowIdx++ {
			// Check if row matches WHERE clause filters
			if len(filters) > 0 {
				matches := e.rowMatchesFilters(record, int(rowIdx), filters, colNameToIdx)
				if !matches {
					continue // Skip this row
				}
			}

			// Build update map for this specific row
			rowUpdates := make(map[string]interface{})

			for column, expr := range assignments {
				// Evaluate the expression for this row
				value, err := e.evaluateExpression(expr, record, int(rowIdx), colNameToIdx)
				if err != nil {
					return nil, fmt.Errorf("failed to evaluate expression for column %s: %w", column, err)
				}
				rowUpdates[column] = value
				logger.Info("Evaluated expression",
					zap.String("column", column),
					zap.Any("value", value),
					zap.Int64("row", rowIdx))
			}

			// Create a filter for this specific row using primary key or all columns
			// For simplicity, create filter using first column (usually id)
			var rowFilter []storage.Filter
			if len(schema.Fields()) > 0 {
				firstCol := schema.Field(0)
				firstColValue := e.getColumnValue(record, 0, int(rowIdx))
				if firstColValue != nil {
					rowFilter = append(rowFilter, storage.Filter{
						Column:   firstCol.Name,
						Operator: "=",
						Value:    firstColValue,
					})
				}
			}

			// Perform update for this specific row
			logger.Info("Applying update to row",
				zap.String("table", tableName),
				zap.Any("updates", rowUpdates),
				zap.Any("filter", rowFilter))

			if err := e.dataManager.UpdateDataWithFilters(dbName, tableName, rowUpdates, rowFilter); err != nil {
				logger.Error("Failed to update row",
					zap.Error(err),
					zap.Any("updates", rowUpdates))
				return nil, fmt.Errorf("failed to update row: %w", err)
			}
			updatedCount++
		}
	}

	logger.Info("Expression UPDATE completed",
		zap.Int64("updated_count", updatedCount))

	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// evaluateExpression evaluates an expression for a specific row
func (e *ExecutorImpl) evaluateExpression(expr interface{}, record arrow.Record, rowIdx int, colNameToIdx map[string]int) (interface{}, error) {
	switch exprNode := expr.(type) {
	case *optimizer.LiteralValue:
		logger.Debug("Evaluating LiteralValue", zap.Any("value", exprNode.Value))
		return exprNode.Value, nil
	case *parser.StringLiteral:
		return exprNode.Value, nil
	case *parser.IntegerLiteral:
		logger.Debug("Evaluating IntegerLiteral", zap.Int64("value", exprNode.Value))
		return exprNode.Value, nil
	case *parser.FloatLiteral:
		logger.Debug("Evaluating FloatLiteral", zap.Float64("value", exprNode.Value))
		return exprNode.Value, nil
	case *parser.BooleanLiteral:
		return exprNode.Value, nil
	case *parser.BinaryExpr:
		// Handle binary expressions like "price * 1.1" or "quantity + 10"
		logger.Info("Evaluating BinaryExpr", zap.String("operator", exprNode.Operator))
		left, err := e.evaluateExpression(exprNode.Left, record, rowIdx, colNameToIdx)
		if err != nil {
			return nil, err
		}
		logger.Info("BinaryExpr left evaluated", zap.Any("value", left))

		right, err := e.evaluateExpression(exprNode.Right, record, rowIdx, colNameToIdx)
		if err != nil {
			return nil, err
		}
		logger.Info("BinaryExpr right evaluated", zap.Any("value", right))

		// Perform the operation
		result, err := e.performBinaryOperation(left, right, exprNode.Operator)
		logger.Info("BinaryExpr result", zap.Any("result", result), zap.Error(err))
		return result, err
	case *parser.ColumnRef:
		// Get the current value of the column for this row
		colIdx, ok := colNameToIdx[exprNode.Column]
		if !ok {
			return nil, fmt.Errorf("column %s not found", exprNode.Column)
		}
		value := e.getColumnValue(record, colIdx, rowIdx)
		logger.Info("Evaluating ColumnRef",
			zap.String("column", exprNode.Column),
			zap.Int("colIdx", colIdx),
			zap.Int("rowIdx", rowIdx),
			zap.Any("value", value))
		return value, nil
	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// performBinaryOperation performs arithmetic operations
func (e *ExecutorImpl) performBinaryOperation(left, right interface{}, operator string) (interface{}, error) {
	// Convert to float64 for arithmetic operations
	leftFloat, leftOk := e.toFloat64(left)
	rightFloat, rightOk := e.toFloat64(right)

	if !leftOk || !rightOk {
		return nil, fmt.Errorf("cannot perform arithmetic on non-numeric values")
	}

	switch operator {
	case "+":
		return leftFloat + rightFloat, nil
	case "-":
		return leftFloat - rightFloat, nil
	case "*":
		return leftFloat * rightFloat, nil
	case "/":
		if rightFloat == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return leftFloat / rightFloat, nil
	default:
		return nil, fmt.Errorf("unsupported operator: %s", operator)
	}
}

// toFloat64 converts a value to float64
func (e *ExecutorImpl) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case float32:
		return float64(v), true
	default:
		return 0, false
	}
}

// rowMatchesFilters checks if a row matches all WHERE clause filters
func (e *ExecutorImpl) rowMatchesFilters(record arrow.Record, rowIdx int, filters []storage.Filter, colNameToIdx map[string]int) bool {
	for _, filter := range filters {
		colIdx, exists := colNameToIdx[filter.Column]
		if !exists {
			return false
		}

		colValue := e.getColumnValue(record, colIdx, rowIdx)
		filterValue := filter.Value

		// Compare values based on operator
		switch filter.Operator {
		case "=":
			if !e.valuesEqual(colValue, filterValue) {
				return false
			}
		case ">":
			if !e.valueGreater(colValue, filterValue) {
				return false
			}
		case "<":
			if !e.valueLess(colValue, filterValue) {
				return false
			}
		case ">=":
			if !e.valueGreaterOrEqual(colValue, filterValue) {
				return false
			}
		case "<=":
			if !e.valueLessOrEqual(colValue, filterValue) {
				return false
			}
		case "!=":
			if e.valuesEqual(colValue, filterValue) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// Helper comparison functions
func (e *ExecutorImpl) valuesEqual(a, b interface{}) bool {
	if a == nil || b == nil {
		return a == b
	}
	// Try numeric comparison
	if aFloat, ok := e.toFloat64(a); ok {
		if bFloat, ok := e.toFloat64(b); ok {
			return aFloat == bFloat
		}
	}
	// String comparison
	return fmt.Sprint(a) == fmt.Sprint(b)
}

func (e *ExecutorImpl) valueGreater(a, b interface{}) bool {
	if aFloat, ok := e.toFloat64(a); ok {
		if bFloat, ok := e.toFloat64(b); ok {
			return aFloat > bFloat
		}
	}
	return false
}

func (e *ExecutorImpl) valueLess(a, b interface{}) bool {
	if aFloat, ok := e.toFloat64(a); ok {
		if bFloat, ok := e.toFloat64(b); ok {
			return aFloat < bFloat
		}
	}
	return false
}

func (e *ExecutorImpl) valueGreaterOrEqual(a, b interface{}) bool {
	if aFloat, ok := e.toFloat64(a); ok {
		if bFloat, ok := e.toFloat64(b); ok {
			return aFloat >= bFloat
		}
	}
	return false
}

func (e *ExecutorImpl) valueLessOrEqual(a, b interface{}) bool {
	if aFloat, ok := e.toFloat64(a); ok {
		if bFloat, ok := e.toFloat64(b); ok {
			return aFloat <= bFloat
		}
	}
	return false
}

// getColumnValue retrieves a column value from a record at a specific row
func (e *ExecutorImpl) getColumnValue(record arrow.Record, colIdx, rowIdx int) interface{} {
	column := record.Column(colIdx)
	switch col := column.(type) {
	case *array.Int64:
		if col.IsNull(rowIdx) {
			return nil
		}
		return col.Value(rowIdx)
	case *array.Float64:
		if col.IsNull(rowIdx) {
			return nil
		}
		return col.Value(rowIdx)
	case *array.String:
		if col.IsNull(rowIdx) {
			return nil
		}
		return col.Value(rowIdx)
	case *array.Boolean:
		if col.IsNull(rowIdx) {
			return nil
		}
		return col.Value(rowIdx)
	default:
		return nil
	}
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

	// 使用 DataManager 删除数据
	// 使用会话中的当前数据库
	currentDB := sess.CurrentDB
	if currentDB == "" {
		currentDB = "default"
	}

	// Convert WHERE condition to storage filters
	var filters []storage.Filter
	if props.Where != nil {
		filters = e.convertWhereToFilters(props.Where)
	}

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

// executeDropTable 执行删除表操作
func (e *ExecutorImpl) executeDropTable(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props := plan.Properties.(*optimizer.DropTableProperties)

	// 使用会话中的当前数据库
	currentDB := sess.CurrentDB
	if currentDB == "" {
		currentDB = "default"
	}

	// 从catalog中删除表元数据
	err := e.catalog.DropTable(currentDB, props.Table)
	if err != nil {
		return nil, fmt.Errorf("failed to drop table: %w", err)
	}

	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// executeCreateIndex 执行创建索引操作
func (e *ExecutorImpl) executeCreateIndex(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props := plan.Properties.(*optimizer.CreateIndexProperties)

	// 使用会话中的当前数据库
	currentDB := sess.CurrentDB
	if currentDB == "" {
		currentDB = "default"
	}

	// 构建IndexMeta
	indexMeta := catalog.IndexMeta{
		Database:  currentDB,
		Table:     props.Table,
		Name:      props.Name,
		Columns:   props.Columns,
		IsUnique:  props.IsUnique,
		IndexType: "BTREE", // 默认使用BTREE索引
	}

	// 调用catalog创建索引
	err := e.catalog.CreateIndex(indexMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// executeDropIndex 执行删除索引操作
func (e *ExecutorImpl) executeDropIndex(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props := plan.Properties.(*optimizer.DropIndexProperties)

	// 使用会话中的当前数据库
	currentDB := sess.CurrentDB
	if currentDB == "" {
		currentDB = "default"
	}

	// 调用catalog删除索引
	err := e.catalog.DropIndex(currentDB, props.Table, props.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to drop index: %w", err)
	}

	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// executeAnalyze 执行ANALYZE TABLE命令
func (e *ExecutorImpl) executeAnalyze(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props := plan.Properties.(*optimizer.AnalyzeProperties)

	// 使用会话中的当前数据库
	currentDB := sess.CurrentDB
	if currentDB == "" {
		currentDB = "default"
	}

	// TODO: 实际实现统计信息收集
	// 当前作为占位符，返回成功
	logger.WithComponent("executor").Info("Analyzing table statistics",
		zap.String("database", currentDB),
		zap.String("table", props.Table),
		zap.Strings("columns", props.Columns))

	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// executeTransaction 执行事务控制命令 (START TRANSACTION, COMMIT, ROLLBACK)
func (e *ExecutorImpl) executeTransaction(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props := plan.Properties.(*optimizer.TransactionProperties)

	// TODO: 实现完整的事务控制
	// 当前作为占位符，记录事务操作
	logger.WithComponent("executor").Info("Transaction operation",
		zap.String("type", props.Type),
		zap.Int64("session_id", sess.ID))

	// 根据事务类型更新会话状态 (未来可用于事务管理)
	switch props.Type {
	case "BEGIN", "START":
		// 标记事务开始
		logger.WithComponent("executor").Debug("Transaction started", zap.Int64("session_id", sess.ID))
	case "COMMIT":
		// 提交事务
		logger.WithComponent("executor").Debug("Transaction committed", zap.Int64("session_id", sess.ID))
	case "ROLLBACK":
		// 回滚事务
		logger.WithComponent("executor").Debug("Transaction rolled back", zap.Int64("session_id", sess.ID))
	}

	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// executeUseDatabase 执行USE DATABASE命令
func (e *ExecutorImpl) executeUseDatabase(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props := plan.Properties.(*optimizer.UseProperties)

	// 检查数据库是否存在
	_, err := e.catalog.GetDatabase(props.Database)
	if err != nil {
		return nil, fmt.Errorf("database '%s' does not exist", props.Database)
	}

	// 更新会话的当前数据库
	sess.CurrentDB = props.Database

	logger.WithComponent("executor").Info("Database switched",
		zap.String("database", props.Database),
		zap.Int64("session_id", sess.ID))

	return &ResultSet{
		Headers: []string{"status"},
		rows:    []*types.Batch{},
		curRow:  -1,
	}, nil
}

// executeExplain 执行EXPLAIN命令
func (e *ExecutorImpl) executeExplain(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	props := plan.Properties.(*optimizer.ExplainProperties)

	// 生成执行计划的文本表示
	planText := e.explainPlan(props.Query, 0)

	// 创建包含执行计划的结果集
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "Query Plan", Type: arrow.BinaryTypes.String},
	}, nil)

	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 添加执行计划文本
	strBuilder := builder.Field(0).(*array.StringBuilder)
	strBuilder.Append(planText)

	record := builder.NewRecord()
	batch := types.NewBatch(record)

	return &ResultSet{
		Headers: []string{"Query Plan"},
		rows:    []*types.Batch{batch},
		curRow:  -1,
	}, nil
}

// explainPlan 生成执行计划的文本表示
func (e *ExecutorImpl) explainPlan(plan *optimizer.Plan, depth int) string {
	if plan == nil {
		return ""
	}

	indent := strings.Repeat("  ", depth)
	result := indent + plan.Type.String()

	if plan.Properties != nil {
		if explainable, ok := plan.Properties.(optimizer.PlanProperties); ok {
			result += "\n" + indent + "  " + explainable.Explain()
		}
	}

	for _, child := range plan.Children {
		childPlan := e.explainPlan(child, depth+1)
		if childPlan != "" {
			result += "\n" + childPlan
		}
	}

	return result
}

// executeShow 执行SHOW命令
func (e *ExecutorImpl) executeShow(plan *optimizer.Plan, sess *session.Session) (*ResultSet, error) {
	// 首先检查是否是 SHOW INDEXES
	if indexProps, ok := plan.Properties.(*optimizer.ShowIndexesProperties); ok {
		return e.executeShowIndexes(indexProps, sess)
	}

	// 否则处理常规 SHOW 命令
	props := plan.Properties.(*optimizer.ShowProperties)

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

// executeShowIndexes 执行SHOW INDEXES命令
func (e *ExecutorImpl) executeShowIndexes(props *optimizer.ShowIndexesProperties, sess *session.Session) (*ResultSet, error) {
	// 获取当前数据库名
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
			Headers: []string{"Index Name", "Table", "Columns", "Unique"},
			rows:    []*types.Batch{},
			curRow:  -1,
		}, nil
	}

	// 创建Arrow schema
	schema := arrow.NewSchema([]arrow.Field{
		{Name: "Index Name", Type: arrow.BinaryTypes.String},
		{Name: "Table", Type: arrow.BinaryTypes.String},
		{Name: "Columns", Type: arrow.BinaryTypes.String},
		{Name: "Unique", Type: arrow.BinaryTypes.String},
	}, nil)

	// 创建Arrow记录
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 填充索引数据
	for _, index := range indexes {
		builder.Field(0).(*array.StringBuilder).Append(index.Name)
		builder.Field(1).(*array.StringBuilder).Append(index.Table)
		builder.Field(2).(*array.StringBuilder).Append(strings.Join(index.Columns, ","))
		if index.IsUnique {
			builder.Field(3).(*array.StringBuilder).Append("YES")
		} else {
			builder.Field(3).(*array.StringBuilder).Append("NO")
		}
	}

	// 构建记录
	record := builder.NewRecord()
	defer record.Release()

	// 创建批次
	batch := types.NewBatch(record)

	return &ResultSet{
		Headers: []string{"Index Name", "Table", "Columns", "Unique"},
		rows:    []*types.Batch{batch},
		curRow:  -1,
	}, nil
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

// convertWhereToFilters 将WHERE条件表达式转换为storage.Filter数组
func (e *ExecutorImpl) convertWhereToFilters(whereExpr interface{}) []storage.Filter {
	var filters []storage.Filter

	// 处理BinaryExpr（如 price > 100, id = 1）
	if binExpr, ok := whereExpr.(*parser.BinaryExpr); ok {
		filter := e.binaryExprToFilter(binExpr)
		if filter != nil {
			filters = append(filters, *filter)
		}
	}

	// 处理InExpr（如 id IN (1, 2, 3)）
	if inExpr, ok := whereExpr.(*parser.InExpr); ok {
		filter := e.inExprToFilter(inExpr)
		if filter != nil {
			filters = append(filters, *filter)
		}
	}

	return filters
}

// binaryExprToFilter 将二元表达式转换为Filter
func (e *ExecutorImpl) binaryExprToFilter(expr *parser.BinaryExpr) *storage.Filter {
	// 提取列名
	var column string
	if colRef, ok := expr.Left.(*parser.ColumnRef); ok {
		column = colRef.Column
	} else {
		return nil // 不支持的左操作数类型
	}

	// 提取值
	var value interface{}
	if intLit, ok := expr.Right.(*parser.IntegerLiteral); ok {
		value = intLit.Value
	} else if strLit, ok := expr.Right.(*parser.StringLiteral); ok {
		value = strLit.Value
	} else if floatLit, ok := expr.Right.(*parser.FloatLiteral); ok {
		value = floatLit.Value
	} else if boolLit, ok := expr.Right.(*parser.BooleanLiteral); ok {
		value = boolLit.Value
	} else {
		return nil // 不支持的右操作数类型
	}

	return &storage.Filter{
		Column:   column,
		Operator: expr.Operator,
		Value:    value,
	}
}

// inExprToFilter 将IN表达式转换为Filter
func (e *ExecutorImpl) inExprToFilter(expr *parser.InExpr) *storage.Filter {
	// 提取列名
	var column string
	if colRef, ok := expr.Left.(*parser.ColumnRef); ok {
		column = colRef.Column
	} else {
		return nil
	}

	// 提取值列表
	var values []interface{}
	for _, val := range expr.Values {
		if intLit, ok := val.(*parser.IntegerLiteral); ok {
			values = append(values, intLit.Value)
		} else if strLit, ok := val.(*parser.StringLiteral); ok {
			values = append(values, strLit.Value)
		} else if floatLit, ok := val.(*parser.FloatLiteral); ok {
			values = append(values, floatLit.Value)
		}
	}

	// 使用已经解析的操作符（IN 或 NOT IN）
	operator := expr.Operator

	return &storage.Filter{
		Column:   column,
		Operator: operator,
		Values:   values,
	}
}

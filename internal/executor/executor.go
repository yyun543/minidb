package executor

import (
	"fmt"
	"strings"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor/operators"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/types"
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
	return &ExecutorImpl{
		catalog:     cat,
		dataManager: NewDataManager(cat),
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
	// 为 DDL/DML 操作特别处理
	switch plan.Type {
	case optimizer.CreateDatabasePlan:
		return e.executeCreateDatabase(plan, sess)
	case optimizer.DropDatabasePlan:
		return e.executeDropDatabase(plan, sess)
	case optimizer.CreateTablePlan:
		return e.executeCreateTable(plan, sess)
	case optimizer.ShowPlan:
		return e.executeShow(plan, sess)
	case optimizer.InsertPlan:
		return e.executeInsert(plan, sess)
	case optimizer.UpdatePlan:
		return e.executeUpdate(plan, sess)
	case optimizer.DeletePlan:
		return e.executeDelete(plan, sess)
	}

	// 创建执行上下文
	ctx := NewContext(e.catalog, sess, e.dataManager)

	// 构建执行算子树
	op, err := e.buildOperator(plan, ctx)
	if err != nil {
		return nil, err
	}

	// 初始化算子
	if err := op.Init(ctx); err != nil {
		return nil, err
	}

	// 执行查询并收集结果
	var batches []*types.Batch
	for {
		batch, err := op.Next()
		if err != nil {
			return nil, err
		}
		if batch == nil {
			break
		}
		batches = append(batches, batch)
	}

	// 关闭算子
	if err := op.Close(); err != nil {
		return nil, err
	}

	// 构建结果集
	headers := e.getResultHeaders(plan, sess)
	return &ResultSet{
		Headers: headers,
		rows:    batches,
		curRow:  -1,
	}, nil
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
		return operators.NewTableScan(currentDB, props.Table, e.catalog), nil

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
			// 从子计划（通常是TableScan）获取schema
			if len(plan.Children) > 0 && plan.Children[0].Type == optimizer.TableScanPlan {
				tableScanProps := plan.Children[0].Properties.(*optimizer.TableScanProperties)
				// 从catalog获取表的schema
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

	// 创建WHERE条件函数，支持动态条件评估
	var whereCondition func(arrow.Record, int) bool
	if props.Where != nil {
		whereCondition = func(record arrow.Record, rowIdx int) bool {
			return e.evaluateWhereCondition(record, rowIdx, props.Where)
		}
	}

	// 使用 DataManager 更新数据
	// 使用会话中的当前数据库
	currentDB := sess.CurrentDB
	if currentDB == "" {
		currentDB = "default"
	}

	err := e.dataManager.UpdateData(currentDB, props.Table, assignments, whereCondition)
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

	// 创建WHERE条件函数，支持动态条件评估
	var whereCondition func(arrow.Record, int) bool
	if props.Where != nil {
		whereCondition = func(record arrow.Record, rowIdx int) bool {
			return e.evaluateWhereCondition(record, rowIdx, props.Where)
		}
	}

	// 使用 DataManager 删除数据
	// 使用会话中的当前数据库
	currentDB := sess.CurrentDB
	if currentDB == "" {
		currentDB = "default"
	}

	err := e.dataManager.DeleteData(currentDB, props.Table, whereCondition)
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

package optimizer

import (
	"fmt"
	"strings"
	"time"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/yyun543/minidb/internal/logger"
	"github.com/yyun543/minidb/internal/parser"
	"go.uber.org/zap"
)

// CatalogInterface 定义optimizer需要的catalog接口
// 这是一个最小接口,避免circular dependency
type CatalogInterface interface {
	GetTable(database, table string) (TableMeta, error)
}

// TableMeta 表元数据(最小定义)
type TableMeta struct {
	Database string
	Table    string
	Schema   *arrow.Schema
}

// Optimizer 查询优化器
type Optimizer struct {
	rules           []Rule           // 优化规则
	catalog         CatalogInterface // catalog接口用于元数据访问(可选)
	currentDatabase string           // 当前数据库名(用于SELECT *展开)
}

// NewOptimizer 创建新的优化器实例
func NewOptimizer() *Optimizer {
	logger.WithComponent("optimizer").Info("Creating new query optimizer")

	opt := &Optimizer{
		rules:           make([]Rule, 0),
		catalog:         nil, // 可选,稍后通过SetCatalog设置
		currentDatabase: "",
	}
	// 添加优化规则
	opt.rules = append(opt.rules,
		&PredicatePushDownRule{}, // 谓词下推
		&JoinReorderRule{},       // Join重排序
		&ProjectionPruningRule{}, // 投影剪枝
	)

	logger.WithComponent("optimizer").Info("Query optimizer created successfully",
		zap.Int("rules_count", len(opt.rules)))

	return opt
}

// SetCatalog 设置catalog接口(可选)
// 设置后optimizer可以在优化时进行SELECT *展开等元数据依赖的优化
func (o *Optimizer) SetCatalog(catalog CatalogInterface) {
	o.catalog = catalog
	logger.WithComponent("optimizer").Debug("Catalog set for optimizer")
}

// SetCurrentDatabase 设置当前数据库
// 用于SELECT *展开时解析表名
func (o *Optimizer) SetCurrentDatabase(database string) {
	o.currentDatabase = database
	logger.WithComponent("optimizer").Debug("Current database set for optimizer",
		zap.String("database", database))
}

// Optimize 优化查询
func (o *Optimizer) Optimize(stmt parser.Node) (*Plan, error) {
	logger.WithComponent("optimizer").Debug("Starting query optimization",
		zap.String("statement_type", fmt.Sprintf("%T", stmt)))

	start := time.Now()

	// 1. 构建初始计划
	buildStart := time.Now()
	plan, err := o.buildPlan(stmt)
	if err != nil {
		logger.WithComponent("optimizer").Error("Failed to build initial plan",
			zap.String("statement_type", fmt.Sprintf("%T", stmt)),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))
		return nil, err
	}
	logger.WithComponent("optimizer").Debug("Initial plan built successfully",
		zap.String("plan_type", plan.Type.String()),
		zap.Duration("build_duration", time.Since(buildStart)))

	// 2. 应用优化规则
	rulesStart := time.Now()
	appliedRules := 0
	for _, rule := range o.rules {
		ruleStart := time.Now()
		originalPlan := plan
		plan = rule.Apply(plan)
		if plan != originalPlan {
			appliedRules++
			logger.WithComponent("optimizer").Debug("Optimization rule applied",
				zap.String("rule_type", fmt.Sprintf("%T", rule)),
				zap.Duration("rule_duration", time.Since(ruleStart)))
		}
	}

	totalDuration := time.Since(start)
	logger.WithComponent("optimizer").Info("Query optimization completed",
		zap.String("statement_type", fmt.Sprintf("%T", stmt)),
		zap.String("final_plan_type", plan.Type.String()),
		zap.Int("rules_applied", appliedRules),
		zap.Int("total_rules", len(o.rules)),
		zap.Duration("rules_duration", time.Since(rulesStart)),
		zap.Duration("total_duration", totalDuration))

	return plan, nil
}

// buildPlan 根据AST节点构建查询计划
func (o *Optimizer) buildPlan(node parser.Node) (*Plan, error) {
	switch n := node.(type) {
	case *parser.SelectStmt:
		return o.buildSelectPlan(n)
	case *parser.InsertStmt:
		return o.buildInsertPlan(n)
	case *parser.UpdateStmt:
		return o.buildUpdatePlan(n)
	case *parser.DeleteStmt:
		return o.buildDeletePlan(n)
	case *parser.CreateDatabaseStmt:
		return o.buildCreateDatabasePlan(n)
	case *parser.CreateTableStmt:
		return o.buildCreateTablePlan(n)
	case *parser.CreateIndexStmt:
		return o.buildCreateIndexPlan(n)
	case *parser.DropIndexStmt:
		return o.buildDropIndexPlan(n)
	case *parser.DropDatabaseStmt:
		return o.buildDropDatabasePlan(n)
	case *parser.DropTableStmt:
		return o.buildDropTablePlan(n)
	case *parser.TransactionStmt:
		return o.buildTransactionPlan(n)
	case *parser.UseStmt:
		return o.buildUsePlan(n)
	case *parser.ShowDatabasesStmt:
		return o.buildShowDatabasesPlan(n)
	case *parser.ShowTablesStmt:
		return o.buildShowTablesPlan(n)
	case *parser.ShowIndexesStmt:
		return o.buildShowIndexesPlan(n)
	case *parser.ExplainStmt:
		return o.buildExplainPlan(n)
	case *parser.AnalyzeStmt:
		return o.buildAnalyzePlan(n)
	default:
		return nil, fmt.Errorf("unsupported statement type: %T", node)
	}
}

// buildSelectPlan 根据AST构建初始查询计划
func (o *Optimizer) buildSelectPlan(stmt *parser.SelectStmt) (*Plan, error) {
	// 1. 创建投影算子
	projectPlan := NewPlan(SelectPlan)

	// 检查是否为 SELECT * 查询
	// 如果没有指定列，或者指定了"*"，则认为是SELECT *
	isSelectAll := len(stmt.Columns) == 0 || (len(stmt.Columns) == 1 && stmt.Columns[0].Column == "*")

	var columns []ColumnRef
	if isSelectAll && o.catalog != nil && stmt.From != "" {
		// P0.2: 如果有catalog且是SELECT *,尝试在优化时展开
		// 这使得投影裁剪优化能够生效,减少I/O
		columns = o.expandSelectAll(stmt.From)
		// 如果展开成功(不只是返回星号占位符),则不再标记为All
		if len(columns) > 0 && columns[0].Column != "*" {
			isSelectAll = false
			logger.WithComponent("optimizer").Debug("SELECT * expanded during optimization",
				zap.String("table", stmt.From),
				zap.Int("columns", len(columns)))
		} else {
			columns = convertSelectItems(stmt.Columns)
		}
	} else {
		columns = convertSelectItems(stmt.Columns)
	}

	projectPlan.Properties = &SelectProperties{
		All:     isSelectAll,
		Columns: columns,
	}

	// 2. 构建表扫描和JOIN
	var currentPlan *Plan
	if stmt.FromSubquery != nil {
		// FROM 子查询：递归优化子查询
		subqueryPlan, err := o.Optimize(stmt.FromSubquery)
		if err != nil {
			return nil, fmt.Errorf("failed to optimize FROM subquery: %w", err)
		}
		// 子查询作为数据源直接使用
		currentPlan = subqueryPlan
	} else if stmt.From != "" {
		if len(stmt.Joins) > 0 {
			currentPlan = o.buildJoinPlan(stmt.From, stmt.FromAlias, stmt.Joins)
		} else {
			currentPlan = NewPlan(TableScanPlan)
			currentPlan.Properties = &TableScanProperties{
				Table:      stmt.From,
				TableAlias: stmt.FromAlias,
			}
		}
	}

	// 3. 构建WHERE过滤
	if stmt.Where != nil {
		filterPlan := NewPlan(FilterPlan)
		filterPlan.Properties = &FilterProperties{
			Condition: convertExpression(stmt.Where.Condition),
		}
		filterPlan.AddChild(currentPlan)
		currentPlan = filterPlan
	}

	// 4. 构建GROUP BY (包括隐式聚合 - 当有聚合函数但没有GROUP BY时)
	hasAggregates := o.hasAggregateFunction(stmt.Columns)

	if len(stmt.GroupBy) > 0 || hasAggregates {
		groupPlan := NewPlan(GroupPlan)

		// 构建GROUP BY键 (如果没有显式GROUP BY，则为空，表示全表聚合)
		groupKeys := make([]ColumnRef, len(stmt.GroupBy))
		for i, expr := range stmt.GroupBy {
			if colRef, ok := expr.(*parser.ColumnRef); ok {
				groupKeys[i] = ColumnRef{
					Column: colRef.Column,
					Table:  colRef.Table,
				}
			}
		}

		// 从SELECT子句中提取聚合表达式
		var aggregations []AggregateExpr
		var selectColumns []ColumnRef

		for _, col := range stmt.Columns {

			selectCol := ColumnRef{
				Column: col.Column,
				Table:  col.Table,
				Alias:  col.Alias,
			}

			// 检查是否是聚合函数
			if col.Expr != nil {
				if funcCall, ok := col.Expr.(*parser.FunctionCall); ok {
					funcName := strings.ToUpper(funcCall.Name)
					if isAggregateFunction(funcName) {
						aggExpr := AggregateExpr{
							Function: funcName,
							Alias:    col.Alias,
							Expr:     convertExpression(col.Expr),
						}

						// 如果函数有参数，提取第一个参数作为列名
						if len(funcCall.Args) > 0 {
							if colRef, ok := funcCall.Args[0].(*parser.ColumnRef); ok {
								aggExpr.Column = colRef.Column
							} else if funcName == "COUNT" && len(funcCall.Args) == 1 {
								// 处理 COUNT(*) 情况
								if _, isAsterisk := funcCall.Args[0].(*parser.Asterisk); isAsterisk {
									aggExpr.Column = "*"
								}
							}
						}

						aggregations = append(aggregations, aggExpr)

						// 为聚合函数设置相应的类型
						selectCol.Type = ColumnRefTypeFunction
						selectCol.FunctionName = funcName
					}
				}
			}

			selectColumns = append(selectColumns, selectCol)
		}

		groupPlan.Properties = &GroupByProperties{
			GroupKeys:     groupKeys,
			Aggregations:  aggregations,
			SelectColumns: selectColumns,
		}
		groupPlan.AddChild(currentPlan)
		currentPlan = groupPlan
	}

	// 5. 构建HAVING (必须在GROUP BY之后)
	if stmt.Having != nil {
		havingPlan := NewPlan(HavingPlan)
		havingPlan.Properties = &HavingProperties{
			Condition: convertExpression(stmt.Having.Condition),
		}
		havingPlan.AddChild(currentPlan)
		currentPlan = havingPlan
	}

	// 6. 构建ORDER BY
	if len(stmt.OrderBy) > 0 {
		orderPlan := NewPlan(OrderPlan)
		orderKeys := make([]OrderKey, len(stmt.OrderBy))
		for i, item := range stmt.OrderBy {
			if colRef, ok := item.Expr.(*parser.ColumnRef); ok {
				// 简单列引用
				orderKeys[i] = OrderKey{
					Column:     colRef.Column,
					Table:      colRef.Table,
					Direction:  item.Direction,
					Expression: nil,
				}
			} else {
				// 表达式 (如 price * quantity)
				orderKeys[i] = OrderKey{
					Column:     "",
					Table:      "",
					Direction:  item.Direction,
					Expression: convertExpression(item.Expr),
				}
			}
		}
		orderPlan.Properties = &OrderByProperties{
			OrderKeys: orderKeys,
		}
		orderPlan.AddChild(currentPlan)
		currentPlan = orderPlan
	}

	// 7. 构建LIMIT
	if stmt.Limit > 0 {
		limitPlan := NewPlan(LimitPlan)
		limitPlan.Properties = &LimitProperties{
			Limit: stmt.Limit,
		}
		limitPlan.AddChild(currentPlan)
		currentPlan = limitPlan
	}

	// 8. 如果不是SELECT *且没有GROUP BY且没有聚合函数，添加投影算子
	// GROUP BY查询或包含聚合函数的查询会自己处理列投影，不需要额外的投影操作符
	if !isSelectAll && len(stmt.GroupBy) == 0 && !hasAggregates {
		projectionPlan := NewPlan(ProjectionPlan)
		projectionPlan.Properties = &ProjectionProperties{
			Columns: columns, // 使用展开后的columns，而不是stmt.Columns
		}
		projectionPlan.AddChild(currentPlan)
		currentPlan = projectionPlan
	}

	// 9. 最后添加顶层SELECT算子
	projectPlan.AddChild(currentPlan)

	return projectPlan, nil
}

// buildJoinPlan 构建JOIN计划
func (o *Optimizer) buildJoinPlan(leftTable string, leftAlias string, joins []*parser.JoinClause) *Plan {
	// 创建左表扫描
	leftScan := NewPlan(TableScanPlan)
	leftScan.Properties = &TableScanProperties{
		Table:      leftTable,
		TableAlias: leftAlias,
	}

	currentPlan := leftScan

	// Track left and right info for the first JOIN
	currentLeft := leftTable
	currentLeftAlias := leftAlias

	// 处理每个JOIN子句
	for joinIdx, join := range joins {
		joinPlan := NewPlan(JoinPlan)

		var rightPlan *Plan
		var rightTable, rightAlias string

		// 检查右侧是子查询还是表引用
		if join.Right.Subquery != nil {
			// 右侧是子查询，递归优化子查询
			var err error
			rightPlan, err = o.Optimize(join.Right.Subquery)
			if err != nil {
				// 如果子查询优化失败，返回原计划（或处理错误）
				// 这里简化处理，使用一个简单的占位符
				rightPlan = NewPlan(TableScanPlan)
				rightPlan.Properties = &TableScanProperties{
					Table:      "",
					TableAlias: join.Right.Alias,
				}
			}
			rightTable = ""               // 子查询没有表名
			rightAlias = join.Right.Alias // 使用子查询的别名
		} else {
			// 右侧是普通表引用
			rightPlan = NewPlan(TableScanPlan)
			rightPlan.Properties = &TableScanProperties{
				Table:      join.Right.Table,
				TableAlias: join.Right.Alias,
			}
			rightTable = join.Right.Table
			rightAlias = join.Right.Alias
		}

		// 设置JOIN属性
		// For multi-table JOINs, after the first JOIN, the left side is the accumulated JOIN result
		// We keep track of this by leaving Left/LeftAlias empty for subsequent JOINs
		if joinIdx == 0 {
			// First JOIN: use the original left table
			joinPlan.Properties = &JoinProperties{
				JoinType:   join.JoinType,
				Left:       currentLeft,
				LeftAlias:  currentLeftAlias,
				Right:      rightTable,
				RightAlias: rightAlias,
				Condition:  convertExpression(join.Condition),
			}
		} else {
			// Subsequent JOINs: left side is the accumulated JOIN result (indicated by empty Left)
			joinPlan.Properties = &JoinProperties{
				JoinType:   join.JoinType,
				Left:       "", // Empty indicates this is a JOIN on accumulated result
				LeftAlias:  "",
				Right:      rightTable,
				RightAlias: rightAlias,
				Condition:  convertExpression(join.Condition),
			}
		}

		// 添加左右子节点
		joinPlan.AddChild(currentPlan)
		joinPlan.AddChild(rightPlan)

		// 更新currentPlan为当前JOIN结果，用于下一个JOIN
		currentPlan = joinPlan
	}

	return currentPlan
}

// buildCreateDatabasePlan 构建CREATE DATABASE语句的查询计划
func (o *Optimizer) buildCreateDatabasePlan(stmt *parser.CreateDatabaseStmt) (*Plan, error) {
	return &Plan{
		Type: CreateDatabasePlan,
		Properties: &CreateDatabaseProperties{
			Database: stmt.Database,
		},
	}, nil
}

// buildCreateTablePlan 构建CREATE TABLE语句的查询计划
func (o *Optimizer) buildCreateTablePlan(stmt *parser.CreateTableStmt) (*Plan, error) {
	columns := make([]ColumnDef, len(stmt.Columns))
	for i, col := range stmt.Columns {
		columns[i] = ColumnDef{
			Name: col.Name,
			Type: col.DataType, // 保存完整的数据类型
		}
	}
	return &Plan{
		Type: CreateTablePlan,
		Properties: &CreateTableProperties{
			Table:   stmt.Table,
			Columns: columns,
		},
	}, nil
}

// buildDropDatabasePlan 构建DROP DATABASE语句的查询计划
func (o *Optimizer) buildDropDatabasePlan(stmt *parser.DropDatabaseStmt) (*Plan, error) {
	return &Plan{
		Type: DropDatabasePlan,
		Properties: &DropDatabaseProperties{
			Database: stmt.Database,
		},
	}, nil
}

// buildDropTablePlan 构建DROP TABLE语句的查询计划
func (o *Optimizer) buildDropTablePlan(stmt *parser.DropTableStmt) (*Plan, error) {
	return &Plan{
		Type: DropTablePlan,
		Properties: &DropTableProperties{
			Table: stmt.Table,
		},
	}, nil
}

// buildTransactionPlan 构建事务语句的查询计划
func (o *Optimizer) buildTransactionPlan(stmt *parser.TransactionStmt) (*Plan, error) {
	return &Plan{
		Type: TransactionPlan,
		Properties: &TransactionProperties{
			Type: stmt.TxType,
		},
	}, nil
}

// buildUsePlan 构建USE DATABASE语句的查询计划
func (o *Optimizer) buildUsePlan(stmt *parser.UseStmt) (*Plan, error) {
	return &Plan{
		Type: UsePlan,
		Properties: &UseProperties{
			Database: stmt.Database,
		},
	}, nil
}

// buildShowDatabasesPlan 构建SHOW DATABASES语句的查询计划
func (o *Optimizer) buildShowDatabasesPlan(stmt *parser.ShowDatabasesStmt) (*Plan, error) {
	return &Plan{
		Type: ShowPlan,
		Properties: &ShowProperties{
			Type: "DATABASES",
		},
	}, nil
}

// buildShowTablesPlan 构建SHOW TABLES语句的查询计划
func (o *Optimizer) buildShowTablesPlan(stmt *parser.ShowTablesStmt) (*Plan, error) {
	return &Plan{
		Type: ShowPlan,
		Properties: &ShowProperties{
			Type: "TABLES",
		},
	}, nil
}

// buildExplainPlan 构建EXPLAIN语句的查询计划
func (o *Optimizer) buildExplainPlan(stmt *parser.ExplainStmt) (*Plan, error) {
	// 首先构建要解释的查询计划
	queryPlan, err := o.buildPlan(stmt.Query)
	if err != nil {
		return nil, err
	}

	return &Plan{
		Type: ExplainPlan,
		Properties: &ExplainProperties{
			Query: queryPlan,
		},
	}, nil
}

// buildAnalyzePlan 构建ANALYZE TABLE语句的查询计划
func (o *Optimizer) buildAnalyzePlan(stmt *parser.AnalyzeStmt) (*Plan, error) {
	return &Plan{
		Type: AnalyzePlan,
		Properties: &AnalyzeProperties{
			Table:   stmt.Table,
			Columns: stmt.Columns,
		},
	}, nil
}

// convertExpression 将AST表达式节点转换为优化器的表达式结构
func convertExpression(expr parser.Node) Expression {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *parser.BinaryExpr:
		// 处理所有BinaryExpr类型（包括LIKE表达式）
		return &BinaryExpression{
			Left:     convertExpression(e.Left),
			Operator: e.Operator,
			Right:    convertExpression(e.Right),
		}
	case *parser.FunctionCall:
		return &FunctionCall{
			Name: e.Name,
			Args: convertFunctionArgs(e.Args),
		}
	case *parser.ColumnRef:
		return &ColumnReference{
			Column: e.Column,
			Table:  e.Table,
		}
	case *parser.IntegerLiteral:
		return &LiteralValue{
			Type:  LiteralTypeInteger,
			Value: e.Value,
		}
	case *parser.FloatLiteral:
		return &LiteralValue{
			Type:  LiteralTypeFloat,
			Value: e.Value,
		}
	case *parser.StringLiteral:
		return &LiteralValue{
			Type:  LiteralTypeString,
			Value: e.Value,
		}
	case *parser.BooleanLiteral:
		return &LiteralValue{
			Type:  LiteralTypeBoolean,
			Value: e.Value,
		}
	case *parser.ComparisonExpr:
		return &BinaryExpression{
			Left:     convertExpression(e.Left),
			Operator: e.Operator,
			Right:    convertExpression(e.Right),
		}
	case *parser.LogicalExpr:
		return &BinaryExpression{
			Left:     convertExpression(e.Left),
			Operator: e.Operator,
			Right:    convertExpression(e.Right),
		}
	case *parser.InExpr:
		// 将IN表达式转换为多个OR条件: age IN (25, 30, 35) -> age = 25 OR age = 30 OR age = 35
		return convertInExpression(e)
	}
	return nil
}

// convertInExpression 转换IN表达式 (优化版本：使用HashSet而非OR链)
func convertInExpression(inExpr *parser.InExpr) Expression {
	if len(inExpr.Values) == 0 {
		return nil
	}

	leftExpr := convertExpression(inExpr.Left)
	if leftExpr == nil {
		return nil
	}

	// 转换所有值表达式
	values := make([]Expression, 0, len(inExpr.Values))
	for _, value := range inExpr.Values {
		valueExpr := convertExpression(value)
		if valueExpr != nil {
			values = append(values, valueExpr)
		}
	}

	if len(values) == 0 {
		return nil
	}

	// 创建优化的InExpression
	inExp := &InExpression{
		Column:  leftExpr,
		Values:  values,
		IsNotIn: inExpr.Operator == "NOT IN",
	}

	// 预构建HashSet以加速执行时的查找
	inExp.BuildValueSet()

	return inExp
}

// convertFunctionArgs 转换函数参数
func convertFunctionArgs(args []parser.Node) []Expression {
	if args == nil {
		return nil
	}
	result := make([]Expression, len(args))
	for i, arg := range args {
		if arg.Type() == parser.AsteriskNode {
			// P0.2修复: 对于聚合函数参数中的星号(如COUNT(*))
			// 星号是特殊标记,不需要展开为具体列,保留Asterisk即可
			// 聚合函数会在执行时特殊处理COUNT(*)
			result[i] = &Asterisk{TableName: ""}
		} else {
			result[i] = convertExpression(arg)
		}
	}
	return result
}

// buildInsertPlan 构建INSERT查询计划
func (o *Optimizer) buildInsertPlan(stmt *parser.InsertStmt) (*Plan, error) {
	plan := NewPlan(InsertPlan)

	props := &InsertProperties{
		Table:   stmt.Table,
		Columns: stmt.Columns,
	}

	// 处理多行INSERT
	if len(stmt.Rows) > 0 {
		// 多行INSERT - 转换所有行
		props.Rows = make([][]Expression, 0, len(stmt.Rows))
		for _, row := range stmt.Rows {
			props.Rows = append(props.Rows, convertInsertValues(row))
		}
	} else {
		// 单行INSERT - 向后兼容
		props.Values = convertInsertValues(stmt.Values)
	}

	plan.Properties = props
	return plan, nil
}

// convertInsertValues 转换INSERT的值
func convertInsertValues(values []parser.Node) []Expression {
	result := make([]Expression, 0, len(values))
	for _, value := range values {
		result = append(result, convertExpression(value))
	}
	return result
}

// buildUpdatePlan 构建UPDATE查询计划
func (o *Optimizer) buildUpdatePlan(stmt *parser.UpdateStmt) (*Plan, error) {
	plan := NewPlan(UpdatePlan)

	assignments := make(map[string]interface{})
	for _, assign := range stmt.Assignments {
		assignments[assign.Column] = assign.Value
	}

	// Handle WHERE clause (might be nil)
	var whereCondition parser.Node
	if stmt.Where != nil {
		whereCondition = stmt.Where.Condition
	}

	plan.Properties = &UpdateProperties{
		Table:       stmt.Table,
		Assignments: assignments,
		Where:       whereCondition,
	}
	return plan, nil
}

// buildDeletePlan 构建DELETE查询计划
func (o *Optimizer) buildDeletePlan(stmt *parser.DeleteStmt) (*Plan, error) {
	plan := NewPlan(DeletePlan)
	var whereCondition parser.Node
	if stmt.Where != nil {
		whereCondition = stmt.Where.Condition
	}
	plan.Properties = &DeleteProperties{
		Table: stmt.Table,
		Where: whereCondition,
	}
	return plan, nil
}

// convertSelectItems 将解析器的列项转换为优化器的列引用
func convertSelectItems(items []*parser.ColumnItem) []ColumnRef {
	refs := make([]ColumnRef, len(items))
	for i, item := range items {
		refs[i] = ColumnRef{
			Column: item.Column,
			Table:  item.Table,
			Alias:  item.Alias,
		}

		// 根据列项类型设置相应的属性
		switch item.Kind {
		case parser.ColumnItemColumn:
			refs[i].Type = ColumnRefTypeColumn
		case parser.ColumnItemFunction:
			refs[i].Type = ColumnRefTypeFunction
			if funcCall, ok := item.Expr.(*parser.FunctionCall); ok {
				refs[i].FunctionName = funcCall.Name
				refs[i].FunctionArgs = convertFunctionArgs(funcCall.Args)
			}
		case parser.ColumnItemExpression:
			refs[i].Type = ColumnRefTypeExpression
			refs[i].Expression = convertExpression(item.Expr)
		}
	}
	return refs
}

// hasAggregateFunction 检查SELECT列中是否包含聚合函数
func (o *Optimizer) hasAggregateFunction(columns []*parser.ColumnItem) bool {
	for _, col := range columns {
		if col.Expr != nil {
			if funcCall, ok := col.Expr.(*parser.FunctionCall); ok {
				if isAggregateFunction(funcCall.Name) {
					return true
				}
			}
		}
	}
	return false
}

// isAggregateFunction 检查函数是否为聚合函数
func isAggregateFunction(funcName string) bool {
	aggregateFunctions := map[string]bool{
		"COUNT": true,
		"SUM":   true,
		"AVG":   true,
		"MIN":   true,
		"MAX":   true,
	}
	return aggregateFunctions[strings.ToUpper(funcName)]
}

// buildCreateIndexPlan 构建 CREATE INDEX 计划
func (o *Optimizer) buildCreateIndexPlan(stmt *parser.CreateIndexStmt) (*Plan, error) {
	logger.WithComponent("optimizer").Debug("Building CREATE INDEX plan",
		zap.String("index_name", stmt.Name),
		zap.String("table", stmt.Table))

	return &Plan{
		Type: CreateIndexPlan,
		Properties: &CreateIndexProperties{
			Name:     stmt.Name,
			Table:    stmt.Table,
			Columns:  stmt.Columns,
			IsUnique: stmt.IsUnique,
		},
		Children: nil,
	}, nil
}

// buildDropIndexPlan 构建 DROP INDEX 计划
func (o *Optimizer) buildDropIndexPlan(stmt *parser.DropIndexStmt) (*Plan, error) {
	logger.WithComponent("optimizer").Debug("Building DROP INDEX plan",
		zap.String("index_name", stmt.Name),
		zap.String("table", stmt.Table))

	return &Plan{
		Type: DropIndexPlan,
		Properties: &DropIndexProperties{
			Name:  stmt.Name,
			Table: stmt.Table,
		},
		Children: nil,
	}, nil
}

// buildShowIndexesPlan 构建 SHOW INDEXES 计划
func (o *Optimizer) buildShowIndexesPlan(stmt *parser.ShowIndexesStmt) (*Plan, error) {
	logger.WithComponent("optimizer").Debug("Building SHOW INDEXES plan",
		zap.String("table", stmt.Table))

	return &Plan{
		Type: ShowPlan,
		Properties: &ShowIndexesProperties{
			Table: stmt.Table,
		},
		Children: nil,
	}, nil
}

// ParseTableName 解析表名，提取数据库名和表名
// 支持格式：
//   - "table_name" -> ("", "table_name")
//   - "database.table_name" -> ("database", "table_name")
func (o *Optimizer) ParseTableName(fullTableName string) (database string, table string) {
	// 查找第一个点号的位置
	dotIndex := strings.Index(fullTableName, ".")
	if dotIndex == -1 {
		// 没有数据库前缀
		return "", fullTableName
	}

	// 有数据库前缀
	database = fullTableName[:dotIndex]
	table = fullTableName[dotIndex+1:]
	return database, table
}

// expandSelectAll 展开SELECT *为具体列列表
// 这是P0.2的核心实现:在优化阶段展开星号,使投影裁剪优化生效
func (o *Optimizer) expandSelectAll(tableName string) []ColumnRef {
	// 如果没有catalog,无法展开,返回星号占位符
	if o.catalog == nil {
		logger.WithComponent("optimizer").Warn("Cannot expand SELECT * without catalog",
			zap.String("table", tableName))
		return []ColumnRef{{
			Column: "*",
			Table:  tableName,
			Type:   ColumnRefTypeColumn,
		}}
	}

	// 解析表名，可能包含数据库前缀（如 sys.db_metadata）
	database, table := o.ParseTableName(tableName)

	// 如果没有指定数据库，使用当前数据库
	if database == "" {
		database = o.currentDatabase
		if database == "" {
			database = "default"
		}
	}

	// 从catalog获取表schema
	tableMeta, err := o.catalog.GetTable(database, table)
	if err != nil {
		logger.WithComponent("optimizer").Error("Failed to get table schema for SELECT * expansion",
			zap.String("database", database),
			zap.String("table", table),
			zap.Error(err))

		// 回退到星号占位符
		return []ColumnRef{{
			Column: "*",
			Table:  table,
			Type:   ColumnRefTypeColumn,
		}}
	}

	if tableMeta.Schema == nil {
		logger.WithComponent("optimizer").Warn("Table schema is nil, cannot expand SELECT *",
			zap.String("database", database),
			zap.String("table", table))
		return []ColumnRef{{
			Column: "*",
			Table:  table,
			Type:   ColumnRefTypeColumn,
		}}
	}

	// 展开为具体列引用列表
	columns := make([]ColumnRef, 0, tableMeta.Schema.NumFields())
	for i := 0; i < tableMeta.Schema.NumFields(); i++ {
		field := tableMeta.Schema.Field(i)
		columns = append(columns, ColumnRef{
			Table:  tableName,
			Column: field.Name,
			Type:   ColumnRefTypeColumn,
		})
	}

	logger.WithComponent("optimizer").Info("SELECT * expanded successfully",
		zap.String("database", database),
		zap.String("table", tableName),
		zap.Int("column_count", len(columns)))

	return columns
}

// expandAsterisk 展开函数参数中的星号(如COUNT(*))
// 修复optimizer.go:605的TODO
func (o *Optimizer) expandAsterisk(tableName string) []Expression {
	// 对于聚合函数如COUNT(*),星号实际上是一个特殊标记,不需要真正展开
	// 返回Asterisk表达式即可
	return []Expression{&Asterisk{TableName: tableName}}
}

package optimizer

import (
	"fmt"
	"strings"
	"time"

	"github.com/yyun543/minidb/internal/logger"
	"github.com/yyun543/minidb/internal/parser"
	"go.uber.org/zap"
)

// Optimizer 查询优化器
type Optimizer struct {
	rules []Rule // 优化规则
}

// NewOptimizer 创建新的优化器实例
func NewOptimizer() *Optimizer {
	logger.WithComponent("optimizer").Info("Creating new query optimizer")

	opt := &Optimizer{
		rules: make([]Rule, 0),
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
		zap.String("plan_type", string(plan.Type)),
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
		zap.String("final_plan_type", string(plan.Type)),
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

	projectPlan.Properties = &SelectProperties{
		All:     isSelectAll,
		Columns: convertSelectItems(stmt.Columns),
	}

	// 2. 构建表扫描和JOIN
	var currentPlan *Plan
	if stmt.From != "" {
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

	// 4. 构建GROUP BY
	if len(stmt.GroupBy) > 0 {
		groupPlan := NewPlan(GroupPlan)
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
				orderKeys[i] = OrderKey{
					Column:    colRef.Column,
					Table:     colRef.Table,
					Direction: item.Direction,
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
	hasAggregateFunction := o.hasAggregateFunction(stmt.Columns)
	if !isSelectAll && len(stmt.GroupBy) == 0 && !hasAggregateFunction {
		projectionPlan := NewPlan(ProjectionPlan)
		projectionPlan.Properties = &ProjectionProperties{
			Columns: convertSelectItems(stmt.Columns),
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

	// 处理每个JOIN子句
	for _, join := range joins {
		joinPlan := NewPlan(JoinPlan)

		// 创建右表扫描
		rightScan := NewPlan(TableScanPlan)
		rightScan.Properties = &TableScanProperties{
			Table:      join.Right.Table,
			TableAlias: join.Right.Alias,
		}

		// 设置JOIN属性
		joinPlan.Properties = &JoinProperties{
			JoinType:   join.JoinType,
			Left:       leftTable,
			LeftAlias:  leftAlias,
			Right:      join.Right.Table,
			RightAlias: join.Right.Alias,
			Condition:  convertExpression(join.Condition),
		}

		// 添加左右子节点
		joinPlan.AddChild(currentPlan)
		joinPlan.AddChild(rightScan)

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
	columns := make([]ColumnRef, len(stmt.Columns))
	for i, col := range stmt.Columns {
		columns[i] = ColumnRef{
			Column: col.Name,
			Type:   ColumnRefTypeColumn,
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

// convertInExpression 转换IN表达式为OR条件链
func convertInExpression(inExpr *parser.InExpr) Expression {
	if len(inExpr.Values) == 0 {
		return nil
	}

	leftExpr := convertExpression(inExpr.Left)
	if leftExpr == nil {
		return nil
	}

	// 为每个值创建等值比较条件
	var orChain Expression

	for _, value := range inExpr.Values {
		valueExpr := convertExpression(value)
		if valueExpr == nil {
			continue
		}

		// 创建等值比较: column = value
		eqExpr := &BinaryExpression{
			Left:     leftExpr,
			Operator: "=",
			Right:    valueExpr,
		}

		// 如果是NOT IN，使用不等于比较
		if inExpr.Operator == "NOT IN" {
			eqExpr.Operator = "!="
		}

		if orChain == nil {
			orChain = eqExpr
		} else {
			// 构建OR链: (... OR column = value)
			logicalOp := "OR"
			if inExpr.Operator == "NOT IN" {
				// NOT IN 使用 AND: column != val1 AND column != val2
				logicalOp = "AND"
			}

			orChain = &BinaryExpression{
				Left:     orChain,
				Operator: logicalOp,
				Right:    eqExpr,
			}
		}
	}

	return orChain
}

// convertFunctionArgs 转换函数参数
func convertFunctionArgs(args []parser.Node) []Expression {
	if args == nil {
		return nil
	}
	result := make([]Expression, len(args))
	for i, arg := range args {
		if arg.Type() == parser.AsteriskNode {
			// TODO 根据元数据补全所有字段信息
			result[i] = &Asterisk{}
		} else {
			result[i] = convertExpression(arg)
		}
	}
	return result
}

// buildInsertPlan 构建INSERT查询计划
func (o *Optimizer) buildInsertPlan(stmt *parser.InsertStmt) (*Plan, error) {
	plan := NewPlan(InsertPlan)
	plan.Properties = &InsertProperties{
		Table:   stmt.Table,
		Columns: stmt.Columns,
		Values:  convertInsertValues(stmt.Values),
	}
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
	plan.Properties = &DeleteProperties{
		Table: stmt.Table,
		Where: stmt.Where.Condition,
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

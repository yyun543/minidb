package optimizer

import (
	"fmt"

	"github.com/yyun543/minidb/internal/parser"
)

// Optimizer 查询优化器
type Optimizer struct {
	rules []Rule // 优化规则
}

// NewOptimizer 创建新的优化器实例
func NewOptimizer() *Optimizer {
	opt := &Optimizer{
		rules: make([]Rule, 0),
	}
	// 添加优化规则
	opt.rules = append(opt.rules,
		&PredicatePushDownRule{}, // 谓词下推
		&JoinReorderRule{},       // Join重排序
		&ProjectionPruningRule{}, // 投影剪枝
	)
	return opt
}

// Optimize 优化查询
func (o *Optimizer) Optimize(stmt parser.Node) (*Plan, error) {
	// 1. 构建初始计划
	plan, err := o.buildPlan(stmt)
	if err != nil {
		return nil, err
	}

	// 2. 应用优化规则
	for _, rule := range o.rules {
		plan = rule.Apply(plan)
	}

	return plan, nil
}

// buildPlan 根据AST构建初始查询计划
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
	default:
		return nil, fmt.Errorf("unsupported statement type: %T", node)
	}
}

// buildSelectPlan 构建SELECT查询计划
func (o *Optimizer) buildSelectPlan(stmt *parser.SelectStmt) (*Plan, error) {
	plan := NewPlan(SelectPlan)

	// 1. 构建SELECT属性
	selectProps := &SelectProperties{
		Columns: make([]ColumnRef, 0),
	}

	// 处理 SELECT *
	if stmt.All {
		selectProps.All = true
	} else {
		// 处理普通列、函数和表达式
		for _, col := range stmt.Columns {
			colRef := ColumnRef{
				Column: col.Column,
				Table:  col.Table,
				Alias:  col.Alias,
			}

			// 根据不同类型处理
			switch col.Kind {
			case parser.ColumnItemColumn:
				colRef.Type = ColumnRefTypeColumn
			case parser.ColumnItemFunction:
				colRef.Type = ColumnRefTypeFunction
				if funcExpr, ok := col.Expr.(*parser.FunctionCall); ok {
					colRef.FunctionName = funcExpr.Name
					// 处理函数参数
					colRef.FunctionArgs = convertFunctionArgs(funcExpr.Args)
				}
			case parser.ColumnItemExpression:
				colRef.Type = ColumnRefTypeExpression
				colRef.Expression = convertExpression(col.Expr)
			}

			selectProps.Columns = append(selectProps.Columns, colRef)
		}
	}
	plan.Properties = selectProps

	// 2. 构建FROM子句的表扫描
	if stmt.From != "" {
		scanPlan := NewPlan(TableScanPlan)
		scanProps := &TableScanProperties{
			Table:      stmt.From,
			TableAlias: stmt.FromAlias,
			Columns:    make([]ColumnRef, 0),
		}

		// 收集所需的列信息
		for _, col := range selectProps.Columns {
			if col.Type == ColumnRefTypeColumn {
				scanProps.Columns = append(scanProps.Columns, col)
			}
		}

		scanPlan.Properties = scanProps
		plan.AddChild(scanPlan)
	}

	// 3. 构建JOIN
	if len(stmt.Joins) > 0 {
		for _, join := range stmt.Joins {
			joinPlan := NewPlan(JoinPlan)
			joinProps := &JoinProperties{
				JoinType:   join.JoinType,
				Left:       join.Left.Table,
				LeftAlias:  join.Left.Alias,
				Right:      join.Right.Table,
				RightAlias: join.Right.Alias,
				Condition:  convertExpression(join.Condition),
			}
			joinPlan.Properties = joinProps
			plan.AddChild(joinPlan)
		}
	}

	// 4. 构建WHERE过滤
	if stmt.Where != nil {
		filterPlan := NewPlan(FilterPlan)
		filterPlan.Properties = &FilterProperties{
			Condition: convertExpression(stmt.Where.Condition),
		}
		plan.AddChild(filterPlan)
	}

	// 5. 构建GROUP BY
	if len(stmt.GroupBy) > 0 {
		groupPlan := NewPlan(GroupPlan)
		groupKeys := make([]string, len(stmt.GroupBy))
		for i, key := range stmt.GroupBy {
			if colRef, ok := key.(*parser.ColumnRef); ok {
				groupKeys[i] = colRef.Column
			}
		}
		groupPlan.Properties = &GroupByProperties{
			GroupKeys: groupKeys,
			Having:    stmt.Having,
		}
		plan.AddChild(groupPlan)
	}

	// 6. 构建ORDER BY
	if len(stmt.OrderBy) > 0 {
		orderPlan := NewPlan(OrderPlan)
		orderKeys := make([]OrderKey, len(stmt.OrderBy))
		for i, item := range stmt.OrderBy {
			if colRef, ok := item.Expr.(*parser.ColumnRef); ok {
				orderKeys[i] = OrderKey{
					Column:    colRef.Column,
					Direction: item.Direction,
				}
			}
		}
		orderPlan.Properties = &OrderByProperties{
			OrderKeys: orderKeys,
		}
		plan.AddChild(orderPlan)
	}

	// 7. 构建LIMIT
	if stmt.Limit > 0 {
		limitPlan := NewPlan(LimitPlan)
		limitPlan.Properties = &LimitProperties{
			Limit: stmt.Limit,
		}
		plan.AddChild(limitPlan)
	}

	return plan, nil
}

// convertExpression 将AST表达式节点转换为优化器的表达式结构
func convertExpression(expr parser.Node) Expression {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *parser.BinaryExpr:
		return &BinaryExpression{
			Left:     convertExpression(e.Left),
			Operator: e.Operator,
			Right:    convertExpression(e.Right),
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
	}
	return nil
}

// convertFunctionArgs 转换函数参数
func convertFunctionArgs(args []parser.Node) []Expression {
	result := make([]Expression, 0, len(args))
	for _, arg := range args {
		if expr := convertExpression(arg); expr != nil {
			result = append(result, expr)
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

	plan.Properties = &UpdateProperties{
		Table:       stmt.Table,
		Assignments: assignments,
		Where:       stmt.Where.Condition,
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

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

// buildSelectPlan 根据AST构建初始查询计划
func (o *Optimizer) buildSelectPlan(stmt *parser.SelectStmt) (*Plan, error) {
	// 1. 创建投影算子
	projectPlan := NewPlan(SelectPlan)
	projectPlan.Properties = &SelectProperties{
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
		groupPlan.Properties = &GroupByProperties{
			GroupKeys: groupKeys,
		}
		groupPlan.AddChild(currentPlan)
		currentPlan = groupPlan
	}

	// 5. 构建HAVING (必须在GROUP BY之后)
	if stmt.Having != nil {
		havingPlan := NewPlan(HavingPlan)
		havingPlan.Properties = &HavingProperties{
			Condition: convertExpression(stmt.Having),
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

	// 8. 最后添加投影算子
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
	}
	return nil
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

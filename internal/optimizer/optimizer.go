package optimizer

import (
	"github.com/yyun543/minidb/internal/parser"
)

// JoinReorder JOIN重排序优化规则
type JoinReorder struct{}

func (r *JoinReorder) Apply(plan *LogicalPlan) *LogicalPlan {
	// Implementation here
	return plan
}

// Optimizer 查询优化器
type Optimizer struct {
	rules []Rule // 优化规则列表
}

// NewOptimizer 创建新的优化器实例
func NewOptimizer() *Optimizer {
	opt := &Optimizer{}
	// 注册优化规则
	opt.rules = []Rule{
		&JoinReorder{}, // JOIN重排序规则
		// 可以添加更多规则
	}
	return opt
}

// Optimize 优化查询
func (o *Optimizer) Optimize(node parser.Node) (*LogicalPlan, error) {
	// 1. 将AST转换为逻辑计划
	plan := o.convertToLogicalPlan(node)

	// 2. 应用优化规则
	for _, rule := range o.rules {
		plan = rule.Apply(plan)
	}

	return plan, nil
}

// convertToLogicalPlan 将AST转换为逻辑计划
func (o *Optimizer) convertToLogicalPlan(node parser.Node) *LogicalPlan {
	switch n := node.(type) {
	case *parser.SelectStmt:
		return o.convertSelect(n)
	case *parser.InsertStmt:
		return o.convertInsert(n)
	case *parser.UpdateStmt:
		return o.convertUpdate(n)
	case *parser.DeleteStmt:
		return o.convertDelete(n)
	default:
		return nil
	}
}

// convertSelect 转换SELECT语句
func (o *Optimizer) convertSelect(stmt *parser.SelectStmt) *LogicalPlan {
	plan := &LogicalPlan{
		Type: SelectPlan,
		Properties: &SelectProperties{
			Columns: stmt.Columns,
		},
	}

	// 处理FROM子句
	if stmt.From != "" {
		plan.Children = append(plan.Children, &LogicalPlan{
			Type: TableScanPlan,
			Properties: &TableScanProperties{
				Table: stmt.From,
			},
		})
	}

	// 处理JOIN
	for _, join := range stmt.Joins {
		plan.Children = append(plan.Children, &LogicalPlan{
			Type: JoinPlan,
			Properties: &JoinProperties{
				JoinType:  join.JoinType,
				Table:     join.Right.Table,
				Condition: join.Condition,
			},
		})
	}

	// 处理WHERE子句
	if stmt.Where != nil {
		plan.Children = append(plan.Children, &LogicalPlan{
			Type: FilterPlan,
			Properties: &FilterProperties{
				Condition: stmt.Where.Condition,
			},
		})
	}

	return plan
}

// convertInsert 转换INSERT语句
func (o *Optimizer) convertInsert(stmt *parser.InsertStmt) *LogicalPlan {
	return &LogicalPlan{
		Type: InsertPlan,
		Properties: &InsertProperties{
			Table:   stmt.Table,
			Columns: stmt.Columns,
			Values:  stmt.Values,
		},
	}
}

// convertUpdate 转换UPDATE语句
func (o *Optimizer) convertUpdate(stmt *parser.UpdateStmt) *LogicalPlan {
	return &LogicalPlan{
		Type: UpdatePlan,
		Properties: &UpdateProperties{
			Table:       stmt.Table,
			Assignments: stmt.Assignments,
			Where:       stmt.Where,
		},
	}
}

// convertDelete 转换DELETE语句
func (o *Optimizer) convertDelete(stmt *parser.DeleteStmt) *LogicalPlan {
	return &LogicalPlan{
		Type: DeletePlan,
		Properties: &DeleteProperties{
			Table: stmt.Table,
			Where: stmt.Where,
		},
	}
}

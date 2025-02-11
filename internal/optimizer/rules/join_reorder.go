package rules

import (
	"github.com/yyun543/minidb/internal/optimizer"
)

// JoinReorder JOIN重排序优化规则
type JoinReorder struct{}

// Apply 应用JOIN重排序规则
func (r *JoinReorder) Apply(plan *optimizer.LogicalPlan) *optimizer.LogicalPlan {
	// 如果不是SELECT计划，直接返回
	if plan.Type != optimizer.SelectPlan {
		return plan
	}

	// 收集所有JOIN节点
	var joins []*optimizer.LogicalPlan
	for _, child := range plan.Children {
		if child.Type == optimizer.JoinPlan {
			joins = append(joins, child)
		}
	}

	// 如果没有JOIN或只有一个JOIN，无需重排序
	if len(joins) <= 1 {
		return plan
	}

	// 简单的贪心策略：根据代价估算重排序
	// 这里使用一个简单的启发式方法：小表在前
	reorderedJoins := r.reorderBySize(joins)

	// 更新计划的子节点
	newChildren := make([]*optimizer.LogicalPlan, 0)
	for _, child := range plan.Children {
		if child.Type != optimizer.JoinPlan {
			newChildren = append(newChildren, child)
		}
	}
	newChildren = append(newChildren, reorderedJoins...)
	plan.Children = newChildren

	return plan
}

// reorderBySize 根据表大小重排序JOIN
func (r *JoinReorder) reorderBySize(joins []*optimizer.LogicalPlan) []*optimizer.LogicalPlan {
	// TODO: 实现更复杂的代价估算
	// 当前简单实现：保持原有顺序
	return joins
}

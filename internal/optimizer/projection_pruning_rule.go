package optimizer

// ProjectionPruningRule 投影剪枝规则
type ProjectionPruningRule struct{}

func (r *ProjectionPruningRule) Apply(plan *Plan) *Plan {
	// 实现投影剪枝逻辑
	// 1. 分析需要的列
	// 2. 剪除不需要的列
	return plan
}

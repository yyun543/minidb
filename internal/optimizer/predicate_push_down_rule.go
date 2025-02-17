package optimizer

// PredicatePushDownRule 谓词下推规则
type PredicatePushDownRule struct{}

func (r *PredicatePushDownRule) Apply(plan *Plan) *Plan {
	// 实现谓词下推逻辑
	// 1. 收集所有过滤条件
	// 2. 尽可能将过滤条件下推到表扫描节点
	// 3. 重构计划树
	return plan
}

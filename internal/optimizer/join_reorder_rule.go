package optimizer

// JoinReorderRule Join重排序规则
type JoinReorderRule struct{}

func (r *JoinReorderRule) Apply(plan *Plan) *Plan {
	// 实现Join重排序逻辑
	// 1. 分析Join图
	// 2. 估算代价
	// 3. 选择最优Join顺序
	return plan
}

package optimizer

// Rule 优化规则接口
type Rule interface {
	Apply(plan *Plan) *Plan
}

package operators

import (
	"github.com/yyun543/minidb/internal/types"
)

// Aggregate 聚合算子
type Aggregate struct {
	functions []string
	groupBy   []string
	child     Operator // Use the new package
	ctx       interface{}
}

// NewAggregate 创建聚合算子
func NewAggregate(functions, groupBy []string, child Operator, ctx interface{}) *Aggregate {
	return &Aggregate{
		functions: functions,
		groupBy:   groupBy,
		child:     child,
		ctx:       ctx,
	}
}

// Init 初始化算子
func (op *Aggregate) Init(ctx interface{}) error {
	return op.child.Init(ctx)
}

// Next 获取下一批数据
func (op *Aggregate) Next() (*types.Batch, error) {
	// 实现聚合计算
	// TODO: 实现聚合逻辑
	return nil, nil
}

// Close 关闭算子
func (op *Aggregate) Close() error {
	return op.child.Close()
}

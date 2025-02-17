package operators

import (
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/types"
)

// Filter 过滤算子
type Filter struct {
	condition optimizer.Expression // 过滤条件
	child     Operator             // Use local Operator interface
	ctx       interface{}          // Use interface{} instead of *executor.Context
}

// NewFilter 创建过滤算子
func NewFilter(condition optimizer.Expression, child Operator, ctx interface{}) *Filter {
	return &Filter{
		condition: condition,
		child:     child,
		ctx:       ctx,
	}
}

// Init 初始化算子
func (op *Filter) Init(ctx interface{}) error {
	return op.child.Init(ctx)
}

// Next 获取下一批数据
func (op *Filter) Next() (*types.Batch, error) {
	// 获取子算子数据
	batch, err := op.child.Next()
	if err != nil {
		return nil, err
	}
	if batch == nil {
		return nil, nil
	}

	// 应用过滤条件
	// TODO: 实现过滤逻辑
	return batch, nil
}

// Close 关闭算子
func (op *Filter) Close() error {
	return op.child.Close()
}

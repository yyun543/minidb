package operators

import (
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/types"
)

// Join 连接算子
type Join struct {
	joinType  string               // 连接类型
	condition optimizer.Expression // 连接条件
	left      Operator             // 左子算子
	right     Operator             // 右子算子
	ctx       interface{}
}

// NewJoin 创建连接算子
func NewJoin(joinType string, condition optimizer.Expression, left, right Operator, ctx interface{}) *Join {
	return &Join{
		joinType:  joinType,
		condition: condition,
		left:      left,
		right:     right,
		ctx:       ctx,
	}
}

// Init 初始化算子
func (op *Join) Init(ctx interface{}) error {
	if err := op.left.Init(ctx); err != nil {
		return err
	}
	return op.right.Init(ctx)
}

// Next 获取下一批数据
func (op *Join) Next() (*types.Batch, error) {
	// 实现嵌套循环连接
	// TODO: 实现连接逻辑
	return nil, nil
}

// Close 关闭算子
func (op *Join) Close() error {
	if err := op.left.Close(); err != nil {
		return err
	}
	return op.right.Close()
}

package operators

import "github.com/yyun543/minidb/internal/types"

// Operator 算子接口
type Operator interface {
	// Init 初始化算子
	Init(ctx interface{}) error
	// Next 获取下一批数据
	Next() (*types.Batch, error)
	// Close 关闭算子
	Close() error
}

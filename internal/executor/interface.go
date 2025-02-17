package executor

import (
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/types"
)

// Executor 执行器接口
type Executor interface {
	// Execute 执行查询计划并返回结果集
	Execute(plan *optimizer.Plan) (*ResultSet, error)
}

// Operator 算子接口
type Operator interface {
	// Init 初始化算子
	Init(ctx *Context) error

	// Next 获取下一批数据
	Next() (*types.Batch, error)

	// Close 关闭算子
	Close() error
}

// ResultSet 查询结果集
type ResultSet struct {
	headers []string       // 列名
	rows    []*types.Batch // 数据批次
	curRow  int            // 当前行
}

// Batches 返回结果集中的数据批次
func (rs *ResultSet) Batches() []*types.Batch {
	return rs.rows
}

// Headers 获取结果集列名
func (rs *ResultSet) Headers() []string {
	return rs.headers
}

// Next 获取下一行数据
func (rs *ResultSet) Next() bool {
	rs.curRow++
	return rs.curRow < len(rs.rows)
}

// Row 获取当前行数据
func (rs *ResultSet) Row() []interface{} {
	if rs.curRow >= len(rs.rows) {
		return nil
	}
	return rs.rows[rs.curRow].Values()
}

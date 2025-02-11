package executor

import (
	"fmt"

	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor/operators"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/types"
)

// ExecutorImpl 执行器实现
type ExecutorImpl struct {
	catalog *catalog.Catalog
}

// NewExecutor 创建执行器实例
func NewExecutor(cat *catalog.Catalog) *ExecutorImpl {
	return &ExecutorImpl{
		catalog: cat,
	}
}

// Execute 执行查询计划
func (e *ExecutorImpl) Execute(plan *optimizer.LogicalPlan) (*ResultSet, error) {
	// 创建执行上下文
	ctx := NewContext(e.catalog)

	// 构建执行算子树
	op, err := e.buildOperator(plan, ctx)
	if err != nil {
		return nil, err
	}

	// 初始化算子
	if err := op.Init(ctx); err != nil {
		return nil, err
	}

	// 执行查询并收集结果
	var batches []*types.Batch
	for {
		batch, err := op.Next()
		if err != nil {
			return nil, err
		}
		if batch == nil {
			break
		}
		batches = append(batches, batch)
	}

	// 关闭算子
	if err := op.Close(); err != nil {
		return nil, err
	}

	// 构建结果集
	headers := e.getResultHeaders(plan)
	return &ResultSet{
		headers: headers,
		rows:    batches,
		curRow:  -1,
	}, nil
}

// buildOperator 根据计划节点构建算子
func (e *ExecutorImpl) buildOperator(plan *optimizer.LogicalPlan, ctx *Context) (operators.Operator, error) {
	switch plan.Type {
	case optimizer.TableScanPlan:
		props := plan.Properties.(*optimizer.TableScanProperties)
		return operators.NewTableScan(props.Table, e.catalog), nil

	case optimizer.JoinPlan:
		props := plan.Properties.(*optimizer.JoinProperties)
		left, err := e.buildOperator(plan.Children[0], ctx)
		if err != nil {
			return nil, err
		}
		right, err := e.buildOperator(plan.Children[1], ctx)
		if err != nil {
			return nil, err
		}
		return operators.NewJoin(props.JoinType, props.Condition, left, right, ctx), nil

	case optimizer.FilterPlan:
		props := plan.Properties.(*optimizer.FilterProperties)
		child, err := e.buildOperator(plan.Children[0], ctx)
		if err != nil {
			return nil, err
		}
		return operators.NewFilter(props.Condition, child, ctx), nil

	default:
		return nil, fmt.Errorf("不支持的计划节点类型: %v", plan.Type)
	}
}

// getResultHeaders 获取结果集列名
func (e *ExecutorImpl) getResultHeaders(plan *optimizer.LogicalPlan) []string {
	switch plan.Type {
	case optimizer.SelectPlan:
		props := plan.Properties.(*optimizer.SelectProperties)
		return props.Columns
	default:
		return nil
	}
}

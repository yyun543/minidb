package executor

import (
	"math"

	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/statistics"
)

// CostBasedOptimizer 基于成本的查询优化器
type CostBasedOptimizer struct {
	statsMgr *statistics.StatisticsManager
	config   *OptimizerConfig
}

// OptimizerConfig 优化器配置
type OptimizerConfig struct {
	// 成本参数
	SeqScanCostFactor    float64 // 顺序扫描成本因子
	IndexScanCostFactor  float64 // 索引扫描成本因子
	HashJoinCostFactor   float64 // 哈希连接成本因子
	NestedLoopCostFactor float64 // 嵌套循环连接成本因子
	SortCostFactor       float64 // 排序成本因子

	// 内存参数
	WorkMemSize    int64 // 工作内存大小
	BufferPoolSize int64 // 缓冲池大小

	// 优化开关
	EnableIndexScan     bool // 启用索引扫描
	EnableHashJoin      bool // 启用哈希连接
	EnableSortMergeJoin bool // 启用排序合并连接
	EnableParallelScan  bool // 启用并行扫描
}

// DefaultOptimizerConfig 默认优化器配置
func DefaultOptimizerConfig() *OptimizerConfig {
	return &OptimizerConfig{
		SeqScanCostFactor:    1.0,
		IndexScanCostFactor:  0.1,
		HashJoinCostFactor:   1.5,
		NestedLoopCostFactor: 2.0,
		SortCostFactor:       1.2,
		WorkMemSize:          64 * 1024 * 1024,  // 64MB
		BufferPoolSize:       256 * 1024 * 1024, // 256MB
		EnableIndexScan:      true,
		EnableHashJoin:       true,
		EnableSortMergeJoin:  true,
		EnableParallelScan:   true,
	}
}

// NewCostBasedOptimizer 创建基于成本的优化器
func NewCostBasedOptimizer(statsMgr *statistics.StatisticsManager) *CostBasedOptimizer {
	return &CostBasedOptimizer{
		statsMgr: statsMgr,
		config:   DefaultOptimizerConfig(),
	}
}

// OptimizePlan 优化查询计划
func (cbo *CostBasedOptimizer) OptimizePlan(plan *optimizer.Plan) (*optimizer.Plan, error) {
	// 创建优化上下文
	ctx := &OptimizationContext{
		statsMgr: cbo.statsMgr,
		config:   cbo.config,
	}

	// 应用优化规则
	optimizedPlan, err := cbo.applyOptimizationRules(plan, ctx)
	if err != nil {
		return plan, err // 如果优化失败，返回原计划
	}

	return optimizedPlan, nil
}

// OptimizationContext 优化上下文
type OptimizationContext struct {
	statsMgr       *statistics.StatisticsManager
	config         *OptimizerConfig
	tableCosts     map[string]*TableCostInfo
	joinOrderCache map[string]*optimizer.Plan
}

// TableCostInfo 表成本信息
type TableCostInfo struct {
	TableName     string
	RowCount      int64
	DataSize      int64
	SeqScanCost   float64
	IndexScanCost map[string]float64 // 索引名 -> 扫描成本
}

// applyOptimizationRules 应用优化规则
func (cbo *CostBasedOptimizer) applyOptimizationRules(plan *optimizer.Plan, ctx *OptimizationContext) (*optimizer.Plan, error) {
	// 递归优化子节点
	optimizedChildren := make([]*optimizer.Plan, len(plan.Children))
	for i, child := range plan.Children {
		optimizedChild, err := cbo.applyOptimizationRules(child, ctx)
		if err != nil {
			return nil, err
		}
		optimizedChildren[i] = optimizedChild
	}

	// 创建优化后的计划副本
	optimizedPlan := &optimizer.Plan{
		Type:       plan.Type,
		Properties: plan.Properties,
		Children:   optimizedChildren,
	}

	// 根据节点类型应用特定优化
	switch plan.Type {
	case optimizer.SelectPlan:
		return cbo.optimizeSelect(optimizedPlan, ctx)
	case optimizer.JoinPlan:
		return cbo.optimizeJoin(optimizedPlan, ctx)
	case optimizer.FilterPlan:
		return cbo.optimizeFilter(optimizedPlan, ctx)
	case optimizer.TableScanPlan:
		return cbo.optimizeTableScan(optimizedPlan, ctx)
	default:
		return optimizedPlan, nil
	}
}

// optimizeSelect 优化SELECT操作
func (cbo *CostBasedOptimizer) optimizeSelect(plan *optimizer.Plan, ctx *OptimizationContext) (*optimizer.Plan, error) {
	// 投影下推：尽早减少列数
	// 这里可以应用列裁剪等优化

	return cbo.pushdownProjection(plan, ctx)
}

// optimizeJoin 优化连接操作
func (cbo *CostBasedOptimizer) optimizeJoin(plan *optimizer.Plan, ctx *OptimizationContext) (*optimizer.Plan, error) {
	// 连接重排序：选择最优的连接顺序
	if len(plan.Children) >= 2 {
		return cbo.optimizeJoinOrder(plan, ctx)
	}

	// 选择最优的连接算法
	return cbo.chooseJoinAlgorithm(plan, ctx)
}

// optimizeFilter 优化过滤操作
func (cbo *CostBasedOptimizer) optimizeFilter(plan *optimizer.Plan, ctx *OptimizationContext) (*optimizer.Plan, error) {
	// 谓词下推：将过滤条件推到更接近数据源的位置
	return cbo.pushdownFilter(plan, ctx)
}

// optimizeTableScan 优化表扫描操作
func (cbo *CostBasedOptimizer) optimizeTableScan(plan *optimizer.Plan, ctx *OptimizationContext) (*optimizer.Plan, error) {
	// 选择最优的扫描方式（顺序扫描 vs 索引扫描）
	return cbo.chooseOptimalScan(plan, ctx)
}

// pushdownProjection 投影下推
func (cbo *CostBasedOptimizer) pushdownProjection(plan *optimizer.Plan, ctx *OptimizationContext) (*optimizer.Plan, error) {
	// 简化实现：将SELECT尽可能下推到叶子节点
	props := plan.Properties.(*optimizer.SelectProperties)

	// 如果子节点是表扫描，可以直接在扫描时进行列裁剪
	if len(plan.Children) == 1 && plan.Children[0].Type == optimizer.TableScanPlan {
		childProps := plan.Children[0].Properties.(*optimizer.TableScanProperties)

		// 更新表扫描的列信息
		childProps.Columns = props.Columns

		// 返回优化后的表扫描节点，去掉多余的SELECT节点
		return plan.Children[0], nil
	}

	return plan, nil
}

// optimizeJoinOrder 优化连接顺序
func (cbo *CostBasedOptimizer) optimizeJoinOrder(plan *optimizer.Plan, ctx *OptimizationContext) (*optimizer.Plan, error) {
	// 对于多表连接，使用动态规划算法找出最优连接顺序

	// 简化实现：只处理两表连接
	if len(plan.Children) == 2 {
		leftCost := cbo.estimatePlanCost(plan.Children[0], ctx)
		rightCost := cbo.estimatePlanCost(plan.Children[1], ctx)

		// 估算两种连接顺序的成本
		props := plan.Properties.(*optimizer.JoinProperties)

		// 方案1: left join right
		joinCost1 := cbo.estimateJoinCost(plan.Children[0], plan.Children[1], props, ctx)
		totalCost1 := leftCost + rightCost + joinCost1

		// 方案2: right join left (如果是内连接才能交换)
		if props.JoinType == "INNER" {
			joinCost2 := cbo.estimateJoinCost(plan.Children[1], plan.Children[0], props, ctx)
			totalCost2 := leftCost + rightCost + joinCost2

			// 选择成本更低的方案
			if totalCost2 < totalCost1 {
				// 交换左右子树
				plan.Children[0], plan.Children[1] = plan.Children[1], plan.Children[0]
			}
		}
	}

	return plan, nil
}

// chooseJoinAlgorithm 选择连接算法
func (cbo *CostBasedOptimizer) chooseJoinAlgorithm(plan *optimizer.Plan, ctx *OptimizationContext) (*optimizer.Plan, error) {
	if len(plan.Children) != 2 {
		return plan, nil
	}

	props := plan.Properties.(*optimizer.JoinProperties)

	// 估算不同连接算法的成本
	nestedLoopCost := cbo.estimateNestedLoopJoinCost(plan.Children[0], plan.Children[1], props, ctx)
	hashJoinCost := cbo.estimateHashJoinCost(plan.Children[0], plan.Children[1], props, ctx)

	// 选择成本最低的算法
	if ctx.config.EnableHashJoin && hashJoinCost < nestedLoopCost {
		// 使用哈希连接
		// 这里可以在计划中添加标记，指示使用哈希连接
		// 简化实现中暂不修改计划结构
	}

	return plan, nil
}

// pushdownFilter 谓词下推
func (cbo *CostBasedOptimizer) pushdownFilter(plan *optimizer.Plan, ctx *OptimizationContext) (*optimizer.Plan, error) {
	props := plan.Properties.(*optimizer.FilterProperties)

	// 如果子节点是表扫描，将过滤条件下推到扫描层
	if len(plan.Children) == 1 && plan.Children[0].Type == optimizer.TableScanPlan {
		// 在实际实现中，这里应该将过滤条件添加到表扫描的属性中
		// 简化实现中保持原有结构
		return plan, nil
	}

	// 如果子节点是连接，尝试将过滤条件下推到连接的一侧
	if len(plan.Children) == 1 && plan.Children[0].Type == optimizer.JoinPlan {
		return cbo.pushdownFilterToJoin(plan, props.Condition, ctx)
	}

	return plan, nil
}

// chooseOptimalScan 选择最优扫描方式
func (cbo *CostBasedOptimizer) chooseOptimalScan(plan *optimizer.Plan, ctx *OptimizationContext) (*optimizer.Plan, error) {
	props := plan.Properties.(*optimizer.TableScanProperties)

	// 获取表统计信息
	tableStats, err := ctx.statsMgr.GetTableStatistics(props.Table)
	if err != nil {
		// 没有统计信息时，使用默认扫描
		return plan, nil
	}

	// 估算顺序扫描成本
	seqScanCost := float64(tableStats.RowCount) * ctx.config.SeqScanCostFactor

	// 如果有索引且启用索引扫描，比较成本
	if ctx.config.EnableIndexScan {
		// TODO: 检查是否有可用的索引
		// 这里需要从catalog中获取索引信息

		// 简化实现：假设存在主键索引
		indexScanCost := float64(tableStats.RowCount) * 0.1 * ctx.config.IndexScanCostFactor

		if indexScanCost < seqScanCost {
			// 在实际实现中，这里应该修改扫描类型
			// 简化实现中暂不修改
		}
	}

	return plan, nil
}

// 成本估算方法

// estimatePlanCost 估算计划成本
func (cbo *CostBasedOptimizer) estimatePlanCost(plan *optimizer.Plan, ctx *OptimizationContext) float64 {
	switch plan.Type {
	case optimizer.TableScanPlan:
		return cbo.estimateTableScanCost(plan, ctx)
	case optimizer.FilterPlan:
		return cbo.estimateFilterCost(plan, ctx)
	case optimizer.JoinPlan:
		if len(plan.Children) == 2 {
			props := plan.Properties.(*optimizer.JoinProperties)
			return cbo.estimateJoinCost(plan.Children[0], plan.Children[1], props, ctx)
		}
	case optimizer.SelectPlan:
		return cbo.estimateProjectCost(plan, ctx)
	}

	// 默认成本
	return 1000.0
}

// estimateTableScanCost 估算表扫描成本
func (cbo *CostBasedOptimizer) estimateTableScanCost(plan *optimizer.Plan, ctx *OptimizationContext) float64 {
	props := plan.Properties.(*optimizer.TableScanProperties)

	tableStats, err := ctx.statsMgr.GetTableStatistics(props.Table)
	if err != nil {
		return 1000.0 // 默认成本
	}

	// 成本 = 行数 * 扫描因子
	return float64(tableStats.RowCount) * ctx.config.SeqScanCostFactor
}

// estimateFilterCost 估算过滤成本
func (cbo *CostBasedOptimizer) estimateFilterCost(plan *optimizer.Plan, ctx *OptimizationContext) float64 {
	if len(plan.Children) == 0 {
		return 0.0
	}

	childCost := cbo.estimatePlanCost(plan.Children[0], ctx)

	// 过滤成本 = 子节点成本 + 过滤处理成本
	// 简化实现：假设过滤成本是子节点成本的10%
	return childCost * 1.1
}

// estimateProjectCost 估算投影成本
func (cbo *CostBasedOptimizer) estimateProjectCost(plan *optimizer.Plan, ctx *OptimizationContext) float64 {
	if len(plan.Children) == 0 {
		return 0.0
	}

	childCost := cbo.estimatePlanCost(plan.Children[0], ctx)

	// 投影成本通常很小
	return childCost * 1.05
}

// estimateJoinCost 估算连接成本
func (cbo *CostBasedOptimizer) estimateJoinCost(left, right *optimizer.Plan, props *optimizer.JoinProperties, ctx *OptimizationContext) float64 {
	leftRows := cbo.estimateRowCount(left, ctx)
	rightRows := cbo.estimateRowCount(right, ctx)

	// 估算连接选择性
	selectivity := cbo.estimateJoinSelectivity(left, right, props, ctx)

	// 嵌套循环连接成本
	nestedLoopCost := leftRows * rightRows * ctx.config.NestedLoopCostFactor

	// 哈希连接成本
	hashJoinCost := (leftRows + rightRows) * ctx.config.HashJoinCostFactor

	// 选择较小的成本
	joinCost := math.Min(nestedLoopCost, hashJoinCost)

	// 考虑选择性
	return joinCost * selectivity
}

// estimateNestedLoopJoinCost 估算嵌套循环连接成本
func (cbo *CostBasedOptimizer) estimateNestedLoopJoinCost(left, right *optimizer.Plan, props *optimizer.JoinProperties, ctx *OptimizationContext) float64 {
	leftRows := cbo.estimateRowCount(left, ctx)
	rightRows := cbo.estimateRowCount(right, ctx)

	return leftRows * rightRows * ctx.config.NestedLoopCostFactor
}

// estimateHashJoinCost 估算哈希连接成本
func (cbo *CostBasedOptimizer) estimateHashJoinCost(left, right *optimizer.Plan, props *optimizer.JoinProperties, ctx *OptimizationContext) float64 {
	leftRows := cbo.estimateRowCount(left, ctx)
	rightRows := cbo.estimateRowCount(right, ctx)

	// 哈希表构建成本 + 探测成本
	buildCost := leftRows * ctx.config.HashJoinCostFactor
	probeCost := rightRows * ctx.config.HashJoinCostFactor * 0.5

	return buildCost + probeCost
}

// estimateRowCount 估算行数
func (cbo *CostBasedOptimizer) estimateRowCount(plan *optimizer.Plan, ctx *OptimizationContext) float64 {
	switch plan.Type {
	case optimizer.TableScanPlan:
		props := plan.Properties.(*optimizer.TableScanProperties)
		if tableStats, err := ctx.statsMgr.GetTableStatistics(props.Table); err == nil {
			return float64(tableStats.RowCount)
		}
	case optimizer.FilterPlan:
		if len(plan.Children) > 0 {
			childRows := cbo.estimateRowCount(plan.Children[0], ctx)
			// 简化：假设过滤后剩余50%的行
			return childRows * 0.5
		}
	case optimizer.JoinPlan:
		if len(plan.Children) == 2 {
			leftRows := cbo.estimateRowCount(plan.Children[0], ctx)
			rightRows := cbo.estimateRowCount(plan.Children[1], ctx)
			props := plan.Properties.(*optimizer.JoinProperties)
			selectivity := cbo.estimateJoinSelectivity(plan.Children[0], plan.Children[1], props, ctx)
			return leftRows * rightRows * selectivity
		}
	}

	return 1000.0 // 默认行数
}

// estimateJoinSelectivity 估算连接选择性
func (cbo *CostBasedOptimizer) estimateJoinSelectivity(left, right *optimizer.Plan, props *optimizer.JoinProperties, ctx *OptimizationContext) float64 {
	// 简化实现：返回固定选择性
	return 0.1
}

// pushdownFilterToJoin 将过滤条件下推到连接
func (cbo *CostBasedOptimizer) pushdownFilterToJoin(plan *optimizer.Plan, condition optimizer.Expression, ctx *OptimizationContext) (*optimizer.Plan, error) {
	// 分析过滤条件，看能否下推到连接的某一侧
	// 简化实现：不进行下推
	return plan, nil
}

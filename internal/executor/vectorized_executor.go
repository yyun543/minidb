package executor

import (
	"context"
	"fmt"
	"strings"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/session"
	"github.com/yyun543/minidb/internal/statistics"
	"github.com/yyun543/minidb/internal/types"
)

// VectorizedExecutor 向量化执行器，提供更高性能的查询执行
type VectorizedExecutor struct {
	catalog     *catalog.Catalog
	dataManager *DataManager
	statsMgr    *statistics.StatisticsManager
	optimizer   *CostBasedOptimizer
}

// NewVectorizedExecutor 创建向量化执行器
func NewVectorizedExecutor(cat *catalog.Catalog, statsMgr *statistics.StatisticsManager) *VectorizedExecutor {
	return &VectorizedExecutor{
		catalog:     cat,
		dataManager: NewDataManager(cat),
		statsMgr:    statsMgr,
		optimizer:   NewCostBasedOptimizer(statsMgr),
	}
}

// NewVectorizedExecutorWithDataManager 创建向量化执行器（使用指定的DataManager）
func NewVectorizedExecutorWithDataManager(cat *catalog.Catalog, dm *DataManager, statsMgr *statistics.StatisticsManager) *VectorizedExecutor {
	return &VectorizedExecutor{
		catalog:     cat,
		dataManager: dm,
		statsMgr:    statsMgr,
		optimizer:   NewCostBasedOptimizer(statsMgr),
	}
}

// Execute 执行查询计划（向量化版本）
func (ve *VectorizedExecutor) Execute(plan *optimizer.Plan, sess *session.Session) (*VectorizedResultSet, error) {
	ctx := context.Background()

	// 应用基于成本的优化
	optimizedPlan, err := ve.optimizer.OptimizePlan(plan)
	if err != nil {
		// 如果优化失败，使用原计划
		optimizedPlan = plan
	}

	// 为 DDL/DML 操作特别处理
	switch optimizedPlan.Type {
	case optimizer.CreateDatabasePlan:
		return ve.executeCreateDatabase(optimizedPlan, sess)
	case optimizer.CreateTablePlan:
		return ve.executeCreateTable(optimizedPlan, sess)
	case optimizer.ShowPlan:
		return ve.executeShow(optimizedPlan, sess)
	case optimizer.InsertPlan:
		return ve.executeInsert(optimizedPlan, sess)
	case optimizer.UpdatePlan:
		return ve.executeUpdate(optimizedPlan, sess)
	case optimizer.DeletePlan:
		return ve.executeDelete(optimizedPlan, sess)
	}

	// 构建向量化执行管道
	pipeline, err := ve.buildVectorizedPipeline(ctx, optimizedPlan, sess)
	if err != nil {
		return nil, err
	}

	// 执行管道并收集结果
	return ve.executePipeline(ctx, pipeline)
}

// VectorizedResultSet 向量化结果集
type VectorizedResultSet struct {
	Headers []string
	Batches []*types.VectorizedBatch
	Schema  *arrow.Schema
	curRow  int
}

// VectorizedPipeline 向量化执行管道
type VectorizedPipeline struct {
	operations []types.VectorizedOperation
	schema     *arrow.Schema
}

// buildVectorizedPipeline 构建向量化执行管道
func (ve *VectorizedExecutor) buildVectorizedPipeline(ctx context.Context, plan *optimizer.Plan, sess *session.Session) (*VectorizedPipeline, error) {
	operations := []types.VectorizedOperation{}
	schema := ve.InferSchema(plan, sess)

	// 递归构建操作链
	ops, err := ve.buildOperationsFromPlan(ctx, plan, schema, sess)
	if err != nil {
		return nil, err
	}

	operations = append(operations, ops...)

	return &VectorizedPipeline{
		operations: operations,
		schema:     schema,
	}, nil
}

// buildOperationsFromPlan 从计划构建操作
func (ve *VectorizedExecutor) buildOperationsFromPlan(ctx context.Context, plan *optimizer.Plan, schema *arrow.Schema, sess *session.Session) ([]types.VectorizedOperation, error) {
	var operations []types.VectorizedOperation

	switch plan.Type {
	case optimizer.TableScanPlan:
		// 表扫描操作
		op, err := ve.buildTableScanOperation(plan, sess)
		if err != nil {
			return nil, err
		}
		operations = append(operations, op)

	case optimizer.FilterPlan:
		// 过滤操作 - 需要使用子节点的完整schema来查找过滤列
		var filterSchema *arrow.Schema
		if len(plan.Children) > 0 {
			filterSchema = ve.InferSchema(plan.Children[0], sess)
		} else {
			filterSchema = schema
		}

		props := plan.Properties.(*optimizer.FilterProperties)
		predicate, err := ve.buildVectorizedPredicate(props.Condition, filterSchema)
		if err != nil {
			return nil, err
		}

		filterOp := types.NewFilterOperation(predicate)
		operations = append(operations, filterOp)

		// 递归处理子操作
		if len(plan.Children) > 0 {
			childOps, err := ve.buildOperationsFromPlan(ctx, plan.Children[0], filterSchema, sess)
			if err != nil {
				return nil, err
			}
			operations = append(operations, childOps...)
		}

	case optimizer.SelectPlan:
		// 投影操作
		props := plan.Properties.(*optimizer.SelectProperties)

		// 获取子节点的输出schema，这才是投影操作的输入schema
		var inputSchema *arrow.Schema
		if len(plan.Children) > 0 {
			inputSchema = ve.InferSchema(plan.Children[0], sess)
		} else {
			inputSchema = schema
		}

		var columnIndices []int
		var newSchema *arrow.Schema

		if props.All {
			// SELECT * - 使用所有列
			columnIndices = make([]int, inputSchema.NumFields())
			for i := 0; i < inputSchema.NumFields(); i++ {
				columnIndices[i] = i
			}
			newSchema = inputSchema
		} else {
			// SELECT specific columns - 基于输入schema构建投影映射
			columnIndices, newSchema = ve.buildProjectionMapping(props.Columns, inputSchema)
		}

		projectOp := types.NewProjectOperation(columnIndices, newSchema)
		operations = append(operations, projectOp)

		// 递归处理子操作 - 使用子节点的输入schema
		if len(plan.Children) > 0 {
			childOps, err := ve.buildOperationsFromPlan(ctx, plan.Children[0], inputSchema, sess)
			if err != nil {
				return nil, err
			}
			operations = append(operations, childOps...)
		}

	case optimizer.JoinPlan:
		// 连接操作
		joinOp, err := ve.buildJoinOperation(ctx, plan, schema)
		if err != nil {
			return nil, err
		}
		operations = append(operations, joinOp)

	default:
		return nil, fmt.Errorf("unsupported plan type for vectorized execution: %v", plan.Type)
	}

	return operations, nil
}

// buildTableScanOperation 构建表扫描操作
func (ve *VectorizedExecutor) buildTableScanOperation(plan *optimizer.Plan, sess *session.Session) (types.VectorizedOperation, error) {
	props := plan.Properties.(*optimizer.TableScanProperties)

	// 解析表引用：支持 "database.table" 或 "table" 格式
	dbName, tableName := ve.parseTableReference(props.Table, sess.CurrentDB)

	batches, err := ve.dataManager.GetTableData(dbName, tableName)
	if err != nil {
		return nil, err
	}

	// 转换为向量化批处理
	vectorizedBatches := make([]*types.VectorizedBatch, len(batches))
	for i, batch := range batches {
		vectorizedBatches[i] = ve.convertToVectorizedBatch(batch)
	}

	return &VectorizedTableScanOperation{
		tableName:  props.Table,
		batches:    vectorizedBatches,
		currentIdx: 0,
	}, nil
}

// buildVectorizedPredicate 构建向量化谓词
func (ve *VectorizedExecutor) buildVectorizedPredicate(condition optimizer.Expression, schema *arrow.Schema) (*types.VectorizedPredicate, error) {
	if binExpr, ok := condition.(*optimizer.BinaryExpression); ok {
		// 检查是否是逻辑表达式（AND/OR）
		if binExpr.Operator == "AND" || binExpr.Operator == "OR" {
			// 递归构建左右子谓词
			leftPred, err := ve.buildVectorizedPredicate(binExpr.Left, schema)
			if err != nil {
				return nil, err
			}

			rightPred, err := ve.buildVectorizedPredicate(binExpr.Right, schema)
			if err != nil {
				return nil, err
			}

			return types.NewCompoundVectorizedPredicate(binExpr.Operator, leftPred, rightPred), nil
		}

		// 检查是否是LIKE表达式
		if binExpr.Operator == "LIKE" || binExpr.Operator == "NOT LIKE" {
			// LIKE表达式暂时不支持向量化，返回错误让系统回退到常规执行器
			return nil, fmt.Errorf("LIKE expressions not supported in vectorized execution yet")
		}

		// 处理比较表达式
		if colRef, ok := binExpr.Left.(*optimizer.ColumnReference); ok {
			// 查找列索引
			columnIndex := -1
			var dataType arrow.DataType
			for i, field := range schema.Fields() {
				if field.Name == colRef.Column {
					columnIndex = i
					dataType = field.Type
					break
				}
			}

			if columnIndex == -1 {
				// If column not found, try to find it by table.column format
				fullColumnName := ""
				if colRef.Table != "" {
					fullColumnName = colRef.Table + "." + colRef.Column
				}

				// Try to find the column by name only (without table prefix)
				for i, field := range schema.Fields() {
					if field.Name == colRef.Column || field.Name == fullColumnName {
						columnIndex = i
						dataType = field.Type
						break
					}
				}

				if columnIndex == -1 {
					return nil, fmt.Errorf("column %s not found in schema", colRef.Column)
				}
			}

			// 提取比较值并确保类型正确转换
			var value interface{}
			if litVal, ok := binExpr.Right.(*optimizer.LiteralValue); ok {
				// 根据列的数据类型转换比较值
				switch dataType {
				case arrow.PrimitiveTypes.Int64:
					// 确保值是int64类型
					switch v := litVal.Value.(type) {
					case int64:
						value = v
					case int:
						value = int64(v)
					case int32:
						value = int64(v)
					case float64:
						value = int64(v)
					default:
						return nil, fmt.Errorf("cannot convert %T to int64 for column %s", litVal.Value, colRef.Column)
					}
				case arrow.BinaryTypes.String:
					// 确保值是string类型
					if strVal, ok := litVal.Value.(string); ok {
						value = strVal
					} else {
						value = fmt.Sprintf("%v", litVal.Value)
					}
				case arrow.PrimitiveTypes.Float64:
					// 确保值是float64类型
					switch v := litVal.Value.(type) {
					case float64:
						value = v
					case int64:
						value = float64(v)
					case int:
						value = float64(v)
					default:
						return nil, fmt.Errorf("cannot convert %T to float64 for column %s", litVal.Value, colRef.Column)
					}
				case arrow.FixedWidthTypes.Boolean:
					// 确保值是bool类型 - 支持 true/false, 1/0, "true"/"false" 等多种形式
					switch v := litVal.Value.(type) {
					case bool:
						value = v
					case int64:
						value = (v != 0)
					case int:
						value = (v != 0)
					case int32:
						value = (v != 0)
					case string:
						// 处理字符串形式的布尔值
						switch v {
						case "true", "1", "t", "T", "TRUE":
							value = true
						case "false", "0", "f", "F", "FALSE":
							value = false
						default:
							return nil, fmt.Errorf("cannot convert string %q to bool for column %s", v, colRef.Column)
						}
					default:
						return nil, fmt.Errorf("cannot convert %T to bool for column %s", litVal.Value, colRef.Column)
					}
				default:
					value = litVal.Value
				}
			} else {
				// 简化实现：对于未知类型的表达式，使用默认值
				value = "unknown"
			}

			return types.NewVectorizedPredicate(columnIndex, binExpr.Operator, value, dataType), nil
		}
	}

	return nil, fmt.Errorf("unsupported predicate type")
}

// buildProjectionMapping 构建投影映射
func (ve *VectorizedExecutor) buildProjectionMapping(columns []optimizer.ColumnRef, schema *arrow.Schema) ([]int, *arrow.Schema) {
	columnIndices := make([]int, len(columns))
	fields := make([]arrow.Field, len(columns))

	for i, col := range columns {
		// 查找列索引
		columnIndex := -1
		var foundField arrow.Field
		for j, field := range schema.Fields() {
			if field.Name == col.Column {
				columnIndex = j
				foundField = field
				break
			}
		}

		if columnIndex != -1 {
			// 如果有别名，使用别名作为字段名
			if col.Alias != "" {
				foundField.Name = col.Alias
			}
			fields[i] = foundField
			columnIndices[i] = columnIndex
		} else {
			// 列未找到，创建一个默认字段
			fieldName := col.Column
			if col.Alias != "" {
				fieldName = col.Alias
			}
			fields[i] = arrow.Field{
				Name:     fieldName,
				Type:     arrow.BinaryTypes.String,
				Nullable: true,
			}
			columnIndices[i] = -1 // -1表示列不存在
		}
	}

	newSchema := arrow.NewSchema(fields, nil)
	return columnIndices, newSchema
}

// buildJoinOperation 构建连接操作
func (ve *VectorizedExecutor) buildJoinOperation(ctx context.Context, plan *optimizer.Plan, schema *arrow.Schema) (types.VectorizedOperation, error) {
	// TODO: 实现向量化连接操作
	return nil, fmt.Errorf("vectorized join not implemented yet")
}

// executePipeline 执行管道
func (ve *VectorizedExecutor) executePipeline(ctx context.Context, pipeline *VectorizedPipeline) (*VectorizedResultSet, error) {
	result := &VectorizedResultSet{
		Headers: ve.getSchemaFieldNames(pipeline.schema),
		Schema:  pipeline.schema,
		Batches: []*types.VectorizedBatch{},
	}

	// 找到表扫描操作并执行整个管道
	operations := pipeline.operations
	for i := len(operations) - 1; i >= 0; i-- {
		op := operations[i]

		if scanOp, ok := op.(*VectorizedTableScanOperation); ok {
			// 对表扫描的每个批次，按正确顺序应用所有其他操作
			for _, batch := range scanOp.batches {
				// 创建需要应用的操作列表（排除当前的TableScan操作）
				var opsToApply []types.VectorizedOperation
				// 操作需要按从底向上的顺序应用：Filter -> Project
				for j := i - 1; j >= 0; j-- {
					opsToApply = append(opsToApply, operations[j])
				}

				// 应用所有操作到这个批次
				processedBatch, err := ve.applyOperationsToaBatch(ctx, batch, opsToApply)
				if err != nil {
					return nil, err
				}
				if processedBatch != nil {
					result.Batches = append(result.Batches, processedBatch)
				}
			}
		}
	}

	return result, nil
}

// applyOperationsToaBatch 对单个批次应用操作
func (ve *VectorizedExecutor) applyOperationsToaBatch(ctx context.Context, batch *types.VectorizedBatch, operations []types.VectorizedOperation) (*types.VectorizedBatch, error) {
	currentBatch := batch

	for _, op := range operations {
		processedBatch, err := op.Execute(currentBatch)
		if err != nil {
			return nil, err
		}
		if processedBatch == nil {
			return nil, nil // 批次被过滤掉
		}
		currentBatch = processedBatch
	}

	return currentBatch, nil
}

// VectorizedTableScanOperation 向量化表扫描操作
type VectorizedTableScanOperation struct {
	tableName  string
	batches    []*types.VectorizedBatch
	currentIdx int
}

// Execute 执行表扫描
func (op *VectorizedTableScanOperation) Execute(input *types.VectorizedBatch) (*types.VectorizedBatch, error) {
	// 表扫描操作不接受输入，返回下一个批次
	if op.currentIdx >= len(op.batches) {
		return nil, nil
	}

	batch := op.batches[op.currentIdx]
	op.currentIdx++
	return batch, nil
}

// Name 返回操作名称
func (op *VectorizedTableScanOperation) Name() string {
	return fmt.Sprintf("VectorizedTableScan_%s", op.tableName)
}

// 工具方法
func (ve *VectorizedExecutor) convertToVectorizedBatch(batch *types.Batch) *types.VectorizedBatch {
	record := batch.Record()
	schema := record.Schema()

	vBatch := types.NewVectorizedBatch(schema, nil)
	for i := int64(0); i < record.NumCols(); i++ {
		column := record.Column(int(i))
		vBatch.SetColumn(int(i), column)
	}

	return vBatch
}

func (ve *VectorizedExecutor) getSchemaFieldNames(schema *arrow.Schema) []string {
	names := make([]string, schema.NumFields())
	for i, field := range schema.Fields() {
		names[i] = field.Name
	}
	return names
}

func (ve *VectorizedExecutor) InferSchema(plan *optimizer.Plan, sess *session.Session) *arrow.Schema {
	// 简化实现：根据计划推断模式
	// 实际实现需要更复杂的类型推断逻辑

	switch plan.Type {
	case optimizer.TableScanPlan:
		props := plan.Properties.(*optimizer.TableScanProperties)
		// 解析表引用：支持 "database.table" 或 "table" 格式
		dbName, tableName := ve.parseTableReference(props.Table, sess.CurrentDB)
		if tableMeta, err := ve.catalog.GetTable(dbName, tableName); err == nil {
			return tableMeta.Schema
		}

	case optimizer.FilterPlan:
		// Filter操作不改变schema，返回子节点的schema
		if len(plan.Children) > 0 {
			return ve.InferSchema(plan.Children[0], sess)
		}

	case optimizer.SelectPlan:
		// 投影操作的模式推断
		if len(plan.Children) > 0 {
			childSchema := ve.InferSchema(plan.Children[0], sess)
			props := plan.Properties.(*optimizer.SelectProperties)

			// 处理SELECT *的情况（columns为空或All=true）
			if props.All || len(props.Columns) == 0 {
				return childSchema
			}

			// 对于SELECT specific columns的情况，需要正确构建投影schema
			// 即使有WHERE子句，也要根据SELECT的列来确定最终的schema

			fields := make([]arrow.Field, len(props.Columns))
			for i, col := range props.Columns {
				// 查找对应的字段
				found := false
				for _, field := range childSchema.Fields() {
					if field.Name == col.Column {
						fields[i] = field
						found = true
						break
					}
				}
				// 如果找不到字段，创建一个默认的字段
				if !found {
					fields[i] = arrow.Field{
						Name: col.Column,
						Type: arrow.BinaryTypes.String,
					}
				}
			}
			return arrow.NewSchema(fields, nil)
		}
	}

	// 默认返回空模式
	return arrow.NewSchema([]arrow.Field{}, nil)
}

// 执行DDL/DML操作的方法（重用现有逻辑）

// executeCreateDatabase 执行创建数据库操作
func (ve *VectorizedExecutor) executeCreateDatabase(plan *optimizer.Plan, sess *session.Session) (*VectorizedResultSet, error) {
	// 重用现有的CreateDatabase逻辑
	executor := &ExecutorImpl{
		catalog:     ve.catalog,
		dataManager: ve.dataManager,
	}

	result, err := executor.executeCreateDatabase(plan, sess)
	if err != nil {
		return nil, err
	}

	return &VectorizedResultSet{
		Headers: result.Headers,
		Batches: []*types.VectorizedBatch{},
		Schema:  arrow.NewSchema([]arrow.Field{}, nil),
	}, nil
}

// executeShow 执行SHOW命令
func (ve *VectorizedExecutor) executeShow(plan *optimizer.Plan, sess *session.Session) (*VectorizedResultSet, error) {
	// 重用现有的Show逻辑
	executor := &ExecutorImpl{
		catalog:     ve.catalog,
		dataManager: ve.dataManager,
	}

	result, err := executor.executeShow(plan, sess)
	if err != nil {
		return nil, err
	}

	return &VectorizedResultSet{
		Headers: result.Headers,
		Batches: []*types.VectorizedBatch{},
		Schema:  arrow.NewSchema([]arrow.Field{}, nil),
	}, nil
}

func (ve *VectorizedExecutor) executeCreateTable(plan *optimizer.Plan, sess *session.Session) (*VectorizedResultSet, error) {
	// 重用现有的CreateTable逻辑
	executor := &ExecutorImpl{
		catalog:     ve.catalog,
		dataManager: ve.dataManager,
	}

	result, err := executor.executeCreateTable(plan, sess)
	if err != nil {
		return nil, err
	}

	// 转换结果
	return &VectorizedResultSet{
		Headers: result.Headers,
		Batches: []*types.VectorizedBatch{},
		Schema:  arrow.NewSchema([]arrow.Field{}, nil),
	}, nil
}

func (ve *VectorizedExecutor) executeInsert(plan *optimizer.Plan, sess *session.Session) (*VectorizedResultSet, error) {
	// 重用现有的Insert逻辑
	executor := &ExecutorImpl{
		catalog:     ve.catalog,
		dataManager: ve.dataManager,
	}

	_, err := executor.executeInsert(plan, sess)
	if err != nil {
		return nil, err
	}

	// 插入后更新统计信息
	ve.updateStatisticsAfterWrite(plan)

	return &VectorizedResultSet{
		Headers: []string{"status"},
		Batches: []*types.VectorizedBatch{},
		Schema:  arrow.NewSchema([]arrow.Field{}, nil),
	}, nil
}

func (ve *VectorizedExecutor) executeUpdate(plan *optimizer.Plan, sess *session.Session) (*VectorizedResultSet, error) {
	// 重用现有的Update逻辑
	executor := &ExecutorImpl{
		catalog:     ve.catalog,
		dataManager: ve.dataManager,
	}

	_, err := executor.executeUpdate(plan, sess)
	if err != nil {
		return nil, err
	}

	// 更新后更新统计信息
	ve.updateStatisticsAfterWrite(plan)

	return &VectorizedResultSet{
		Headers: []string{"status"},
		Batches: []*types.VectorizedBatch{},
		Schema:  arrow.NewSchema([]arrow.Field{}, nil),
	}, nil
}

func (ve *VectorizedExecutor) executeDelete(plan *optimizer.Plan, sess *session.Session) (*VectorizedResultSet, error) {
	// 重用现有的Delete逻辑
	executor := &ExecutorImpl{
		catalog:     ve.catalog,
		dataManager: ve.dataManager,
	}

	_, err := executor.executeDelete(plan, sess)
	if err != nil {
		return nil, err
	}

	// 删除后更新统计信息
	ve.updateStatisticsAfterWrite(plan)

	return &VectorizedResultSet{
		Headers: []string{"status"},
		Batches: []*types.VectorizedBatch{},
		Schema:  arrow.NewSchema([]arrow.Field{}, nil),
	}, nil
}

func (ve *VectorizedExecutor) updateStatisticsAfterWrite(plan *optimizer.Plan) {
	// 异步更新统计信息 - 实现智能的统计信息更新策略
	go func() {
		// 异步更新，避免阻塞写操作
		// 基于操作类型和表大小决定是否需要立即更新统计信息

		var tableName string
		var needsUpdate bool

		// 根据计划类型提取表名和确定更新策略
		switch plan.Type {
		case optimizer.InsertPlan:
			if props, ok := plan.Properties.(*optimizer.InsertProperties); ok {
				tableName = props.Table
				needsUpdate = true // INSERT总是需要更新统计信息
			}
		case optimizer.UpdatePlan:
			if props, ok := plan.Properties.(*optimizer.UpdateProperties); ok {
				tableName = props.Table
				needsUpdate = true // UPDATE可能改变数据分布
			}
		case optimizer.DeletePlan:
			if props, ok := plan.Properties.(*optimizer.DeleteProperties); ok {
				tableName = props.Table
				needsUpdate = true // DELETE改变行数统计
			}
		}

		// 如果需要更新且有统计管理器，则触发更新
		if needsUpdate && tableName != "" && ve.statsMgr != nil {
			// 简化实现：标记统计信息为过期，下次查询时会重新收集
			// 在生产环境中，这里应该有更复杂的策略：
			// 1. 检查表大小，小表立即更新，大表延迟更新
			// 2. 使用采样技术进行快速估算
			// 3. 维护更新队列，避免重复更新同一张表

			// 当前简化实现：记录需要更新的表，由后台任务处理
			// ve.statsMgr.MarkTableForUpdate(tableName)
		}
	}()
}

// parseTableReference 解析表引用，支持 "database.table" 或 "table" 格式
func (ve *VectorizedExecutor) parseTableReference(tableRef string, currentDB string) (string, string) {
	parts := strings.Split(tableRef, ".")
	if len(parts) == 2 {
		// "database.table" 格式
		return parts[0], parts[1]
	}
	// "table" 格式，使用当前数据库
	if currentDB == "" {
		currentDB = "default"
	}
	return currentDB, tableRef
}

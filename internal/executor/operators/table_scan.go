package operators

import (
	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/types"
)

// DataProvider 数据提供者接口
// 用于提供表数据，支持系统表和普通表的统一访问
type DataProvider interface {
	GetTableData(dbName, tableName string) ([]*types.Batch, error)
}

// TableScan 表扫描算子 (v2.0)
// 使用 DataProvider 统一获取系统表和普通表数据
type TableScan struct {
	database     string
	table        string
	catalog      *catalog.Catalog
	dataProvider DataProvider
	schema       *arrow.Schema
	pool         *memory.GoAllocator
	batchSize    int
	curBatch     int
	dataBatches  []*types.Batch
}

// NewTableScan 创建表扫描算子
func NewTableScan(database, table string, catalog *catalog.Catalog, dataProvider DataProvider) *TableScan {
	return &TableScan{
		database:     database,
		table:        table,
		catalog:      catalog,
		dataProvider: dataProvider,
		pool:         memory.NewGoAllocator(),
		batchSize:    1024,
	}
}

// Init 初始化算子 (v2.0)
func (op *TableScan) Init(ctx interface{}) error {
	// 获取表结构
	table, err := op.catalog.GetTable(op.database, op.table)
	if err != nil {
		return err
	}

	// 使用表的 Schema
	op.schema = table.Schema

	// 从 DataProvider 读取数据 (统一处理系统表和普通表)
	if ctx != nil {
		batches, err := op.getTableData()
		if err != nil {
			return err
		}
		op.dataBatches = batches
	}

	return nil
}

// Next 获取下一批数据
func (op *TableScan) Next() (*types.Batch, error) {
	// 检查是否还有数据批次要返回
	if op.curBatch >= len(op.dataBatches) {
		return nil, nil // 表示没有更多数据
	}

	// 获取当前批次数据
	batch := op.dataBatches[op.curBatch]
	op.curBatch++

	return batch, nil
}

// Close 关闭算子
func (op *TableScan) Close() error {
	return nil
}

// getTableData 从 DataProvider 获取表数据 (v2.0)
// 统一处理系统表和普通表，不再区分
func (op *TableScan) getTableData() ([]*types.Batch, error) {
	if op.dataProvider == nil {
		// 如果没有 DataProvider，返回空结果
		return []*types.Batch{}, nil
	}

	// 使用 DataProvider.GetTableData() 获取表数据
	// DataProvider 内部会判断是系统表还是普通表，并采用相应的方式获取数据
	batches, err := op.dataProvider.GetTableData(op.database, op.table)
	if err != nil {
		return nil, err
	}

	return batches, nil
}

// convertType 转换数据类型到Arrow类型
func (op *TableScan) convertType(t string) arrow.DataType {
	switch t {
	case "INT":
		return arrow.PrimitiveTypes.Int64
	case "STRING":
		return arrow.BinaryTypes.String
	// 可以添加更多类型支持
	default:
		return arrow.BinaryTypes.String
	}
}

package operators

import (
	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/types"
)

// TableScan 表扫描算子
type TableScan struct {
	table     string
	catalog   *catalog.Catalog
	schema    *arrow.Schema
	pool      *memory.GoAllocator
	batchSize int
	curBatch  int
}

// NewTableScan 创建表扫描算子
func NewTableScan(table string, catalog *catalog.Catalog) *TableScan {
	return &TableScan{
		table:     table,
		catalog:   catalog,
		pool:      memory.NewGoAllocator(),
		batchSize: 1024,
	}
}

// Init 初始化算子
func (op *TableScan) Init(ctx interface{}) error {
	// 获取表结构
	table, err := op.catalog.GetTable(op.table)
	if err != nil {
		return err
	}

	// 构建Arrow Schema
	fields := make([]arrow.Field, len(table.Columns))
	for i, col := range table.Columns {
		fields[i] = arrow.Field{Name: col.Name, Type: op.convertType(col.Type)}
	}
	op.schema = arrow.NewSchema(fields, nil)
	return nil
}

// Next 获取下一批数据
func (op *TableScan) Next() (*types.Batch, error) {
	// 从存储引擎读取数据
	table, err := op.catalog.GetTable(op.table)
	if err != nil {
		return nil, err
	}
	if table == nil {
		return nil, nil
	}

	// 构建RecordBatch
	builders := make([]array.Builder, len(op.schema.Fields()))
	for i, field := range op.schema.Fields() {
		builders[i] = array.NewBuilder(op.pool, field.Type)
	}

	// 读取数据并构建RecordBatch
	batch := types.NewBatch(op.schema, op.batchSize)
	// TODO: 实现实际的数据读取逻辑

	return batch, nil
}

// Close 关闭算子
func (op *TableScan) Close() error {
	return nil
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

package operators

import (
	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/storage"
	"github.com/yyun543/minidb/internal/types"
)

// TableScan 表扫描算子
type TableScan struct {
	database    string
	table       string
	catalog     *catalog.Catalog
	schema      *arrow.Schema
	pool        *memory.GoAllocator
	batchSize   int
	curBatch    int
	dataBatches []*types.Batch
}

// NewTableScan 创建表扫描算子
func NewTableScan(database, table string, catalog *catalog.Catalog) *TableScan {
	return &TableScan{
		database:  database,
		table:     table,
		catalog:   catalog,
		pool:      memory.NewGoAllocator(),
		batchSize: 1024,
	}
}

// Init 初始化算子
func (op *TableScan) Init(ctx interface{}) error {
	// 获取表结构
	table, err := op.catalog.GetTable(op.database, op.table)
	if err != nil {
		return err
	}

	// 使用表的 Schema
	op.schema = table.Schema

	// 直接从存储引擎读取数据
	if ctxType := ctx; ctxType != nil {
		batches, err := op.getTableDataFromStorage()
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

// getTableDataFromStorage 从存储引擎获取表数据
func (op *TableScan) getTableDataFromStorage() ([]*types.Batch, error) {
	// 获取表元数据
	tableMeta, err := op.catalog.GetTable(op.database, op.table)
	if err != nil {
		return nil, err
	}

	// 创建KeyManager
	keyManager := storage.NewKeyManager()
	engine := op.catalog.GetEngine()

	// 获取所有数据块
	var batches []*types.Batch
	maxChunks := tableMeta.ChunkCount
	if maxChunks == 0 {
		maxChunks = 1
	}

	for i := int64(0); i < maxChunks; i++ {
		key := keyManager.TableChunkKey(op.database, op.table, i)
		record, err := engine.Get(key)
		if err != nil {
			return nil, err
		}

		if record != nil && record.NumRows() > 0 {
			batch := types.NewBatch(record)
			batches = append(batches, batch)
		}
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

package operators

import (
	"github.com/yyun543/minidb/internal/types"
)

// Limit LIMIT算子
type Limit struct {
	limit      int64    // LIMIT数量
	offset     int64    // OFFSET数量（可选，默认0）
	child      Operator // 子算子
	ctx        interface{}
	rowsRead   int64 // 已读取的行数
	rowsOutput int64 // 已输出的行数
}

// NewLimit 创建LIMIT算子
func NewLimit(limit int64, offset int64, child Operator, ctx interface{}) *Limit {
	return &Limit{
		limit:      limit,
		offset:     offset,
		child:      child,
		ctx:        ctx,
		rowsRead:   0,
		rowsOutput: 0,
	}
}

// Init 初始化算子
func (op *Limit) Init(ctx interface{}) error {
	return op.child.Init(ctx)
}

// Next 获取下一批数据
func (op *Limit) Next() (*types.Batch, error) {
	// 如果已经输出了足够的行，返回nil
	if op.rowsOutput >= op.limit {
		return nil, nil
	}

	for {
		batch, err := op.child.Next()
		if err != nil {
			return nil, err
		}
		if batch == nil {
			return nil, nil
		}

		batchRows := batch.NumRows()

		// 如果还在跳过offset范围内
		if op.rowsRead < op.offset {
			remainingOffset := op.offset - op.rowsRead
			if batchRows <= remainingOffset {
				// 整个batch都在offset范围内，跳过
				op.rowsRead += batchRows
				continue
			} else {
				// batch部分在offset范围内
				// 需要从batch中取出offset之后的行
				startRow := int(remainingOffset)
				op.rowsRead += remainingOffset

				// 计算需要输出的行数
				remainingLimit := op.limit - op.rowsOutput
				outputRows := batchRows - int64(startRow)
				if outputRows > remainingLimit {
					outputRows = remainingLimit
				}

				// 从batch中提取需要的行
				limitedBatch, err := extractRows(batch, startRow, int(outputRows))
				if err != nil {
					return nil, err
				}

				op.rowsRead += outputRows
				op.rowsOutput += outputRows
				return limitedBatch, nil
			}
		}

		// 已经过了offset阶段，现在需要限制输出行数
		remainingLimit := op.limit - op.rowsOutput
		if batchRows <= remainingLimit {
			// 整个batch都需要输出
			op.rowsRead += batchRows
			op.rowsOutput += batchRows
			return batch, nil
		} else {
			// 只需要输出batch的一部分
			limitedBatch, err := extractRows(batch, 0, int(remainingLimit))
			if err != nil {
				return nil, err
			}

			op.rowsRead += remainingLimit
			op.rowsOutput += remainingLimit
			return limitedBatch, nil
		}
	}
}

// Close 关闭算子
func (op *Limit) Close() error {
	return op.child.Close()
}

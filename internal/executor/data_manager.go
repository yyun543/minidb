package executor

import (
	"fmt"
	"sync"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/storage"
	"github.com/yyun543/minidb/internal/types"
)

// DataManager 管理表数据的存储和查询
type DataManager struct {
	catalog    *catalog.Catalog
	engine     storage.Engine
	keyManager *storage.KeyManager
	mu         sync.RWMutex
}

// NewDataManager 创建新的数据管理器
func NewDataManager(catalog *catalog.Catalog) *DataManager {
	return &DataManager{
		catalog:    catalog,
		engine:     catalog.GetEngine(),
		keyManager: storage.NewKeyManager(),
	}
}

// InsertData 插入数据到表中
func (dm *DataManager) InsertData(dbName, tableName string, columns []string, values []interface{}) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 获取表元数据
	tableMeta, err := dm.catalog.GetTable(dbName, tableName)
	if err != nil {
		return fmt.Errorf("table not found: %w", err)
	}

	// 获取现有数据
	chunks, err := dm.getTableChunks(dbName, tableName, tableMeta.ChunkCount)
	if err != nil {
		return err
	}

	// 创建新记录
	newRecord, err := dm.createRecord(tableMeta.Schema, columns, values)
	if err != nil {
		return err
	}
	defer newRecord.Release()

	// 如果有现有数据，合并记录
	var finalRecord arrow.Record
	if len(chunks) > 0 && chunks[0] != nil {
		finalRecord, err = dm.mergeRecords(chunks[0].Record(), newRecord)
		if err != nil {
			return err
		}
		defer finalRecord.Release()
	} else {
		newRecord.Retain()
		finalRecord = newRecord
	}

	// 创建Chunk并存储
	chunk := types.NewChunk(finalRecord, 0, int64(finalRecord.NumRows()))
	defer chunk.Release()

	key := dm.keyManager.TableChunkKey(dbName, tableName, 0)
	record := chunk.Record()
	err = dm.engine.Put(key, &record)
	if err != nil {
		return err
	}

	// 更新表元数据中的ChunkCount（如果需要）
	if tableMeta.ChunkCount == 0 {
		// 更新catalog中的表信息
		tableMeta.ChunkCount = 1
		err := dm.catalog.UpdateTable(dbName, tableMeta)
		if err != nil {
			return fmt.Errorf("failed to update table metadata: %w", err)
		}
	}

	return nil
}

// GetTableData 获取表的所有数据
func (dm *DataManager) GetTableData(dbName, tableName string) ([]*types.Batch, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	// 获取表元数据
	tableMeta, err := dm.catalog.GetTable(dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("table not found: %w", err)
	}

	// 获取所有数据块
	chunks, err := dm.getTableChunks(dbName, tableName, tableMeta.ChunkCount)
	if err != nil {
		return nil, err
	}

	// 转换为Batch
	var batches []*types.Batch
	for _, chunk := range chunks {
		if chunk != nil && chunk.NumRows() > 0 {
			batch := types.NewBatch(chunk.Record())
			batches = append(batches, batch)
		}
	}

	return batches, nil
}

// UpdateData 更新表中的数据
func (dm *DataManager) UpdateData(dbName, tableName string, assignments map[string]interface{}, whereCondition func(arrow.Record, int) bool) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 获取表元数据
	tableMeta, err := dm.catalog.GetTable(dbName, tableName)
	if err != nil {
		return fmt.Errorf("table not found: %w", err)
	}

	// 获取现有数据
	chunks, err := dm.getTableChunks(dbName, tableName, tableMeta.ChunkCount)
	if err != nil {
		return err
	}

	if len(chunks) == 0 || chunks[0] == nil {
		return nil // 没有数据可更新
	}

	oldRecord := chunks[0].Record()
	newRecord, err := dm.updateRecord(oldRecord, assignments, whereCondition)
	if err != nil {
		return err
	}
	defer newRecord.Release()

	// 存储更新后的数据
	chunk := types.NewChunk(newRecord, 0, int64(newRecord.NumRows()))
	defer chunk.Release()

	key := dm.keyManager.TableChunkKey(dbName, tableName, 0)
	record := chunk.Record()
	return dm.engine.Put(key, &record)
}

// DeleteData 删除表中的数据
func (dm *DataManager) DeleteData(dbName, tableName string, whereCondition func(arrow.Record, int) bool) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 获取表元数据
	tableMeta, err := dm.catalog.GetTable(dbName, tableName)
	if err != nil {
		return fmt.Errorf("table not found: %w", err)
	}

	// 获取现有数据
	chunks, err := dm.getTableChunks(dbName, tableName, tableMeta.ChunkCount)
	if err != nil {
		return err
	}

	if len(chunks) == 0 || chunks[0] == nil {
		return nil // 没有数据可删除
	}

	oldRecord := chunks[0].Record()
	newRecord, err := dm.filterRecord(oldRecord, whereCondition)
	if err != nil {
		return err
	}
	defer newRecord.Release()

	// 存储过滤后的数据
	chunk := types.NewChunk(newRecord, 0, int64(newRecord.NumRows()))
	defer chunk.Release()

	key := dm.keyManager.TableChunkKey(dbName, tableName, 0)
	record := chunk.Record()
	return dm.engine.Put(key, &record)
}

// 获取表的所有数据块
func (dm *DataManager) getTableChunks(dbName, tableName string, chunkCount int64) ([]*types.Chunk, error) {
	var chunks []*types.Chunk

	// 至少尝试读取第一个chunk
	maxChunks := chunkCount
	if maxChunks == 0 {
		maxChunks = 1
	}

	for i := int64(0); i < maxChunks; i++ {
		key := dm.keyManager.TableChunkKey(dbName, tableName, i)
		record, err := dm.engine.Get(key)
		if err != nil {
			return nil, err
		}

		if record != nil {
			chunk := types.NewChunk(record, i, int64(record.NumRows()))
			chunks = append(chunks, chunk)
		} else {
			chunks = append(chunks, nil)
		}
	}

	return chunks, nil
}

// 创建新记录
func (dm *DataManager) createRecord(schema *arrow.Schema, columns []string, values []interface{}) (arrow.Record, error) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, schema)
	defer builder.Release()

	// 创建列名到索引的映射
	fieldMap := make(map[string]int)
	for i, field := range schema.Fields() {
		fieldMap[field.Name] = i
	}

	// 为每个字段添加值
	for i, column := range columns {
		if fieldIdx, exists := fieldMap[column]; exists {
			field := builder.Field(fieldIdx)
			if i < len(values) {
				err := dm.appendValue(field, values[i])
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return builder.NewRecord(), nil
}

// 合并两个记录
func (dm *DataManager) mergeRecords(oldRecord, newRecord arrow.Record) (arrow.Record, error) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, oldRecord.Schema())
	defer builder.Release()

	// 复制旧数据
	for colIdx := int64(0); colIdx < oldRecord.NumCols(); colIdx++ {
		oldCol := oldRecord.Column(int(colIdx))
		field := builder.Field(int(colIdx))

		for rowIdx := int64(0); rowIdx < oldRecord.NumRows(); rowIdx++ {
			value := dm.getValueFromArray(oldCol, int(rowIdx))
			err := dm.appendValue(field, value)
			if err != nil {
				return nil, err
			}
		}
	}

	// 复制新数据
	for colIdx := int64(0); colIdx < newRecord.NumCols(); colIdx++ {
		newCol := newRecord.Column(int(colIdx))
		field := builder.Field(int(colIdx))

		for rowIdx := int64(0); rowIdx < newRecord.NumRows(); rowIdx++ {
			value := dm.getValueFromArray(newCol, int(rowIdx))
			err := dm.appendValue(field, value)
			if err != nil {
				return nil, err
			}
		}
	}

	return builder.NewRecord(), nil
}

// 更新记录
func (dm *DataManager) updateRecord(oldRecord arrow.Record, assignments map[string]interface{}, whereCondition func(arrow.Record, int) bool) (arrow.Record, error) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, oldRecord.Schema())
	defer builder.Release()

	// 创建字段名到索引的映射
	fieldMap := make(map[string]int)
	for i, field := range oldRecord.Schema().Fields() {
		fieldMap[field.Name] = i
	}

	// 处理每一行
	for rowIdx := int64(0); rowIdx < oldRecord.NumRows(); rowIdx++ {
		shouldUpdate := whereCondition == nil || whereCondition(oldRecord, int(rowIdx))

		for colIdx := int64(0); colIdx < oldRecord.NumCols(); colIdx++ {
			field := builder.Field(int(colIdx))
			fieldName := oldRecord.Schema().Field(int(colIdx)).Name

			var value interface{}
			if shouldUpdate && assignments[fieldName] != nil {
				value = assignments[fieldName]
			} else {
				oldCol := oldRecord.Column(int(colIdx))
				value = dm.getValueFromArray(oldCol, int(rowIdx))
			}

			err := dm.appendValue(field, value)
			if err != nil {
				return nil, err
			}
		}
	}

	return builder.NewRecord(), nil
}

// 过滤记录
func (dm *DataManager) filterRecord(oldRecord arrow.Record, whereCondition func(arrow.Record, int) bool) (arrow.Record, error) {
	pool := memory.NewGoAllocator()
	builder := array.NewRecordBuilder(pool, oldRecord.Schema())
	defer builder.Release()

	// 处理每一行
	for rowIdx := int64(0); rowIdx < oldRecord.NumRows(); rowIdx++ {
		shouldKeep := whereCondition == nil || !whereCondition(oldRecord, int(rowIdx))

		if shouldKeep {
			for colIdx := int64(0); colIdx < oldRecord.NumCols(); colIdx++ {
				field := builder.Field(int(colIdx))
				oldCol := oldRecord.Column(int(colIdx))
				value := dm.getValueFromArray(oldCol, int(rowIdx))

				err := dm.appendValue(field, value)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return builder.NewRecord(), nil
}

// 从Array中获取值
func (dm *DataManager) getValueFromArray(arr arrow.Array, index int) interface{} {
	switch arr := arr.(type) {
	case *array.Int64:
		return arr.Value(index)
	case *array.Float64:
		return arr.Value(index)
	case *array.String:
		return arr.Value(index)
	case *array.Boolean:
		return arr.Value(index)
	default:
		return nil
	}
}

// 向字段添加值
func (dm *DataManager) appendValue(field array.Builder, value interface{}) error {
	switch field := field.(type) {
	case *array.Int64Builder:
		if v, ok := value.(int64); ok {
			field.Append(v)
		} else if v, ok := value.(int); ok {
			field.Append(int64(v))
		} else {
			field.AppendNull()
		}
	case *array.Float64Builder:
		if v, ok := value.(float64); ok {
			field.Append(v)
		} else {
			field.AppendNull()
		}
	case *array.StringBuilder:
		if v, ok := value.(string); ok {
			field.Append(v)
		} else {
			field.AppendNull()
		}
	case *array.BooleanBuilder:
		if v, ok := value.(bool); ok {
			field.Append(v)
		} else {
			field.AppendNull()
		}
	default:
		return fmt.Errorf("unsupported field type")
	}
	return nil
}

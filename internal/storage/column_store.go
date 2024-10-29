package storage

import (
	"fmt"
	"sync"
)

type Column struct {
	Type string
	Data []string
}

type ColumnStore struct {
	Schema   Row
	Columns  map[string]*Column
	RowCount int
	mu       sync.RWMutex
}

func NewColumnStore(schema Row) *ColumnStore {
	columns := make(map[string]*Column)
	for name, typ := range schema {
		columns[name] = &Column{
			Type: typ,
			Data: make([]string, 0),
		}
	}

	return &ColumnStore{
		Schema:   schema,
		Columns:  columns,
		RowCount: 0,
	}
}

func (cs *ColumnStore) Insert(values []string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if len(values) != len(cs.Schema) {
		return fmt.Errorf("column count mismatch: expected %d, got %d", len(cs.Schema), len(values))
	}

	i := 0
	for col := range cs.Schema {
		cs.Columns[col].Data = append(cs.Columns[col].Data, values[i])
		i++
	}
	cs.RowCount++
	return nil
}

func (cs *ColumnStore) Select(columns []string, where string) ([]Row, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	result := make([]Row, 0)

	// 构建行
	for i := 0; i < cs.RowCount; i++ {
		row := make(Row)
		for _, col := range columns {
			if col == "*" {
				for colName, column := range cs.Columns {
					row[colName] = column.Data[i]
				}
				break
			}
			if column, exists := cs.Columns[col]; exists {
				row[col] = column.Data[i]
			}
		}
		if evaluateWhere(row, where) {
			result = append(result, row)
		}
	}

	return result, nil
}

func (cs *ColumnStore) Update(column string, value string, where string) (int, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if _, exists := cs.Columns[column]; !exists {
		return 0, fmt.Errorf("column %s does not exist", column)
	}

	count := 0
	for i := 0; i < cs.RowCount; i++ {
		row := make(Row)
		for col, c := range cs.Columns {
			row[col] = c.Data[i]
		}

		if evaluateWhere(row, where) {
			cs.Columns[column].Data[i] = value
			count++
		}
	}

	return count, nil
}

func (cs *ColumnStore) Delete(where string) (int, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	toDelete := make([]int, 0)

	// 找出要删除的行
	for i := 0; i < cs.RowCount; i++ {
		row := make(Row)
		for col, c := range cs.Columns {
			row[col] = c.Data[i]
		}

		if evaluateWhere(row, where) {
			toDelete = append(toDelete, i)
		}
	}

	// 删除行
	if len(toDelete) > 0 {
		for col := range cs.Columns {
			newData := make([]string, 0, cs.RowCount-len(toDelete))
			deleteIdx := 0
			for i := 0; i < cs.RowCount; i++ {
				if deleteIdx < len(toDelete) && i == toDelete[deleteIdx] {
					deleteIdx++
					continue
				}
				newData = append(newData, cs.Columns[col].Data[i])
			}
			cs.Columns[col].Data = newData
		}
		cs.RowCount -= len(toDelete)
	}

	return len(toDelete), nil
}

package storage

import (
	"fmt"
	"strings"
	"sync"
)

type Row map[string]string

type Table struct {
	Schema Row
	Rows   []Row
	mu     sync.RWMutex
}

func NewTable(schema Row) *Table {
	return &Table{
		Schema: schema,
		Rows:   make([]Row, 0),
	}
}

func (t *Table) Insert(values []string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(values) != len(t.Schema) {
		return fmt.Errorf("column count mismatch: expected %d, got %d", len(t.Schema), len(values))
	}

	row := make(Row)
	i := 0
	for col := range t.Schema {
		row[col] = values[i]
		i++
	}

	t.Rows = append(t.Rows, row)
	return nil
}

func (t *Table) Select(columns []string, where string) ([]Row, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]Row, 0)
	for _, row := range t.Rows {
		if evaluateWhere(row, where) {
			selectedRow := make(Row)
			for _, col := range columns {
				if col == "*" {
					for k, v := range row {
						selectedRow[k] = v
					}
					break
				}
				if val, exists := row[col]; exists {
					selectedRow[col] = val
				}
			}
			result = append(result, selectedRow)
		}
	}
	return result, nil
}

func (t *Table) Update(column string, value string, where string) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.Schema[column]; !exists {
		return 0, fmt.Errorf("column %s does not exist", column)
	}

	count := 0
	for i := range t.Rows {
		if evaluateWhere(t.Rows[i], where) {
			t.Rows[i][column] = value
			count++
		}
	}
	return count, nil
}

func (t *Table) Delete(where string) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	newRows := make([]Row, 0)
	count := 0

	for _, row := range t.Rows {
		if !evaluateWhere(row, where) {
			newRows = append(newRows, row)
		} else {
			count++
		}
	}

	t.Rows = newRows
	return count, nil
}

func evaluateWhere(row Row, where string) bool {
	if where == "" {
		return true
	}

	// 支持复杂条件
	parts := strings.Split(where, " AND ")
	for _, part := range parts {
		if !evaluateCondition(row, part) {
			return false
		}
	}
	return true
}

func evaluateCondition(row Row, condition string) bool {
	// 支持各种操作符
	for _, op := range []string{">=", "<=", "<>", "=", ">", "<", "LIKE", "IN"} {
		if parts := strings.Split(condition, op); len(parts) == 2 {
			column := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, "'\"")

			actual, exists := row[column]
			if !exists {
				return false
			}

			switch op {
			case "=":
				return actual == value
			case ">":
				return actual > value
			case "<":
				return actual < value
			case ">=":
				return actual >= value
			case "<=":
				return actual <= value
			case "<>":
				return actual != value
			case "LIKE":
				return matchLikePattern(actual, value)
			case "IN":
				return evaluateIn(actual, value)
			}
		}
	}
	return false
}

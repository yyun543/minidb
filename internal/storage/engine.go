package storage

import (
	"fmt"
	"strings"
	"sync"
)

type Row map[string]string

type Table struct {
	Rows []Row
	mu   sync.RWMutex
}

type Engine struct {
	Tables map[string]*Table
	mu     sync.RWMutex
}

func NewEngine() *Engine {
	return &Engine{
		Tables: make(map[string]*Table),
	}
}

func (e *Engine) CreateTable(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.Tables[name]; exists {
		return fmt.Errorf("table %s already exists", name)
	}

	e.Tables[name] = &Table{
		Rows: make([]Row, 0),
	}
	return nil
}

func (e *Engine) Select(table string, fields []string) ([]Row, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	t, exists := e.Tables[table]
	if !exists {
		return nil, fmt.Errorf("table %s does not exist", table)
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	if fields[0] == "*" {
		return t.Rows, nil
	}

	result := make([]Row, len(t.Rows))
	for i, row := range t.Rows {
		result[i] = make(Row)
		for _, field := range fields {
			if value, exists := row[field]; exists {
				result[i][field] = value
			}
		}
	}
	return result, nil
}

func (e *Engine) Insert(table string, values []string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	t, exists := e.Tables[table]
	if !exists {
		return fmt.Errorf("table %s does not exist", table)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	row := make(Row)
	for i, value := range values {
		row[fmt.Sprintf("col%d", i+1)] = strings.TrimSpace(value)
	}
	t.Rows = append(t.Rows, row)
	return nil
}

func (e *Engine) Update(table, field, value, where string) (int, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	t, exists := e.Tables[table]
	if !exists {
		return 0, fmt.Errorf("table %s does not exist", table)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	count := 0
	for i := range t.Rows {
		if evaluateWhere(t.Rows[i], where) {
			t.Rows[i][field] = value
			count++
		}
	}
	return count, nil
}

func (e *Engine) Delete(table, where string) (int, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	t, exists := e.Tables[table]
	if !exists {
		return 0, fmt.Errorf("table %s does not exist", table)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	count := 0
	newRows := make([]Row, 0, len(t.Rows))
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
	// 简化实现，仅支持简单的等于条件
	parts := strings.Split(where, "=")
	if len(parts) != 2 {
		return false
	}
	field := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	return row[field] == value
}


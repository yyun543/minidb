package storage

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Row map[string]string

type Table struct {
	Rows   []Row
	Schema Row // Stores column names and their types
	mu     sync.RWMutex
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

	// Use lowercase table names consistently
	name = strings.ToLower(name)

	if _, exists := e.Tables[name]; exists {
		return fmt.Errorf("table %s already exists", name)
	}

	e.Tables[name] = &Table{
		Rows: make([]Row, 0),
	}
	return nil
}

func (e *Engine) CreateTableWithColumns(name string, columns []string, types []string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Use lowercase table names consistently
	name = strings.ToLower(name)

	if _, exists := e.Tables[name]; exists {
		return fmt.Errorf("table %s already exists", name)
	}

	// Create table schema with lowercase column names
	schema := make(Row)
	for i, col := range columns {
		schema[strings.ToLower(col)] = types[i]
	}

	e.Tables[name] = &Table{
		Rows:   make([]Row, 0),
		Schema: schema,
	}
	return nil
}

func (e *Engine) Select(table string, fields []string) ([]Row, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Use lowercase table names consistently
	table = strings.ToLower(table)

	t, exists := e.Tables[table]
	if !exists {
		return nil, fmt.Errorf("table %s does not exist", table)
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	// Convert field names to lowercase for comparison
	lowerFields := make([]string, len(fields))
	for i, field := range fields {
		lowerFields[i] = strings.ToLower(field)
	}

	// Validate fields against schema
	if lowerFields[0] != "*" {
		for _, field := range lowerFields {
			if _, exists := t.Schema[field]; !exists {
				return nil, fmt.Errorf("column %s does not exist in table %s", field, table)
			}
		}
	}

	result := make([]Row, len(t.Rows))
	for i, row := range t.Rows {
		result[i] = make(Row)
		if lowerFields[0] == "*" {
			// Copy all fields
			for col, val := range row {
				result[i][col] = val
			}
		} else {
			// Copy selected fields
			for _, field := range lowerFields {
				if val, exists := row[field]; exists {
					result[i][field] = val
				}
			}
		}
	}
	return result, nil
}

func (e *Engine) Insert(table string, values []string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Use lowercase table names consistently
	table = strings.ToLower(table)

	t, exists := e.Tables[table]
	if !exists {
		return fmt.Errorf("table %s does not exist", table)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if schema exists
	if t.Schema == nil {
		return fmt.Errorf("table %s has no schema defined", table)
	}

	// Get ordered column names from schema
	columns := make([]string, 0, len(t.Schema))
	for col := range t.Schema {
		columns = append(columns, col)
	}
	sort.Strings(columns) // Sort for consistent order

	// Validate number of values matches schema
	if len(values) != len(columns) {
		return fmt.Errorf("invalid number of values: expected %d, got %d", len(columns), len(values))
	}

	// Create new row using schema column names
	row := make(Row)
	for i, col := range columns {
		// 确保使用小写列名存储
		row[col] = strings.TrimSpace(values[i])
	}

	t.Rows = append(t.Rows, row)
	return nil
}

func (e *Engine) Update(table, field, value, where string) (int, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	table = strings.ToLower(table)
	field = strings.ToLower(strings.TrimSpace(field))
	value = strings.TrimSpace(value)

	t, exists := e.Tables[table]
	if !exists {
		return 0, fmt.Errorf("table %s does not exist", table)
	}

	if _, exists := t.Schema[field]; !exists {
		return 0, fmt.Errorf("column %s does not exist in table %s", field, table)
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

	// Use lowercase table names consistently
	table = strings.ToLower(table)

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
	if where == "" {
		return true
	}

	parts := strings.Split(where, "=")
	if len(parts) != 2 {
		return false
	}

	field := strings.ToLower(strings.TrimSpace(parts[0]))
	expectedValue := strings.TrimSpace(parts[1])
	// 移除可能存在的引号
	expectedValue = strings.Trim(expectedValue, "'\"")

	actualValue, exists := row[field]
	if !exists {
		return false
	}

	// 移除实际值中可能存在的引号
	actualValue = strings.Trim(actualValue, "'\"")

	// 首先尝试数值比较
	expectedNum, expectedErr := strconv.ParseFloat(expectedValue, 64)
	actualNum, actualErr := strconv.ParseFloat(actualValue, 64)

	if expectedErr == nil && actualErr == nil {
		return expectedNum == actualNum
	}

	// 如果不是数值，进行字符串比较
	return actualValue == expectedValue
}

func (e *Engine) ShowTables() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	tables := make([]string, 0, len(e.Tables))
	for name := range e.Tables {
		tables = append(tables, name)
	}
	return tables
}

func (e *Engine) DropTable(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Use lowercase table names consistently
	name = strings.ToLower(name)

	if _, exists := e.Tables[name]; !exists {
		return fmt.Errorf("table %s does not exist", name)
	}

	delete(e.Tables, name)
	return nil
}

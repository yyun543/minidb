package storage

import (
	"fmt"
	"sort"
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

	// Create table schema
	schema := make(Row)
	for i, col := range columns {
		schema[col] = types[i]
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

	// Validate fields against schema
	if fields[0] != "*" {
		for _, field := range fields {
			if _, exists := t.Schema[field]; !exists {
				return nil, fmt.Errorf("column %s does not exist in table %s", field, table)
			}
		}
	}

	result := make([]Row, len(t.Rows))
	for i, row := range t.Rows {
		result[i] = make(Row)
		if fields[0] == "*" {
			// Copy all fields
			for col := range t.Schema {
				result[i][col] = row[col]
			}
		} else {
			// Copy selected fields
			for _, field := range fields {
				result[i][field] = row[field]
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
		row[col] = strings.TrimSpace(values[i])
	}

	t.Rows = append(t.Rows, row)
	return nil
}

func (e *Engine) Update(table, field, value, where string) (int, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Use lowercase table names consistently
	table = strings.ToLower(table)
	field = strings.TrimSpace(field)

	t, exists := e.Tables[table]
	if !exists {
		return 0, fmt.Errorf("table %s does not exist", table)
	}

	// Validate field exists in schema
	if _, exists := t.Schema[field]; !exists {
		return 0, fmt.Errorf("column %s does not exist in table %s", field, table)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	count := 0
	for i := range t.Rows {
		if evaluateWhere(t.Rows[i], where) {
			t.Rows[i][field] = strings.TrimSpace(value)
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

	field := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	// Remove possible quotes
	value = strings.Trim(value, "'\"")

	actualValue, exists := row[field]
	if !exists {
		return false
	}

	// Case-insensitive comparison
	return strings.EqualFold(strings.TrimSpace(actualValue), strings.TrimSpace(value))
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

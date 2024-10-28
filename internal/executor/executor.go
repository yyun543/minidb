package executor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/storage"
)

type Executor struct {
	storage *storage.Engine
}

func NewExecutor(storage *storage.Engine) *Executor {
	return &Executor{storage: storage}
}

func (e *Executor) Execute(query *parser.Query) (string, error) {
	switch query.Type {
	// DDL Operations
	case parser.CREATE:
		return e.executeCreate(query)
	case parser.DROP:
		return e.executeDropTable(query)
	case parser.SHOW:
		return e.executeShow(query)
	// DML Operations
	case parser.SELECT:
		return e.executeSelect(query)
	case parser.INSERT:
		return e.executeInsert(query)
	case parser.UPDATE:
		return e.executeUpdate(query)
	case parser.DELETE:
		return e.executeDelete(query)
	default:
		return "", fmt.Errorf("unsupported query type")
	}
}

// formatTable formats field names and data rows into an ASCII table
func formatTable(headers []string, rows []storage.Row) string {
	if len(rows) == 0 {
		return "Empty set (0 rows)"
	}

	// If SELECT *, get all column names from first row and sort them
	if len(headers) == 1 && headers[0] == "*" {
		headers = make([]string, 0)
		for key := range rows[0] {
			headers = append(headers, key)
		}
		// Optional: sort column names for consistent display order
		sort.Strings(headers)
	}

	// Calculate maximum width for each column
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
		for _, row := range rows {
			// Use header instead of index i to access map
			width := len(fmt.Sprintf("%v", row[header]))
			if width > colWidths[i] {
				colWidths[i] = width
			}
		}
	}

	var result strings.Builder

	// Helper function to draw separator line
	drawLine := func() {
		result.WriteString("+")
		for _, width := range colWidths {
			result.WriteString(strings.Repeat("-", width+2))
			result.WriteString("+")
		}
		result.WriteString("\n")
	}

	// Write headers
	drawLine()
	result.WriteString("|")
	for i, header := range headers {
		result.WriteString(fmt.Sprintf(" %-*s |", colWidths[i], header))
	}
	result.WriteString("\n")
	drawLine()

	// Write data rows
	for _, row := range rows {
		result.WriteString("|")
		for i, header := range headers {
			// Use header instead of index i to access map
			result.WriteString(fmt.Sprintf(" %-*v |", colWidths[i], row[header]))
		}
		result.WriteString("\n")
	}
	drawLine()

	// Add statistics
	result.WriteString(fmt.Sprintf("\nTotal: %d rows", len(rows)))

	return result.String()
}

func (e *Executor) executeSelect(query *parser.Query) (string, error) {
	rows, err := e.storage.Select(query.Table, query.Fields)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "No rows found", nil
	}
	return formatTable(query.Fields, rows), nil
}

func (e *Executor) executeInsert(query *parser.Query) (string, error) {
	err := e.storage.Insert(query.Table, query.Values)
	if err != nil {
		return "", err
	}
	return "Inserted 1 row", nil
}

func (e *Executor) executeUpdate(query *parser.Query) (string, error) {
	count, err := e.storage.Update(query.Table, query.Fields[0], query.Values[0], query.Where)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Updated %d rows", count), nil
}

func (e *Executor) executeDelete(query *parser.Query) (string, error) {
	count, err := e.storage.Delete(query.Table, query.Where)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Deleted %d rows", count), nil
}

func (e *Executor) executeCreate(query *parser.Query) (string, error) {
	err := e.storage.CreateTableWithColumns(query.Table, query.Fields, query.Values)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Table %s created successfully", query.Table), nil
}

func (e *Executor) executeDropTable(query *parser.Query) (string, error) {
	err := e.storage.DropTable(query.Table)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Table %s dropped successfully", query.Table), nil
}

func (e *Executor) executeShow(query *parser.Query) (string, error) {
	if query.Command == "TABLES" {
		tables := e.storage.ShowTables()
		if len(tables) == 0 {
			return "No tables found", nil
		}

		// Format the output as a table
		var result strings.Builder
		result.WriteString("+----------------+\n")
		result.WriteString("| Table Name     |\n")
		result.WriteString("+----------------+\n")

		for _, table := range tables {
			result.WriteString(fmt.Sprintf("| %-14s |\n", table))
		}

		result.WriteString("+----------------+\n")
		result.WriteString(fmt.Sprintf("\nTotal: %d tables", len(tables)))
		return result.String(), nil
	}
	return "", fmt.Errorf("unsupported SHOW command")
}

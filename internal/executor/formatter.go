package executor

import (
	"fmt"
	"strings"

	"github.com/yyun543/minidb/internal/storage"
)

func formatTable(headers []string, rows []storage.Row) string {
	if len(rows) == 0 {
		return "Empty set"
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
		for _, row := range rows {
			if width := len(fmt.Sprintf("%v", row[header])); width > colWidths[i] {
				colWidths[i] = width
			}
		}
	}

	var result strings.Builder

	// Draw line
	drawLine := func() {
		result.WriteString("+")
		for _, width := range colWidths {
			result.WriteString(strings.Repeat("-", width+2))
			result.WriteString("+")
		}
		result.WriteString("\n")
	}

	// Draw header
	drawLine()
	result.WriteString("|")
	for i, header := range headers {
		result.WriteString(fmt.Sprintf(" %-*s |", colWidths[i], header))
	}
	result.WriteString("\n")
	drawLine()

	// Draw rows
	for _, row := range rows {
		result.WriteString("|")
		for i, header := range headers {
			result.WriteString(fmt.Sprintf(" %-*v |", colWidths[i], row[header]))
		}
		result.WriteString("\n")
	}
	drawLine()

	result.WriteString(fmt.Sprintf("\n%d rows in set", len(rows)))
	return result.String()
}

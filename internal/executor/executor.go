package executor

import (
	"fmt"
	"strings"
	"sort"

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

// formatTable 将字段名和数据行格式化为 ASCII 表格
func formatTable(headers []string, rows []storage.Row) string {
	if len(rows) == 0 {
		return "Empty set (0 rows)"
	}

	// 如果是 SELECT *，需要从第一行数据中获取所有列名并排序
	if len(headers) == 1 && headers[0] == "*" {
		headers = make([]string, 0)
		for key := range rows[0] {
			headers = append(headers, key)
		}
		// 可选：对列名进行排序以保持一致的显示顺序
		sort.Strings(headers)
	}

	// 计算每列的最大宽度
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
		for _, row := range rows {
				// 使用 header 而不是索引 i 来访问 map
			width := len(fmt.Sprintf("%v", row[header]))
			if width > colWidths[i] {
				colWidths[i] = width
			}
		}
	}

	var result strings.Builder

	// 绘制分隔线的辅助函数
	drawLine := func() {
		result.WriteString("+")
		for _, width := range colWidths {
			result.WriteString(strings.Repeat("-", width+2))
			result.WriteString("+")
		}
		result.WriteString("\n")
	}

	// 写入表头
	drawLine()
	result.WriteString("|")
	for i, header := range headers {
		result.WriteString(fmt.Sprintf(" %-*s |", colWidths[i], header))
	}
	result.WriteString("\n")
	drawLine()

	// 写入数据行
	for _, row := range rows {
		result.WriteString("|")
		for i, header := range headers {
			// 使用 header 而不是索引 i 来访问 map
			result.WriteString(fmt.Sprintf(" %-*v |", colWidths[i], row[header]))
		}
		result.WriteString("\n")
	}
	drawLine()

	// 添加统计信息
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

package main

import (
	"fmt"
	"strings"

	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
)

// QueryHandler 处理SQL查询请求
type QueryHandler struct {
	catalog  *catalog.Catalog
	executor *executor.ExecutorImpl
}

// NewQueryHandler 创建新的查询处理器
func NewQueryHandler() *QueryHandler {
	cat := catalog.NewCatalog()
	exec := executor.NewExecutor(cat)
	return &QueryHandler{
		catalog:  cat,
		executor: exec,
	}
}

// HandleQuery 处理单个SQL查询
func (h *QueryHandler) HandleQuery(sql string) (string, error) {
	// 1. 解析SQL
	ast, err := parser.Parse(sql)
	if err != nil {
		return "", fmt.Errorf("解析错误: %v", err)
	}

	// 2. 优化查询
	opt := optimizer.NewOptimizer()
	plan, err := opt.Optimize(ast)
	if err != nil {
		return "", fmt.Errorf("优化错误: %v", err)
	}

	// 3. 执行查询
	result, err := h.executor.Execute(plan)
	if err != nil {
		return "", fmt.Errorf("执行错误: %v", err)
	}

	// 4. 格式化结果
	return formatResult(result), nil
}

// formatResult 将查询结果格式化为字符串
func formatResult(result *executor.ResultSet) string {
	if result == nil {
		return "OK"
	}

	var sb strings.Builder

	// 写入列名
	headers := result.Headers()
	sb.WriteString("|")
	for _, header := range headers {
		sb.WriteString(fmt.Sprintf(" %-15s |", header))
	}
	sb.WriteString("\n")

	// 写入分隔线
	sb.WriteString("+")
	for range headers {
		sb.WriteString("-----------------+")
	}
	sb.WriteString("\n")

	// 写入数据行
	for result.Next() {
		row := result.Row()
		sb.WriteString("|")
		for _, value := range row {
			sb.WriteString(fmt.Sprintf(" %-15v |", value))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

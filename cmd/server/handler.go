package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/yyun543/minidb/internal/catalog"
	"github.com/yyun543/minidb/internal/executor"
	"github.com/yyun543/minidb/internal/optimizer"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/session"
)

// QueryHandler 处理SQL查询请求
type QueryHandler struct {
	catalog        *catalog.Catalog
	executor       *executor.BaseExecutor
	sessionManager *session.SessionManager
}

// NewQueryHandler 创建新的查询处理器
func NewQueryHandler() (*QueryHandler, error) {
	cat, err := catalog.NewCatalog()
	if err != nil {
		return nil, fmt.Errorf("Failed to create a catalog: %v", err)
	}
	exec := executor.NewExecutor(cat)

	sessMgr, err := session.NewSessionManager()
	if err != nil {
		return nil, fmt.Errorf("Failure to create session manager: %v", err)
	}

	// 启动定期清理过期会话的goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			sessMgr.CleanupExpiredSessions(2 * time.Hour)
		}
	}()

	return &QueryHandler{
		catalog:        cat,
		executor:       exec,
		sessionManager: sessMgr,
	}, nil
}

// HandleQuery 处理单个SQL查询
func (h *QueryHandler) HandleQuery(sessionID int64, sql string) (string, error) {
	// 获取或创建会话
	sess, ok := h.sessionManager.GetSession(sessionID)
	if !ok {
		return "", fmt.Errorf("Invalid session ID: %d", sessionID)
	}

	// 1. 解析SQL
	ast, err := parser.Parse(sql)
	if err != nil {
		return "", fmt.Errorf("parsing error: %v", err)
	}

	// 2. 处理USE语句，更新会话的当前数据库
	if useStmt, ok := ast.(*parser.UseStmt); ok {
		sess.CurrentDB = useStmt.Database
		return fmt.Sprintf("Switched to database: %s", useStmt.Database), nil
	}

	// 3. 优化查询
	opt := optimizer.NewOptimizer()
	plan, err := opt.Optimize(ast)
	if err != nil {
		return "", fmt.Errorf("optimization error: %v", err)
	}

	// 4. 执行查询
	result, err := h.executor.Execute(plan, sess)
	if err != nil {
		return "", fmt.Errorf("execution error: %v", err)
	}

	// 5. 格式化结果
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

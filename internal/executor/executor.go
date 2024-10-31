package executor

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/yyun543/minidb/internal/cache"
	"github.com/yyun543/minidb/internal/index"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/storage"
)

// Executor 是SQL执行引擎的主要结构体
type Executor struct {
	storage *storage.Engine // 存储引擎
	cache   *cache.Cache    // 查询缓存
	index   *index.Manager  // 索引管理器
	stats   ExecutionStats  // 执行统计信息
}

// New 创建一个新的执行器实例
func New(storage *storage.Engine, index *index.Manager, cache *cache.Cache) *Executor {
	return &Executor{
		storage: storage,
		cache:   cache,
		index:   index,
	}
}

// ExecutorError 是执行器错误类型
type ExecutorError struct {
	Operation string
	Err       error
	SQL       string
	Timestamp time.Time
}

func (e *ExecutorError) Error() string {
	return fmt.Sprintf("[%s] %s failed: %v (SQL: %s)",
		e.Timestamp.Format(time.RFC3339),
		e.Operation,
		e.Err,
		e.SQL)
}

// ExecutionStats 是执行统计信息类型
type ExecutionStats struct {
	ParseTime   time.Duration
	ExecuteTime time.Duration
	CacheHits   int64
	CacheMisses int64
	ErrorCount  int64
}

// GetStats 返回执行统计信息
func (e *Executor) GetStats() ExecutionStats {
	return e.stats
}

// Execute 执行一条SQL语句并返回结果
func (e *Executor) Execute(sql string) (string, error) {
	start := time.Now()
	log.Printf("Executing SQL: %s", sql)

	// 1. 解析SQL语句
	stmt, err := e.parseSQL(sql)
	if err != nil {
		log.Printf("Parse error: %v", err)
		return "", &ExecutorError{
			Operation: "Parse",
			Err:       err,
			SQL:       sql,
			Timestamp: start,
		}
	}

	// 2. 检查缓存
	if result, ok := e.checkCache(sql); ok {
		log.Printf("Cache hit for SQL: %s", sql)
		return result, nil
	}

	// 3. 执行语句
	result, err := e.executeStatement(stmt)
	if err != nil {
		log.Printf("Execution error: %v", err)
		return "", &ExecutorError{
			Operation: "Execute",
			Err:       err,
			SQL:       sql,
			Timestamp: start,
		}
	}

	// 4. 缓存结果
	if _, ok := stmt.(*parser.SelectStmt); ok {
		e.cache.Set(sql, result)
		log.Printf("Cached result for SQL: %s", sql)
	}

	duration := time.Since(start)
	log.Printf("Executed SQL in %v: %s", duration, sql)
	return result, nil
}

// parseSQL 解析SQL语句
func (e *Executor) parseSQL(sql string) (parser.Statement, error) {
	p := parser.NewParser(sql)
	return p.Parse()
}

// checkCache 检查查询缓存
func (e *Executor) checkCache(sql string) (string, bool) {
	if value, ok := e.cache.Get(sql); ok {
		if result, ok := value.(string); ok {
			return result, true
		}
	}
	return "", false
}

// executeStatement 执行具体的SQL语句
func (e *Executor) executeStatement(stmt parser.Statement) (string, error) {
	switch s := stmt.(type) {
	case *parser.CreateTableStmt:
		return e.executeCreate(s)
	case *parser.SelectStmt:
		return e.executeSelect(s)
	case *parser.InsertStmt:
		return e.executeInsert(s)
	case *parser.UpdateStmt:
		return e.executeUpdate(s)
	case *parser.DeleteStmt:
		return e.executeDelete(s)
	default:
		return "", fmt.Errorf("unsupported statement type: %T", stmt)
	}
}

// executeCreate 执行CREATE TABLE语句
func (e *Executor) executeCreate(stmt *parser.CreateTableStmt) (string, error) {
	// 转换列定义为存储引擎需要的格式
	schema := make(storage.Schema)
	for _, col := range stmt.Columns {
		schema[col.Name] = storage.Column{
			Type:     col.DataType,
			Nullable: !col.NotNull,
		}
	}

	err := e.storage.CreateTable(stmt.TableName, schema)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Table %s created successfully", stmt.TableName), nil
}

// executeSelect 执行SELECT语句
func (e *Executor) executeSelect(stmt *parser.SelectStmt) (string, error) {
	// 转换字段列表
	columns := make([]string, len(stmt.Fields))
	for i, field := range stmt.Fields {
		columns[i] = field.String()
	}

	// 转换WHERE条件
	where := ""
	if stmt.Where != nil {
		where = stmt.Where.String()
	}

	rows, err := e.storage.Select(stmt.Table, columns, where)
	if err != nil {
		return "", err
	}

	return formatResults(rows), nil
}

// executeInsert 执行INSERT语句
func (e *Executor) executeInsert(stmt *parser.InsertStmt) (string, error) {
	// 转换值为map
	values := make(map[string]parser.Expression)
	for i, col := range stmt.Columns {
		values[col] = stmt.Values[i].String()
	}

	err := e.storage.Insert(stmt.Table, values)
	if err != nil {
		return "", err
	}

	return "Insert successful", nil
}

// executeUpdate 执行UPDATE语句
func (e *Executor) executeUpdate(stmt *parser.UpdateStmt) (string, error) {
	// 转换更新值为map[string]interface{}
	values := make(map[string]interface{})
	for key, expr := range stmt.Values {
		value, err := extractValue(expr)
		if err != nil {
			return "", fmt.Errorf("failed to extract value: %v", err)
		}
		values[key] = value
	}

	// 转换WHERE条件
	where := ""
	if stmt.Where != nil {
		where = stmt.Where.String()
	}

	// 执行更新
	count, err := e.storage.Update(stmt.Table, values, where)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d rows updated", count), nil
}

// extractValue 从表达式中提取值
func extractValue(expr parser.Expression) (interface{}, error) {
	switch e := expr.(type) {
	case *parser.Literal:
		return convertLiteralValue(e)
	case *parser.Identifier:
		return e.Name, nil
	case *parser.ValueExpression:
		return e.Value(), nil
	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// convertLiteralValue 转换字面量值
func convertLiteralValue(lit *parser.Literal) (interface{}, error) {
	switch lit.Type {
	case "string":
		return lit.Value, nil
	case "number":
		// 尝试转换为整数
		if i, err := strconv.ParseInt(lit.Value, 10, 64); err == nil {
			return i, nil
		}
		// 尝试转换为浮点数
		if f, err := strconv.ParseFloat(lit.Value, 64); err == nil {
			return f, nil
		}
		return nil, fmt.Errorf("invalid number format: %s", lit.Value)
	case "boolean":
		return strconv.ParseBool(lit.Value)
	default:
		return lit.Value, nil
	}
}

// executeDelete 执行DELETE语句
func (e *Executor) executeDelete(stmt *parser.DeleteStmt) (string, error) {
	// 将 Expression 类型的 Where 转换为字符串
	whereStr := ""
	if stmt.Where != nil {
		whereStr = stmt.Where.String()
	}

	count, err := e.storage.Delete(stmt.Table, whereStr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d rows deleted", count), nil
}

// findUsableIndex 查找可用的索引
func (e *Executor) findUsableIndex(stmt *parser.SelectStmt) *index.Index {
	// 简化的索引选择逻辑
	if stmt.Where == nil {
		return nil
	}
	return e.index.FindBestIndex(stmt.Table, stmt.Where)
}

// executeIndexedSelect 使用索引执行SELECT
func (e *Executor) executeIndexedSelect(stmt *parser.SelectStmt, idx *index.Index) (string, error) {
	// 从WHERE子句中提取值
	value := extractValueFromWhere(stmt.Where)

	// 使用索引查找记录
	rowIDs := idx.Find(value)

	// 使用行ID获取完整记录
	var result []storage.Row
	for _, id := range rowIDs {
		row, err := e.storage.GetRow(stmt.Table, id)
		if err != nil {
			continue
		}
		result = append(result, row)
	}

	return formatResults(result), nil
}

// 新增辅助函数
func extractValueFromWhere(expr parser.Expression) (interface{}, error) {
	switch e := expr.(type) {
	case *parser.ValueExpression:
		return e.Value(), nil
	case *parser.Literal:
		return e.Value, nil
	case *parser.Identifier:
		return e.Name, nil
	default:
		return nil, fmt.Errorf("unsupported expression type for value extraction: %T", expr)
	}
}

// updateIndexes 更新表的索引
func (e *Executor) updateIndexes(table string, values []string) {
	indexes := e.index.GetTableIndexes(table)
	for _, idx := range indexes {
		err := idx.Update(values)
		if err != nil {
			return
		}
	}
}

// formatResults 格式化查询结果
func formatResults(rows []storage.Row) string {
	// 简单的表格格式化
	if len(rows) == 0 {
		return "Empty set"
	}
	// ... 格式化逻辑 ...
	return "Results formatted as table"
}

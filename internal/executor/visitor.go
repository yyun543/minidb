package executor

import (
	"fmt"
	"strings"

	"github.com/yyun543/minidb/internal/cache"
	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/storage"
)

type QueryExecutor struct {
	storage *storage.Engine
	cache   *cache.Cache
}

func NewQueryExecutor(storage *storage.Engine, cache *cache.Cache) *QueryExecutor {
	return &QueryExecutor{
		storage: storage,
		cache:   cache,
	}
}

func (e *QueryExecutor) VisitCreateTable(stmt *parser.CreateTableStmt) interface{} {
	schema := make(storage.Row)
	for _, col := range stmt.Columns {
		schema[col.Name] = col.DataType
	}

	err := e.storage.CreateTable(stmt.TableName, schema)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}
	return fmt.Sprintf("Table %s created successfully", stmt.TableName)
}

func (e *QueryExecutor) VisitDropTable(stmt *parser.DropTableStmt) interface{} {
	err := e.storage.DropTable(stmt.TableName)
	if err != nil {
		return fmt.Errorf("failed to drop table: %v", err)
	}
	return fmt.Sprintf("Table %s dropped successfully", stmt.TableName)
}

func (e *QueryExecutor) VisitShowTables(stmt *parser.ShowTablesStmt) interface{} {
	tables := e.storage.GetTables()
	if len(tables) == 0 {
		return "No tables"
	}
	return fmt.Sprintf("Tables:\n%s", strings.Join(tables, "\n"))
}

func (e *QueryExecutor) VisitSelect(stmt *parser.SelectStmt) interface{} {
	// 检查缓存
	cacheKey := stmt.String()
	if result, ok := e.cache.Get(cacheKey); ok {
		return result
	}

	// 获取字段名
	fields := make([]string, len(stmt.Fields))
	for i, field := range stmt.Fields {
		if ident, ok := field.(*parser.Identifier); ok {
			fields[i] = ident.Name
		}
	}

	// 构建WHERE子句
	where := ""
	if stmt.Where != nil {
		where = e.buildWhereClause(stmt.Where)
	}

	// 执行查询
	rows, err := e.storage.Select(stmt.From, fields, where, stmt.IsAnalytic)
	if err != nil {
		return fmt.Errorf("select failed: %v", err)
	}

	// 处理JOIN
	if stmt.JoinType != parser.NO_JOIN {
		rows, err = e.processJoin(rows, stmt)
		if err != nil {
			return fmt.Errorf("join failed: %v", err)
		}
	}

	// 处理GROUP BY
	if len(stmt.GroupBy) > 0 {
		rows, err = e.processGroupBy(rows, stmt)
		if err != nil {
			return fmt.Errorf("group by failed: %v", err)
		}
	}

	// 处理ORDER BY
	if len(stmt.OrderBy) > 0 {
		rows, err = e.processOrderBy(rows, stmt)
		if err != nil {
			return fmt.Errorf("order by failed: %v", err)
		}
	}

	// 处理LIMIT和OFFSET
	if stmt.Limit != nil {
		limit := *stmt.Limit
		offset := 0
		if stmt.Offset != nil {
			offset = *stmt.Offset
		}
		if offset < len(rows) {
			end := offset + limit
			if end > len(rows) {
				end = len(rows)
			}
			rows = rows[offset:end]
		} else {
			rows = []storage.Row{}
		}
	}

	result := formatTable(fields, rows)

	// 缓存结果
	e.cache.Set(cacheKey, result)

	return result
}

func (e *QueryExecutor) VisitInsert(stmt *parser.InsertStmt) interface{} {
	values := make([]string, len(stmt.Values))
	for i, val := range stmt.Values {
		if lit, ok := val.(*parser.Literal); ok {
			values[i] = lit.Value
		}
	}

	err := e.storage.Insert(stmt.Table, values)
	if err != nil {
		return fmt.Errorf("insert failed: %v", err)
	}
	return "1 row inserted"
}

func (e *QueryExecutor) VisitUpdate(stmt *parser.UpdateStmt) interface{} {
	// 构建SET和WHERE子句
	for col, val := range stmt.Set {
		if lit, ok := val.(*parser.Literal); ok {
			where := ""
			if stmt.Where != nil {
				where = e.buildWhereClause(stmt.Where)
			}
			count, err := e.storage.Update(stmt.Table, col, lit.Value, where)
			if err != nil {
				return fmt.Errorf("update failed: %v", err)
			}
			return fmt.Sprintf("%d rows updated", count)
		}
	}
	return fmt.Errorf("invalid update statement")
}

func (e *QueryExecutor) VisitDelete(stmt *parser.DeleteStmt) interface{} {
	where := ""
	if stmt.Where != nil {
		where = e.buildWhereClause(stmt.Where)
	}

	count, err := e.storage.Delete(stmt.Table, where)
	if err != nil {
		return fmt.Errorf("delete failed: %v", err)
	}
	return fmt.Sprintf("%d rows deleted", count)
}

func (e *QueryExecutor) VisitIdentifier(node *parser.Identifier) interface{} {
	return node.Name
}

func (e *QueryExecutor) VisitLiteral(node *parser.Literal) interface{} {
	return node.Value
}

func (e *QueryExecutor) VisitComparison(node *parser.ComparisonExpr) interface{} {
	left := e.visitNode(node.Left)
	right := e.visitNode(node.Right)
	return fmt.Sprintf("%v %s %v", left, node.Operator, right)
}

func (e *QueryExecutor) VisitFunction(node *parser.FunctionExpr) interface{} {
	args := make([]string, len(node.Args))
	for i, arg := range node.Args {
		args[i] = fmt.Sprintf("%v", e.visitNode(arg))
	}
	return fmt.Sprintf("%s(%s)", node.Name, strings.Join(args, ", "))
}

func (e *QueryExecutor) visitNode(node parser.Node) interface{} {
	switch n := node.(type) {
	case *parser.Identifier:
		return e.VisitIdentifier(n)
	case *parser.Literal:
		return e.VisitLiteral(n)
	case *parser.ComparisonExpr:
		return e.VisitComparison(n)
	case *parser.FunctionExpr:
		return e.VisitFunction(n)
	default:
		return fmt.Errorf("unknown node type: %T", node)
	}
}

func (e *QueryExecutor) buildWhereClause(expr parser.Expression) string {
	if expr == nil {
		return ""
	}
	return fmt.Sprintf("%v", e.visitNode(expr))
}

func (e *QueryExecutor) processJoin(rows []storage.Row, stmt *parser.SelectStmt) ([]storage.Row, error) {
	if stmt.JoinTable == "" {
		return rows, nil
	}

	rightRows, err := e.storage.Select(stmt.JoinTable, []string{"*"}, "", false)
	if err != nil {
		return nil, err
	}

	result := make([]storage.Row, 0)
	for _, leftRow := range rows {
		for _, rightRow := range rightRows {
			if e.evaluateJoinCondition(leftRow, rightRow, stmt.JoinOn) {
				mergedRow := make(storage.Row)
				// 合并左右表行
				for k, v := range leftRow {
					mergedRow[k] = v
				}
				for k, v := range rightRow {
					mergedRow[stmt.JoinTable+"."+k] = v
				}
				result = append(result, mergedRow)
			}
		}
	}
	return result, nil
}

func (e *QueryExecutor) processGroupBy(rows []storage.Row, stmt *parser.SelectStmt) ([]storage.Row, error) {
	if len(stmt.GroupBy) == 0 {
		return rows, nil
	}

	// 按GROUP BY字段分组
	groups := make(map[string][]storage.Row)
	for _, row := range rows {
		key := makeGroupKey(row, stmt.GroupBy)
		groups[key] = append(groups[key], row)
	}

	// 处理每个分组
	result := make([]storage.Row, 0)
	for _, group := range groups {
		aggregatedRow := aggregateGroup(group, stmt.Fields)
		result = append(result, aggregatedRow)
	}

	return result, nil
}

func makeGroupKey(row storage.Row, groupBy []string) string {
	parts := make([]string, len(groupBy))
	for i, col := range groupBy {
		parts[i] = row[col]
	}
	return strings.Join(parts, "|")
}

func aggregateGroup(group []storage.Row, fields []parser.Expression) storage.Row {
	result := make(storage.Row)
	// ... 实现聚合函数
	return result
}

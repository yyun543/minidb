package executor

import (
	"fmt"

	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/storage"
)

// QueryVisitor 实现SQL语句的访问者模式
type QueryVisitor struct {
	storage *storage.Engine
}

// NewQueryVisitor 创建新的查询访问者
func NewQueryVisitor(storage *storage.Engine) *QueryVisitor {
	return &QueryVisitor{storage: storage}
}

// VisitCreateTable 处理CREATE TABLE语句
func (v *QueryVisitor) VisitCreateTable(stmt *parser.CreateTableStmt) interface{} {
	// 构建表结构
	schema := make(storage.Schema)
	for _, col := range stmt.Columns {
		schema[col.Name] = storage.Column{
			Type:     col.DataType,
			Nullable: !col.NotNull,
		}
	}

	// 创建表
	err := v.storage.CreateTable(stmt.TableName, schema)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	return fmt.Sprintf("Table %s created successfully", stmt.TableName)
}

// VisitSelect 处理SELECT语句
func (v *QueryVisitor) VisitSelect(stmt *parser.SelectStmt) interface{} {
	// 处理查询字段
	fields := v.processSelectFields(stmt.Fields)

	// 处理WHERE条件
	where := ""
	if stmt.Where != nil {
		where = v.processWhereClause(stmt.Where)
	}

	// 执行查询
	rows, err := v.storage.Select(stmt.Table, fields, where)
	if err != nil {
		return fmt.Errorf("select failed: %v", err)
	}

	// 处理ORDER BY
	if len(stmt.OrderBy) > 0 {
		rows = v.processOrderBy(rows, stmt.OrderBy)
	}

	// 处理LIMIT
	if stmt.Limit != nil {
		rows = v.processLimit(rows, *stmt.Limit)
	}

	return formatResults(rows)
}

// VisitInsert 处理INSERT语句
func (v *QueryVisitor) VisitInsert(stmt *parser.InsertStmt) interface{} {
	// 验证列和值的数量是否匹配
	if len(stmt.Columns) != len(stmt.Values) {
		return fmt.Errorf("column count doesn't match value count")
	}

	// 构建要插入的数据
	data := make(map[string]string)
	for i, col := range stmt.Columns {
		data[col] = stmt.Values[i].String()
	}

	// 执行插入
	err := v.storage.Insert(stmt.Table, data)
	if err != nil {
		return fmt.Errorf("insert failed: %v", err)
	}

	return "Insert successful"
}

// VisitUpdate 处理UPDATE语句
func (v *QueryVisitor) VisitUpdate(stmt *parser.UpdateStmt) interface{} {
	// 处理SET子句
	updates := make(map[string]string)
	for col, expr := range stmt.Set {
		updates[col] = expr.String()
	}

	// 处理WHERE条件
	where := ""
	if stmt.Where != nil {
		where = v.processWhereClause(stmt.Where)
	}

	// 执行更新
	count, err := v.storage.Update(stmt.Table, updates, where)
	if err != nil {
		return fmt.Errorf("update failed: %v", err)
	}

	return fmt.Sprintf("%d rows updated", count)
}

// VisitDelete 处理DELETE语句
func (v *QueryVisitor) VisitDelete(stmt *parser.DeleteStmt) interface{} {
	// 处理WHERE条件
	where := ""
	if stmt.Where != nil {
		where = v.processWhereClause(stmt.Where)
	}

	// 执行删除
	count, err := v.storage.Delete(stmt.Table, where)
	if err != nil {
		return fmt.Errorf("delete failed: %v", err)
	}

	return fmt.Sprintf("%d rows deleted", count)
}

// 辅助方法

// processSelectFields 处理SELECT字段列表
func (v *QueryVisitor) processSelectFields(fields []parser.Expression) []string {
	result := make([]string, len(fields))
	for i, field := range fields {
		result[i] = field.String()
	}
	return result
}

// processWhereClause 处理WHERE子句
func (v *QueryVisitor) processWhereClause(expr parser.Expression) string {
	switch e := expr.(type) {
	case *parser.ComparisonExpr:
		return fmt.Sprintf("%s %s %s",
			e.Left.String(),
			e.Operator,
			e.Right.String())
	case *parser.BinaryExpr:
		return fmt.Sprintf("(%s %s %s)",
			v.processWhereClause(e.Left),
			e.Operator,
			v.processWhereClause(e.Right))
	default:
		return expr.String()
	}
}

// processOrderBy 处理ORDER BY子句
func (v *QueryVisitor) processOrderBy(rows []storage.Row, orderBy []parser.OrderByExpr) []storage.Row {
	// 简化的排序实现
	// 实际实现应该使用sort.Slice并处理多个排序键
	return rows
}

// processLimit 处理LIMIT子句
func (v *QueryVisitor) processLimit(rows []storage.Row, limit int) []storage.Row {
	if limit > len(rows) {
		return rows
	}
	return rows[:limit]
}

// VisitChildren 访问子节点
func (v *QueryVisitor) VisitChildren(node parser.Node) interface{} {
	// 默认的子节点访问实现
	return nil
}

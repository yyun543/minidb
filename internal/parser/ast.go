package parser

import (
	"fmt"
	"strings"
)

// Node 表示AST中的节点接口
type Node interface {
	String() string // 用于调试和显示
}

// Statement 表示SQL语句接口
type Statement interface {
	Node
	statementNode() // 标记方法
}

// Expression 表示表达式接口
type Expression interface {
	Node
	expressionNode() // 标记方法
}

// BaseNode 所有节点的基础结构
type BaseNode struct {
	pos Position // 位置信息
}

// Position 表示源代码中的位置
type Position struct {
	Line   int
	Column int
	Offset int
}

// CreateTableStmt 表示CREATE TABLE语句
type CreateTableStmt struct {
	BaseNode
	TableName string
	Columns   []ColumnDef
}

type ColumnDef struct {
	Name     string
	DataType string
	NotNull  bool
}

func (s *CreateTableStmt) statementNode() {}
func (s *CreateTableStmt) String() string {
	cols := make([]string, len(s.Columns))
	for i, col := range s.Columns {
		constraint := ""
		if col.NotNull {
			constraint = " NOT NULL"
		}
		cols[i] = fmt.Sprintf("%s %s%s", col.Name, col.DataType, constraint)
	}
	return fmt.Sprintf("CREATE TABLE %s (%s)", s.TableName, strings.Join(cols, ", "))
}

// SelectStmt 表示SELECT语句
type SelectStmt struct {
	BaseNode
	Fields  []Expression  // 选择的字段
	Table   string        // FROM子句
	Where   Expression    // WHERE子句
	OrderBy []OrderByExpr // ORDER BY子句
	Limit   *int          // LIMIT子句
}

type OrderByExpr struct {
	Expr      Expression
	Ascending bool
}

func (s *SelectStmt) statementNode() {}
func (s *SelectStmt) String() string {
	parts := []string{fmt.Sprintf("SELECT %s", expressionsToString(s.Fields))}
	parts = append(parts, fmt.Sprintf("FROM %s", s.Table))

	if s.Where != nil {
		parts = append(parts, fmt.Sprintf("WHERE %s", s.Where.String()))
	}

	if len(s.OrderBy) > 0 {
		orders := make([]string, len(s.OrderBy))
		for i, o := range s.OrderBy {
			dir := "ASC"
			if !o.Ascending {
				dir = "DESC"
			}
			orders[i] = fmt.Sprintf("%s %s", o.Expr.String(), dir)
		}
		parts = append(parts, fmt.Sprintf("ORDER BY %s", strings.Join(orders, ", ")))
	}

	if s.Limit != nil {
		parts = append(parts, fmt.Sprintf("LIMIT %d", *s.Limit))
	}

	return strings.Join(parts, " ")
}

// InsertStmt 表示INSERT语句
type InsertStmt struct {
	BaseNode
	Table   string
	Columns []string
	Values  []Expression
}

func (s *InsertStmt) statementNode() {}
func (s *InsertStmt) String() string {
	if len(s.Columns) > 0 {
		return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			s.Table,
			strings.Join(s.Columns, ", "),
			expressionsToString(s.Values))
	}
	return fmt.Sprintf("INSERT INTO %s VALUES (%s)",
		s.Table,
		expressionsToString(s.Values))
}

// UpdateStmt 表示UPDATE语句
type UpdateStmt struct {
	BaseNode
	Table string
	Set   map[string]Expression
	Where Expression
}

func (s *UpdateStmt) statementNode() {}
func (s *UpdateStmt) String() string {
	sets := make([]string, 0, len(s.Set))
	for col, val := range s.Set {
		sets = append(sets, fmt.Sprintf("%s = %s", col, val.String()))
	}
	result := fmt.Sprintf("UPDATE %s SET %s", s.Table, strings.Join(sets, ", "))
	if s.Where != nil {
		result += fmt.Sprintf(" WHERE %s", s.Where.String())
	}
	return result
}

// DeleteStmt 表示DELETE语句
type DeleteStmt struct {
	BaseNode
	Table string
	Where Expression
}

func (s *DeleteStmt) statementNode() {}
func (s *DeleteStmt) String() string {
	result := fmt.Sprintf("DELETE FROM %s", s.Table)
	if s.Where != nil {
		result += fmt.Sprintf(" WHERE %s", s.Where.String())
	}
	return result
}

// 表达式节点

// Identifier 表示标识符
type Identifier struct {
	BaseNode
	Name string
}

func (e *Identifier) expressionNode() {}
func (e *Identifier) String() string  { return e.Name }

// Literal 表示字面量
type Literal struct {
	BaseNode
	Value string
	Type  string // 可以是 "string", "number", "boolean" 等
}

func (e *Literal) expressionNode() {}
func (e *Literal) String() string  { return e.Value }

// BinaryExpr 表示二元表达式
type BinaryExpr struct {
	BaseNode
	Left     Expression
	Operator string
	Right    Expression
}

func (e *BinaryExpr) expressionNode() {}
func (e *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", e.Left.String(), e.Operator, e.Right.String())
}

// ComparisonExpr 表示比较表达式
type ComparisonExpr struct {
	BaseNode
	Left     Expression
	Operator string
	Right    Expression
}

func (e *ComparisonExpr) expressionNode() {}
func (e *ComparisonExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.Left.String(), e.Operator, e.Right.String())
}

// FunctionExpr 表示函数调用
type FunctionExpr struct {
	BaseNode
	Name string
	Args []Expression
}

func (e *FunctionExpr) expressionNode() {}
func (e *FunctionExpr) String() string {
	args := make([]string, len(e.Args))
	for i, arg := range e.Args {
		args[i] = arg.String()
	}
	return fmt.Sprintf("%s(%s)", e.Name, strings.Join(args, ", "))
}

// 辅助函数

// expressionsToString 将表达式列表转换为字符串
func expressionsToString(exprs []Expression) string {
	parts := make([]string, len(exprs))
	for i, expr := range exprs {
		parts[i] = expr.String()
	}
	return strings.Join(parts, ", ")
}

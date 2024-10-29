package parser

import (
	"fmt"
	"strings"
)

// NodeType 定义AST节点类型
type NodeType int

const (
	// 语句类型
	CREATE_TABLE NodeType = iota
	DROP_TABLE
	SHOW_TABLES
	SELECT
	INSERT
	UPDATE
	DELETE

	// 表达式类型
	IDENTIFIER // 标识符(表名、列名等)
	STRING_LIT // 字符串字面量
	NUMBER_LIT // 数字字面量
	COMPARISON // 比较表达式
	AND        // AND表达式
	OR         // OR表达式
	FUNCTION   // 函数调用
)

// JoinType 定义JOIN类型
type JoinType int

const (
	NO_JOIN JoinType = iota
	INNER_JOIN
	LEFT_JOIN
	RIGHT_JOIN
)

// Node 接口定义AST节点的基本行为
type Node interface {
	Type() NodeType
	String() string
}

// Statement 接口定义语句节点
type Statement interface {
	Node
	statementNode()
}

// Expression 接口定义表达式节点
type Expression interface {
	Node
	expressionNode()
}

// BaseNode 提供基础节点实现
type BaseNode struct {
	nodeType NodeType
}

func (n BaseNode) Type() NodeType {
	return n.nodeType
}

// CreateTableStmt 表示CREATE TABLE语句
type CreateTableStmt struct {
	BaseNode
	TableName string
	Columns   []ColumnDef
}

type ColumnDef struct {
	Name        string
	DataType    string
	Constraints []string
}

func (s *CreateTableStmt) statementNode() {}
func (s *CreateTableStmt) String() string {
	cols := make([]string, len(s.Columns))
	for i, col := range s.Columns {
		constraints := ""
		if len(col.Constraints) > 0 {
			constraints = " " + strings.Join(col.Constraints, " ")
		}
		cols[i] = fmt.Sprintf("%s %s%s", col.Name, col.DataType, constraints)
	}
	return fmt.Sprintf("CREATE TABLE %s (%s)", s.TableName, strings.Join(cols, ", "))
}

// DropTableStmt 表示DROP TABLE语句
type DropTableStmt struct {
	BaseNode
	TableName string
}

func (s *DropTableStmt) statementNode() {}
func (s *DropTableStmt) String() string {
	return fmt.Sprintf("DROP TABLE %s", s.TableName)
}

// ShowTablesStmt 表示SHOW TABLES语句
type ShowTablesStmt struct {
	BaseNode
}

func (s *ShowTablesStmt) statementNode() {}
func (s *ShowTablesStmt) String() string {
	return "SHOW TABLES"
}

// SelectStmt 表示SELECT语句
type SelectStmt struct {
	BaseNode
	Fields     []Expression
	From       string
	Where      Expression
	JoinType   JoinType
	JoinTable  string
	JoinOn     Expression
	GroupBy    []string
	Having     Expression
	OrderBy    []OrderByExpr
	Limit      *int
	Offset     *int
	IsAnalytic bool
}

type OrderByExpr struct {
	Expr      Expression
	Ascending bool
}

func (s *SelectStmt) statementNode() {}
func (s *SelectStmt) String() string {
	var parts []string

	// SELECT clause
	fields := make([]string, len(s.Fields))
	for i, f := range s.Fields {
		fields[i] = f.String()
	}
	parts = append(parts, fmt.Sprintf("SELECT %s", strings.Join(fields, ", ")))

	// FROM clause
	parts = append(parts, fmt.Sprintf("FROM %s", s.From))

	// JOIN clause
	if s.JoinType != NO_JOIN {
		switch s.JoinType {
		case INNER_JOIN:
			parts = append(parts, fmt.Sprintf("JOIN %s ON %s", s.JoinTable, s.JoinOn.String()))
		case LEFT_JOIN:
			parts = append(parts, fmt.Sprintf("LEFT JOIN %s ON %s", s.JoinTable, s.JoinOn.String()))
		case RIGHT_JOIN:
			parts = append(parts, fmt.Sprintf("RIGHT JOIN %s ON %s", s.JoinTable, s.JoinOn.String()))
		}
	}

	// WHERE clause
	if s.Where != nil {
		parts = append(parts, fmt.Sprintf("WHERE %s", s.Where.String()))
	}

	// GROUP BY clause
	if len(s.GroupBy) > 0 {
		parts = append(parts, fmt.Sprintf("GROUP BY %s", strings.Join(s.GroupBy, ", ")))
	}

	// HAVING clause
	if s.Having != nil {
		parts = append(parts, fmt.Sprintf("HAVING %s", s.Having.String()))
	}

	// ORDER BY clause
	if len(s.OrderBy) > 0 {
		orderBy := make([]string, len(s.OrderBy))
		for i, o := range s.OrderBy {
			dir := "ASC"
			if !o.Ascending {
				dir = "DESC"
			}
			orderBy[i] = fmt.Sprintf("%s %s", o.Expr.String(), dir)
		}
		parts = append(parts, fmt.Sprintf("ORDER BY %s", strings.Join(orderBy, ", ")))
	}

	// LIMIT and OFFSET
	if s.Limit != nil {
		parts = append(parts, fmt.Sprintf("LIMIT %d", *s.Limit))
		if s.Offset != nil {
			parts = append(parts, fmt.Sprintf("OFFSET %d", *s.Offset))
		}
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
	values := make([]string, len(s.Values))
	for i, v := range s.Values {
		values[i] = v.String()
	}
	if len(s.Columns) > 0 {
		return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			s.Table,
			strings.Join(s.Columns, ", "),
			strings.Join(values, ", "))
	}
	return fmt.Sprintf("INSERT INTO %s VALUES (%s)",
		s.Table,
		strings.Join(values, ", "))
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
}

func (e *Literal) expressionNode() {}
func (e *Literal) String() string  { return e.Value }

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

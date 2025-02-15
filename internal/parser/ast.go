package parser

import "time"

// NodeType 定义AST节点类型
type NodeType int

const (
	// 语句节点类型
	SelectNode NodeType = iota
	InsertNode
	UpdateNode
	DeleteNode
	CreateDatabaseNode
	CreateTableNode
	CreateIndexNode
	DropTableNode
	UseNode
	ShowDatabasesNode
	ShowTablesNode
	ExplainNode

	// 表达式节点类型
	BinaryExprNode     // 二元表达式
	ComparisonExprNode // 比较表达式
	LogicalExprNode    // 逻辑表达式
	ColumnRefNode      // 列引用
	LiteralNode        // 字面量
	FunctionCallNode   // 函数调用
	IdentifierNode     // 标识符
)

// Constraint types
const (
	NotNullConstraint    = "NOT NULL"
	PrimaryKeyConstraint = "PRIMARY KEY"
	UniqueConstraint     = "UNIQUE"
	DefaultConstraint    = "DEFAULT"
)

// Node AST节点接口
type Node interface {
	Type() NodeType
}

// BaseNode 基础节点结构，实现Node接口
type BaseNode struct {
	nodeType NodeType
}

func (n *BaseNode) Type() NodeType {
	return n.nodeType
}

// SelectStmt SELECT语句节点
type SelectStmt struct {
	BaseNode
	Columns []string       // 选择的列
	From    string         // FROM子句表名
	Joins   []*JoinClause  // JOIN子句
	Where   *WhereClause   // WHERE子句
	GroupBy []string       // GROUP BY子句
	Having  Node           // HAVING子句
	OrderBy []*OrderByItem // ORDER BY子句
	Limit   int64          // LIMIT子句
}

// JoinClause JOIN子句节点
type JoinClause struct {
	BaseNode
	JoinType  string // JOIN类型(INNER/LEFT/RIGHT/FULL)
	Table     string // 连接表名
	Condition Node   // ON条件
}

// WhereClause WHERE子句节点
type WhereClause struct {
	BaseNode
	Condition Node // WHERE条件
}

// OrderByItem ORDER BY项节点
type OrderByItem struct {
	BaseNode
	Column    string // 排序列
	Ascending bool   // 升序/降序
}

// InsertStmt INSERT语句节点
type InsertStmt struct {
	BaseNode
	Table   string   // 表名
	Columns []string // 列名列表
	Values  []Node   // 值列表
}

// UpdateStmt UPDATE语句节点
type UpdateStmt struct {
	BaseNode
	Table       string              // 表名
	Assignments []*UpdateAssignment // 赋值列表
	Where       *WhereClause        // WHERE子句
}

// UpdateAssignment UPDATE赋值节点
type UpdateAssignment struct {
	BaseNode
	Column string // 列名
	Value  Node   // 值表达式
}

// DeleteStmt DELETE语句节点
type DeleteStmt struct {
	BaseNode
	Table string       // 表名
	Where *WhereClause // WHERE子句
}

// CreateDatabaseStmt CREATE DATABASE语句节点
type CreateDatabaseStmt struct {
	BaseNode
	Database string // 数据库名
}

// CreateTableStmt CREATE TABLE语句节点
type CreateTableStmt struct {
	BaseNode
	Table       string        // 表名
	Columns     []*ColumnDef  // 列定义
	Constraints []*Constraint // 表约束
}

// ColumnDef 列定义节点
type ColumnDef struct {
	BaseNode
	Name        string        // 列名
	DataType    string        // 数据类型
	Constraints []*Constraint // 列约束
}

// Constraint 约束节点
type Constraint struct {
	BaseNode
	Type    string   // 约束类型(PRIMARY KEY/NOT NULL等)
	Columns []string // 涉及的列
}

// CreateIndexStmt CREATE INDEX语句节点
type CreateIndexStmt struct {
	BaseNode
	Name     string   // 索引名
	Table    string   // 表名
	Columns  []string // 索引列
	IsUnique bool     // 是否唯一索引
}

// BinaryExpr 二元表达式节点
type BinaryExpr struct {
	BaseNode
	Left     Node   // 左操作数
	Operator string // 运算符
	Right    Node   // 右操作数
}

// ComparisonExpr 比较表达式节点
type ComparisonExpr struct {
	BaseNode
	Left     Node   // 左操作数
	Operator string // 比较运算符
	Right    Node   // 右操作数
}

// LogicalExpr 逻辑表达式节点
type LogicalExpr struct {
	BaseNode
	Left     Node   // 左操作数
	Operator string // 逻辑运算符
	Right    Node   // 右操作数
}

// ColumnRef 列引用节点
type ColumnRef struct {
	BaseNode
	Table  string // 表名(可选)
	Column string // 列名
}

// Literal 字面量节点
type Literal struct {
	BaseNode
	Value interface{} // 字面量值
}

// IntLiteral 整型字面量节点
type IntLiteral struct {
	BaseNode
	Value int64 // 整数值
}

// StringLiteral 字符串字面量节点
type StringLiteral struct {
	BaseNode
	Value string // 字符串值
}

// FloatLiteral 浮点字面量节点
type FloatLiteral struct {
	BaseNode
	Value float64 // 浮点数值
}

// BooleanLiteral 布尔字面量节点
type BooleanLiteral struct {
	BaseNode
	Value bool // 布尔值
}

// TimestampLiteral 时间戳字面量节点
type TimestampLiteral struct {
	BaseNode
	Value time.Time // 时间戳值
}

// FunctionCall 函数调用节点
type FunctionCall struct {
	BaseNode
	Name string // 函数名
	Args []Node // 参数列表
}

// Identifier 标识符节点
type Identifier struct {
	BaseNode
	Value string // 标识符值
}

package parser

import "time"

// NodeType 定义AST节点类型
type NodeType int

const (
	// 语句节点类型
	SelectNode NodeType = iota
	InsertNode
	UpdateNode
	SelectItemNode
	UpdateAssignmentNode
	DeleteNode
	WhereNode
	LikeExprNode
	InExprNode
	TableRefNode
	JoinNode
	OrderByItemNode
	ColumnItemNode
	CreateDatabaseNode
	CreateTableNode
	CreateIndexNode
	DropIndexNode
	DropTableNode
	DropDatabaseNode
	UseNode
	ShowDatabasesNode
	ShowTablesNode
	ShowIndexesNode
	ExplainNode
	ErrorNode

	// 表达式节点类型
	BinaryExprNode       // 二元表达式
	ComparisonExprNode   // 比较表达式
	LogicalExprNode      // 逻辑表达式
	ColumnRefNode        // 列引用
	LiteralNode          // 字面量
	IntegerLiteralNode   // 整数字面量
	FloatLiteralNode     // 浮点数字面量
	StringLiteralNode    // 字符串字面量
	BooleanLiteralNode   // 布尔字面量
	TimestampLiteralNode // 时间戳字面量
	FunctionCallNode     // 函数调用
	IdentifierNode       // 标识符
	AsteriskNode

	// 新增节点类型
	PartitionMethodNode

	// 新增事务节点类型
	TransactionNode

	// 新增HavingClause节点类型
	HavingNode
)

// ColumnItemType 定义了SELECT列项的类型
type ColumnItemType int

const (
	ColumnItemUnknown    ColumnItemType = iota
	ColumnItemColumn                    // 普通列引用
	ColumnItemFunction                  // 函数调用
	ColumnItemExpression                // 计算表达式（非函数、非直接列引用）
	ColumnItemLiteral                   // 字面量
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
	All       bool           // 是否选择所有列
	Columns   []*ColumnItem  // 选择的列
	From      string         // FROM子句表名
	FromAlias string         // FROM子句表别名
	Joins     []*JoinClause  // JOIN子句
	Where     *WhereClause   // WHERE子句
	GroupBy   []Node         // GROUP BY子句
	Having    *HavingClause  // HAVING子句
	OrderBy   []*OrderByItem // ORDER BY子句
	Limit     int64          // LIMIT子句
}

// UseStmt USE语句节点
type UseStmt struct {
	BaseNode
	Database string // 数据库名
}

// ShowTablesStmt SHOW TABLES语句节点
type ShowTablesStmt struct {
	BaseNode
	Database string // 可选的数据库名
}

// SelectItem 用于构建 SELECT 项的临时结构
type SelectItem struct {
	BaseNode
	All    bool   // 是否选择所有列 *
	Table  string // 表名(可选)
	Column string // 列名
	Expr   Node   // 表达式
	Alias  string // 别名(可选)
}

// ColumnItem 表示 SELECT 列项的 AST 节点
type ColumnItem struct {
	BaseNode
	Table  string         // 如果限定了表名则记录
	Column string         // 如果是列引用则保存列名
	Expr   Node           // 如果是函数调用、表达式或字面量则记录整个表达式节点
	Alias  string         // 别名（可选）
	Kind   ColumnItemType // 新增字段，用于标识列项的来源类型
}

// TableRef 用于构建表引用的临时结构
type TableRef struct {
	BaseNode
	Table    string        // 表名（基本表引用时使用）
	Alias    string        // 别名
	Subquery *SelectStmt   // 子查询（子查询时使用）
	Joins    []*JoinClause // JOIN子句列表
}

// JoinClause JOIN子句节点
type JoinClause struct {
	BaseNode
	JoinType  string    // JOIN类型(INNER/LEFT/RIGHT/FULL)
	Left      *TableRef // 左表
	Right     *TableRef // 右表
	Condition Node      // ON条件
}

// WhereClause WHERE子句节点
type WhereClause struct {
	BaseNode
	Condition Node // WHERE条件
}

// OrderByItem ORDER BY 子句中的排序项
type OrderByItem struct {
	BaseNode
	Expr      Node   // 排序表达式
	Direction string // 排序方向：ASC 或 DESC
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

// DropIndexStmt DROP INDEX语句节点
type DropIndexStmt struct {
	BaseNode
	Name  string // 索引名
	Table string // 表名
}

// ShowIndexesStmt SHOW INDEXES语句节点
type ShowIndexesStmt struct {
	BaseNode
	Table string // 表名
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

// IntegerLiteral 整数字面量节点
type IntegerLiteral struct {
	BaseNode
	Value int64
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

// DropTableStmt DROP TABLE语句节点
type DropTableStmt struct {
	BaseNode
	Table string // 表名
}

// DropDatabaseStmt DROP DATABASE语句节点
type DropDatabaseStmt struct {
	BaseNode
	Database string // 数据库名
}

// InExpr IN表达式节点
type InExpr struct {
	BaseNode
	Left     Node   // 左操作数
	Operator string // IN 或 NOT IN
	Values   []Node // 值列表
}

// Asterisk 表示 SELECT * 中的星号
type Asterisk struct {
	BaseNode
}

// ShowDatabasesStmt SHOW DATABASES语句节点
type ShowDatabasesStmt struct {
	BaseNode
}

// PartitionMethod 分区方法节点
type PartitionMethod struct {
	BaseNode
	Type         string   // 分区类型(HASH/RANGE)
	Columns      []string // 分区键列
	PartitionNum int      // 分区数量(HASH分区)
}

// TransactionStmt 事务语句节点
type TransactionStmt struct {
	BaseNode
	TxType string // BEGIN/COMMIT/ROLLBACK
}

// ExplainStmt EXPLAIN语句节点
type ExplainStmt struct {
	BaseNode
	Query Node // 要解释的查询语句
}

// ErrorStmt 错误语句节点
type ErrorStmt struct {
	BaseNode
	Message string // 错误信息
}

// HavingClause HAVING子句节点
type HavingClause struct {
	BaseNode
	Condition Node // HAVING条件表达式
}

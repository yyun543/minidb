package parser

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
)

// AST节点类型枚举
type NodeType int

const (
	SelectNode NodeType = iota
	CreateTableNode
	InsertNode
	UpdateNode
	DeleteNode
	JoinNode
	WhereNode
	ExpressionNode
)

// AST节点接口
type Node interface {
	Type() NodeType
}

// SQL语句的基础节点结构
type BaseNode struct {
	nodeType NodeType
}

func (n *BaseNode) Type() NodeType {
	return n.nodeType
}

// Select语句节点
type SelectStmt struct {
	BaseNode
	Columns []string      // 选择的列
	From    string        // 表名
	Joins   []*JoinClause // JOIN子句
	Where   *WhereClause  // WHERE子句
	GroupBy []string      // GROUP BY子句
	OrderBy []string      // ORDER BY子句
	Limit   int           // LIMIT子句
}

// JOIN子句节点
type JoinClause struct {
	BaseNode
	JoinType  string // JOIN类型(INNER/LEFT)
	Table     string // 连接表名
	Condition Node   // 连接条件
}

// WHERE子句节点
type WhereClause struct {
	BaseNode
	Condition Node // 条件表达式
}

// 表达式节点
type Expression struct {
	BaseNode
	Left         Node        // 左操作数
	Operator     string      // 操作符
	Right        Node        // 右操作数
	Value        interface{} // 字面量值
	FunctionArgs []Node      // 函数参数
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
	Assignments []*UpdateAssignment // 更新赋值列表
	Where       *WhereClause        // WHERE子句
}

// UpdateAssignment 更新赋值
type UpdateAssignment struct {
	Column string // 列名
	Value  Node   // 新值
}

// DeleteStmt DELETE语句节点
type DeleteStmt struct {
	BaseNode
	Table string       // 表名
	Where *WhereClause // WHERE子句
}

// MiniQL访问器实现
type MiniQLVisitorImpl struct {
	BaseMiniQLVisitor
}

// Parse 函数是对外的主要接口
func Parse(sql string) (Node, error) {
	// 创建输入流
	input := antlr.NewInputStream(sql)

	// 创建词法分析器
	lexer := NewMiniQLLexer(input)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(antlr.NewDiagnosticErrorListener(true))

	// 创建语法分析器
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := NewMiniQLParser(stream)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(antlr.NewDiagnosticErrorListener(true))

	// 获取语法树
	tree := parser.Parse()

	// 创建访问器
	visitor := &MiniQLVisitorImpl{}

	// 访问语法树并构建AST
	result := visitor.Visit(tree)
	if result == nil {
		return nil, fmt.Errorf("解析SQL失败")
	}

	// 类型断言为Node接口
	node, ok := result.(Node)
	if !ok {
		return nil, fmt.Errorf("无法将解析结果转换为AST节点")
	}

	return node, nil
}

// Visit 实现通用访问方法
func (v *MiniQLVisitorImpl) Visit(tree antlr.ParseTree) interface{} {
	switch node := tree.(type) {
	case *ParseContext:
		return v.VisitParse(node)
	case *SqlStatementContext:
		return v.VisitSqlStatement(node)
	default:
		return nil
	}
}

// VisitParse 访问根节点
func (v *MiniQLVisitorImpl) VisitParse(ctx *ParseContext) interface{} {
	if len(ctx.AllSqlStatement()) > 0 {
		return v.Visit(ctx.SqlStatement(0))
	}
	return nil
}

// VisitSqlStatement 访问SQL语句节点
func (v *MiniQLVisitorImpl) VisitSqlStatement(ctx *SqlStatementContext) interface{} {
	if ctx.DqlStatement() != nil {
		return v.Visit(ctx.DqlStatement())
	}
	if ctx.DmlStatement() != nil {
		return v.Visit(ctx.DmlStatement())
	}
	if ctx.DdlStatement() != nil {
		return v.Visit(ctx.DdlStatement())
	}
	return nil
}

// VisitSelectStatement 访问SELECT语句节点
func (v *MiniQLVisitorImpl) VisitSelectStatement(ctx *SelectStatementContext) interface{} {
	stmt := &SelectStmt{
		BaseNode: BaseNode{nodeType: SelectNode},
	}

	// 解析选择的列
	for _, item := range ctx.AllSelectItem() {
		if colName, ok := v.Visit(item).(string); ok {
			stmt.Columns = append(stmt.Columns, colName)
		}
	}

	// 解析FROM子句
	if tableRef := ctx.TableReference(); tableRef != nil {
		if tableName, ok := v.Visit(tableRef).(string); ok {
			stmt.From = tableName
		}
	}

	// 解析WHERE子句
	if where := ctx.Expression(0); where != nil {
		if condition, ok := v.Visit(where).(Node); ok {
			stmt.Where = &WhereClause{
				BaseNode:  BaseNode{nodeType: WhereNode},
				Condition: condition,
			}
		}
	}

	return stmt
}

// VisitExpression 访问表达式节点
func (v *MiniQLVisitorImpl) VisitBinaryArithExpr(ctx *BinaryArithExprContext) interface{} {
	expr := &Expression{
		BaseNode: BaseNode{nodeType: ExpressionNode},
		Operator: ctx.GetOperator().GetText(),
	}

	// 解析左操作数
	if left := ctx.GetLeft(); left != nil {
		if leftNode, ok := v.Visit(left).(Node); ok {
			expr.Left = leftNode
		}
	}

	// 解析右操作数
	if right := ctx.GetRight(); right != nil {
		if rightNode, ok := v.Visit(right).(Node); ok {
			expr.Right = rightNode
		}
	}

	return expr
}

// VisitTableReference 访问表引用节点
func (v *MiniQLVisitorImpl) VisitTableRefBase(ctx *TableRefBaseContext) interface{} {
	// 获取表名
	tableName := v.Visit(ctx.TableName()).(string)

	// 处理可选的别名
	if ctx.Identifier() != nil {
		// 返回带别名的表引用
		return fmt.Sprintf("%s AS %s", tableName, ctx.Identifier().GetText())
	}
	return tableName
}

// VisitTableRefJoin 访问JOIN表引用节点
func (v *MiniQLVisitorImpl) VisitTableRefJoin(ctx *TableRefJoinContext) interface{} {
	join := &JoinClause{
		BaseNode: BaseNode{nodeType: JoinNode},
	}

	// 获取JOIN类型
	if joinType := ctx.JoinType(); joinType != nil {
		if joinType.LEFT() != nil {
			join.JoinType = "LEFT"
		} else {
			join.JoinType = "INNER"
		}
	} else {
		join.JoinType = "INNER" // 默认INNER JOIN
	}

	// 获取右表
	rightTable := v.Visit(ctx.TableReference(1)).(string)
	join.Table = rightTable

	// 获取JOIN条件
	if expr := ctx.Expression(); expr != nil {
		if condition, ok := v.Visit(expr).(Node); ok {
			join.Condition = condition
		}
	}

	return join
}

// VisitTableName 访问表名节点
func (v *MiniQLVisitorImpl) VisitTableName(ctx *TableNameContext) interface{} {
	return ctx.Identifier().GetText()
}

// VisitLiteralExpr 访问字面量表达式
func (v *MiniQLVisitorImpl) VisitLiteralExpr(ctx *LiteralExprContext) interface{} {
	return &Expression{
		BaseNode: BaseNode{nodeType: ExpressionNode},
		Left:     nil,
		Operator: "LITERAL",
		Right:    nil,
		Value:    v.Visit(ctx.Literal()),
	}
}

// VisitColumnRefExpr 访问列引用表达式
func (v *MiniQLVisitorImpl) VisitColumnRefExpr(ctx *ColumnRefExprContext) interface{} {
	return &Expression{
		BaseNode: BaseNode{nodeType: ExpressionNode},
		Left:     nil,
		Operator: "COLUMN_REF",
		Right:    nil,
		Value:    ctx.Identifier().GetText(),
	}
}

// VisitQualifiedColumnRef 访问限定列引用表达式
func (v *MiniQLVisitorImpl) VisitQualifiedColumnRef(ctx *QualifiedColumnRefContext) interface{} {
	tableName := v.Visit(ctx.TableName()).(string)
	columnName := ctx.Identifier().GetText()
	return &Expression{
		BaseNode: BaseNode{nodeType: ExpressionNode},
		Left:     nil,
		Operator: "QUALIFIED_COLUMN_REF",
		Right:    nil,
		Value:    fmt.Sprintf("%s.%s", tableName, columnName),
	}
}

// VisitComparisonExpr 访问比较表达式
func (v *MiniQLVisitorImpl) VisitComparisonExpr(ctx *ComparisonExprContext) interface{} {
	expr := &Expression{
		BaseNode: BaseNode{nodeType: ExpressionNode},
		Operator: ctx.COMPARISON_OP().GetText(),
	}

	// 解析左操作数
	if leftExpr := ctx.Expression(0); leftExpr != nil {
		if left, ok := v.Visit(leftExpr).(Node); ok {
			expr.Left = left
		}
	}

	// 解析右操作数
	if rightExpr := ctx.Expression(1); rightExpr != nil {
		if right, ok := v.Visit(rightExpr).(Node); ok {
			expr.Right = right
		}
	}

	return expr
}

// VisitLogicalExpr 访问逻辑表达式
func (v *MiniQLVisitorImpl) VisitLogicalExpr(ctx *LogicalExprContext) interface{} {
	var operator string
	if ctx.AND() != nil {
		operator = "AND"
	} else if ctx.OR() != nil {
		operator = "OR"
	}

	expr := &Expression{
		BaseNode: BaseNode{nodeType: ExpressionNode},
		Operator: operator,
	}

	// 解析左操作数
	if leftExpr := ctx.Expression(0); leftExpr != nil {
		if left, ok := v.Visit(leftExpr).(Node); ok {
			expr.Left = left
		}
	}

	// 解析右操作数
	if rightExpr := ctx.Expression(1); rightExpr != nil {
		if right, ok := v.Visit(rightExpr).(Node); ok {
			expr.Right = right
		}
	}

	return expr
}

// VisitFunctionCall 访问函数调用
func (v *MiniQLVisitorImpl) VisitFunctionCall(ctx *FunctionCallContext) interface{} {
	funcName := v.Visit(ctx.FunctionName()).(string)

	expr := &Expression{
		BaseNode: BaseNode{nodeType: ExpressionNode},
		Operator: "FUNCTION_CALL",
		Value:    funcName,
	}

	// 收集函数参数
	var args []Node
	for _, argCtx := range ctx.AllExpression() {
		if arg, ok := v.Visit(argCtx).(Node); ok {
			args = append(args, arg)
		}
	}

	// 将参数列表存储在FunctionArgs字段中
	expr.FunctionArgs = args

	return expr
}

// VisitCreateTable 访问CREATE TABLE语句
func (v *MiniQLVisitorImpl) VisitCreateTable(ctx *CreateTableContext) interface{} {
	stmt := &CreateTableStmt{
		BaseNode:  BaseNode{nodeType: CreateTableNode},
		TableName: ctx.TableName().GetText(),
	}

	// 解析列定义
	for _, colDef := range ctx.AllColumnDef() {
		if col, ok := v.Visit(colDef).(*ColumnDef); ok {
			stmt.Columns = append(stmt.Columns, col)
		}
	}

	// 解析表约束
	for _, constraint := range ctx.AllTableConstraint() {
		if tc, ok := v.Visit(constraint).(*TableConstraint); ok {
			stmt.Constraints = append(stmt.Constraints, tc)
		}
	}

	return stmt
}

// VisitInsertStatement 访问INSERT语句
func (v *MiniQLVisitorImpl) VisitInsertStatement(ctx *InsertStatementContext) interface{} {
	stmt := &InsertStmt{
		BaseNode: BaseNode{nodeType: InsertNode},
		Table:    ctx.TableName().GetText(),
	}

	// 解析列名列表(如果有)
	if cols := ctx.IdentifierList(); cols != nil {
		for _, id := range cols.AllIdentifier() {
			stmt.Columns = append(stmt.Columns, id.GetText())
		}
	}

	// 解析值列表
	if valueList := ctx.ValueList(0); valueList != nil {
		for _, expr := range valueList.AllExpression() {
			if value, ok := v.Visit(expr).(Node); ok {
				stmt.Values = append(stmt.Values, value)
			}
		}
	}

	return stmt
}

// VisitUpdateStatement 访问UPDATE语句
func (v *MiniQLVisitorImpl) VisitUpdateStatement(ctx *UpdateStatementContext) interface{} {
	stmt := &UpdateStmt{
		BaseNode: BaseNode{nodeType: UpdateNode},
		Table:    ctx.TableName().GetText(),
	}

	// 解析SET子句
	for _, assignment := range ctx.AllUpdateAssignment() {
		if assign, ok := v.Visit(assignment).(*UpdateAssignment); ok {
			stmt.Assignments = append(stmt.Assignments, assign)
		}
	}

	// 解析WHERE子句
	if where := ctx.Expression(); where != nil {
		if condition, ok := v.Visit(where).(Node); ok {
			stmt.Where = &WhereClause{
				BaseNode:  BaseNode{nodeType: WhereNode},
				Condition: condition,
			}
		}
	}

	return stmt
}

// VisitDeleteStatement 访问DELETE语句
func (v *MiniQLVisitorImpl) VisitDeleteStatement(ctx *DeleteStatementContext) interface{} {
	stmt := &DeleteStmt{
		BaseNode: BaseNode{nodeType: DeleteNode},
		Table:    ctx.TableName().GetText(),
	}

	// 解析WHERE子句
	if where := ctx.Expression(); where != nil {
		if condition, ok := v.Visit(where).(Node); ok {
			stmt.Where = &WhereClause{
				BaseNode:  BaseNode{nodeType: WhereNode},
				Condition: condition,
			}
		}
	}

	return stmt
}

// VisitLiteral 访问字面量
func (v *MiniQLVisitorImpl) VisitLiteral(ctx *LiteralContext) interface{} {
	if ctx.STRING() != nil {
		return ctx.STRING().GetText()
	}
	if ctx.INTEGER() != nil {
		return ctx.INTEGER().GetText()
	}
	if ctx.FLOAT() != nil {
		return ctx.FLOAT().GetText()
	}
	if ctx.NULL() != nil {
		return nil
	}
	return nil
}

// VisitIdentifier 访问标识符
func (v *MiniQLVisitorImpl) VisitIdentifier(ctx *IdentifierContext) interface{} {
	return ctx.IDENTIFIER().GetText()
}

// CreateTableStmt CREATE TABLE语句节点
type CreateTableStmt struct {
	BaseNode
	TableName   string             // 表名
	Columns     []*ColumnDef       // 列定义
	Constraints []*TableConstraint // 表约束
}

// ColumnDef 列定义
type ColumnDef struct {
	Name     string   // 列名
	DataType string   // 数据类型
	Options  []string // 列选项(NOT NULL等)
}

// TableConstraint 表约束
type TableConstraint struct {
	Type       string   // 约束类型(PRIMARY KEY等)
	Columns    []string // 涉及的列
	Definition string   // 约束定义
}

// TODO: 其他必要的访问方法实现...

package parser

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
)

// MiniQLVisitorImpl 是 MiniQL 的访问器实现
type MiniQLVisitorImpl struct {
	BaseMiniQLVisitor
	currentDatabase string // 当前选中的数据库
}

// Parse 是对外主要接口，传入 SQL 字符串返回 AST 节点
func Parse(sql string) (Node, error) {
	// 创建词法分析器
	input := antlr.NewInputStream(sql)
	lexer := NewMiniQLLexer(input)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(antlr.NewDiagnosticErrorListener(true))

	// 创建语法分析器
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := NewMiniQLParser(stream)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(antlr.NewDiagnosticErrorListener(true))

	// 生成语法树，并开始访问
	parseTree := parser.Parse()
	visitor := &MiniQLVisitorImpl{}
	result := visitor.Visit(parseTree)
	if result == nil {
		return nil, fmt.Errorf("解析SQL失败")
	}

	node, ok := result.(Node)
	if !ok {
		return nil, fmt.Errorf("解析结果类型错误")
	}

	return node, nil
}

// Visit 实现通用访问方法
func (v *MiniQLVisitorImpl) Visit(tree antlr.ParseTree) interface{} {
	return tree.Accept(v)
}

// VisitParse 访问根语法规则 parse
func (v *MiniQLVisitorImpl) VisitParse(ctx *ParseContext) interface{} {
	if len(ctx.AllSqlStatement()) > 0 {
		return v.Visit(ctx.SqlStatement(0))
	}
	return nil
}

// VisitSqlStatement 访问 SQL 语句节点
func (v *MiniQLVisitorImpl) VisitSqlStatement(ctx *SqlStatementContext) interface{} {
	// 按照语法规则定义的顺序检查
	if ctx.DmlStatement() != nil {
		return v.Visit(ctx.DmlStatement())
	}
	if ctx.DqlStatement() != nil {
		return v.Visit(ctx.DqlStatement())
	}
	if ctx.DdlStatement() != nil {
		return v.Visit(ctx.DdlStatement())
	}
	if ctx.DclStatement() != nil {
		return v.Visit(ctx.DclStatement())
	}
	if ctx.UtilityStatement() != nil {
		return v.Visit(ctx.UtilityStatement())
	}
	return nil
}

// VisitDqlStatement 处理 DQL 语句（SELECT）
func (v *MiniQLVisitorImpl) VisitDqlStatement(ctx *DqlStatementContext) interface{} {
	if ctx.SelectStatement() != nil {
		return v.Visit(ctx.SelectStatement())
	}
	return nil
}

// VisitDmlStatement 处理 DML 语句（INSERT、UPDATE、DELETE）
func (v *MiniQLVisitorImpl) VisitDmlStatement(ctx *DmlStatementContext) interface{} {
	if ctx.InsertStatement() != nil {
		return v.Visit(ctx.InsertStatement())
	}
	if ctx.UpdateStatement() != nil {
		return v.Visit(ctx.UpdateStatement())
	}
	if ctx.DeleteStatement() != nil {
		return v.Visit(ctx.DeleteStatement())
	}
	return nil
}

// VisitDdlStatement 处理 DDL 语句（CREATE、DROP、ALTER）
func (v *MiniQLVisitorImpl) VisitDdlStatement(ctx *DdlStatementContext) interface{} {
	if ctx.CreateDatabase() != nil {
		return v.Visit(ctx.CreateDatabase())
	}
	if ctx.CreateTable() != nil {
		return v.Visit(ctx.CreateTable())
	}
	if ctx.CreateIndex() != nil {
		return v.Visit(ctx.CreateIndex())
	}
	if ctx.DropDatabase() != nil {
		return v.Visit(ctx.DropDatabase())
	}
	if ctx.DropTable() != nil {
		return v.Visit(ctx.DropTable())
	}
	return nil
}

// VisitDclStatement 处理 DCL 语句（事务）
func (v *MiniQLVisitorImpl) VisitDclStatement(ctx *DclStatementContext) interface{} {
	if ctx.TransactionStatement() != nil {
		return v.Visit(ctx.TransactionStatement())
	}
	return nil
}

// VisitUtilityStatement 处理工具语句（USE、SHOW）
func (v *MiniQLVisitorImpl) VisitUtilityStatement(ctx *UtilityStatementContext) interface{} {
	if ctx.UseStatement() != nil {
		return v.Visit(ctx.UseStatement())
	}
	if ctx.ShowDatabases() != nil {
		return v.Visit(ctx.ShowDatabases())
	}
	if ctx.ShowTables() != nil {
		return v.Visit(ctx.ShowTables())
	}
	if ctx.ExplainStatement() != nil {
		return v.Visit(ctx.ExplainStatement())
	}
	return nil
}

// VisitCreateDatabase 访问 CREATE DATABASE 语句
func (v *MiniQLVisitorImpl) VisitCreateDatabase(ctx *CreateDatabaseContext) interface{} {
	// 1. 创建 CreateDatabaseStmt 节点
	stmt := &CreateDatabaseStmt{
		BaseNode: BaseNode{nodeType: CreateDatabaseNode},
	}

	// 2. 获取数据库名称
	// 根据语法规则，数据库名是一个标识符
	if ctx.Identifier() != nil {
		if result := v.Visit(ctx.Identifier()); result != nil {
			switch t := result.(type) {
			case string:
				stmt.Database = t
			case *Identifier:
				stmt.Database = t.Value
			}
		}
	}

	return stmt
}

// VisitCreateTable 访问 CREATE TABLE 语句
func (v *MiniQLVisitorImpl) VisitCreateTable(ctx *CreateTableContext) interface{} {
	// 创建 CreateTableStmt 节点
	stmt := &CreateTableStmt{
		BaseNode: BaseNode{nodeType: CreateTableNode},
	}

	// 获取表名
	if ctx.TableName() != nil {
		if result := v.Visit(ctx.TableName()); result != nil {
			switch t := result.(type) {
			case string:
				stmt.Table = t
			case *Identifier:
				stmt.Table = t.Value
			}
		}
	}

	// 处理列定义
	for _, colDef := range ctx.AllColumnDef() {
		if result := v.Visit(colDef); result != nil {
			if col, ok := result.(*ColumnDef); ok {
				stmt.Columns = append(stmt.Columns, col)
			}
		}
	}

	// 处理表级约束
	for _, constraint := range ctx.AllTableConstraint() {
		if result := v.Visit(constraint); result != nil {
			if tc, ok := result.(*Constraint); ok {
				stmt.Constraints = append(stmt.Constraints, tc)
			}
		}
	}

	return stmt
}

// VisitColumnDef 访问列定义
func (v *MiniQLVisitorImpl) VisitColumnDef(ctx *ColumnDefContext) interface{} {
	colDef := &ColumnDef{}

	// 获取列名
	if ctx.Identifier() != nil {
		colDef.Name = ctx.Identifier().GetText()
	}

	// 获取数据类型
	if ctx.DataType() != nil {
		if result := v.Visit(ctx.DataType()); result != nil {
			switch t := result.(type) {
			case string:
				colDef.DataType = t
			}
		}
	}

	// 处理列级约束
	for _, constraint := range ctx.AllColumnConstraint() {
		if result := v.Visit(constraint); result != nil {
			if cc, ok := result.(*Constraint); ok {
				colDef.Constraints = append(colDef.Constraints, cc)
			}
		}
	}

	return colDef
}

// VisitColumnConstraint 访问列级约束
func (v *MiniQLVisitorImpl) VisitColumnConstraint(ctx *ColumnConstraintContext) interface{} {
	constraint := &Constraint{}

	if ctx.PRIMARY() != nil && ctx.KEY() != nil {
		constraint.Type = PrimaryKeyConstraint
	} else if ctx.NOT() != nil && ctx.NULL() != nil {
		constraint.Type = NotNullConstraint
	}

	return constraint
}

// VisitTableConstraint 访问表级约束
func (v *MiniQLVisitorImpl) VisitTableConstraint(ctx *TableConstraintContext) interface{} {
	constraint := &Constraint{}

	if ctx.PRIMARY() != nil && ctx.KEY() != nil {
		constraint.Type = PrimaryKeyConstraint
		// 获取主键列名列表
		if ctx.IdentifierList() != nil {
			if result := v.Visit(ctx.IdentifierList()); result != nil {
				switch t := result.(type) {
				case []string:
					constraint.Columns = t
				}
			}
		}
	}

	return constraint
}

// TODO 访问 CreateIndex 语句
func (v *MiniQLVisitorImpl) VisitCreateIndex(ctx *CreateIndexContext) interface{} {
	return nil
}

// TODO 访问 DropTable 语句
func (v *MiniQLVisitorImpl) VisitDropTable(ctx *DropTableContext) interface{} {
	return nil
}

// TODO 访问 DropDatabase 语句
func (v *MiniQLVisitorImpl) VisitDropDatabase(ctx *DropDatabaseContext) interface{} {
	return nil
}

// TODO 访问 InsertStatement 语句
func (v *MiniQLVisitorImpl) VisitInsertStatement(ctx *InsertStatementContext) interface{} {
	return nil
}

// TODO 访问 UpdateStatement 语句
func (v *MiniQLVisitorImpl) VisitUpdateStatement(ctx *UpdateStatementContext) interface{} {
	return nil
}

// TODO 访问 DeleteStatement 语句
func (v *MiniQLVisitorImpl) VisitDeleteStatement(ctx *DeleteStatementContext) interface{} {
	return nil
}

// TODO 访问 SelectStatement 语句
func (v *MiniQLVisitorImpl) VisitSelectStatement(ctx *SelectStatementContext) interface{} {
	return nil
}

// TODO 访问 SelectAll 语句
func (v *MiniQLVisitorImpl) VisitSelectAll(ctx *SelectAllContext) interface{} {
	return nil
}

// TODO 访问 SelectExpr 语句
func (v *MiniQLVisitorImpl) VisitSelectExpr(ctx *SelectExprContext) interface{} {
	return nil
}

// TODO 访问 TableRefBase 语句
func (v *MiniQLVisitorImpl) VisitTableRefBase(ctx *TableRefBaseContext) interface{} {
	return nil
}

// TODO 访问 TableRefJoin 语句
func (v *MiniQLVisitorImpl) VisitTableRefJoin(ctx *TableRefJoinContext) interface{} {
	return nil
}

// TODO 访问 TableRefSubquery 语句
func (v *MiniQLVisitorImpl) VisitTableRefSubquery(ctx *TableRefSubqueryContext) interface{} {
	return nil
}

// TODO 访问 JoinType 语句
func (v *MiniQLVisitorImpl) VisitJoinType(ctx *JoinTypeContext) interface{} {
	return nil
}

// TODO 访问 PrimaryExpression 语句
func (v *MiniQLVisitorImpl) VisitPrimaryExpression(ctx *PrimaryExpressionContext) interface{} {
	return nil
}

// TODO 访问 OrExpression 语句
func (v *MiniQLVisitorImpl) VisitOrExpression(ctx *OrExpressionContext) interface{} {
	return nil
}

// TODO 访问 AndExpression 语句
func (v *MiniQLVisitorImpl) VisitAndExpression(ctx *AndExpressionContext) interface{} {
	return nil
}

// TODO 访问 InExpression 语句
func (v *MiniQLVisitorImpl) VisitInExpression(ctx *InExpressionContext) interface{} {
	return nil
}

// TODO 访问 LikeExpression 语句
func (v *MiniQLVisitorImpl) VisitLikeExpression(ctx *LikeExpressionContext) interface{} {
	return nil
}

// TODO 访问 ComparisonExpression 语句
func (v *MiniQLVisitorImpl) VisitComparisonExpression(ctx *ComparisonExpressionContext) interface{} {
	return nil
}

// TODO 访问 LiteralExpr 语句
func (v *MiniQLVisitorImpl) VisitLiteralExpr(ctx *LiteralExprContext) interface{} {
	return nil
}

// TODO 访问 ColumnRefExpr 语句
func (v *MiniQLVisitorImpl) VisitColumnRefExpr(ctx *ColumnRefExprContext) interface{} {
	return nil
}

// TODO 访问 FunctionCallExpr 语句
func (v *MiniQLVisitorImpl) VisitFunctionCallExpr(ctx *FunctionCallExprContext) interface{} {
	return nil
}

// TODO 访问 ParenExpr 语句
func (v *MiniQLVisitorImpl) VisitParenExpr(ctx *ParenExprContext) interface{} {
	return nil
}

// TODO 访问 ComparisonOperator 语句
func (v *MiniQLVisitorImpl) VisitComparisonOperator(ctx *ComparisonOperatorContext) interface{} {
	return nil
}

// TODO 访问 ColumnRef 语句
func (v *MiniQLVisitorImpl) VisitColumnRef(ctx *ColumnRefContext) interface{} {
	return nil
}

// TODO 访问 UpdateAssignment 语句
func (v *MiniQLVisitorImpl) VisitUpdateAssignment(ctx *UpdateAssignmentContext) interface{} {
	return nil
}

// TODO 访问 GroupByItem 语句
func (v *MiniQLVisitorImpl) VisitGroupByItem(ctx *GroupByItemContext) interface{} {
	return nil
}

// TODO 访问 OrderByItem 语句
func (v *MiniQLVisitorImpl) VisitOrderByItem(ctx *OrderByItemContext) interface{} {
	return nil
}

// TODO 访问 FunctionCall 语句
func (v *MiniQLVisitorImpl) VisitFunctionCall(ctx *FunctionCallContext) interface{} {
	return nil
}

// TODO 访问 PartitionMethod 语句
func (v *MiniQLVisitorImpl) VisitPartitionMethod(ctx *PartitionMethodContext) interface{} {
	return nil
}

// TODO 访问 TransactionStatement 语句
func (v *MiniQLVisitorImpl) VisitTransactionStatement(ctx *TransactionStatementContext) interface{} {
	return nil
}

// TODO 访问 UseStatement 语句
func (v *MiniQLVisitorImpl) VisitUseStatement(ctx *UseStatementContext) interface{} {
	return nil
}

// TODO 访问 ShowDatabases 语句
func (v *MiniQLVisitorImpl) VisitShowDatabases(ctx *ShowDatabasesContext) interface{} {
	return nil
}

// TODO 访问 ShowTables 语句
func (v *MiniQLVisitorImpl) VisitShowTables(ctx *ShowTablesContext) interface{} {
	return nil
}

// TODO 访问 ExplainStatement 语句
func (v *MiniQLVisitorImpl) VisitExplainStatement(ctx *ExplainStatementContext) interface{} {
	return nil
}

// VisitIdentifierList 访问标识符列表
func (v *MiniQLVisitorImpl) VisitIdentifierList(ctx *IdentifierListContext) interface{} {
	identifiers := make([]string, 0)
	for _, id := range ctx.AllIdentifier() {
		if result := v.Visit(id); result != nil {
			switch t := result.(type) {
			case string:
				identifiers = append(identifiers, t)
			}
		}
	}
	return identifiers
}

// TODO 访问 ValueList 语句
func (v *MiniQLVisitorImpl) VisitValueList(ctx *ValueListContext) interface{} {
	return nil
}

// VisitTableName 访问表名节点
func (v *MiniQLVisitorImpl) VisitTableName(ctx *TableNameContext) interface{} {
	if ctx == nil {
		return nil
	}
	// 直接访问标识符节点并返回其文本值
	if id := ctx.Identifier(); id != nil {
		return v.Visit(id)
	}
	return nil
}

// VisitIdentifier 访问标识符节点
func (v *MiniQLVisitorImpl) VisitIdentifier(ctx *IdentifierContext) interface{} {
	if ctx == nil {
		return nil
	}

	// 直接返回标识符文本值，避免创建不必要的中间对象
	// 只有在确实需要 Identifier 节点的地方才创建
	return ctx.GetText()
}

// VisitDataType 访问数据类型
func (v *MiniQLVisitorImpl) VisitDataType(ctx *DataTypeContext) interface{} {
	if ctx.INTEGER_TYPE() != nil {
		return "INTEGER"
	}
	if ctx.VARCHAR_TYPE() != nil {
		if ctx.INTEGER_LITERAL() != nil {
			return fmt.Sprintf("VARCHAR(%s)", ctx.INTEGER_LITERAL().GetText())
		}
		return "VARCHAR"
	}
	if ctx.BOOLEAN_TYPE() != nil {
		return "BOOLEAN"
	}
	if ctx.DOUBLE_TYPE() != nil {
		return "DOUBLE"
	}
	if ctx.TIMESTAMP_TYPE() != nil {
		return "TIMESTAMP"
	}
	return nil
}

// TODO 访问 Literal 语句
func (v *MiniQLVisitorImpl) VisitLiteral(ctx *LiteralContext) interface{} {
	return nil
}

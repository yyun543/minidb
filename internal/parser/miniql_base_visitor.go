// Code generated from /Users/yasonlee/codes/minidb/internal/parser/MiniQL.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // MiniQL

import "github.com/antlr4-go/antlr/v4"


type BaseMiniQLVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseMiniQLVisitor) VisitParse(ctx *ParseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitError(ctx *ErrorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitSqlStatement(ctx *SqlStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitDdlStatement(ctx *DdlStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitDmlStatement(ctx *DmlStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitDqlStatement(ctx *DqlStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitUtilityStatement(ctx *UtilityStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitCreateDatabase(ctx *CreateDatabaseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitCreateTable(ctx *CreateTableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitColumnDef(ctx *ColumnDefContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitColumnConstraint(ctx *ColumnConstraintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitTableConstraint(ctx *TableConstraintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitCreateIndex(ctx *CreateIndexContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitDropTable(ctx *DropTableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitDropDatabase(ctx *DropDatabaseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitInsertStatement(ctx *InsertStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitUpdateStatement(ctx *UpdateStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitDeleteStatement(ctx *DeleteStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitSelectStatement(ctx *SelectStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitSelectAll(ctx *SelectAllContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitSelectExpr(ctx *SelectExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitTableRefBase(ctx *TableRefBaseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitTableRefJoin(ctx *TableRefJoinContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitTableRefSubquery(ctx *TableRefSubqueryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitTableName(ctx *TableNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitIdentifierList(ctx *IdentifierListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitValueList(ctx *ValueListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitUpdateAssignment(ctx *UpdateAssignmentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitGroupByItem(ctx *GroupByItemContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitOrderByItem(ctx *OrderByItemContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitFunctionName(ctx *FunctionNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitLiteral(ctx *LiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitIdentifier(ctx *IdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitPartitionMethod(ctx *PartitionMethodContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitJoinType(ctx *JoinTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitUseStatement(ctx *UseStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitShowDatabases(ctx *ShowDatabasesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitShowTables(ctx *ShowTablesContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitExplainStatement(ctx *ExplainStatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitDataType(ctx *DataTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitBinaryArithExpr(ctx *BinaryArithExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitQualifiedColumnRef(ctx *QualifiedColumnRefContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitLiteralExpr(ctx *LiteralExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitFunctionCall(ctx *FunctionCallContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitLogicalExpr(ctx *LogicalExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitNestedExpr(ctx *NestedExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitComparisonExpr(ctx *ComparisonExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMiniQLVisitor) VisitColumnRefExpr(ctx *ColumnRefExprContext) interface{} {
	return v.VisitChildren(ctx)
}

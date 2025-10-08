// Code generated from MiniQL.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // MiniQL

import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by MiniQLParser.
type MiniQLVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by MiniQLParser#parse.
	VisitParse(ctx *ParseContext) interface{}

	// Visit a parse tree produced by MiniQLParser#sqlStatement.
	VisitSqlStatement(ctx *SqlStatementContext) interface{}

	// Visit a parse tree produced by MiniQLParser#ddlStatement.
	VisitDdlStatement(ctx *DdlStatementContext) interface{}

	// Visit a parse tree produced by MiniQLParser#dmlStatement.
	VisitDmlStatement(ctx *DmlStatementContext) interface{}

	// Visit a parse tree produced by MiniQLParser#dqlStatement.
	VisitDqlStatement(ctx *DqlStatementContext) interface{}

	// Visit a parse tree produced by MiniQLParser#dclStatement.
	VisitDclStatement(ctx *DclStatementContext) interface{}

	// Visit a parse tree produced by MiniQLParser#utilityStatement.
	VisitUtilityStatement(ctx *UtilityStatementContext) interface{}

	// Visit a parse tree produced by MiniQLParser#createDatabase.
	VisitCreateDatabase(ctx *CreateDatabaseContext) interface{}

	// Visit a parse tree produced by MiniQLParser#createTable.
	VisitCreateTable(ctx *CreateTableContext) interface{}

	// Visit a parse tree produced by MiniQLParser#columnDef.
	VisitColumnDef(ctx *ColumnDefContext) interface{}

	// Visit a parse tree produced by MiniQLParser#columnConstraint.
	VisitColumnConstraint(ctx *ColumnConstraintContext) interface{}

	// Visit a parse tree produced by MiniQLParser#tableConstraint.
	VisitTableConstraint(ctx *TableConstraintContext) interface{}

	// Visit a parse tree produced by MiniQLParser#createIndex.
	VisitCreateIndex(ctx *CreateIndexContext) interface{}

	// Visit a parse tree produced by MiniQLParser#dropIndex.
	VisitDropIndex(ctx *DropIndexContext) interface{}

	// Visit a parse tree produced by MiniQLParser#dropTable.
	VisitDropTable(ctx *DropTableContext) interface{}

	// Visit a parse tree produced by MiniQLParser#dropDatabase.
	VisitDropDatabase(ctx *DropDatabaseContext) interface{}

	// Visit a parse tree produced by MiniQLParser#insertStatement.
	VisitInsertStatement(ctx *InsertStatementContext) interface{}

	// Visit a parse tree produced by MiniQLParser#updateStatement.
	VisitUpdateStatement(ctx *UpdateStatementContext) interface{}

	// Visit a parse tree produced by MiniQLParser#deleteStatement.
	VisitDeleteStatement(ctx *DeleteStatementContext) interface{}

	// Visit a parse tree produced by MiniQLParser#selectStatement.
	VisitSelectStatement(ctx *SelectStatementContext) interface{}

	// Visit a parse tree produced by MiniQLParser#selectAll.
	VisitSelectAll(ctx *SelectAllContext) interface{}

	// Visit a parse tree produced by MiniQLParser#selectExpr.
	VisitSelectExpr(ctx *SelectExprContext) interface{}

	// Visit a parse tree produced by MiniQLParser#tableReference.
	VisitTableReference(ctx *TableReferenceContext) interface{}

	// Visit a parse tree produced by MiniQLParser#tableRefBase.
	VisitTableRefBase(ctx *TableRefBaseContext) interface{}

	// Visit a parse tree produced by MiniQLParser#tableRefSubquery.
	VisitTableRefSubquery(ctx *TableRefSubqueryContext) interface{}

	// Visit a parse tree produced by MiniQLParser#joinType.
	VisitJoinType(ctx *JoinTypeContext) interface{}

	// Visit a parse tree produced by MiniQLParser#primaryExpression.
	VisitPrimaryExpression(ctx *PrimaryExpressionContext) interface{}

	// Visit a parse tree produced by MiniQLParser#orExpression.
	VisitOrExpression(ctx *OrExpressionContext) interface{}

	// Visit a parse tree produced by MiniQLParser#andExpression.
	VisitAndExpression(ctx *AndExpressionContext) interface{}

	// Visit a parse tree produced by MiniQLParser#inExpression.
	VisitInExpression(ctx *InExpressionContext) interface{}

	// Visit a parse tree produced by MiniQLParser#likeExpression.
	VisitLikeExpression(ctx *LikeExpressionContext) interface{}

	// Visit a parse tree produced by MiniQLParser#comparisonExpression.
	VisitComparisonExpression(ctx *ComparisonExpressionContext) interface{}

	// Visit a parse tree produced by MiniQLParser#literalExpr.
	VisitLiteralExpr(ctx *LiteralExprContext) interface{}

	// Visit a parse tree produced by MiniQLParser#columnRefExpr.
	VisitColumnRefExpr(ctx *ColumnRefExprContext) interface{}

	// Visit a parse tree produced by MiniQLParser#functionCallExpr.
	VisitFunctionCallExpr(ctx *FunctionCallExprContext) interface{}

	// Visit a parse tree produced by MiniQLParser#parenExpr.
	VisitParenExpr(ctx *ParenExprContext) interface{}

	// Visit a parse tree produced by MiniQLParser#comparisonOperator.
	VisitComparisonOperator(ctx *ComparisonOperatorContext) interface{}

	// Visit a parse tree produced by MiniQLParser#columnRef.
	VisitColumnRef(ctx *ColumnRefContext) interface{}

	// Visit a parse tree produced by MiniQLParser#updateAssignment.
	VisitUpdateAssignment(ctx *UpdateAssignmentContext) interface{}

	// Visit a parse tree produced by MiniQLParser#groupByItem.
	VisitGroupByItem(ctx *GroupByItemContext) interface{}

	// Visit a parse tree produced by MiniQLParser#orderByItem.
	VisitOrderByItem(ctx *OrderByItemContext) interface{}

	// Visit a parse tree produced by MiniQLParser#functionCall.
	VisitFunctionCall(ctx *FunctionCallContext) interface{}

	// Visit a parse tree produced by MiniQLParser#partitionMethod.
	VisitPartitionMethod(ctx *PartitionMethodContext) interface{}

	// Visit a parse tree produced by MiniQLParser#transactionStatement.
	VisitTransactionStatement(ctx *TransactionStatementContext) interface{}

	// Visit a parse tree produced by MiniQLParser#useStatement.
	VisitUseStatement(ctx *UseStatementContext) interface{}

	// Visit a parse tree produced by MiniQLParser#showDatabases.
	VisitShowDatabases(ctx *ShowDatabasesContext) interface{}

	// Visit a parse tree produced by MiniQLParser#showTables.
	VisitShowTables(ctx *ShowTablesContext) interface{}

	// Visit a parse tree produced by MiniQLParser#showIndexes.
	VisitShowIndexes(ctx *ShowIndexesContext) interface{}

	// Visit a parse tree produced by MiniQLParser#explainStatement.
	VisitExplainStatement(ctx *ExplainStatementContext) interface{}

	// Visit a parse tree produced by MiniQLParser#identifierList.
	VisitIdentifierList(ctx *IdentifierListContext) interface{}

	// Visit a parse tree produced by MiniQLParser#valueList.
	VisitValueList(ctx *ValueListContext) interface{}

	// Visit a parse tree produced by MiniQLParser#tableName.
	VisitTableName(ctx *TableNameContext) interface{}

	// Visit a parse tree produced by MiniQLParser#identifier.
	VisitIdentifier(ctx *IdentifierContext) interface{}

	// Visit a parse tree produced by MiniQLParser#dataType.
	VisitDataType(ctx *DataTypeContext) interface{}

	// Visit a parse tree produced by MiniQLParser#literal.
	VisitLiteral(ctx *LiteralContext) interface{}
}

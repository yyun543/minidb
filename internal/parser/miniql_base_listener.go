// Code generated from MiniQL.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // MiniQL
import "github.com/antlr4-go/antlr/v4"

// BaseMiniQLListener is a complete listener for a parse tree produced by MiniQLParser.
type BaseMiniQLListener struct{}

var _ MiniQLListener = &BaseMiniQLListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseMiniQLListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseMiniQLListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseMiniQLListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseMiniQLListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterParse is called when production parse is entered.
func (s *BaseMiniQLListener) EnterParse(ctx *ParseContext) {}

// ExitParse is called when production parse is exited.
func (s *BaseMiniQLListener) ExitParse(ctx *ParseContext) {}

// EnterSqlStatement is called when production sqlStatement is entered.
func (s *BaseMiniQLListener) EnterSqlStatement(ctx *SqlStatementContext) {}

// ExitSqlStatement is called when production sqlStatement is exited.
func (s *BaseMiniQLListener) ExitSqlStatement(ctx *SqlStatementContext) {}

// EnterDdlStatement is called when production ddlStatement is entered.
func (s *BaseMiniQLListener) EnterDdlStatement(ctx *DdlStatementContext) {}

// ExitDdlStatement is called when production ddlStatement is exited.
func (s *BaseMiniQLListener) ExitDdlStatement(ctx *DdlStatementContext) {}

// EnterDmlStatement is called when production dmlStatement is entered.
func (s *BaseMiniQLListener) EnterDmlStatement(ctx *DmlStatementContext) {}

// ExitDmlStatement is called when production dmlStatement is exited.
func (s *BaseMiniQLListener) ExitDmlStatement(ctx *DmlStatementContext) {}

// EnterDqlStatement is called when production dqlStatement is entered.
func (s *BaseMiniQLListener) EnterDqlStatement(ctx *DqlStatementContext) {}

// ExitDqlStatement is called when production dqlStatement is exited.
func (s *BaseMiniQLListener) ExitDqlStatement(ctx *DqlStatementContext) {}

// EnterDclStatement is called when production dclStatement is entered.
func (s *BaseMiniQLListener) EnterDclStatement(ctx *DclStatementContext) {}

// ExitDclStatement is called when production dclStatement is exited.
func (s *BaseMiniQLListener) ExitDclStatement(ctx *DclStatementContext) {}

// EnterUtilityStatement is called when production utilityStatement is entered.
func (s *BaseMiniQLListener) EnterUtilityStatement(ctx *UtilityStatementContext) {}

// ExitUtilityStatement is called when production utilityStatement is exited.
func (s *BaseMiniQLListener) ExitUtilityStatement(ctx *UtilityStatementContext) {}

// EnterCreateDatabase is called when production createDatabase is entered.
func (s *BaseMiniQLListener) EnterCreateDatabase(ctx *CreateDatabaseContext) {}

// ExitCreateDatabase is called when production createDatabase is exited.
func (s *BaseMiniQLListener) ExitCreateDatabase(ctx *CreateDatabaseContext) {}

// EnterCreateTable is called when production createTable is entered.
func (s *BaseMiniQLListener) EnterCreateTable(ctx *CreateTableContext) {}

// ExitCreateTable is called when production createTable is exited.
func (s *BaseMiniQLListener) ExitCreateTable(ctx *CreateTableContext) {}

// EnterColumnDef is called when production columnDef is entered.
func (s *BaseMiniQLListener) EnterColumnDef(ctx *ColumnDefContext) {}

// ExitColumnDef is called when production columnDef is exited.
func (s *BaseMiniQLListener) ExitColumnDef(ctx *ColumnDefContext) {}

// EnterColumnConstraint is called when production columnConstraint is entered.
func (s *BaseMiniQLListener) EnterColumnConstraint(ctx *ColumnConstraintContext) {}

// ExitColumnConstraint is called when production columnConstraint is exited.
func (s *BaseMiniQLListener) ExitColumnConstraint(ctx *ColumnConstraintContext) {}

// EnterTableConstraint is called when production tableConstraint is entered.
func (s *BaseMiniQLListener) EnterTableConstraint(ctx *TableConstraintContext) {}

// ExitTableConstraint is called when production tableConstraint is exited.
func (s *BaseMiniQLListener) ExitTableConstraint(ctx *TableConstraintContext) {}

// EnterCreateIndex is called when production createIndex is entered.
func (s *BaseMiniQLListener) EnterCreateIndex(ctx *CreateIndexContext) {}

// ExitCreateIndex is called when production createIndex is exited.
func (s *BaseMiniQLListener) ExitCreateIndex(ctx *CreateIndexContext) {}

// EnterDropIndex is called when production dropIndex is entered.
func (s *BaseMiniQLListener) EnterDropIndex(ctx *DropIndexContext) {}

// ExitDropIndex is called when production dropIndex is exited.
func (s *BaseMiniQLListener) ExitDropIndex(ctx *DropIndexContext) {}

// EnterDropTable is called when production dropTable is entered.
func (s *BaseMiniQLListener) EnterDropTable(ctx *DropTableContext) {}

// ExitDropTable is called when production dropTable is exited.
func (s *BaseMiniQLListener) ExitDropTable(ctx *DropTableContext) {}

// EnterDropDatabase is called when production dropDatabase is entered.
func (s *BaseMiniQLListener) EnterDropDatabase(ctx *DropDatabaseContext) {}

// ExitDropDatabase is called when production dropDatabase is exited.
func (s *BaseMiniQLListener) ExitDropDatabase(ctx *DropDatabaseContext) {}

// EnterInsertStatement is called when production insertStatement is entered.
func (s *BaseMiniQLListener) EnterInsertStatement(ctx *InsertStatementContext) {}

// ExitInsertStatement is called when production insertStatement is exited.
func (s *BaseMiniQLListener) ExitInsertStatement(ctx *InsertStatementContext) {}

// EnterUpdateStatement is called when production updateStatement is entered.
func (s *BaseMiniQLListener) EnterUpdateStatement(ctx *UpdateStatementContext) {}

// ExitUpdateStatement is called when production updateStatement is exited.
func (s *BaseMiniQLListener) ExitUpdateStatement(ctx *UpdateStatementContext) {}

// EnterDeleteStatement is called when production deleteStatement is entered.
func (s *BaseMiniQLListener) EnterDeleteStatement(ctx *DeleteStatementContext) {}

// ExitDeleteStatement is called when production deleteStatement is exited.
func (s *BaseMiniQLListener) ExitDeleteStatement(ctx *DeleteStatementContext) {}

// EnterSelectStatement is called when production selectStatement is entered.
func (s *BaseMiniQLListener) EnterSelectStatement(ctx *SelectStatementContext) {}

// ExitSelectStatement is called when production selectStatement is exited.
func (s *BaseMiniQLListener) ExitSelectStatement(ctx *SelectStatementContext) {}

// EnterSelectAll is called when production selectAll is entered.
func (s *BaseMiniQLListener) EnterSelectAll(ctx *SelectAllContext) {}

// ExitSelectAll is called when production selectAll is exited.
func (s *BaseMiniQLListener) ExitSelectAll(ctx *SelectAllContext) {}

// EnterSelectExpr is called when production selectExpr is entered.
func (s *BaseMiniQLListener) EnterSelectExpr(ctx *SelectExprContext) {}

// ExitSelectExpr is called when production selectExpr is exited.
func (s *BaseMiniQLListener) ExitSelectExpr(ctx *SelectExprContext) {}

// EnterTableReference is called when production tableReference is entered.
func (s *BaseMiniQLListener) EnterTableReference(ctx *TableReferenceContext) {}

// ExitTableReference is called when production tableReference is exited.
func (s *BaseMiniQLListener) ExitTableReference(ctx *TableReferenceContext) {}

// EnterTableRefBase is called when production tableRefBase is entered.
func (s *BaseMiniQLListener) EnterTableRefBase(ctx *TableRefBaseContext) {}

// ExitTableRefBase is called when production tableRefBase is exited.
func (s *BaseMiniQLListener) ExitTableRefBase(ctx *TableRefBaseContext) {}

// EnterTableRefSubquery is called when production tableRefSubquery is entered.
func (s *BaseMiniQLListener) EnterTableRefSubquery(ctx *TableRefSubqueryContext) {}

// ExitTableRefSubquery is called when production tableRefSubquery is exited.
func (s *BaseMiniQLListener) ExitTableRefSubquery(ctx *TableRefSubqueryContext) {}

// EnterJoinType is called when production joinType is entered.
func (s *BaseMiniQLListener) EnterJoinType(ctx *JoinTypeContext) {}

// ExitJoinType is called when production joinType is exited.
func (s *BaseMiniQLListener) ExitJoinType(ctx *JoinTypeContext) {}

// EnterPrimaryExpression is called when production primaryExpression is entered.
func (s *BaseMiniQLListener) EnterPrimaryExpression(ctx *PrimaryExpressionContext) {}

// ExitPrimaryExpression is called when production primaryExpression is exited.
func (s *BaseMiniQLListener) ExitPrimaryExpression(ctx *PrimaryExpressionContext) {}

// EnterOrExpression is called when production orExpression is entered.
func (s *BaseMiniQLListener) EnterOrExpression(ctx *OrExpressionContext) {}

// ExitOrExpression is called when production orExpression is exited.
func (s *BaseMiniQLListener) ExitOrExpression(ctx *OrExpressionContext) {}

// EnterAndExpression is called when production andExpression is entered.
func (s *BaseMiniQLListener) EnterAndExpression(ctx *AndExpressionContext) {}

// ExitAndExpression is called when production andExpression is exited.
func (s *BaseMiniQLListener) ExitAndExpression(ctx *AndExpressionContext) {}

// EnterInExpression is called when production inExpression is entered.
func (s *BaseMiniQLListener) EnterInExpression(ctx *InExpressionContext) {}

// ExitInExpression is called when production inExpression is exited.
func (s *BaseMiniQLListener) ExitInExpression(ctx *InExpressionContext) {}

// EnterLikeExpression is called when production likeExpression is entered.
func (s *BaseMiniQLListener) EnterLikeExpression(ctx *LikeExpressionContext) {}

// ExitLikeExpression is called when production likeExpression is exited.
func (s *BaseMiniQLListener) ExitLikeExpression(ctx *LikeExpressionContext) {}

// EnterComparisonExpression is called when production comparisonExpression is entered.
func (s *BaseMiniQLListener) EnterComparisonExpression(ctx *ComparisonExpressionContext) {}

// ExitComparisonExpression is called when production comparisonExpression is exited.
func (s *BaseMiniQLListener) ExitComparisonExpression(ctx *ComparisonExpressionContext) {}

// EnterLiteralExpr is called when production literalExpr is entered.
func (s *BaseMiniQLListener) EnterLiteralExpr(ctx *LiteralExprContext) {}

// ExitLiteralExpr is called when production literalExpr is exited.
func (s *BaseMiniQLListener) ExitLiteralExpr(ctx *LiteralExprContext) {}

// EnterColumnRefExpr is called when production columnRefExpr is entered.
func (s *BaseMiniQLListener) EnterColumnRefExpr(ctx *ColumnRefExprContext) {}

// ExitColumnRefExpr is called when production columnRefExpr is exited.
func (s *BaseMiniQLListener) ExitColumnRefExpr(ctx *ColumnRefExprContext) {}

// EnterFunctionCallExpr is called when production functionCallExpr is entered.
func (s *BaseMiniQLListener) EnterFunctionCallExpr(ctx *FunctionCallExprContext) {}

// ExitFunctionCallExpr is called when production functionCallExpr is exited.
func (s *BaseMiniQLListener) ExitFunctionCallExpr(ctx *FunctionCallExprContext) {}

// EnterParenExpr is called when production parenExpr is entered.
func (s *BaseMiniQLListener) EnterParenExpr(ctx *ParenExprContext) {}

// ExitParenExpr is called when production parenExpr is exited.
func (s *BaseMiniQLListener) ExitParenExpr(ctx *ParenExprContext) {}

// EnterComparisonOperator is called when production comparisonOperator is entered.
func (s *BaseMiniQLListener) EnterComparisonOperator(ctx *ComparisonOperatorContext) {}

// ExitComparisonOperator is called when production comparisonOperator is exited.
func (s *BaseMiniQLListener) ExitComparisonOperator(ctx *ComparisonOperatorContext) {}

// EnterColumnRef is called when production columnRef is entered.
func (s *BaseMiniQLListener) EnterColumnRef(ctx *ColumnRefContext) {}

// ExitColumnRef is called when production columnRef is exited.
func (s *BaseMiniQLListener) ExitColumnRef(ctx *ColumnRefContext) {}

// EnterUpdateAssignment is called when production updateAssignment is entered.
func (s *BaseMiniQLListener) EnterUpdateAssignment(ctx *UpdateAssignmentContext) {}

// ExitUpdateAssignment is called when production updateAssignment is exited.
func (s *BaseMiniQLListener) ExitUpdateAssignment(ctx *UpdateAssignmentContext) {}

// EnterGroupByItem is called when production groupByItem is entered.
func (s *BaseMiniQLListener) EnterGroupByItem(ctx *GroupByItemContext) {}

// ExitGroupByItem is called when production groupByItem is exited.
func (s *BaseMiniQLListener) ExitGroupByItem(ctx *GroupByItemContext) {}

// EnterOrderByItem is called when production orderByItem is entered.
func (s *BaseMiniQLListener) EnterOrderByItem(ctx *OrderByItemContext) {}

// ExitOrderByItem is called when production orderByItem is exited.
func (s *BaseMiniQLListener) ExitOrderByItem(ctx *OrderByItemContext) {}

// EnterFunctionCall is called when production functionCall is entered.
func (s *BaseMiniQLListener) EnterFunctionCall(ctx *FunctionCallContext) {}

// ExitFunctionCall is called when production functionCall is exited.
func (s *BaseMiniQLListener) ExitFunctionCall(ctx *FunctionCallContext) {}

// EnterPartitionMethod is called when production partitionMethod is entered.
func (s *BaseMiniQLListener) EnterPartitionMethod(ctx *PartitionMethodContext) {}

// ExitPartitionMethod is called when production partitionMethod is exited.
func (s *BaseMiniQLListener) ExitPartitionMethod(ctx *PartitionMethodContext) {}

// EnterTransactionStatement is called when production transactionStatement is entered.
func (s *BaseMiniQLListener) EnterTransactionStatement(ctx *TransactionStatementContext) {}

// ExitTransactionStatement is called when production transactionStatement is exited.
func (s *BaseMiniQLListener) ExitTransactionStatement(ctx *TransactionStatementContext) {}

// EnterUseStatement is called when production useStatement is entered.
func (s *BaseMiniQLListener) EnterUseStatement(ctx *UseStatementContext) {}

// ExitUseStatement is called when production useStatement is exited.
func (s *BaseMiniQLListener) ExitUseStatement(ctx *UseStatementContext) {}

// EnterShowDatabases is called when production showDatabases is entered.
func (s *BaseMiniQLListener) EnterShowDatabases(ctx *ShowDatabasesContext) {}

// ExitShowDatabases is called when production showDatabases is exited.
func (s *BaseMiniQLListener) ExitShowDatabases(ctx *ShowDatabasesContext) {}

// EnterShowTables is called when production showTables is entered.
func (s *BaseMiniQLListener) EnterShowTables(ctx *ShowTablesContext) {}

// ExitShowTables is called when production showTables is exited.
func (s *BaseMiniQLListener) ExitShowTables(ctx *ShowTablesContext) {}

// EnterShowIndexes is called when production showIndexes is entered.
func (s *BaseMiniQLListener) EnterShowIndexes(ctx *ShowIndexesContext) {}

// ExitShowIndexes is called when production showIndexes is exited.
func (s *BaseMiniQLListener) ExitShowIndexes(ctx *ShowIndexesContext) {}

// EnterExplainStatement is called when production explainStatement is entered.
func (s *BaseMiniQLListener) EnterExplainStatement(ctx *ExplainStatementContext) {}

// ExitExplainStatement is called when production explainStatement is exited.
func (s *BaseMiniQLListener) ExitExplainStatement(ctx *ExplainStatementContext) {}

// EnterAnalyzeStatement is called when production analyzeStatement is entered.
func (s *BaseMiniQLListener) EnterAnalyzeStatement(ctx *AnalyzeStatementContext) {}

// ExitAnalyzeStatement is called when production analyzeStatement is exited.
func (s *BaseMiniQLListener) ExitAnalyzeStatement(ctx *AnalyzeStatementContext) {}

// EnterColumnList is called when production columnList is entered.
func (s *BaseMiniQLListener) EnterColumnList(ctx *ColumnListContext) {}

// ExitColumnList is called when production columnList is exited.
func (s *BaseMiniQLListener) ExitColumnList(ctx *ColumnListContext) {}

// EnterIdentifierList is called when production identifierList is entered.
func (s *BaseMiniQLListener) EnterIdentifierList(ctx *IdentifierListContext) {}

// ExitIdentifierList is called when production identifierList is exited.
func (s *BaseMiniQLListener) ExitIdentifierList(ctx *IdentifierListContext) {}

// EnterValueList is called when production valueList is entered.
func (s *BaseMiniQLListener) EnterValueList(ctx *ValueListContext) {}

// ExitValueList is called when production valueList is exited.
func (s *BaseMiniQLListener) ExitValueList(ctx *ValueListContext) {}

// EnterTableName is called when production tableName is entered.
func (s *BaseMiniQLListener) EnterTableName(ctx *TableNameContext) {}

// ExitTableName is called when production tableName is exited.
func (s *BaseMiniQLListener) ExitTableName(ctx *TableNameContext) {}

// EnterIdentifier is called when production identifier is entered.
func (s *BaseMiniQLListener) EnterIdentifier(ctx *IdentifierContext) {}

// ExitIdentifier is called when production identifier is exited.
func (s *BaseMiniQLListener) ExitIdentifier(ctx *IdentifierContext) {}

// EnterDataType is called when production dataType is entered.
func (s *BaseMiniQLListener) EnterDataType(ctx *DataTypeContext) {}

// ExitDataType is called when production dataType is exited.
func (s *BaseMiniQLListener) ExitDataType(ctx *DataTypeContext) {}

// EnterLiteral is called when production literal is entered.
func (s *BaseMiniQLListener) EnterLiteral(ctx *LiteralContext) {}

// ExitLiteral is called when production literal is exited.
func (s *BaseMiniQLListener) ExitLiteral(ctx *LiteralContext) {}

package parser

import (
	"fmt"
	"strconv"
	"time"

	"github.com/antlr4-go/antlr/v4"
	"github.com/yyun543/minidb/internal/logger"
	"go.uber.org/zap"
)

// MiniQLVisitorImpl 是 MiniQL 的访问器实现
type MiniQLVisitorImpl struct {
	BaseMiniQLVisitor
}

// Parse 是对外主要接口，传入 SQL 字符串返回 AST 节点
func Parse(sql string) (Node, error) {
	logger.WithComponent("parser").Debug("Starting SQL parsing",
		zap.String("sql", sql),
		zap.Int("sql_length", len(sql)))

	start := time.Now()

	// 创建词法分析器
	lexerStart := time.Now()
	input := antlr.NewInputStream(sql)
	lexer := NewMiniQLLexer(input)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(antlr.NewDiagnosticErrorListener(true))
	logger.WithComponent("parser").Debug("Lexer created successfully",
		zap.Duration("lexer_creation_time", time.Since(lexerStart)))

	// 创建语法分析器
	parserStart := time.Now()
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := NewMiniQLParser(stream)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(antlr.NewDiagnosticErrorListener(true))
	logger.WithComponent("parser").Debug("Parser created successfully",
		zap.Duration("parser_creation_time", time.Since(parserStart)))

	// 生成语法树，并开始访问
	parseTreeStart := time.Now()
	parseTree := parser.Parse()
	logger.WithComponent("parser").Debug("Parse tree generated",
		zap.Duration("parse_tree_generation_time", time.Since(parseTreeStart)))

	visitorStart := time.Now()
	visitor := &MiniQLVisitorImpl{}
	result := visitor.Visit(parseTree)
	if result == nil {
		logger.WithComponent("parser").Error("Failed to parse SQL - visitor returned nil",
			zap.String("sql", sql),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("解析SQL失败")
	}

	logger.WithComponent("parser").Debug("AST visitor completed",
		zap.Duration("visitor_duration", time.Since(visitorStart)))

	node, ok := result.(Node)
	if !ok {
		logger.WithComponent("parser").Error("Parse result type assertion failed",
			zap.String("sql", sql),
			zap.String("result_type", fmt.Sprintf("%T", result)),
			zap.Duration("duration", time.Since(start)))
		return nil, fmt.Errorf("解析结果类型错误")
	}

	totalDuration := time.Since(start)
	logger.WithComponent("parser").Info("SQL parsing completed successfully",
		zap.String("sql", sql),
		zap.String("node_type", fmt.Sprintf("%T", node)),
		zap.Duration("total_parsing_time", totalDuration))

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
	logger.WithComponent("parser").Debug("Visiting SQL statement",
		zap.String("statement_text", ctx.GetText()))

	// 按照语法规则定义的顺序检查
	if ctx.DmlStatement() != nil {
		logger.WithComponent("parser").Debug("Processing DML statement")
		return v.Visit(ctx.DmlStatement())
	}
	if ctx.DqlStatement() != nil {
		logger.WithComponent("parser").Debug("Processing DQL statement")
		return v.Visit(ctx.DqlStatement())
	}
	if ctx.DdlStatement() != nil {
		logger.WithComponent("parser").Debug("Processing DDL statement")
		return v.Visit(ctx.DdlStatement())
	}
	if ctx.DclStatement() != nil {
		logger.WithComponent("parser").Debug("Processing DCL statement")
		return v.Visit(ctx.DclStatement())
	}
	if ctx.UtilityStatement() != nil {
		logger.WithComponent("parser").Debug("Processing utility statement")
		return v.Visit(ctx.UtilityStatement())
	}

	logger.WithComponent("parser").Warn("Unrecognized SQL statement type",
		zap.String("statement_text", ctx.GetText()))
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
	if ctx.DropIndex() != nil {
		return v.Visit(ctx.DropIndex())
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
	if ctx.ShowIndexes() != nil {
		return v.Visit(ctx.ShowIndexes())
	}
	if ctx.ExplainStatement() != nil {
		return v.Visit(ctx.ExplainStatement())
	}
	return nil
}

// VisitCreateDatabase 访问 CREATE DATABASE 语句
func (v *MiniQLVisitorImpl) VisitCreateDatabase(ctx *CreateDatabaseContext) interface{} {
	logger.WithComponent("parser").Debug("Processing CREATE DATABASE statement")

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
			logger.WithComponent("parser").Info("CREATE DATABASE statement parsed successfully",
				zap.String("database_name", stmt.Database))
		}
	} else {
		logger.WithComponent("parser").Error("CREATE DATABASE missing database name")
	}

	return stmt
}

// VisitCreateTable 访问 CREATE TABLE 语句
func (v *MiniQLVisitorImpl) VisitCreateTable(ctx *CreateTableContext) interface{} {
	logger.WithComponent("parser").Debug("Processing CREATE TABLE statement")

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

	logger.WithComponent("parser").Info("CREATE TABLE statement parsed successfully",
		zap.String("table_name", stmt.Table),
		zap.Int("column_count", len(stmt.Columns)),
		zap.Int("constraint_count", len(stmt.Constraints)))

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

// VisitCreateIndex 访问 CREATE INDEX 语句
func (v *MiniQLVisitorImpl) VisitCreateIndex(ctx *CreateIndexContext) interface{} {
	logger.WithComponent("parser").Debug("Processing CREATE INDEX statement")

	// 创建 CreateIndexStmt 节点
	stmt := &CreateIndexStmt{
		BaseNode: BaseNode{nodeType: CreateIndexNode},
	}

	// 检查是否是 UNIQUE INDEX
	stmt.IsUnique = ctx.UNIQUE() != nil

	// 获取索引名
	if ctx.Identifier() != nil {
		if result := v.Visit(ctx.Identifier()); result != nil {
			switch t := result.(type) {
			case string:
				stmt.Name = t
			case *Identifier:
				stmt.Name = t.Value
			}
		}
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

	// 获取索引列列表
	if ctx.IdentifierList() != nil {
		if result := v.Visit(ctx.IdentifierList()); result != nil {
			if cols, ok := result.([]string); ok {
				stmt.Columns = cols
			}
		}
	}

	logger.WithComponent("parser").Info("CREATE INDEX statement parsed successfully",
		zap.String("index_name", stmt.Name),
		zap.String("table_name", stmt.Table),
		zap.Strings("columns", stmt.Columns),
		zap.Bool("is_unique", stmt.IsUnique))

	return stmt
}

// VisitDropIndex 访问 DROP INDEX 语句
func (v *MiniQLVisitorImpl) VisitDropIndex(ctx *DropIndexContext) interface{} {
	logger.WithComponent("parser").Debug("Processing DROP INDEX statement")

	// 创建 DropIndexStmt 节点
	stmt := &DropIndexStmt{
		BaseNode: BaseNode{nodeType: DropIndexNode},
	}

	// 获取索引名
	if ctx.Identifier() != nil {
		stmt.Name = ctx.Identifier().GetText()
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

	logger.WithComponent("parser").Info("DROP INDEX statement parsed successfully",
		zap.String("index_name", stmt.Name),
		zap.String("table_name", stmt.Table))

	return stmt
}

// VisitDropTable 访问 DROP TABLE 语句节点
func (v *MiniQLVisitorImpl) VisitDropTable(ctx *DropTableContext) interface{} {
	// 创建 DROP TABLE 语句节点
	stmt := &DropTableStmt{
		BaseNode: BaseNode{nodeType: DropTableNode},
	}

	// 获取表名
	if ctx.TableName() != nil {
		if tableName := v.Visit(ctx.TableName()); tableName != nil {
			switch t := tableName.(type) {
			case string:
				stmt.Table = t
			case *Identifier:
				stmt.Table = t.Value
			}
		}
	}

	return stmt
}

// VisitDropDatabase 访问 DROP DATABASE 语句
func (v *MiniQLVisitorImpl) VisitDropDatabase(ctx *DropDatabaseContext) interface{} {
	// 创建 DropDatabaseStmt 节点
	stmt := &DropDatabaseStmt{
		BaseNode: BaseNode{nodeType: DropDatabaseNode},
	}

	// 获取数据库名称
	if ctx.Identifier() != nil {
		// 直接获取标识符文本值，避免创建中间对象
		stmt.Database = ctx.Identifier().GetText()
	}

	return stmt
}

// VisitInsertStatement 访问 INSERT 语句节点
func (v *MiniQLVisitorImpl) VisitInsertStatement(ctx *InsertStatementContext) interface{} {
	logger.WithComponent("parser").Debug("Processing INSERT statement")

	// 创建 InsertStmt 节点
	stmt := &InsertStmt{
		BaseNode: BaseNode{nodeType: InsertNode},
	}

	// 获取表名
	if ctx.TableName() != nil {
		if result := v.Visit(ctx.TableName()); result != nil {
			stmt.Table = result.(string)
		}
	}

	// 获取列名列表（可选）
	if ctx.IdentifierList() != nil {
		if result := v.Visit(ctx.IdentifierList()); result != nil {
			stmt.Columns = result.([]string)
		}
	}

	// 获取值列表
	for _, valueList := range ctx.AllValueList() {
		if result := v.Visit(valueList); result != nil {
			// ValueList 返回的是 []Node
			stmt.Values = result.([]Node)
		}
	}

	logger.WithComponent("parser").Info("INSERT statement parsed successfully",
		zap.String("table_name", stmt.Table),
		zap.Int("column_count", len(stmt.Columns)),
		zap.Int("value_count", len(stmt.Values)))

	return stmt
}

// VisitUpdateStatement 访问UPDATE语句节点
func (v *MiniQLVisitorImpl) VisitUpdateStatement(ctx *UpdateStatementContext) interface{} {
	if ctx == nil {
		return nil
	}

	logger.WithComponent("parser").Debug("Processing UPDATE statement")

	// 创建UPDATE语句节点
	stmt := &UpdateStmt{
		BaseNode: BaseNode{nodeType: UpdateNode},
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

	// 处理赋值列表
	if ctx.AllUpdateAssignment() != nil {
		assignments := make([]*UpdateAssignment, 0, len(ctx.AllUpdateAssignment()))
		for _, assignCtx := range ctx.AllUpdateAssignment() {
			if result := v.Visit(assignCtx); result != nil {
				if assignment, ok := result.(*UpdateAssignment); ok {
					assignments = append(assignments, assignment)
				}
			}
		}
		stmt.Assignments = assignments
	}

	// 处理WHERE子句
	if ctx.WHERE() != nil && ctx.Expression() != nil {
		if result := v.Visit(ctx.Expression()); result != nil {
			if expr, ok := result.(Node); ok {
				stmt.Where = &WhereClause{
					BaseNode:  BaseNode{nodeType: WhereNode},
					Condition: expr,
				}
			}
		}
	}

	logger.WithComponent("parser").Info("UPDATE statement parsed successfully",
		zap.String("table_name", stmt.Table),
		zap.Int("assignment_count", len(stmt.Assignments)),
		zap.Bool("has_where", stmt.Where != nil))

	return stmt
}

// VisitDeleteStatement 访问 DELETE 语句节点
func (v *MiniQLVisitorImpl) VisitDeleteStatement(ctx *DeleteStatementContext) interface{} {
	logger.WithComponent("parser").Debug("Processing DELETE statement")

	// 创建 DeleteStmt 节点
	stmt := &DeleteStmt{
		BaseNode: BaseNode{nodeType: DeleteNode},
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

	// 处理 WHERE 子句
	if ctx.WHERE() != nil && ctx.Expression() != nil {
		if result := v.Visit(ctx.Expression()); result != nil {
			if expr, ok := result.(Node); ok {
				stmt.Where = &WhereClause{
					BaseNode:  BaseNode{nodeType: WhereNode},
					Condition: expr,
				}
			}
		}
	}

	logger.WithComponent("parser").Info("DELETE statement parsed successfully",
		zap.String("table_name", stmt.Table),
		zap.Bool("has_where", stmt.Where != nil))

	return stmt
}

// VisitSelectStatement 访问 SELECT 语句节点
func (v *MiniQLVisitorImpl) VisitSelectStatement(ctx *SelectStatementContext) interface{} {
	if ctx == nil {
		return nil
	}

	logger.WithComponent("parser").Debug("Processing SELECT statement",
		zap.String("select_text", ctx.GetText()))

	start := time.Now()

	// 创建 SelectStmt 节点
	stmt := &SelectStmt{
		BaseNode: BaseNode{nodeType: SelectNode},
	}

	// 处理 SELECT 列表
	selectItems := ctx.AllSelectItem()
	for _, itemCtx := range selectItems {
		if result := v.Visit(itemCtx); result != nil {
			switch item := result.(type) {
			case *SelectItem:
				// 如果是 SELECT *
				if item.All {
					stmt.All = true
					// TODO 如果是 SELECT table.*
					continue
				}
				// 普通列
				if colItem, ok := item.Expr.(*ColumnItem); ok {
					stmt.Columns = append(stmt.Columns, colItem)
				}
			case *ColumnItem:
				stmt.Columns = append(stmt.Columns, item)
			}
		}
	}

	// 处理 FROM 子句
	if ctx.TableReference() != nil {
		if result := v.Visit(ctx.TableReference()); result != nil {
			switch t := result.(type) {
			case *TableRef:
				stmt.From = t.Table
				stmt.Joins = t.Joins
				stmt.FromAlias = t.Alias
			}
		}
	}

	// 处理 WHERE 子句
	if ctx.WHERE() != nil && ctx.Expression(0) != nil {
		if result := v.Visit(ctx.Expression(0)); result != nil {
			if expr, ok := result.(Node); ok {
				stmt.Where = &WhereClause{
					BaseNode:  BaseNode{nodeType: WhereNode},
					Condition: expr,
				}
			}
		}
	}

	// 处理 GROUP BY 子句
	if ctx.GROUP() != nil && ctx.AllGroupByItem() != nil {
		for _, item := range ctx.AllGroupByItem() {
			if result := v.Visit(item); result != nil {
				if expr, ok := result.(Node); ok {
					stmt.GroupBy = append(stmt.GroupBy, expr)
				}
			}
		}
	}

	// 处理 HAVING 子句
	if ctx.HAVING() != nil {
		// 获取所有表达式
		expressions := ctx.AllExpression()
		if len(expressions) > 0 {
			// HAVING 表达式总是最后一个表达式
			// 因为它在语法上是最后处理的条件
			havingExpr := expressions[len(expressions)-1]
			if result := v.Visit(havingExpr); result != nil {
				if expr, ok := result.(Node); ok {
					stmt.Having = &HavingClause{
						BaseNode:  BaseNode{nodeType: HavingNode},
						Condition: expr,
					}
				}
			}
		}
	}

	// 处理 ORDER BY 子句
	if ctx.ORDER() != nil && ctx.AllOrderByItem() != nil {
		for _, item := range ctx.AllOrderByItem() {
			if result := v.Visit(item); result != nil {
				if orderItem, ok := result.(*OrderByItem); ok {
					stmt.OrderBy = append(stmt.OrderBy, orderItem)
				}
			}
		}
	}

	// 处理 LIMIT 子句
	if ctx.LIMIT() != nil && ctx.INTEGER_LITERAL() != nil {
		limit, err := strconv.ParseInt(ctx.INTEGER_LITERAL().GetText(), 10, 64)
		if err == nil {
			stmt.Limit = limit
		}
	}

	duration := time.Since(start)
	logger.WithComponent("parser").Info("SELECT statement parsed successfully",
		zap.String("from_table", stmt.From),
		zap.String("from_alias", stmt.FromAlias),
		zap.Int("column_count", len(stmt.Columns)),
		zap.Int("join_count", len(stmt.Joins)),
		zap.Bool("has_where", stmt.Where != nil),
		zap.Bool("has_group_by", len(stmt.GroupBy) > 0),
		zap.Bool("has_having", stmt.Having != nil),
		zap.Bool("has_order_by", len(stmt.OrderBy) > 0),
		zap.Int64("limit_value", stmt.Limit),
		zap.Duration("parse_duration", duration))

	return stmt
}

// VisitSelectAll 访问 SELECT * 语句
func (v *MiniQLVisitorImpl) VisitSelectAll(ctx *SelectAllContext) interface{} {
	item := &SelectItem{
		All: true,
	}

	// 如果有表名限定，设置表名
	if ctx.TableName() != nil {
		if result := v.Visit(ctx.TableName()); result != nil {
			switch t := result.(type) {
			case string:
				item.Table = t
			case *Identifier:
				item.Table = t.Value
			}
		}
	}

	return item
}

// VisitSelectExpr 访问 SELECT 表达式
func (v *MiniQLVisitorImpl) VisitSelectExpr(ctx *SelectExprContext) interface{} {
	// 创建SelectItem节点，初始状态未知类型
	item := &ColumnItem{
		BaseNode: BaseNode{nodeType: SelectItemNode},
		Kind:     ColumnItemUnknown,
	}

	// 处理表达式
	if ctx.Expression() != nil {
		if result := v.Visit(ctx.Expression()); result != nil {
			switch expr := result.(type) {
			case *ColumnRef:
				// 如果是列引用，直接设置表和列名
				item.Column = expr.Column
				if expr.Table != "" {
					item.Table = expr.Table
				}
				item.Kind = ColumnItemColumn
			case *FunctionCall:
				// 如果是函数调用，记录整个函数调用节点
				item.Expr = expr
				item.Kind = ColumnItemFunction
			case *BinaryExpr:
				// 二元表达式
				item.Expr = expr
				item.Kind = ColumnItemExpression
			case *IntegerLiteral, *FloatLiteral, *StringLiteral, *Literal:
				// 如果直接是字面量
				if node, ok := expr.(Node); ok {
					item.Expr = node
				}
				item.Kind = ColumnItemLiteral
			default:
				// 其他类型的表达式节点
				if node, ok := expr.(Node); ok {
					item.Expr = node
					item.Kind = ColumnItemExpression
				}
			}
		}
	}

	// 处理可能的别名
	if ctx.Identifier() != nil {
		if result := v.Visit(ctx.Identifier()); result != nil {
			switch t := result.(type) {
			case string:
				item.Alias = t
			case *Identifier:
				item.Alias = t.Value
			}
		}
	}

	return item
}

// VisitTableRefBase 访问基本表引用
func (v *MiniQLVisitorImpl) VisitTableRefBase(ctx *TableRefBaseContext) interface{} {
	tableRef := &TableRef{
		BaseNode: BaseNode{nodeType: TableRefNode},
	}

	// 处理表名
	if ctx.TableName() != nil {
		if result := v.Visit(ctx.TableName()); result != nil {
			switch t := result.(type) {
			case string:
				tableRef.Table = t
			case *Identifier:
				tableRef.Table = t.Value
			}
		}
	}

	// 处理可选的别名
	if ctx.Identifier() != nil {
		if result := v.Visit(ctx.Identifier()); result != nil {
			switch t := result.(type) {
			case string:
				tableRef.Alias = t
			case *Identifier:
				tableRef.Alias = t.Value
			}
		}
	}

	return tableRef
}

// VisitTableReference 访问表引用节点
func (v *MiniQLVisitorImpl) VisitTableReference(ctx *TableReferenceContext) interface{} {
	if ctx == nil {
		return nil
	}

	// 处理基本表引用（无JOIN）
	if len(ctx.GetChildren()) == 1 {
		return v.Visit(ctx.TableReferenceAtom())
	}

	// 处理带JOIN的表引用
	baseRef := &TableRef{
		BaseNode: BaseNode{nodeType: TableRefNode},
	}

	// 获取左表信息
	if result := v.Visit(ctx.TableReference()); result != nil {
		if leftRef, ok := result.(*TableRef); ok {
			baseRef.Table = leftRef.Table
			baseRef.Alias = leftRef.Alias
			baseRef.Joins = leftRef.Joins
		}
	}

	// 构建JOIN信息
	join := &JoinClause{
		BaseNode: BaseNode{nodeType: JoinNode},
	}

	// 设置JOIN类型
	if ctx.JoinType() != nil {
		if joinType := v.Visit(ctx.JoinType()); joinType != nil {
			join.JoinType = joinType.(string)
		}
	} else {
		join.JoinType = "INNER" // 默认为INNER JOIN
	}

	// 获取右表信息
	if result := v.Visit(ctx.TableReferenceAtom()); result != nil {
		if rightRef, ok := result.(*TableRef); ok {
			join.Right = rightRef
		}
	}

	// 设置JOIN条件
	if ctx.Expression() != nil {
		if result := v.Visit(ctx.Expression()); result != nil {
			if expr, ok := result.(Node); ok {
				join.Condition = expr
			}
		}
	}

	// 将新的JOIN添加到JOIN列表中
	baseRef.Joins = append(baseRef.Joins, join)

	return baseRef
}

// VisitTableRefSubquery 访问子查询表引用
func (v *MiniQLVisitorImpl) VisitTableRefSubquery(ctx *TableRefSubqueryContext) interface{} {
	tableRef := &TableRef{
		BaseNode: BaseNode{nodeType: TableRefNode},
	}

	// 处理子查询
	if ctx.SelectStatement() != nil {
		if result := v.Visit(ctx.SelectStatement()); result != nil {
			if subquery, ok := result.(*SelectStmt); ok {
				tableRef.Subquery = subquery
			}
		}
	}

	// 处理别名（必须有）
	if ctx.Identifier() != nil {
		if result := v.Visit(ctx.Identifier()); result != nil {
			switch t := result.(type) {
			case string:
				tableRef.Alias = t
			case *Identifier:
				tableRef.Alias = t.Value
			}
		}
	}

	return tableRef
}

// VisitJoinType 访问连接类型
func (v *MiniQLVisitorImpl) VisitJoinType(ctx *JoinTypeContext) interface{} {
	if ctx.INNER() != nil {
		return "INNER"
	}
	if ctx.LEFT() != nil {
		return "LEFT"
	}
	if ctx.RIGHT() != nil {
		return "RIGHT"
	}
	if ctx.FULL() != nil {
		return "FULL"
	}
	return "INNER" // 默认为内连接
}

// VisitPrimaryExpression 访问基本表达式
func (v *MiniQLVisitorImpl) VisitPrimaryExpression(ctx *PrimaryExpressionContext) interface{} {
	return v.Visit(ctx.PrimaryExpr())
}

// VisitOrExpression 访问OR表达式
func (v *MiniQLVisitorImpl) VisitOrExpression(ctx *OrExpressionContext) interface{} {
	left := v.Visit(ctx.Expression(0))
	right := v.Visit(ctx.Expression(1))

	return &BinaryExpr{
		BaseNode: BaseNode{nodeType: LogicalExprNode},
		Left:     left.(Node),
		Operator: "OR",
		Right:    right.(Node),
	}
}

// VisitAndExpression 访问AND表达式
func (v *MiniQLVisitorImpl) VisitAndExpression(ctx *AndExpressionContext) interface{} {
	left := v.Visit(ctx.Expression(0))
	right := v.Visit(ctx.Expression(1))

	return &BinaryExpr{
		BaseNode: BaseNode{nodeType: LogicalExprNode},
		Left:     left.(Node),
		Operator: "AND",
		Right:    right.(Node),
	}
}

// VisitInExpression 访问IN表达式
func (v *MiniQLVisitorImpl) VisitInExpression(ctx *InExpressionContext) interface{} {
	left := v.Visit(ctx.Expression())
	values := v.Visit(ctx.ValueList()).([]Node)
	operator := "IN"
	if ctx.NOT() != nil {
		operator = "NOT IN"
	}

	return &InExpr{
		BaseNode: BaseNode{nodeType: InExprNode},
		Left:     left.(Node),
		Operator: operator,
		Values:   values,
	}
}

// VisitLikeExpression 访问LIKE表达式
func (v *MiniQLVisitorImpl) VisitLikeExpression(ctx *LikeExpressionContext) interface{} {
	left := v.Visit(ctx.Expression(0))
	right := v.Visit(ctx.Expression(1))
	operator := "LIKE"
	if ctx.NOT() != nil {
		operator = "NOT LIKE"
	}

	return &BinaryExpr{
		BaseNode: BaseNode{nodeType: LikeExprNode},
		Left:     left.(Node),
		Operator: operator,
		Right:    right.(Node),
	}
}

// VisitComparisonExpression 访问比较表达式
func (v *MiniQLVisitorImpl) VisitComparisonExpression(ctx *ComparisonExpressionContext) interface{} {
	left := v.Visit(ctx.Expression(0))
	right := v.Visit(ctx.Expression(1))
	operator := v.Visit(ctx.ComparisonOperator()).(string)

	return &BinaryExpr{
		BaseNode: BaseNode{nodeType: ComparisonExprNode},
		Left:     left.(Node),
		Operator: operator,
		Right:    right.(Node),
	}
}

// VisitLiteralExpr 访问字面量表达式
func (v *MiniQLVisitorImpl) VisitLiteralExpr(ctx *LiteralExprContext) interface{} {
	if ctx.Literal() != nil {
		return v.Visit(ctx.Literal())
	}
	return nil
}

// VisitColumnRefExpr 访问列引用表达式
func (v *MiniQLVisitorImpl) VisitColumnRefExpr(ctx *ColumnRefExprContext) interface{} {
	if ctx.ColumnRef() != nil {
		return v.Visit(ctx.ColumnRef())
	}
	return nil
}

// VisitFunctionCallExpr 访问函数调用表达式
func (v *MiniQLVisitorImpl) VisitFunctionCallExpr(ctx *FunctionCallExprContext) interface{} {
	if ctx.FunctionCall() != nil {
		return v.Visit(ctx.FunctionCall())
	}
	return nil
}

// VisitParenExpr 访问括号表达式
func (v *MiniQLVisitorImpl) VisitParenExpr(ctx *ParenExprContext) interface{} {
	// 对于括号表达式，我们直接返回内部表达式的结果
	// 因为括号只影响优先级，不需要在AST中体现
	if ctx.Expression() != nil {
		return v.Visit(ctx.Expression())
	}
	return nil
}

// VisitComparisonOperator 访问比较运算符
func (v *MiniQLVisitorImpl) VisitComparisonOperator(ctx *ComparisonOperatorContext) interface{} {
	return ctx.GetText()
}

// VisitColumnRef 访问列引用
func (v *MiniQLVisitorImpl) VisitColumnRef(ctx *ColumnRefContext) interface{} {
	var tableName, columnName string

	// 处理限定列名 (table.column)
	if ctx.GetChildCount() == 3 {
		tableName = ctx.Identifier(0).GetText()
		columnName = ctx.Identifier(1).GetText()
	} else {
		columnName = ctx.Identifier(0).GetText()
	}

	return &ColumnRef{
		BaseNode: BaseNode{nodeType: ColumnRefNode},
		Table:    tableName,
		Column:   columnName,
	}
}

// VisitUpdateAssignment 访问UPDATE赋值表达式节点
func (v *MiniQLVisitorImpl) VisitUpdateAssignment(ctx *UpdateAssignmentContext) interface{} {
	if ctx == nil {
		return nil
	}

	assignment := &UpdateAssignment{
		BaseNode: BaseNode{nodeType: UpdateAssignmentNode},
	}

	// 获取列名
	if ctx.Identifier() != nil {
		if result := v.Visit(ctx.Identifier()); result != nil {
			switch t := result.(type) {
			case string:
				assignment.Column = t
			case *Identifier:
				assignment.Column = t.Value
			}
		}
	}

	// 获取表达式值
	if ctx.Expression() != nil {
		if result := v.Visit(ctx.Expression()); result != nil {
			if expr, ok := result.(Node); ok {
				assignment.Value = expr
			}
		}
	}

	return assignment
}

// VisitGroupByItem 访问 GROUP BY 项
func (v *MiniQLVisitorImpl) VisitGroupByItem(ctx *GroupByItemContext) interface{} {
	if ctx == nil {
		return nil
	}

	// 直接返回表达式节点，不需要额外包装
	// GROUP BY 子句中的表达式通常是列引用或函数调用
	if ctx.Expression() != nil {
		return v.Visit(ctx.Expression())
	}

	return nil
}

// VisitOrderByItem 访问 ORDER BY 项
func (v *MiniQLVisitorImpl) VisitOrderByItem(ctx *OrderByItemContext) interface{} {
	if ctx == nil {
		return nil
	}

	// 创建 OrderByItem 节点
	orderByItem := &OrderByItem{
		BaseNode: BaseNode{nodeType: OrderByItemNode},
	}

	// 处理表达式
	if ctx.Expression() != nil {
		if result := v.Visit(ctx.Expression()); result != nil {
			if expr, ok := result.(Node); ok {
				orderByItem.Expr = expr
			}
		}
	}

	// 处理排序方向
	if ctx.DESC() != nil {
		orderByItem.Direction = "DESC"
	} else {
		// ASC 是默认值，即使没有显式指定 ASC 也使用升序
		orderByItem.Direction = "ASC"
	}

	return orderByItem
}

// VisitFunctionCall 访问函数调用
func (v *MiniQLVisitorImpl) VisitFunctionCall(ctx *FunctionCallContext) interface{} {
	funcName := ctx.Identifier().GetText()
	var args []Node

	// 处理参数
	if ctx.ASTERISK() != nil {
		args = append(args, &Asterisk{BaseNode: BaseNode{nodeType: AsteriskNode}})
	} else {
		for _, expr := range ctx.AllExpression() {
			if arg := v.Visit(expr); arg != nil {
				args = append(args, arg.(Node))
			}
		}
	}

	return &FunctionCall{
		BaseNode: BaseNode{nodeType: FunctionCallNode},
		Name:     funcName,
		Args:     args,
	}
}

// VisitPartitionMethod 访问分区方法节点
func (v *MiniQLVisitorImpl) VisitPartitionMethod(ctx *PartitionMethodContext) interface{} {
	if ctx == nil {
		return nil
	}

	// 创建分区方法节点
	partition := &PartitionMethod{
		BaseNode: BaseNode{nodeType: PartitionMethodNode},
	}

	// 处理分区类型
	if ctx.HASH() != nil {
		partition.Type = "HASH"
	} else if ctx.RANGE() != nil {
		partition.Type = "RANGE"
	}

	// 处理分区键列
	if ctx.IdentifierList() != nil {
		if result := v.Visit(ctx.IdentifierList()); result != nil {
			if columns, ok := result.([]string); ok {
				partition.Columns = columns
			}
		}
	}

	// TODO 处理分区数量（仅用于 HASH 分区）
	/*if ctx.INTEGER_LITERAL() != nil {
		if num, err := strconv.ParseInt(ctx.INTEGER_LITERAL().GetText(), 10, 64); err == nil {
			partition.PartitionNum = int(num)
		}
	}*/

	return partition
}

// VisitTransactionStatement 访问事务语句节点
func (v *MiniQLVisitorImpl) VisitTransactionStatement(ctx *TransactionStatementContext) interface{} {
	if ctx == nil {
		return nil
	}

	// 创建事务语句节点
	stmt := &TransactionStmt{
		BaseNode: BaseNode{nodeType: TransactionNode},
	}

	// 根据具体的事务命令设置类型
	switch {
	case ctx.START() != nil && ctx.TRANSACTION() != nil:
		// 必须同时存在 START 和 TRANSACTION 关键字
		stmt.TxType = "BEGIN"
	case ctx.COMMIT() != nil:
		stmt.TxType = "COMMIT"
	case ctx.ROLLBACK() != nil:
		stmt.TxType = "ROLLBACK"
	default:
		return nil
	}

	return stmt
}

// VisitUseStatement 访问 USE 语句节点
func (v *MiniQLVisitorImpl) VisitUseStatement(ctx *UseStatementContext) interface{} {
	if ctx == nil {
		return nil
	}

	logger.WithComponent("parser").Debug("Processing USE statement")

	// 获取数据库名
	var dbName string
	if ctx.Identifier() != nil {
		if result := v.Visit(ctx.Identifier()); result != nil {
			dbName = result.(string)
		}
	}

	logger.WithComponent("parser").Info("USE statement parsed successfully",
		zap.String("database_name", dbName))

	// 返回 UseStmt 节点
	return &UseStmt{
		BaseNode: BaseNode{nodeType: UseNode},
		Database: dbName,
	}
}

// VisitShowDatabases 访问 SHOW DATABASES 语句
func (v *MiniQLVisitorImpl) VisitShowDatabases(ctx *ShowDatabasesContext) interface{} {
	logger.WithComponent("parser").Debug("Processing SHOW DATABASES statement")

	// 直接返回 ShowDatabasesStmt 节点，无需任何中间对象
	stmt := &ShowDatabasesStmt{
		BaseNode: BaseNode{nodeType: ShowDatabasesNode},
	}

	logger.WithComponent("parser").Info("SHOW DATABASES statement parsed successfully")

	return stmt
}

// VisitShowTables 访问 SHOW TABLES 语句节点
func (v *MiniQLVisitorImpl) VisitShowTables(ctx *ShowTablesContext) interface{} {
	if ctx == nil {
		return nil
	}

	logger.WithComponent("parser").Debug("Processing SHOW TABLES statement")

	// 返回 ShowTablesStmt 节点
	stmt := &ShowTablesStmt{
		BaseNode: BaseNode{nodeType: ShowTablesNode},
		Database: "",
	}

	logger.WithComponent("parser").Info("SHOW TABLES statement parsed successfully")

	return stmt
}

// VisitShowIndexes 访问 SHOW INDEXES 语句节点
func (v *MiniQLVisitorImpl) VisitShowIndexes(ctx *ShowIndexesContext) interface{} {
	if ctx == nil {
		return nil
	}

	logger.WithComponent("parser").Debug("Processing SHOW INDEXES statement")

	// 创建 ShowIndexesStmt 节点
	stmt := &ShowIndexesStmt{
		BaseNode: BaseNode{nodeType: ShowIndexesNode},
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

	logger.WithComponent("parser").Info("SHOW INDEXES statement parsed successfully",
		zap.String("table_name", stmt.Table))

	return stmt
}

// VisitExplainStatement 访问 EXPLAIN 语句节点
func (v *MiniQLVisitorImpl) VisitExplainStatement(ctx *ExplainStatementContext) interface{} {
	if ctx == nil {
		return nil
	}

	// 创建 EXPLAIN 语句节点
	stmt := &ExplainStmt{
		BaseNode: BaseNode{nodeType: ExplainNode},
	}

	// 获取要解释的查询语句
	// EXPLAIN 后面只能跟 SELECT 语句
	if ctx.SelectStatement() != nil {
		if result := v.Visit(ctx.SelectStatement()); result != nil {
			if query, ok := result.(Node); ok {
				stmt.Query = query
			}
		}
	}

	return stmt
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

// VisitValueList 访问值列表节点
func (v *MiniQLVisitorImpl) VisitValueList(ctx *ValueListContext) interface{} {
	if ctx == nil {
		return nil
	}

	// 直接构建值节点列表，无需中间对象
	values := make([]Node, 0, len(ctx.AllLiteral()))

	// 遍历所有字面量
	for _, literalCtx := range ctx.AllLiteral() {
		if result := v.Visit(literalCtx); result != nil {
			// 确保返回的是 Node 类型
			if node, ok := result.(Node); ok {
				values = append(values, node)
			}
		}
	}

	return values
}

// VisitTableName 访问表名节点，支持 database.table 格式
func (v *MiniQLVisitorImpl) VisitTableName(ctx *TableNameContext) interface{} {
	if ctx == nil {
		return nil
	}
	// 获取所有标识符（可能是 [table] 或 [database, table]）
	identifiers := ctx.AllIdentifier()
	if len(identifiers) == 0 {
		return nil
	}

	if len(identifiers) == 1 {
		// 单个标识符：table
		return v.Visit(identifiers[0])
	}

	// 两个标识符：database.table
	database := v.Visit(identifiers[0]).(string)
	table := v.Visit(identifiers[1]).(string)
	return database + "." + table
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

// VisitLiteral 访问字面量节点
func (v *MiniQLVisitorImpl) VisitLiteral(ctx *LiteralContext) interface{} {
	if ctx == nil {
		return nil
	}

	// 根据字面量类型创建对应的节点
	switch {
	case ctx.INTEGER_LITERAL() != nil:
		text := ctx.INTEGER_LITERAL().GetText()
		val, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			logger.WithComponent("parser").Error("Failed to parse integer literal",
				zap.String("text", text),
				zap.Error(err))
			return nil
		}
		return &IntegerLiteral{
			BaseNode: BaseNode{nodeType: IntegerLiteralNode},
			Value:    val,
		}
	case ctx.FLOAT_LITERAL() != nil:
		text := ctx.FLOAT_LITERAL().GetText()
		val, err := strconv.ParseFloat(text, 64)
		if err != nil {
			logger.WithComponent("parser").Error("Failed to parse float literal",
				zap.String("text", text),
				zap.Error(err))
			return nil
		}
		return &FloatLiteral{
			BaseNode: BaseNode{nodeType: FloatLiteralNode},
			Value:    val,
		}
	case ctx.STRING_LITERAL() != nil:
		// 去除字符串两端的引号
		text := ctx.STRING_LITERAL().GetText()
		if len(text) < 2 {
			logger.WithComponent("parser").Error("Invalid string literal format",
				zap.String("text", text))
			return nil
		}
		value := text[1 : len(text)-1]
		return &StringLiteral{
			BaseNode: BaseNode{nodeType: StringLiteralNode},
			Value:    value,
		}
	default:
		logger.WithComponent("parser").Warn("Unrecognized literal type",
			zap.String("context", ctx.GetText()))
		return nil
	}
}

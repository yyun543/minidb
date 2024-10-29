package parser

// Visitor 定义了访问AST节点的接口
type Visitor interface {
	// 语句访问方法
	VisitCreateTable(*CreateTableStmt) interface{}
	VisitDropTable(*DropTableStmt) interface{}
	VisitShowTables(*ShowTablesStmt) interface{}
	VisitSelect(*SelectStmt) interface{}
	VisitInsert(*InsertStmt) interface{}
	VisitUpdate(*UpdateStmt) interface{}
	VisitDelete(*DeleteStmt) interface{}

	// 表达式访问方法
	VisitIdentifier(*Identifier) interface{}
	VisitLiteral(*Literal) interface{}
	VisitComparisonExpr(*ComparisonExpr) interface{}
	VisitBinaryExpr(*BinaryExpr) interface{}
	VisitFunctionExpr(*FunctionExpr) interface{}
}

// Visitable 定义了可访问的节点接口
type Visitable interface {
	Accept(Visitor) interface{}
}

// Node接口的Accept方法实现
func (s *CreateTableStmt) Accept(v Visitor) interface{} {
	return v.VisitCreateTable(s)
}

func (s *DropTableStmt) Accept(v Visitor) interface{} {
	return v.VisitDropTable(s)
}

func (s *ShowTablesStmt) Accept(v Visitor) interface{} {
	return v.VisitShowTables(s)
}

func (s *SelectStmt) Accept(v Visitor) interface{} {
	return v.VisitSelect(s)
}

func (s *InsertStmt) Accept(v Visitor) interface{} {
	return v.VisitInsert(s)
}

func (s *UpdateStmt) Accept(v Visitor) interface{} {
	return v.VisitUpdate(s)
}

func (s *DeleteStmt) Accept(v Visitor) interface{} {
	return v.VisitDelete(s)
}

func (e *Identifier) Accept(v Visitor) interface{} {
	return v.VisitIdentifier(e)
}

func (e *Literal) Accept(v Visitor) interface{} {
	return v.VisitLiteral(e)
}

func (e *ComparisonExpr) Accept(v Visitor) interface{} {
	return v.VisitComparisonExpr(e)
}

func (e *BinaryExpr) Accept(v Visitor) interface{} {
	return v.VisitBinaryExpr(e)
}

func (e *FunctionExpr) Accept(v Visitor) interface{} {
	return v.VisitFunctionExpr(e)
}

// BaseVisitor 提供Visitor接口的默认实现
type BaseVisitor struct{}

func (v *BaseVisitor) VisitCreateTable(*CreateTableStmt) interface{}   { return nil }
func (v *BaseVisitor) VisitDropTable(*DropTableStmt) interface{}       { return nil }
func (v *BaseVisitor) VisitShowTables(*ShowTablesStmt) interface{}     { return nil }
func (v *BaseVisitor) VisitSelect(*SelectStmt) interface{}             { return nil }
func (v *BaseVisitor) VisitInsert(*InsertStmt) interface{}             { return nil }
func (v *BaseVisitor) VisitUpdate(*UpdateStmt) interface{}             { return nil }
func (v *BaseVisitor) VisitDelete(*DeleteStmt) interface{}             { return nil }
func (v *BaseVisitor) VisitIdentifier(*Identifier) interface{}         { return nil }
func (v *BaseVisitor) VisitLiteral(*Literal) interface{}               { return nil }
func (v *BaseVisitor) VisitComparisonExpr(*ComparisonExpr) interface{} { return nil }
func (v *BaseVisitor) VisitBinaryExpr(*BinaryExpr) interface{}         { return nil }
func (v *BaseVisitor) VisitFunctionExpr(*FunctionExpr) interface{}     { return nil }

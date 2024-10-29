package executor

import (
	"fmt"

	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/storage"
)

type Executor struct {
	storage *storage.Engine
	visitor *QueryExecutor
}

func NewExecutor(storage *storage.Engine) *Executor {
	return &Executor{
		storage: storage,
		visitor: NewQueryExecutor(storage),
	}
}

func (e *Executor) Execute(sql string) (string, error) {
	// Parse SQL
	p := parser.NewParser(sql)
	stmt, err := p.Parse()
	if err != nil {
		return "", fmt.Errorf("parse error: %v", err)
	}

	// Execute statement
	result := e.executeStatement(stmt)

	// Format result
	switch v := result.(type) {
	case string:
		return v, nil
	case error:
		return "", v
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func (e *Executor) executeStatement(stmt parser.Statement) interface{} {
	switch s := stmt.(type) {
	case *parser.CreateTableStmt:
		return e.visitor.VisitCreateTable(s)
	case *parser.SelectStmt:
		return e.visitor.VisitSelect(s)
	case *parser.InsertStmt:
		return e.visitor.VisitInsert(s)
	case *parser.UpdateStmt:
		return e.visitor.VisitUpdate(s)
	case *parser.DeleteStmt:
		return e.visitor.VisitDelete(s)
	default:
		return fmt.Errorf("unknown statement type: %T", stmt)
	}
}

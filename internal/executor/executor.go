package executor

import (
	"fmt"

	"github.com/yyun543/minidb/internal/parser"
	"github.com/yyun543/minidb/internal/storage"
)

type Executor struct {
	storage *storage.Engine
}

func NewExecutor(storage *storage.Engine) *Executor {
	return &Executor{storage: storage}
}

func (e *Executor) Execute(query *parser.Query) (string, error) {
	switch query.Type {
	case parser.SELECT:
		return e.executeSelect(query)
	case parser.INSERT:
		return e.executeInsert(query)
	case parser.UPDATE:
		return e.executeUpdate(query)
	case parser.DELETE:
		return e.executeDelete(query)
	default:
		return "", fmt.Errorf("unsupported query type")
	}
}

func (e *Executor) executeSelect(query *parser.Query) (string, error) {
	rows, err := e.storage.Select(query.Table, query.Fields)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Selected %d rows", len(rows)), nil
}

func (e *Executor) executeInsert(query *parser.Query) (string, error) {
	err := e.storage.Insert(query.Table, query.Values)
	if err != nil {
		return "", err
	}
	return "Inserted 1 row", nil
}

func (e *Executor) executeUpdate(query *parser.Query) (string, error) {
	count, err := e.storage.Update(query.Table, query.Fields[0], query.Values[0], query.Where)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Updated %d rows", count), nil
}

func (e *Executor) executeDelete(query *parser.Query) (string, error) {
	count, err := e.storage.Delete(query.Table, query.Where)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Deleted %d rows", count), nil
}


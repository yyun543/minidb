package parser

import (
	"errors"
	"strings"
)

type QueryType int

const (
	// DDL Operations
	CREATE QueryType = iota
	DROP
	SHOW
	// DML Operations
	SELECT
	INSERT
	UPDATE
	DELETE
)

type Query struct {
	Type    QueryType
	Table   string
	Fields  []string // For SELECT and CREATE TABLE (column names)
	Values  []string // For INSERT and column types in CREATE TABLE
	Where   string   // For SELECT, UPDATE, DELETE conditions
	Command string   // For special commands like SHOW TABLES
}

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(query string) (*Query, error) {
	// Remove trailing semicolon and clean whitespace
	query = strings.TrimSpace(strings.TrimSuffix(query, ";"))

	parts := strings.Fields(query)
	if len(parts) == 0 {
		return nil, errors.New("empty query")
	}

	// Only convert command keywords to uppercase for matching
	switch strings.ToUpper(parts[0]) {
	// DDL Operations
	case "CREATE":
		return p.parseCreate(parts)
	case "DROP":
		return p.parseDropTable(parts)
	case "SHOW":
		return p.parseShow(parts)
	// DML Operations
	case "SELECT":
		return p.parseSelect(parts)
	case "INSERT":
		return p.parseInsert(parts)
	case "UPDATE":
		return p.parseUpdate(parts)
	case "DELETE":
		return p.parseDelete(parts)
	default:
		return nil, errors.New("unsupported query type")
	}
}

func (p *Parser) parseSelect(parts []string) (*Query, error) {
	if len(parts) < 4 || strings.ToUpper(parts[2]) != "FROM" {
		return nil, errors.New("invalid SELECT query")
	}

	// Extract fields between SELECT and FROM
	fieldsStr := strings.Join(parts[1:], " ")
	fromIndex := strings.Index(strings.ToUpper(fieldsStr), "FROM")
	if fromIndex == -1 {
		return nil, errors.New("invalid SELECT query: missing FROM clause")
	}
	fieldsStr = fieldsStr[:fromIndex]

	// Split fields by comma and clean
	var fields []string
	if fieldsStr == "*" {
		fields = []string{"*"}
	} else {
		fields = strings.Split(fieldsStr, ",")
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
			if fields[i] == "" {
				return nil, errors.New("empty field name in SELECT query")
			}
		}
	}

	// Get table name and WHERE clause
	remainingStr := strings.TrimSpace(strings.Join(parts[3:], " "))
	tableParts := strings.Fields(remainingStr)
	if len(tableParts) == 0 {
		return nil, errors.New("missing table name")
	}

	table := tableParts[0]
	var where string
	if len(tableParts) > 1 && strings.ToUpper(tableParts[1]) == "WHERE" {
		where = strings.Join(tableParts[2:], " ")
	}

	return &Query{
		Type:   SELECT,
		Table:  table,
		Fields: fields,
		Where:  where,
	}, nil
}

func (p *Parser) parseInsert(parts []string) (*Query, error) {
	if len(parts) < 5 || strings.ToUpper(parts[1]) != "INTO" || strings.ToUpper(parts[3]) != "VALUES" {
		return nil, errors.New("invalid INSERT query")
	}

	// Extract values within parentheses
	valuesStr := strings.Join(parts[4:], " ")
	valuesStr = strings.Trim(valuesStr, "()")

	// Split values and clean
	values := strings.Split(valuesStr, ",")
	// Remove possible quotes
	for i := range values {
		values[i] = strings.Trim(values[i], "'\"")
	}

	return &Query{
		Type:   INSERT,
		Table:  parts[2],
		Values: values,
	}, nil
}

func (p *Parser) parseUpdate(parts []string) (*Query, error) {
	// Check basic syntax
	if len(parts) < 4 {
		return nil, errors.New("invalid UPDATE query")
	}

	// Reconstruct full query string for better handling of equals sign
	fullQuery := strings.Join(parts, " ")

	// Split main parts
	sections := strings.SplitN(fullQuery, "SET", 2)
	if len(sections) != 2 {
		return nil, errors.New("UPDATE query must have SET clause")
	}

	// Get table name
	table := strings.TrimSpace(strings.TrimPrefix(sections[0], "UPDATE"))

	// Handle SET and WHERE parts
	setAndWhere := strings.SplitN(sections[1], "WHERE", 2)
	if len(setAndWhere) < 2 {
		return nil, errors.New("UPDATE query must have WHERE clause")
	}

	// Parse SET clause
	setClause := strings.TrimSpace(setAndWhere[0])
	setParts := strings.Split(setClause, "=")
	if len(setParts) != 2 {
		return nil, errors.New("invalid SET clause in UPDATE query")
	}

	field := strings.TrimSpace(setParts[0])
	value := strings.TrimSpace(setParts[1])
	// Remove possible quotes
	value = strings.Trim(value, "'\"")

	// Get WHERE condition
	whereCondition := strings.TrimSpace(setAndWhere[1])

	return &Query{
		Type:   UPDATE,
		Table:  table,
		Fields: []string{field},
		Values: []string{value},
		Where:  whereCondition,
	}, nil
}

func (p *Parser) parseDelete(parts []string) (*Query, error) {
	// Check basic syntax
	if len(parts) < 4 || strings.ToUpper(parts[1]) != "FROM" {
		return nil, errors.New("invalid DELETE query: must be in format 'DELETE FROM table WHERE condition'")
	}

	// Reconstruct full query string
	fullQuery := strings.Join(parts, " ")

	// Split main parts
	sections := strings.SplitN(fullQuery, "WHERE", 2)
	if len(sections) != 2 {
		return nil, errors.New("DELETE query must have WHERE clause")
	}

	// Get table name
	tableParts := strings.Fields(sections[0])
	if len(tableParts) != 3 { // DELETE FROM table
		return nil, errors.New("invalid DELETE query format")
	}
	table := strings.TrimSpace(tableParts[2])

	// Get WHERE condition
	whereCondition := strings.TrimSpace(sections[1])

	return &Query{
		Type:  DELETE,
		Table: table,
		Where: whereCondition,
	}, nil
}

// parseCreate handles CREATE TABLE statements
func (p *Parser) parseCreate(parts []string) (*Query, error) {
	if len(parts) < 4 || strings.ToUpper(parts[1]) != "TABLE" {
		return nil, errors.New("invalid CREATE TABLE syntax")
	}

	tableName := parts[2]

	// Extract column definitions within parentheses
	columnsStr := strings.Join(parts[3:], " ")
	columnsStr = strings.Trim(columnsStr, "()")

	// Split column definitions
	columnDefs := strings.Split(columnsStr, ",")
	fields := make([]string, 0)
	types := make([]string, 0)

	for _, def := range columnDefs {
		parts := strings.Fields(strings.TrimSpace(def))
		if len(parts) < 2 {
			return nil, errors.New("invalid column definition")
		}
		fields = append(fields, parts[0])
		types = append(types, parts[1])
	}

	return &Query{
		Type:   CREATE,
		Table:  tableName,
		Fields: fields,
		Values: types,
	}, nil
}

// parseDropTable handles DROP TABLE statements
func (p *Parser) parseDropTable(parts []string) (*Query, error) {
	if len(parts) != 3 || strings.ToUpper(parts[1]) != "TABLE" {
		return nil, errors.New("invalid DROP TABLE syntax")
	}

	return &Query{
		Type:  DROP,
		Table: parts[2],
	}, nil
}

// parseShow handles SHOW TABLES statement
func (p *Parser) parseShow(parts []string) (*Query, error) {
	if len(parts) != 2 || strings.ToUpper(parts[1]) != "TABLES" {
		return nil, errors.New("invalid SHOW command")
	}

	return &Query{
		Type:    SHOW,
		Command: "TABLES",
	}, nil
}

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
	// 重建完整查询字符串
	query := strings.Join(parts, " ")

	// 分割 SELECT 和 FROM 部分
	sections := strings.SplitN(strings.ToUpper(query), "FROM", 2)
	if len(sections) != 2 {
		return nil, errors.New("invalid SELECT query: missing FROM clause")
	}

	// 处理字段部分
	fieldsStr := strings.TrimSpace(strings.TrimPrefix(sections[0], "SELECT"))
	rawFields := strings.Split(fieldsStr, ",")
	fields := make([]string, len(rawFields))
	for i := range rawFields {
		fields[i] = strings.TrimSpace(rawFields[i])
	}

	// 处理表名和 WHERE 部分
	fromParts := strings.Fields(sections[1])
	if len(fromParts) == 0 {
		return nil, errors.New("invalid SELECT query: missing table name")
	}

	table := fromParts[0]
	var where string
	if len(fromParts) > 1 && strings.ToUpper(fromParts[1]) == "WHERE" {
		where = strings.Join(fromParts[2:], " ")
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
	// 重建完整查询字符串
	query := strings.Join(parts, " ")

	// 分割 UPDATE 和 SET 部分
	sections := strings.SplitN(strings.ToUpper(query), "SET", 2)
	if len(sections) != 2 {
		return nil, errors.New("invalid UPDATE query: missing SET clause")
	}

	// 获取表名
	table := strings.TrimSpace(strings.TrimPrefix(sections[0], "UPDATE"))

	// 处理 SET 和 WHERE 部分
	setAndWhere := strings.SplitN(sections[1], "WHERE", 2)

	// 解析 SET 子句
	setClause := strings.TrimSpace(setAndWhere[0])
	setParts := strings.Split(setClause, "=")
	if len(setParts) != 2 {
		return nil, errors.New("invalid SET clause")
	}

	field := strings.TrimSpace(setParts[0])
	value := strings.TrimSpace(setParts[1])

	var where string
	if len(setAndWhere) > 1 {
		where = strings.TrimSpace(setAndWhere[1])
	}

	return &Query{
		Type:   UPDATE,
		Table:  table,
		Fields: []string{field},
		Values: []string{value},
		Where:  where,
	}, nil
}

func (p *Parser) parseDelete(parts []string) (*Query, error) {
	if len(parts) < 4 || strings.ToUpper(parts[1]) != "FROM" {
		return nil, errors.New("invalid DELETE query")
	}

	table := parts[2]
	var where string

	// Fix: Properly handle WHERE clause
	if len(parts) > 3 {
		if strings.ToUpper(parts[3]) == "WHERE" && len(parts) > 4 {
			where = strings.Join(parts[4:], " ")
		}
	}

	return &Query{
		Type:  DELETE,
		Table: table,
		Where: where,
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

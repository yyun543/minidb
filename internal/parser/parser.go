package parser

import (
	"errors"
	"strings"
)

type QueryType int

const (
	SELECT QueryType = iota
	INSERT
	UPDATE
	DELETE
)

type Query struct {
	Type   QueryType
	Table  string
	Fields []string
	Values []string
	Where  string
}

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(query string) (*Query, error) {
	parts := strings.Fields(strings.ToUpper(query))
	if len(parts) == 0 {
		return nil, errors.New("empty query")
	}

	switch parts[0] {
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
	// 简化实现，仅支持 "SELECT * FROM table_name"
	if len(parts) != 4 || parts[1] != "*" || parts[2] != "FROM" {
		return nil, errors.New("invalid SELECT query")
	}
	return &Query{
		Type:   SELECT,
		Table:  parts[3],
		Fields: []string{"*"},
	}, nil
}

func (p *Parser) parseInsert(parts []string) (*Query, error) {
	// 简化实现，仅支持 "INSERT INTO table_name VALUES (value1, value2, ...)"
	if len(parts) < 5 || parts[1] != "INTO" || parts[3] != "VALUES" {
		return nil, errors.New("invalid INSERT query")
	}
	values := strings.Join(parts[4:], " ")
	values = strings.Trim(values, "()")
	return &Query{
		Type:   INSERT,
		Table:  parts[2],
		Values: strings.Split(values, ","),
	}, nil
}

func (p *Parser) parseUpdate(parts []string) (*Query, error) {
	// 修改 UPDATE 语句的解析逻辑
	if len(parts) < 7 || parts[2] != "SET" {
		return nil, errors.New("invalid UPDATE query")
	}

	// 查找 WHERE 子句的位置
	whereIndex := -1
	for i, part := range parts {
		if part == "WHERE" {
			whereIndex = i
			break
		}
	}
	
	if whereIndex == -1 {
		return nil, errors.New("UPDATE query must have WHERE clause")
	}

	// 解析 SET 子句
	setClause := strings.Split(parts[3], "=")
	if len(setClause) != 2 {
		return nil, errors.New("invalid SET clause in UPDATE query")
	}

	// 解析 WHERE 子句
	whereCondition := strings.Join(parts[whereIndex+1:], " ")

	return &Query{
		Type:   UPDATE,
		Table:  parts[1],
		Fields: []string{setClause[0]},
		Values: []string{setClause[1]},
		Where:  whereCondition,
	}, nil
}

func (p *Parser) parseDelete(parts []string) (*Query, error) {
	// 简化实现，仅支持 "DELETE FROM table_name WHERE condition"
	if len(parts) != 5 || parts[1] != "FROM" || parts[3] != "WHERE" {
		return nil, errors.New("invalid DELETE query")
	}
	return &Query{
		Type:  DELETE,
		Table: parts[2],
		Where: parts[4],
	}, nil
}

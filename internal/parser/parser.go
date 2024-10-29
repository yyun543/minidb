package parser

import (
	"fmt"
	"strconv"
)

type Parser struct {
	lexer        *Lexer
	currentToken Token
	peekToken    Token
	errors       []string
}

func NewParser(input string) *Parser {
	p := &Parser{
		lexer:  NewLexer(input),
		errors: []string{},
	}
	// 读取两个token以初始化current和peek
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("expected next token to be %v, got %v instead at line %d, column %d",
		t, p.peekToken.Type, p.peekToken.Line, p.peekToken.Column)
	p.errors = append(p.errors, msg)
}

func (p *Parser) Parse() (Statement, error) {
	switch p.currentToken.Type {
	case SELECT:
		return p.parseSelect()
	case INSERT:
		return p.parseInsert()
	case UPDATE:
		return p.parseUpdate()
	case DELETE:
		return p.parseDelete()
	case CREATE:
		return p.parseCreateTable()
	case DROP:
		return p.parseDropTable()
	case SHOW:
		return p.parseShowTables()
	default:
		return nil, fmt.Errorf("unexpected token %v at line %d, column %d",
			p.currentToken.Type, p.currentToken.Line, p.currentToken.Column)
	}
}

func (p *Parser) parseSelect() (*SelectStmt, error) {
	stmt := &SelectStmt{BaseNode: BaseNode{nodeType: SELECT}}

	// Parse fields
	if !p.expectPeek(MULTIPLY) && !p.expectPeek(IDENT) {
		return nil, fmt.Errorf("expected field name or * after SELECT")
	}

	stmt.Fields = make([]Expression, 0)
	for {
		var expr Expression
		if p.currentToken.Type == MULTIPLY {
			expr = &Identifier{
				BaseNode: BaseNode{nodeType: IDENTIFIER},
				Name:     "*",
			}
		} else {
			expr = &Identifier{
				BaseNode: BaseNode{nodeType: IDENTIFIER},
				Name:     p.currentToken.Literal,
			}

			// Check for alias
			if p.peekToken.Type == AS {
				p.nextToken() // consume AS
				if !p.expectPeek(IDENT) {
					return nil, fmt.Errorf("expected identifier after AS")
				}
				expr = &Identifier{
					BaseNode: BaseNode{nodeType: IDENTIFIER},
					Name:     fmt.Sprintf("%s AS %s", expr.(*Identifier).Name, p.currentToken.Literal),
				}
			}
		}
		stmt.Fields = append(stmt.Fields, expr)

		if p.peekToken.Type != COMMA {
			break
		}
		p.nextToken() // consume comma
		p.nextToken() // move to next field
	}

	// Parse FROM clause
	if !p.expectPeek(FROM) {
		return nil, fmt.Errorf("expected FROM after SELECT fields")
	}
	if !p.expectPeek(IDENT) {
		return nil, fmt.Errorf("expected table name after FROM")
	}
	stmt.From = p.currentToken.Literal

	// Parse optional JOIN
	if p.peekToken.Type == JOIN || p.peekToken.Type == LEFT || p.peekToken.Type == RIGHT {
		p.nextToken()
		if err := p.parseJoin(stmt); err != nil {
			return nil, err
		}
	}

	// Parse optional WHERE
	if p.peekToken.Type == WHERE {
		p.nextToken() // consume WHERE
		p.nextToken() // move to first token of expression
		where, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		stmt.Where = where
	}

	// Parse optional GROUP BY
	if p.peekToken.Type == GROUP {
		p.nextToken() // consume GROUP
		if !p.expectPeek(BY) {
			return nil, fmt.Errorf("expected BY after GROUP")
		}

		stmt.GroupBy = make([]string, 0)
		for {
			if !p.expectPeek(IDENT) {
				return nil, fmt.Errorf("expected identifier in GROUP BY")
			}
			stmt.GroupBy = append(stmt.GroupBy, p.currentToken.Literal)

			if p.peekToken.Type != COMMA {
				break
			}
			p.nextToken() // consume comma
		}

		// Parse optional HAVING
		if p.peekToken.Type == HAVING {
			p.nextToken() // consume HAVING
			p.nextToken() // move to first token of expression
			having, err := p.parseExpression(LOWEST)
			if err != nil {
				return nil, err
			}
			stmt.Having = having
		}
	}

	// Parse optional ORDER BY
	if p.peekToken.Type == ORDER {
		p.nextToken() // consume ORDER
		if !p.expectPeek(BY) {
			return nil, fmt.Errorf("expected BY after ORDER")
		}

		stmt.OrderBy = make([]OrderByExpr, 0)
		for {
			if !p.expectPeek(IDENT) {
				return nil, fmt.Errorf("expected identifier in ORDER BY")
			}
			expr := OrderByExpr{
				Expr: &Identifier{
					BaseNode: BaseNode{nodeType: IDENTIFIER},
					Name:     p.currentToken.Literal,
				},
				Ascending: true,
			}

			// Check for optional ASC/DESC
			if p.peekToken.Type == ASC || p.peekToken.Type == DESC {
				p.nextToken()
				expr.Ascending = p.currentToken.Type == ASC
			}

			stmt.OrderBy = append(stmt.OrderBy, expr)

			if p.peekToken.Type != COMMA {
				break
			}
			p.nextToken() // consume comma
		}
	}

	// Parse optional LIMIT and OFFSET
	if p.peekToken.Type == LIMIT {
		p.nextToken() // consume LIMIT
		if !p.expectPeek(NUMBER) {
			return nil, fmt.Errorf("expected number after LIMIT")
		}
		limit, err := strconv.Atoi(p.currentToken.Literal)
		if err != nil {
			return nil, fmt.Errorf("invalid LIMIT value: %s", p.currentToken.Literal)
		}
		stmt.Limit = &limit

		// Parse optional OFFSET
		if p.peekToken.Type == OFFSET {
			p.nextToken() // consume OFFSET
			if !p.expectPeek(NUMBER) {
				return nil, fmt.Errorf("expected number after OFFSET")
			}
			offset, err := strconv.Atoi(p.currentToken.Literal)
			if err != nil {
				return nil, fmt.Errorf("invalid OFFSET value: %s", p.currentToken.Literal)
			}
			stmt.Offset = &offset
		}
	}

	return stmt, nil
}

func (p *Parser) parseJoin(stmt *SelectStmt) error {
	switch p.currentToken.Type {
	case JOIN:
		stmt.JoinType = INNER_JOIN
	case LEFT:
		if !p.expectPeek(JOIN) {
			return fmt.Errorf("expected JOIN after LEFT")
		}
		stmt.JoinType = LEFT_JOIN
	case RIGHT:
		if !p.expectPeek(JOIN) {
			return fmt.Errorf("expected JOIN after RIGHT")
		}
		stmt.JoinType = RIGHT_JOIN
	}

	if !p.expectPeek(IDENT) {
		return fmt.Errorf("expected table name after JOIN")
	}
	stmt.JoinTable = p.currentToken.Literal

	if !p.expectPeek(ON) {
		return fmt.Errorf("expected ON after JOIN table")
	}

	p.nextToken() // move to first token of join condition
	joinCond, err := p.parseExpression(LOWEST)
	if err != nil {
		return err
	}
	stmt.JoinOn = joinCond

	return nil
}

const (
	_ int = iota
	LOWEST
	OR          // OR
	AND         // AND
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

var precedences = map[TokenType]int{
	EQ:       EQUALS,
	NEQ:      EQUALS,
	LT:       LESSGREATER,
	GT:       LESSGREATER,
	LTE:      LESSGREATER,
	GTE:      LESSGREATER,
	PLUS:     SUM,
	MINUS:    SUM,
	MULTIPLY: PRODUCT,
	DIVIDE:   PRODUCT,
	MOD:      PRODUCT,
	AND:      AND,
	OR:       OR,
	LPAREN:   CALL,
}

func (p *Parser) parseExpression(precedence int) (Expression, error) {
	prefix := p.prefixParseFn(p.currentToken.Type)
	if prefix == nil {
		return nil, fmt.Errorf("no prefix parse function for %s found", p.currentToken.Type)
	}

	leftExp, err := prefix()
	if err != nil {
		return nil, err
	}

	for precedence < p.peekPrecedence() {
		infix := p.infixParseFn(p.peekToken.Type)
		if infix == nil {
			return leftExp, nil
		}

		p.nextToken()

		leftExp, err = infix(leftExp)
		if err != nil {
			return nil, err
		}
	}

	return leftExp, nil
}

func (p *Parser) prefixParseFn(tokenType TokenType) func() (Expression, error) {
	switch tokenType {
	case IDENT:
		return p.parseIdentifier
	case STRING:
		return p.parseStringLiteral
	case NUMBER:
		return p.parseNumberLiteral
	case LPAREN:
		return p.parseGroupedExpression
	case MINUS:
		return p.parsePrefixExpression
	case NOT:
		return p.parsePrefixExpression
	}
	return nil
}

func (p *Parser) infixParseFn(tokenType TokenType) func(Expression) (Expression, error) {
	switch tokenType {
	case PLUS, MINUS, MULTIPLY, DIVIDE, MOD,
		EQ, NEQ, LT, GT, LTE, GTE,
		AND, OR:
		return p.parseInfixExpression
	case LPAREN:
		return p.parseFunctionCall
	}
	return nil
}

func (p *Parser) parseIdentifier() (Expression, error) {
	return &Identifier{
		BaseNode: BaseNode{nodeType: IDENTIFIER},
		Name:     p.currentToken.Literal,
	}, nil
}

func (p *Parser) parseStringLiteral() (Expression, error) {
	return &Literal{
		BaseNode: BaseNode{nodeType: STRING_LIT},
		Value:    p.currentToken.Literal,
	}, nil
}

func (p *Parser) parseNumberLiteral() (Expression, error) {
	return &Literal{
		BaseNode: BaseNode{nodeType: NUMBER_LIT},
		Value:    p.currentToken.Literal,
	}, nil
}

func (p *Parser) parseGroupedExpression() (Expression, error) {
	p.nextToken()

	exp, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}

	if !p.expectPeek(RPAREN) {
		return nil, fmt.Errorf("expected )")
	}

	return exp, nil
}

func (p *Parser) parsePrefixExpression() (Expression, error) {
	operator := p.currentToken.Literal
	p.nextToken()

	right, err := p.parseExpression(PREFIX)
	if err != nil {
		return nil, err
	}

	return &ComparisonExpr{
		BaseNode: BaseNode{nodeType: COMPARISON},
		Left:     nil,
		Operator: operator,
		Right:    right,
	}, nil
}

func (p *Parser) parseInfixExpression(left Expression) (Expression, error) {
	operator := p.currentToken.Literal
	precedence := p.curPrecedence()
	p.nextToken()

	right, err := p.parseExpression(precedence)
	if err != nil {
		return nil, err
	}

	return &ComparisonExpr{
		BaseNode: BaseNode{nodeType: COMPARISON},
		Left:     left,
		Operator: operator,
		Right:    right,
	}, nil
}

func (p *Parser) parseFunctionCall(function Expression) (Expression, error) {
	ident, ok := function.(*Identifier)
	if !ok {
		return nil, fmt.Errorf("expected function name")
	}

	args, err := p.parseFunctionArguments()
	if err != nil {
		return nil, err
	}

	return &FunctionExpr{
		BaseNode: BaseNode{nodeType: FUNCTION},
		Name:     ident.Name,
		Args:     args,
	}, nil
}

func (p *Parser) parseFunctionArguments() ([]Expression, error) {
	args := make([]Expression, 0)

	if p.peekToken.Type == RPAREN {
		p.nextToken()
		return args, nil
	}

	p.nextToken()
	arg, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, err
	}
	args = append(args, arg)

	for p.peekToken.Type == COMMA {
		p.nextToken()
		p.nextToken()
		arg, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	if !p.expectPeek(RPAREN) {
		return nil, fmt.Errorf("expected ) after function arguments")
	}

	return args, nil
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.currentToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseCreateTable() (*CreateTableStmt, error) {
	stmt := &CreateTableStmt{BaseNode: BaseNode{nodeType: CREATE_TABLE}}

	// Parse TABLE keyword
	if !p.expectPeek(TABLE) {
		return nil, fmt.Errorf("expected TABLE after CREATE")
	}

	// Parse table name
	if !p.expectPeek(IDENT) {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.TableName = p.currentToken.Literal

	// Parse column definitions
	if !p.expectPeek(LPAREN) {
		return nil, fmt.Errorf("expected ( after table name")
	}

	stmt.Columns = make([]ColumnDef, 0)
	for {
		if p.currentToken.Type == RPAREN {
			break
		}

		// 解析列名
		if p.currentToken.Type != IDENT {
			return nil, fmt.Errorf("expected column name")
		}
		colName := p.currentToken.Literal

		// 解析数据类型
		if !p.expectPeek(IDENT) {
			return nil, fmt.Errorf("expected data type")
		}
		dataType := p.currentToken.Literal

		// 解析约束
		constraints := make([]string, 0)
		for p.peekToken.Type != COMMA && p.peekToken.Type != RPAREN {
			p.nextToken()
			constraints = append(constraints, p.currentToken.Literal)
		}

		col := ColumnDef{
			Name:        colName,
			DataType:    dataType,
			Constraints: constraints,
		}
		stmt.Columns = append(stmt.Columns, col)

		if p.peekToken.Type == RPAREN {
			break
		}
		if !p.expectPeek(COMMA) {
			return nil, fmt.Errorf("expected comma or right paren")
		}
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropTable() (*DropTableStmt, error) {
	stmt := &DropTableStmt{BaseNode: BaseNode{nodeType: DROP_TABLE}}

	// Parse TABLE keyword
	if !p.expectPeek(TABLE) {
		return nil, fmt.Errorf("expected TABLE after DROP")
	}

	// Parse table name
	if !p.expectPeek(IDENT) {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.TableName = p.currentToken.Literal

	return stmt, nil
}

func (p *Parser) parseInsert() (*InsertStmt, error) {
	stmt := &InsertStmt{BaseNode: BaseNode{nodeType: INSERT}}

	// Parse INTO keyword
	if !p.expectPeek(INTO) {
		return nil, fmt.Errorf("expected INTO after INSERT")
	}

	// Parse table name
	if !p.expectPeek(IDENT) {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.Table = p.currentToken.Literal

	// Parse optional column list
	if p.peekToken.Type == LPAREN {
		p.nextToken()
		stmt.Columns = make([]string, 0)
		for {
			if !p.expectPeek(IDENT) {
				return nil, fmt.Errorf("expected column name")
			}
			stmt.Columns = append(stmt.Columns, p.currentToken.Literal)

			if p.peekToken.Type == RPAREN {
				p.nextToken()
				break
			}
			if !p.expectPeek(COMMA) {
				return nil, fmt.Errorf("expected , or ) after column name")
			}
		}
	}

	// Parse VALUES keyword
	if !p.expectPeek(VALUES) {
		return nil, fmt.Errorf("expected VALUES")
	}

	// Parse value list
	if !p.expectPeek(LPAREN) {
		return nil, fmt.Errorf("expected ( after VALUES")
	}

	stmt.Values = make([]Expression, 0)
	for {
		p.nextToken()
		if p.currentToken.Type == RPAREN {
			break
		}

		expr, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		stmt.Values = append(stmt.Values, expr)

		if p.peekToken.Type == RPAREN {
			p.nextToken()
			break
		}
		if !p.expectPeek(COMMA) {
			return nil, fmt.Errorf("expected , or ) after value")
		}
	}

	return stmt, nil
}

func (p *Parser) parseUpdate() (*UpdateStmt, error) {
	stmt := &UpdateStmt{
		BaseNode: BaseNode{nodeType: UPDATE},
		Set:      make(map[string]Expression),
	}

	// Parse table name
	if !p.expectPeek(IDENT) {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.Table = p.currentToken.Literal

	// Parse SET keyword
	if !p.expectPeek(SET) {
		return nil, fmt.Errorf("expected SET")
	}

	// Parse set assignments
	for {
		if !p.expectPeek(IDENT) {
			return nil, fmt.Errorf("expected column name")
		}
		column := p.currentToken.Literal

		if !p.expectPeek(EQ) {
			return nil, fmt.Errorf("expected = after column name")
		}

		p.nextToken()
		value, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		stmt.Set[column] = value

		if p.peekToken.Type != COMMA {
			break
		}
		p.nextToken()
	}

	// Parse optional WHERE clause
	if p.peekToken.Type == WHERE {
		p.nextToken()
		p.nextToken()
		where, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		stmt.Where = where
	}

	return stmt, nil
}

func (p *Parser) parseDelete() (*DeleteStmt, error) {
	stmt := &DeleteStmt{BaseNode: BaseNode{nodeType: DELETE}}

	// Parse FROM keyword
	if !p.expectPeek(FROM) {
		return nil, fmt.Errorf("expected FROM after DELETE")
	}

	// Parse table name
	if !p.expectPeek(IDENT) {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.Table = p.currentToken.Literal

	// Parse optional WHERE clause
	if p.peekToken.Type == WHERE {
		p.nextToken()
		p.nextToken()
		where, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		stmt.Where = where
	}

	return stmt, nil
}

func (p *Parser) parseShowTables() (*ShowTablesStmt, error) {
	return &ShowTablesStmt{BaseNode: BaseNode{nodeType: SHOW_TABLES}}, nil
}

package parser

import (
	"fmt"
	"strconv"
)

// Parser SQL解析器
type Parser struct {
	lexer        *Lexer   // 词法分析器
	currentToken Token    // 当前token
	peekToken    Token    // 下一个token
	errors       []string // 解析错误
}

// NewParser 创建新的解析器
func NewParser(input string) *Parser {
	p := &Parser{
		lexer: NewLexer(input),
	}
	// 读取两个token，设置current和peek
	p.nextToken()
	p.nextToken()
	return p
}

// Parse 解析SQL语句
func (p *Parser) Parse() (Statement, error) {
	switch p.currentToken.Type {
	case TOK_SELECT:
		return p.parseSelect()
	case TOK_INSERT:
		return p.parseInsert()
	case TOK_UPDATE:
		return p.parseUpdate()
	case TOK_DELETE:
		return p.parseDelete()
	case TOK_CREATE:
		return p.parseCreate()
	default:
		return nil, fmt.Errorf("unexpected token: %s", p.currentToken)
	}
}

// parseSelect 解析SELECT语句
func (p *Parser) parseSelect() (*SelectStmt, error) {
	stmt := &SelectStmt{}

	// 检查SELECT后是否有字段
	if p.peekTokenIs(TOK_FROM) {
		return nil, &ParseError{
			Message:  "no fields specified",
			Line:     p.peekToken.Line,
			Column:   p.peekToken.Column,
			Token:    p.peekToken,
			Expected: "field list",
		}
	}

	// 解析字段列表
	fields, err := p.parseExpressionList()
	if err != nil {
		return nil, err
	}
	stmt.Fields = fields

	// 检查是否有FROM子句
	if !p.expectPeek(TOK_FROM) {
		return nil, &ParseError{
			Message:  "missing FROM clause",
			Line:     p.currentToken.Line,
			Column:   p.currentToken.Column,
			Token:    p.currentToken,
			Expected: "FROM",
		}
	}

	// 添加类型检查
	for _, field := range stmt.Fields {
		if err := p.validateExpression(field); err != nil {
			return nil, err
		}
	}

	// 解析FROM子句
	if !p.expectPeek(TOK_FROM) {
		return nil, fmt.Errorf("expected FROM")
	}
	if !p.expectPeek(TOK_IDENT) {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.Table = p.currentToken.Literal

	// 解析WHERE子句
	if p.peekTokenIs(TOK_WHERE) {
		p.nextToken()
		where, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		stmt.Where = where
	}

	// 解析ORDER BY子句
	if p.peekTokenIs(TOK_ORDER) {
		p.nextToken()
		if !p.expectPeek(TOK_BY) {
			return nil, fmt.Errorf("expected BY after ORDER")
		}
		orderBy, err := p.parseOrderBy()
		if err != nil {
			return nil, err
		}
		stmt.OrderBy = orderBy
	}

	// 解析LIMIT子句
	if p.peekTokenIs(TOK_LIMIT) {
		p.nextToken()
		if !p.expectPeek(TOK_NUMBER) {
			return nil, fmt.Errorf("expected number after LIMIT")
		}
		limit, err := strconv.Atoi(p.currentToken.Literal)
		if err != nil {
			return nil, fmt.Errorf("invalid LIMIT value")
		}
		stmt.Limit = &limit
	}

	return stmt, nil
}

// parseInsert 解析INSERT语句
func (p *Parser) parseInsert() (*InsertStmt, error) {
	stmt := &InsertStmt{}

	// 解析INTO
	if !p.expectPeek(TOK_INTO) {
		return nil, fmt.Errorf("expected INTO")
	}

	// 解析表名
	if !p.expectPeek(TOK_IDENT) {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.Table = p.currentToken.Literal

	// 解析列名列表
	if p.peekTokenIs(TOK_LPAREN) {
		p.nextToken()
		columns, err := p.parseIdentList()
		if err != nil {
			return nil, err
		}
		stmt.Columns = columns
	}

	// 解析VALUES
	if !p.expectPeek(TOK_VALUES) {
		return nil, fmt.Errorf("expected VALUES")
	}

	// 解析值列表
	if !p.expectPeek(TOK_LPAREN) {
		return nil, fmt.Errorf("expected (")
	}
	values, err := p.parseExpressionList()
	if err != nil {
		return nil, err
	}
	stmt.Values = values

	return stmt, nil
}

// parseUpdate 解析UPDATE语句
func (p *Parser) parseUpdate() (*UpdateStmt, error) {
	stmt := &UpdateStmt{
		Set: make(map[string]Expression),
	}

	// 解析表名
	if !p.expectPeek(TOK_IDENT) {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.Table = p.currentToken.Literal

	// 解析SET
	if !p.expectPeek(TOK_SET) {
		return nil, fmt.Errorf("expected SET")
	}

	// 解析赋值列表
	for {
		if !p.expectPeek(TOK_IDENT) {
			return nil, fmt.Errorf("expected column name")
		}
		column := p.currentToken.Literal

		if !p.expectPeek(TOK_EQ) {
			return nil, fmt.Errorf("expected =")
		}

		p.nextToken()
		value, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		stmt.Set[column] = value

		if !p.peekTokenIs(TOK_COMMA) {
			break
		}
		p.nextToken()
	}

	// 解析WHERE子句
	if p.peekTokenIs(TOK_WHERE) {
		p.nextToken()
		where, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		stmt.Where = where
	}

	return stmt, nil
}

// parseDelete 解析DELETE语句
func (p *Parser) parseDelete() (*DeleteStmt, error) {
	stmt := &DeleteStmt{}

	// 解析FROM
	if !p.expectPeek(TOK_FROM) {
		return nil, fmt.Errorf("expected FROM")
	}

	// 解析表名
	if !p.expectPeek(TOK_IDENT) {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.Table = p.currentToken.Literal

	// 解析WHERE子句
	if p.peekTokenIs(TOK_WHERE) {
		p.nextToken()
		where, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		stmt.Where = where
	}

	return stmt, nil
}

// 表达式优先级
const (
	LOWEST      = iota
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

var precedences = map[TokenType]int{
	TOK_EQ:       EQUALS,
	TOK_NEQ:      EQUALS,
	TOK_LT:       LESSGREATER,
	TOK_GT:       LESSGREATER,
	TOK_LTE:      LESSGREATER,
	TOK_GTE:      LESSGREATER,
	TOK_PLUS:     SUM,
	TOK_MINUS:    SUM,
	TOK_MULTIPLY: PRODUCT,
	TOK_DIVIDE:   PRODUCT,
	TOK_LPAREN:   CALL,
}

// parseExpression 解析表达式
func (p *Parser) parseExpression(precedence int) (Expression, error) {
	prefix := p.prefixParseFns(p.currentToken.Type)
	if prefix == nil {
		return nil, fmt.Errorf("no prefix parse function for %s", p.currentToken.Type)
	}

	leftExp := prefix()

	for !p.peekTokenIs(TOK_SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns(p.peekToken.Type)
		if infix == nil {
			return leftExp, nil
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp, nil
}

// 辅助方法

func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) currentTokenIs(t TokenType) bool {
	return p.currentToken.Type == t
}

func (p *Parser) peekTokenIs(t TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) Errors() []string {
	return p.errors
}

// prefixParseFns 返回前缀表达式解析函数
func (p *Parser) prefixParseFns(tokenType TokenType) func() Expression {
	switch tokenType {
	case TOK_IDENT:
		return p.parseIdentifier
	case TOK_STRING, TOK_NUMBER:
		return p.parseLiteral
	case TOK_MINUS:
		return p.parsePrefixExpression
	case TOK_LPAREN:
		return p.parseGroupedExpression
	default:
		return nil
	}
}

// infixParseFns 返回中缀表达式解析函数
func (p *Parser) infixParseFns(tokenType TokenType) func(Expression) Expression {
	switch tokenType {
	case TOK_PLUS, TOK_MINUS, TOK_MULTIPLY, TOK_DIVIDE:
		return p.parseBinaryExpression
	case TOK_EQ, TOK_NEQ, TOK_LT, TOK_GT, TOK_LTE, TOK_GTE:
		return p.parseComparisonExpression
	default:
		return nil
	}
}

// 具体的解析函数实现...

// parseCreate 解析CREATE语句
func (p *Parser) parseCreate() (*CreateTableStmt, error) {
	stmt := &CreateTableStmt{}

	// 期望下一个token是TABLE
	if !p.expectPeek(TOK_TABLE) {
		return nil, fmt.Errorf("expected TABLE after CREATE")
	}

	// 解析表名
	if !p.expectPeek(TOK_IDENT) {
		return nil, fmt.Errorf("expected table name")
	}
	stmt.TableName = p.currentToken.Literal

	// 解析列定义
	if !p.expectPeek(TOK_LPAREN) {
		return nil, fmt.Errorf("expected (")
	}

	columns, err := p.parseColumnDefs()
	if err != nil {
		return nil, err
	}
	stmt.Columns = columns

	return stmt, nil
}

// parseColumnDefs 解析列定义
func (p *Parser) parseColumnDefs() ([]ColumnDef, error) {
	var columns []ColumnDef

	for {
		if p.peekTokenIs(TOK_RPAREN) {
			p.nextToken()
			break
		}

		col, err := p.parseColumnDef()
		if err != nil {
			return nil, err
		}
		columns = append(columns, col)

		if !p.peekTokenIs(TOK_COMMA) {
			if !p.expectPeek(TOK_RPAREN) {
				return nil, fmt.Errorf("expected , or )")
			}
			break
		}
		p.nextToken()
	}

	return columns, nil
}

// parseColumnDef 解析单个列定义
func (p *Parser) parseColumnDef() (ColumnDef, error) {
	var col ColumnDef

	if !p.expectPeek(TOK_IDENT) {
		return col, fmt.Errorf("expected column name")
	}
	col.Name = p.currentToken.Literal

	if !p.expectPeek(TOK_IDENT) {
		return col, fmt.Errorf("expected data type")
	}
	col.DataType = p.currentToken.Literal

	// 检查是否有NOT NULL约束
	if p.peekTokenIs(TOK_IDENT) {
		p.nextToken()
		if p.currentToken.Literal == "NOT" {
			if !p.expectPeek(TOK_IDENT) || p.currentToken.Literal != "NULL" {
				return col, fmt.Errorf("expected NULL after NOT")
			}
			col.NotNull = true
		}
	}

	return col, nil
}

// parseIdentifier 解析标识符
func (p *Parser) parseIdentifier() Expression {
	return &Identifier{Name: p.currentToken.Literal}
}

// parseLiteral 解析字面量
func (p *Parser) parseLiteral() Expression {
	return &Literal{
		Value: p.currentToken.Literal,
		Type:  string(p.currentToken.Type),
	}
}

// parsePrefixExpression 解析前缀表达式
func (p *Parser) parsePrefixExpression() Expression {
	expression := &BinaryExpr{
		Operator: p.currentToken.Literal,
	}

	p.nextToken()
	right, err := p.parseExpression(PREFIX)
	if err != nil {
		return nil
	}
	expression.Right = right

	return expression
}

// parseGroupedExpression 解析括号表达式
func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()
	exp, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil
	}

	if !p.expectPeek(TOK_RPAREN) {
		return nil
	}

	return exp
}

// parseBinaryExpression 解析二元表达式
func (p *Parser) parseBinaryExpression(left Expression) Expression {
	expression := &BinaryExpr{
		Left:     left,
		Operator: p.currentToken.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	right, err := p.parseExpression(precedence)
	if err != nil {
		return nil
	}
	expression.Right = right

	return expression
}

// parseComparisonExpression 解析比较表达式
func (p *Parser) parseComparisonExpression(left Expression) Expression {
	expression := &ComparisonExpr{
		Left:     left,
		Operator: p.currentToken.Literal,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	right, err := p.parseExpression(precedence)
	if err != nil {
		return nil
	}
	expression.Right = right

	return expression
}

// peekPrecedence 获取下一个token的优先级
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// curPrecedence 获取当前token的优先级
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.currentToken.Type]; ok {
		return p
	}
	return LOWEST
}

// 添加详细的错误类型
type ParseError struct {
	Message  string
	Line     int
	Column   int
	Token    Token
	Expected string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("syntax error at line %d, column %d: %s (got %s, expected %s)",
		e.Line, e.Column, e.Message, e.Token.Literal, e.Expected)
}

// 添加表达式验证
func (p *Parser) validateExpression(expr Expression) error {
	switch e := expr.(type) {
	case *BinaryExpr:
		if err := p.validateBinaryOperator(e.Operator); err != nil {
			return err
		}
		if err := p.validateExpression(e.Left); err != nil {
			return err
		}
		return p.validateExpression(e.Right)

	case *ComparisonExpr:
		if err := p.validateComparisonOperator(e.Operator); err != nil {
			return err
		}
		if err := p.validateExpression(e.Left); err != nil {
			return err
		}
		return p.validateExpression(e.Right)
	}
	return nil
}

// parseExpressionList 解析表达式列表
func (p *Parser) parseExpressionList() ([]Expression, error) {
	var expressions []Expression

	for {
		expr, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, expr)

		if !p.peekTokenIs(TOK_COMMA) {
			break
		}
		p.nextToken()
	}

	return expressions, nil
}

// parseOrderBy 解析ORDER BY子句
func (p *Parser) parseOrderBy() ([]OrderByExpr, error) {
	var orderBy []OrderByExpr

	for {
		expr, err := p.parseExpression(LOWEST)
		if err != nil {
			return nil, err
		}

		ascending := true
		if p.peekTokenIs(TOK_DESC) {
			ascending = false
			p.nextToken()
		}

		orderBy = append(orderBy, OrderByExpr{
			Expr:      expr,
			Ascending: ascending,
		})

		if !p.peekTokenIs(TOK_COMMA) {
			break
		}
		p.nextToken()
	}

	return orderBy, nil
}

// validateBinaryOperator 验证二元运算符
func (p *Parser) validateBinaryOperator(operator string) error {
	validOperators := map[string]bool{
		"+": true,
		"-": true,
		"*": true,
		"/": true,
	}

	if !validOperators[operator] {
		return fmt.Errorf("invalid binary operator: %s", operator)
	}
	return nil
}

// validateComparisonOperator 验证比较运算符
func (p *Parser) validateComparisonOperator(operator string) error {
	validOperators := map[string]bool{
		"=":  true,
		"!=": true,
		"<":  true,
		">":  true,
		"<=": true,
		">=": true,
	}

	if !validOperators[operator] {
		return fmt.Errorf("invalid comparison operator: %s", operator)
	}
	return nil
}

// parseIdentList 解析标识符列表
func (p *Parser) parseIdentList() ([]string, error) {
	var identifiers []string

	// 解析第一个标识符
	if !p.expectPeek(TOK_IDENT) {
		return nil, fmt.Errorf("expected identifier, got %s", p.peekToken.Type)
	}
	identifiers = append(identifiers, p.currentToken.Literal)

	// 解析后续的标识符
	for p.peekTokenIs(TOK_COMMA) {
		p.nextToken() // 跳过逗号
		if !p.expectPeek(TOK_IDENT) {
			return nil, fmt.Errorf("expected identifier after comma, got %s", p.peekToken.Type)
		}
		identifiers = append(identifiers, p.currentToken.Literal)
	}

	// 检查右括号
	if !p.expectPeek(TOK_RPAREN) {
		return nil, fmt.Errorf("expected ), got %s", p.peekToken.Type)
	}

	return identifiers, nil
}

// ParseWhereExpression 专门用于解析WHERE条件表达式
func (p *Parser) ParseWhereExpression() (Expression, error) {
	expr, err := p.parseExpression(LOWEST)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WHERE expression: %v", err)
	}
	return expr, nil
}

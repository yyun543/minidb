package parser

import (
	"fmt"
	"strings"
)

type TokenType int

const (
	ILLEGAL TokenType = iota
	EOF
	WS

	// 标识符和字面量
	IDENT  // 表名、列名等
	STRING // 字符串字面量
	NUMBER // 数字字面量

	// 关键字
	SELECT
	INSERT
	UPDATE
	DELETE
	CREATE
	DROP
	TABLE
	FROM
	WHERE
	SET
	INTO
	VALUES
	AND
	OR
	JOIN
	LEFT
	RIGHT
	INNER
	ON
	GROUP
	BY
	HAVING
	ORDER
	ASC
	DESC
	LIMIT
	OFFSET
	AS
	LIKE
	IN
	NOT
	NULL
	IS
	TRUE
	FALSE

	// 运算符和分隔符
	COMMA     // ,
	SEMICOLON // ;
	LPAREN    // (
	RPAREN    // )
	EQ        // =
	NEQ       // !=
	LT        // <
	GT        // >
	LTE       // <=
	GTE       // >=
	PLUS      // +
	MINUS     // -
	MULTIPLY  // *
	DIVIDE    // /
	MOD       // %
	DOT       // .
)

var keywords = map[string]TokenType{
	"select": SELECT,
	"insert": INSERT,
	"update": UPDATE,
	"delete": DELETE,
	"create": CREATE,
	"drop":   DROP,
	"table":  TABLE,
	"from":   FROM,
	"where":  WHERE,
	"set":    SET,
	"into":   INTO,
	"values": VALUES,
	"and":    AND,
	"or":     OR,
	"join":   JOIN,
	"left":   LEFT,
	"right":  RIGHT,
	"inner":  INNER,
	"on":     ON,
	"group":  GROUP,
	"by":     BY,
	"having": HAVING,
	"order":  ORDER,
	"asc":    ASC,
	"desc":   DESC,
	"limit":  LIMIT,
	"offset": OFFSET,
	"as":     AS,
	"like":   LIKE,
	"in":     IN,
	"not":    NOT,
	"null":   NULL,
	"is":     IS,
	"true":   TRUE,
	"false":  FALSE,
}

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %v, Literal: %q, Line: %d, Column: %d}", t.Type, t.Literal, t.Line, t.Column)
}

type Lexer struct {
	input      string
	pos        int  // 当前位置
	readPos    int  // 下一个要读取的位置
	ch         byte // 当前字符
	line       int  // 当前行号
	column     int  // 当前列号
	lastColumn int  // 上一个token的列号
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++

	if l.ch == '\n' {
		l.line++
		l.lastColumn = l.column
		l.column = 0
	} else {
		l.column++
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	// 记录token的起始位置
	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '=':
		tok = l.newToken(EQ, string(l.ch))
	case ',':
		tok = l.newToken(COMMA, string(l.ch))
	case ';':
		tok = l.newToken(SEMICOLON, string(l.ch))
	case '(':
		tok = l.newToken(LPAREN, string(l.ch))
	case ')':
		tok = l.newToken(RPAREN, string(l.ch))
	case '+':
		tok = l.newToken(PLUS, string(l.ch))
	case '-':
		tok = l.newToken(MINUS, string(l.ch))
	case '*':
		tok = l.newToken(MULTIPLY, string(l.ch))
	case '/':
		if l.peekChar() == '/' {
			l.skipSingleLineComment()
			return l.NextToken()
		} else if l.peekChar() == '*' {
			l.skipMultiLineComment()
			return l.NextToken()
		}
		tok = l.newToken(DIVIDE, string(l.ch))
	case '%':
		tok = l.newToken(MOD, string(l.ch))
	case '.':
		tok = l.newToken(DOT, string(l.ch))
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(LTE, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(LT, string(l.ch))
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(GTE, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(GT, string(l.ch))
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(NEQ, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(ILLEGAL, string(l.ch))
		}
	case '"', '\'':
		tok.Type = STRING
		tok.Literal = l.readString(l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = lookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = NUMBER
			return tok
		} else {
			tok = l.newToken(ILLEGAL, string(l.ch))
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipSingleLineComment() {
	l.readChar() // 跳过第二个 /
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	if l.ch == '\n' {
		l.readChar()
	}
}

func (l *Lexer) skipMultiLineComment() {
	l.readChar() // 跳过 *
	for {
		if l.ch == 0 {
			return
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // 跳过 *
			l.readChar() // 跳过 /
			return
		}
		l.readChar()
	}
}

func (l *Lexer) readString(quote byte) string {
	position := l.pos + 1
	for {
		l.readChar()
		if l.ch == quote || l.ch == 0 {
			break
		}
		if l.ch == '\\' {
			l.readChar() // 跳过转义字符
		}
	}
	if l.ch == 0 {
		return l.input[position:l.pos]
	}
	return l.input[position:l.pos]
}

func (l *Lexer) readIdentifier() string {
	position := l.pos
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.pos]
}

func (l *Lexer) readNumber() string {
	position := l.pos
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar() // 读取小数点
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	return l.input[position:l.pos]
}

func (l *Lexer) newToken(tokenType TokenType, literal string) Token {
	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.line,
		Column:  l.column,
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func lookupIdent(ident string) TokenType {
	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}

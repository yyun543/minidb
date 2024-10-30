package parser

import (
	"fmt"
	"unicode"
)

// Lexer 词法分析器
type Lexer struct {
	input   string // 输入的SQL字符串
	pos     int    // 当前位置
	readPos int    // 下一个要读取的位置
	ch      byte   // 当前字符
	line    int    // 当前行号
	column  int    // 当前列号
}

// NewLexer 创建新的词法分析器
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar() // 读取第一个字符
	return l
}

// readChar 读取下一个字符
func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++

	// 更新行号和列号
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

// peekChar 查看下一个字符但不移动位置
func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

// NextToken 获取下一个token
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	var tok Token
	tok.Line = l.line
	tok.Column = l.column
	tok.Offset = l.pos

	switch l.ch {
	// 单字符token
	case '=':
		tok = l.newToken(TOK_EQ, string(l.ch))
	case '+':
		tok = l.newToken(TOK_PLUS, string(l.ch))
	case '-':
		tok = l.newToken(TOK_MINUS, string(l.ch))
	case '*':
		tok = l.newToken(TOK_MULTIPLY, string(l.ch))
	case '/':
		tok = l.newToken(TOK_DIVIDE, string(l.ch))
	case '(':
		tok = l.newToken(TOK_LPAREN, string(l.ch))
	case ')':
		tok = l.newToken(TOK_RPAREN, string(l.ch))
	case ',':
		tok = l.newToken(TOK_COMMA, string(l.ch))
	case ';':
		tok = l.newToken(TOK_SEMICOLON, string(l.ch))
	case '.':
		tok = l.newToken(TOK_DOT, string(l.ch))

	// 双字符token
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TOK_LTE, Literal: string(ch) + string(l.ch)}
		} else {
			tok = l.newToken(TOK_LT, string(l.ch))
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TOK_GTE, Literal: string(ch) + string(l.ch)}
		} else {
			tok = l.newToken(TOK_GT, string(l.ch))
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: TOK_NEQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = l.newToken(TOK_ILLEGAL, string(l.ch))
		}

	// 字符串
	case '"', '\'':
		tok.Type = TOK_STRING
		tok.Literal = l.readString(l.ch)

	// EOF
	case 0:
		tok.Type = TOK_EOF
		tok.Literal = ""

	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = TOK_NUMBER
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = l.newToken(TOK_ILLEGAL, string(l.ch))
		}
	}

	l.readChar()
	return tok
}

// 辅助方法

// newToken 创建新token
func (l *Lexer) newToken(tokenType TokenType, literal string) Token {
	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.line,
		Column:  l.column,
		Offset:  l.pos,
	}
}

// readIdentifier 读取标识符
func (l *Lexer) readIdentifier() string {
	position := l.pos
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.pos]
}

// readNumber 读取数字
func (l *Lexer) readNumber() string {
	position := l.pos
	for isDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	return l.input[position:l.pos]
}

// readString 读取字符串
func (l *Lexer) readString(quote byte) string {
	position := l.pos + 1
	for {
		l.readChar()
		if l.ch == quote || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.pos]
}

// skipWhitespace 跳过空白字符
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// isLetter 判断是否是字母
func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

// isDigit 判断是否是数字
func isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}

// String 返回Token的字符串表示
func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %v, Literal: %q, Line: %d, Column: %d}",
		t.Type, t.Literal, t.Line, t.Column)
}

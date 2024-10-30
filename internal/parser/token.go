package parser

// TokenType 表示词法单元的类型
type TokenType string

// Token 表示词法单元
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
	Offset  int
}

// 定义所有的token类型常量
const (
	TOK_ILLEGAL = "ILLEGAL"
	TOK_EOF     = "EOF"

	// 标识符和字面量
	TOK_IDENT  = "IDENT"
	TOK_NUMBER = "NUMBER"
	TOK_STRING = "STRING"

	// 运算符
	TOK_EQ       = "="
	TOK_NEQ      = "!="
	TOK_LT       = "<"
	TOK_GT       = ">"
	TOK_LTE      = "<="
	TOK_GTE      = ">="
	TOK_PLUS     = "+"
	TOK_MINUS    = "-"
	TOK_MULTIPLY = "*"
	TOK_DIVIDE   = "/"

	// 分隔符
	TOK_COMMA     = ","
	TOK_SEMICOLON = ";"
	TOK_LPAREN    = "("
	TOK_RPAREN    = ")"
	TOK_DOT       = "."

	// 关键字
	TOK_SELECT = "SELECT"
	TOK_FROM   = "FROM"
	TOK_WHERE  = "WHERE"
	TOK_INSERT = "INSERT"
	TOK_INTO   = "INTO"
	TOK_VALUES = "VALUES"
	TOK_UPDATE = "UPDATE"
	TOK_SET    = "SET"
	TOK_DELETE = "DELETE"
	TOK_CREATE = "CREATE"
	TOK_TABLE  = "TABLE"
	TOK_ORDER  = "ORDER"
	TOK_BY     = "BY"
	TOK_LIMIT  = "LIMIT"
)

// keywords 保存所有SQL关键字
var keywords = map[string]TokenType{
	"SELECT": TOK_SELECT,
	"FROM":   TOK_FROM,
	"WHERE":  TOK_WHERE,
	"INSERT": TOK_INSERT,
	"INTO":   TOK_INTO,
	"VALUES": TOK_VALUES,
	"UPDATE": TOK_UPDATE,
	"SET":    TOK_SET,
	"DELETE": TOK_DELETE,
	"CREATE": TOK_CREATE,
	"TABLE":  TOK_TABLE,
	"ORDER":  TOK_ORDER,
	"BY":     TOK_BY,
	"LIMIT":  TOK_LIMIT,
}

// LookupIdent 检查标识符是否是关键字
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TOK_IDENT
}

// Code generated from MiniQL.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // MiniQL
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type MiniQLParser struct {
	*antlr.BaseParser
}

var MiniQLParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func miniqlParserInit() {
	staticData := &MiniQLParserStaticData
	staticData.LiteralNames = []string{
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "'='", "'!='",
		"'>'", "'>='", "'<'", "'<='", "'+'", "'-'", "", "'/'", "'.'", "','",
		"';'", "'('", "')'",
	}
	staticData.SymbolicNames = []string{
		"", "SINGLE_LINE_COMMENT", "MULTI_LINE_COMMENT", "SELECT", "FROM", "WHERE",
		"GROUP", "BY", "HAVING", "ORDER", "LIMIT", "INSERT", "INTO", "VALUES",
		"UPDATE", "SET", "DELETE", "CREATE", "TABLE", "DATABASE", "DROP", "PRIMARY",
		"KEY", "NOT", "NULL", "TRUE", "FALSE", "AS", "LIKE", "IN", "AND", "OR",
		"JOIN", "ON", "PARTITION", "ASC", "DESC", "INNER", "LEFT", "RIGHT",
		"FULL", "OUTER", "USE", "SHOW", "DATABASES", "TABLES", "EXPLAIN", "ANALYZE",
		"VERBOSE", "UNIQUE", "DEFAULT", "INDEX", "INDEXES", "INT_TYPE", "INTEGER_TYPE",
		"VARCHAR_TYPE", "BOOLEAN_TYPE", "DOUBLE_TYPE", "TIMESTAMP_TYPE", "START",
		"TRANSACTION", "COMMIT", "ROLLBACK", "HASH", "RANGE", "ASTERISK", "EQUAL",
		"NOT_EQUAL", "GREATER", "GREATER_EQUAL", "LESS", "LESS_EQUAL", "PLUS",
		"MINUS", "MULTIPLY", "DIVIDE", "DOT", "COMMA", "SEMICOLON", "LEFT_PAREN",
		"RIGHT_PAREN", "IDENTIFIER", "INTEGER_LITERAL", "FLOAT_LITERAL", "STRING_LITERAL",
		"WS",
	}
	staticData.RuleNames = []string{
		"parse", "sqlStatement", "ddlStatement", "dmlStatement", "dqlStatement",
		"dclStatement", "utilityStatement", "createDatabase", "createTable",
		"columnDef", "columnConstraint", "tableConstraint", "createIndex", "dropIndex",
		"dropTable", "dropDatabase", "insertStatement", "updateStatement", "deleteStatement",
		"selectStatement", "selectItem", "tableReference", "tableReferenceAtom",
		"joinType", "expression", "primaryExpr", "comparisonOperator", "columnRef",
		"updateAssignment", "groupByItem", "orderByItem", "functionCall", "partitionMethod",
		"transactionStatement", "useStatement", "showDatabases", "showTables",
		"showIndexes", "explainStatement", "analyzeStatement", "columnList",
		"identifierList", "valueList", "tableName", "identifier", "dataType",
		"literal",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 85, 546, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7, 20, 2,
		21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2, 25, 7, 25, 2, 26,
		7, 26, 2, 27, 7, 27, 2, 28, 7, 28, 2, 29, 7, 29, 2, 30, 7, 30, 2, 31, 7,
		31, 2, 32, 7, 32, 2, 33, 7, 33, 2, 34, 7, 34, 2, 35, 7, 35, 2, 36, 7, 36,
		2, 37, 7, 37, 2, 38, 7, 38, 2, 39, 7, 39, 2, 40, 7, 40, 2, 41, 7, 41, 2,
		42, 7, 42, 2, 43, 7, 43, 2, 44, 7, 44, 2, 45, 7, 45, 2, 46, 7, 46, 1, 0,
		5, 0, 96, 8, 0, 10, 0, 12, 0, 99, 9, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1,
		1, 1, 1, 3, 1, 108, 8, 1, 1, 1, 3, 1, 111, 8, 1, 1, 2, 1, 2, 1, 2, 1, 2,
		1, 2, 1, 2, 3, 2, 119, 8, 2, 1, 3, 1, 3, 1, 3, 3, 3, 124, 8, 3, 1, 4, 1,
		4, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 3, 6, 136, 8, 6, 1,
		7, 1, 7, 1, 7, 1, 7, 1, 8, 1, 8, 1, 8, 1, 8, 1, 8, 1, 8, 1, 8, 5, 8, 149,
		8, 8, 10, 8, 12, 8, 152, 9, 8, 1, 8, 1, 8, 5, 8, 156, 8, 8, 10, 8, 12,
		8, 159, 9, 8, 1, 8, 1, 8, 1, 8, 1, 8, 3, 8, 165, 8, 8, 1, 9, 1, 9, 1, 9,
		5, 9, 170, 8, 9, 10, 9, 12, 9, 173, 9, 9, 1, 10, 3, 10, 176, 8, 10, 1,
		10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 3, 10, 184, 8, 10, 1, 11, 1, 11,
		1, 11, 1, 11, 1, 11, 1, 11, 1, 12, 1, 12, 3, 12, 194, 8, 12, 1, 12, 1,
		12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 13, 1, 13, 1, 13, 1, 13,
		1, 13, 1, 13, 1, 14, 1, 14, 1, 14, 1, 14, 1, 15, 1, 15, 1, 15, 1, 15, 1,
		16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 3, 16, 225, 8, 16, 1, 16,
		1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 5, 16, 236, 8,
		16, 10, 16, 12, 16, 239, 9, 16, 1, 17, 1, 17, 1, 17, 1, 17, 1, 17, 1, 17,
		5, 17, 247, 8, 17, 10, 17, 12, 17, 250, 9, 17, 1, 17, 1, 17, 3, 17, 254,
		8, 17, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18, 3, 18, 261, 8, 18, 1, 19, 1,
		19, 1, 19, 1, 19, 5, 19, 267, 8, 19, 10, 19, 12, 19, 270, 9, 19, 1, 19,
		1, 19, 1, 19, 1, 19, 3, 19, 276, 8, 19, 1, 19, 1, 19, 1, 19, 1, 19, 1,
		19, 5, 19, 283, 8, 19, 10, 19, 12, 19, 286, 9, 19, 3, 19, 288, 8, 19, 1,
		19, 1, 19, 3, 19, 292, 8, 19, 1, 19, 1, 19, 1, 19, 1, 19, 1, 19, 5, 19,
		299, 8, 19, 10, 19, 12, 19, 302, 9, 19, 3, 19, 304, 8, 19, 1, 19, 1, 19,
		3, 19, 308, 8, 19, 1, 20, 1, 20, 1, 20, 3, 20, 313, 8, 20, 1, 20, 1, 20,
		1, 20, 3, 20, 318, 8, 20, 1, 20, 3, 20, 321, 8, 20, 3, 20, 323, 8, 20,
		1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 3, 21, 330, 8, 21, 1, 21, 1, 21, 1,
		21, 1, 21, 1, 21, 5, 21, 337, 8, 21, 10, 21, 12, 21, 340, 9, 21, 1, 22,
		1, 22, 3, 22, 344, 8, 22, 1, 22, 3, 22, 347, 8, 22, 1, 22, 1, 22, 1, 22,
		1, 22, 3, 22, 353, 8, 22, 1, 22, 1, 22, 3, 22, 357, 8, 22, 1, 23, 1, 23,
		1, 23, 3, 23, 362, 8, 23, 1, 23, 1, 23, 3, 23, 366, 8, 23, 1, 23, 1, 23,
		3, 23, 370, 8, 23, 3, 23, 372, 8, 23, 1, 24, 1, 24, 1, 24, 1, 24, 1, 24,
		1, 24, 1, 24, 1, 24, 1, 24, 1, 24, 1, 24, 1, 24, 1, 24, 1, 24, 1, 24, 1,
		24, 1, 24, 1, 24, 1, 24, 1, 24, 1, 24, 3, 24, 395, 8, 24, 1, 24, 1, 24,
		1, 24, 1, 24, 3, 24, 401, 8, 24, 1, 24, 1, 24, 1, 24, 1, 24, 1, 24, 5,
		24, 408, 8, 24, 10, 24, 12, 24, 411, 9, 24, 1, 25, 1, 25, 1, 25, 1, 25,
		1, 25, 1, 25, 1, 25, 3, 25, 420, 8, 25, 1, 26, 1, 26, 1, 27, 1, 27, 1,
		27, 1, 27, 1, 27, 3, 27, 429, 8, 27, 1, 28, 1, 28, 1, 28, 1, 28, 1, 29,
		1, 29, 1, 30, 1, 30, 3, 30, 439, 8, 30, 1, 31, 1, 31, 1, 31, 1, 31, 1,
		31, 1, 31, 5, 31, 447, 8, 31, 10, 31, 12, 31, 450, 9, 31, 3, 31, 452, 8,
		31, 1, 31, 1, 31, 1, 32, 1, 32, 1, 32, 1, 32, 1, 32, 1, 32, 1, 32, 1, 32,
		1, 32, 1, 32, 3, 32, 466, 8, 32, 1, 33, 1, 33, 1, 33, 1, 33, 3, 33, 472,
		8, 33, 1, 34, 1, 34, 1, 34, 1, 35, 1, 35, 1, 35, 1, 36, 1, 36, 1, 36, 1,
		37, 1, 37, 1, 37, 1, 37, 1, 37, 1, 38, 1, 38, 1, 38, 1, 39, 1, 39, 1, 39,
		1, 39, 1, 39, 1, 39, 1, 39, 3, 39, 498, 8, 39, 1, 40, 1, 40, 1, 40, 5,
		40, 503, 8, 40, 10, 40, 12, 40, 506, 9, 40, 1, 41, 1, 41, 1, 41, 5, 41,
		511, 8, 41, 10, 41, 12, 41, 514, 9, 41, 1, 42, 1, 42, 1, 42, 5, 42, 519,
		8, 42, 10, 42, 12, 42, 522, 9, 42, 1, 43, 1, 43, 1, 43, 3, 43, 527, 8,
		43, 1, 44, 1, 44, 1, 45, 1, 45, 1, 45, 1, 45, 1, 45, 1, 45, 3, 45, 537,
		8, 45, 1, 45, 1, 45, 1, 45, 3, 45, 542, 8, 45, 1, 46, 1, 46, 1, 46, 0,
		2, 42, 48, 47, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30,
		32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56, 58, 60, 62, 64, 66,
		68, 70, 72, 74, 76, 78, 80, 82, 84, 86, 88, 90, 92, 0, 6, 2, 0, 65, 65,
		75, 75, 1, 0, 72, 73, 1, 0, 66, 71, 1, 0, 35, 36, 2, 0, 4, 4, 33, 33, 2,
		0, 24, 26, 82, 84, 585, 0, 97, 1, 0, 0, 0, 2, 107, 1, 0, 0, 0, 4, 118,
		1, 0, 0, 0, 6, 123, 1, 0, 0, 0, 8, 125, 1, 0, 0, 0, 10, 127, 1, 0, 0, 0,
		12, 135, 1, 0, 0, 0, 14, 137, 1, 0, 0, 0, 16, 141, 1, 0, 0, 0, 18, 166,
		1, 0, 0, 0, 20, 183, 1, 0, 0, 0, 22, 185, 1, 0, 0, 0, 24, 191, 1, 0, 0,
		0, 26, 203, 1, 0, 0, 0, 28, 209, 1, 0, 0, 0, 30, 213, 1, 0, 0, 0, 32, 217,
		1, 0, 0, 0, 34, 240, 1, 0, 0, 0, 36, 255, 1, 0, 0, 0, 38, 262, 1, 0, 0,
		0, 40, 322, 1, 0, 0, 0, 42, 324, 1, 0, 0, 0, 44, 356, 1, 0, 0, 0, 46, 371,
		1, 0, 0, 0, 48, 373, 1, 0, 0, 0, 50, 419, 1, 0, 0, 0, 52, 421, 1, 0, 0,
		0, 54, 428, 1, 0, 0, 0, 56, 430, 1, 0, 0, 0, 58, 434, 1, 0, 0, 0, 60, 436,
		1, 0, 0, 0, 62, 440, 1, 0, 0, 0, 64, 465, 1, 0, 0, 0, 66, 471, 1, 0, 0,
		0, 68, 473, 1, 0, 0, 0, 70, 476, 1, 0, 0, 0, 72, 479, 1, 0, 0, 0, 74, 482,
		1, 0, 0, 0, 76, 487, 1, 0, 0, 0, 78, 490, 1, 0, 0, 0, 80, 499, 1, 0, 0,
		0, 82, 507, 1, 0, 0, 0, 84, 515, 1, 0, 0, 0, 86, 523, 1, 0, 0, 0, 88, 528,
		1, 0, 0, 0, 90, 541, 1, 0, 0, 0, 92, 543, 1, 0, 0, 0, 94, 96, 3, 2, 1,
		0, 95, 94, 1, 0, 0, 0, 96, 99, 1, 0, 0, 0, 97, 95, 1, 0, 0, 0, 97, 98,
		1, 0, 0, 0, 98, 100, 1, 0, 0, 0, 99, 97, 1, 0, 0, 0, 100, 101, 5, 0, 0,
		1, 101, 1, 1, 0, 0, 0, 102, 108, 3, 4, 2, 0, 103, 108, 3, 6, 3, 0, 104,
		108, 3, 8, 4, 0, 105, 108, 3, 10, 5, 0, 106, 108, 3, 12, 6, 0, 107, 102,
		1, 0, 0, 0, 107, 103, 1, 0, 0, 0, 107, 104, 1, 0, 0, 0, 107, 105, 1, 0,
		0, 0, 107, 106, 1, 0, 0, 0, 108, 110, 1, 0, 0, 0, 109, 111, 5, 78, 0, 0,
		110, 109, 1, 0, 0, 0, 110, 111, 1, 0, 0, 0, 111, 3, 1, 0, 0, 0, 112, 119,
		3, 14, 7, 0, 113, 119, 3, 16, 8, 0, 114, 119, 3, 24, 12, 0, 115, 119, 3,
		26, 13, 0, 116, 119, 3, 28, 14, 0, 117, 119, 3, 30, 15, 0, 118, 112, 1,
		0, 0, 0, 118, 113, 1, 0, 0, 0, 118, 114, 1, 0, 0, 0, 118, 115, 1, 0, 0,
		0, 118, 116, 1, 0, 0, 0, 118, 117, 1, 0, 0, 0, 119, 5, 1, 0, 0, 0, 120,
		124, 3, 32, 16, 0, 121, 124, 3, 34, 17, 0, 122, 124, 3, 36, 18, 0, 123,
		120, 1, 0, 0, 0, 123, 121, 1, 0, 0, 0, 123, 122, 1, 0, 0, 0, 124, 7, 1,
		0, 0, 0, 125, 126, 3, 38, 19, 0, 126, 9, 1, 0, 0, 0, 127, 128, 3, 66, 33,
		0, 128, 11, 1, 0, 0, 0, 129, 136, 3, 68, 34, 0, 130, 136, 3, 70, 35, 0,
		131, 136, 3, 72, 36, 0, 132, 136, 3, 74, 37, 0, 133, 136, 3, 76, 38, 0,
		134, 136, 3, 78, 39, 0, 135, 129, 1, 0, 0, 0, 135, 130, 1, 0, 0, 0, 135,
		131, 1, 0, 0, 0, 135, 132, 1, 0, 0, 0, 135, 133, 1, 0, 0, 0, 135, 134,
		1, 0, 0, 0, 136, 13, 1, 0, 0, 0, 137, 138, 5, 17, 0, 0, 138, 139, 5, 19,
		0, 0, 139, 140, 3, 88, 44, 0, 140, 15, 1, 0, 0, 0, 141, 142, 5, 17, 0,
		0, 142, 143, 5, 18, 0, 0, 143, 144, 3, 86, 43, 0, 144, 145, 5, 79, 0, 0,
		145, 150, 3, 18, 9, 0, 146, 147, 5, 77, 0, 0, 147, 149, 3, 18, 9, 0, 148,
		146, 1, 0, 0, 0, 149, 152, 1, 0, 0, 0, 150, 148, 1, 0, 0, 0, 150, 151,
		1, 0, 0, 0, 151, 157, 1, 0, 0, 0, 152, 150, 1, 0, 0, 0, 153, 154, 5, 77,
		0, 0, 154, 156, 3, 22, 11, 0, 155, 153, 1, 0, 0, 0, 156, 159, 1, 0, 0,
		0, 157, 155, 1, 0, 0, 0, 157, 158, 1, 0, 0, 0, 158, 160, 1, 0, 0, 0, 159,
		157, 1, 0, 0, 0, 160, 164, 5, 80, 0, 0, 161, 162, 5, 34, 0, 0, 162, 163,
		5, 7, 0, 0, 163, 165, 3, 64, 32, 0, 164, 161, 1, 0, 0, 0, 164, 165, 1,
		0, 0, 0, 165, 17, 1, 0, 0, 0, 166, 167, 3, 88, 44, 0, 167, 171, 3, 90,
		45, 0, 168, 170, 3, 20, 10, 0, 169, 168, 1, 0, 0, 0, 170, 173, 1, 0, 0,
		0, 171, 169, 1, 0, 0, 0, 171, 172, 1, 0, 0, 0, 172, 19, 1, 0, 0, 0, 173,
		171, 1, 0, 0, 0, 174, 176, 5, 23, 0, 0, 175, 174, 1, 0, 0, 0, 175, 176,
		1, 0, 0, 0, 176, 177, 1, 0, 0, 0, 177, 184, 5, 24, 0, 0, 178, 179, 5, 21,
		0, 0, 179, 184, 5, 22, 0, 0, 180, 184, 5, 49, 0, 0, 181, 182, 5, 50, 0,
		0, 182, 184, 3, 92, 46, 0, 183, 175, 1, 0, 0, 0, 183, 178, 1, 0, 0, 0,
		183, 180, 1, 0, 0, 0, 183, 181, 1, 0, 0, 0, 184, 21, 1, 0, 0, 0, 185, 186,
		5, 21, 0, 0, 186, 187, 5, 22, 0, 0, 187, 188, 5, 79, 0, 0, 188, 189, 3,
		82, 41, 0, 189, 190, 5, 80, 0, 0, 190, 23, 1, 0, 0, 0, 191, 193, 5, 17,
		0, 0, 192, 194, 5, 49, 0, 0, 193, 192, 1, 0, 0, 0, 193, 194, 1, 0, 0, 0,
		194, 195, 1, 0, 0, 0, 195, 196, 5, 51, 0, 0, 196, 197, 3, 88, 44, 0, 197,
		198, 5, 33, 0, 0, 198, 199, 3, 86, 43, 0, 199, 200, 5, 79, 0, 0, 200, 201,
		3, 82, 41, 0, 201, 202, 5, 80, 0, 0, 202, 25, 1, 0, 0, 0, 203, 204, 5,
		20, 0, 0, 204, 205, 5, 51, 0, 0, 205, 206, 3, 88, 44, 0, 206, 207, 5, 33,
		0, 0, 207, 208, 3, 86, 43, 0, 208, 27, 1, 0, 0, 0, 209, 210, 5, 20, 0,
		0, 210, 211, 5, 18, 0, 0, 211, 212, 3, 86, 43, 0, 212, 29, 1, 0, 0, 0,
		213, 214, 5, 20, 0, 0, 214, 215, 5, 19, 0, 0, 215, 216, 3, 88, 44, 0, 216,
		31, 1, 0, 0, 0, 217, 218, 5, 11, 0, 0, 218, 219, 5, 12, 0, 0, 219, 224,
		3, 86, 43, 0, 220, 221, 5, 79, 0, 0, 221, 222, 3, 82, 41, 0, 222, 223,
		5, 80, 0, 0, 223, 225, 1, 0, 0, 0, 224, 220, 1, 0, 0, 0, 224, 225, 1, 0,
		0, 0, 225, 226, 1, 0, 0, 0, 226, 227, 5, 13, 0, 0, 227, 228, 5, 79, 0,
		0, 228, 229, 3, 84, 42, 0, 229, 237, 5, 80, 0, 0, 230, 231, 5, 77, 0, 0,
		231, 232, 5, 79, 0, 0, 232, 233, 3, 84, 42, 0, 233, 234, 5, 80, 0, 0, 234,
		236, 1, 0, 0, 0, 235, 230, 1, 0, 0, 0, 236, 239, 1, 0, 0, 0, 237, 235,
		1, 0, 0, 0, 237, 238, 1, 0, 0, 0, 238, 33, 1, 0, 0, 0, 239, 237, 1, 0,
		0, 0, 240, 241, 5, 14, 0, 0, 241, 242, 3, 86, 43, 0, 242, 243, 5, 15, 0,
		0, 243, 248, 3, 56, 28, 0, 244, 245, 5, 77, 0, 0, 245, 247, 3, 56, 28,
		0, 246, 244, 1, 0, 0, 0, 247, 250, 1, 0, 0, 0, 248, 246, 1, 0, 0, 0, 248,
		249, 1, 0, 0, 0, 249, 253, 1, 0, 0, 0, 250, 248, 1, 0, 0, 0, 251, 252,
		5, 5, 0, 0, 252, 254, 3, 48, 24, 0, 253, 251, 1, 0, 0, 0, 253, 254, 1,
		0, 0, 0, 254, 35, 1, 0, 0, 0, 255, 256, 5, 16, 0, 0, 256, 257, 5, 4, 0,
		0, 257, 260, 3, 86, 43, 0, 258, 259, 5, 5, 0, 0, 259, 261, 3, 48, 24, 0,
		260, 258, 1, 0, 0, 0, 260, 261, 1, 0, 0, 0, 261, 37, 1, 0, 0, 0, 262, 263,
		5, 3, 0, 0, 263, 268, 3, 40, 20, 0, 264, 265, 5, 77, 0, 0, 265, 267, 3,
		40, 20, 0, 266, 264, 1, 0, 0, 0, 267, 270, 1, 0, 0, 0, 268, 266, 1, 0,
		0, 0, 268, 269, 1, 0, 0, 0, 269, 271, 1, 0, 0, 0, 270, 268, 1, 0, 0, 0,
		271, 272, 5, 4, 0, 0, 272, 275, 3, 42, 21, 0, 273, 274, 5, 5, 0, 0, 274,
		276, 3, 48, 24, 0, 275, 273, 1, 0, 0, 0, 275, 276, 1, 0, 0, 0, 276, 287,
		1, 0, 0, 0, 277, 278, 5, 6, 0, 0, 278, 279, 5, 7, 0, 0, 279, 284, 3, 58,
		29, 0, 280, 281, 5, 77, 0, 0, 281, 283, 3, 58, 29, 0, 282, 280, 1, 0, 0,
		0, 283, 286, 1, 0, 0, 0, 284, 282, 1, 0, 0, 0, 284, 285, 1, 0, 0, 0, 285,
		288, 1, 0, 0, 0, 286, 284, 1, 0, 0, 0, 287, 277, 1, 0, 0, 0, 287, 288,
		1, 0, 0, 0, 288, 291, 1, 0, 0, 0, 289, 290, 5, 8, 0, 0, 290, 292, 3, 48,
		24, 0, 291, 289, 1, 0, 0, 0, 291, 292, 1, 0, 0, 0, 292, 303, 1, 0, 0, 0,
		293, 294, 5, 9, 0, 0, 294, 295, 5, 7, 0, 0, 295, 300, 3, 60, 30, 0, 296,
		297, 5, 77, 0, 0, 297, 299, 3, 60, 30, 0, 298, 296, 1, 0, 0, 0, 299, 302,
		1, 0, 0, 0, 300, 298, 1, 0, 0, 0, 300, 301, 1, 0, 0, 0, 301, 304, 1, 0,
		0, 0, 302, 300, 1, 0, 0, 0, 303, 293, 1, 0, 0, 0, 303, 304, 1, 0, 0, 0,
		304, 307, 1, 0, 0, 0, 305, 306, 5, 10, 0, 0, 306, 308, 5, 82, 0, 0, 307,
		305, 1, 0, 0, 0, 307, 308, 1, 0, 0, 0, 308, 39, 1, 0, 0, 0, 309, 310, 3,
		86, 43, 0, 310, 311, 5, 76, 0, 0, 311, 313, 1, 0, 0, 0, 312, 309, 1, 0,
		0, 0, 312, 313, 1, 0, 0, 0, 313, 314, 1, 0, 0, 0, 314, 323, 5, 65, 0, 0,
		315, 320, 3, 48, 24, 0, 316, 318, 5, 27, 0, 0, 317, 316, 1, 0, 0, 0, 317,
		318, 1, 0, 0, 0, 318, 319, 1, 0, 0, 0, 319, 321, 3, 88, 44, 0, 320, 317,
		1, 0, 0, 0, 320, 321, 1, 0, 0, 0, 321, 323, 1, 0, 0, 0, 322, 312, 1, 0,
		0, 0, 322, 315, 1, 0, 0, 0, 323, 41, 1, 0, 0, 0, 324, 325, 6, 21, -1, 0,
		325, 326, 3, 44, 22, 0, 326, 338, 1, 0, 0, 0, 327, 329, 10, 1, 0, 0, 328,
		330, 3, 46, 23, 0, 329, 328, 1, 0, 0, 0, 329, 330, 1, 0, 0, 0, 330, 331,
		1, 0, 0, 0, 331, 332, 5, 32, 0, 0, 332, 333, 3, 44, 22, 0, 333, 334, 5,
		33, 0, 0, 334, 335, 3, 48, 24, 0, 335, 337, 1, 0, 0, 0, 336, 327, 1, 0,
		0, 0, 337, 340, 1, 0, 0, 0, 338, 336, 1, 0, 0, 0, 338, 339, 1, 0, 0, 0,
		339, 43, 1, 0, 0, 0, 340, 338, 1, 0, 0, 0, 341, 346, 3, 86, 43, 0, 342,
		344, 5, 27, 0, 0, 343, 342, 1, 0, 0, 0, 343, 344, 1, 0, 0, 0, 344, 345,
		1, 0, 0, 0, 345, 347, 3, 88, 44, 0, 346, 343, 1, 0, 0, 0, 346, 347, 1,
		0, 0, 0, 347, 357, 1, 0, 0, 0, 348, 349, 5, 79, 0, 0, 349, 350, 3, 38,
		19, 0, 350, 352, 5, 80, 0, 0, 351, 353, 5, 27, 0, 0, 352, 351, 1, 0, 0,
		0, 352, 353, 1, 0, 0, 0, 353, 354, 1, 0, 0, 0, 354, 355, 3, 88, 44, 0,
		355, 357, 1, 0, 0, 0, 356, 341, 1, 0, 0, 0, 356, 348, 1, 0, 0, 0, 357,
		45, 1, 0, 0, 0, 358, 372, 5, 37, 0, 0, 359, 361, 5, 38, 0, 0, 360, 362,
		5, 41, 0, 0, 361, 360, 1, 0, 0, 0, 361, 362, 1, 0, 0, 0, 362, 372, 1, 0,
		0, 0, 363, 365, 5, 39, 0, 0, 364, 366, 5, 41, 0, 0, 365, 364, 1, 0, 0,
		0, 365, 366, 1, 0, 0, 0, 366, 372, 1, 0, 0, 0, 367, 369, 5, 40, 0, 0, 368,
		370, 5, 41, 0, 0, 369, 368, 1, 0, 0, 0, 369, 370, 1, 0, 0, 0, 370, 372,
		1, 0, 0, 0, 371, 358, 1, 0, 0, 0, 371, 359, 1, 0, 0, 0, 371, 363, 1, 0,
		0, 0, 371, 367, 1, 0, 0, 0, 372, 47, 1, 0, 0, 0, 373, 374, 6, 24, -1, 0,
		374, 375, 3, 50, 25, 0, 375, 409, 1, 0, 0, 0, 376, 377, 10, 7, 0, 0, 377,
		378, 7, 0, 0, 0, 378, 408, 3, 48, 24, 8, 379, 380, 10, 6, 0, 0, 380, 381,
		7, 1, 0, 0, 381, 408, 3, 48, 24, 7, 382, 383, 10, 5, 0, 0, 383, 384, 3,
		52, 26, 0, 384, 385, 3, 48, 24, 6, 385, 408, 1, 0, 0, 0, 386, 387, 10,
		4, 0, 0, 387, 388, 5, 30, 0, 0, 388, 408, 3, 48, 24, 5, 389, 390, 10, 3,
		0, 0, 390, 391, 5, 31, 0, 0, 391, 408, 3, 48, 24, 4, 392, 394, 10, 2, 0,
		0, 393, 395, 5, 23, 0, 0, 394, 393, 1, 0, 0, 0, 394, 395, 1, 0, 0, 0, 395,
		396, 1, 0, 0, 0, 396, 397, 5, 28, 0, 0, 397, 408, 3, 48, 24, 3, 398, 400,
		10, 1, 0, 0, 399, 401, 5, 23, 0, 0, 400, 399, 1, 0, 0, 0, 400, 401, 1,
		0, 0, 0, 401, 402, 1, 0, 0, 0, 402, 403, 5, 29, 0, 0, 403, 404, 5, 79,
		0, 0, 404, 405, 3, 84, 42, 0, 405, 406, 5, 80, 0, 0, 406, 408, 1, 0, 0,
		0, 407, 376, 1, 0, 0, 0, 407, 379, 1, 0, 0, 0, 407, 382, 1, 0, 0, 0, 407,
		386, 1, 0, 0, 0, 407, 389, 1, 0, 0, 0, 407, 392, 1, 0, 0, 0, 407, 398,
		1, 0, 0, 0, 408, 411, 1, 0, 0, 0, 409, 407, 1, 0, 0, 0, 409, 410, 1, 0,
		0, 0, 410, 49, 1, 0, 0, 0, 411, 409, 1, 0, 0, 0, 412, 420, 3, 92, 46, 0,
		413, 420, 3, 54, 27, 0, 414, 420, 3, 62, 31, 0, 415, 416, 5, 79, 0, 0,
		416, 417, 3, 48, 24, 0, 417, 418, 5, 80, 0, 0, 418, 420, 1, 0, 0, 0, 419,
		412, 1, 0, 0, 0, 419, 413, 1, 0, 0, 0, 419, 414, 1, 0, 0, 0, 419, 415,
		1, 0, 0, 0, 420, 51, 1, 0, 0, 0, 421, 422, 7, 2, 0, 0, 422, 53, 1, 0, 0,
		0, 423, 429, 3, 88, 44, 0, 424, 425, 3, 88, 44, 0, 425, 426, 5, 76, 0,
		0, 426, 427, 3, 88, 44, 0, 427, 429, 1, 0, 0, 0, 428, 423, 1, 0, 0, 0,
		428, 424, 1, 0, 0, 0, 429, 55, 1, 0, 0, 0, 430, 431, 3, 88, 44, 0, 431,
		432, 5, 66, 0, 0, 432, 433, 3, 48, 24, 0, 433, 57, 1, 0, 0, 0, 434, 435,
		3, 48, 24, 0, 435, 59, 1, 0, 0, 0, 436, 438, 3, 48, 24, 0, 437, 439, 7,
		3, 0, 0, 438, 437, 1, 0, 0, 0, 438, 439, 1, 0, 0, 0, 439, 61, 1, 0, 0,
		0, 440, 441, 3, 88, 44, 0, 441, 451, 5, 79, 0, 0, 442, 452, 5, 65, 0, 0,
		443, 448, 3, 48, 24, 0, 444, 445, 5, 77, 0, 0, 445, 447, 3, 48, 24, 0,
		446, 444, 1, 0, 0, 0, 447, 450, 1, 0, 0, 0, 448, 446, 1, 0, 0, 0, 448,
		449, 1, 0, 0, 0, 449, 452, 1, 0, 0, 0, 450, 448, 1, 0, 0, 0, 451, 442,
		1, 0, 0, 0, 451, 443, 1, 0, 0, 0, 451, 452, 1, 0, 0, 0, 452, 453, 1, 0,
		0, 0, 453, 454, 5, 80, 0, 0, 454, 63, 1, 0, 0, 0, 455, 456, 5, 63, 0, 0,
		456, 457, 5, 79, 0, 0, 457, 458, 3, 82, 41, 0, 458, 459, 5, 80, 0, 0, 459,
		466, 1, 0, 0, 0, 460, 461, 5, 64, 0, 0, 461, 462, 5, 79, 0, 0, 462, 463,
		3, 82, 41, 0, 463, 464, 5, 80, 0, 0, 464, 466, 1, 0, 0, 0, 465, 455, 1,
		0, 0, 0, 465, 460, 1, 0, 0, 0, 466, 65, 1, 0, 0, 0, 467, 468, 5, 59, 0,
		0, 468, 472, 5, 60, 0, 0, 469, 472, 5, 61, 0, 0, 470, 472, 5, 62, 0, 0,
		471, 467, 1, 0, 0, 0, 471, 469, 1, 0, 0, 0, 471, 470, 1, 0, 0, 0, 472,
		67, 1, 0, 0, 0, 473, 474, 5, 42, 0, 0, 474, 475, 3, 88, 44, 0, 475, 69,
		1, 0, 0, 0, 476, 477, 5, 43, 0, 0, 477, 478, 5, 44, 0, 0, 478, 71, 1, 0,
		0, 0, 479, 480, 5, 43, 0, 0, 480, 481, 5, 45, 0, 0, 481, 73, 1, 0, 0, 0,
		482, 483, 5, 43, 0, 0, 483, 484, 5, 52, 0, 0, 484, 485, 7, 4, 0, 0, 485,
		486, 3, 86, 43, 0, 486, 75, 1, 0, 0, 0, 487, 488, 5, 46, 0, 0, 488, 489,
		3, 38, 19, 0, 489, 77, 1, 0, 0, 0, 490, 491, 5, 47, 0, 0, 491, 492, 5,
		18, 0, 0, 492, 497, 3, 86, 43, 0, 493, 494, 5, 79, 0, 0, 494, 495, 3, 80,
		40, 0, 495, 496, 5, 80, 0, 0, 496, 498, 1, 0, 0, 0, 497, 493, 1, 0, 0,
		0, 497, 498, 1, 0, 0, 0, 498, 79, 1, 0, 0, 0, 499, 504, 3, 88, 44, 0, 500,
		501, 5, 77, 0, 0, 501, 503, 3, 88, 44, 0, 502, 500, 1, 0, 0, 0, 503, 506,
		1, 0, 0, 0, 504, 502, 1, 0, 0, 0, 504, 505, 1, 0, 0, 0, 505, 81, 1, 0,
		0, 0, 506, 504, 1, 0, 0, 0, 507, 512, 3, 88, 44, 0, 508, 509, 5, 77, 0,
		0, 509, 511, 3, 88, 44, 0, 510, 508, 1, 0, 0, 0, 511, 514, 1, 0, 0, 0,
		512, 510, 1, 0, 0, 0, 512, 513, 1, 0, 0, 0, 513, 83, 1, 0, 0, 0, 514, 512,
		1, 0, 0, 0, 515, 520, 3, 92, 46, 0, 516, 517, 5, 77, 0, 0, 517, 519, 3,
		92, 46, 0, 518, 516, 1, 0, 0, 0, 519, 522, 1, 0, 0, 0, 520, 518, 1, 0,
		0, 0, 520, 521, 1, 0, 0, 0, 521, 85, 1, 0, 0, 0, 522, 520, 1, 0, 0, 0,
		523, 526, 3, 88, 44, 0, 524, 525, 5, 76, 0, 0, 525, 527, 3, 88, 44, 0,
		526, 524, 1, 0, 0, 0, 526, 527, 1, 0, 0, 0, 527, 87, 1, 0, 0, 0, 528, 529,
		5, 81, 0, 0, 529, 89, 1, 0, 0, 0, 530, 542, 5, 53, 0, 0, 531, 542, 5, 54,
		0, 0, 532, 536, 5, 55, 0, 0, 533, 534, 5, 79, 0, 0, 534, 535, 5, 82, 0,
		0, 535, 537, 5, 80, 0, 0, 536, 533, 1, 0, 0, 0, 536, 537, 1, 0, 0, 0, 537,
		542, 1, 0, 0, 0, 538, 542, 5, 56, 0, 0, 539, 542, 5, 57, 0, 0, 540, 542,
		5, 58, 0, 0, 541, 530, 1, 0, 0, 0, 541, 531, 1, 0, 0, 0, 541, 532, 1, 0,
		0, 0, 541, 538, 1, 0, 0, 0, 541, 539, 1, 0, 0, 0, 541, 540, 1, 0, 0, 0,
		542, 91, 1, 0, 0, 0, 543, 544, 7, 5, 0, 0, 544, 93, 1, 0, 0, 0, 58, 97,
		107, 110, 118, 123, 135, 150, 157, 164, 171, 175, 183, 193, 224, 237, 248,
		253, 260, 268, 275, 284, 287, 291, 300, 303, 307, 312, 317, 320, 322, 329,
		338, 343, 346, 352, 356, 361, 365, 369, 371, 394, 400, 407, 409, 419, 428,
		438, 448, 451, 465, 471, 497, 504, 512, 520, 526, 536, 541,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// MiniQLParserInit initializes any static state used to implement MiniQLParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewMiniQLParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func MiniQLParserInit() {
	staticData := &MiniQLParserStaticData
	staticData.once.Do(miniqlParserInit)
}

// NewMiniQLParser produces a new parser instance for the optional input antlr.TokenStream.
func NewMiniQLParser(input antlr.TokenStream) *MiniQLParser {
	MiniQLParserInit()
	this := new(MiniQLParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &MiniQLParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "MiniQL.g4"

	return this
}

// MiniQLParser tokens.
const (
	MiniQLParserEOF                 = antlr.TokenEOF
	MiniQLParserSINGLE_LINE_COMMENT = 1
	MiniQLParserMULTI_LINE_COMMENT  = 2
	MiniQLParserSELECT              = 3
	MiniQLParserFROM                = 4
	MiniQLParserWHERE               = 5
	MiniQLParserGROUP               = 6
	MiniQLParserBY                  = 7
	MiniQLParserHAVING              = 8
	MiniQLParserORDER               = 9
	MiniQLParserLIMIT               = 10
	MiniQLParserINSERT              = 11
	MiniQLParserINTO                = 12
	MiniQLParserVALUES              = 13
	MiniQLParserUPDATE              = 14
	MiniQLParserSET                 = 15
	MiniQLParserDELETE              = 16
	MiniQLParserCREATE              = 17
	MiniQLParserTABLE               = 18
	MiniQLParserDATABASE            = 19
	MiniQLParserDROP                = 20
	MiniQLParserPRIMARY             = 21
	MiniQLParserKEY                 = 22
	MiniQLParserNOT                 = 23
	MiniQLParserNULL                = 24
	MiniQLParserTRUE                = 25
	MiniQLParserFALSE               = 26
	MiniQLParserAS                  = 27
	MiniQLParserLIKE                = 28
	MiniQLParserIN                  = 29
	MiniQLParserAND                 = 30
	MiniQLParserOR                  = 31
	MiniQLParserJOIN                = 32
	MiniQLParserON                  = 33
	MiniQLParserPARTITION           = 34
	MiniQLParserASC                 = 35
	MiniQLParserDESC                = 36
	MiniQLParserINNER               = 37
	MiniQLParserLEFT                = 38
	MiniQLParserRIGHT               = 39
	MiniQLParserFULL                = 40
	MiniQLParserOUTER               = 41
	MiniQLParserUSE                 = 42
	MiniQLParserSHOW                = 43
	MiniQLParserDATABASES           = 44
	MiniQLParserTABLES              = 45
	MiniQLParserEXPLAIN             = 46
	MiniQLParserANALYZE             = 47
	MiniQLParserVERBOSE             = 48
	MiniQLParserUNIQUE              = 49
	MiniQLParserDEFAULT             = 50
	MiniQLParserINDEX               = 51
	MiniQLParserINDEXES             = 52
	MiniQLParserINT_TYPE            = 53
	MiniQLParserINTEGER_TYPE        = 54
	MiniQLParserVARCHAR_TYPE        = 55
	MiniQLParserBOOLEAN_TYPE        = 56
	MiniQLParserDOUBLE_TYPE         = 57
	MiniQLParserTIMESTAMP_TYPE      = 58
	MiniQLParserSTART               = 59
	MiniQLParserTRANSACTION         = 60
	MiniQLParserCOMMIT              = 61
	MiniQLParserROLLBACK            = 62
	MiniQLParserHASH                = 63
	MiniQLParserRANGE               = 64
	MiniQLParserASTERISK            = 65
	MiniQLParserEQUAL               = 66
	MiniQLParserNOT_EQUAL           = 67
	MiniQLParserGREATER             = 68
	MiniQLParserGREATER_EQUAL       = 69
	MiniQLParserLESS                = 70
	MiniQLParserLESS_EQUAL          = 71
	MiniQLParserPLUS                = 72
	MiniQLParserMINUS               = 73
	MiniQLParserMULTIPLY            = 74
	MiniQLParserDIVIDE              = 75
	MiniQLParserDOT                 = 76
	MiniQLParserCOMMA               = 77
	MiniQLParserSEMICOLON           = 78
	MiniQLParserLEFT_PAREN          = 79
	MiniQLParserRIGHT_PAREN         = 80
	MiniQLParserIDENTIFIER          = 81
	MiniQLParserINTEGER_LITERAL     = 82
	MiniQLParserFLOAT_LITERAL       = 83
	MiniQLParserSTRING_LITERAL      = 84
	MiniQLParserWS                  = 85
)

// MiniQLParser rules.
const (
	MiniQLParserRULE_parse                = 0
	MiniQLParserRULE_sqlStatement         = 1
	MiniQLParserRULE_ddlStatement         = 2
	MiniQLParserRULE_dmlStatement         = 3
	MiniQLParserRULE_dqlStatement         = 4
	MiniQLParserRULE_dclStatement         = 5
	MiniQLParserRULE_utilityStatement     = 6
	MiniQLParserRULE_createDatabase       = 7
	MiniQLParserRULE_createTable          = 8
	MiniQLParserRULE_columnDef            = 9
	MiniQLParserRULE_columnConstraint     = 10
	MiniQLParserRULE_tableConstraint      = 11
	MiniQLParserRULE_createIndex          = 12
	MiniQLParserRULE_dropIndex            = 13
	MiniQLParserRULE_dropTable            = 14
	MiniQLParserRULE_dropDatabase         = 15
	MiniQLParserRULE_insertStatement      = 16
	MiniQLParserRULE_updateStatement      = 17
	MiniQLParserRULE_deleteStatement      = 18
	MiniQLParserRULE_selectStatement      = 19
	MiniQLParserRULE_selectItem           = 20
	MiniQLParserRULE_tableReference       = 21
	MiniQLParserRULE_tableReferenceAtom   = 22
	MiniQLParserRULE_joinType             = 23
	MiniQLParserRULE_expression           = 24
	MiniQLParserRULE_primaryExpr          = 25
	MiniQLParserRULE_comparisonOperator   = 26
	MiniQLParserRULE_columnRef            = 27
	MiniQLParserRULE_updateAssignment     = 28
	MiniQLParserRULE_groupByItem          = 29
	MiniQLParserRULE_orderByItem          = 30
	MiniQLParserRULE_functionCall         = 31
	MiniQLParserRULE_partitionMethod      = 32
	MiniQLParserRULE_transactionStatement = 33
	MiniQLParserRULE_useStatement         = 34
	MiniQLParserRULE_showDatabases        = 35
	MiniQLParserRULE_showTables           = 36
	MiniQLParserRULE_showIndexes          = 37
	MiniQLParserRULE_explainStatement     = 38
	MiniQLParserRULE_analyzeStatement     = 39
	MiniQLParserRULE_columnList           = 40
	MiniQLParserRULE_identifierList       = 41
	MiniQLParserRULE_valueList            = 42
	MiniQLParserRULE_tableName            = 43
	MiniQLParserRULE_identifier           = 44
	MiniQLParserRULE_dataType             = 45
	MiniQLParserRULE_literal              = 46
)

// IParseContext is an interface to support dynamic dispatch.
type IParseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
	AllSqlStatement() []ISqlStatementContext
	SqlStatement(i int) ISqlStatementContext

	// IsParseContext differentiates from other interfaces.
	IsParseContext()
}

type ParseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyParseContext() *ParseContext {
	var p = new(ParseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_parse
	return p
}

func InitEmptyParseContext(p *ParseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_parse
}

func (*ParseContext) IsParseContext() {}

func NewParseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ParseContext {
	var p = new(ParseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_parse

	return p
}

func (s *ParseContext) GetParser() antlr.Parser { return s.parser }

func (s *ParseContext) EOF() antlr.TerminalNode {
	return s.GetToken(MiniQLParserEOF, 0)
}

func (s *ParseContext) AllSqlStatement() []ISqlStatementContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISqlStatementContext); ok {
			len++
		}
	}

	tst := make([]ISqlStatementContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISqlStatementContext); ok {
			tst[i] = t.(ISqlStatementContext)
			i++
		}
	}

	return tst
}

func (s *ParseContext) SqlStatement(i int) ISqlStatementContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISqlStatementContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISqlStatementContext)
}

func (s *ParseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ParseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ParseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitParse(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) Parse() (localctx IParseContext) {
	localctx = NewParseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, MiniQLParserRULE_parse)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(97)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&7494214080317835272) != 0 {
		{
			p.SetState(94)
			p.SqlStatement()
		}

		p.SetState(99)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(100)
		p.Match(MiniQLParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISqlStatementContext is an interface to support dynamic dispatch.
type ISqlStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	DdlStatement() IDdlStatementContext
	DmlStatement() IDmlStatementContext
	DqlStatement() IDqlStatementContext
	DclStatement() IDclStatementContext
	UtilityStatement() IUtilityStatementContext
	SEMICOLON() antlr.TerminalNode

	// IsSqlStatementContext differentiates from other interfaces.
	IsSqlStatementContext()
}

type SqlStatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySqlStatementContext() *SqlStatementContext {
	var p = new(SqlStatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_sqlStatement
	return p
}

func InitEmptySqlStatementContext(p *SqlStatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_sqlStatement
}

func (*SqlStatementContext) IsSqlStatementContext() {}

func NewSqlStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SqlStatementContext {
	var p = new(SqlStatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_sqlStatement

	return p
}

func (s *SqlStatementContext) GetParser() antlr.Parser { return s.parser }

func (s *SqlStatementContext) DdlStatement() IDdlStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDdlStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDdlStatementContext)
}

func (s *SqlStatementContext) DmlStatement() IDmlStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDmlStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDmlStatementContext)
}

func (s *SqlStatementContext) DqlStatement() IDqlStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDqlStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDqlStatementContext)
}

func (s *SqlStatementContext) DclStatement() IDclStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDclStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDclStatementContext)
}

func (s *SqlStatementContext) UtilityStatement() IUtilityStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUtilityStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUtilityStatementContext)
}

func (s *SqlStatementContext) SEMICOLON() antlr.TerminalNode {
	return s.GetToken(MiniQLParserSEMICOLON, 0)
}

func (s *SqlStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SqlStatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SqlStatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitSqlStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) SqlStatement() (localctx ISqlStatementContext) {
	localctx = NewSqlStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, MiniQLParserRULE_sqlStatement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(107)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MiniQLParserCREATE, MiniQLParserDROP:
		{
			p.SetState(102)
			p.DdlStatement()
		}

	case MiniQLParserINSERT, MiniQLParserUPDATE, MiniQLParserDELETE:
		{
			p.SetState(103)
			p.DmlStatement()
		}

	case MiniQLParserSELECT:
		{
			p.SetState(104)
			p.DqlStatement()
		}

	case MiniQLParserSTART, MiniQLParserCOMMIT, MiniQLParserROLLBACK:
		{
			p.SetState(105)
			p.DclStatement()
		}

	case MiniQLParserUSE, MiniQLParserSHOW, MiniQLParserEXPLAIN, MiniQLParserANALYZE:
		{
			p.SetState(106)
			p.UtilityStatement()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	p.SetState(110)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MiniQLParserSEMICOLON {
		{
			p.SetState(109)
			p.Match(MiniQLParserSEMICOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IDdlStatementContext is an interface to support dynamic dispatch.
type IDdlStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CreateDatabase() ICreateDatabaseContext
	CreateTable() ICreateTableContext
	CreateIndex() ICreateIndexContext
	DropIndex() IDropIndexContext
	DropTable() IDropTableContext
	DropDatabase() IDropDatabaseContext

	// IsDdlStatementContext differentiates from other interfaces.
	IsDdlStatementContext()
}

type DdlStatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDdlStatementContext() *DdlStatementContext {
	var p = new(DdlStatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_ddlStatement
	return p
}

func InitEmptyDdlStatementContext(p *DdlStatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_ddlStatement
}

func (*DdlStatementContext) IsDdlStatementContext() {}

func NewDdlStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DdlStatementContext {
	var p = new(DdlStatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_ddlStatement

	return p
}

func (s *DdlStatementContext) GetParser() antlr.Parser { return s.parser }

func (s *DdlStatementContext) CreateDatabase() ICreateDatabaseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICreateDatabaseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICreateDatabaseContext)
}

func (s *DdlStatementContext) CreateTable() ICreateTableContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICreateTableContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICreateTableContext)
}

func (s *DdlStatementContext) CreateIndex() ICreateIndexContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICreateIndexContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICreateIndexContext)
}

func (s *DdlStatementContext) DropIndex() IDropIndexContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDropIndexContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDropIndexContext)
}

func (s *DdlStatementContext) DropTable() IDropTableContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDropTableContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDropTableContext)
}

func (s *DdlStatementContext) DropDatabase() IDropDatabaseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDropDatabaseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDropDatabaseContext)
}

func (s *DdlStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DdlStatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DdlStatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitDdlStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) DdlStatement() (localctx IDdlStatementContext) {
	localctx = NewDdlStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, MiniQLParserRULE_ddlStatement)
	p.SetState(118)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 3, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(112)
			p.CreateDatabase()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(113)
			p.CreateTable()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(114)
			p.CreateIndex()
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(115)
			p.DropIndex()
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(116)
			p.DropTable()
		}

	case 6:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(117)
			p.DropDatabase()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IDmlStatementContext is an interface to support dynamic dispatch.
type IDmlStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	InsertStatement() IInsertStatementContext
	UpdateStatement() IUpdateStatementContext
	DeleteStatement() IDeleteStatementContext

	// IsDmlStatementContext differentiates from other interfaces.
	IsDmlStatementContext()
}

type DmlStatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDmlStatementContext() *DmlStatementContext {
	var p = new(DmlStatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_dmlStatement
	return p
}

func InitEmptyDmlStatementContext(p *DmlStatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_dmlStatement
}

func (*DmlStatementContext) IsDmlStatementContext() {}

func NewDmlStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DmlStatementContext {
	var p = new(DmlStatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_dmlStatement

	return p
}

func (s *DmlStatementContext) GetParser() antlr.Parser { return s.parser }

func (s *DmlStatementContext) InsertStatement() IInsertStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInsertStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInsertStatementContext)
}

func (s *DmlStatementContext) UpdateStatement() IUpdateStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUpdateStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUpdateStatementContext)
}

func (s *DmlStatementContext) DeleteStatement() IDeleteStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDeleteStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDeleteStatementContext)
}

func (s *DmlStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DmlStatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DmlStatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitDmlStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) DmlStatement() (localctx IDmlStatementContext) {
	localctx = NewDmlStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, MiniQLParserRULE_dmlStatement)
	p.SetState(123)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MiniQLParserINSERT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(120)
			p.InsertStatement()
		}

	case MiniQLParserUPDATE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(121)
			p.UpdateStatement()
		}

	case MiniQLParserDELETE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(122)
			p.DeleteStatement()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IDqlStatementContext is an interface to support dynamic dispatch.
type IDqlStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SelectStatement() ISelectStatementContext

	// IsDqlStatementContext differentiates from other interfaces.
	IsDqlStatementContext()
}

type DqlStatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDqlStatementContext() *DqlStatementContext {
	var p = new(DqlStatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_dqlStatement
	return p
}

func InitEmptyDqlStatementContext(p *DqlStatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_dqlStatement
}

func (*DqlStatementContext) IsDqlStatementContext() {}

func NewDqlStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DqlStatementContext {
	var p = new(DqlStatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_dqlStatement

	return p
}

func (s *DqlStatementContext) GetParser() antlr.Parser { return s.parser }

func (s *DqlStatementContext) SelectStatement() ISelectStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelectStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelectStatementContext)
}

func (s *DqlStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DqlStatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DqlStatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitDqlStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) DqlStatement() (localctx IDqlStatementContext) {
	localctx = NewDqlStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, MiniQLParserRULE_dqlStatement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(125)
		p.SelectStatement()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IDclStatementContext is an interface to support dynamic dispatch.
type IDclStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TransactionStatement() ITransactionStatementContext

	// IsDclStatementContext differentiates from other interfaces.
	IsDclStatementContext()
}

type DclStatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDclStatementContext() *DclStatementContext {
	var p = new(DclStatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_dclStatement
	return p
}

func InitEmptyDclStatementContext(p *DclStatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_dclStatement
}

func (*DclStatementContext) IsDclStatementContext() {}

func NewDclStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DclStatementContext {
	var p = new(DclStatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_dclStatement

	return p
}

func (s *DclStatementContext) GetParser() antlr.Parser { return s.parser }

func (s *DclStatementContext) TransactionStatement() ITransactionStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITransactionStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITransactionStatementContext)
}

func (s *DclStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DclStatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DclStatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitDclStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) DclStatement() (localctx IDclStatementContext) {
	localctx = NewDclStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, MiniQLParserRULE_dclStatement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(127)
		p.TransactionStatement()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IUtilityStatementContext is an interface to support dynamic dispatch.
type IUtilityStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	UseStatement() IUseStatementContext
	ShowDatabases() IShowDatabasesContext
	ShowTables() IShowTablesContext
	ShowIndexes() IShowIndexesContext
	ExplainStatement() IExplainStatementContext
	AnalyzeStatement() IAnalyzeStatementContext

	// IsUtilityStatementContext differentiates from other interfaces.
	IsUtilityStatementContext()
}

type UtilityStatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUtilityStatementContext() *UtilityStatementContext {
	var p = new(UtilityStatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_utilityStatement
	return p
}

func InitEmptyUtilityStatementContext(p *UtilityStatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_utilityStatement
}

func (*UtilityStatementContext) IsUtilityStatementContext() {}

func NewUtilityStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *UtilityStatementContext {
	var p = new(UtilityStatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_utilityStatement

	return p
}

func (s *UtilityStatementContext) GetParser() antlr.Parser { return s.parser }

func (s *UtilityStatementContext) UseStatement() IUseStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUseStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUseStatementContext)
}

func (s *UtilityStatementContext) ShowDatabases() IShowDatabasesContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IShowDatabasesContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IShowDatabasesContext)
}

func (s *UtilityStatementContext) ShowTables() IShowTablesContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IShowTablesContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IShowTablesContext)
}

func (s *UtilityStatementContext) ShowIndexes() IShowIndexesContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IShowIndexesContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IShowIndexesContext)
}

func (s *UtilityStatementContext) ExplainStatement() IExplainStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExplainStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExplainStatementContext)
}

func (s *UtilityStatementContext) AnalyzeStatement() IAnalyzeStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAnalyzeStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAnalyzeStatementContext)
}

func (s *UtilityStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UtilityStatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *UtilityStatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitUtilityStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) UtilityStatement() (localctx IUtilityStatementContext) {
	localctx = NewUtilityStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, MiniQLParserRULE_utilityStatement)
	p.SetState(135)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(129)
			p.UseStatement()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(130)
			p.ShowDatabases()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(131)
			p.ShowTables()
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(132)
			p.ShowIndexes()
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(133)
			p.ExplainStatement()
		}

	case 6:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(134)
			p.AnalyzeStatement()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICreateDatabaseContext is an interface to support dynamic dispatch.
type ICreateDatabaseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CREATE() antlr.TerminalNode
	DATABASE() antlr.TerminalNode
	Identifier() IIdentifierContext

	// IsCreateDatabaseContext differentiates from other interfaces.
	IsCreateDatabaseContext()
}

type CreateDatabaseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCreateDatabaseContext() *CreateDatabaseContext {
	var p = new(CreateDatabaseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_createDatabase
	return p
}

func InitEmptyCreateDatabaseContext(p *CreateDatabaseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_createDatabase
}

func (*CreateDatabaseContext) IsCreateDatabaseContext() {}

func NewCreateDatabaseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CreateDatabaseContext {
	var p = new(CreateDatabaseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_createDatabase

	return p
}

func (s *CreateDatabaseContext) GetParser() antlr.Parser { return s.parser }

func (s *CreateDatabaseContext) CREATE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserCREATE, 0)
}

func (s *CreateDatabaseContext) DATABASE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDATABASE, 0)
}

func (s *CreateDatabaseContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *CreateDatabaseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CreateDatabaseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CreateDatabaseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitCreateDatabase(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) CreateDatabase() (localctx ICreateDatabaseContext) {
	localctx = NewCreateDatabaseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, MiniQLParserRULE_createDatabase)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(137)
		p.Match(MiniQLParserCREATE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(138)
		p.Match(MiniQLParserDATABASE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(139)
		p.Identifier()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICreateTableContext is an interface to support dynamic dispatch.
type ICreateTableContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CREATE() antlr.TerminalNode
	TABLE() antlr.TerminalNode
	TableName() ITableNameContext
	LEFT_PAREN() antlr.TerminalNode
	AllColumnDef() []IColumnDefContext
	ColumnDef(i int) IColumnDefContext
	RIGHT_PAREN() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	AllTableConstraint() []ITableConstraintContext
	TableConstraint(i int) ITableConstraintContext
	PARTITION() antlr.TerminalNode
	BY() antlr.TerminalNode
	PartitionMethod() IPartitionMethodContext

	// IsCreateTableContext differentiates from other interfaces.
	IsCreateTableContext()
}

type CreateTableContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCreateTableContext() *CreateTableContext {
	var p = new(CreateTableContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_createTable
	return p
}

func InitEmptyCreateTableContext(p *CreateTableContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_createTable
}

func (*CreateTableContext) IsCreateTableContext() {}

func NewCreateTableContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CreateTableContext {
	var p = new(CreateTableContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_createTable

	return p
}

func (s *CreateTableContext) GetParser() antlr.Parser { return s.parser }

func (s *CreateTableContext) CREATE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserCREATE, 0)
}

func (s *CreateTableContext) TABLE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserTABLE, 0)
}

func (s *CreateTableContext) TableName() ITableNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
}

func (s *CreateTableContext) LEFT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserLEFT_PAREN, 0)
}

func (s *CreateTableContext) AllColumnDef() []IColumnDefContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IColumnDefContext); ok {
			len++
		}
	}

	tst := make([]IColumnDefContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IColumnDefContext); ok {
			tst[i] = t.(IColumnDefContext)
			i++
		}
	}

	return tst
}

func (s *CreateTableContext) ColumnDef(i int) IColumnDefContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumnDefContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IColumnDefContext)
}

func (s *CreateTableContext) RIGHT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserRIGHT_PAREN, 0)
}

func (s *CreateTableContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MiniQLParserCOMMA)
}

func (s *CreateTableContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MiniQLParserCOMMA, i)
}

func (s *CreateTableContext) AllTableConstraint() []ITableConstraintContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITableConstraintContext); ok {
			len++
		}
	}

	tst := make([]ITableConstraintContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITableConstraintContext); ok {
			tst[i] = t.(ITableConstraintContext)
			i++
		}
	}

	return tst
}

func (s *CreateTableContext) TableConstraint(i int) ITableConstraintContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableConstraintContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableConstraintContext)
}

func (s *CreateTableContext) PARTITION() antlr.TerminalNode {
	return s.GetToken(MiniQLParserPARTITION, 0)
}

func (s *CreateTableContext) BY() antlr.TerminalNode {
	return s.GetToken(MiniQLParserBY, 0)
}

func (s *CreateTableContext) PartitionMethod() IPartitionMethodContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPartitionMethodContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPartitionMethodContext)
}

func (s *CreateTableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CreateTableContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CreateTableContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitCreateTable(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) CreateTable() (localctx ICreateTableContext) {
	localctx = NewCreateTableContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, MiniQLParserRULE_createTable)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(141)
		p.Match(MiniQLParserCREATE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(142)
		p.Match(MiniQLParserTABLE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(143)
		p.TableName()
	}
	{
		p.SetState(144)
		p.Match(MiniQLParserLEFT_PAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(145)
		p.ColumnDef()
	}
	p.SetState(150)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 6, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(146)
				p.Match(MiniQLParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(147)
				p.ColumnDef()
			}

		}
		p.SetState(152)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 6, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	p.SetState(157)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == MiniQLParserCOMMA {
		{
			p.SetState(153)
			p.Match(MiniQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(154)
			p.TableConstraint()
		}

		p.SetState(159)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(160)
		p.Match(MiniQLParserRIGHT_PAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(164)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MiniQLParserPARTITION {
		{
			p.SetState(161)
			p.Match(MiniQLParserPARTITION)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(162)
			p.Match(MiniQLParserBY)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(163)
			p.PartitionMethod()
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IColumnDefContext is an interface to support dynamic dispatch.
type IColumnDefContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext
	DataType() IDataTypeContext
	AllColumnConstraint() []IColumnConstraintContext
	ColumnConstraint(i int) IColumnConstraintContext

	// IsColumnDefContext differentiates from other interfaces.
	IsColumnDefContext()
}

type ColumnDefContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyColumnDefContext() *ColumnDefContext {
	var p = new(ColumnDefContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_columnDef
	return p
}

func InitEmptyColumnDefContext(p *ColumnDefContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_columnDef
}

func (*ColumnDefContext) IsColumnDefContext() {}

func NewColumnDefContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ColumnDefContext {
	var p = new(ColumnDefContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_columnDef

	return p
}

func (s *ColumnDefContext) GetParser() antlr.Parser { return s.parser }

func (s *ColumnDefContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *ColumnDefContext) DataType() IDataTypeContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDataTypeContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDataTypeContext)
}

func (s *ColumnDefContext) AllColumnConstraint() []IColumnConstraintContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IColumnConstraintContext); ok {
			len++
		}
	}

	tst := make([]IColumnConstraintContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IColumnConstraintContext); ok {
			tst[i] = t.(IColumnConstraintContext)
			i++
		}
	}

	return tst
}

func (s *ColumnDefContext) ColumnConstraint(i int) IColumnConstraintContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumnConstraintContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IColumnConstraintContext)
}

func (s *ColumnDefContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ColumnDefContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ColumnDefContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitColumnDef(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) ColumnDef() (localctx IColumnDefContext) {
	localctx = NewColumnDefContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, MiniQLParserRULE_columnDef)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(166)
		p.Identifier()
	}
	{
		p.SetState(167)
		p.DataType()
	}
	p.SetState(171)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&1688849887526912) != 0 {
		{
			p.SetState(168)
			p.ColumnConstraint()
		}

		p.SetState(173)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IColumnConstraintContext is an interface to support dynamic dispatch.
type IColumnConstraintContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NULL() antlr.TerminalNode
	NOT() antlr.TerminalNode
	PRIMARY() antlr.TerminalNode
	KEY() antlr.TerminalNode
	UNIQUE() antlr.TerminalNode
	DEFAULT() antlr.TerminalNode
	Literal() ILiteralContext

	// IsColumnConstraintContext differentiates from other interfaces.
	IsColumnConstraintContext()
}

type ColumnConstraintContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyColumnConstraintContext() *ColumnConstraintContext {
	var p = new(ColumnConstraintContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_columnConstraint
	return p
}

func InitEmptyColumnConstraintContext(p *ColumnConstraintContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_columnConstraint
}

func (*ColumnConstraintContext) IsColumnConstraintContext() {}

func NewColumnConstraintContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ColumnConstraintContext {
	var p = new(ColumnConstraintContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_columnConstraint

	return p
}

func (s *ColumnConstraintContext) GetParser() antlr.Parser { return s.parser }

func (s *ColumnConstraintContext) NULL() antlr.TerminalNode {
	return s.GetToken(MiniQLParserNULL, 0)
}

func (s *ColumnConstraintContext) NOT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserNOT, 0)
}

func (s *ColumnConstraintContext) PRIMARY() antlr.TerminalNode {
	return s.GetToken(MiniQLParserPRIMARY, 0)
}

func (s *ColumnConstraintContext) KEY() antlr.TerminalNode {
	return s.GetToken(MiniQLParserKEY, 0)
}

func (s *ColumnConstraintContext) UNIQUE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserUNIQUE, 0)
}

func (s *ColumnConstraintContext) DEFAULT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDEFAULT, 0)
}

func (s *ColumnConstraintContext) Literal() ILiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILiteralContext)
}

func (s *ColumnConstraintContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ColumnConstraintContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ColumnConstraintContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitColumnConstraint(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) ColumnConstraint() (localctx IColumnConstraintContext) {
	localctx = NewColumnConstraintContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, MiniQLParserRULE_columnConstraint)
	var _la int

	p.SetState(183)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MiniQLParserNOT, MiniQLParserNULL:
		p.EnterOuterAlt(localctx, 1)
		p.SetState(175)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == MiniQLParserNOT {
			{
				p.SetState(174)
				p.Match(MiniQLParserNOT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(177)
			p.Match(MiniQLParserNULL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case MiniQLParserPRIMARY:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(178)
			p.Match(MiniQLParserPRIMARY)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(179)
			p.Match(MiniQLParserKEY)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case MiniQLParserUNIQUE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(180)
			p.Match(MiniQLParserUNIQUE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case MiniQLParserDEFAULT:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(181)
			p.Match(MiniQLParserDEFAULT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(182)
			p.Literal()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITableConstraintContext is an interface to support dynamic dispatch.
type ITableConstraintContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	PRIMARY() antlr.TerminalNode
	KEY() antlr.TerminalNode
	LEFT_PAREN() antlr.TerminalNode
	IdentifierList() IIdentifierListContext
	RIGHT_PAREN() antlr.TerminalNode

	// IsTableConstraintContext differentiates from other interfaces.
	IsTableConstraintContext()
}

type TableConstraintContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTableConstraintContext() *TableConstraintContext {
	var p = new(TableConstraintContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_tableConstraint
	return p
}

func InitEmptyTableConstraintContext(p *TableConstraintContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_tableConstraint
}

func (*TableConstraintContext) IsTableConstraintContext() {}

func NewTableConstraintContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TableConstraintContext {
	var p = new(TableConstraintContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_tableConstraint

	return p
}

func (s *TableConstraintContext) GetParser() antlr.Parser { return s.parser }

func (s *TableConstraintContext) PRIMARY() antlr.TerminalNode {
	return s.GetToken(MiniQLParserPRIMARY, 0)
}

func (s *TableConstraintContext) KEY() antlr.TerminalNode {
	return s.GetToken(MiniQLParserKEY, 0)
}

func (s *TableConstraintContext) LEFT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserLEFT_PAREN, 0)
}

func (s *TableConstraintContext) IdentifierList() IIdentifierListContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierListContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierListContext)
}

func (s *TableConstraintContext) RIGHT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserRIGHT_PAREN, 0)
}

func (s *TableConstraintContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TableConstraintContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TableConstraintContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitTableConstraint(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) TableConstraint() (localctx ITableConstraintContext) {
	localctx = NewTableConstraintContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, MiniQLParserRULE_tableConstraint)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(185)
		p.Match(MiniQLParserPRIMARY)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(186)
		p.Match(MiniQLParserKEY)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(187)
		p.Match(MiniQLParserLEFT_PAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(188)
		p.IdentifierList()
	}
	{
		p.SetState(189)
		p.Match(MiniQLParserRIGHT_PAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICreateIndexContext is an interface to support dynamic dispatch.
type ICreateIndexContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CREATE() antlr.TerminalNode
	INDEX() antlr.TerminalNode
	Identifier() IIdentifierContext
	ON() antlr.TerminalNode
	TableName() ITableNameContext
	LEFT_PAREN() antlr.TerminalNode
	IdentifierList() IIdentifierListContext
	RIGHT_PAREN() antlr.TerminalNode
	UNIQUE() antlr.TerminalNode

	// IsCreateIndexContext differentiates from other interfaces.
	IsCreateIndexContext()
}

type CreateIndexContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCreateIndexContext() *CreateIndexContext {
	var p = new(CreateIndexContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_createIndex
	return p
}

func InitEmptyCreateIndexContext(p *CreateIndexContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_createIndex
}

func (*CreateIndexContext) IsCreateIndexContext() {}

func NewCreateIndexContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CreateIndexContext {
	var p = new(CreateIndexContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_createIndex

	return p
}

func (s *CreateIndexContext) GetParser() antlr.Parser { return s.parser }

func (s *CreateIndexContext) CREATE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserCREATE, 0)
}

func (s *CreateIndexContext) INDEX() antlr.TerminalNode {
	return s.GetToken(MiniQLParserINDEX, 0)
}

func (s *CreateIndexContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *CreateIndexContext) ON() antlr.TerminalNode {
	return s.GetToken(MiniQLParserON, 0)
}

func (s *CreateIndexContext) TableName() ITableNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
}

func (s *CreateIndexContext) LEFT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserLEFT_PAREN, 0)
}

func (s *CreateIndexContext) IdentifierList() IIdentifierListContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierListContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierListContext)
}

func (s *CreateIndexContext) RIGHT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserRIGHT_PAREN, 0)
}

func (s *CreateIndexContext) UNIQUE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserUNIQUE, 0)
}

func (s *CreateIndexContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CreateIndexContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CreateIndexContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitCreateIndex(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) CreateIndex() (localctx ICreateIndexContext) {
	localctx = NewCreateIndexContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, MiniQLParserRULE_createIndex)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(191)
		p.Match(MiniQLParserCREATE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(193)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MiniQLParserUNIQUE {
		{
			p.SetState(192)
			p.Match(MiniQLParserUNIQUE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(195)
		p.Match(MiniQLParserINDEX)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(196)
		p.Identifier()
	}
	{
		p.SetState(197)
		p.Match(MiniQLParserON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(198)
		p.TableName()
	}
	{
		p.SetState(199)
		p.Match(MiniQLParserLEFT_PAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(200)
		p.IdentifierList()
	}
	{
		p.SetState(201)
		p.Match(MiniQLParserRIGHT_PAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IDropIndexContext is an interface to support dynamic dispatch.
type IDropIndexContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	DROP() antlr.TerminalNode
	INDEX() antlr.TerminalNode
	Identifier() IIdentifierContext
	ON() antlr.TerminalNode
	TableName() ITableNameContext

	// IsDropIndexContext differentiates from other interfaces.
	IsDropIndexContext()
}

type DropIndexContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDropIndexContext() *DropIndexContext {
	var p = new(DropIndexContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_dropIndex
	return p
}

func InitEmptyDropIndexContext(p *DropIndexContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_dropIndex
}

func (*DropIndexContext) IsDropIndexContext() {}

func NewDropIndexContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DropIndexContext {
	var p = new(DropIndexContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_dropIndex

	return p
}

func (s *DropIndexContext) GetParser() antlr.Parser { return s.parser }

func (s *DropIndexContext) DROP() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDROP, 0)
}

func (s *DropIndexContext) INDEX() antlr.TerminalNode {
	return s.GetToken(MiniQLParserINDEX, 0)
}

func (s *DropIndexContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *DropIndexContext) ON() antlr.TerminalNode {
	return s.GetToken(MiniQLParserON, 0)
}

func (s *DropIndexContext) TableName() ITableNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
}

func (s *DropIndexContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DropIndexContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DropIndexContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitDropIndex(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) DropIndex() (localctx IDropIndexContext) {
	localctx = NewDropIndexContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, MiniQLParserRULE_dropIndex)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(203)
		p.Match(MiniQLParserDROP)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(204)
		p.Match(MiniQLParserINDEX)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(205)
		p.Identifier()
	}
	{
		p.SetState(206)
		p.Match(MiniQLParserON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(207)
		p.TableName()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IDropTableContext is an interface to support dynamic dispatch.
type IDropTableContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	DROP() antlr.TerminalNode
	TABLE() antlr.TerminalNode
	TableName() ITableNameContext

	// IsDropTableContext differentiates from other interfaces.
	IsDropTableContext()
}

type DropTableContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDropTableContext() *DropTableContext {
	var p = new(DropTableContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_dropTable
	return p
}

func InitEmptyDropTableContext(p *DropTableContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_dropTable
}

func (*DropTableContext) IsDropTableContext() {}

func NewDropTableContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DropTableContext {
	var p = new(DropTableContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_dropTable

	return p
}

func (s *DropTableContext) GetParser() antlr.Parser { return s.parser }

func (s *DropTableContext) DROP() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDROP, 0)
}

func (s *DropTableContext) TABLE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserTABLE, 0)
}

func (s *DropTableContext) TableName() ITableNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
}

func (s *DropTableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DropTableContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DropTableContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitDropTable(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) DropTable() (localctx IDropTableContext) {
	localctx = NewDropTableContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, MiniQLParserRULE_dropTable)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(209)
		p.Match(MiniQLParserDROP)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(210)
		p.Match(MiniQLParserTABLE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(211)
		p.TableName()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IDropDatabaseContext is an interface to support dynamic dispatch.
type IDropDatabaseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	DROP() antlr.TerminalNode
	DATABASE() antlr.TerminalNode
	Identifier() IIdentifierContext

	// IsDropDatabaseContext differentiates from other interfaces.
	IsDropDatabaseContext()
}

type DropDatabaseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDropDatabaseContext() *DropDatabaseContext {
	var p = new(DropDatabaseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_dropDatabase
	return p
}

func InitEmptyDropDatabaseContext(p *DropDatabaseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_dropDatabase
}

func (*DropDatabaseContext) IsDropDatabaseContext() {}

func NewDropDatabaseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DropDatabaseContext {
	var p = new(DropDatabaseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_dropDatabase

	return p
}

func (s *DropDatabaseContext) GetParser() antlr.Parser { return s.parser }

func (s *DropDatabaseContext) DROP() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDROP, 0)
}

func (s *DropDatabaseContext) DATABASE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDATABASE, 0)
}

func (s *DropDatabaseContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *DropDatabaseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DropDatabaseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DropDatabaseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitDropDatabase(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) DropDatabase() (localctx IDropDatabaseContext) {
	localctx = NewDropDatabaseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, MiniQLParserRULE_dropDatabase)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(213)
		p.Match(MiniQLParserDROP)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(214)
		p.Match(MiniQLParserDATABASE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(215)
		p.Identifier()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IInsertStatementContext is an interface to support dynamic dispatch.
type IInsertStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	INSERT() antlr.TerminalNode
	INTO() antlr.TerminalNode
	TableName() ITableNameContext
	VALUES() antlr.TerminalNode
	AllLEFT_PAREN() []antlr.TerminalNode
	LEFT_PAREN(i int) antlr.TerminalNode
	AllValueList() []IValueListContext
	ValueList(i int) IValueListContext
	AllRIGHT_PAREN() []antlr.TerminalNode
	RIGHT_PAREN(i int) antlr.TerminalNode
	IdentifierList() IIdentifierListContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsInsertStatementContext differentiates from other interfaces.
	IsInsertStatementContext()
}

type InsertStatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInsertStatementContext() *InsertStatementContext {
	var p = new(InsertStatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_insertStatement
	return p
}

func InitEmptyInsertStatementContext(p *InsertStatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_insertStatement
}

func (*InsertStatementContext) IsInsertStatementContext() {}

func NewInsertStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *InsertStatementContext {
	var p = new(InsertStatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_insertStatement

	return p
}

func (s *InsertStatementContext) GetParser() antlr.Parser { return s.parser }

func (s *InsertStatementContext) INSERT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserINSERT, 0)
}

func (s *InsertStatementContext) INTO() antlr.TerminalNode {
	return s.GetToken(MiniQLParserINTO, 0)
}

func (s *InsertStatementContext) TableName() ITableNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
}

func (s *InsertStatementContext) VALUES() antlr.TerminalNode {
	return s.GetToken(MiniQLParserVALUES, 0)
}

func (s *InsertStatementContext) AllLEFT_PAREN() []antlr.TerminalNode {
	return s.GetTokens(MiniQLParserLEFT_PAREN)
}

func (s *InsertStatementContext) LEFT_PAREN(i int) antlr.TerminalNode {
	return s.GetToken(MiniQLParserLEFT_PAREN, i)
}

func (s *InsertStatementContext) AllValueList() []IValueListContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValueListContext); ok {
			len++
		}
	}

	tst := make([]IValueListContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValueListContext); ok {
			tst[i] = t.(IValueListContext)
			i++
		}
	}

	return tst
}

func (s *InsertStatementContext) ValueList(i int) IValueListContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueListContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueListContext)
}

func (s *InsertStatementContext) AllRIGHT_PAREN() []antlr.TerminalNode {
	return s.GetTokens(MiniQLParserRIGHT_PAREN)
}

func (s *InsertStatementContext) RIGHT_PAREN(i int) antlr.TerminalNode {
	return s.GetToken(MiniQLParserRIGHT_PAREN, i)
}

func (s *InsertStatementContext) IdentifierList() IIdentifierListContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierListContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierListContext)
}

func (s *InsertStatementContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MiniQLParserCOMMA)
}

func (s *InsertStatementContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MiniQLParserCOMMA, i)
}

func (s *InsertStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InsertStatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *InsertStatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitInsertStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) InsertStatement() (localctx IInsertStatementContext) {
	localctx = NewInsertStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 32, MiniQLParserRULE_insertStatement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(217)
		p.Match(MiniQLParserINSERT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(218)
		p.Match(MiniQLParserINTO)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(219)
		p.TableName()
	}
	p.SetState(224)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MiniQLParserLEFT_PAREN {
		{
			p.SetState(220)
			p.Match(MiniQLParserLEFT_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(221)
			p.IdentifierList()
		}
		{
			p.SetState(222)
			p.Match(MiniQLParserRIGHT_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(226)
		p.Match(MiniQLParserVALUES)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(227)
		p.Match(MiniQLParserLEFT_PAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(228)
		p.ValueList()
	}
	{
		p.SetState(229)
		p.Match(MiniQLParserRIGHT_PAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(237)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == MiniQLParserCOMMA {
		{
			p.SetState(230)
			p.Match(MiniQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(231)
			p.Match(MiniQLParserLEFT_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(232)
			p.ValueList()
		}
		{
			p.SetState(233)
			p.Match(MiniQLParserRIGHT_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(239)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IUpdateStatementContext is an interface to support dynamic dispatch.
type IUpdateStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	UPDATE() antlr.TerminalNode
	TableName() ITableNameContext
	SET() antlr.TerminalNode
	AllUpdateAssignment() []IUpdateAssignmentContext
	UpdateAssignment(i int) IUpdateAssignmentContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	WHERE() antlr.TerminalNode
	Expression() IExpressionContext

	// IsUpdateStatementContext differentiates from other interfaces.
	IsUpdateStatementContext()
}

type UpdateStatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUpdateStatementContext() *UpdateStatementContext {
	var p = new(UpdateStatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_updateStatement
	return p
}

func InitEmptyUpdateStatementContext(p *UpdateStatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_updateStatement
}

func (*UpdateStatementContext) IsUpdateStatementContext() {}

func NewUpdateStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *UpdateStatementContext {
	var p = new(UpdateStatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_updateStatement

	return p
}

func (s *UpdateStatementContext) GetParser() antlr.Parser { return s.parser }

func (s *UpdateStatementContext) UPDATE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserUPDATE, 0)
}

func (s *UpdateStatementContext) TableName() ITableNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
}

func (s *UpdateStatementContext) SET() antlr.TerminalNode {
	return s.GetToken(MiniQLParserSET, 0)
}

func (s *UpdateStatementContext) AllUpdateAssignment() []IUpdateAssignmentContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IUpdateAssignmentContext); ok {
			len++
		}
	}

	tst := make([]IUpdateAssignmentContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IUpdateAssignmentContext); ok {
			tst[i] = t.(IUpdateAssignmentContext)
			i++
		}
	}

	return tst
}

func (s *UpdateStatementContext) UpdateAssignment(i int) IUpdateAssignmentContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUpdateAssignmentContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUpdateAssignmentContext)
}

func (s *UpdateStatementContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MiniQLParserCOMMA)
}

func (s *UpdateStatementContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MiniQLParserCOMMA, i)
}

func (s *UpdateStatementContext) WHERE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserWHERE, 0)
}

func (s *UpdateStatementContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *UpdateStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UpdateStatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *UpdateStatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitUpdateStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) UpdateStatement() (localctx IUpdateStatementContext) {
	localctx = NewUpdateStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 34, MiniQLParserRULE_updateStatement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(240)
		p.Match(MiniQLParserUPDATE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(241)
		p.TableName()
	}
	{
		p.SetState(242)
		p.Match(MiniQLParserSET)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(243)
		p.UpdateAssignment()
	}
	p.SetState(248)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == MiniQLParserCOMMA {
		{
			p.SetState(244)
			p.Match(MiniQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(245)
			p.UpdateAssignment()
		}

		p.SetState(250)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(253)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MiniQLParserWHERE {
		{
			p.SetState(251)
			p.Match(MiniQLParserWHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(252)
			p.expression(0)
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IDeleteStatementContext is an interface to support dynamic dispatch.
type IDeleteStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	DELETE() antlr.TerminalNode
	FROM() antlr.TerminalNode
	TableName() ITableNameContext
	WHERE() antlr.TerminalNode
	Expression() IExpressionContext

	// IsDeleteStatementContext differentiates from other interfaces.
	IsDeleteStatementContext()
}

type DeleteStatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDeleteStatementContext() *DeleteStatementContext {
	var p = new(DeleteStatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_deleteStatement
	return p
}

func InitEmptyDeleteStatementContext(p *DeleteStatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_deleteStatement
}

func (*DeleteStatementContext) IsDeleteStatementContext() {}

func NewDeleteStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DeleteStatementContext {
	var p = new(DeleteStatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_deleteStatement

	return p
}

func (s *DeleteStatementContext) GetParser() antlr.Parser { return s.parser }

func (s *DeleteStatementContext) DELETE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDELETE, 0)
}

func (s *DeleteStatementContext) FROM() antlr.TerminalNode {
	return s.GetToken(MiniQLParserFROM, 0)
}

func (s *DeleteStatementContext) TableName() ITableNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
}

func (s *DeleteStatementContext) WHERE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserWHERE, 0)
}

func (s *DeleteStatementContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *DeleteStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DeleteStatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DeleteStatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitDeleteStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) DeleteStatement() (localctx IDeleteStatementContext) {
	localctx = NewDeleteStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 36, MiniQLParserRULE_deleteStatement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(255)
		p.Match(MiniQLParserDELETE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(256)
		p.Match(MiniQLParserFROM)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(257)
		p.TableName()
	}
	p.SetState(260)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MiniQLParserWHERE {
		{
			p.SetState(258)
			p.Match(MiniQLParserWHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(259)
			p.expression(0)
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISelectStatementContext is an interface to support dynamic dispatch.
type ISelectStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SELECT() antlr.TerminalNode
	AllSelectItem() []ISelectItemContext
	SelectItem(i int) ISelectItemContext
	FROM() antlr.TerminalNode
	TableReference() ITableReferenceContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode
	WHERE() antlr.TerminalNode
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	GROUP() antlr.TerminalNode
	AllBY() []antlr.TerminalNode
	BY(i int) antlr.TerminalNode
	AllGroupByItem() []IGroupByItemContext
	GroupByItem(i int) IGroupByItemContext
	HAVING() antlr.TerminalNode
	ORDER() antlr.TerminalNode
	AllOrderByItem() []IOrderByItemContext
	OrderByItem(i int) IOrderByItemContext
	LIMIT() antlr.TerminalNode
	INTEGER_LITERAL() antlr.TerminalNode

	// IsSelectStatementContext differentiates from other interfaces.
	IsSelectStatementContext()
}

type SelectStatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySelectStatementContext() *SelectStatementContext {
	var p = new(SelectStatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_selectStatement
	return p
}

func InitEmptySelectStatementContext(p *SelectStatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_selectStatement
}

func (*SelectStatementContext) IsSelectStatementContext() {}

func NewSelectStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SelectStatementContext {
	var p = new(SelectStatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_selectStatement

	return p
}

func (s *SelectStatementContext) GetParser() antlr.Parser { return s.parser }

func (s *SelectStatementContext) SELECT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserSELECT, 0)
}

func (s *SelectStatementContext) AllSelectItem() []ISelectItemContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISelectItemContext); ok {
			len++
		}
	}

	tst := make([]ISelectItemContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISelectItemContext); ok {
			tst[i] = t.(ISelectItemContext)
			i++
		}
	}

	return tst
}

func (s *SelectStatementContext) SelectItem(i int) ISelectItemContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelectItemContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelectItemContext)
}

func (s *SelectStatementContext) FROM() antlr.TerminalNode {
	return s.GetToken(MiniQLParserFROM, 0)
}

func (s *SelectStatementContext) TableReference() ITableReferenceContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableReferenceContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableReferenceContext)
}

func (s *SelectStatementContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MiniQLParserCOMMA)
}

func (s *SelectStatementContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MiniQLParserCOMMA, i)
}

func (s *SelectStatementContext) WHERE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserWHERE, 0)
}

func (s *SelectStatementContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *SelectStatementContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *SelectStatementContext) GROUP() antlr.TerminalNode {
	return s.GetToken(MiniQLParserGROUP, 0)
}

func (s *SelectStatementContext) AllBY() []antlr.TerminalNode {
	return s.GetTokens(MiniQLParserBY)
}

func (s *SelectStatementContext) BY(i int) antlr.TerminalNode {
	return s.GetToken(MiniQLParserBY, i)
}

func (s *SelectStatementContext) AllGroupByItem() []IGroupByItemContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IGroupByItemContext); ok {
			len++
		}
	}

	tst := make([]IGroupByItemContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IGroupByItemContext); ok {
			tst[i] = t.(IGroupByItemContext)
			i++
		}
	}

	return tst
}

func (s *SelectStatementContext) GroupByItem(i int) IGroupByItemContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IGroupByItemContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IGroupByItemContext)
}

func (s *SelectStatementContext) HAVING() antlr.TerminalNode {
	return s.GetToken(MiniQLParserHAVING, 0)
}

func (s *SelectStatementContext) ORDER() antlr.TerminalNode {
	return s.GetToken(MiniQLParserORDER, 0)
}

func (s *SelectStatementContext) AllOrderByItem() []IOrderByItemContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IOrderByItemContext); ok {
			len++
		}
	}

	tst := make([]IOrderByItemContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IOrderByItemContext); ok {
			tst[i] = t.(IOrderByItemContext)
			i++
		}
	}

	return tst
}

func (s *SelectStatementContext) OrderByItem(i int) IOrderByItemContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrderByItemContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOrderByItemContext)
}

func (s *SelectStatementContext) LIMIT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserLIMIT, 0)
}

func (s *SelectStatementContext) INTEGER_LITERAL() antlr.TerminalNode {
	return s.GetToken(MiniQLParserINTEGER_LITERAL, 0)
}

func (s *SelectStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SelectStatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SelectStatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitSelectStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) SelectStatement() (localctx ISelectStatementContext) {
	localctx = NewSelectStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 38, MiniQLParserRULE_selectStatement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(262)
		p.Match(MiniQLParserSELECT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(263)
		p.SelectItem()
	}
	p.SetState(268)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == MiniQLParserCOMMA {
		{
			p.SetState(264)
			p.Match(MiniQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(265)
			p.SelectItem()
		}

		p.SetState(270)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(271)
		p.Match(MiniQLParserFROM)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(272)
		p.tableReference(0)
	}
	p.SetState(275)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MiniQLParserWHERE {
		{
			p.SetState(273)
			p.Match(MiniQLParserWHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(274)
			p.expression(0)
		}

	}
	p.SetState(287)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MiniQLParserGROUP {
		{
			p.SetState(277)
			p.Match(MiniQLParserGROUP)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(278)
			p.Match(MiniQLParserBY)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(279)
			p.GroupByItem()
		}
		p.SetState(284)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == MiniQLParserCOMMA {
			{
				p.SetState(280)
				p.Match(MiniQLParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(281)
				p.GroupByItem()
			}

			p.SetState(286)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	}
	p.SetState(291)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MiniQLParserHAVING {
		{
			p.SetState(289)
			p.Match(MiniQLParserHAVING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(290)
			p.expression(0)
		}

	}
	p.SetState(303)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MiniQLParserORDER {
		{
			p.SetState(293)
			p.Match(MiniQLParserORDER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(294)
			p.Match(MiniQLParserBY)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(295)
			p.OrderByItem()
		}
		p.SetState(300)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == MiniQLParserCOMMA {
			{
				p.SetState(296)
				p.Match(MiniQLParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(297)
				p.OrderByItem()
			}

			p.SetState(302)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	}
	p.SetState(307)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MiniQLParserLIMIT {
		{
			p.SetState(305)
			p.Match(MiniQLParserLIMIT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(306)
			p.Match(MiniQLParserINTEGER_LITERAL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISelectItemContext is an interface to support dynamic dispatch.
type ISelectItemContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsSelectItemContext differentiates from other interfaces.
	IsSelectItemContext()
}

type SelectItemContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySelectItemContext() *SelectItemContext {
	var p = new(SelectItemContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_selectItem
	return p
}

func InitEmptySelectItemContext(p *SelectItemContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_selectItem
}

func (*SelectItemContext) IsSelectItemContext() {}

func NewSelectItemContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SelectItemContext {
	var p = new(SelectItemContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_selectItem

	return p
}

func (s *SelectItemContext) GetParser() antlr.Parser { return s.parser }

func (s *SelectItemContext) CopyAll(ctx *SelectItemContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *SelectItemContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SelectItemContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type SelectAllContext struct {
	SelectItemContext
}

func NewSelectAllContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SelectAllContext {
	var p = new(SelectAllContext)

	InitEmptySelectItemContext(&p.SelectItemContext)
	p.parser = parser
	p.CopyAll(ctx.(*SelectItemContext))

	return p
}

func (s *SelectAllContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SelectAllContext) ASTERISK() antlr.TerminalNode {
	return s.GetToken(MiniQLParserASTERISK, 0)
}

func (s *SelectAllContext) TableName() ITableNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
}

func (s *SelectAllContext) DOT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDOT, 0)
}

func (s *SelectAllContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitSelectAll(s)

	default:
		return t.VisitChildren(s)
	}
}

type SelectExprContext struct {
	SelectItemContext
}

func NewSelectExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *SelectExprContext {
	var p = new(SelectExprContext)

	InitEmptySelectItemContext(&p.SelectItemContext)
	p.parser = parser
	p.CopyAll(ctx.(*SelectItemContext))

	return p
}

func (s *SelectExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SelectExprContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *SelectExprContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *SelectExprContext) AS() antlr.TerminalNode {
	return s.GetToken(MiniQLParserAS, 0)
}

func (s *SelectExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitSelectExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) SelectItem() (localctx ISelectItemContext) {
	localctx = NewSelectItemContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 40, MiniQLParserRULE_selectItem)
	var _la int

	p.SetState(322)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 29, p.GetParserRuleContext()) {
	case 1:
		localctx = NewSelectAllContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		p.SetState(312)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == MiniQLParserIDENTIFIER {
			{
				p.SetState(309)
				p.TableName()
			}
			{
				p.SetState(310)
				p.Match(MiniQLParserDOT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(314)
			p.Match(MiniQLParserASTERISK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewSelectExprContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(315)
			p.expression(0)
		}
		p.SetState(320)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == MiniQLParserAS || _la == MiniQLParserIDENTIFIER {
			p.SetState(317)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			if _la == MiniQLParserAS {
				{
					p.SetState(316)
					p.Match(MiniQLParserAS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			{
				p.SetState(319)
				p.Identifier()
			}

		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITableReferenceContext is an interface to support dynamic dispatch.
type ITableReferenceContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	TableReferenceAtom() ITableReferenceAtomContext
	TableReference() ITableReferenceContext
	JOIN() antlr.TerminalNode
	ON() antlr.TerminalNode
	Expression() IExpressionContext
	JoinType() IJoinTypeContext

	// IsTableReferenceContext differentiates from other interfaces.
	IsTableReferenceContext()
}

type TableReferenceContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTableReferenceContext() *TableReferenceContext {
	var p = new(TableReferenceContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_tableReference
	return p
}

func InitEmptyTableReferenceContext(p *TableReferenceContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_tableReference
}

func (*TableReferenceContext) IsTableReferenceContext() {}

func NewTableReferenceContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TableReferenceContext {
	var p = new(TableReferenceContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_tableReference

	return p
}

func (s *TableReferenceContext) GetParser() antlr.Parser { return s.parser }

func (s *TableReferenceContext) TableReferenceAtom() ITableReferenceAtomContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableReferenceAtomContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableReferenceAtomContext)
}

func (s *TableReferenceContext) TableReference() ITableReferenceContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableReferenceContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableReferenceContext)
}

func (s *TableReferenceContext) JOIN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserJOIN, 0)
}

func (s *TableReferenceContext) ON() antlr.TerminalNode {
	return s.GetToken(MiniQLParserON, 0)
}

func (s *TableReferenceContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *TableReferenceContext) JoinType() IJoinTypeContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IJoinTypeContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IJoinTypeContext)
}

func (s *TableReferenceContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TableReferenceContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TableReferenceContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitTableReference(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) TableReference() (localctx ITableReferenceContext) {
	return p.tableReference(0)
}

func (p *MiniQLParser) tableReference(_p int) (localctx ITableReferenceContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewTableReferenceContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx ITableReferenceContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 42
	p.EnterRecursionRule(localctx, 42, MiniQLParserRULE_tableReference, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(325)
		p.TableReferenceAtom()
	}

	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(338)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 31, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			localctx = NewTableReferenceContext(p, _parentctx, _parentState)
			p.PushNewRecursionContext(localctx, _startState, MiniQLParserRULE_tableReference)
			p.SetState(327)

			if !(p.Precpred(p.GetParserRuleContext(), 1)) {
				p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 1)", ""))
				goto errorExit
			}
			p.SetState(329)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2061584302080) != 0 {
				{
					p.SetState(328)
					p.JoinType()
				}

			}
			{
				p.SetState(331)
				p.Match(MiniQLParserJOIN)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(332)
				p.TableReferenceAtom()
			}
			{
				p.SetState(333)
				p.Match(MiniQLParserON)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(334)
				p.expression(0)
			}

		}
		p.SetState(340)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 31, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITableReferenceAtomContext is an interface to support dynamic dispatch.
type ITableReferenceAtomContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsTableReferenceAtomContext differentiates from other interfaces.
	IsTableReferenceAtomContext()
}

type TableReferenceAtomContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTableReferenceAtomContext() *TableReferenceAtomContext {
	var p = new(TableReferenceAtomContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_tableReferenceAtom
	return p
}

func InitEmptyTableReferenceAtomContext(p *TableReferenceAtomContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_tableReferenceAtom
}

func (*TableReferenceAtomContext) IsTableReferenceAtomContext() {}

func NewTableReferenceAtomContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TableReferenceAtomContext {
	var p = new(TableReferenceAtomContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_tableReferenceAtom

	return p
}

func (s *TableReferenceAtomContext) GetParser() antlr.Parser { return s.parser }

func (s *TableReferenceAtomContext) CopyAll(ctx *TableReferenceAtomContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *TableReferenceAtomContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TableReferenceAtomContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type TableRefBaseContext struct {
	TableReferenceAtomContext
}

func NewTableRefBaseContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *TableRefBaseContext {
	var p = new(TableRefBaseContext)

	InitEmptyTableReferenceAtomContext(&p.TableReferenceAtomContext)
	p.parser = parser
	p.CopyAll(ctx.(*TableReferenceAtomContext))

	return p
}

func (s *TableRefBaseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TableRefBaseContext) TableName() ITableNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
}

func (s *TableRefBaseContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *TableRefBaseContext) AS() antlr.TerminalNode {
	return s.GetToken(MiniQLParserAS, 0)
}

func (s *TableRefBaseContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitTableRefBase(s)

	default:
		return t.VisitChildren(s)
	}
}

type TableRefSubqueryContext struct {
	TableReferenceAtomContext
}

func NewTableRefSubqueryContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *TableRefSubqueryContext {
	var p = new(TableRefSubqueryContext)

	InitEmptyTableReferenceAtomContext(&p.TableReferenceAtomContext)
	p.parser = parser
	p.CopyAll(ctx.(*TableReferenceAtomContext))

	return p
}

func (s *TableRefSubqueryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TableRefSubqueryContext) LEFT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserLEFT_PAREN, 0)
}

func (s *TableRefSubqueryContext) SelectStatement() ISelectStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelectStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelectStatementContext)
}

func (s *TableRefSubqueryContext) RIGHT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserRIGHT_PAREN, 0)
}

func (s *TableRefSubqueryContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *TableRefSubqueryContext) AS() antlr.TerminalNode {
	return s.GetToken(MiniQLParserAS, 0)
}

func (s *TableRefSubqueryContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitTableRefSubquery(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) TableReferenceAtom() (localctx ITableReferenceAtomContext) {
	localctx = NewTableReferenceAtomContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 44, MiniQLParserRULE_tableReferenceAtom)
	var _la int

	p.SetState(356)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MiniQLParserIDENTIFIER:
		localctx = NewTableRefBaseContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(341)
			p.TableName()
		}
		p.SetState(346)
		p.GetErrorHandler().Sync(p)

		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 33, p.GetParserRuleContext()) == 1 {
			p.SetState(343)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			if _la == MiniQLParserAS {
				{
					p.SetState(342)
					p.Match(MiniQLParserAS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			{
				p.SetState(345)
				p.Identifier()
			}

		} else if p.HasError() { // JIM
			goto errorExit
		}

	case MiniQLParserLEFT_PAREN:
		localctx = NewTableRefSubqueryContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(348)
			p.Match(MiniQLParserLEFT_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(349)
			p.SelectStatement()
		}
		{
			p.SetState(350)
			p.Match(MiniQLParserRIGHT_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(352)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == MiniQLParserAS {
			{
				p.SetState(351)
				p.Match(MiniQLParserAS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(354)
			p.Identifier()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IJoinTypeContext is an interface to support dynamic dispatch.
type IJoinTypeContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	INNER() antlr.TerminalNode
	LEFT() antlr.TerminalNode
	OUTER() antlr.TerminalNode
	RIGHT() antlr.TerminalNode
	FULL() antlr.TerminalNode

	// IsJoinTypeContext differentiates from other interfaces.
	IsJoinTypeContext()
}

type JoinTypeContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyJoinTypeContext() *JoinTypeContext {
	var p = new(JoinTypeContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_joinType
	return p
}

func InitEmptyJoinTypeContext(p *JoinTypeContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_joinType
}

func (*JoinTypeContext) IsJoinTypeContext() {}

func NewJoinTypeContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *JoinTypeContext {
	var p = new(JoinTypeContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_joinType

	return p
}

func (s *JoinTypeContext) GetParser() antlr.Parser { return s.parser }

func (s *JoinTypeContext) INNER() antlr.TerminalNode {
	return s.GetToken(MiniQLParserINNER, 0)
}

func (s *JoinTypeContext) LEFT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserLEFT, 0)
}

func (s *JoinTypeContext) OUTER() antlr.TerminalNode {
	return s.GetToken(MiniQLParserOUTER, 0)
}

func (s *JoinTypeContext) RIGHT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserRIGHT, 0)
}

func (s *JoinTypeContext) FULL() antlr.TerminalNode {
	return s.GetToken(MiniQLParserFULL, 0)
}

func (s *JoinTypeContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *JoinTypeContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *JoinTypeContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitJoinType(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) JoinType() (localctx IJoinTypeContext) {
	localctx = NewJoinTypeContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 46, MiniQLParserRULE_joinType)
	var _la int

	p.SetState(371)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MiniQLParserINNER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(358)
			p.Match(MiniQLParserINNER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case MiniQLParserLEFT:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(359)
			p.Match(MiniQLParserLEFT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(361)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == MiniQLParserOUTER {
			{
				p.SetState(360)
				p.Match(MiniQLParserOUTER)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}

	case MiniQLParserRIGHT:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(363)
			p.Match(MiniQLParserRIGHT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(365)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == MiniQLParserOUTER {
			{
				p.SetState(364)
				p.Match(MiniQLParserOUTER)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}

	case MiniQLParserFULL:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(367)
			p.Match(MiniQLParserFULL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(369)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == MiniQLParserOUTER {
			{
				p.SetState(368)
				p.Match(MiniQLParserOUTER)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExpressionContext is an interface to support dynamic dispatch.
type IExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsExpressionContext differentiates from other interfaces.
	IsExpressionContext()
}

type ExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpressionContext() *ExpressionContext {
	var p = new(ExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_expression
	return p
}

func InitEmptyExpressionContext(p *ExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_expression
}

func (*ExpressionContext) IsExpressionContext() {}

func NewExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionContext {
	var p = new(ExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_expression

	return p
}

func (s *ExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionContext) CopyAll(ctx *ExpressionContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type PrimaryExpressionContext struct {
	ExpressionContext
}

func NewPrimaryExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *PrimaryExpressionContext {
	var p = new(PrimaryExpressionContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *PrimaryExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PrimaryExpressionContext) PrimaryExpr() IPrimaryExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPrimaryExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPrimaryExprContext)
}

func (s *PrimaryExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitPrimaryExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

type OrExpressionContext struct {
	ExpressionContext
}

func NewOrExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *OrExpressionContext {
	var p = new(OrExpressionContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *OrExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OrExpressionContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *OrExpressionContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *OrExpressionContext) OR() antlr.TerminalNode {
	return s.GetToken(MiniQLParserOR, 0)
}

func (s *OrExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitOrExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

type AndExpressionContext struct {
	ExpressionContext
}

func NewAndExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AndExpressionContext {
	var p = new(AndExpressionContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *AndExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AndExpressionContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *AndExpressionContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *AndExpressionContext) AND() antlr.TerminalNode {
	return s.GetToken(MiniQLParserAND, 0)
}

func (s *AndExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitAndExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

type InExpressionContext struct {
	ExpressionContext
}

func NewInExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *InExpressionContext {
	var p = new(InExpressionContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *InExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InExpressionContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *InExpressionContext) IN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserIN, 0)
}

func (s *InExpressionContext) LEFT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserLEFT_PAREN, 0)
}

func (s *InExpressionContext) ValueList() IValueListContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueListContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueListContext)
}

func (s *InExpressionContext) RIGHT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserRIGHT_PAREN, 0)
}

func (s *InExpressionContext) NOT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserNOT, 0)
}

func (s *InExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitInExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

type AdditiveExpressionContext struct {
	ExpressionContext
}

func NewAdditiveExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AdditiveExpressionContext {
	var p = new(AdditiveExpressionContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *AdditiveExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AdditiveExpressionContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *AdditiveExpressionContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *AdditiveExpressionContext) PLUS() antlr.TerminalNode {
	return s.GetToken(MiniQLParserPLUS, 0)
}

func (s *AdditiveExpressionContext) MINUS() antlr.TerminalNode {
	return s.GetToken(MiniQLParserMINUS, 0)
}

func (s *AdditiveExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitAdditiveExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

type LikeExpressionContext struct {
	ExpressionContext
}

func NewLikeExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LikeExpressionContext {
	var p = new(LikeExpressionContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *LikeExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LikeExpressionContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *LikeExpressionContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *LikeExpressionContext) LIKE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserLIKE, 0)
}

func (s *LikeExpressionContext) NOT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserNOT, 0)
}

func (s *LikeExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitLikeExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

type ComparisonExpressionContext struct {
	ExpressionContext
}

func NewComparisonExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ComparisonExpressionContext {
	var p = new(ComparisonExpressionContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *ComparisonExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ComparisonExpressionContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *ComparisonExpressionContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ComparisonExpressionContext) ComparisonOperator() IComparisonOperatorContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IComparisonOperatorContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IComparisonOperatorContext)
}

func (s *ComparisonExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitComparisonExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

type MultiplicativeExpressionContext struct {
	ExpressionContext
}

func NewMultiplicativeExpressionContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MultiplicativeExpressionContext {
	var p = new(MultiplicativeExpressionContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *MultiplicativeExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MultiplicativeExpressionContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *MultiplicativeExpressionContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *MultiplicativeExpressionContext) ASTERISK() antlr.TerminalNode {
	return s.GetToken(MiniQLParserASTERISK, 0)
}

func (s *MultiplicativeExpressionContext) DIVIDE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDIVIDE, 0)
}

func (s *MultiplicativeExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitMultiplicativeExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) Expression() (localctx IExpressionContext) {
	return p.expression(0)
}

func (p *MiniQLParser) expression(_p int) (localctx IExpressionContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewExpressionContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IExpressionContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 48
	p.EnterRecursionRule(localctx, 48, MiniQLParserRULE_expression, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	localctx = NewPrimaryExpressionContext(p, localctx)
	p.SetParserRuleContext(localctx)
	_prevctx = localctx

	{
		p.SetState(374)
		p.PrimaryExpr()
	}

	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(409)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 43, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(407)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 42, p.GetParserRuleContext()) {
			case 1:
				localctx = NewMultiplicativeExpressionContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, MiniQLParserRULE_expression)
				p.SetState(376)

				if !(p.Precpred(p.GetParserRuleContext(), 7)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 7)", ""))
					goto errorExit
				}
				{
					p.SetState(377)
					_la = p.GetTokenStream().LA(1)

					if !(_la == MiniQLParserASTERISK || _la == MiniQLParserDIVIDE) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(378)
					p.expression(8)
				}

			case 2:
				localctx = NewAdditiveExpressionContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, MiniQLParserRULE_expression)
				p.SetState(379)

				if !(p.Precpred(p.GetParserRuleContext(), 6)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 6)", ""))
					goto errorExit
				}
				{
					p.SetState(380)
					_la = p.GetTokenStream().LA(1)

					if !(_la == MiniQLParserPLUS || _la == MiniQLParserMINUS) {
						p.GetErrorHandler().RecoverInline(p)
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(381)
					p.expression(7)
				}

			case 3:
				localctx = NewComparisonExpressionContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, MiniQLParserRULE_expression)
				p.SetState(382)

				if !(p.Precpred(p.GetParserRuleContext(), 5)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 5)", ""))
					goto errorExit
				}
				{
					p.SetState(383)
					p.ComparisonOperator()
				}
				{
					p.SetState(384)
					p.expression(6)
				}

			case 4:
				localctx = NewAndExpressionContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, MiniQLParserRULE_expression)
				p.SetState(386)

				if !(p.Precpred(p.GetParserRuleContext(), 4)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 4)", ""))
					goto errorExit
				}
				{
					p.SetState(387)
					p.Match(MiniQLParserAND)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(388)
					p.expression(5)
				}

			case 5:
				localctx = NewOrExpressionContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, MiniQLParserRULE_expression)
				p.SetState(389)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
					goto errorExit
				}
				{
					p.SetState(390)
					p.Match(MiniQLParserOR)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(391)
					p.expression(4)
				}

			case 6:
				localctx = NewLikeExpressionContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, MiniQLParserRULE_expression)
				p.SetState(392)

				if !(p.Precpred(p.GetParserRuleContext(), 2)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
					goto errorExit
				}
				p.SetState(394)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)

				if _la == MiniQLParserNOT {
					{
						p.SetState(393)
						p.Match(MiniQLParserNOT)
						if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
						}
					}

				}
				{
					p.SetState(396)
					p.Match(MiniQLParserLIKE)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(397)
					p.expression(3)
				}

			case 7:
				localctx = NewInExpressionContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, MiniQLParserRULE_expression)
				p.SetState(398)

				if !(p.Precpred(p.GetParserRuleContext(), 1)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 1)", ""))
					goto errorExit
				}
				p.SetState(400)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)

				if _la == MiniQLParserNOT {
					{
						p.SetState(399)
						p.Match(MiniQLParserNOT)
						if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
						}
					}

				}
				{
					p.SetState(402)
					p.Match(MiniQLParserIN)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(403)
					p.Match(MiniQLParserLEFT_PAREN)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(404)
					p.ValueList()
				}
				{
					p.SetState(405)
					p.Match(MiniQLParserRIGHT_PAREN)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(411)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 43, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPrimaryExprContext is an interface to support dynamic dispatch.
type IPrimaryExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsPrimaryExprContext differentiates from other interfaces.
	IsPrimaryExprContext()
}

type PrimaryExprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPrimaryExprContext() *PrimaryExprContext {
	var p = new(PrimaryExprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_primaryExpr
	return p
}

func InitEmptyPrimaryExprContext(p *PrimaryExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_primaryExpr
}

func (*PrimaryExprContext) IsPrimaryExprContext() {}

func NewPrimaryExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PrimaryExprContext {
	var p = new(PrimaryExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_primaryExpr

	return p
}

func (s *PrimaryExprContext) GetParser() antlr.Parser { return s.parser }

func (s *PrimaryExprContext) CopyAll(ctx *PrimaryExprContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *PrimaryExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PrimaryExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type LiteralExprContext struct {
	PrimaryExprContext
}

func NewLiteralExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LiteralExprContext {
	var p = new(LiteralExprContext)

	InitEmptyPrimaryExprContext(&p.PrimaryExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*PrimaryExprContext))

	return p
}

func (s *LiteralExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LiteralExprContext) Literal() ILiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILiteralContext)
}

func (s *LiteralExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitLiteralExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type FunctionCallExprContext struct {
	PrimaryExprContext
}

func NewFunctionCallExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *FunctionCallExprContext {
	var p = new(FunctionCallExprContext)

	InitEmptyPrimaryExprContext(&p.PrimaryExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*PrimaryExprContext))

	return p
}

func (s *FunctionCallExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FunctionCallExprContext) FunctionCall() IFunctionCallContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionCallContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionCallContext)
}

func (s *FunctionCallExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitFunctionCallExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type ColumnRefExprContext struct {
	PrimaryExprContext
}

func NewColumnRefExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ColumnRefExprContext {
	var p = new(ColumnRefExprContext)

	InitEmptyPrimaryExprContext(&p.PrimaryExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*PrimaryExprContext))

	return p
}

func (s *ColumnRefExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ColumnRefExprContext) ColumnRef() IColumnRefContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumnRefContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IColumnRefContext)
}

func (s *ColumnRefExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitColumnRefExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type ParenExprContext struct {
	PrimaryExprContext
}

func NewParenExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ParenExprContext {
	var p = new(ParenExprContext)

	InitEmptyPrimaryExprContext(&p.PrimaryExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*PrimaryExprContext))

	return p
}

func (s *ParenExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ParenExprContext) LEFT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserLEFT_PAREN, 0)
}

func (s *ParenExprContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ParenExprContext) RIGHT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserRIGHT_PAREN, 0)
}

func (s *ParenExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitParenExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) PrimaryExpr() (localctx IPrimaryExprContext) {
	localctx = NewPrimaryExprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 50, MiniQLParserRULE_primaryExpr)
	p.SetState(419)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 44, p.GetParserRuleContext()) {
	case 1:
		localctx = NewLiteralExprContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(412)
			p.Literal()
		}

	case 2:
		localctx = NewColumnRefExprContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(413)
			p.ColumnRef()
		}

	case 3:
		localctx = NewFunctionCallExprContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(414)
			p.FunctionCall()
		}

	case 4:
		localctx = NewParenExprContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(415)
			p.Match(MiniQLParserLEFT_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(416)
			p.expression(0)
		}
		{
			p.SetState(417)
			p.Match(MiniQLParserRIGHT_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IComparisonOperatorContext is an interface to support dynamic dispatch.
type IComparisonOperatorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EQUAL() antlr.TerminalNode
	NOT_EQUAL() antlr.TerminalNode
	GREATER() antlr.TerminalNode
	GREATER_EQUAL() antlr.TerminalNode
	LESS() antlr.TerminalNode
	LESS_EQUAL() antlr.TerminalNode

	// IsComparisonOperatorContext differentiates from other interfaces.
	IsComparisonOperatorContext()
}

type ComparisonOperatorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyComparisonOperatorContext() *ComparisonOperatorContext {
	var p = new(ComparisonOperatorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_comparisonOperator
	return p
}

func InitEmptyComparisonOperatorContext(p *ComparisonOperatorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_comparisonOperator
}

func (*ComparisonOperatorContext) IsComparisonOperatorContext() {}

func NewComparisonOperatorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ComparisonOperatorContext {
	var p = new(ComparisonOperatorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_comparisonOperator

	return p
}

func (s *ComparisonOperatorContext) GetParser() antlr.Parser { return s.parser }

func (s *ComparisonOperatorContext) EQUAL() antlr.TerminalNode {
	return s.GetToken(MiniQLParserEQUAL, 0)
}

func (s *ComparisonOperatorContext) NOT_EQUAL() antlr.TerminalNode {
	return s.GetToken(MiniQLParserNOT_EQUAL, 0)
}

func (s *ComparisonOperatorContext) GREATER() antlr.TerminalNode {
	return s.GetToken(MiniQLParserGREATER, 0)
}

func (s *ComparisonOperatorContext) GREATER_EQUAL() antlr.TerminalNode {
	return s.GetToken(MiniQLParserGREATER_EQUAL, 0)
}

func (s *ComparisonOperatorContext) LESS() antlr.TerminalNode {
	return s.GetToken(MiniQLParserLESS, 0)
}

func (s *ComparisonOperatorContext) LESS_EQUAL() antlr.TerminalNode {
	return s.GetToken(MiniQLParserLESS_EQUAL, 0)
}

func (s *ComparisonOperatorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ComparisonOperatorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ComparisonOperatorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitComparisonOperator(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) ComparisonOperator() (localctx IComparisonOperatorContext) {
	localctx = NewComparisonOperatorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 52, MiniQLParserRULE_comparisonOperator)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(421)
		_la = p.GetTokenStream().LA(1)

		if !((int64((_la-66)) & ^0x3f) == 0 && ((int64(1)<<(_la-66))&63) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IColumnRefContext is an interface to support dynamic dispatch.
type IColumnRefContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllIdentifier() []IIdentifierContext
	Identifier(i int) IIdentifierContext
	DOT() antlr.TerminalNode

	// IsColumnRefContext differentiates from other interfaces.
	IsColumnRefContext()
}

type ColumnRefContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyColumnRefContext() *ColumnRefContext {
	var p = new(ColumnRefContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_columnRef
	return p
}

func InitEmptyColumnRefContext(p *ColumnRefContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_columnRef
}

func (*ColumnRefContext) IsColumnRefContext() {}

func NewColumnRefContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ColumnRefContext {
	var p = new(ColumnRefContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_columnRef

	return p
}

func (s *ColumnRefContext) GetParser() antlr.Parser { return s.parser }

func (s *ColumnRefContext) AllIdentifier() []IIdentifierContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIdentifierContext); ok {
			len++
		}
	}

	tst := make([]IIdentifierContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIdentifierContext); ok {
			tst[i] = t.(IIdentifierContext)
			i++
		}
	}

	return tst
}

func (s *ColumnRefContext) Identifier(i int) IIdentifierContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *ColumnRefContext) DOT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDOT, 0)
}

func (s *ColumnRefContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ColumnRefContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ColumnRefContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitColumnRef(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) ColumnRef() (localctx IColumnRefContext) {
	localctx = NewColumnRefContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 54, MiniQLParserRULE_columnRef)
	p.SetState(428)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 45, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(423)
			p.Identifier()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(424)
			p.Identifier()
		}
		{
			p.SetState(425)
			p.Match(MiniQLParserDOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(426)
			p.Identifier()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IUpdateAssignmentContext is an interface to support dynamic dispatch.
type IUpdateAssignmentContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext
	EQUAL() antlr.TerminalNode
	Expression() IExpressionContext

	// IsUpdateAssignmentContext differentiates from other interfaces.
	IsUpdateAssignmentContext()
}

type UpdateAssignmentContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUpdateAssignmentContext() *UpdateAssignmentContext {
	var p = new(UpdateAssignmentContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_updateAssignment
	return p
}

func InitEmptyUpdateAssignmentContext(p *UpdateAssignmentContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_updateAssignment
}

func (*UpdateAssignmentContext) IsUpdateAssignmentContext() {}

func NewUpdateAssignmentContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *UpdateAssignmentContext {
	var p = new(UpdateAssignmentContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_updateAssignment

	return p
}

func (s *UpdateAssignmentContext) GetParser() antlr.Parser { return s.parser }

func (s *UpdateAssignmentContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *UpdateAssignmentContext) EQUAL() antlr.TerminalNode {
	return s.GetToken(MiniQLParserEQUAL, 0)
}

func (s *UpdateAssignmentContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *UpdateAssignmentContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UpdateAssignmentContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *UpdateAssignmentContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitUpdateAssignment(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) UpdateAssignment() (localctx IUpdateAssignmentContext) {
	localctx = NewUpdateAssignmentContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 56, MiniQLParserRULE_updateAssignment)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(430)
		p.Identifier()
	}
	{
		p.SetState(431)
		p.Match(MiniQLParserEQUAL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(432)
		p.expression(0)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IGroupByItemContext is an interface to support dynamic dispatch.
type IGroupByItemContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Expression() IExpressionContext

	// IsGroupByItemContext differentiates from other interfaces.
	IsGroupByItemContext()
}

type GroupByItemContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyGroupByItemContext() *GroupByItemContext {
	var p = new(GroupByItemContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_groupByItem
	return p
}

func InitEmptyGroupByItemContext(p *GroupByItemContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_groupByItem
}

func (*GroupByItemContext) IsGroupByItemContext() {}

func NewGroupByItemContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *GroupByItemContext {
	var p = new(GroupByItemContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_groupByItem

	return p
}

func (s *GroupByItemContext) GetParser() antlr.Parser { return s.parser }

func (s *GroupByItemContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *GroupByItemContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *GroupByItemContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *GroupByItemContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitGroupByItem(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) GroupByItem() (localctx IGroupByItemContext) {
	localctx = NewGroupByItemContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 58, MiniQLParserRULE_groupByItem)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(434)
		p.expression(0)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IOrderByItemContext is an interface to support dynamic dispatch.
type IOrderByItemContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Expression() IExpressionContext
	ASC() antlr.TerminalNode
	DESC() antlr.TerminalNode

	// IsOrderByItemContext differentiates from other interfaces.
	IsOrderByItemContext()
}

type OrderByItemContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOrderByItemContext() *OrderByItemContext {
	var p = new(OrderByItemContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_orderByItem
	return p
}

func InitEmptyOrderByItemContext(p *OrderByItemContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_orderByItem
}

func (*OrderByItemContext) IsOrderByItemContext() {}

func NewOrderByItemContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OrderByItemContext {
	var p = new(OrderByItemContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_orderByItem

	return p
}

func (s *OrderByItemContext) GetParser() antlr.Parser { return s.parser }

func (s *OrderByItemContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *OrderByItemContext) ASC() antlr.TerminalNode {
	return s.GetToken(MiniQLParserASC, 0)
}

func (s *OrderByItemContext) DESC() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDESC, 0)
}

func (s *OrderByItemContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OrderByItemContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OrderByItemContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitOrderByItem(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) OrderByItem() (localctx IOrderByItemContext) {
	localctx = NewOrderByItemContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 60, MiniQLParserRULE_orderByItem)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(436)
		p.expression(0)
	}
	p.SetState(438)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MiniQLParserASC || _la == MiniQLParserDESC {
		{
			p.SetState(437)
			_la = p.GetTokenStream().LA(1)

			if !(_la == MiniQLParserASC || _la == MiniQLParserDESC) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFunctionCallContext is an interface to support dynamic dispatch.
type IFunctionCallContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext
	LEFT_PAREN() antlr.TerminalNode
	RIGHT_PAREN() antlr.TerminalNode
	ASTERISK() antlr.TerminalNode
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsFunctionCallContext differentiates from other interfaces.
	IsFunctionCallContext()
}

type FunctionCallContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFunctionCallContext() *FunctionCallContext {
	var p = new(FunctionCallContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_functionCall
	return p
}

func InitEmptyFunctionCallContext(p *FunctionCallContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_functionCall
}

func (*FunctionCallContext) IsFunctionCallContext() {}

func NewFunctionCallContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FunctionCallContext {
	var p = new(FunctionCallContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_functionCall

	return p
}

func (s *FunctionCallContext) GetParser() antlr.Parser { return s.parser }

func (s *FunctionCallContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *FunctionCallContext) LEFT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserLEFT_PAREN, 0)
}

func (s *FunctionCallContext) RIGHT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserRIGHT_PAREN, 0)
}

func (s *FunctionCallContext) ASTERISK() antlr.TerminalNode {
	return s.GetToken(MiniQLParserASTERISK, 0)
}

func (s *FunctionCallContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *FunctionCallContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *FunctionCallContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MiniQLParserCOMMA)
}

func (s *FunctionCallContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MiniQLParserCOMMA, i)
}

func (s *FunctionCallContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FunctionCallContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FunctionCallContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitFunctionCall(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) FunctionCall() (localctx IFunctionCallContext) {
	localctx = NewFunctionCallContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 62, MiniQLParserRULE_functionCall)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(440)
		p.Identifier()
	}
	{
		p.SetState(441)
		p.Match(MiniQLParserLEFT_PAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(451)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	switch p.GetTokenStream().LA(1) {
	case MiniQLParserASTERISK:
		{
			p.SetState(442)
			p.Match(MiniQLParserASTERISK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case MiniQLParserNULL, MiniQLParserTRUE, MiniQLParserFALSE, MiniQLParserLEFT_PAREN, MiniQLParserIDENTIFIER, MiniQLParserINTEGER_LITERAL, MiniQLParserFLOAT_LITERAL, MiniQLParserSTRING_LITERAL:
		{
			p.SetState(443)
			p.expression(0)
		}
		p.SetState(448)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == MiniQLParserCOMMA {
			{
				p.SetState(444)
				p.Match(MiniQLParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(445)
				p.expression(0)
			}

			p.SetState(450)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	case MiniQLParserRIGHT_PAREN:

	default:
	}
	{
		p.SetState(453)
		p.Match(MiniQLParserRIGHT_PAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPartitionMethodContext is an interface to support dynamic dispatch.
type IPartitionMethodContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	HASH() antlr.TerminalNode
	LEFT_PAREN() antlr.TerminalNode
	IdentifierList() IIdentifierListContext
	RIGHT_PAREN() antlr.TerminalNode
	RANGE() antlr.TerminalNode

	// IsPartitionMethodContext differentiates from other interfaces.
	IsPartitionMethodContext()
}

type PartitionMethodContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPartitionMethodContext() *PartitionMethodContext {
	var p = new(PartitionMethodContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_partitionMethod
	return p
}

func InitEmptyPartitionMethodContext(p *PartitionMethodContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_partitionMethod
}

func (*PartitionMethodContext) IsPartitionMethodContext() {}

func NewPartitionMethodContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PartitionMethodContext {
	var p = new(PartitionMethodContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_partitionMethod

	return p
}

func (s *PartitionMethodContext) GetParser() antlr.Parser { return s.parser }

func (s *PartitionMethodContext) HASH() antlr.TerminalNode {
	return s.GetToken(MiniQLParserHASH, 0)
}

func (s *PartitionMethodContext) LEFT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserLEFT_PAREN, 0)
}

func (s *PartitionMethodContext) IdentifierList() IIdentifierListContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierListContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierListContext)
}

func (s *PartitionMethodContext) RIGHT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserRIGHT_PAREN, 0)
}

func (s *PartitionMethodContext) RANGE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserRANGE, 0)
}

func (s *PartitionMethodContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PartitionMethodContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PartitionMethodContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitPartitionMethod(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) PartitionMethod() (localctx IPartitionMethodContext) {
	localctx = NewPartitionMethodContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 64, MiniQLParserRULE_partitionMethod)
	p.SetState(465)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MiniQLParserHASH:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(455)
			p.Match(MiniQLParserHASH)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(456)
			p.Match(MiniQLParserLEFT_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(457)
			p.IdentifierList()
		}
		{
			p.SetState(458)
			p.Match(MiniQLParserRIGHT_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case MiniQLParserRANGE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(460)
			p.Match(MiniQLParserRANGE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(461)
			p.Match(MiniQLParserLEFT_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(462)
			p.IdentifierList()
		}
		{
			p.SetState(463)
			p.Match(MiniQLParserRIGHT_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITransactionStatementContext is an interface to support dynamic dispatch.
type ITransactionStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	START() antlr.TerminalNode
	TRANSACTION() antlr.TerminalNode
	COMMIT() antlr.TerminalNode
	ROLLBACK() antlr.TerminalNode

	// IsTransactionStatementContext differentiates from other interfaces.
	IsTransactionStatementContext()
}

type TransactionStatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTransactionStatementContext() *TransactionStatementContext {
	var p = new(TransactionStatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_transactionStatement
	return p
}

func InitEmptyTransactionStatementContext(p *TransactionStatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_transactionStatement
}

func (*TransactionStatementContext) IsTransactionStatementContext() {}

func NewTransactionStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TransactionStatementContext {
	var p = new(TransactionStatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_transactionStatement

	return p
}

func (s *TransactionStatementContext) GetParser() antlr.Parser { return s.parser }

func (s *TransactionStatementContext) START() antlr.TerminalNode {
	return s.GetToken(MiniQLParserSTART, 0)
}

func (s *TransactionStatementContext) TRANSACTION() antlr.TerminalNode {
	return s.GetToken(MiniQLParserTRANSACTION, 0)
}

func (s *TransactionStatementContext) COMMIT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserCOMMIT, 0)
}

func (s *TransactionStatementContext) ROLLBACK() antlr.TerminalNode {
	return s.GetToken(MiniQLParserROLLBACK, 0)
}

func (s *TransactionStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TransactionStatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TransactionStatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitTransactionStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) TransactionStatement() (localctx ITransactionStatementContext) {
	localctx = NewTransactionStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 66, MiniQLParserRULE_transactionStatement)
	p.SetState(471)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MiniQLParserSTART:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(467)
			p.Match(MiniQLParserSTART)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(468)
			p.Match(MiniQLParserTRANSACTION)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case MiniQLParserCOMMIT:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(469)
			p.Match(MiniQLParserCOMMIT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case MiniQLParserROLLBACK:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(470)
			p.Match(MiniQLParserROLLBACK)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IUseStatementContext is an interface to support dynamic dispatch.
type IUseStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	USE() antlr.TerminalNode
	Identifier() IIdentifierContext

	// IsUseStatementContext differentiates from other interfaces.
	IsUseStatementContext()
}

type UseStatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUseStatementContext() *UseStatementContext {
	var p = new(UseStatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_useStatement
	return p
}

func InitEmptyUseStatementContext(p *UseStatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_useStatement
}

func (*UseStatementContext) IsUseStatementContext() {}

func NewUseStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *UseStatementContext {
	var p = new(UseStatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_useStatement

	return p
}

func (s *UseStatementContext) GetParser() antlr.Parser { return s.parser }

func (s *UseStatementContext) USE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserUSE, 0)
}

func (s *UseStatementContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *UseStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UseStatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *UseStatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitUseStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) UseStatement() (localctx IUseStatementContext) {
	localctx = NewUseStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 68, MiniQLParserRULE_useStatement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(473)
		p.Match(MiniQLParserUSE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(474)
		p.Identifier()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IShowDatabasesContext is an interface to support dynamic dispatch.
type IShowDatabasesContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SHOW() antlr.TerminalNode
	DATABASES() antlr.TerminalNode

	// IsShowDatabasesContext differentiates from other interfaces.
	IsShowDatabasesContext()
}

type ShowDatabasesContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyShowDatabasesContext() *ShowDatabasesContext {
	var p = new(ShowDatabasesContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_showDatabases
	return p
}

func InitEmptyShowDatabasesContext(p *ShowDatabasesContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_showDatabases
}

func (*ShowDatabasesContext) IsShowDatabasesContext() {}

func NewShowDatabasesContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ShowDatabasesContext {
	var p = new(ShowDatabasesContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_showDatabases

	return p
}

func (s *ShowDatabasesContext) GetParser() antlr.Parser { return s.parser }

func (s *ShowDatabasesContext) SHOW() antlr.TerminalNode {
	return s.GetToken(MiniQLParserSHOW, 0)
}

func (s *ShowDatabasesContext) DATABASES() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDATABASES, 0)
}

func (s *ShowDatabasesContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ShowDatabasesContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ShowDatabasesContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitShowDatabases(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) ShowDatabases() (localctx IShowDatabasesContext) {
	localctx = NewShowDatabasesContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 70, MiniQLParserRULE_showDatabases)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(476)
		p.Match(MiniQLParserSHOW)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(477)
		p.Match(MiniQLParserDATABASES)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IShowTablesContext is an interface to support dynamic dispatch.
type IShowTablesContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SHOW() antlr.TerminalNode
	TABLES() antlr.TerminalNode

	// IsShowTablesContext differentiates from other interfaces.
	IsShowTablesContext()
}

type ShowTablesContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyShowTablesContext() *ShowTablesContext {
	var p = new(ShowTablesContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_showTables
	return p
}

func InitEmptyShowTablesContext(p *ShowTablesContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_showTables
}

func (*ShowTablesContext) IsShowTablesContext() {}

func NewShowTablesContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ShowTablesContext {
	var p = new(ShowTablesContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_showTables

	return p
}

func (s *ShowTablesContext) GetParser() antlr.Parser { return s.parser }

func (s *ShowTablesContext) SHOW() antlr.TerminalNode {
	return s.GetToken(MiniQLParserSHOW, 0)
}

func (s *ShowTablesContext) TABLES() antlr.TerminalNode {
	return s.GetToken(MiniQLParserTABLES, 0)
}

func (s *ShowTablesContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ShowTablesContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ShowTablesContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitShowTables(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) ShowTables() (localctx IShowTablesContext) {
	localctx = NewShowTablesContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 72, MiniQLParserRULE_showTables)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(479)
		p.Match(MiniQLParserSHOW)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(480)
		p.Match(MiniQLParserTABLES)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IShowIndexesContext is an interface to support dynamic dispatch.
type IShowIndexesContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SHOW() antlr.TerminalNode
	INDEXES() antlr.TerminalNode
	TableName() ITableNameContext
	ON() antlr.TerminalNode
	FROM() antlr.TerminalNode

	// IsShowIndexesContext differentiates from other interfaces.
	IsShowIndexesContext()
}

type ShowIndexesContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyShowIndexesContext() *ShowIndexesContext {
	var p = new(ShowIndexesContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_showIndexes
	return p
}

func InitEmptyShowIndexesContext(p *ShowIndexesContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_showIndexes
}

func (*ShowIndexesContext) IsShowIndexesContext() {}

func NewShowIndexesContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ShowIndexesContext {
	var p = new(ShowIndexesContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_showIndexes

	return p
}

func (s *ShowIndexesContext) GetParser() antlr.Parser { return s.parser }

func (s *ShowIndexesContext) SHOW() antlr.TerminalNode {
	return s.GetToken(MiniQLParserSHOW, 0)
}

func (s *ShowIndexesContext) INDEXES() antlr.TerminalNode {
	return s.GetToken(MiniQLParserINDEXES, 0)
}

func (s *ShowIndexesContext) TableName() ITableNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
}

func (s *ShowIndexesContext) ON() antlr.TerminalNode {
	return s.GetToken(MiniQLParserON, 0)
}

func (s *ShowIndexesContext) FROM() antlr.TerminalNode {
	return s.GetToken(MiniQLParserFROM, 0)
}

func (s *ShowIndexesContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ShowIndexesContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ShowIndexesContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitShowIndexes(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) ShowIndexes() (localctx IShowIndexesContext) {
	localctx = NewShowIndexesContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 74, MiniQLParserRULE_showIndexes)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(482)
		p.Match(MiniQLParserSHOW)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(483)
		p.Match(MiniQLParserINDEXES)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(484)
		_la = p.GetTokenStream().LA(1)

		if !(_la == MiniQLParserFROM || _la == MiniQLParserON) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	{
		p.SetState(485)
		p.TableName()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExplainStatementContext is an interface to support dynamic dispatch.
type IExplainStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EXPLAIN() antlr.TerminalNode
	SelectStatement() ISelectStatementContext

	// IsExplainStatementContext differentiates from other interfaces.
	IsExplainStatementContext()
}

type ExplainStatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExplainStatementContext() *ExplainStatementContext {
	var p = new(ExplainStatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_explainStatement
	return p
}

func InitEmptyExplainStatementContext(p *ExplainStatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_explainStatement
}

func (*ExplainStatementContext) IsExplainStatementContext() {}

func NewExplainStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExplainStatementContext {
	var p = new(ExplainStatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_explainStatement

	return p
}

func (s *ExplainStatementContext) GetParser() antlr.Parser { return s.parser }

func (s *ExplainStatementContext) EXPLAIN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserEXPLAIN, 0)
}

func (s *ExplainStatementContext) SelectStatement() ISelectStatementContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelectStatementContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelectStatementContext)
}

func (s *ExplainStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExplainStatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExplainStatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitExplainStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) ExplainStatement() (localctx IExplainStatementContext) {
	localctx = NewExplainStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 76, MiniQLParserRULE_explainStatement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(487)
		p.Match(MiniQLParserEXPLAIN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(488)
		p.SelectStatement()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAnalyzeStatementContext is an interface to support dynamic dispatch.
type IAnalyzeStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ANALYZE() antlr.TerminalNode
	TABLE() antlr.TerminalNode
	TableName() ITableNameContext
	LEFT_PAREN() antlr.TerminalNode
	ColumnList() IColumnListContext
	RIGHT_PAREN() antlr.TerminalNode

	// IsAnalyzeStatementContext differentiates from other interfaces.
	IsAnalyzeStatementContext()
}

type AnalyzeStatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAnalyzeStatementContext() *AnalyzeStatementContext {
	var p = new(AnalyzeStatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_analyzeStatement
	return p
}

func InitEmptyAnalyzeStatementContext(p *AnalyzeStatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_analyzeStatement
}

func (*AnalyzeStatementContext) IsAnalyzeStatementContext() {}

func NewAnalyzeStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AnalyzeStatementContext {
	var p = new(AnalyzeStatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_analyzeStatement

	return p
}

func (s *AnalyzeStatementContext) GetParser() antlr.Parser { return s.parser }

func (s *AnalyzeStatementContext) ANALYZE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserANALYZE, 0)
}

func (s *AnalyzeStatementContext) TABLE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserTABLE, 0)
}

func (s *AnalyzeStatementContext) TableName() ITableNameContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
}

func (s *AnalyzeStatementContext) LEFT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserLEFT_PAREN, 0)
}

func (s *AnalyzeStatementContext) ColumnList() IColumnListContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumnListContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IColumnListContext)
}

func (s *AnalyzeStatementContext) RIGHT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserRIGHT_PAREN, 0)
}

func (s *AnalyzeStatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AnalyzeStatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AnalyzeStatementContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitAnalyzeStatement(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) AnalyzeStatement() (localctx IAnalyzeStatementContext) {
	localctx = NewAnalyzeStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 78, MiniQLParserRULE_analyzeStatement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(490)
		p.Match(MiniQLParserANALYZE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(491)
		p.Match(MiniQLParserTABLE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(492)
		p.TableName()
	}
	p.SetState(497)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == MiniQLParserLEFT_PAREN {
		{
			p.SetState(493)
			p.Match(MiniQLParserLEFT_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(494)
			p.ColumnList()
		}
		{
			p.SetState(495)
			p.Match(MiniQLParserRIGHT_PAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IColumnListContext is an interface to support dynamic dispatch.
type IColumnListContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllIdentifier() []IIdentifierContext
	Identifier(i int) IIdentifierContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsColumnListContext differentiates from other interfaces.
	IsColumnListContext()
}

type ColumnListContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyColumnListContext() *ColumnListContext {
	var p = new(ColumnListContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_columnList
	return p
}

func InitEmptyColumnListContext(p *ColumnListContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_columnList
}

func (*ColumnListContext) IsColumnListContext() {}

func NewColumnListContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ColumnListContext {
	var p = new(ColumnListContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_columnList

	return p
}

func (s *ColumnListContext) GetParser() antlr.Parser { return s.parser }

func (s *ColumnListContext) AllIdentifier() []IIdentifierContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIdentifierContext); ok {
			len++
		}
	}

	tst := make([]IIdentifierContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIdentifierContext); ok {
			tst[i] = t.(IIdentifierContext)
			i++
		}
	}

	return tst
}

func (s *ColumnListContext) Identifier(i int) IIdentifierContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *ColumnListContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MiniQLParserCOMMA)
}

func (s *ColumnListContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MiniQLParserCOMMA, i)
}

func (s *ColumnListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ColumnListContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ColumnListContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitColumnList(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) ColumnList() (localctx IColumnListContext) {
	localctx = NewColumnListContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 80, MiniQLParserRULE_columnList)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(499)
		p.Identifier()
	}
	p.SetState(504)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == MiniQLParserCOMMA {
		{
			p.SetState(500)
			p.Match(MiniQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(501)
			p.Identifier()
		}

		p.SetState(506)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIdentifierListContext is an interface to support dynamic dispatch.
type IIdentifierListContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllIdentifier() []IIdentifierContext
	Identifier(i int) IIdentifierContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsIdentifierListContext differentiates from other interfaces.
	IsIdentifierListContext()
}

type IdentifierListContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIdentifierListContext() *IdentifierListContext {
	var p = new(IdentifierListContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_identifierList
	return p
}

func InitEmptyIdentifierListContext(p *IdentifierListContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_identifierList
}

func (*IdentifierListContext) IsIdentifierListContext() {}

func NewIdentifierListContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IdentifierListContext {
	var p = new(IdentifierListContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_identifierList

	return p
}

func (s *IdentifierListContext) GetParser() antlr.Parser { return s.parser }

func (s *IdentifierListContext) AllIdentifier() []IIdentifierContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIdentifierContext); ok {
			len++
		}
	}

	tst := make([]IIdentifierContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIdentifierContext); ok {
			tst[i] = t.(IIdentifierContext)
			i++
		}
	}

	return tst
}

func (s *IdentifierListContext) Identifier(i int) IIdentifierContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *IdentifierListContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MiniQLParserCOMMA)
}

func (s *IdentifierListContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MiniQLParserCOMMA, i)
}

func (s *IdentifierListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IdentifierListContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IdentifierListContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitIdentifierList(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) IdentifierList() (localctx IIdentifierListContext) {
	localctx = NewIdentifierListContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 82, MiniQLParserRULE_identifierList)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(507)
		p.Identifier()
	}
	p.SetState(512)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == MiniQLParserCOMMA {
		{
			p.SetState(508)
			p.Match(MiniQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(509)
			p.Identifier()
		}

		p.SetState(514)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IValueListContext is an interface to support dynamic dispatch.
type IValueListContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllLiteral() []ILiteralContext
	Literal(i int) ILiteralContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsValueListContext differentiates from other interfaces.
	IsValueListContext()
}

type ValueListContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyValueListContext() *ValueListContext {
	var p = new(ValueListContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_valueList
	return p
}

func InitEmptyValueListContext(p *ValueListContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_valueList
}

func (*ValueListContext) IsValueListContext() {}

func NewValueListContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ValueListContext {
	var p = new(ValueListContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_valueList

	return p
}

func (s *ValueListContext) GetParser() antlr.Parser { return s.parser }

func (s *ValueListContext) AllLiteral() []ILiteralContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ILiteralContext); ok {
			len++
		}
	}

	tst := make([]ILiteralContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ILiteralContext); ok {
			tst[i] = t.(ILiteralContext)
			i++
		}
	}

	return tst
}

func (s *ValueListContext) Literal(i int) ILiteralContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteralContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILiteralContext)
}

func (s *ValueListContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(MiniQLParserCOMMA)
}

func (s *ValueListContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(MiniQLParserCOMMA, i)
}

func (s *ValueListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValueListContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ValueListContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitValueList(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) ValueList() (localctx IValueListContext) {
	localctx = NewValueListContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 84, MiniQLParserRULE_valueList)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(515)
		p.Literal()
	}
	p.SetState(520)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == MiniQLParserCOMMA {
		{
			p.SetState(516)
			p.Match(MiniQLParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(517)
			p.Literal()
		}

		p.SetState(522)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITableNameContext is an interface to support dynamic dispatch.
type ITableNameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllIdentifier() []IIdentifierContext
	Identifier(i int) IIdentifierContext
	DOT() antlr.TerminalNode

	// IsTableNameContext differentiates from other interfaces.
	IsTableNameContext()
}

type TableNameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTableNameContext() *TableNameContext {
	var p = new(TableNameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_tableName
	return p
}

func InitEmptyTableNameContext(p *TableNameContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_tableName
}

func (*TableNameContext) IsTableNameContext() {}

func NewTableNameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TableNameContext {
	var p = new(TableNameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_tableName

	return p
}

func (s *TableNameContext) GetParser() antlr.Parser { return s.parser }

func (s *TableNameContext) AllIdentifier() []IIdentifierContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IIdentifierContext); ok {
			len++
		}
	}

	tst := make([]IIdentifierContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IIdentifierContext); ok {
			tst[i] = t.(IIdentifierContext)
			i++
		}
	}

	return tst
}

func (s *TableNameContext) Identifier(i int) IIdentifierContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *TableNameContext) DOT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDOT, 0)
}

func (s *TableNameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TableNameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TableNameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitTableName(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) TableName() (localctx ITableNameContext) {
	localctx = NewTableNameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 86, MiniQLParserRULE_tableName)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(523)
		p.Identifier()
	}
	p.SetState(526)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 55, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(524)
			p.Match(MiniQLParserDOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(525)
			p.Identifier()
		}

	} else if p.HasError() { // JIM
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIdentifierContext is an interface to support dynamic dispatch.
type IIdentifierContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode

	// IsIdentifierContext differentiates from other interfaces.
	IsIdentifierContext()
}

type IdentifierContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIdentifierContext() *IdentifierContext {
	var p = new(IdentifierContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_identifier
	return p
}

func InitEmptyIdentifierContext(p *IdentifierContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_identifier
}

func (*IdentifierContext) IsIdentifierContext() {}

func NewIdentifierContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IdentifierContext {
	var p = new(IdentifierContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_identifier

	return p
}

func (s *IdentifierContext) GetParser() antlr.Parser { return s.parser }

func (s *IdentifierContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(MiniQLParserIDENTIFIER, 0)
}

func (s *IdentifierContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IdentifierContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IdentifierContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitIdentifier(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) Identifier() (localctx IIdentifierContext) {
	localctx = NewIdentifierContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 88, MiniQLParserRULE_identifier)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(528)
		p.Match(MiniQLParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IDataTypeContext is an interface to support dynamic dispatch.
type IDataTypeContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	INT_TYPE() antlr.TerminalNode
	INTEGER_TYPE() antlr.TerminalNode
	VARCHAR_TYPE() antlr.TerminalNode
	LEFT_PAREN() antlr.TerminalNode
	INTEGER_LITERAL() antlr.TerminalNode
	RIGHT_PAREN() antlr.TerminalNode
	BOOLEAN_TYPE() antlr.TerminalNode
	DOUBLE_TYPE() antlr.TerminalNode
	TIMESTAMP_TYPE() antlr.TerminalNode

	// IsDataTypeContext differentiates from other interfaces.
	IsDataTypeContext()
}

type DataTypeContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDataTypeContext() *DataTypeContext {
	var p = new(DataTypeContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_dataType
	return p
}

func InitEmptyDataTypeContext(p *DataTypeContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_dataType
}

func (*DataTypeContext) IsDataTypeContext() {}

func NewDataTypeContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DataTypeContext {
	var p = new(DataTypeContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_dataType

	return p
}

func (s *DataTypeContext) GetParser() antlr.Parser { return s.parser }

func (s *DataTypeContext) INT_TYPE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserINT_TYPE, 0)
}

func (s *DataTypeContext) INTEGER_TYPE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserINTEGER_TYPE, 0)
}

func (s *DataTypeContext) VARCHAR_TYPE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserVARCHAR_TYPE, 0)
}

func (s *DataTypeContext) LEFT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserLEFT_PAREN, 0)
}

func (s *DataTypeContext) INTEGER_LITERAL() antlr.TerminalNode {
	return s.GetToken(MiniQLParserINTEGER_LITERAL, 0)
}

func (s *DataTypeContext) RIGHT_PAREN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserRIGHT_PAREN, 0)
}

func (s *DataTypeContext) BOOLEAN_TYPE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserBOOLEAN_TYPE, 0)
}

func (s *DataTypeContext) DOUBLE_TYPE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDOUBLE_TYPE, 0)
}

func (s *DataTypeContext) TIMESTAMP_TYPE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserTIMESTAMP_TYPE, 0)
}

func (s *DataTypeContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DataTypeContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DataTypeContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitDataType(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) DataType() (localctx IDataTypeContext) {
	localctx = NewDataTypeContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 90, MiniQLParserRULE_dataType)
	var _la int

	p.SetState(541)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MiniQLParserINT_TYPE:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(530)
			p.Match(MiniQLParserINT_TYPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case MiniQLParserINTEGER_TYPE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(531)
			p.Match(MiniQLParserINTEGER_TYPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case MiniQLParserVARCHAR_TYPE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(532)
			p.Match(MiniQLParserVARCHAR_TYPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(536)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == MiniQLParserLEFT_PAREN {
			{
				p.SetState(533)
				p.Match(MiniQLParserLEFT_PAREN)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(534)
				p.Match(MiniQLParserINTEGER_LITERAL)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(535)
				p.Match(MiniQLParserRIGHT_PAREN)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}

	case MiniQLParserBOOLEAN_TYPE:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(538)
			p.Match(MiniQLParserBOOLEAN_TYPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case MiniQLParserDOUBLE_TYPE:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(539)
			p.Match(MiniQLParserDOUBLE_TYPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case MiniQLParserTIMESTAMP_TYPE:
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(540)
			p.Match(MiniQLParserTIMESTAMP_TYPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILiteralContext is an interface to support dynamic dispatch.
type ILiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	INTEGER_LITERAL() antlr.TerminalNode
	FLOAT_LITERAL() antlr.TerminalNode
	STRING_LITERAL() antlr.TerminalNode
	TRUE() antlr.TerminalNode
	FALSE() antlr.TerminalNode
	NULL() antlr.TerminalNode

	// IsLiteralContext differentiates from other interfaces.
	IsLiteralContext()
}

type LiteralContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLiteralContext() *LiteralContext {
	var p = new(LiteralContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_literal
	return p
}

func InitEmptyLiteralContext(p *LiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_literal
}

func (*LiteralContext) IsLiteralContext() {}

func NewLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LiteralContext {
	var p = new(LiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_literal

	return p
}

func (s *LiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *LiteralContext) INTEGER_LITERAL() antlr.TerminalNode {
	return s.GetToken(MiniQLParserINTEGER_LITERAL, 0)
}

func (s *LiteralContext) FLOAT_LITERAL() antlr.TerminalNode {
	return s.GetToken(MiniQLParserFLOAT_LITERAL, 0)
}

func (s *LiteralContext) STRING_LITERAL() antlr.TerminalNode {
	return s.GetToken(MiniQLParserSTRING_LITERAL, 0)
}

func (s *LiteralContext) TRUE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserTRUE, 0)
}

func (s *LiteralContext) FALSE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserFALSE, 0)
}

func (s *LiteralContext) NULL() antlr.TerminalNode {
	return s.GetToken(MiniQLParserNULL, 0)
}

func (s *LiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LiteralContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *MiniQLParser) Literal() (localctx ILiteralContext) {
	localctx = NewLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 92, MiniQLParserRULE_literal)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(543)
		_la = p.GetTokenStream().LA(1)

		if !((int64((_la-24)) & ^0x3f) == 0 && ((int64(1)<<(_la-24))&2017612633061982215) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

func (p *MiniQLParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 21:
		var t *TableReferenceContext = nil
		if localctx != nil {
			t = localctx.(*TableReferenceContext)
		}
		return p.TableReference_Sempred(t, predIndex)

	case 24:
		var t *ExpressionContext = nil
		if localctx != nil {
			t = localctx.(*ExpressionContext)
		}
		return p.Expression_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *MiniQLParser) TableReference_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 1)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}

func (p *MiniQLParser) Expression_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 1:
		return p.Precpred(p.GetParserRuleContext(), 7)

	case 2:
		return p.Precpred(p.GetParserRuleContext(), 6)

	case 3:
		return p.Precpred(p.GetParserRuleContext(), 5)

	case 4:
		return p.Precpred(p.GetParserRuleContext(), 4)

	case 5:
		return p.Precpred(p.GetParserRuleContext(), 3)

	case 6:
		return p.Precpred(p.GetParserRuleContext(), 2)

	case 7:
		return p.Precpred(p.GetParserRuleContext(), 1)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}

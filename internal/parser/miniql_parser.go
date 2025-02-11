// Code generated from /Users/yasonlee/codes/minidb/internal/parser/MiniQL.g4 by ANTLR 4.13.2. DO NOT EDIT.

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
    "", "';'", "'('", "','", "')'", "'.'", "'*'", "'='", "'/'", "'+'", "'-'",
  }
  staticData.SymbolicNames = []string{
    "", "", "", "", "", "", "", "", "", "", "", "CREATE", "DATABASE", "TABLE", 
    "INDEX", "DROP", "INSERT", "INTO", "VALUES", "SELECT", "FROM", "WHERE", 
    "GROUP", "BY", "HAVING", "ORDER", "LIMIT", "JOIN", "INNER", "LEFT", 
    "ON", "AS", "USE", "SHOW", "EXPLAIN", "PRIMARY", "KEY", "PARTITION", 
    "INT", "BIGINT", "VARCHAR", "DATE", "DOUBLE", "AND", "OR", "ASC", "DESC", 
    "NULL", "HASH", "RANGE", "ANALYZE", "VERBOSE", "UPDATE", "SET", "DELETE", 
    "DATABASES", "TABLES", "NOT", "UNIQUE", "DEFAULT", "IDENTIFIER", "STRING", 
    "INTEGER", "FLOAT", "COMPARISON_OP", "WS",
  }
  staticData.RuleNames = []string{
    "parse", "error", "sqlStatement", "ddlStatement", "dmlStatement", "dqlStatement", 
    "utilityStatement", "createDatabase", "createTable", "columnDef", "columnConstraint", 
    "tableConstraint", "createIndex", "dropTable", "dropDatabase", "insertStatement", 
    "updateStatement", "deleteStatement", "selectStatement", "selectItem", 
    "tableReference", "tableName", "identifierList", "valueList", "updateAssignment", 
    "groupByItem", "orderByItem", "functionName", "literal", "identifier", 
    "partitionMethod", "joinType", "useStatement", "showDatabases", "showTables", 
    "explainStatement", "dataType", "expression",
  }
  staticData.PredictionContextCache = antlr.NewPredictionContextCache()
  staticData.serializedATN = []int32{
	4, 1, 65, 462, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7, 
	4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7, 
	10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15, 
	2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7, 20, 2, 
	21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2, 25, 7, 25, 2, 26, 
	7, 26, 2, 27, 7, 27, 2, 28, 7, 28, 2, 29, 7, 29, 2, 30, 7, 30, 2, 31, 7, 
	31, 2, 32, 7, 32, 2, 33, 7, 33, 2, 34, 7, 34, 2, 35, 7, 35, 2, 36, 7, 36, 
	2, 37, 7, 37, 1, 0, 1, 0, 5, 0, 79, 8, 0, 10, 0, 12, 0, 82, 9, 0, 1, 0, 
	1, 0, 1, 1, 4, 1, 87, 8, 1, 11, 1, 12, 1, 88, 1, 1, 1, 1, 1, 2, 1, 2, 1, 
	2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 3, 2, 105, 8, 
	2, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 3, 3, 112, 8, 3, 1, 4, 1, 4, 1, 4, 3, 
	4, 117, 8, 4, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 1, 6, 3, 6, 125, 8, 6, 1, 7, 
	1, 7, 1, 7, 1, 7, 1, 8, 1, 8, 1, 8, 1, 8, 1, 8, 1, 8, 1, 8, 5, 8, 138, 
	8, 8, 10, 8, 12, 8, 141, 9, 8, 1, 8, 1, 8, 5, 8, 145, 8, 8, 10, 8, 12, 
	8, 148, 9, 8, 1, 8, 1, 8, 1, 8, 1, 8, 3, 8, 154, 8, 8, 1, 9, 1, 9, 1, 9, 
	5, 9, 159, 8, 9, 10, 9, 12, 9, 162, 9, 9, 1, 10, 3, 10, 165, 8, 10, 1, 
	10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 3, 10, 173, 8, 10, 1, 11, 1, 11, 
	1, 11, 1, 11, 1, 11, 1, 11, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 
	12, 1, 12, 1, 12, 1, 13, 1, 13, 1, 13, 1, 13, 1, 14, 1, 14, 1, 14, 1, 14, 
	1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 3, 15, 205, 8, 15, 1, 
	15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 5, 15, 216, 
	8, 15, 10, 15, 12, 15, 219, 9, 15, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 
	16, 5, 16, 227, 8, 16, 10, 16, 12, 16, 230, 9, 16, 1, 16, 1, 16, 3, 16, 
	234, 8, 16, 1, 17, 1, 17, 1, 17, 1, 17, 1, 17, 3, 17, 241, 8, 17, 1, 18, 
	1, 18, 1, 18, 1, 18, 5, 18, 247, 8, 18, 10, 18, 12, 18, 250, 9, 18, 1, 
	18, 1, 18, 1, 18, 1, 18, 3, 18, 256, 8, 18, 1, 18, 1, 18, 1, 18, 1, 18, 
	1, 18, 5, 18, 263, 8, 18, 10, 18, 12, 18, 266, 9, 18, 3, 18, 268, 8, 18, 
	1, 18, 1, 18, 3, 18, 272, 8, 18, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18, 5, 
	18, 279, 8, 18, 10, 18, 12, 18, 282, 9, 18, 3, 18, 284, 8, 18, 1, 18, 1, 
	18, 3, 18, 288, 8, 18, 1, 19, 1, 19, 1, 19, 3, 19, 293, 8, 19, 1, 19, 1, 
	19, 1, 19, 3, 19, 298, 8, 19, 1, 19, 3, 19, 301, 8, 19, 3, 19, 303, 8, 
	19, 1, 20, 1, 20, 1, 20, 3, 20, 308, 8, 20, 1, 20, 3, 20, 311, 8, 20, 1, 
	20, 1, 20, 1, 20, 1, 20, 3, 20, 317, 8, 20, 1, 20, 1, 20, 3, 20, 321, 8, 
	20, 1, 20, 1, 20, 1, 20, 1, 20, 1, 20, 1, 20, 1, 20, 5, 20, 330, 8, 20, 
	10, 20, 12, 20, 333, 9, 20, 1, 21, 1, 21, 1, 22, 1, 22, 1, 22, 5, 22, 340, 
	8, 22, 10, 22, 12, 22, 343, 9, 22, 1, 23, 1, 23, 1, 23, 5, 23, 348, 8, 
	23, 10, 23, 12, 23, 351, 9, 23, 1, 24, 1, 24, 1, 24, 1, 24, 1, 25, 1, 25, 
	1, 26, 1, 26, 3, 26, 361, 8, 26, 1, 27, 1, 27, 1, 28, 1, 28, 1, 29, 1, 
	29, 1, 30, 1, 30, 1, 30, 1, 30, 1, 30, 1, 30, 1, 30, 1, 30, 1, 30, 1, 30, 
	3, 30, 379, 8, 30, 1, 31, 3, 31, 382, 8, 31, 1, 31, 3, 31, 385, 8, 31, 
	1, 32, 1, 32, 1, 32, 1, 33, 1, 33, 1, 33, 1, 34, 1, 34, 1, 34, 1, 35, 1, 
	35, 3, 35, 398, 8, 35, 1, 35, 3, 35, 401, 8, 35, 1, 35, 1, 35, 1, 36, 1, 
	36, 1, 36, 1, 36, 1, 36, 1, 36, 1, 36, 1, 36, 3, 36, 413, 8, 36, 1, 37, 
	1, 37, 1, 37, 1, 37, 1, 37, 1, 37, 1, 37, 1, 37, 1, 37, 1, 37, 1, 37, 1, 
	37, 1, 37, 1, 37, 1, 37, 1, 37, 5, 37, 431, 8, 37, 10, 37, 12, 37, 434, 
	9, 37, 3, 37, 436, 8, 37, 1, 37, 1, 37, 3, 37, 440, 8, 37, 1, 37, 1, 37, 
	1, 37, 1, 37, 1, 37, 1, 37, 1, 37, 1, 37, 1, 37, 1, 37, 1, 37, 1, 37, 1, 
	37, 1, 37, 1, 37, 5, 37, 457, 8, 37, 10, 37, 12, 37, 460, 9, 37, 1, 37, 
	1, 88, 2, 40, 74, 38, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 
	28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56, 58, 60, 62, 
	64, 66, 68, 70, 72, 74, 0, 4, 1, 0, 45, 46, 2, 0, 47, 47, 61, 63, 2, 0, 
	6, 6, 8, 8, 1, 0, 9, 10, 491, 0, 80, 1, 0, 0, 0, 2, 86, 1, 0, 0, 0, 4, 
	104, 1, 0, 0, 0, 6, 111, 1, 0, 0, 0, 8, 116, 1, 0, 0, 0, 10, 118, 1, 0, 
	0, 0, 12, 124, 1, 0, 0, 0, 14, 126, 1, 0, 0, 0, 16, 130, 1, 0, 0, 0, 18, 
	155, 1, 0, 0, 0, 20, 172, 1, 0, 0, 0, 22, 174, 1, 0, 0, 0, 24, 180, 1, 
	0, 0, 0, 26, 189, 1, 0, 0, 0, 28, 193, 1, 0, 0, 0, 30, 197, 1, 0, 0, 0, 
	32, 220, 1, 0, 0, 0, 34, 235, 1, 0, 0, 0, 36, 242, 1, 0, 0, 0, 38, 302, 
	1, 0, 0, 0, 40, 320, 1, 0, 0, 0, 42, 334, 1, 0, 0, 0, 44, 336, 1, 0, 0, 
	0, 46, 344, 1, 0, 0, 0, 48, 352, 1, 0, 0, 0, 50, 356, 1, 0, 0, 0, 52, 358, 
	1, 0, 0, 0, 54, 362, 1, 0, 0, 0, 56, 364, 1, 0, 0, 0, 58, 366, 1, 0, 0, 
	0, 60, 378, 1, 0, 0, 0, 62, 384, 1, 0, 0, 0, 64, 386, 1, 0, 0, 0, 66, 389, 
	1, 0, 0, 0, 68, 392, 1, 0, 0, 0, 70, 395, 1, 0, 0, 0, 72, 412, 1, 0, 0, 
	0, 74, 439, 1, 0, 0, 0, 76, 79, 3, 4, 2, 0, 77, 79, 3, 2, 1, 0, 78, 76, 
	1, 0, 0, 0, 78, 77, 1, 0, 0, 0, 79, 82, 1, 0, 0, 0, 80, 78, 1, 0, 0, 0, 
	80, 81, 1, 0, 0, 0, 81, 83, 1, 0, 0, 0, 82, 80, 1, 0, 0, 0, 83, 84, 5, 
	0, 0, 1, 84, 1, 1, 0, 0, 0, 85, 87, 9, 0, 0, 0, 86, 85, 1, 0, 0, 0, 87, 
	88, 1, 0, 0, 0, 88, 89, 1, 0, 0, 0, 88, 86, 1, 0, 0, 0, 89, 90, 1, 0, 0, 
	0, 90, 91, 5, 1, 0, 0, 91, 3, 1, 0, 0, 0, 92, 93, 3, 6, 3, 0, 93, 94, 5, 
	1, 0, 0, 94, 105, 1, 0, 0, 0, 95, 96, 3, 8, 4, 0, 96, 97, 5, 1, 0, 0, 97, 
	105, 1, 0, 0, 0, 98, 99, 3, 10, 5, 0, 99, 100, 5, 1, 0, 0, 100, 105, 1, 
	0, 0, 0, 101, 102, 3, 12, 6, 0, 102, 103, 5, 1, 0, 0, 103, 105, 1, 0, 0, 
	0, 104, 92, 1, 0, 0, 0, 104, 95, 1, 0, 0, 0, 104, 98, 1, 0, 0, 0, 104, 
	101, 1, 0, 0, 0, 105, 5, 1, 0, 0, 0, 106, 112, 3, 14, 7, 0, 107, 112, 3, 
	16, 8, 0, 108, 112, 3, 24, 12, 0, 109, 112, 3, 26, 13, 0, 110, 112, 3, 
	28, 14, 0, 111, 106, 1, 0, 0, 0, 111, 107, 1, 0, 0, 0, 111, 108, 1, 0, 
	0, 0, 111, 109, 1, 0, 0, 0, 111, 110, 1, 0, 0, 0, 112, 7, 1, 0, 0, 0, 113, 
	117, 3, 30, 15, 0, 114, 117, 3, 32, 16, 0, 115, 117, 3, 34, 17, 0, 116, 
	113, 1, 0, 0, 0, 116, 114, 1, 0, 0, 0, 116, 115, 1, 0, 0, 0, 117, 9, 1, 
	0, 0, 0, 118, 119, 3, 36, 18, 0, 119, 11, 1, 0, 0, 0, 120, 125, 3, 64, 
	32, 0, 121, 125, 3, 66, 33, 0, 122, 125, 3, 68, 34, 0, 123, 125, 3, 70, 
	35, 0, 124, 120, 1, 0, 0, 0, 124, 121, 1, 0, 0, 0, 124, 122, 1, 0, 0, 0, 
	124, 123, 1, 0, 0, 0, 125, 13, 1, 0, 0, 0, 126, 127, 5, 11, 0, 0, 127, 
	128, 5, 12, 0, 0, 128, 129, 3, 58, 29, 0, 129, 15, 1, 0, 0, 0, 130, 131, 
	5, 11, 0, 0, 131, 132, 5, 13, 0, 0, 132, 133, 3, 42, 21, 0, 133, 134, 5, 
	2, 0, 0, 134, 139, 3, 18, 9, 0, 135, 136, 5, 3, 0, 0, 136, 138, 3, 18, 
	9, 0, 137, 135, 1, 0, 0, 0, 138, 141, 1, 0, 0, 0, 139, 137, 1, 0, 0, 0, 
	139, 140, 1, 0, 0, 0, 140, 146, 1, 0, 0, 0, 141, 139, 1, 0, 0, 0, 142, 
	143, 5, 3, 0, 0, 143, 145, 3, 22, 11, 0, 144, 142, 1, 0, 0, 0, 145, 148, 
	1, 0, 0, 0, 146, 144, 1, 0, 0, 0, 146, 147, 1, 0, 0, 0, 147, 149, 1, 0, 
	0, 0, 148, 146, 1, 0, 0, 0, 149, 153, 5, 4, 0, 0, 150, 151, 5, 37, 0, 0, 
	151, 152, 5, 23, 0, 0, 152, 154, 3, 60, 30, 0, 153, 150, 1, 0, 0, 0, 153, 
	154, 1, 0, 0, 0, 154, 17, 1, 0, 0, 0, 155, 156, 3, 58, 29, 0, 156, 160, 
	3, 72, 36, 0, 157, 159, 3, 20, 10, 0, 158, 157, 1, 0, 0, 0, 159, 162, 1, 
	0, 0, 0, 160, 158, 1, 0, 0, 0, 160, 161, 1, 0, 0, 0, 161, 19, 1, 0, 0, 
	0, 162, 160, 1, 0, 0, 0, 163, 165, 5, 57, 0, 0, 164, 163, 1, 0, 0, 0, 164, 
	165, 1, 0, 0, 0, 165, 166, 1, 0, 0, 0, 166, 173, 5, 47, 0, 0, 167, 168, 
	5, 35, 0, 0, 168, 173, 5, 36, 0, 0, 169, 173, 5, 58, 0, 0, 170, 171, 5, 
	59, 0, 0, 171, 173, 3, 56, 28, 0, 172, 164, 1, 0, 0, 0, 172, 167, 1, 0, 
	0, 0, 172, 169, 1, 0, 0, 0, 172, 170, 1, 0, 0, 0, 173, 21, 1, 0, 0, 0, 
	174, 175, 5, 35, 0, 0, 175, 176, 5, 36, 0, 0, 176, 177, 5, 2, 0, 0, 177, 
	178, 3, 44, 22, 0, 178, 179, 5, 4, 0, 0, 179, 23, 1, 0, 0, 0, 180, 181, 
	5, 11, 0, 0, 181, 182, 5, 14, 0, 0, 182, 183, 3, 58, 29, 0, 183, 184, 5, 
	30, 0, 0, 184, 185, 3, 42, 21, 0, 185, 186, 5, 2, 0, 0, 186, 187, 3, 44, 
	22, 0, 187, 188, 5, 4, 0, 0, 188, 25, 1, 0, 0, 0, 189, 190, 5, 15, 0, 0, 
	190, 191, 5, 13, 0, 0, 191, 192, 3, 42, 21, 0, 192, 27, 1, 0, 0, 0, 193, 
	194, 5, 15, 0, 0, 194, 195, 5, 12, 0, 0, 195, 196, 3, 58, 29, 0, 196, 29, 
	1, 0, 0, 0, 197, 198, 5, 16, 0, 0, 198, 199, 5, 17, 0, 0, 199, 204, 3, 
	42, 21, 0, 200, 201, 5, 2, 0, 0, 201, 202, 3, 44, 22, 0, 202, 203, 5, 4, 
	0, 0, 203, 205, 1, 0, 0, 0, 204, 200, 1, 0, 0, 0, 204, 205, 1, 0, 0, 0, 
	205, 206, 1, 0, 0, 0, 206, 207, 5, 18, 0, 0, 207, 208, 5, 2, 0, 0, 208, 
	209, 3, 46, 23, 0, 209, 217, 5, 4, 0, 0, 210, 211, 5, 3, 0, 0, 211, 212, 
	5, 2, 0, 0, 212, 213, 3, 46, 23, 0, 213, 214, 5, 4, 0, 0, 214, 216, 1, 
	0, 0, 0, 215, 210, 1, 0, 0, 0, 216, 219, 1, 0, 0, 0, 217, 215, 1, 0, 0, 
	0, 217, 218, 1, 0, 0, 0, 218, 31, 1, 0, 0, 0, 219, 217, 1, 0, 0, 0, 220, 
	221, 5, 52, 0, 0, 221, 222, 3, 42, 21, 0, 222, 223, 5, 53, 0, 0, 223, 228, 
	3, 48, 24, 0, 224, 225, 5, 3, 0, 0, 225, 227, 3, 48, 24, 0, 226, 224, 1, 
	0, 0, 0, 227, 230, 1, 0, 0, 0, 228, 226, 1, 0, 0, 0, 228, 229, 1, 0, 0, 
	0, 229, 233, 1, 0, 0, 0, 230, 228, 1, 0, 0, 0, 231, 232, 5, 21, 0, 0, 232, 
	234, 3, 74, 37, 0, 233, 231, 1, 0, 0, 0, 233, 234, 1, 0, 0, 0, 234, 33, 
	1, 0, 0, 0, 235, 236, 5, 54, 0, 0, 236, 237, 5, 20, 0, 0, 237, 240, 3, 
	42, 21, 0, 238, 239, 5, 21, 0, 0, 239, 241, 3, 74, 37, 0, 240, 238, 1, 
	0, 0, 0, 240, 241, 1, 0, 0, 0, 241, 35, 1, 0, 0, 0, 242, 243, 5, 19, 0, 
	0, 243, 248, 3, 38, 19, 0, 244, 245, 5, 3, 0, 0, 245, 247, 3, 38, 19, 0, 
	246, 244, 1, 0, 0, 0, 247, 250, 1, 0, 0, 0, 248, 246, 1, 0, 0, 0, 248, 
	249, 1, 0, 0, 0, 249, 251, 1, 0, 0, 0, 250, 248, 1, 0, 0, 0, 251, 252, 
	5, 20, 0, 0, 252, 255, 3, 40, 20, 0, 253, 254, 5, 21, 0, 0, 254, 256, 3, 
	74, 37, 0, 255, 253, 1, 0, 0, 0, 255, 256, 1, 0, 0, 0, 256, 267, 1, 0, 
	0, 0, 257, 258, 5, 22, 0, 0, 258, 259, 5, 23, 0, 0, 259, 264, 3, 50, 25, 
	0, 260, 261, 5, 3, 0, 0, 261, 263, 3, 50, 25, 0, 262, 260, 1, 0, 0, 0, 
	263, 266, 1, 0, 0, 0, 264, 262, 1, 0, 0, 0, 264, 265, 1, 0, 0, 0, 265, 
	268, 1, 0, 0, 0, 266, 264, 1, 0, 0, 0, 267, 257, 1, 0, 0, 0, 267, 268, 
	1, 0, 0, 0, 268, 271, 1, 0, 0, 0, 269, 270, 5, 24, 0, 0, 270, 272, 3, 74, 
	37, 0, 271, 269, 1, 0, 0, 0, 271, 272, 1, 0, 0, 0, 272, 283, 1, 0, 0, 0, 
	273, 274, 5, 25, 0, 0, 274, 275, 5, 23, 0, 0, 275, 280, 3, 52, 26, 0, 276, 
	277, 5, 3, 0, 0, 277, 279, 3, 52, 26, 0, 278, 276, 1, 0, 0, 0, 279, 282, 
	1, 0, 0, 0, 280, 278, 1, 0, 0, 0, 280, 281, 1, 0, 0, 0, 281, 284, 1, 0, 
	0, 0, 282, 280, 1, 0, 0, 0, 283, 273, 1, 0, 0, 0, 283, 284, 1, 0, 0, 0, 
	284, 287, 1, 0, 0, 0, 285, 286, 5, 26, 0, 0, 286, 288, 5, 62, 0, 0, 287, 
	285, 1, 0, 0, 0, 287, 288, 1, 0, 0, 0, 288, 37, 1, 0, 0, 0, 289, 290, 3, 
	42, 21, 0, 290, 291, 5, 5, 0, 0, 291, 293, 1, 0, 0, 0, 292, 289, 1, 0, 
	0, 0, 292, 293, 1, 0, 0, 0, 293, 294, 1, 0, 0, 0, 294, 303, 5, 6, 0, 0, 
	295, 300, 3, 74, 37, 0, 296, 298, 5, 31, 0, 0, 297, 296, 1, 0, 0, 0, 297, 
	298, 1, 0, 0, 0, 298, 299, 1, 0, 0, 0, 299, 301, 3, 58, 29, 0, 300, 297, 
	1, 0, 0, 0, 300, 301, 1, 0, 0, 0, 301, 303, 1, 0, 0, 0, 302, 292, 1, 0, 
	0, 0, 302, 295, 1, 0, 0, 0, 303, 39, 1, 0, 0, 0, 304, 305, 6, 20, -1, 0, 
	305, 310, 3, 42, 21, 0, 306, 308, 5, 31, 0, 0, 307, 306, 1, 0, 0, 0, 307, 
	308, 1, 0, 0, 0, 308, 309, 1, 0, 0, 0, 309, 311, 3, 58, 29, 0, 310, 307, 
	1, 0, 0, 0, 310, 311, 1, 0, 0, 0, 311, 321, 1, 0, 0, 0, 312, 313, 5, 2, 
	0, 0, 313, 314, 3, 36, 18, 0, 314, 316, 5, 4, 0, 0, 315, 317, 5, 31, 0, 
	0, 316, 315, 1, 0, 0, 0, 316, 317, 1, 0, 0, 0, 317, 318, 1, 0, 0, 0, 318, 
	319, 3, 58, 29, 0, 319, 321, 1, 0, 0, 0, 320, 304, 1, 0, 0, 0, 320, 312, 
	1, 0, 0, 0, 321, 331, 1, 0, 0, 0, 322, 323, 10, 1, 0, 0, 323, 324, 3, 62, 
	31, 0, 324, 325, 5, 27, 0, 0, 325, 326, 3, 40, 20, 0, 326, 327, 5, 30, 
	0, 0, 327, 328, 3, 74, 37, 0, 328, 330, 1, 0, 0, 0, 329, 322, 1, 0, 0, 
	0, 330, 333, 1, 0, 0, 0, 331, 329, 1, 0, 0, 0, 331, 332, 1, 0, 0, 0, 332, 
	41, 1, 0, 0, 0, 333, 331, 1, 0, 0, 0, 334, 335, 3, 58, 29, 0, 335, 43, 
	1, 0, 0, 0, 336, 341, 3, 58, 29, 0, 337, 338, 5, 3, 0, 0, 338, 340, 3, 
	58, 29, 0, 339, 337, 1, 0, 0, 0, 340, 343, 1, 0, 0, 0, 341, 339, 1, 0, 
	0, 0, 341, 342, 1, 0, 0, 0, 342, 45, 1, 0, 0, 0, 343, 341, 1, 0, 0, 0, 
	344, 349, 3, 74, 37, 0, 345, 346, 5, 3, 0, 0, 346, 348, 3, 74, 37, 0, 347, 
	345, 1, 0, 0, 0, 348, 351, 1, 0, 0, 0, 349, 347, 1, 0, 0, 0, 349, 350, 
	1, 0, 0, 0, 350, 47, 1, 0, 0, 0, 351, 349, 1, 0, 0, 0, 352, 353, 3, 58, 
	29, 0, 353, 354, 5, 7, 0, 0, 354, 355, 3, 74, 37, 0, 355, 49, 1, 0, 0, 
	0, 356, 357, 3, 74, 37, 0, 357, 51, 1, 0, 0, 0, 358, 360, 3, 74, 37, 0, 
	359, 361, 7, 0, 0, 0, 360, 359, 1, 0, 0, 0, 360, 361, 1, 0, 0, 0, 361, 
	53, 1, 0, 0, 0, 362, 363, 3, 58, 29, 0, 363, 55, 1, 0, 0, 0, 364, 365, 
	7, 1, 0, 0, 365, 57, 1, 0, 0, 0, 366, 367, 5, 60, 0, 0, 367, 59, 1, 0, 
	0, 0, 368, 369, 5, 48, 0, 0, 369, 370, 5, 2, 0, 0, 370, 371, 3, 44, 22, 
	0, 371, 372, 5, 4, 0, 0, 372, 379, 1, 0, 0, 0, 373, 374, 5, 49, 0, 0, 374, 
	375, 5, 2, 0, 0, 375, 376, 3, 74, 37, 0, 376, 377, 5, 4, 0, 0, 377, 379, 
	1, 0, 0, 0, 378, 368, 1, 0, 0, 0, 378, 373, 1, 0, 0, 0, 379, 61, 1, 0, 
	0, 0, 380, 382, 5, 28, 0, 0, 381, 380, 1, 0, 0, 0, 381, 382, 1, 0, 0, 0, 
	382, 385, 1, 0, 0, 0, 383, 385, 5, 29, 0, 0, 384, 381, 1, 0, 0, 0, 384, 
	383, 1, 0, 0, 0, 385, 63, 1, 0, 0, 0, 386, 387, 5, 32, 0, 0, 387, 388, 
	3, 58, 29, 0, 388, 65, 1, 0, 0, 0, 389, 390, 5, 33, 0, 0, 390, 391, 5, 
	55, 0, 0, 391, 67, 1, 0, 0, 0, 392, 393, 5, 33, 0, 0, 393, 394, 5, 56, 
	0, 0, 394, 69, 1, 0, 0, 0, 395, 397, 5, 34, 0, 0, 396, 398, 5, 50, 0, 0, 
	397, 396, 1, 0, 0, 0, 397, 398, 1, 0, 0, 0, 398, 400, 1, 0, 0, 0, 399, 
	401, 5, 51, 0, 0, 400, 399, 1, 0, 0, 0, 400, 401, 1, 0, 0, 0, 401, 402, 
	1, 0, 0, 0, 402, 403, 3, 4, 2, 0, 403, 71, 1, 0, 0, 0, 404, 413, 5, 38, 
	0, 0, 405, 413, 5, 39, 0, 0, 406, 407, 5, 40, 0, 0, 407, 408, 5, 2, 0, 
	0, 408, 409, 5, 62, 0, 0, 409, 413, 5, 4, 0, 0, 410, 413, 5, 41, 0, 0, 
	411, 413, 5, 42, 0, 0, 412, 404, 1, 0, 0, 0, 412, 405, 1, 0, 0, 0, 412, 
	406, 1, 0, 0, 0, 412, 410, 1, 0, 0, 0, 412, 411, 1, 0, 0, 0, 413, 73, 1, 
	0, 0, 0, 414, 415, 6, 37, -1, 0, 415, 440, 3, 56, 28, 0, 416, 440, 3, 58, 
	29, 0, 417, 418, 3, 42, 21, 0, 418, 419, 5, 5, 0, 0, 419, 420, 3, 58, 29, 
	0, 420, 440, 1, 0, 0, 0, 421, 422, 5, 2, 0, 0, 422, 423, 3, 74, 37, 0, 
	423, 424, 5, 4, 0, 0, 424, 440, 1, 0, 0, 0, 425, 426, 3, 54, 27, 0, 426, 
	435, 5, 2, 0, 0, 427, 432, 3, 74, 37, 0, 428, 429, 5, 3, 0, 0, 429, 431, 
	3, 74, 37, 0, 430, 428, 1, 0, 0, 0, 431, 434, 1, 0, 0, 0, 432, 430, 1, 
	0, 0, 0, 432, 433, 1, 0, 0, 0, 433, 436, 1, 0, 0, 0, 434, 432, 1, 0, 0, 
	0, 435, 427, 1, 0, 0, 0, 435, 436, 1, 0, 0, 0, 436, 437, 1, 0, 0, 0, 437, 
	438, 5, 4, 0, 0, 438, 440, 1, 0, 0, 0, 439, 414, 1, 0, 0, 0, 439, 416, 
	1, 0, 0, 0, 439, 417, 1, 0, 0, 0, 439, 421, 1, 0, 0, 0, 439, 425, 1, 0, 
	0, 0, 440, 458, 1, 0, 0, 0, 441, 442, 10, 6, 0, 0, 442, 443, 7, 2, 0, 0, 
	443, 457, 3, 74, 37, 7, 444, 445, 10, 5, 0, 0, 445, 446, 7, 3, 0, 0, 446, 
	457, 3, 74, 37, 6, 447, 448, 10, 4, 0, 0, 448, 449, 5, 64, 0, 0, 449, 457, 
	3, 74, 37, 5, 450, 451, 10, 3, 0, 0, 451, 452, 5, 43, 0, 0, 452, 457, 3, 
	74, 37, 4, 453, 454, 10, 2, 0, 0, 454, 455, 5, 44, 0, 0, 455, 457, 3, 74, 
	37, 3, 456, 441, 1, 0, 0, 0, 456, 444, 1, 0, 0, 0, 456, 447, 1, 0, 0, 0, 
	456, 450, 1, 0, 0, 0, 456, 453, 1, 0, 0, 0, 457, 460, 1, 0, 0, 0, 458, 
	456, 1, 0, 0, 0, 458, 459, 1, 0, 0, 0, 459, 75, 1, 0, 0, 0, 460, 458, 1, 
	0, 0, 0, 49, 78, 80, 88, 104, 111, 116, 124, 139, 146, 153, 160, 164, 172, 
	204, 217, 228, 233, 240, 248, 255, 264, 267, 271, 280, 283, 287, 292, 297, 
	300, 302, 307, 310, 316, 320, 331, 341, 349, 360, 378, 381, 384, 397, 400, 
	412, 432, 435, 439, 456, 458,
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
	MiniQLParserEOF = antlr.TokenEOF
	MiniQLParserT__0 = 1
	MiniQLParserT__1 = 2
	MiniQLParserT__2 = 3
	MiniQLParserT__3 = 4
	MiniQLParserT__4 = 5
	MiniQLParserT__5 = 6
	MiniQLParserT__6 = 7
	MiniQLParserT__7 = 8
	MiniQLParserT__8 = 9
	MiniQLParserT__9 = 10
	MiniQLParserCREATE = 11
	MiniQLParserDATABASE = 12
	MiniQLParserTABLE = 13
	MiniQLParserINDEX = 14
	MiniQLParserDROP = 15
	MiniQLParserINSERT = 16
	MiniQLParserINTO = 17
	MiniQLParserVALUES = 18
	MiniQLParserSELECT = 19
	MiniQLParserFROM = 20
	MiniQLParserWHERE = 21
	MiniQLParserGROUP = 22
	MiniQLParserBY = 23
	MiniQLParserHAVING = 24
	MiniQLParserORDER = 25
	MiniQLParserLIMIT = 26
	MiniQLParserJOIN = 27
	MiniQLParserINNER = 28
	MiniQLParserLEFT = 29
	MiniQLParserON = 30
	MiniQLParserAS = 31
	MiniQLParserUSE = 32
	MiniQLParserSHOW = 33
	MiniQLParserEXPLAIN = 34
	MiniQLParserPRIMARY = 35
	MiniQLParserKEY = 36
	MiniQLParserPARTITION = 37
	MiniQLParserINT = 38
	MiniQLParserBIGINT = 39
	MiniQLParserVARCHAR = 40
	MiniQLParserDATE = 41
	MiniQLParserDOUBLE = 42
	MiniQLParserAND = 43
	MiniQLParserOR = 44
	MiniQLParserASC = 45
	MiniQLParserDESC = 46
	MiniQLParserNULL = 47
	MiniQLParserHASH = 48
	MiniQLParserRANGE = 49
	MiniQLParserANALYZE = 50
	MiniQLParserVERBOSE = 51
	MiniQLParserUPDATE = 52
	MiniQLParserSET = 53
	MiniQLParserDELETE = 54
	MiniQLParserDATABASES = 55
	MiniQLParserTABLES = 56
	MiniQLParserNOT = 57
	MiniQLParserUNIQUE = 58
	MiniQLParserDEFAULT = 59
	MiniQLParserIDENTIFIER = 60
	MiniQLParserSTRING = 61
	MiniQLParserINTEGER = 62
	MiniQLParserFLOAT = 63
	MiniQLParserCOMPARISON_OP = 64
	MiniQLParserWS = 65
)

// MiniQLParser rules.
const (
	MiniQLParserRULE_parse = 0
	MiniQLParserRULE_error = 1
	MiniQLParserRULE_sqlStatement = 2
	MiniQLParserRULE_ddlStatement = 3
	MiniQLParserRULE_dmlStatement = 4
	MiniQLParserRULE_dqlStatement = 5
	MiniQLParserRULE_utilityStatement = 6
	MiniQLParserRULE_createDatabase = 7
	MiniQLParserRULE_createTable = 8
	MiniQLParserRULE_columnDef = 9
	MiniQLParserRULE_columnConstraint = 10
	MiniQLParserRULE_tableConstraint = 11
	MiniQLParserRULE_createIndex = 12
	MiniQLParserRULE_dropTable = 13
	MiniQLParserRULE_dropDatabase = 14
	MiniQLParserRULE_insertStatement = 15
	MiniQLParserRULE_updateStatement = 16
	MiniQLParserRULE_deleteStatement = 17
	MiniQLParserRULE_selectStatement = 18
	MiniQLParserRULE_selectItem = 19
	MiniQLParserRULE_tableReference = 20
	MiniQLParserRULE_tableName = 21
	MiniQLParserRULE_identifierList = 22
	MiniQLParserRULE_valueList = 23
	MiniQLParserRULE_updateAssignment = 24
	MiniQLParserRULE_groupByItem = 25
	MiniQLParserRULE_orderByItem = 26
	MiniQLParserRULE_functionName = 27
	MiniQLParserRULE_literal = 28
	MiniQLParserRULE_identifier = 29
	MiniQLParserRULE_partitionMethod = 30
	MiniQLParserRULE_joinType = 31
	MiniQLParserRULE_useStatement = 32
	MiniQLParserRULE_showDatabases = 33
	MiniQLParserRULE_showTables = 34
	MiniQLParserRULE_explainStatement = 35
	MiniQLParserRULE_dataType = 36
	MiniQLParserRULE_expression = 37
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
	AllError_() []IErrorContext
	Error_(i int) IErrorContext

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

func InitEmptyParseContext(p *ParseContext)  {
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
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISqlStatementContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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

func (s *ParseContext) AllError_() []IErrorContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IErrorContext); ok {
			len++
		}
	}

	tst := make([]IErrorContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IErrorContext); ok {
			tst[i] = t.(IErrorContext)
			i++
		}
	}

	return tst
}

func (s *ParseContext) Error_(i int) IErrorContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IErrorContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IErrorContext)
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
	p.SetState(80)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for ((int64(_la) & ^0x3f) == 0 && ((int64(1) << _la) & -2) != 0) || _la == MiniQLParserCOMPARISON_OP || _la == MiniQLParserWS {
		p.SetState(78)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 0, p.GetParserRuleContext()) {
		case 1:
			{
				p.SetState(76)
				p.SqlStatement()
			}


		case 2:
			{
				p.SetState(77)
				p.Error_()
			}

		case antlr.ATNInvalidAltNumber:
			goto errorExit
		}

		p.SetState(82)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(83)
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


// IErrorContext is an interface to support dynamic dispatch.
type IErrorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsErrorContext differentiates from other interfaces.
	IsErrorContext()
}

type ErrorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyErrorContext() *ErrorContext {
	var p = new(ErrorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_error
	return p
}

func InitEmptyErrorContext(p *ErrorContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_error
}

func (*ErrorContext) IsErrorContext() {}

func NewErrorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ErrorContext {
	var p = new(ErrorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_error

	return p
}

func (s *ErrorContext) GetParser() antlr.Parser { return s.parser }
func (s *ErrorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ErrorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *ErrorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitError(s)

	default:
		return t.VisitChildren(s)
	}
}




func (p *MiniQLParser) Error_() (localctx IErrorContext) {
	localctx = NewErrorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, MiniQLParserRULE_error)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(86)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = 1+1
	for ok := true; ok; ok = _alt != 1 && _alt != antlr.ATNInvalidAltNumber {
		switch _alt {
		case 1+1:
				p.SetState(85)
				p.MatchWildcard()





		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(88)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	{
		p.SetState(90)
		p.Match(MiniQLParserT__0)
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
	UtilityStatement() IUtilityStatementContext

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

func InitEmptySqlStatementContext(p *SqlStatementContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDdlStatementContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDdlStatementContext)
}

func (s *SqlStatementContext) DmlStatement() IDmlStatementContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDmlStatementContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDmlStatementContext)
}

func (s *SqlStatementContext) DqlStatement() IDqlStatementContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDqlStatementContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDqlStatementContext)
}

func (s *SqlStatementContext) UtilityStatement() IUtilityStatementContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUtilityStatementContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUtilityStatementContext)
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
	p.EnterRule(localctx, 4, MiniQLParserRULE_sqlStatement)
	p.SetState(104)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MiniQLParserCREATE, MiniQLParserDROP:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(92)
			p.DdlStatement()
		}
		{
			p.SetState(93)
			p.Match(MiniQLParserT__0)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MiniQLParserINSERT, MiniQLParserUPDATE, MiniQLParserDELETE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(95)
			p.DmlStatement()
		}
		{
			p.SetState(96)
			p.Match(MiniQLParserT__0)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MiniQLParserSELECT:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(98)
			p.DqlStatement()
		}
		{
			p.SetState(99)
			p.Match(MiniQLParserT__0)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MiniQLParserUSE, MiniQLParserSHOW, MiniQLParserEXPLAIN:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(101)
			p.UtilityStatement()
		}
		{
			p.SetState(102)
			p.Match(MiniQLParserT__0)
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


// IDdlStatementContext is an interface to support dynamic dispatch.
type IDdlStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CreateDatabase() ICreateDatabaseContext
	CreateTable() ICreateTableContext
	CreateIndex() ICreateIndexContext
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

func InitEmptyDdlStatementContext(p *DdlStatementContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICreateDatabaseContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICreateDatabaseContext)
}

func (s *DdlStatementContext) CreateTable() ICreateTableContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICreateTableContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICreateTableContext)
}

func (s *DdlStatementContext) CreateIndex() ICreateIndexContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICreateIndexContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICreateIndexContext)
}

func (s *DdlStatementContext) DropTable() IDropTableContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDropTableContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IDropTableContext)
}

func (s *DdlStatementContext) DropDatabase() IDropDatabaseContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDropDatabaseContext); ok {
			t = ctx.(antlr.RuleContext);
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
	p.EnterRule(localctx, 6, MiniQLParserRULE_ddlStatement)
	p.SetState(111)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(106)
			p.CreateDatabase()
		}


	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(107)
			p.CreateTable()
		}


	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(108)
			p.CreateIndex()
		}


	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(109)
			p.DropTable()
		}


	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(110)
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

func InitEmptyDmlStatementContext(p *DmlStatementContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IInsertStatementContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IInsertStatementContext)
}

func (s *DmlStatementContext) UpdateStatement() IUpdateStatementContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUpdateStatementContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUpdateStatementContext)
}

func (s *DmlStatementContext) DeleteStatement() IDeleteStatementContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDeleteStatementContext); ok {
			t = ctx.(antlr.RuleContext);
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
	p.EnterRule(localctx, 8, MiniQLParserRULE_dmlStatement)
	p.SetState(116)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MiniQLParserINSERT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(113)
			p.InsertStatement()
		}


	case MiniQLParserUPDATE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(114)
			p.UpdateStatement()
		}


	case MiniQLParserDELETE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(115)
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

func InitEmptyDqlStatementContext(p *DqlStatementContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelectStatementContext); ok {
			t = ctx.(antlr.RuleContext);
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
	p.EnterRule(localctx, 10, MiniQLParserRULE_dqlStatement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(118)
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


// IUtilityStatementContext is an interface to support dynamic dispatch.
type IUtilityStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	UseStatement() IUseStatementContext
	ShowDatabases() IShowDatabasesContext
	ShowTables() IShowTablesContext
	ExplainStatement() IExplainStatementContext

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

func InitEmptyUtilityStatementContext(p *UtilityStatementContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUseStatementContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUseStatementContext)
}

func (s *UtilityStatementContext) ShowDatabases() IShowDatabasesContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IShowDatabasesContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IShowDatabasesContext)
}

func (s *UtilityStatementContext) ShowTables() IShowTablesContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IShowTablesContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IShowTablesContext)
}

func (s *UtilityStatementContext) ExplainStatement() IExplainStatementContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExplainStatementContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExplainStatementContext)
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
	p.SetState(124)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 6, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(120)
			p.UseStatement()
		}


	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(121)
			p.ShowDatabases()
		}


	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(122)
			p.ShowTables()
		}


	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(123)
			p.ExplainStatement()
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

func InitEmptyCreateDatabaseContext(p *CreateDatabaseContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext);
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
		p.SetState(126)
		p.Match(MiniQLParserCREATE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(127)
		p.Match(MiniQLParserDATABASE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(128)
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
	AllColumnDef() []IColumnDefContext
	ColumnDef(i int) IColumnDefContext
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

func InitEmptyCreateTableContext(p *CreateTableContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
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
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumnDefContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableConstraintContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPartitionMethodContext); ok {
			t = ctx.(antlr.RuleContext);
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
		p.SetState(130)
		p.Match(MiniQLParserCREATE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(131)
		p.Match(MiniQLParserTABLE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(132)
		p.TableName()
	}
	{
		p.SetState(133)
		p.Match(MiniQLParserT__1)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(134)
		p.ColumnDef()
	}
	p.SetState(139)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 7, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			{
				p.SetState(135)
				p.Match(MiniQLParserT__2)
				if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
				}
			}
			{
				p.SetState(136)
				p.ColumnDef()
			}


		}
		p.SetState(141)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 7, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}
	p.SetState(146)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for _la == MiniQLParserT__2 {
		{
			p.SetState(142)
			p.Match(MiniQLParserT__2)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(143)
			p.TableConstraint()
		}


		p.SetState(148)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(149)
		p.Match(MiniQLParserT__3)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(153)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MiniQLParserPARTITION {
		{
			p.SetState(150)
			p.Match(MiniQLParserPARTITION)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(151)
			p.Match(MiniQLParserBY)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(152)
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

func InitEmptyColumnDefContext(p *ColumnDefContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *ColumnDefContext) DataType() IDataTypeContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IDataTypeContext); ok {
			t = ctx.(antlr.RuleContext);
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
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColumnConstraintContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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
		p.SetState(155)
		p.Identifier()
	}
	{
		p.SetState(156)
		p.DataType()
	}
	p.SetState(160)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for ((int64(_la) & ^0x3f) == 0 && ((int64(1) << _la) & 1008947088379084800) != 0) {
		{
			p.SetState(157)
			p.ColumnConstraint()
		}


		p.SetState(162)
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

func InitEmptyColumnConstraintContext(p *ColumnConstraintContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteralContext); ok {
			t = ctx.(antlr.RuleContext);
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

	p.SetState(172)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MiniQLParserNULL, MiniQLParserNOT:
		p.EnterOuterAlt(localctx, 1)
		p.SetState(164)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)


		if _la == MiniQLParserNOT {
			{
				p.SetState(163)
				p.Match(MiniQLParserNOT)
				if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
				}
			}

		}
		{
			p.SetState(166)
			p.Match(MiniQLParserNULL)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MiniQLParserPRIMARY:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(167)
			p.Match(MiniQLParserPRIMARY)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(168)
			p.Match(MiniQLParserKEY)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MiniQLParserUNIQUE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(169)
			p.Match(MiniQLParserUNIQUE)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MiniQLParserDEFAULT:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(170)
			p.Match(MiniQLParserDEFAULT)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(171)
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
	IdentifierList() IIdentifierListContext

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

func InitEmptyTableConstraintContext(p *TableConstraintContext)  {
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

func (s *TableConstraintContext) IdentifierList() IIdentifierListContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierListContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierListContext)
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
		p.SetState(174)
		p.Match(MiniQLParserPRIMARY)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(175)
		p.Match(MiniQLParserKEY)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(176)
		p.Match(MiniQLParserT__1)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(177)
		p.IdentifierList()
	}
	{
		p.SetState(178)
		p.Match(MiniQLParserT__3)
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
	IdentifierList() IIdentifierListContext

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

func InitEmptyCreateIndexContext(p *CreateIndexContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext);
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
}

func (s *CreateIndexContext) IdentifierList() IIdentifierListContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierListContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierListContext)
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
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(180)
		p.Match(MiniQLParserCREATE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(181)
		p.Match(MiniQLParserINDEX)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(182)
		p.Identifier()
	}
	{
		p.SetState(183)
		p.Match(MiniQLParserON)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(184)
		p.TableName()
	}
	{
		p.SetState(185)
		p.Match(MiniQLParserT__1)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(186)
		p.IdentifierList()
	}
	{
		p.SetState(187)
		p.Match(MiniQLParserT__3)
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

func InitEmptyDropTableContext(p *DropTableContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext);
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
	p.EnterRule(localctx, 26, MiniQLParserRULE_dropTable)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(189)
		p.Match(MiniQLParserDROP)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(190)
		p.Match(MiniQLParserTABLE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(191)
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

func InitEmptyDropDatabaseContext(p *DropDatabaseContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext);
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
	p.EnterRule(localctx, 28, MiniQLParserRULE_dropDatabase)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(193)
		p.Match(MiniQLParserDROP)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(194)
		p.Match(MiniQLParserDATABASE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(195)
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
	AllValueList() []IValueListContext
	ValueList(i int) IValueListContext
	IdentifierList() IIdentifierListContext

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

func InitEmptyInsertStatementContext(p *InsertStatementContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext);
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
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueListContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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

func (s *InsertStatementContext) IdentifierList() IIdentifierListContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierListContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierListContext)
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
	p.EnterRule(localctx, 30, MiniQLParserRULE_insertStatement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(197)
		p.Match(MiniQLParserINSERT)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(198)
		p.Match(MiniQLParserINTO)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(199)
		p.TableName()
	}
	p.SetState(204)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MiniQLParserT__1 {
		{
			p.SetState(200)
			p.Match(MiniQLParserT__1)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(201)
			p.IdentifierList()
		}
		{
			p.SetState(202)
			p.Match(MiniQLParserT__3)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}

	}
	{
		p.SetState(206)
		p.Match(MiniQLParserVALUES)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(207)
		p.Match(MiniQLParserT__1)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(208)
		p.ValueList()
	}
	{
		p.SetState(209)
		p.Match(MiniQLParserT__3)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(217)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for _la == MiniQLParserT__2 {
		{
			p.SetState(210)
			p.Match(MiniQLParserT__2)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(211)
			p.Match(MiniQLParserT__1)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(212)
			p.ValueList()
		}
		{
			p.SetState(213)
			p.Match(MiniQLParserT__3)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


		p.SetState(219)
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

func InitEmptyUpdateStatementContext(p *UpdateStatementContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext);
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
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUpdateAssignmentContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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

func (s *UpdateStatementContext) WHERE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserWHERE, 0)
}

func (s *UpdateStatementContext) Expression() IExpressionContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext);
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
	p.EnterRule(localctx, 32, MiniQLParserRULE_updateStatement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(220)
		p.Match(MiniQLParserUPDATE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(221)
		p.TableName()
	}
	{
		p.SetState(222)
		p.Match(MiniQLParserSET)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(223)
		p.UpdateAssignment()
	}
	p.SetState(228)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for _la == MiniQLParserT__2 {
		{
			p.SetState(224)
			p.Match(MiniQLParserT__2)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(225)
			p.UpdateAssignment()
		}


		p.SetState(230)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(233)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MiniQLParserWHERE {
		{
			p.SetState(231)
			p.Match(MiniQLParserWHERE)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(232)
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

func InitEmptyDeleteStatementContext(p *DeleteStatementContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext);
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext);
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
	p.EnterRule(localctx, 34, MiniQLParserRULE_deleteStatement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(235)
		p.Match(MiniQLParserDELETE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(236)
		p.Match(MiniQLParserFROM)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(237)
		p.TableName()
	}
	p.SetState(240)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MiniQLParserWHERE {
		{
			p.SetState(238)
			p.Match(MiniQLParserWHERE)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(239)
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
	INTEGER() antlr.TerminalNode

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

func InitEmptySelectStatementContext(p *SelectStatementContext)  {
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
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelectItemContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableReferenceContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableReferenceContext)
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
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IGroupByItemContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrderByItemContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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

func (s *SelectStatementContext) INTEGER() antlr.TerminalNode {
	return s.GetToken(MiniQLParserINTEGER, 0)
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
	p.EnterRule(localctx, 36, MiniQLParserRULE_selectStatement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(242)
		p.Match(MiniQLParserSELECT)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(243)
		p.SelectItem()
	}
	p.SetState(248)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for _la == MiniQLParserT__2 {
		{
			p.SetState(244)
			p.Match(MiniQLParserT__2)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(245)
			p.SelectItem()
		}


		p.SetState(250)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(251)
		p.Match(MiniQLParserFROM)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(252)
		p.tableReference(0)
	}
	p.SetState(255)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MiniQLParserWHERE {
		{
			p.SetState(253)
			p.Match(MiniQLParserWHERE)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(254)
			p.expression(0)
		}

	}
	p.SetState(267)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MiniQLParserGROUP {
		{
			p.SetState(257)
			p.Match(MiniQLParserGROUP)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(258)
			p.Match(MiniQLParserBY)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(259)
			p.GroupByItem()
		}
		p.SetState(264)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)


		for _la == MiniQLParserT__2 {
			{
				p.SetState(260)
				p.Match(MiniQLParserT__2)
				if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
				}
			}
			{
				p.SetState(261)
				p.GroupByItem()
			}


			p.SetState(266)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
		    	goto errorExit
		    }
			_la = p.GetTokenStream().LA(1)
		}

	}
	p.SetState(271)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MiniQLParserHAVING {
		{
			p.SetState(269)
			p.Match(MiniQLParserHAVING)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(270)
			p.expression(0)
		}

	}
	p.SetState(283)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MiniQLParserORDER {
		{
			p.SetState(273)
			p.Match(MiniQLParserORDER)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(274)
			p.Match(MiniQLParserBY)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(275)
			p.OrderByItem()
		}
		p.SetState(280)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)


		for _la == MiniQLParserT__2 {
			{
				p.SetState(276)
				p.Match(MiniQLParserT__2)
				if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
				}
			}
			{
				p.SetState(277)
				p.OrderByItem()
			}


			p.SetState(282)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
		    	goto errorExit
		    }
			_la = p.GetTokenStream().LA(1)
		}

	}
	p.SetState(287)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MiniQLParserLIMIT {
		{
			p.SetState(285)
			p.Match(MiniQLParserLIMIT)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(286)
			p.Match(MiniQLParserINTEGER)
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

func InitEmptySelectItemContext(p *SelectItemContext)  {
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

func (s *SelectAllContext) TableName() ITableNameContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *SelectExprContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext);
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
	p.EnterRule(localctx, 38, MiniQLParserRULE_selectItem)
	var _la int

	p.SetState(302)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 29, p.GetParserRuleContext()) {
	case 1:
		localctx = NewSelectAllContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		p.SetState(292)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)


		if _la == MiniQLParserIDENTIFIER {
			{
				p.SetState(289)
				p.TableName()
			}
			{
				p.SetState(290)
				p.Match(MiniQLParserT__4)
				if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
				}
			}

		}
		{
			p.SetState(294)
			p.Match(MiniQLParserT__5)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case 2:
		localctx = NewSelectExprContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(295)
			p.expression(0)
		}
		p.SetState(300)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)


		if _la == MiniQLParserAS || _la == MiniQLParserIDENTIFIER {
			p.SetState(297)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)


			if _la == MiniQLParserAS {
				{
					p.SetState(296)
					p.Match(MiniQLParserAS)
					if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
					}
				}

			}
			{
				p.SetState(299)
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

func InitEmptyTableReferenceContext(p *TableReferenceContext)  {
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

func (s *TableReferenceContext) CopyAll(ctx *TableReferenceContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *TableReferenceContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TableReferenceContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}





type TableRefBaseContext struct {
	TableReferenceContext
}

func NewTableRefBaseContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *TableRefBaseContext {
	var p = new(TableRefBaseContext)

	InitEmptyTableReferenceContext(&p.TableReferenceContext)
	p.parser = parser
	p.CopyAll(ctx.(*TableReferenceContext))

	return p
}

func (s *TableRefBaseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TableRefBaseContext) TableName() ITableNameContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
}

func (s *TableRefBaseContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext);
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


type TableRefJoinContext struct {
	TableReferenceContext
}

func NewTableRefJoinContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *TableRefJoinContext {
	var p = new(TableRefJoinContext)

	InitEmptyTableReferenceContext(&p.TableReferenceContext)
	p.parser = parser
	p.CopyAll(ctx.(*TableReferenceContext))

	return p
}

func (s *TableRefJoinContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TableRefJoinContext) AllTableReference() []ITableReferenceContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITableReferenceContext); ok {
			len++
		}
	}

	tst := make([]ITableReferenceContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITableReferenceContext); ok {
			tst[i] = t.(ITableReferenceContext)
			i++
		}
	}

	return tst
}

func (s *TableRefJoinContext) TableReference(i int) ITableReferenceContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableReferenceContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableReferenceContext)
}

func (s *TableRefJoinContext) JoinType() IJoinTypeContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IJoinTypeContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IJoinTypeContext)
}

func (s *TableRefJoinContext) JOIN() antlr.TerminalNode {
	return s.GetToken(MiniQLParserJOIN, 0)
}

func (s *TableRefJoinContext) ON() antlr.TerminalNode {
	return s.GetToken(MiniQLParserON, 0)
}

func (s *TableRefJoinContext) Expression() IExpressionContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}


func (s *TableRefJoinContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitTableRefJoin(s)

	default:
		return t.VisitChildren(s)
	}
}


type TableRefSubqueryContext struct {
	TableReferenceContext
}

func NewTableRefSubqueryContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *TableRefSubqueryContext {
	var p = new(TableRefSubqueryContext)

	InitEmptyTableReferenceContext(&p.TableReferenceContext)
	p.parser = parser
	p.CopyAll(ctx.(*TableReferenceContext))

	return p
}

func (s *TableRefSubqueryContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TableRefSubqueryContext) SelectStatement() ISelectStatementContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISelectStatementContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISelectStatementContext)
}

func (s *TableRefSubqueryContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext);
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



func (p *MiniQLParser) TableReference() (localctx ITableReferenceContext) {
	return p.tableReference(0)
}

func (p *MiniQLParser) tableReference(_p int) (localctx ITableReferenceContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewTableReferenceContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx ITableReferenceContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 40
	p.EnterRecursionRule(localctx, 40, MiniQLParserRULE_tableReference, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(320)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MiniQLParserIDENTIFIER:
		localctx = NewTableRefBaseContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(305)
			p.TableName()
		}
		p.SetState(310)
		p.GetErrorHandler().Sync(p)


		if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 31, p.GetParserRuleContext()) == 1 {
			p.SetState(307)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)


			if _la == MiniQLParserAS {
				{
					p.SetState(306)
					p.Match(MiniQLParserAS)
					if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
					}
				}

			}
			{
				p.SetState(309)
				p.Identifier()
			}

			} else if p.HasError() { // JIM
				goto errorExit
		}


	case MiniQLParserT__1:
		localctx = NewTableRefSubqueryContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(312)
			p.Match(MiniQLParserT__1)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(313)
			p.SelectStatement()
		}
		{
			p.SetState(314)
			p.Match(MiniQLParserT__3)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		p.SetState(316)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)


		if _la == MiniQLParserAS {
			{
				p.SetState(315)
				p.Match(MiniQLParserAS)
				if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
				}
			}

		}
		{
			p.SetState(318)
			p.Identifier()
		}



	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(331)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 34, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			localctx = NewTableRefJoinContext(p, NewTableReferenceContext(p, _parentctx, _parentState))
			p.PushNewRecursionContext(localctx, _startState, MiniQLParserRULE_tableReference)
			p.SetState(322)

			if !(p.Precpred(p.GetParserRuleContext(), 1)) {
				p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 1)", ""))
				goto errorExit
			}

			{
				p.SetState(323)
				p.JoinType()
			}
			{
				p.SetState(324)
				p.Match(MiniQLParserJOIN)
				if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
				}
			}
			{
				p.SetState(325)
				p.tableReference(0)
			}
			{
				p.SetState(326)
				p.Match(MiniQLParserON)
				if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
				}
			}
			{
				p.SetState(327)
				p.expression(0)
			}



		}
		p.SetState(333)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 34, p.GetParserRuleContext())
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


// ITableNameContext is an interface to support dynamic dispatch.
type ITableNameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext

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

func InitEmptyTableNameContext(p *TableNameContext)  {
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

func (s *TableNameContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
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
	p.EnterRule(localctx, 42, MiniQLParserRULE_tableName)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(334)
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


// IIdentifierListContext is an interface to support dynamic dispatch.
type IIdentifierListContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllIdentifier() []IIdentifierContext
	Identifier(i int) IIdentifierContext

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

func InitEmptyIdentifierListContext(p *IdentifierListContext)  {
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
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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
	p.EnterRule(localctx, 44, MiniQLParserRULE_identifierList)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(336)
		p.Identifier()
	}
	p.SetState(341)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for _la == MiniQLParserT__2 {
		{
			p.SetState(337)
			p.Match(MiniQLParserT__2)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(338)
			p.Identifier()
		}


		p.SetState(343)
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
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext

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

func InitEmptyValueListContext(p *ValueListContext)  {
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

func (s *ValueListContext) AllExpression() []IExpressionContext {
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

func (s *ValueListContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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
	p.EnterRule(localctx, 46, MiniQLParserRULE_valueList)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(344)
		p.expression(0)
	}
	p.SetState(349)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	for _la == MiniQLParserT__2 {
		{
			p.SetState(345)
			p.Match(MiniQLParserT__2)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(346)
			p.expression(0)
		}


		p.SetState(351)
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


// IUpdateAssignmentContext is an interface to support dynamic dispatch.
type IUpdateAssignmentContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext
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

func InitEmptyUpdateAssignmentContext(p *UpdateAssignmentContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *UpdateAssignmentContext) Expression() IExpressionContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext);
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
	p.EnterRule(localctx, 48, MiniQLParserRULE_updateAssignment)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(352)
		p.Identifier()
	}
	{
		p.SetState(353)
		p.Match(MiniQLParserT__6)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(354)
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

func InitEmptyGroupByItemContext(p *GroupByItemContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext);
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
	p.EnterRule(localctx, 50, MiniQLParserRULE_groupByItem)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(356)
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

func InitEmptyOrderByItemContext(p *OrderByItemContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext);
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
	p.EnterRule(localctx, 52, MiniQLParserRULE_orderByItem)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(358)
		p.expression(0)
	}
	p.SetState(360)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MiniQLParserASC || _la == MiniQLParserDESC {
		{
			p.SetState(359)
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


// IFunctionNameContext is an interface to support dynamic dispatch.
type IFunctionNameContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Identifier() IIdentifierContext

	// IsFunctionNameContext differentiates from other interfaces.
	IsFunctionNameContext()
}

type FunctionNameContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFunctionNameContext() *FunctionNameContext {
	var p = new(FunctionNameContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_functionName
	return p
}

func InitEmptyFunctionNameContext(p *FunctionNameContext)  {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = MiniQLParserRULE_functionName
}

func (*FunctionNameContext) IsFunctionNameContext() {}

func NewFunctionNameContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FunctionNameContext {
	var p = new(FunctionNameContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = MiniQLParserRULE_functionName

	return p
}

func (s *FunctionNameContext) GetParser() antlr.Parser { return s.parser }

func (s *FunctionNameContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *FunctionNameContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FunctionNameContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}


func (s *FunctionNameContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitFunctionName(s)

	default:
		return t.VisitChildren(s)
	}
}




func (p *MiniQLParser) FunctionName() (localctx IFunctionNameContext) {
	localctx = NewFunctionNameContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 54, MiniQLParserRULE_functionName)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(362)
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


// ILiteralContext is an interface to support dynamic dispatch.
type ILiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STRING() antlr.TerminalNode
	INTEGER() antlr.TerminalNode
	FLOAT() antlr.TerminalNode
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

func InitEmptyLiteralContext(p *LiteralContext)  {
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

func (s *LiteralContext) STRING() antlr.TerminalNode {
	return s.GetToken(MiniQLParserSTRING, 0)
}

func (s *LiteralContext) INTEGER() antlr.TerminalNode {
	return s.GetToken(MiniQLParserINTEGER, 0)
}

func (s *LiteralContext) FLOAT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserFLOAT, 0)
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
	p.EnterRule(localctx, 56, MiniQLParserRULE_literal)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(364)
		_la = p.GetTokenStream().LA(1)

		if !(((int64(_la) & ^0x3f) == 0 && ((int64(1) << _la) & -2305702271725338624) != 0)) {
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

func InitEmptyIdentifierContext(p *IdentifierContext)  {
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
	p.EnterRule(localctx, 58, MiniQLParserRULE_identifier)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(366)
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


// IPartitionMethodContext is an interface to support dynamic dispatch.
type IPartitionMethodContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	HASH() antlr.TerminalNode
	IdentifierList() IIdentifierListContext
	RANGE() antlr.TerminalNode
	Expression() IExpressionContext

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

func InitEmptyPartitionMethodContext(p *PartitionMethodContext)  {
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

func (s *PartitionMethodContext) IdentifierList() IIdentifierListContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierListContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierListContext)
}

func (s *PartitionMethodContext) RANGE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserRANGE, 0)
}

func (s *PartitionMethodContext) Expression() IExpressionContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
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
	p.EnterRule(localctx, 60, MiniQLParserRULE_partitionMethod)
	p.SetState(378)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MiniQLParserHASH:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(368)
			p.Match(MiniQLParserHASH)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(369)
			p.Match(MiniQLParserT__1)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(370)
			p.IdentifierList()
		}
		{
			p.SetState(371)
			p.Match(MiniQLParserT__3)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MiniQLParserRANGE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(373)
			p.Match(MiniQLParserRANGE)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(374)
			p.Match(MiniQLParserT__1)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(375)
			p.expression(0)
		}
		{
			p.SetState(376)
			p.Match(MiniQLParserT__3)
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


// IJoinTypeContext is an interface to support dynamic dispatch.
type IJoinTypeContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	INNER() antlr.TerminalNode
	LEFT() antlr.TerminalNode

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

func InitEmptyJoinTypeContext(p *JoinTypeContext)  {
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
	p.EnterRule(localctx, 62, MiniQLParserRULE_joinType)
	var _la int

	p.SetState(384)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MiniQLParserJOIN, MiniQLParserINNER:
		p.EnterOuterAlt(localctx, 1)
		p.SetState(381)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)


		if _la == MiniQLParserINNER {
			{
				p.SetState(380)
				p.Match(MiniQLParserINNER)
				if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
				}
			}

		}


	case MiniQLParserLEFT:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(383)
			p.Match(MiniQLParserLEFT)
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

func InitEmptyUseStatementContext(p *UseStatementContext)  {
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
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext);
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
	p.EnterRule(localctx, 64, MiniQLParserRULE_useStatement)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(386)
		p.Match(MiniQLParserUSE)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(387)
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

func InitEmptyShowDatabasesContext(p *ShowDatabasesContext)  {
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
	p.EnterRule(localctx, 66, MiniQLParserRULE_showDatabases)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(389)
		p.Match(MiniQLParserSHOW)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(390)
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

func InitEmptyShowTablesContext(p *ShowTablesContext)  {
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
	p.EnterRule(localctx, 68, MiniQLParserRULE_showTables)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(392)
		p.Match(MiniQLParserSHOW)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	{
		p.SetState(393)
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


// IExplainStatementContext is an interface to support dynamic dispatch.
type IExplainStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EXPLAIN() antlr.TerminalNode
	SqlStatement() ISqlStatementContext
	ANALYZE() antlr.TerminalNode
	VERBOSE() antlr.TerminalNode

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

func InitEmptyExplainStatementContext(p *ExplainStatementContext)  {
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

func (s *ExplainStatementContext) SqlStatement() ISqlStatementContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISqlStatementContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISqlStatementContext)
}

func (s *ExplainStatementContext) ANALYZE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserANALYZE, 0)
}

func (s *ExplainStatementContext) VERBOSE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserVERBOSE, 0)
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
	p.EnterRule(localctx, 70, MiniQLParserRULE_explainStatement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(395)
		p.Match(MiniQLParserEXPLAIN)
		if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
		}
	}
	p.SetState(397)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MiniQLParserANALYZE {
		{
			p.SetState(396)
			p.Match(MiniQLParserANALYZE)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}

	}
	p.SetState(400)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)


	if _la == MiniQLParserVERBOSE {
		{
			p.SetState(399)
			p.Match(MiniQLParserVERBOSE)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}

	}
	{
		p.SetState(402)
		p.SqlStatement()
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
	INT() antlr.TerminalNode
	BIGINT() antlr.TerminalNode
	VARCHAR() antlr.TerminalNode
	INTEGER() antlr.TerminalNode
	DATE() antlr.TerminalNode
	DOUBLE() antlr.TerminalNode

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

func InitEmptyDataTypeContext(p *DataTypeContext)  {
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

func (s *DataTypeContext) INT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserINT, 0)
}

func (s *DataTypeContext) BIGINT() antlr.TerminalNode {
	return s.GetToken(MiniQLParserBIGINT, 0)
}

func (s *DataTypeContext) VARCHAR() antlr.TerminalNode {
	return s.GetToken(MiniQLParserVARCHAR, 0)
}

func (s *DataTypeContext) INTEGER() antlr.TerminalNode {
	return s.GetToken(MiniQLParserINTEGER, 0)
}

func (s *DataTypeContext) DATE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDATE, 0)
}

func (s *DataTypeContext) DOUBLE() antlr.TerminalNode {
	return s.GetToken(MiniQLParserDOUBLE, 0)
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
	p.EnterRule(localctx, 72, MiniQLParserRULE_dataType)
	p.SetState(412)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case MiniQLParserINT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(404)
			p.Match(MiniQLParserINT)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MiniQLParserBIGINT:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(405)
			p.Match(MiniQLParserBIGINT)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MiniQLParserVARCHAR:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(406)
			p.Match(MiniQLParserVARCHAR)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(407)
			p.Match(MiniQLParserT__1)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(408)
			p.Match(MiniQLParserINTEGER)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(409)
			p.Match(MiniQLParserT__3)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MiniQLParserDATE:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(410)
			p.Match(MiniQLParserDATE)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case MiniQLParserDOUBLE:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(411)
			p.Match(MiniQLParserDOUBLE)
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

func InitEmptyExpressionContext(p *ExpressionContext)  {
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





type BinaryArithExprContext struct {
	ExpressionContext
	left IExpressionContext 
	operator antlr.Token
	right IExpressionContext 
}

func NewBinaryArithExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *BinaryArithExprContext {
	var p = new(BinaryArithExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}


func (s *BinaryArithExprContext) GetOperator() antlr.Token { return s.operator }


func (s *BinaryArithExprContext) SetOperator(v antlr.Token) { s.operator = v }


func (s *BinaryArithExprContext) GetLeft() IExpressionContext { return s.left }

func (s *BinaryArithExprContext) GetRight() IExpressionContext { return s.right }


func (s *BinaryArithExprContext) SetLeft(v IExpressionContext) { s.left = v }

func (s *BinaryArithExprContext) SetRight(v IExpressionContext) { s.right = v }

func (s *BinaryArithExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BinaryArithExprContext) AllExpression() []IExpressionContext {
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

func (s *BinaryArithExprContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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


func (s *BinaryArithExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitBinaryArithExpr(s)

	default:
		return t.VisitChildren(s)
	}
}


type QualifiedColumnRefContext struct {
	ExpressionContext
}

func NewQualifiedColumnRefContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *QualifiedColumnRefContext {
	var p = new(QualifiedColumnRefContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *QualifiedColumnRefContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *QualifiedColumnRefContext) TableName() ITableNameContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITableNameContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITableNameContext)
}

func (s *QualifiedColumnRefContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}


func (s *QualifiedColumnRefContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitQualifiedColumnRef(s)

	default:
		return t.VisitChildren(s)
	}
}


type LiteralExprContext struct {
	ExpressionContext
}

func NewLiteralExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LiteralExprContext {
	var p = new(LiteralExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *LiteralExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LiteralExprContext) Literal() ILiteralContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteralContext); ok {
			t = ctx.(antlr.RuleContext);
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


type FunctionCallContext struct {
	ExpressionContext
}

func NewFunctionCallContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *FunctionCallContext {
	var p = new(FunctionCallContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *FunctionCallContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FunctionCallContext) FunctionName() IFunctionNameContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionNameContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionNameContext)
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
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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


func (s *FunctionCallContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitFunctionCall(s)

	default:
		return t.VisitChildren(s)
	}
}


type LogicalExprContext struct {
	ExpressionContext
}

func NewLogicalExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LogicalExprContext {
	var p = new(LogicalExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *LogicalExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LogicalExprContext) AllExpression() []IExpressionContext {
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

func (s *LogicalExprContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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

func (s *LogicalExprContext) AND() antlr.TerminalNode {
	return s.GetToken(MiniQLParserAND, 0)
}

func (s *LogicalExprContext) OR() antlr.TerminalNode {
	return s.GetToken(MiniQLParserOR, 0)
}


func (s *LogicalExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitLogicalExpr(s)

	default:
		return t.VisitChildren(s)
	}
}


type NestedExprContext struct {
	ExpressionContext
}

func NewNestedExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *NestedExprContext {
	var p = new(NestedExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *NestedExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NestedExprContext) Expression() IExpressionContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}


func (s *NestedExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitNestedExpr(s)

	default:
		return t.VisitChildren(s)
	}
}


type ComparisonExprContext struct {
	ExpressionContext
	left IExpressionContext 
	operator antlr.Token
	right IExpressionContext 
}

func NewComparisonExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ComparisonExprContext {
	var p = new(ComparisonExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}


func (s *ComparisonExprContext) GetOperator() antlr.Token { return s.operator }


func (s *ComparisonExprContext) SetOperator(v antlr.Token) { s.operator = v }


func (s *ComparisonExprContext) GetLeft() IExpressionContext { return s.left }

func (s *ComparisonExprContext) GetRight() IExpressionContext { return s.right }


func (s *ComparisonExprContext) SetLeft(v IExpressionContext) { s.left = v }

func (s *ComparisonExprContext) SetRight(v IExpressionContext) { s.right = v }

func (s *ComparisonExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ComparisonExprContext) AllExpression() []IExpressionContext {
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

func (s *ComparisonExprContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext;
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext);
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

func (s *ComparisonExprContext) COMPARISON_OP() antlr.TerminalNode {
	return s.GetToken(MiniQLParserCOMPARISON_OP, 0)
}


func (s *ComparisonExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitComparisonExpr(s)

	default:
		return t.VisitChildren(s)
	}
}


type ColumnRefExprContext struct {
	ExpressionContext
}

func NewColumnRefExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ColumnRefExprContext {
	var p = new(ColumnRefExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *ColumnRefExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ColumnRefExprContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext;
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext);
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}


func (s *ColumnRefExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case MiniQLVisitor:
		return t.VisitColumnRefExpr(s)

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
	_startState := 74
	p.EnterRecursionRule(localctx, 74, MiniQLParserRULE_expression, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(439)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 46, p.GetParserRuleContext()) {
	case 1:
		localctx = NewLiteralExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(415)
			p.Literal()
		}


	case 2:
		localctx = NewColumnRefExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(416)
			p.Identifier()
		}


	case 3:
		localctx = NewQualifiedColumnRefContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(417)
			p.TableName()
		}
		{
			p.SetState(418)
			p.Match(MiniQLParserT__4)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(419)
			p.Identifier()
		}


	case 4:
		localctx = NewNestedExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(421)
			p.Match(MiniQLParserT__1)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		{
			p.SetState(422)
			p.expression(0)
		}
		{
			p.SetState(423)
			p.Match(MiniQLParserT__3)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}


	case 5:
		localctx = NewFunctionCallContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(425)
			p.FunctionName()
		}
		{
			p.SetState(426)
			p.Match(MiniQLParserT__1)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}
		p.SetState(435)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)


		if ((int64(_la) & ^0x3f) == 0 && ((int64(1) << _la) & -1152780767118491644) != 0) {
			{
				p.SetState(427)
				p.expression(0)
			}
			p.SetState(432)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)


			for _la == MiniQLParserT__2 {
				{
					p.SetState(428)
					p.Match(MiniQLParserT__2)
					if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
					}
				}
				{
					p.SetState(429)
					p.expression(0)
				}


				p.SetState(434)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
			    	goto errorExit
			    }
				_la = p.GetTokenStream().LA(1)
			}

		}
		{
			p.SetState(437)
			p.Match(MiniQLParserT__3)
			if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(458)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 48, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(456)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 47, p.GetParserRuleContext()) {
			case 1:
				localctx = NewBinaryArithExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				localctx.(*BinaryArithExprContext).left = _prevctx


				p.PushNewRecursionContext(localctx, _startState, MiniQLParserRULE_expression)
				p.SetState(441)

				if !(p.Precpred(p.GetParserRuleContext(), 6)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 6)", ""))
					goto errorExit
				}
				{
					p.SetState(442)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*BinaryArithExprContext).operator = _lt

					_la = p.GetTokenStream().LA(1)

					if !(_la == MiniQLParserT__5 || _la == MiniQLParserT__7) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*BinaryArithExprContext).operator = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(443)

					var _x = p.expression(7)

					localctx.(*BinaryArithExprContext).right = _x
				}


			case 2:
				localctx = NewBinaryArithExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				localctx.(*BinaryArithExprContext).left = _prevctx


				p.PushNewRecursionContext(localctx, _startState, MiniQLParserRULE_expression)
				p.SetState(444)

				if !(p.Precpred(p.GetParserRuleContext(), 5)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 5)", ""))
					goto errorExit
				}
				{
					p.SetState(445)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*BinaryArithExprContext).operator = _lt

					_la = p.GetTokenStream().LA(1)

					if !(_la == MiniQLParserT__8 || _la == MiniQLParserT__9) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*BinaryArithExprContext).operator = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(446)

					var _x = p.expression(6)

					localctx.(*BinaryArithExprContext).right = _x
				}


			case 3:
				localctx = NewComparisonExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				localctx.(*ComparisonExprContext).left = _prevctx


				p.PushNewRecursionContext(localctx, _startState, MiniQLParserRULE_expression)
				p.SetState(447)

				if !(p.Precpred(p.GetParserRuleContext(), 4)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 4)", ""))
					goto errorExit
				}
				{
					p.SetState(448)

					var _m = p.Match(MiniQLParserCOMPARISON_OP)

					localctx.(*ComparisonExprContext).operator = _m
					if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
					}
				}
				{
					p.SetState(449)

					var _x = p.expression(5)

					localctx.(*ComparisonExprContext).right = _x
				}


			case 4:
				localctx = NewLogicalExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, MiniQLParserRULE_expression)
				p.SetState(450)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
					goto errorExit
				}
				{
					p.SetState(451)
					p.Match(MiniQLParserAND)
					if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
					}
				}
				{
					p.SetState(452)
					p.expression(4)
				}


			case 5:
				localctx = NewLogicalExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, MiniQLParserRULE_expression)
				p.SetState(453)

				if !(p.Precpred(p.GetParserRuleContext(), 2)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
					goto errorExit
				}
				{
					p.SetState(454)
					p.Match(MiniQLParserOR)
					if p.HasError() {
							// Recognition error - abort rule
							goto errorExit
					}
				}
				{
					p.SetState(455)
					p.expression(3)
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(460)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
	    	goto errorExit
	    }
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 48, p.GetParserRuleContext())
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


func (p *MiniQLParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 20:
			var t *TableReferenceContext = nil
			if localctx != nil { t = localctx.(*TableReferenceContext) }
			return p.TableReference_Sempred(t, predIndex)

	case 37:
			var t *ExpressionContext = nil
			if localctx != nil { t = localctx.(*ExpressionContext) }
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
			return p.Precpred(p.GetParserRuleContext(), 6)

	case 2:
			return p.Precpred(p.GetParserRuleContext(), 5)

	case 3:
			return p.Precpred(p.GetParserRuleContext(), 4)

	case 4:
			return p.Precpred(p.GetParserRuleContext(), 3)

	case 5:
			return p.Precpred(p.GetParserRuleContext(), 2)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}


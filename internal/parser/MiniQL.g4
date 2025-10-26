/*
 * MiniQL语法定义文件
 * 这是一个简化版SQL语法，支持基本的DDL、DML、DQL和DCL操作
 */

grammar MiniQL;

/*====================================================
 * 词法规则部分
 * 注意：为了避免关键词和标识符冲突，将关键词定义放在IDENTIFIER前。
 *====================================================*/

// 注释规则（放在最前面，优先级最高）
SINGLE_LINE_COMMENT: '--' ~[\r\n]* -> skip;
MULTI_LINE_COMMENT: '/*' .*? '*/' -> skip;

// 关键字
SELECT: S E L E C T;
FROM: F R O M;
WHERE: W H E R E;
GROUP: G R O U P;
BY: B Y;
HAVING: H A V I N G;
ORDER: O R D E R;
LIMIT: L I M I T;
INSERT: I N S E R T;
INTO: I N T O;
VALUES: V A L U E S;
UPDATE: U P D A T E;
SET: S E T;
DELETE: D E L E T E;
CREATE: C R E A T E;
TABLE: T A B L E;
DATABASE: D A T A B A S E;
DROP: D R O P;
PRIMARY: P R I M A R Y;
KEY: K E Y;
NOT: N O T;
NULL: N U L L;
AS: A S;
LIKE: L I K E;
IN: I N;
AND: A N D;
OR: O R;
JOIN: J O I N;
ON: O N;
PARTITION: P A R T I T I O N;
ASC: A S C;
DESC: D E S C;
INNER: I N N E R;
LEFT: L E F T;
RIGHT: R I G H T;
FULL: F U L L;
OUTER: O U T E R;
USE: U S E;
SHOW: S H O W;
DATABASES: D A T A B A S E S;
TABLES: T A B L E S;
EXPLAIN: E X P L A I N;
ANALYZE: A N A L Y Z E;
VERBOSE: V E R B O S E;
UNIQUE: U N I Q U E;
DEFAULT: D E F A U L T;
INDEX: I N D E X;
INDEXES: I N D E X E S;

// 数据类型关键字（必须放在IDENTIFIER前）
INTEGER_TYPE: I N T E G E R;
VARCHAR_TYPE: V A R C H A R;
BOOLEAN_TYPE: B O O L E A N;
DOUBLE_TYPE: D O U B L E;
TIMESTAMP_TYPE: T I M E S T A M P;

// 事务相关关键字
START: S T A R T;
TRANSACTION: T R A N S A C T I O N;
COMMIT: C O M M I T;
ROLLBACK: R O L L B A C K;

// 其他关键字
HASH: H A S H;
RANGE: R A N G E;

// 运算符和标点符号
ASTERISK: '*';
EQUAL: '=';
NOT_EQUAL: '!=';
GREATER: '>';
GREATER_EQUAL: '>=';
LESS: '<';
LESS_EQUAL: '<=';
PLUS: '+';
MINUS: '-';
MULTIPLY: '*';
DIVIDE: '/';
DOT: '.';
COMMA: ',';
SEMICOLON: ';';
LEFT_PAREN: '(';
RIGHT_PAREN: ')';

// 标识符
IDENTIFIER: [a-zA-Z_][a-zA-Z0-9_]*;

// 字面量
INTEGER_LITERAL: [0-9]+;
FLOAT_LITERAL: [0-9]+ '.' [0-9]*;
STRING_LITERAL: '\'' (~['\\] | '\\' .)* '\'';

// 空白字符
WS: [ \t\r\n]+ -> skip;

// 大小写不敏感的字母匹配
fragment A: [aA];
fragment B: [bB];
fragment C: [cC];
fragment D: [dD];
fragment E: [eE];
fragment F: [fF];
fragment G: [gG];
fragment H: [hH];
fragment I: [iI];
fragment J: [jJ];
fragment K: [kK];
fragment L: [lL];
fragment M: [mM];
fragment N: [nN];
fragment O: [oO];
fragment P: [pP];
fragment Q: [qQ];
fragment R: [rR];
fragment S: [sS];
fragment T: [tT];
fragment U: [uU];
fragment V: [vV];
fragment W: [wW];
fragment X: [xX];
fragment Y: [yY];
fragment Z: [zZ];

/*====================================================
 * 语法规则部分
 *====================================================*/

parse
 : sqlStatement* EOF
 ;

sqlStatement
 : (ddlStatement | dmlStatement | dqlStatement | dclStatement | utilityStatement) SEMICOLON?
 ;

// 语句类型定义
ddlStatement
 : createDatabase
 | createTable
 | createIndex
 | dropIndex
 | dropTable
 | dropDatabase
 ;

dmlStatement
 : insertStatement
 | updateStatement
 | deleteStatement
 ;

dqlStatement
 : selectStatement
 ;

dclStatement
 : transactionStatement
 ;

utilityStatement
 : useStatement
 | showDatabases
 | showTables
 | showIndexes
 | explainStatement
 | analyzeStatement
 ;

// DDL规则
createDatabase
 : CREATE DATABASE identifier
 ;

createTable
 : CREATE TABLE tableName
   LEFT_PAREN columnDef (COMMA columnDef)* (COMMA tableConstraint)* RIGHT_PAREN
   (PARTITION BY partitionMethod)?
 ;

columnDef
 : identifier dataType columnConstraint*
 ;

columnConstraint
 : (NOT? NULL)
 | PRIMARY KEY
 | UNIQUE
 | DEFAULT literal
 ;

tableConstraint
 : PRIMARY KEY LEFT_PAREN identifierList RIGHT_PAREN
 ;

createIndex
 : CREATE UNIQUE? INDEX identifier ON tableName LEFT_PAREN identifierList RIGHT_PAREN
 ;

dropIndex
 : DROP INDEX identifier ON tableName
 ;

dropTable
 : DROP TABLE tableName
 ;

dropDatabase
 : DROP DATABASE identifier
 ;

// DML规则
insertStatement
 : INSERT INTO tableName (LEFT_PAREN identifierList RIGHT_PAREN)? 
   VALUES LEFT_PAREN valueList (COMMA valueList)* RIGHT_PAREN
 ;

updateStatement
 : UPDATE tableName SET updateAssignment (COMMA updateAssignment)* 
   (WHERE expression)?
 ;

deleteStatement
 : DELETE FROM tableName (WHERE expression)?
 ;

// DQL规则
selectStatement
 : SELECT selectItem (COMMA selectItem)*
   FROM tableReference
   (WHERE expression)?
   (GROUP BY groupByItem (COMMA groupByItem)*)?
   (HAVING expression)?
   (ORDER BY orderByItem (COMMA orderByItem)*)?
   (LIMIT INTEGER_LITERAL)?
 ;

// 查询项定义
selectItem
 : (tableName DOT)? ASTERISK    #selectAll
 | expression (AS? identifier)?  #selectExpr
 ;

// 表引用
tableReference
 : tableReferenceAtom
 | tableReference joinType? JOIN tableReferenceAtom ON expression
 ;

tableReferenceAtom
 : tableName ( AS? identifier )?                                      #tableRefBase
 | LEFT_PAREN selectStatement RIGHT_PAREN AS? identifier             #tableRefSubquery
 ;

// JOIN类型
joinType
 : INNER
 | LEFT OUTER?
 | RIGHT OUTER?
 | FULL OUTER?
 ;

// 表达式规则
expression
 : primaryExpr                                                      #primaryExpression
 | expression comparisonOperator expression                         #comparisonExpression
 | expression AND expression                                        #andExpression
 | expression OR expression                                         #orExpression
 | expression (NOT)? LIKE expression                               #likeExpression
 | expression (NOT)? IN LEFT_PAREN valueList RIGHT_PAREN           #inExpression
 ;

primaryExpr
 : literal                                                         #literalExpr
 | columnRef                                                       #columnRefExpr
 | functionCall                                                    #functionCallExpr
 | LEFT_PAREN expression RIGHT_PAREN                              #parenExpr
 ;

comparisonOperator
 : EQUAL
 | NOT_EQUAL
 | GREATER
 | GREATER_EQUAL
 | LESS
 | LESS_EQUAL
 ;

// 其他规则定义
columnRef
 : identifier
 | identifier DOT identifier
 ;

updateAssignment
 : identifier EQUAL expression
 ;

groupByItem
 : expression
 ;

orderByItem
 : expression (ASC | DESC)?
 ;

functionCall
 : identifier LEFT_PAREN (ASTERISK | expression (COMMA expression)*)? RIGHT_PAREN
 ;

partitionMethod
 : HASH LEFT_PAREN identifierList RIGHT_PAREN
 | RANGE LEFT_PAREN identifierList RIGHT_PAREN
 ;

// DCL语句（事务控制）
transactionStatement
 : START TRANSACTION
 | COMMIT
 | ROLLBACK
 ;

// 使用、显示、解释语句
useStatement
 : USE identifier
 ;

showDatabases
 : SHOW DATABASES
 ;

showTables
 : SHOW TABLES
 ;

showIndexes
 : SHOW INDEXES (ON | FROM) tableName
 ;

explainStatement
 : EXPLAIN selectStatement
 ;

analyzeStatement
 : ANALYZE TABLE tableName (LEFT_PAREN columnList RIGHT_PAREN)?
 ;

columnList
 : identifier (COMMA identifier)*
 ;

// 辅助规则
identifierList
 : identifier (COMMA identifier)*
 ;

valueList
 : literal (COMMA literal)*
 ;

tableName
 : identifier (DOT identifier)?
 ;

identifier
 : IDENTIFIER
 ;

dataType
 : INTEGER_TYPE
 | VARCHAR_TYPE (LEFT_PAREN INTEGER_LITERAL RIGHT_PAREN)?
 | BOOLEAN_TYPE
 | DOUBLE_TYPE
 | TIMESTAMP_TYPE
 ;

literal
 : INTEGER_LITERAL
 | FLOAT_LITERAL
 | STRING_LITERAL
 ;
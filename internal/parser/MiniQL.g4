/*
 * MiniQL语法定义文件
 * 这是一个简化版SQL语法，支持基本的DDL、DML和DQL操作
 */

grammar MiniQL;

/*====================================================
 * 顶层语法规则
 *====================================================*/

// 解析的入口点，可以处理多个SQL语句或错误，直到文件结束
parse
 : ( sqlStatement | error )* EOF
 ;

// 错误处理规则，匹配任意字符直到分号
error
 : .+? ';'
 ;

// SQL语句的主要类型
sqlStatement
 : ddlStatement ';'      // 数据定义语言(CREATE, DROP等)
 | dmlStatement ';'      // 数据操作语言(INSERT, UPDATE, DELETE)
 | dqlStatement ';'      // 数据查询语言(SELECT)
 | utilityStatement ';'  // 实用语句(USE, SHOW等)
 ;

/*====================================================
 * 语句类型定义
 *====================================================*/

// DDL(数据定义语言)语句
ddlStatement
 : createDatabase
 | createTable
 | createIndex
 | dropTable
 | dropDatabase
 ;

// DML(数据操作语言)语句
dmlStatement
 : insertStatement
 | updateStatement
 | deleteStatement
 ;

// DQL(数据查询语言)语句
dqlStatement
 : selectStatement
 ;

// 实用语句
utilityStatement
 : useStatement
 | showDatabases
 | showTables
 | explainStatement
 ;

/*====================================================
 * DDL语句详细定义
 *====================================================*/

// 创建数据库
createDatabase
 : CREATE DATABASE identifier
 ;

// 创建表，支持列定义和表约束
createTable
 : CREATE TABLE tableName
   '(' columnDef ( ',' columnDef )* ( ',' tableConstraint )* ')'
   ( PARTITION BY partitionMethod )? // 分区支持（解析但不实现）
 ;

// 列定义，包含列名、数据类型和约束
columnDef
 : identifier dataType columnConstraint*
 ;

// 列级约束
columnConstraint
 : NOT? NULL         // 空值约束
 | PRIMARY KEY       // 主键约束
 | UNIQUE           // 唯一约束
 | DEFAULT literal   // 默认值
 ;

// 表级约束
tableConstraint
 : PRIMARY KEY '(' identifierList ')'  // 表级主键约束
 ;

// 创建索引
createIndex
 : CREATE INDEX identifier ON tableName '(' identifierList ')'
 ;

// 删除表
dropTable
 : DROP TABLE tableName
 ;

// 删除数据库
dropDatabase
 : DROP DATABASE identifier
 ;

/*====================================================
 * DML语句详细定义
 *====================================================*/

// INSERT语句，支持单行或多行插入
insertStatement
 : INSERT INTO tableName ( '(' identifierList ')' )? 
   VALUES '(' valueList ')' ( ',' '(' valueList ')' )*
 ;

// UPDATE语句，支持WHERE条件
updateStatement
 : UPDATE tableName SET updateAssignment ( ',' updateAssignment )* 
   ( WHERE expression )?
 ;

// DELETE语句，支持WHERE条件
deleteStatement
 : DELETE FROM tableName ( WHERE expression )?
 ;

/*====================================================
 * DQL语句详细定义
 *====================================================*/

// SELECT查询语句，支持完整的查询功能
selectStatement
 : SELECT selectItem ( ',' selectItem )*
   FROM tableReference
   ( WHERE expression )?
   ( GROUP BY groupByItem ( ',' groupByItem )* )?
   ( HAVING expression )?
   ( ORDER BY orderByItem ( ',' orderByItem )* )?
   ( LIMIT INTEGER )?
 ;

// 查询项定义
selectItem
 : ( tableName '.' )? '*'               #selectAll    // 选择所有列
 | expression ( AS? identifier )?       #selectExpr   // 带别名的表达式
 ;

// 表引用，支持基本表、子查询和连接
tableReference
 : tableName ( AS? identifier )?                            #tableRefBase     // 基本表引用
 | '(' selectStatement ')' AS? identifier                   #tableRefSubquery // 子查询
 | tableReference ( joinType JOIN tableReference ON expression ) #tableRefJoin    // 表连接
 ;

/*====================================================
 * 通用组件定义
 *====================================================*/

// 表名定义
tableName
 : identifier
 ;

// 标识符列表
identifierList
 : identifier (',' identifier)*
 ;

// 值列表
valueList
 : expression (',' expression)*
 ;

// 更新赋值
updateAssignment
 : identifier '=' expression
 ;

// GROUP BY项
groupByItem
 : expression
 ;

// ORDER BY项
orderByItem
 : expression (ASC | DESC)?
 ;

// 函数名
functionName
 : identifier
 ;

// 字面量
literal
 : STRING    // 字符串
 | INTEGER   // 整数
 | FLOAT     // 浮点数
 | NULL      // 空值
 ;

// 标识符
identifier
 : IDENTIFIER
 ;

// 分区方法
partitionMethod
 : HASH '(' identifierList ')'   // 哈希分区
 | RANGE '(' expression ')'      // 范围分区
 ;

// 连接类型
joinType
 : INNER?   // 内连接
 | LEFT     // 左连接
 ;

/*====================================================
 * 实用语句详细定义
 *====================================================*/

// USE语句
useStatement
 : USE identifier
 ;

// SHOW DATABASES语句
showDatabases
 : SHOW DATABASES
 ;

// SHOW TABLES语句
showTables
 : SHOW TABLES
 ;

// EXPLAIN语句
explainStatement
 : EXPLAIN ( ANALYZE )? ( VERBOSE )? sqlStatement
 ;

/*====================================================
 * 数据类型和表达式
 *====================================================*/

// 支持的数据类型
dataType
 : INT
 | BIGINT
 | VARCHAR '(' INTEGER ')'
 | DATE
 | DOUBLE
 ;

// 表达式定义
expression
 : literal                                            #literalExpr         // 字面量
 | identifier                                         #columnRefExpr       // 列引用
 | tableName '.' identifier                           #qualifiedColumnRef  // 限定列引用
 | '(' expression ')'                                 #nestedExpr          // 嵌套表达式
 | left=expression operator=('*'|'/') right=expression  #binaryArithExpr    // 乘除运算
 | left=expression operator=('+'|'-') right=expression  #binaryArithExpr    // 加减运算
 | left=expression operator=COMPARISON_OP right=expression #comparisonExpr    // 比较运算
 | expression AND expression                          #logicalExpr         // 逻辑AND
 | expression OR expression                           #logicalExpr         // 逻辑OR
 | functionName '(' ( expression ( ',' expression )* )? ')' #functionCall  // 函数调用
 ;

/*====================================================
 * 词法规则
 *====================================================*/

// 关键字定义
CREATE : C R E A T E;
DATABASE : D A T A B A S E;
TABLE : T A B L E;
INDEX : I N D E X;
DROP : D R O P;
INSERT : I N S E R T;
INTO : I N T O;
VALUES : V A L U E S;
SELECT : S E L E C T;
FROM : F R O M;
WHERE : W H E R E;
GROUP : G R O U P;
BY : B Y;
HAVING : H A V I N G;
ORDER : O R D E R;
LIMIT : L I M I T;
JOIN : J O I N;
INNER : I N N E R;
LEFT : L E F T;
ON : O N;
AS : A S;
USE : U S E;
SHOW : S H O W;
EXPLAIN : E X P L A I N;
PRIMARY : P R I M A R Y;
KEY : K E Y;
PARTITION : P A R T I T I O N;

// 数据类型关键字
INT : I N T;
BIGINT : B I G I N T;
VARCHAR : V A R C H A R;
DATE : D A T E;
DOUBLE : D O U B L E;

// 其他关键字
AND : A N D;
OR : O R;
ASC : A S C;
DESC : D E S C;
NULL : N U L L;
HASH : H A S H;
RANGE : R A N G E;
ANALYZE : A N A L Y Z E;
VERBOSE : V E R B O S E;
UPDATE : U P D A T E;
SET : S E T;
DELETE : D E L E T E;
DATABASES : D A T A B A S E S;
TABLES : T A B L E S;
NOT : N O T;
UNIQUE : U N I Q U E;
DEFAULT : D E F A U L T;

// 标识符规则：以字母或下划线开头，后跟字母、数字或下划线
IDENTIFIER
 : [a-zA-Z_][a-zA-Z_0-9]*
 ;

// 字符串字面量：支持单引号字符串，允许使用两个单引号表示字符串内的单引号
STRING
 : '\'' ( ~'\'' | '\'\'' )* '\''
 ;

// 整数字面量
INTEGER
 : [0-9]+
 ;

// 浮点数字面量
FLOAT
 : [0-9]+ '.' [0-9]* 
 | '.' [0-9]+
 ;

// 比较运算符
COMPARISON_OP
 : '=' | '>' | '<' | '>=' | '<=' | '<>' | '!='
 ;

// 忽略空白字符
WS : [ \t\r\n]+ -> skip;

/*====================================================
 * 字母片段定义：用于构建关键字
 *====================================================*/
fragment A : [aA];
fragment B : [bB];
fragment C : [cC];
fragment D : [dD];
fragment E : [eE];
fragment F : [fF];
fragment G : [gG];
fragment H : [hH];
fragment I : [iI];
fragment J : [jJ];
fragment K : [kK];
fragment L : [lL];
fragment M : [mM];
fragment N : [nN];
fragment O : [oO];
fragment P : [pP];
fragment Q : [qQ];
fragment R : [rR];
fragment S : [sS];
fragment T : [tT];
fragment U : [uU];
fragment V : [vV];
fragment W : [wW];
fragment X : [xX];
fragment Y : [yY];
fragment Z : [zZ];
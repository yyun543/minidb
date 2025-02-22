package storage

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/apache/arrow/go/v18/arrow"
)

// KeyManager 用于管理存储引擎的key
type KeyManager struct {
	sysPrefix  string
	userPrefix string
}

// NewKeyManager 创建KeyManager实例
func NewKeyManager() *KeyManager {
	return &KeyManager{
		sysPrefix:  "sys:",
		userPrefix: "user:",
	}
}

const (
	SYS_DATABASE = "system"

	// 系统表名
	SYS_DATABASES = "sys_databases"
	SYS_TABLES    = "sys_tables"
	SYS_COLUMNS   = "sys_columns"
	SYS_INDEXES   = "sys_indexes"

	// 每个chunk的目标大小(4KB)
	TARGET_CHUNK_SIZE = 4 * 1024
)

// 定义key类型常量
const (
	keyTypeDatabase = "db"     // 数据库元数据
	keyTypeTable    = "table"  // 表元数据
	keyTypeChunk    = "chunk"  // 表数据chunk
	keyTypeIndex    = "index"  // 索引数据
	keyTypeSchema   = "schema" // 表结构
)

// DatabaseKey 生成数据库元数据的key
func (km *KeyManager) DatabaseKey(dbName string) []byte {
	if dbName == SYS_DATABASE {
		return []byte(fmt.Sprintf("%s%s:%s", km.sysPrefix, keyTypeDatabase, dbName))
	}
	return []byte(fmt.Sprintf("%s%s:%s", km.userPrefix, keyTypeDatabase, dbName))
}

// TableSchemaKey 生成表Schema的key
func (km *KeyManager) TableSchemaKey(dbName, tableName string) []byte {
	if dbName == SYS_DATABASE {
		return []byte(fmt.Sprintf("%s%s:%s:%s", km.sysPrefix, keyTypeSchema, dbName, tableName))
	}
	return []byte(fmt.Sprintf("%s%s:%s:%s", km.userPrefix, keyTypeSchema, dbName, tableName))
}

// TableChunkKey 生成表数据chunk的key
func (km *KeyManager) TableChunkKey(dbName, tableName string, chunkId int64) []byte {
	if dbName == SYS_DATABASE {
		return []byte(fmt.Sprintf("%s%s:%s:%s:%d", km.sysPrefix, keyTypeChunk, dbName, tableName, chunkId))
	}
	return []byte(fmt.Sprintf("%s%s:%s:%s:%d", km.userPrefix, keyTypeChunk, dbName, tableName, chunkId))
}

// TableIndexKey 生成表索引记录的key
func (km *KeyManager) TableIndexKey(dbName, tableName, indexName string) []byte {
	if dbName == SYS_DATABASE {
		return []byte(fmt.Sprintf("%s%s:%s:%s:%s", km.sysPrefix, keyTypeIndex, dbName, tableName, indexName))
	}
	return []byte(fmt.Sprintf("%s%s:%s:%s:%s", km.userPrefix, keyTypeIndex, dbName, tableName, indexName))
}

// CalculateChunkSize 计算给定schema的每行大小，返回每个chunk可以容纳的最大行数
func (km *KeyManager) CalculateChunkSize(schema *arrow.Schema) int64 {
	rowSize := int64(0)
	for _, field := range schema.Fields() {
		switch field.Type.ID() {
		case arrow.BOOL:
			rowSize += 1
		case arrow.INT8, arrow.UINT8:
			rowSize += 1
		case arrow.INT16, arrow.UINT16:
			rowSize += 2
		case arrow.INT32, arrow.UINT32, arrow.FLOAT32:
			rowSize += 4
		case arrow.INT64, arrow.UINT64, arrow.FLOAT64:
			rowSize += 8
		case arrow.STRING, arrow.BINARY:
			// 对于变长类型，假设平均长度为32字节
			rowSize += 32
		default:
			// 其他类型默认按8字节计算
			rowSize += 8
		}
	}

	// 计算每个chunk可以容纳的最大行数
	maxRows := TARGET_CHUNK_SIZE / rowSize
	if maxRows < 1 {
		maxRows = 1 // 确保至少有1行
	}
	return maxRows
}

// 根据当前最大行数和chunk可以容纳的最大行数，计算出当前chunk的ID
func (km *KeyManager) CalculateChunkId(chunkSize int64, numRows int64) int64 {
	// 计算当前chunk的ID
	return (numRows + chunkSize - 1) / chunkSize
}

// ParseKey 解析key，返回其组成部分
func (km *KeyManager) ParseKey(key string) map[string]interface{} {
	result := make(map[string]interface{})
	parts := strings.Split(key, ":")

	// 检查前缀
	if len(parts) < 2 {
		return result
	}

	// 设置前缀
	result["prefix"] = parts[0]

	// 根据第二个部分（类型）进行解析
	switch parts[1] {
	case keyTypeDatabase:
		// sys:db:system 或 user:db:testdb
		if len(parts) >= 3 {
			result["type"] = "db"
			result["database"] = parts[2]
		}
	case keyTypeChunk:
		// sys:chunk:system:sys_tables:0
		if len(parts) >= 5 {
			result["type"] = "chunk"
			result["database"] = parts[2]
			result["table"] = parts[3]
			chunkId, err := strconv.ParseInt(parts[4], 10, 64)
			if err == nil {
				result["chunk_id"] = chunkId
			}
		}
	case keyTypeTable:
		// sys:table:system:sys_tables 或 user:table:testdb:testtable
		if len(parts) >= 4 {
			result["type"] = "table"
			result["database"] = parts[2]
			result["table"] = parts[3]
		}
	case keyTypeSchema:
		// sys:schema:system:sys_tables 或 user:schema:testdb:testtable
		if len(parts) >= 4 {
			result["type"] = "schema"
			result["database"] = parts[2]
			result["table"] = parts[3]
		}
	case keyTypeIndex:
		// sys:index:system:sys_tables:idx_name 或 user:index:testdb:testtable:idx_name
		if len(parts) >= 4 {
			result["type"] = "index"
			result["database"] = parts[2]
			result["table"] = parts[3]
			if len(parts) > 4 {
				result["index"] = parts[4]
			}
		}
	}

	return result
}

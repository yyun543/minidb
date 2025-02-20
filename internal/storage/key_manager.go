package storage

import (
	"fmt"
	"strings"
)

// KeyManager 用于管理存储引擎的key
type KeyManager struct {
	// 系统表前缀常量
	sysPrefix string
}

// NewKeyManager 创建KeyManager实例
func NewKeyManager() *KeyManager {
	return &KeyManager{
		sysPrefix: "sys:",
	}
}

const (
	SYS_DATABASE = "system"

	// 系统表名
	SYS_DATABASES = "sys_databases"
	SYS_TABLES    = "sys_tables"
	SYS_COLUMNS   = "sys_columns"
	SYS_INDEXES   = "sys_indexes"
)

// 定义key类型常量
const (
	keyTypeDatabase = "db"
	keyTypeTable    = "table"
	keyTypeColumn   = "column"
	keyTypeIndex    = "index"
)

// DatabaseKey 生成数据库元数据的key
func (km *KeyManager) DatabaseKey(dbName string) []byte {
	return []byte(fmt.Sprintf("%s:%s", keyTypeDatabase, dbName))
}

// TableKey 生成表元数据的key
func (km *KeyManager) TableKey(dbName, tableName string) []byte {
	if dbName == SYS_DATABASE {
		return []byte(fmt.Sprintf("%s%s:%s:%s", km.sysPrefix, keyTypeTable, dbName, tableName))
	}
	return []byte(fmt.Sprintf("%s:%s:%s", keyTypeTable, dbName, tableName))
}

// SysTableKey 生成系统表记录的key
func (km *KeyManager) SysTableKey(tableID int64) []byte {
	return []byte(fmt.Sprintf("%ssys_tables:%d", km.sysPrefix, tableID))
}

// GetKeyRange 返回以指定前缀的扫描范围：起始键为prefix，结束键为prefix的字典序后继。
func (km *KeyManager) GetKeyRange(prefix []byte) (start, end []byte) {
	start = prefix
	end = make([]byte, len(prefix))
	copy(end, prefix)
	end[len(end)-1]++ // 将最后一字节加1，假设不会溢出
	return
}

// ParseKey 解析key获取组成部分
func (km *KeyManager) ParseKey(key string) map[string]string {
	parts := strings.Split(key, ":")
	result := make(map[string]string)

	if len(parts) < 2 {
		return result
	}

	result["type"] = parts[0]

	switch parts[0] {
	case keyTypeDatabase:
		if len(parts) >= 2 {
			result["database"] = parts[1]
		}
	case keyTypeTable:
		if len(parts) >= 3 {
			result["database"] = parts[1]
			result["table"] = parts[2]
		}
	}

	return result
}

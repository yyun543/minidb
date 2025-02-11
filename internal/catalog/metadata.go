package catalog

import (
	"encoding/json"
	"time"
)

// TODO TableMeta, ColumnMeta 等元数据定义

// TableMeta 表元数据
type TableMeta struct {
	ID          int64        // 表ID
	Name        string       // 表名
	Columns     []ColumnMeta // 列定义
	Constraints []Constraint // 约束
	Indexes     []IndexMeta  // 索引
	CreateTime  time.Time    // 创建时间
	UpdateTime  time.Time    // 更新时间
}

// Column 列定义
type ColumnMeta struct {
	ID         int64     // 列ID
	Name       string    // 列名
	Type       string    // 数据类型
	NotNull    bool      // 非空约束
	Default    string    // 默认值
	Comment    string    // 列注释
	CreateTime time.Time // 创建时间
}

// Constraint 表约束
type Constraint struct {
	ID         int64    // 约束ID
	Name       string   // 约束名
	Type       string   // 约束类型(PRIMARY/UNIQUE/FOREIGN/CHECK)
	Columns    []string // 相关列
	RefTable   string   // 外键引用表
	RefColumns []string // 外键引用列
}

// IndexMeta 索引元数据
type IndexMeta struct {
	ID         int64     // 索引ID
	Name       string    // 索引名
	Columns    []string  // 索引列
	Type       string    // 索引类型(BTREE/HASH)
	Unique     bool      // 是否唯一索引
	CreateTime time.Time // 创建时间
}

// DatabaseMeta 数据库元数据
type DatabaseMeta struct {
	ID         int64       // 数据库ID
	Name       string      // 数据库名
	Tables     []TableMeta // 表列表
	CreateTime time.Time   // 创建时间
	UpdateTime time.Time   // 更新时间
}

// Serialize 序列化元数据
func (m *TableMeta) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

// Deserialize 反序列化元数据
func (m *TableMeta) Deserialize(data []byte) error {
	return json.Unmarshal(data, m)
}

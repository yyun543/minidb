package catalog

import "time"

// DatabaseMeta 数据库元数据
type DatabaseMeta struct {
	ID         int64     // 数据库ID
	Name       string    // 数据库名称
	CreateTime time.Time // 创建时间
	UpdateTime time.Time // 更新时间
}

// TableMeta 表元数据
type TableMeta struct {
	ID          int64        // 表ID
	Name        string       // 表名
	DatabaseID  int64        // 所属数据库ID
	Columns     []ColumnMeta // 列定义
	Indexes     []IndexMeta  // 索引定义
	Constraints []Constraint // 约束定义
	CreateTime  time.Time    // 创建时间
	UpdateTime  time.Time    // 更新时间
}

// ColumnMeta 列元数据
type ColumnMeta struct {
	ID         int64     // 列ID
	Name       string    // 列名
	Type       string    // 数据类型
	Length     int       // 长度(可选)
	NotNull    bool      // 是否允许为空
	Default    string    // 默认值
	CreateTime time.Time // 创建时间
	UpdateTime time.Time // 更新时间
}

// IndexMeta 索引元数据
type IndexMeta struct {
	ID         int64     // 索引ID
	Name       string    // 索引名称
	Type       string    // 索引类型(BTREE/HASH)
	Columns    []string  // 索引列
	IsUnique   bool      // 是否唯一索引
	IsPrimary  bool      // 是否主键索引
	CreateTime time.Time // 创建时间
	UpdateTime time.Time // 更新时间
}

// Constraint 约束定义
type Constraint struct {
	Name       string   // 约束名称
	Type       string   // 约束类型(PRIMARY/UNIQUE/FOREIGN)
	Columns    []string // 涉及的列
	RefTable   string   // 引用表(仅FOREIGN约束)
	RefColumns []string // 引用列(仅FOREIGN约束)
}

package catalog

import (
	"fmt"
	"sync"
)

// SystemTables 系统表管理器
type SystemTables struct {
	mu sync.RWMutex

	// 系统表定义
	tables map[string]*TableMeta
}

// 系统表名常量
const (
	SYS_DATABASES = "sys_databases" // 数据库信息表
	SYS_TABLES    = "sys_tables"    // 表信息表
	SYS_COLUMNS   = "sys_columns"   // 列信息表
	SYS_INDEXES   = "sys_indexes"   // 索引信息表
)

// NewSystemTables 创建系统表管理器
func NewSystemTables() *SystemTables {
	st := &SystemTables{
		tables: make(map[string]*TableMeta),
	}
	st.initSystemTables()
	return st
}

// initSystemTables 初始化系统表
func (st *SystemTables) initSystemTables() {
	// 数据库信息表
	st.tables[SYS_DATABASES] = &TableMeta{
		Name: SYS_DATABASES,
		Columns: []ColumnMeta{
			{Name: "id", Type: "INT64", NotNull: true},
			{Name: "name", Type: "STRING", NotNull: true},
			{Name: "create_time", Type: "TIMESTAMP", NotNull: true},
			{Name: "update_time", Type: "TIMESTAMP", NotNull: true},
		},
		Constraints: []Constraint{
			{Name: "pk_database", Type: "PRIMARY", Columns: []string{"id"}},
			{Name: "uk_database_name", Type: "UNIQUE", Columns: []string{"name"}},
		},
	}

	// 表信息表
	st.tables[SYS_TABLES] = &TableMeta{
		Name: SYS_TABLES,
		Columns: []ColumnMeta{
			{Name: "id", Type: "INT64", NotNull: true},
			{Name: "database_id", Type: "INT64", NotNull: true},
			{Name: "name", Type: "STRING", NotNull: true},
			{Name: "create_time", Type: "TIMESTAMP", NotNull: true},
			{Name: "update_time", Type: "TIMESTAMP", NotNull: true},
		},
		Constraints: []Constraint{
			{Name: "pk_table", Type: "PRIMARY", Columns: []string{"id"}},
			{Name: "uk_table_name", Type: "UNIQUE", Columns: []string{"database_id", "name"}},
		},
	}

	// 列信息表
	st.tables[SYS_COLUMNS] = &TableMeta{
		Name: SYS_COLUMNS,
		Columns: []ColumnMeta{
			{Name: "id", Type: "INT64", NotNull: true},
			{Name: "table_id", Type: "INT64", NotNull: true},
			{Name: "name", Type: "STRING", NotNull: true},
			{Name: "type", Type: "STRING", NotNull: true},
			{Name: "not_null", Type: "BOOL", NotNull: true},
			{Name: "default_value", Type: "STRING"},
			{Name: "comment", Type: "STRING"},
			{Name: "create_time", Type: "TIMESTAMP", NotNull: true},
		},
		Constraints: []Constraint{
			{Name: "pk_column", Type: "PRIMARY", Columns: []string{"id"}},
			{Name: "uk_column_name", Type: "UNIQUE", Columns: []string{"table_id", "name"}},
		},
	}
}

// GetTable 获取系统表定义
func (st *SystemTables) GetTable(name string) (*TableMeta, error) {
	st.mu.RLock()
	defer st.mu.RUnlock()

	table, ok := st.tables[name]
	if !ok {
		return nil, fmt.Errorf("系统表不存在: %s", name)
	}
	return table, nil
}

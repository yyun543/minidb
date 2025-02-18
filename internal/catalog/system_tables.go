package catalog

import (
	"github.com/yyun543/minidb/internal/storage"
	"time"
)

// SystemTables 系统表管理器

// initSystemTables 初始化系统表
func (c *Catalog) initSystemTables() error {
	// 创建系统数据库
	sysDB := &DatabaseMeta{
		ID:         time.Now().UnixNano(),
		Name:       storage.SYS_DATABASE,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	if err := c.storage.Put(c.keyManager.DatabaseKey(sysDB.Name), encodeDatabase(sysDB)); err != nil {
		return err
	}

	// 创建系统表: sys_databases
	if err := c.createSysDatabasesTable(sysDB.ID); err != nil {
		return err
	}

	// 创建系统表: sys_tables
	if err := c.createSysTablesTable(); err != nil {
		return err
	}

	// 创建系统表: sys_columns
	if err := c.createSysColumnsTable(); err != nil {
		return err
	}

	// 创建系统表: sys_indexes
	if err := c.createSysIndexesTable(); err != nil {
		return err
	}

	return nil
}

func (c *Catalog) createSysDatabasesTable(sysDBID int64) error {
	table := &TableMeta{
		ID:         time.Now().UnixNano(),
		Name:       storage.SYS_DATABASES,
		DatabaseID: sysDBID,
		Columns: []ColumnMeta{
			{
				ID:         time.Now().UnixNano(),
				Name:       "id",
				Type:       "INTEGER",
				NotNull:    true,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			},
			{
				ID:         time.Now().UnixNano(),
				Name:       "name",
				Type:       "VARCHAR",
				NotNull:    true,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			},
			{
				ID:         time.Now().UnixNano(),
				Name:       "create_time",
				Type:       "TIMESTAMP",
				NotNull:    true,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			},
		},
		Constraints: []Constraint{
			{
				Name:    "pk_sys_databases",
				Type:    "PRIMARY",
				Columns: []string{"id"},
			},
		},
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	return c.storage.Put(c.keyManager.TableKey(storage.SYS_DATABASE, table.Name), encodeTable(table))
}

// createSysTablesTable 创建系统表 sys_tables
func (c *Catalog) createSysTablesTable() error {
	// 创建系统表 sys_tables 的元数据
	sysTablesTable := &TableMeta{
		ID:   time.Now().UnixNano(),
		Name: "sys_tables",
		Columns: []ColumnMeta{
			{
				ID:      1,
				Name:    "table_id",
				Type:    "INTEGER",
				NotNull: true,
			},
			{
				ID:      2,
				Name:    "database_name",
				Type:    "VARCHAR",
				NotNull: true,
			},
			{
				ID:      3,
				Name:    "table_name",
				Type:    "VARCHAR",
				NotNull: true,
			},
			{
				ID:      4,
				Name:    "table_type",
				Type:    "VARCHAR",
				NotNull: true,
			},
			{
				ID:      5,
				Name:    "create_time",
				Type:    "TIMESTAMP",
				NotNull: true,
			},
			{
				ID:      6,
				Name:    "update_time",
				Type:    "TIMESTAMP",
				NotNull: true,
			},
		},
		Constraints: []Constraint{
			{
				Name:    "pk_sys_tables",
				Type:    "PRIMARY",
				Columns: []string{"table_id"},
			},
			{
				Name:    "uk_sys_tables_name",
				Type:    "UNIQUE",
				Columns: []string{"database_name", "table_name"},
			},
		},
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	// 存储系统表元数据
	return c.storage.Put(c.keyManager.TableKey(storage.SYS_DATABASE, sysTablesTable.Name), encodeTable(sysTablesTable))
}

// createSysColumnsTable 创建系统表 sys_columns
func (c *Catalog) createSysColumnsTable() error {
	// 创建系统表 sys_columns 的元数据
	sysColumnsTable := &TableMeta{
		ID:   time.Now().UnixNano(),
		Name: "sys_columns",
		Columns: []ColumnMeta{
			{
				ID:      1,
				Name:    "column_id",
				Type:    "INTEGER",
				NotNull: true,
			},
			{
				ID:      2,
				Name:    "table_id",
				Type:    "INTEGER",
				NotNull: true,
			},
			{
				ID:      3,
				Name:    "column_name",
				Type:    "VARCHAR",
				NotNull: true,
			},
			{
				ID:      4,
				Name:    "data_type",
				Type:    "VARCHAR",
				NotNull: true,
			},
			{
				ID:      5,
				Name:    "is_nullable",
				Type:    "BOOLEAN",
				NotNull: true,
			},
			{
				ID:      6,
				Name:    "default_value",
				Type:    "VARCHAR",
				NotNull: false,
			},
			{
				ID:      7,
				Name:    "ordinal_position",
				Type:    "INTEGER",
				NotNull: true,
			},
			{
				ID:      8,
				Name:    "create_time",
				Type:    "TIMESTAMP",
				NotNull: true,
			},
			{
				ID:      9,
				Name:    "update_time",
				Type:    "TIMESTAMP",
				NotNull: true,
			},
		},
		Constraints: []Constraint{
			{
				Name:    "pk_sys_columns",
				Type:    "PRIMARY",
				Columns: []string{"column_id"},
			},
			{
				Name:       "fk_sys_columns_table",
				Type:       "FOREIGN",
				Columns:    []string{"table_id"},
				RefTable:   "sys_tables",
				RefColumns: []string{"table_id"},
			},
			{
				Name:    "uk_sys_columns_name",
				Type:    "UNIQUE",
				Columns: []string{"table_id", "column_name"},
			},
		},
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	// 存储系统表元数据
	return c.storage.Put(c.keyManager.TableKey(storage.SYS_DATABASE, sysColumnsTable.Name), encodeTable(sysColumnsTable))
}

func (c *Catalog) createSysIndexesTable() error {
	sysIndexesTable := &TableMeta{
		ID:   time.Now().UnixNano(),
		Name: "sys_indexes",
		Columns: []ColumnMeta{
			{
				ID:      1,
				Name:    "index_id",
				Type:    "INTEGER",
				NotNull: true,
			},
			{
				ID:      2,
				Name:    "table_id",
				Type:    "INTEGER",
				NotNull: true,
			},
			{
				ID:      3,
				Name:    "index_name",
				Type:    "VARCHAR",
				NotNull: true,
			},
			{
				ID:      4,
				Name:    "index_type",
				Type:    "VARCHAR",
				NotNull: true,
			},
			{
				ID:      5,
				Name:    "is_unique",
				Type:    "BOOLEAN",
				NotNull: true,
			},
			{
				ID:      6,
				Name:    "is_primary",
				Type:    "BOOLEAN",
				NotNull: true,
			},
			{
				ID:      7,
				Name:    "column_names",
				Type:    "VARCHAR",
				NotNull: true,
			},
			{
				ID:      8,
				Name:    "create_time",
				Type:    "TIMESTAMP",
				NotNull: true,
			},
			{
				ID:      9,
				Name:    "update_time",
				Type:    "TIMESTAMP",
				NotNull: true,
			},
		},
		Constraints: []Constraint{
			{
				Name:    "pk_sys_indexes",
				Type:    "PRIMARY",
				Columns: []string{"index_id"},
			},
			{
				Name:       "fk_sys_indexes_table",
				Type:       "FOREIGN",
				Columns:    []string{"table_id"},
				RefTable:   "sys_tables",
				RefColumns: []string{"table_id"},
			},
			{
				Name:    "uk_sys_indexes_name",
				Type:    "UNIQUE",
				Columns: []string{"table_id", "index_name"},
			},
		},
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	return c.storage.Put(c.keyManager.TableKey(storage.SYS_DATABASE, sysIndexesTable.Name), encodeTable(sysIndexesTable))
}

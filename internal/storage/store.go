package storage

type Store interface {
	CreateTable(name string, schema Schema) error
	DropTable(name string) error
	Insert(tableName string, values map[string]interface{}) error
	Select(tableName string, columns []string, where string) ([]Row, error)
	Update(tableName string, values map[string]interface{}, where string) (int, error)
	Delete(tableName string, where string) (int, error)
}

// 将通用的验证函数移到这里
// DataType 表示支持的数据类型
type DataType string

const (
	TypeString   DataType = "string"
	TypeInt      DataType = "int"
	TypeFloat    DataType = "float"
	TypeBool     DataType = "bool"
	TypeDateTime DataType = "datetime"
)

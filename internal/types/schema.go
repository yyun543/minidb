package types

import (
	"fmt"

	"github.com/apache/arrow/go/v18/arrow"
)

// DataType 定义MiniDB支持的数据类型，确保类型安全和确定性
type DataType uint8

const (
	UnknownType DataType = iota
	BooleanType
	Int8Type
	Int16Type
	Int32Type
	Int64Type
	Float32Type
	Float64Type
	StringType
	BinaryType
	DateType
	TimestampType
	DecimalType
)

// String 返回数据类型的字符串表示
func (dt DataType) String() string {
	switch dt {
	case BooleanType:
		return "BOOLEAN"
	case Int8Type:
		return "TINYINT"
	case Int16Type:
		return "SMALLINT"
	case Int32Type:
		return "INTEGER"
	case Int64Type:
		return "BIGINT"
	case Float32Type:
		return "REAL"
	case Float64Type:
		return "DOUBLE"
	case StringType:
		return "VARCHAR"
	case BinaryType:
		return "BINARY"
	case DateType:
		return "DATE"
	case TimestampType:
		return "TIMESTAMP"
	case DecimalType:
		return "DECIMAL"
	default:
		return "UNKNOWN"
	}
}

// ToArrowType 转换为Arrow数据类型
func (dt DataType) ToArrowType() arrow.DataType {
	switch dt {
	case BooleanType:
		return arrow.FixedWidthTypes.Boolean
	case Int8Type:
		return arrow.PrimitiveTypes.Int8
	case Int16Type:
		return arrow.PrimitiveTypes.Int16
	case Int32Type:
		return arrow.PrimitiveTypes.Int32
	case Int64Type:
		return arrow.PrimitiveTypes.Int64
	case Float32Type:
		return arrow.PrimitiveTypes.Float32
	case Float64Type:
		return arrow.PrimitiveTypes.Float64
	case StringType:
		return arrow.BinaryTypes.String
	case BinaryType:
		return arrow.BinaryTypes.Binary
	case DateType:
		return arrow.FixedWidthTypes.Date32
	case TimestampType:
		return arrow.FixedWidthTypes.Timestamp_ns
	default:
		return arrow.BinaryTypes.String // 默认为字符串
	}
}

// FromArrowType 从Arrow数据类型转换
func FromArrowType(arrowType arrow.DataType) DataType {
	switch arrowType.ID() {
	case arrow.BOOL:
		return BooleanType
	case arrow.INT8:
		return Int8Type
	case arrow.INT16:
		return Int16Type
	case arrow.INT32:
		return Int32Type
	case arrow.INT64:
		return Int64Type
	case arrow.FLOAT32:
		return Float32Type
	case arrow.FLOAT64:
		return Float64Type
	case arrow.STRING:
		return StringType
	case arrow.BINARY:
		return BinaryType
	case arrow.DATE32:
		return DateType
	case arrow.TIMESTAMP:
		return TimestampType
	default:
		return StringType // 默认为字符串
	}
}

// ColumnSchema 列模式定义，提供强类型约束
type ColumnSchema struct {
	Name     string   // 列名
	Type     DataType // 数据类型
	Nullable bool     // 是否允许NULL
	Default  any      // 默认值
}

// Validate 验证列模式的有效性
func (cs *ColumnSchema) Validate() error {
	if cs.Name == "" {
		return fmt.Errorf("column name cannot be empty")
	}
	if cs.Type == UnknownType {
		return fmt.Errorf("column type cannot be unknown")
	}
	return nil
}

// ToArrowField 转换为Arrow字段
func (cs *ColumnSchema) ToArrowField() arrow.Field {
	return arrow.Field{
		Name:     cs.Name,
		Type:     cs.Type.ToArrowType(),
		Nullable: cs.Nullable,
	}
}

// TableSchema 表模式定义，确保强类型和确定性
type TableSchema struct {
	Name    string          // 表名
	Columns []*ColumnSchema // 列定义
	Primary []string        // 主键列名（为分布式做准备）
}

// Validate 验证表模式
func (ts *TableSchema) Validate() error {
	if ts.Name == "" {
		return fmt.Errorf("table name cannot be empty")
	}
	if len(ts.Columns) == 0 {
		return fmt.Errorf("table must have at least one column")
	}

	// 检查列名唯一性
	nameSet := make(map[string]bool)
	for _, col := range ts.Columns {
		if err := col.Validate(); err != nil {
			return fmt.Errorf("invalid column %s: %w", col.Name, err)
		}
		if nameSet[col.Name] {
			return fmt.Errorf("duplicate column name: %s", col.Name)
		}
		nameSet[col.Name] = true
	}

	return nil
}

// ToArrowSchema 转换为Arrow模式
func (ts *TableSchema) ToArrowSchema() *arrow.Schema {
	fields := make([]arrow.Field, len(ts.Columns))
	for i, col := range ts.Columns {
		fields[i] = col.ToArrowField()
	}
	return arrow.NewSchema(fields, nil)
}

// FindColumn 查找列定义
func (ts *TableSchema) FindColumn(name string) (*ColumnSchema, int, error) {
	for i, col := range ts.Columns {
		if col.Name == name {
			return col, i, nil
		}
	}
	return nil, -1, fmt.Errorf("column %s not found", name)
}

// GetColumnNames 获取所有列名
func (ts *TableSchema) GetColumnNames() []string {
	names := make([]string, len(ts.Columns))
	for i, col := range ts.Columns {
		names[i] = col.Name
	}
	return names
}

// GetPrimaryKeyIndices 获取主键列的索引（为分片键做准备）
func (ts *TableSchema) GetPrimaryKeyIndices() []int {
	if len(ts.Primary) == 0 {
		return nil
	}

	indices := make([]int, 0, len(ts.Primary))
	for _, pkCol := range ts.Primary {
		if _, idx, err := ts.FindColumn(pkCol); err == nil {
			indices = append(indices, idx)
		}
	}
	return indices
}

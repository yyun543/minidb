package optimizer

import (
	"fmt"
)

// PlanType 定义了查询计划节点的类型
type PlanType int

const (
	UnknownPlan PlanType = iota
	SelectPlan
	TableScanPlan
	FilterPlan
	JoinPlan
	InsertPlan
	UpdatePlan
	DeletePlan
	OrderPlan
	LimitPlan
	GroupPlan
)

// String 返回 PlanType 的字符串描述，便于调试和日志记录
func (pt PlanType) String() string {
	switch pt {
	case SelectPlan:
		return "Select"
	case TableScanPlan:
		return "TableScan"
	case FilterPlan:
		return "Filter"
	case JoinPlan:
		return "Join"
	case InsertPlan:
		return "Insert"
	case UpdatePlan:
		return "Update"
	case DeletePlan:
		return "Delete"
	case OrderPlan:
		return "OrderBy"
	case LimitPlan:
		return "Limit"
	case GroupPlan:
		return "GroupBy"
	default:
		return "Unknown"
	}
}

// PlanProperties 定义了计划节点属性需要实现的接口，用于解释和调试
type PlanProperties interface {
	Explain() string
}

// Plan 定义了查询计划树的基础结构
type Plan struct {
	Type       PlanType       // 节点类型
	Properties PlanProperties // 节点专有属性
	Children   []*Plan        // 子节点列表
}

// NewPlan 创建一个新的 Plan 节点
func NewPlan(pt PlanType) *Plan {
	return &Plan{
		Type:     pt,
		Children: []*Plan{},
	}
}

// AddChild 添加子节点到当前 Plan 节点中
func (p *Plan) AddChild(child *Plan) {
	p.Children = append(p.Children, child)
}

// Explain 递归输出整个计划树的结构，用于调试或日志追踪
func (p *Plan) Explain(indent string) string {
	explanation := fmt.Sprintf("%s%s", indent, p.Type.String())
	if p.Properties != nil {
		explanation += " {" + p.Properties.Explain() + "}"
	}
	explanation += "\n"
	for _, child := range p.Children {
		explanation += child.Explain(indent + "  ")
	}
	return explanation
}

// -----------------------------------------------------------------------------
// 以下为各个计划节点对应的属性定义，均实现 PlanProperties 接口
// -----------------------------------------------------------------------------

// ColumnRefType 定义列引用类型
type ColumnRefType int

const (
	ColumnRefTypeColumn ColumnRefType = iota
	ColumnRefTypeFunction
	ColumnRefTypeExpression
)

// ColumnRef 定义列引用
type ColumnRef struct {
	Column       string
	Table        string
	Alias        string
	Type         ColumnRefType
	FunctionName string
	FunctionArgs []Expression
	Expression   Expression
}

// SelectProperties 用于SELECT计划
type SelectProperties struct {
	All     bool        // 是否为SELECT *
	Columns []ColumnRef // 选择的列
}

func (sp *SelectProperties) Explain() string {
	return fmt.Sprintf("Columns: %v", sp.Columns)
}

// TableScanProperties 用于表扫描计划
type TableScanProperties struct {
	Table      string      // 表名
	TableAlias string      // 表别名
	Columns    []ColumnRef // 需要扫描的列
}

func (tp *TableScanProperties) Explain() string {
	return fmt.Sprintf("Table: %s", tp.Table)
}

// FilterProperties 用于过滤（WHERE）条件计划
type FilterProperties struct {
	Condition Expression // 条件表达式
}

func (fp *FilterProperties) Explain() string {
	return fmt.Sprintf("Condition: %v", fp.Condition)
}

// OrderKey 定义排序键
type OrderKey struct {
	Column    string // 排序列
	Direction string // 排序方向 (ASC/DESC)
}

// OrderByProperties 用于 ORDER BY 排序计划
type OrderByProperties struct {
	OrderKeys []OrderKey // 排序关键字列表
}

func (op *OrderByProperties) Explain() string {
	return fmt.Sprintf("OrderKeys: %v", op.OrderKeys)
}

// LimitProperties 用于 LIMIT 限制计划
type LimitProperties struct {
	Limit int64
}

func (lp *LimitProperties) Explain() string {
	return fmt.Sprintf("Limit: %d", lp.Limit)
}

// InsertProperties 用于 INSERT 计划
type InsertProperties struct {
	Table   string       // 表名
	Columns []string     // 列名列表
	Values  []Expression // 插入数据
}

func (ip *InsertProperties) Explain() string {
	return fmt.Sprintf("Table: %s, Columns: %v, Values: %v", ip.Table, ip.Columns, ip.Values)
}

// UpdateProperties 用于 UPDATE 计划
type UpdateProperties struct {
	Table       string                 // 表名
	Assignments map[string]interface{} // 列更新赋值
	Where       interface{}            // WHERE 条件表达式
}

func (up *UpdateProperties) Explain() string {
	return fmt.Sprintf("Table: %s, Assignments: %v, Where: %v", up.Table, up.Assignments, up.Where)
}

// DeleteProperties 用于 DELETE 计划
type DeleteProperties struct {
	Table string      // 表名
	Where interface{} // WHERE 条件表达式
}

func (dp *DeleteProperties) Explain() string {
	return fmt.Sprintf("Table: %s, Where: %v", dp.Table, dp.Where)
}

// JoinProperties 用于JOIN计划
type JoinProperties struct {
	JoinType   string     // JOIN类型(INNER/LEFT/RIGHT)
	Left       string     // 左表名
	LeftAlias  string     // 左表别名
	Right      string     // 右表名
	RightAlias string     // 右表别名
	Condition  Expression // JOIN条件
}

func (jp *JoinProperties) Explain() string {
	return fmt.Sprintf("Type: %s, Left: %s, Right: %s", jp.JoinType, jp.Left, jp.Right)
}

// GroupByProperties 用于 GROUP BY 计划
type GroupByProperties struct {
	GroupKeys []string    // 分组键列表
	Having    interface{} // HAVING 条件
}

func (gp *GroupByProperties) Explain() string {
	return fmt.Sprintf("GroupKeys: %v, Having: %v", gp.GroupKeys, gp.Having)
}

// Expression 表达式接口
type Expression interface {
	String() string
}

// BinaryExpression 二元表达式
type BinaryExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (e *BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", e.Left, e.Operator, e.Right)
}

// ColumnReference 列引用表达式
type ColumnReference struct {
	Column string
	Table  string
}

func (e *ColumnReference) String() string {
	if e.Table != "" {
		return fmt.Sprintf("%s.%s", e.Table, e.Column)
	}
	return e.Column
}

// LiteralValue 字面量
type LiteralValue struct {
	Type  LiteralType
	Value interface{}
}

func (e *LiteralValue) String() string {
	return fmt.Sprintf("%v", e.Value)
}

type LiteralType int

const (
	LiteralTypeInteger LiteralType = iota
	LiteralTypeFloat
	LiteralTypeString
	LiteralTypeBoolean
)

package optimizer

import (
	"github.com/yyun543/minidb/internal/parser"
)

// PlanType 计划节点类型
type PlanType int

const (
	SelectPlan PlanType = iota
	TableScanPlan
	JoinPlan
	FilterPlan
	InsertPlan
	UpdatePlan
	DeletePlan
)

// LogicalPlan 逻辑计划节点
type LogicalPlan struct {
	Type       PlanType       // 节点类型
	Properties PlanProperties // 节点属性
	Children   []*LogicalPlan // 子节点
	Cost       float64        // 估算代价
}

// PlanProperties 计划节点属性接口
type PlanProperties interface {
	Type() PlanType
}

// SelectProperties SELECT节点属性
type SelectProperties struct {
	Columns []string
}

func (p *SelectProperties) Type() PlanType {
	return SelectPlan
}

// TableScanProperties 表扫描节点属性
type TableScanProperties struct {
	Table string
}

func (p *TableScanProperties) Type() PlanType {
	return TableScanPlan
}

// JoinProperties JOIN节点属性
type JoinProperties struct {
	JoinType  string
	Table     string
	Condition parser.Node
}

func (p *JoinProperties) Type() PlanType {
	return JoinPlan
}

// FilterProperties 过滤节点属性
type FilterProperties struct {
	Condition parser.Node
}

func (p *FilterProperties) Type() PlanType {
	return FilterPlan
}

// InsertProperties INSERT节点属性
type InsertProperties struct {
	Table   string
	Columns []string
	Values  []parser.Node
}

func (p *InsertProperties) Type() PlanType {
	return InsertPlan
}

// UpdateProperties UPDATE节点属性
type UpdateProperties struct {
	Table       string
	Assignments []*parser.UpdateAssignment
	Where       *parser.WhereClause
}

func (p *UpdateProperties) Type() PlanType {
	return UpdatePlan
}

// DeleteProperties DELETE节点属性
type DeleteProperties struct {
	Table string
	Where *parser.WhereClause
}

func (p *DeleteProperties) Type() PlanType {
	return DeletePlan
}

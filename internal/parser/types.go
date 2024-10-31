package parser

import (
	"fmt"
)

// ExpressionType 表示表达式的类型
type ExpressionType interface {
	Expression
	Type() string
	Value() interface{}
}

// ValueExpression 实现了 ExpressionType 接口
type ValueExpression struct {
	value interface{}
	typ   string
	val   interface{}
}

func NewValueExpression(value interface{}, typ string) *ValueExpression {
	return &ValueExpression{
		value: value,
		typ:   typ,
	}
}

func (v *ValueExpression) expressionNode()    {}
func (v *ValueExpression) String() string     { return fmt.Sprintf("%v", v.value) }
func (v *ValueExpression) Type() string       { return v.typ }
func (v *ValueExpression) Value() interface{} { return v.value }

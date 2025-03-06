package inverted

import "fmt"

type FilterOp int

const (
	OpEqual FilterOp = iota
	OpNotEqual
	OpGreaterThan
	OpGreaterThanEqual
	OpLessThan
	OpLessThanEqual
)

type LogicalOp int

const (
	LogicalAnd LogicalOp = iota
	LogicalOr
)

type Filter struct {
	IndexName string
	Op        FilterOp
	Value     interface{}
}

func NewFilter(indexName string, op FilterOp, value interface{}) *Filter {
	return &Filter{
		IndexName: indexName,
		Op:        op,
		Value:     value,
	}
}

func (f *Filter) String() string {
	return fmt.Sprintf("Filter[IndexName=%s, Op=%d, Value=%v]", f.IndexName, f.Op, f.Value)
}

type FilterExpression struct {
	Single    *Filter
	Composite *CompositeFilter
}

func NewSingleExpression(f *Filter) *FilterExpression {
	return &FilterExpression{Single: f}
}

func NewCompositeExpression(op LogicalOp, exprs ...*FilterExpression) *FilterExpression {
	return &FilterExpression{
		Composite: &CompositeFilter{
			Op:          op,
			Expressions: exprs,
		},
	}
}

func (fe *FilterExpression) String() string {
	if fe.Single != nil {
		return fmt.Sprintf("Single[%v %v]", fe.Single.IndexName, fe.Single.Value)
	} else if fe.Composite != nil {
		return fmt.Sprintf("Composite[%v]", fe.Composite)
	}
	return "Empty"
}

type CompositeFilter struct {
	Op          LogicalOp
	Expressions []*FilterExpression
}

func (cf *CompositeFilter) String() string {
	opStr := "AND"
	if cf.Op == LogicalOr {
		opStr = "OR"
	}
	return fmt.Sprintf("%s(%v)", opStr, cf.Expressions)
}

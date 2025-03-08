// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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

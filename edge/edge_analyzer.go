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

package edge

import (
	"errors"
	"fmt"

	"github.com/sjy-dv/coltt/gen/protoc/v4/edgepb"
	"github.com/sjy-dv/coltt/pkg/inverted"
)

func standardAnalyzer(metadata map[string]interface{}, analyzer map[string]IndexFeature) error {
	for _, column := range analyzer {

		value, ok := metadata[column.IndexName]
		if !ok {
			if column.EnableNull {
				if column.PrimaryKey {
					return fmt.Errorf("primaryKey %s must not be empty", column.IndexName)
				}
				baseValue := defaultType(column.IndexType)
				if baseValue == nil {
					return fmt.Errorf("index: %s design error, type: %d", column.IndexName, column.IndexType)
				}
				value = baseValue
			} else {
				return fmt.Errorf("index: %s is null, but index design not allowed null value", column.IndexName)
			}
		}
		if column.PrimaryKey {
			_, ok := value.(string)
			if !ok {
				return fmt.Errorf("primaryKey [%s] must be string", column.IndexName)
			}
			continue
		}
		switch column.IndexType {
		case 0:
			_, ok := value.(string)
			if !ok {
				return fmt.Errorf("index: [%s] type error, expect Type: %s", column.IndexName, edgepb.IndexType_name[column.IndexType])
			}
		case 1:
			switch v := value.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			case float64:
				if v != float64(int64(v)) {
					return fmt.Errorf("index: [%s] type error, expect Type: %s", column.IndexName, edgepb.IndexType_name[column.IndexType])
				}
				//prevent map forced convert int => float
				metadata[column.IndexName] = int64(v)
			default:
				return fmt.Errorf("index: [%s] type error, expect Type: %s", column.IndexName, edgepb.IndexType_name[column.IndexType])
			}
		case 2:
			switch value.(type) {
			case float32, float64:
			default:
				return fmt.Errorf("index: [%s] type error, expect Type: %s", column.IndexName, edgepb.IndexType_name[column.IndexType])
			}
		case 3:
			_, ok := value.(bool)
			if !ok {
				return fmt.Errorf("index: [%s] type error, expect Type: %s", column.IndexName, edgepb.IndexType_name[column.IndexType])
			}
		}
	}
	return nil
}
func defaultType(typeLevel int32) interface{} {
	switch typeLevel {
	case 0:
		return ""
	case 1:
		return 0
	case 2:
		return float64(0)
	case 3:
		return false
	default:
		return nil
	}
}

/*
	실제 proto type과 매칭해봐야함

if edgepb.IndexType_String == edgepb.IndexType(value.IndexType) {

	}

이런식으로 일단은 향후 개발
현재로서 rule만 어긋나지 않으면 크게 문제없음
*/
func dropKeyAnalyzer(dropKey map[string]interface{}, analyzer map[string]IndexFeature) error {
	for indexName, indexValue := range dropKey {

		value, ok := analyzer[indexName]
		if !ok {
			return errors.New("ErrNotDefinedIndex")
		}
		switch indexValue {
		case 0:
			_, ok := indexValue.(string)
			if !ok {
				return fmt.Errorf("index: [%s] type error, expect Type: %s", indexName, edgepb.IndexType_name[value.IndexType])
			}
		case 1:
			switch v := indexValue.(type) {
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			case float64:
				if v != float64(int64(v)) {
					return fmt.Errorf("index: [%s] type error, expect Type: %s", indexName, edgepb.IndexType_name[value.IndexType])
				}
			default:
				return fmt.Errorf("index: [%s] type error, expect Type: %s", indexName, edgepb.IndexType_name[value.IndexType])
			}
		case 2:
			switch indexValue.(type) {
			case float32, float64:
			default:
				return fmt.Errorf("index: [%s] type error, expect Type: %s", indexName, edgepb.IndexType_name[value.IndexType])
			}
		case 3:
			_, ok := indexValue.(bool)
			if !ok {
				return fmt.Errorf("index: [%s] type error, expect Type: %s", indexName, edgepb.IndexType_name[value.IndexType])
			}
		}
	}

	return nil
}

func queryExprAnalyzer(protoExpr *edgepb.FilterExpression) (*inverted.FilterExpression, error) {
	if protoExpr.GetFilter() != nil {
		f := protoExpr.GetFilter()
		var value interface{}
		switch v := f.Value.(type) {
		case *edgepb.SearchFilter_StringVal:
			value = v.StringVal
		case *edgepb.SearchFilter_IntVal:
			value = v.IntVal
		case *edgepb.SearchFilter_FloatVal:
			value = v.FloatVal
		case *edgepb.SearchFilter_BoolVal:
			value = v.BoolVal
		default:
			return nil, fmt.Errorf("unsupported filter value type")
		}
		return &inverted.FilterExpression{
			Single: &inverted.Filter{
				IndexName: f.IndexName,
				Op:        convertProtoOp(f.Op),
				Value:     value,
			},
		}, nil
	} else if protoExpr.GetComposite() != nil {
		comp := protoExpr.GetComposite()
		var exprs []*inverted.FilterExpression
		for _, pe := range comp.Expressions {
			fe, err := queryExprAnalyzer(pe)
			if err != nil {
				return nil, err
			}
			exprs = append(exprs, fe)
		}
		return &inverted.FilterExpression{
			Composite: &inverted.CompositeFilter{
				Op:          convertProtoLogicalOperator(comp.Op),
				Expressions: exprs,
			},
		}, nil
	}
	return nil, nil
}

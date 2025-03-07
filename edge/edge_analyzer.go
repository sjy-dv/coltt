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
				return fmt.Errorf("primaryKey [%s] is must string", column.IndexName)
			}
			continue
		}
		switch column.IndexType {
		case 0:
			_, ok := value.(string)
			if !ok {
				return fmt.Errorf("index: [%s] type error, expect Type: %s", column.IndexName,
					edgepb.IndexChagedType_name[column.IndexType])
			}
			break
		case 1:
			_, ok := value.(int)
			if ok {
				break
			}
			_, ok = value.(int8)
			if ok {
				break
			}
			_, ok = value.(int16)
			if ok {
				break
			}
			_, ok = value.(int32)
			if ok {
				break
			}
			_, ok = value.(int64)
			if ok {
				break
			}
			_, ok = value.(uint)
			if ok {
				break
			}
			_, ok = value.(uint8)
			if ok {
				break
			}
			_, ok = value.(uint16)
			if ok {
				break
			}
			_, ok = value.(uint32)
			if ok {
				break
			}
			_, ok = value.(uint64)
			if ok {
				break
			}
			return fmt.Errorf("index: [%s] type error, expect Type: %s", column.IndexName,
				edgepb.IndexChagedType_name[column.IndexType])
		case 2:
			_, ok := value.(float32)
			if ok {
				break
			}
			_, ok = value.(float64)
			if ok {
				break
			}
			return fmt.Errorf("index: [%s] type error, expect Type: %s", column.IndexName,
				edgepb.IndexChagedType_name[column.IndexType])
		case 3:
			_, ok := value.(bool)
			if !ok {
				return fmt.Errorf("index: [%s] type error, expect Type: %s", column.IndexName,
					edgepb.IndexChagedType_name[column.IndexType])
			}
			break
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
				return fmt.Errorf("index: [%s] type error, expect Type: %s", indexName,
					edgepb.IndexChagedType_name[value.IndexType])
			}
			break
		case 1:
			_, ok := indexValue.(int)
			if ok {
				break
			}
			_, ok = indexValue.(int8)
			if ok {
				break
			}
			_, ok = indexValue.(int16)
			if ok {
				break
			}
			_, ok = indexValue.(int32)
			if ok {
				break
			}
			_, ok = indexValue.(int64)
			if ok {
				break
			}
			_, ok = indexValue.(uint)
			if ok {
				break
			}
			_, ok = indexValue.(uint8)
			if ok {
				break
			}
			_, ok = indexValue.(uint16)
			if ok {
				break
			}
			_, ok = indexValue.(uint32)
			if ok {
				break
			}
			_, ok = indexValue.(uint64)
			if ok {
				break
			}
			return fmt.Errorf("index: [%s] type error, expect Type: %s", indexName,
				edgepb.IndexChagedType_name[value.IndexType])
		case 2:
			_, ok := indexValue.(float32)
			if ok {
				break
			}
			_, ok = indexValue.(float64)
			if ok {
				break
			}
			return fmt.Errorf("index: [%s] type error, expect Type: %s", indexName,
				edgepb.IndexChagedType_name[value.IndexType])
		case 3:
			_, ok := indexValue.(bool)
			if !ok {
				return fmt.Errorf("index: [%s] type error, expect Type: %s", indexName,
					edgepb.IndexChagedType_name[value.IndexType])
			}
			break
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

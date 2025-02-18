package experimental

import (
	"errors"
	"fmt"

	"github.com/sjy-dv/coltt/gen/protoc/v3/experimentalproto"
)

func basicConditionAnalyzer(indexes []*experimentalproto.Index) error {
	for _, index := range indexes {
		if index.IndexType == experimentalproto.IndexType_Vector {
			if index.EnableNull {
				return fmt.Errorf("index: [%s] is vector, vector is not allowed null", index.IndexName)
			}
		}
	}
	return nil
}

func metadataAnalyzer(inMap map[string]interface{}, analyzer map[string]IndexFeature) error {
	for _, column := range analyzer {
		value, ok := inMap[column.IndexName]
		if !ok {
			if column.EnableNull {
				baseValue := defaultType(column.IndexType)
				if baseValue == nil {
					return fmt.Errorf("index: %s design error, type: %d", column.IndexName, column.IndexType)
				}
				value = baseValue
			} else {
				return fmt.Errorf("index: %s is null, but index design not allowed null value", column.IndexName)
			}
		}
		switch column.IndexType {
		case 0:
			_, ok := value.(string)
			if !ok {
				return fmt.Errorf("index: [%s] type error, expect Type: %s", column.IndexName,
					experimentalproto.IndexChagedType_name[column.IndexType])
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
				experimentalproto.IndexChagedType_name[column.IndexType])
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
				experimentalproto.IndexChagedType_name[column.IndexType])
		case 3:
			_, ok := value.(bool)
			if !ok {
				return fmt.Errorf("index: [%s] type error, expect Type: %s", column.IndexName,
					experimentalproto.IndexChagedType_name[column.IndexType])
			}
			break
		case 4:
			_, ok := value.([]float32)
			if !ok {
				return fmt.Errorf("index: [%s] type error, expect Type: %s", column.IndexName,
					experimentalproto.IndexChagedType_name[column.IndexType])
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
	case 4:
		return []float32{}
	default:
		return nil
	}
}

func validateRatio(multiVectors []*experimentalproto.MultiVectorIndex) error {
	ratio := 0.0
	for _, vector := range multiVectors {
		if vector.IncludeOrNot {
			ratio += float64(vector.GetRatio())
		}
	}
	if ratio > 1 || ratio < 1 {
		return errors.New("The sum of the ratios must be 1.")
	}
	return nil
}

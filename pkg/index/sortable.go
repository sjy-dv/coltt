package index

import (
	"encoding/binary"
	"fmt"
	"math"
)

func toByteSortable[T Invertable](v T) ([]byte, error) {
	switch v := any(v).(type) {
	case string:
		return []byte(v), nil
	case uint64:
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], v)
		return buf[:], nil
	case int64:
		var buf [8]byte
		vv := uint64(v ^ math.MinInt64)
		binary.BigEndian.PutUint64(buf[:], vv)
		return buf[:], nil
	case float64:
		bits := math.Float64bits(v)
		if v >= 0 {
			bits ^= 0x8000000000000000 // math.MinInt64
		} else {
			bits ^= 0xffffffffffffffff // math.MaxUint64
		}
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], bits)
		return buf[:], nil
	}
	return nil, fmt.Errorf("unsupported sortable type %T", v)
}

func fromByteSortable[T Invertable](b []byte, v *T) error {
	switch v := any(v).(type) {
	case *string:
		*v = string(b)
	case *uint64:
		*v = binary.BigEndian.Uint64(b)
	case *int64:
		vv := binary.BigEndian.Uint64(b)
		*v = int64(vv) ^ math.MinInt64
	case *float64:
		bits := binary.BigEndian.Uint64(b)
		// Check sign bit
		if bits&0x8000000000000000 != 0 {
			bits ^= 0x8000000000000000 // math.MinInt64
		} else {
			bits ^= 0xffffffffffffffff // math.MaxUint64
		}
		*v = math.Float64frombits(bits)
	default:
		return fmt.Errorf("unsupported sortable type %T", v)
	}
	return nil
}

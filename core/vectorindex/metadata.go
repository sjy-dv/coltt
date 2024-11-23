package vectorindex

import (
	"math"
	"reflect"
)

type Metadata map[string]any

func Normalize(v []float32) []float32 {
	var norm float32
	out := make([]float32, len(v))
	for i := range v {
		norm += v[i] * v[i]
	}
	if norm == 0 {
		return out
	}

	norm = float32(math.Sqrt(float64(norm)))
	for i := range v {
		out[i] = v[i] / norm
	}

	return out
}

func (xx Metadata) byteSize() uint64 {
	var n uint64 = 0
	for k, v := range xx {
		n += uint64(len(k))
		n += calculateSize(v)
	}
	return n
}

func calculateSize(v interface{}) uint64 {
	var size uint64
	seen := make(map[uintptr]bool)
	size = traverse(reflect.ValueOf(v), seen)
	return size
}

func traverse(v reflect.Value, seen map[uintptr]bool) uint64 {
	if !v.IsValid() {
		return 0
	}

	// Handle pointers and avoid infinite recursion
	if v.Kind() == reflect.Ptr {
		ptr := v.Pointer()
		if ptr == 0 || seen[ptr] {
			return 0
		}
		seen[ptr] = true
		return traverse(v.Elem(), seen)
	}

	var size uint64

	switch v.Kind() {
	case reflect.String:
		size += uint64(len(v.String()))
	case reflect.Bool:
		size += 1
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		size += 8
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		size += 8
	case reflect.Float32, reflect.Float64:
		size += 8
	case reflect.Map:
		for _, key := range v.MapKeys() {
			size += traverse(key, seen)
			size += traverse(v.MapIndex(key), seen)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			size += traverse(v.Index(i), seen)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			size += traverse(v.Field(i), seen)
		}
	}

	return size
}

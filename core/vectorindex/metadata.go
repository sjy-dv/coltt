package vectorindex

import (
	"encoding/binary"
	"io"
	"math"
	"reflect"

	"github.com/vmihailenco/msgpack/v5"
)

type Metadata map[string]any

func (this Metadata) save(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, uint16(len(this))); err != nil {
		return err
	}
	for k, v := range this {
		if err := this.saveKV(w, k, v); err != nil {
			return err
		}
	}
	return nil
}

func (this Metadata) load(r io.Reader) error {
	var length uint16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return err
	}
	for i := 0; i < int(length); i++ {
		k, v, err := this.loadKV(r)
		if err != nil {
			return err
		}
		this[k] = v
	}
	return nil
}

func (this Metadata) saveKV(w io.Writer, k string, v any) error {
	if err := binary.Write(w, binary.BigEndian, uint8(len(k))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, k); err != nil {
		return err
	}

	valBytes, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}

	if err := binary.Write(w, binary.BigEndian, uint16(len(valBytes))); err != nil {
		return err
	}
	if _, err := w.Write(valBytes); err != nil {
		return err
	}
	return nil
}

func (this *Metadata) loadKV(r io.Reader) (string, any, error) {
	var keyLength uint8
	if err := binary.Read(r, binary.BigEndian, &keyLength); err != nil {
		return "", nil, err
	}
	keyBytes := make([]byte, keyLength)
	if _, err := r.Read(keyBytes); err != nil {
		return "", nil, err
	}

	var valLength uint16
	if err := binary.Read(r, binary.BigEndian, &valLength); err != nil {
		return "", nil, err
	}
	valBytes := make([]byte, valLength)
	if _, err := r.Read(valBytes); err != nil {
		return "", nil, err
	}

	var value any
	if err := msgpack.Unmarshal(valBytes, &value); err != nil {
		return "", nil, err
	}

	return string(keyBytes), value, nil
}

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

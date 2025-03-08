package inverted

import (
	"fmt"
	"strconv"
	"sync"

	roaring "github.com/RoaringBitmap/roaring/v2/roaring64"
)

type BitmapIndex struct {
	Shards    map[string]*IndexShard
	shardLock sync.RWMutex
}

type IndexShard struct {
	ShardIndex map[interface{}]*roaring.Bitmap
	rmu        sync.RWMutex
}

func NewBitmapIndex() *BitmapIndex {
	return &BitmapIndex{
		Shards: make(map[string]*IndexShard),
	}
}

func (idx *BitmapIndex) getShard(key string) *IndexShard {
	idx.shardLock.RLock()
	shard, exists := idx.Shards[key]
	idx.shardLock.RUnlock()
	if exists {
		return shard
	}
	idx.shardLock.Lock()
	defer idx.shardLock.Unlock()
	shard, exists = idx.Shards[key]
	if !exists {
		shard = &IndexShard{
			ShardIndex: make(map[interface{}]*roaring.Bitmap),
		}
		idx.Shards[key] = shard
	}
	return shard
}

func (idx *BitmapIndex) Add(nodeId uint64, metadata map[string]interface{}) error {
	for key, val := range metadata {
		shard := idx.getShard(key)
		shard.rmu.Lock()
		if _, exists := shard.ShardIndex[val]; !exists {
			shard.ShardIndex[val] = roaring.New()
		}
		shard.ShardIndex[val].Add(nodeId)
		shard.rmu.Unlock()
	}
	return nil
}

func (idx *BitmapIndex) Remove(nodeId uint64, metadata map[string]interface{}) error {

	for key, val := range metadata {
		shard := idx.getShard(key)
		shard.rmu.Lock()
		if bm, exists := shard.ShardIndex[val]; exists {
			bm.Remove(nodeId)
			if bm.IsEmpty() {
				delete(shard.ShardIndex, val)
			}
		}
		if len(shard.ShardIndex) == 0 {
			shard.rmu.Unlock()
			idx.shardLock.Lock()
			delete(idx.Shards, key)
			idx.shardLock.Unlock()
			continue
		}
		shard.rmu.Unlock()
	}
	return nil
}

func compareValues(a, b interface{}) (int, error) {
	// a가 int64인 경우
	switch va := a.(type) {
	case int64:
		switch vb := b.(type) {
		case int64:
			if va < vb {
				return -1, nil
			} else if va > vb {
				return 1, nil
			}
			return 0, nil
		case int:
			bInt := int64(vb)
			if va < bInt {
				return -1, nil
			} else if va > bInt {
				return 1, nil
			}
			return 0, nil
		case float64:
			if float64(va) < vb {
				return -1, nil
			} else if float64(va) > vb {
				return 1, nil
			}
			return 0, nil
		case float32:
			if float64(va) < float64(vb) {
				return -1, nil
			} else if float64(va) > float64(vb) {
				return 1, nil
			}
			return 0, nil
		case string:
			// 문자열인 경우, 파싱하여 비교
			num, err := strconv.ParseInt(vb, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("cannot convert string %q to int64: %v", vb, err)
			}
			if va < num {
				return -1, nil
			} else if va > num {
				return 1, nil
			}
			return 0, nil
		default:
			return 0, fmt.Errorf("type mismatch: %T vs %T", a, b)
		}
	case int:
		// a가 int인 경우, 변환해서 처리
		aInt := int64(va)
		switch vb := b.(type) {
		case int:
			bInt := int64(vb)
			if aInt < bInt {
				return -1, nil
			} else if aInt > bInt {
				return 1, nil
			}
			return 0, nil
		case int64:
			if aInt < vb {
				return -1, nil
			} else if aInt > vb {
				return 1, nil
			}
			return 0, nil
		case float64:
			if float64(aInt) < vb {
				return -1, nil
			} else if float64(aInt) > vb {
				return 1, nil
			}
			return 0, nil
		case float32:
			if float64(aInt) < float64(vb) {
				return -1, nil
			} else if float64(aInt) > float64(vb) {
				return 1, nil
			}
			return 0, nil
		case string:
			num, err := strconv.ParseInt(vb, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("cannot convert string %q to int64: %v", vb, err)
			}
			if aInt < num {
				return -1, nil
			} else if aInt > num {
				return 1, nil
			}
			return 0, nil
		default:
			return 0, fmt.Errorf("type mismatch: %T vs %T", a, b)
		}
	case float64:
		switch vb := b.(type) {
		case float64:
			if va < vb {
				return -1, nil
			} else if va > vb {
				return 1, nil
			}
			return 0, nil
		case float32:
			if va < float64(vb) {
				return -1, nil
			} else if va > float64(vb) {
				return 1, nil
			}
			return 0, nil
		case int:
			if va < float64(vb) {
				return -1, nil
			} else if va > float64(vb) {
				return 1, nil
			}
			return 0, nil
		case int64:
			if va < float64(vb) {
				return -1, nil
			} else if va > float64(vb) {
				return 1, nil
			}
			return 0, nil
		case string:
			num, err := strconv.ParseFloat(vb, 64)
			if err != nil {
				return 0, fmt.Errorf("cannot convert string %q to float64: %v", vb, err)
			}
			if va < num {
				return -1, nil
			} else if va > num {
				return 1, nil
			}
			return 0, nil
		default:
			return 0, fmt.Errorf("type mismatch: %T vs %T", a, b)
		}
	case float32:
		aFloat := float64(va)
		switch vb := b.(type) {
		case float32:
			bFloat := float64(vb)
			if aFloat < bFloat {
				return -1, nil
			} else if aFloat > bFloat {
				return 1, nil
			}
			return 0, nil
		case float64:
			if aFloat < vb {
				return -1, nil
			} else if aFloat > vb {
				return 1, nil
			}
			return 0, nil
		case int:
			if aFloat < float64(vb) {
				return -1, nil
			} else if aFloat > float64(vb) {
				return 1, nil
			}
			return 0, nil
		case int64:
			if aFloat < float64(vb) {
				return -1, nil
			} else if aFloat > float64(vb) {
				return 1, nil
			}
			return 0, nil
		case string:
			num, err := strconv.ParseFloat(vb, 64)
			if err != nil {
				return 0, fmt.Errorf("cannot convert string %q to float64: %v", vb, err)
			}
			if aFloat < num {
				return -1, nil
			} else if aFloat > num {
				return 1, nil
			}
			return 0, nil
		default:
			return 0, fmt.Errorf("type mismatch: %T vs %T", a, b)
		}
	case string:
		// a가 string인 경우
		switch vb := b.(type) {
		case string:
			if va < vb {
				return -1, nil
			} else if va > vb {
				return 1, nil
			}
			return 0, nil
		case int:
			num, err := strconv.Atoi(va)
			if err != nil {
				return 0, fmt.Errorf("cannot convert string %q to int: %v", va, err)
			}
			if num < vb {
				return -1, nil
			} else if num > vb {
				return 1, nil
			}
			return 0, nil
		case int64:
			num, err := strconv.ParseInt(va, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("cannot convert string %q to int64: %v", va, err)
			}
			if num < vb {
				return -1, nil
			} else if num > vb {
				return 1, nil
			}
			return 0, nil
		case float64:
			num, err := strconv.ParseFloat(va, 64)
			if err != nil {
				return 0, fmt.Errorf("cannot convert string %q to float64: %v", va, err)
			}
			if num < vb {
				return -1, nil
			} else if num > vb {
				return 1, nil
			}
			return 0, nil
		case float32:
			num, err := strconv.ParseFloat(va, 32)
			if err != nil {
				return 0, fmt.Errorf("cannot convert string %q to float32: %v", va, err)
			}
			if num < float64(vb) {
				return -1, nil
			} else if num > float64(vb) {
				return 1, nil
			}
			return 0, nil
		default:
			return 0, fmt.Errorf("type mismatch: %T vs %T", a, b)
		}
	case bool:
		vb, ok := b.(bool)
		if !ok {
			return 0, fmt.Errorf("type mismatch: %T vs %T", a, b)
		}
		if !va && vb {
			return -1, nil
		} else if va && !vb {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", a)
	}
}

func satisfiesOp(f *Filter, key interface{}) (bool, error) {
	cmp, err := compareValues(key, f.Value)
	if err != nil {
		return false, err
	}
	switch f.Op {
	case OpEqual:
		return cmp == 0, nil
	case OpNotEqual:
		return cmp != 0, nil
	case OpGreaterThan:
		return cmp > 0, nil
	case OpGreaterThanEqual:
		return cmp >= 0, nil
	case OpLessThan:
		return cmp < 0, nil
	case OpLessThanEqual:
		return cmp <= 0, nil
	default:
		return false, fmt.Errorf("unsupported op")
	}
}

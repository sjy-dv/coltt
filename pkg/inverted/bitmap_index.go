package inverted

import (
	"fmt"
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
	switch va := a.(type) {
	case int:
		vb, ok := b.(int)
		if !ok {
			return 0, fmt.Errorf("type mismatch")
		}
		if va < vb {
			return -1, nil
		}
		if va > vb {
			return 1, nil
		}
		return 0, nil
	case int64:
		vb, ok := b.(int64)
		if !ok {
			return 0, fmt.Errorf("type mismatch")
		}
		if va < vb {
			return -1, nil
		}
		if va > vb {
			return 1, nil
		}
		return 0, nil
	case float64:
		vb, ok := b.(float64)
		if !ok {
			return 0, fmt.Errorf("type mismatch")
		}
		if va < vb {
			return -1, nil
		}
		if va > vb {
			return 1, nil
		}
		return 0, nil
	case string:
		vb, ok := b.(string)
		if !ok {
			return 0, fmt.Errorf("type mismatch")
		}
		if va < vb {
			return -1, nil
		}
		if va > vb {
			return 1, nil
		}
		return 0, nil
	case bool:
		vb, ok := b.(bool)
		if !ok {
			return 0, fmt.Errorf("type mismatch")
		}
		if !va && vb {
			return -1, nil
		}
		if va && !vb {
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

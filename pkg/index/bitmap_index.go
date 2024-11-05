package index

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	roaring "github.com/RoaringBitmap/roaring/roaring64"
)

type BitmapIndex struct {
	Shards             map[string]*IndexShard
	shardLock          sync.RWMutex
	optimizationTicker *time.Ticker
	stopOptimization   chan bool
}

type IndexShard struct {
	ShardIndex map[string]*roaring.Bitmap
	rmu        sync.RWMutex
}

func NewBitmapIndex() *BitmapIndex {
	return &BitmapIndex{
		Shards:           make(map[string]*IndexShard),
		stopOptimization: make(chan bool),
	}
}

func forcedStringTypeChanger(x interface{}) string {
	switch val := x.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (idx *BitmapIndex) getShard(key string) *IndexShard {
	idx.shardLock.RLock()
	shard, exists := idx.Shards[key]
	idx.shardLock.RUnlock()
	if exists {
		return shard
	}

	//create new shard
	idx.shardLock.Lock()
	defer idx.shardLock.Unlock()
	shard, exists = idx.Shards[key]
	if !exists {
		shard = &IndexShard{
			ShardIndex: make(map[string]*roaring.Bitmap),
		}
		idx.Shards[key] = shard
	}
	return shard
}

func (idx *BitmapIndex) Add(nodeId uint64, metadata map[string]interface{}) error {
	for key, val := range metadata {
		shard := idx.getShard(key)
		shard.rmu.Lock()
		if _, exists := shard.ShardIndex[forcedStringTypeChanger(val)]; !exists {
			shard.ShardIndex[forcedStringTypeChanger(val)] = roaring.New()
		}
		shard.ShardIndex[forcedStringTypeChanger(val)].Add(nodeId)
		shard.rmu.Unlock()
	}
	return nil
}

func (idx *BitmapIndex) Remove(nodeId uint64, metadata map[string]interface{}) error {
	for key, value := range metadata {
		val := forcedStringTypeChanger(value)

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

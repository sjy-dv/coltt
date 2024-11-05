package index

import roaring "github.com/RoaringBitmap/roaring/roaring64"

// using hybrid search
func (idx *BitmapIndex) SearchWitCandidates(
	candidatsIds []uint64, filter map[string]string,
) []uint64 {
	if len(filter) == 0 {
		return candidatsIds
	}

	newBitmap := roaring.New()
	newBitmap.AddMany(candidatsIds)

	for key, value := range filter {
		shard := idx.getShard(key)
		shard.rmu.RLock()
		bm, exists := shard.ShardIndex[value]
		if !exists {
			shard.rmu.RUnlock()
			return []uint64{}
		}
		newBitmap.And(bm)
		shard.rmu.RUnlock()
	}
	return newBitmap.ToArray()
}

// using pure search
func (idx *BitmapIndex) PureSearch(filter map[string]string) []uint64 {
	var result *roaring.Bitmap
	first := true

	for key, value := range filter {
		shard := idx.getShard(key)
		shard.rmu.RLock()
		bm, exists := shard.ShardIndex[value]
		if !exists {
			shard.rmu.RUnlock()
			return []uint64{}
		}
		if first {
			result = bm.Clone()
			first = false
		} else {
			result.And(bm)
		}
		shard.rmu.RUnlock()
	}
	if result != nil {
		return result.ToArray()
	}
	return []uint64{}
}

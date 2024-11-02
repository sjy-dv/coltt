package index

import "github.com/RoaringBitmap/roaring"

// using hybrid search
func (idx *BitmapIndex) SearchWitCandidates(
	candidatsIds []uint32, filter map[string]string,
) []uint32 {
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
			return []uint32{}
		}
		newBitmap.And(bm)
		shard.rmu.RUnlock()
	}
	return newBitmap.ToArray()
}

// using pure search
func (idx *BitmapIndex) PureSearch(filter map[string]string) []uint32 {
	var result *roaring.Bitmap
	first := true

	for key, value := range filter {
		shard := idx.getShard(key)
		shard.rmu.RLock()
		bm, exists := shard.ShardIndex[value]
		if !exists {
			shard.rmu.RUnlock()
			return []uint32{}
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
	return []uint32{}
}

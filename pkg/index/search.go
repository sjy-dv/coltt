// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package index

import roaring "github.com/RoaringBitmap/roaring/v2/roaring64"

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

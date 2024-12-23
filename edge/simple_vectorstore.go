package edge

import (
	"sync"

	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/sharding"
)

type simplevecSpace struct {
	dimension int
	// vectors        map[uint64]Vector
	// vectors        *concurrentmap.Map[uint64, Vector]
	vertices       [EDGE_MAP_SHARD_COUNT]map[uint64]ENode
	verticesMu     [EDGE_MAP_SHARD_COUNT]*sync.RWMutex
	collectionName string
	distance       distance.Space
	quantization   NoQuantization
	// lock           sync.RWMutex
}

func newSimpleVectorstore(config CollectionConfig) *simplevecSpace {
	vecspace := &simplevecSpace{
		dimension:      config.Dimension,
		collectionName: config.CollectionName,
		distance: func() distance.Space {
			if config.Distance == COSINE {
				return distance.NewCosine()
			} else if config.Distance == EUCLIDEAN {
				return distance.NewEuclidean()
			}
			return distance.NewCosine()
		}(),
		quantization: NoQuantization{},
	}
	for i := 0; i < EDGE_MAP_SHARD_COUNT; i++ {
		vecspace.vertices[i] = make(map[uint64]ENode)
		vecspace.verticesMu[i] = &sync.RWMutex{}
	}
	return vecspace
}

func (qx *simplevecSpace) InsertVector(collectionName string, commitId uint64, data ENode) error {
	if qx.distance.Type() == T_COSINE {
		data.Vector = Normalize(data.Vector)
	}
	shardIdx := sharding.ShardVertex(commitId, uint64(EDGE_MAP_SHARD_COUNT))
	qx.verticesMu[shardIdx].Lock()
	defer qx.verticesMu[shardIdx].Unlock()
	qx.vertices[shardIdx][commitId] = data
	return nil
}

func (qx *simplevecSpace) UpdateVector(collectionName string, id uint64, data ENode) error {
	if qx.distance.Type() == T_COSINE {
		data.Vector = Normalize(data.Vector)
	}
	shardIdx := sharding.ShardVertex(id, uint64(EDGE_MAP_SHARD_COUNT))
	qx.verticesMu[shardIdx].Lock()
	defer qx.verticesMu[shardIdx].Unlock()
	qx.vertices[shardIdx][id] = data
	return nil
}

func (qx *simplevecSpace) RemoveVector(collectionName string, id uint64) error {
	shardIdx := sharding.ShardVertex(id, uint64(EDGE_MAP_SHARD_COUNT))
	qx.verticesMu[shardIdx].Lock()
	defer qx.verticesMu[shardIdx].Unlock()
	delete(qx.vertices[shardIdx], id)
	return nil
}

func (qx *simplevecSpace) FullScan(collectionName string, target Vector, topK int,
) ([]*SearchResultItem, error) {
	if qx.distance.Type() == T_COSINE {
		target = Normalize(target)
	}
	pq := NewPriorityQueue(topK)
	for shard := 0; shard < EDGE_MAP_SHARD_COUNT; shard++ {
		qx.verticesMu[shard].RLock()
		for uid, node := range qx.vertices[shard] {
			sim := qx.quantization.Similarity(target, node.Vector, qx.distance)
			pq.Add(&SearchResultItem{
				Id:       uid,
				Score:    sim,
				Metadata: node.Metadata,
			})
		}
		qx.verticesMu[shard].RUnlock()
	}
	return pq.ToSlice(), nil
}

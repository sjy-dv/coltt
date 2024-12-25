package edge

import (
	"fmt"
	"sync"

	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/sharding"
)

type f16vecSpace struct {
	dimension      int
	vertices       [EDGE_MAP_SHARD_COUNT]map[uint64]ENodeF16
	verticesMu     [EDGE_MAP_SHARD_COUNT]*sync.RWMutex
	collectionName string
	distance       distance.Space
	quantization   Float16Quantization
	lock           sync.RWMutex
}

func newF16Vectorstore(config CollectionConfig) *f16vecSpace {
	vecspace := &f16vecSpace{
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
		quantization: Float16Quantization{},
	}
	for i := 0; i < EDGE_MAP_SHARD_COUNT; i++ {
		vecspace.vertices[i] = make(map[uint64]ENodeF16)
		vecspace.verticesMu[i] = &sync.RWMutex{}
	}
	return vecspace
}

func (qx *f16vecSpace) InsertVector(collectionName string, commitId uint64, data ENode) error {
	if qx.distance.Type() == T_COSINE {
		data.Vector = Normalize(data.Vector)
	}
	lower, err := qx.quantization.Lower(data.Vector)
	if err != nil {
		return fmt.Errorf(ErrQuantizedFailed, err)
	}

	shardIdx := sharding.ShardVertex(commitId, uint64(EDGE_MAP_SHARD_COUNT))
	qx.verticesMu[shardIdx].Lock()
	defer qx.verticesMu[shardIdx].Unlock()
	qx.vertices[shardIdx][commitId] = ENodeF16{Vector: lower, Metadata: data.Metadata}
	return nil
}

func (qx *f16vecSpace) UpdateVector(collectionName string, id uint64, data ENode) error {
	if qx.distance.Type() == T_COSINE {
		data.Vector = Normalize(data.Vector)
	}
	lower, err := qx.quantization.Lower(data.Vector)
	if err != nil {
		return fmt.Errorf(ErrQuantizedFailed, err)
	}
	shardIdx := sharding.ShardVertex(id, uint64(EDGE_MAP_SHARD_COUNT))
	qx.verticesMu[shardIdx].Lock()
	defer qx.verticesMu[shardIdx].Unlock()
	qx.vertices[shardIdx][id] = ENodeF16{Vector: lower, Metadata: data.Metadata}
	return nil
}

func (qx *f16vecSpace) RemoveVector(collectionName string, id uint64) error {
	shardIdx := sharding.ShardVertex(id, uint64(EDGE_MAP_SHARD_COUNT))
	qx.verticesMu[shardIdx].Lock()
	defer qx.verticesMu[shardIdx].Unlock()
	delete(qx.vertices[shardIdx], id)
	return nil
}

func (qx *f16vecSpace) FullScan(collectionName string, target Vector, topK int, highCpu bool,
) ([]*SearchResultItem, error) {
	if qx.distance.Type() == T_COSINE {
		target = Normalize(target)
	}
	lower, err := qx.quantization.Lower(target)
	if err != nil {
		return nil, fmt.Errorf(ErrQuantizedFailed, err)
	}
	pq := NewPriorityQueue(topK)
	if !highCpu {
		for shard := 0; shard < EDGE_MAP_SHARD_COUNT; shard++ {
			qx.verticesMu[shard].RLock()
			for uid, node := range qx.vertices[shard] {
				sim := qx.quantization.Similarity(lower, node.Vector, qx.distance)
				pq.Add(&SearchResultItem{
					Id:       uid,
					Score:    sim,
					Metadata: node.Metadata,
				})
			}
			qx.verticesMu[shard].RUnlock()
		}
	} else {
		type shardResult struct {
			Items []*SearchResultItem
		}
		results := make([]shardResult, EDGE_MAP_SHARD_COUNT)
		var concurrenyWorker sync.WaitGroup
		concurrenyWorker.Add(EDGE_MAP_SHARD_COUNT)
		for shard := 0; shard < EDGE_MAP_SHARD_COUNT; shard++ {
			go func(shard int) {
				defer concurrenyWorker.Done()
				localpq := NewPriorityQueue(topK)
				qx.verticesMu[shard].RLock()
				for uid, node := range qx.vertices[shard] {
					sim := qx.quantization.Similarity(lower, node.Vector, qx.distance)
					localpq.Add(&SearchResultItem{
						Id:       uid,
						Score:    sim,
						Metadata: node.Metadata,
					})
				}
				qx.verticesMu[shard].RUnlock()
				results[shard].Items = localpq.ToSlice()
			}(shard)
		}
		concurrenyWorker.Wait()
		for _, res := range results {
			for _, item := range res.Items {
				pq.Add(item)
			}
		}
	}
	return pq.ToSlice(), nil
}

package edge

import (
	"sync"

	"github.com/sjy-dv/nnv/pkg/concurrentmap"
	"github.com/sjy-dv/nnv/pkg/distancer"
)

type simplevecSpace struct {
	dimension int
	// vectors        map[uint64]Vector
	vectors        *concurrentmap.Map[uint64, Vector]
	collectionName string
	distance       distancer.Provider
	quantization   NoQuantization
	lock           sync.RWMutex
}

func newSimpleVectorstore(config CollectionConfig) *simplevecSpace {
	return &simplevecSpace{
		dimension:      config.Dimension,
		vectors:        concurrentmap.New[uint64, Vector](),
		collectionName: config.CollectionName,
		distance: func() distancer.Provider {
			if config.Distance == COSINE {
				return distancer.NewCosineDistanceProvider()
			} else if config.Distance == EUCLIDEAN {
				return distancer.NewL2SquaredProvider()
			}
			return distancer.NewCosineDistanceProvider()
		}(),
		quantization: NoQuantization{},
	}
}

func (qx *simplevecSpace) InsertVector(collectionName string, commitId uint64, vector Vector) error {
	if qx.distance.Type() == "cosine-dot" {
		vector = Normalize(vector)
	}
	qx.vectors.Set(commitId, vector)
	return nil
}

func (qx *simplevecSpace) UpdateVector(collectionName string, id uint64, vector Vector) error {
	if qx.distance.Type() == "cosine-dot" {
		vector = Normalize(vector)
	}
	qx.vectors.Set(id, vector)
	return nil
}

func (qx *simplevecSpace) RemoveVector(collectionName string, id uint64) error {
	qx.vectors.Del(id)
	return nil
}

func (qx *simplevecSpace) FullScan(collectionName string, target Vector, topK int,
) (*ResultSet, error) {
	if qx.distance.Type() == "cosine-dot" {
		target = Normalize(target)
	}
	rs := NewResultSet(topK)
	qx.vectors.ForEach(func(u uint64, v Vector) bool {
		sim, _ := qx.quantization.Similarity(target, v, qx.distance)
		rs.AddResult(ID(u), sim)
		return true
	})
	return rs, nil
}

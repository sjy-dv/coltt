package edge

import (
	"fmt"
	"sync"

	"github.com/sjy-dv/nnv/pkg/concurrentmap"
	"github.com/sjy-dv/nnv/pkg/distance"
)

type f8vecSpace struct {
	dimension      int
	vectors        *concurrentmap.Map[uint64, float8Vec]
	collectionName string
	distance       distance.Space
	quantization   Float8Quantization
	lock           sync.RWMutex
}

func newF8Vectorstore(config CollectionConfig) *f8vecSpace {
	return &f8vecSpace{
		dimension:      config.Dimension,
		vectors:        concurrentmap.New[uint64, float8Vec](),
		collectionName: config.CollectionName,
		distance: func() distance.Space {
			if config.Distance == COSINE {
				return distance.NewCosine()
			} else if config.Distance == EUCLIDEAN {
				return distance.NewEuclidean()
			}
			return distance.NewCosine()
		}(),
		quantization: Float8Quantization{},
	}
}

func (qx *f8vecSpace) InsertVector(collectionName string, commitId uint64, vector Vector) error {
	if qx.distance.Type() == "cosine-dot" {
		vector = Normalize(vector)
	}
	lower, err := qx.quantization.Lower(vector)
	if err != nil {
		return fmt.Errorf(ErrQuantizedFailed, err)
	}
	// qx.lock.Lock()
	// qx.vectors[commitId] = lower
	// qx.lock.Unlock()
	qx.vectors.Set(commitId, lower)
	return nil
}

func (qx *f8vecSpace) UpdateVector(collectionName string, id uint64, vector Vector) error {
	if qx.distance.Type() == "cosine-dot" {
		vector = Normalize(vector)
	}
	lower, err := qx.quantization.Lower(vector)
	if err != nil {
		return fmt.Errorf(ErrQuantizedFailed, err)
	}
	// qx.lock.Lock()
	// qx.vectors[id] = lower
	// qx.lock.Unlock()
	qx.vectors.Set(id, lower)
	return nil
}

func (qx *f8vecSpace) RemoveVector(collectionName string, id uint64) error {
	// qx.lock.Lock()
	// delete(qx.vectors, id)
	// qx.lock.Unlock()
	qx.vectors.Del(id)
	return nil
}

func (qx *f8vecSpace) FullScan(collectionName string, target Vector, topK int,
) (*ResultSet, error) {
	if qx.distance.Type() == "cosine-dot" {
		target = Normalize(target)
	}
	rs := NewResultSet(topK)

	lower, err := qx.quantization.Lower(target)
	if err != nil {
		return nil, fmt.Errorf(ErrQuantizedFailed, err)
	}
	// qx.lock.RLock()
	// defer qx.lock.RUnlock()
	// for index, qvec := range qx.vectors {
	// 	sim := qx.quantization.Similarity(lower, qvec, qx.distance)
	// 	rs.AddResult(ID(index), sim)
	// }
	qx.vectors.ForEach(func(u uint64, fv float8Vec) bool {
		sim := qx.quantization.Similarity(lower, fv, qx.distance)
		rs.AddResult(ID(u), sim)
		return true
	})
	return rs, nil
}

package edge

import (
	"fmt"
	"sync"

	"github.com/sjy-dv/nnv/pkg/concurrentmap"
	"github.com/sjy-dv/nnv/pkg/distance"
)

type bf16vecSpace struct {
	dimension      int
	vectors        *concurrentmap.Map[uint64, bfloat16Vec]
	collectionName string
	distance       distance.Space
	quantization   BFloat16Quantization
	lock           sync.RWMutex
}

func newBF16Vectorstore(config CollectionConfig) *bf16vecSpace {
	return &bf16vecSpace{
		dimension:      config.Dimension,
		vectors:        concurrentmap.New[uint64, bfloat16Vec](),
		collectionName: config.CollectionName,
		distance: func() distance.Space {
			if config.Distance == COSINE {
				return distance.NewCosine()
			} else if config.Distance == EUCLIDEAN {
				return distance.NewEuclidean()
			}
			return distance.NewCosine()
		}(),
		quantization: BFloat16Quantization{},
	}
}

func (qx *bf16vecSpace) InsertVector(collectionName string, commitId uint64, vector Vector) error {
	if qx.distance.Type() == "cosine-dot" {
		vector = Normalize(vector)
	}
	lower, err := qx.quantization.Lower(vector)
	if err != nil {
		return fmt.Errorf(ErrQuantizedFailed, err)
	}
	qx.vectors.Set(commitId, lower)
	return nil
}

func (qx *bf16vecSpace) UpdateVector(collectionName string, id uint64, vector Vector) error {
	if qx.distance.Type() == "cosine-dot" {
		vector = Normalize(vector)
	}
	lower, err := qx.quantization.Lower(vector)
	if err != nil {
		return fmt.Errorf(ErrQuantizedFailed, err)
	}
	qx.vectors.Set(id, lower)
	return nil
}

func (qx *bf16vecSpace) RemoveVector(collectionName string, id uint64) error {
	qx.vectors.Del(id)
	return nil
}

func (qx *bf16vecSpace) FullScan(collectionName string, target Vector, topK int,
) (*ResultSet, error) {
	if qx.distance.Type() == "cosine-dot" {
		target = Normalize(target)
	}
	rs := NewResultSet(topK)

	lower, err := qx.quantization.Lower(target)
	if err != nil {
		return nil, fmt.Errorf(ErrQuantizedFailed, err)
	}
	qx.vectors.ForEach(func(u uint64, fv bfloat16Vec) bool {
		sim := qx.quantization.Similarity(lower, fv, qx.distance)
		rs.AddResult(ID(u), sim)
		return true
	})
	return rs, nil
}

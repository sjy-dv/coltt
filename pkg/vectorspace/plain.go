package vectorspace

import (
	"fmt"
	"math"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/nnv/pkg/cache"
	"github.com/sjy-dv/nnv/pkg/conversion"
	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/storage"
)

type plainStore struct {
	items  *cache.ItemCache[uint64, plainPoint]
	distFn distance.FloatDistFunc
}

func (ps plainStore) Exists(id uint64) bool {
	_, err := ps.items.Get(id)
	return err == nil
}

func (ps plainStore) Get(id uint64) (VectorStorePoint, error) {
	return ps.items.Get(id)
}

func (ps plainStore) GetMany(ids ...uint64) ([]VectorStorePoint, error) {
	points, err := ps.items.GetMany(ids...)
	if err != nil {
		return nil, err
	}
	// Amazing casting here, why not just return points, nil? Oh well, it's not
	// the same type.
	ret := make([]VectorStorePoint, len(points))
	for i, p := range points {
		ret[i] = p
	}
	return ret, nil
}

func (ps plainStore) ForEach(fn func(VectorStorePoint) error) error {
	return ps.items.ForEach(func(id uint64, point plainPoint) error {
		return fn(point)
	})
}

func (ps plainStore) SizeInMemory() int64 {
	return ps.items.SizeInMemory()
}

func (ps plainStore) UpdateStorage(storage storage.Storage) {
	ps.items.UpdateStorage(storage)
}

func (ps plainStore) Set(id uint64, vector []float32) (VectorStorePoint, error) {
	point := plainPoint{
		id:     id,
		Vector: vector,
	}
	ps.items.Put(id, point)
	return point, nil
}

func (ps plainStore) Delete(ids ...uint64) error {
	return ps.items.Delete(ids...)

}

func (ps plainStore) Fit() error {
	return nil
}

func (ps plainStore) DistanceFromFloat(x []float32) PointIdDistFn {
	return func(y VectorStorePoint) float32 {
		point, ok := y.(plainPoint)
		if !ok {
			log.Warn().Uint64("id", y.Id()).Msg("point not found for distance calculation")
			return math.MaxFloat32
		}
		return ps.distFn(x, point.Vector)
	}
}

func (ps plainStore) DistanceFromPoint(x VectorStorePoint) PointIdDistFn {
	pointX, okX := x.(plainPoint)
	return func(y VectorStorePoint) float32 {
		pointY, okY := y.(plainPoint)
		if !okX || !okY {
			log.Warn().Uint64("idX", x.Id()).Uint64("idY", y.Id()).Msg("point not found for distance calculation")
			return math.MaxFloat32
		}
		return ps.distFn(pointX.Vector, pointY.Vector)
	}
}

func (ps plainStore) Flush() error {
	return ps.items.Flush()
}

type plainPoint struct {
	id     uint64
	Vector []float32
}

func (pp plainPoint) Id() uint64 {
	return pp.id
}

func (pp plainPoint) IdFromKey(key []byte) (uint64, bool) {
	return conversion.NodeIdFromKey(key, 'v')
}

func (pp plainPoint) SizeInMemory() int64 {
	return int64(8 + 4*len(pp.Vector))
}

// Always returns false as we don't track dirty state.
func (pp plainPoint) CheckAndClearDirty() bool {
	return false
}

func (pp plainPoint) ReadFrom(id uint64, storage storage.Storage) (point plainPoint, err error) {
	point.id = id
	vectorBytes := storage.Get(conversion.NodeKey(id, 'v'))
	if vectorBytes == nil {
		err = cache.ErrNotFound
		return
	}
	point.Vector = conversion.BytesToFloat32(vectorBytes)
	return
}

func (pp plainPoint) WriteTo(id uint64, storage storage.Storage) error {
	if err := storage.Put(conversion.NodeKey(id, 'v'), conversion.Float32ToBytes(pp.Vector)); err != nil {
		return fmt.Errorf("could not write plain point vector: %w", err)
	}
	return nil
}
func (pp plainPoint) DeleteFrom(id uint64, storage storage.Storage) error {
	if err := storage.Delete(conversion.NodeKey(id, 'v')); err != nil {
		return fmt.Errorf("could not delete plain point vector: %w", err)
	}
	return nil
}

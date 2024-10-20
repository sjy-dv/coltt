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

package vectorspace

import (
	"fmt"
	"math"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/nnv/pkg/cache"
	"github.com/sjy-dv/nnv/pkg/conversion"
	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/models"
	"github.com/sjy-dv/nnv/storage"
)

const binaryQuantizerThresholdKey = "_binaryQuantizerThreshold"

type binaryQuantizer struct {
	threshold   []float32
	params      models.BinaryQuantizerParamaters
	items       *cache.ItemCache[uint64, *binaryQuantizedPoint]
	storage     storage.Storage
	floatDistFn distance.FloatDistFunc
	bitDistFn   distance.BitDistFunc
}

func newBinaryQuantizer(storage storage.Storage, floatDistFn distance.FloatDistFunc, params models.BinaryQuantizerParamaters, vectorLen int) (*binaryQuantizer, error) {
	// ---------------------------
	bitDistFn, err := distance.GetBitDistanceFn(params.DistanceMetric)
	if err != nil {
		return nil, fmt.Errorf("failed to get bit distance function: %w", err)
	}
	// ---------------------------
	bq := &binaryQuantizer{
		items:       cache.NewItemCache[uint64, *binaryQuantizedPoint](storage),
		params:      params,
		floatDistFn: floatDistFn,
		bitDistFn:   bitDistFn,
		storage:     storage,
	}
	// Setup the threshold, if given
	if params.Threshold != nil {
		bq.threshold = make([]float32, vectorLen)
		for i := range bq.threshold {
			bq.threshold[i] = *params.Threshold
		}
	} else {
		// Check storage for stored threshold
		floatBytes := storage.Get([]byte(binaryQuantizerThresholdKey))
		if floatBytes != nil {
			bq.threshold = conversion.BytesToFloat32(floatBytes)
		}
	}
	return bq, nil
}

func (bq *binaryQuantizer) Exists(id uint64) bool {
	_, err := bq.items.Get(id)
	return err == nil
}

func (bq *binaryQuantizer) Get(id uint64) (VectorStorePoint, error) {
	return bq.items.Get(id)
}

func (bq *binaryQuantizer) GetMany(ids ...uint64) ([]VectorStorePoint, error) {
	points, err := bq.items.GetMany(ids...)
	if err != nil {
		return nil, err
	}
	ret := make([]VectorStorePoint, len(points))
	for i, p := range points {
		ret[i] = p
	}
	return ret, nil

}

func (bq *binaryQuantizer) ForEach(fn func(VectorStorePoint) error) error {
	return bq.items.ForEach(func(id uint64, point *binaryQuantizedPoint) error {
		return fn(point)
	})
}

func (bq *binaryQuantizer) SizeInMemory() int64 {
	return bq.items.SizeInMemory()
}

func (bq *binaryQuantizer) UpdateStorage(storage storage.Storage) {
	bq.items.UpdateStorage(storage)
	bq.storage = storage
}

func (bq *binaryQuantizer) encode(vector []float32) []uint64 {
	if bq.threshold == nil {
		return nil
	}
	// How many uint64s do we need?
	numUint64s := len(vector) / 64
	if len(vector)%64 != 0 {
		numUint64s++
	}
	encoded := make([]uint64, numUint64s)
	/* Our goal here is to convert the float32 vector into a binary vector. We
	 * do this by setting the bit at position i in the binary vector to 1 if the
	 * value at position i in the float32 vector is greater than the threshold.
	 *
	 * For example: if the threshold is 0.5 and the float32 vector is [0.1, 0.6,
	 * 0.7, 0.4], the binary vector would be [0, 1, 1, 0] in LittleEndian.
	 *
	 * This is then encoded into a uint64 array where each uint64 represents 64
	 * bits of the binary vector.
	 */
	for i, v := range vector {
		if v > bq.threshold[i] {
			encoded[i/64] |= 1 << (i % 64)
		}
	}
	return encoded
}

func (bq *binaryQuantizer) Set(id uint64, vector []float32) (VectorStorePoint, error) {
	point := &binaryQuantizedPoint{
		id:           id,
		Vector:       vector,
		BinaryVector: bq.encode(vector),
	}
	bq.items.Put(id, point)
	return point, nil
}

func (bq *binaryQuantizer) Delete(ids ...uint64) error {
	return bq.items.Delete(ids...)
}

func (bq *binaryQuantizer) Fit() error {
	// Have we already fitted the quantizer or are there enough points to fit it? The short-circuiting
	// here is important to avoid unnecessary work of counting the items.
	if bq.threshold != nil || bq.items.Count() < bq.params.TriggerThreshold {
		return nil
	}
	// ---------------------------
	/* Time to fit. We are doing two passes. First pass computes the mean of the
	 * vectors. The second pass encodes the vectors. */
	count := 0
	var sum []float32
	startTime := time.Now()
	err := bq.items.ForEach(func(id uint64, point *binaryQuantizedPoint) error {
		if sum == nil {
			sum = make([]float32, len(point.Vector))
		}
		for i, v := range point.Vector {
			sum[i] += v
		}
		count++
		return nil
	})
	if err != nil {
		return err
	}
	for i := range sum {
		sum[i] /= float32(count)
	}
	bq.threshold = sum
	// ---------------------------
	// Second pass to encode
	err = bq.items.ForEach(func(id uint64, point *binaryQuantizedPoint) error {
		point.BinaryVector = bq.encode(point.Vector)
		point.isDirty = true
		return nil
	})
	log.Debug().Dur("duration", time.Since(startTime)).Int("thresholdLen", len(bq.threshold)).Msg("fitted binary quantizer")
	// ---------------------------
	return err

}

func (bq *binaryQuantizer) DistanceFromFloat(x []float32) PointIdDistFn {
	// It's okay to duplicate code inside the distance function here because it
	// avoids the if statement check for each distance calculation. Recall that
	// there are a lot of distance calculations in vector stores.
	if bq.threshold != nil {
		encodedX := bq.encode(x)
		return func(y VectorStorePoint) float32 {
			pointY, ok := y.(*binaryQuantizedPoint)
			if !ok {
				log.Warn().Uint64("id", y.Id()).Msg("point not found for distance calculation")
				return math.MaxFloat32
			}
			return bq.bitDistFn(encodedX, pointY.BinaryVector)
		}
	}
	/* Here we fall back to the original vector if the threshold is not set. */
	return func(y VectorStorePoint) float32 {
		pointY, ok := y.(*binaryQuantizedPoint)
		if !ok {
			log.Warn().Uint64("id", y.Id()).Msg("point not found for distance calculation")
			return math.MaxFloat32
		}
		return bq.floatDistFn(x, pointY.Vector)
	}
}

func (bq *binaryQuantizer) DistanceFromPoint(x VectorStorePoint) PointIdDistFn {
	pointX, okX := x.(*binaryQuantizedPoint)
	if bq.threshold != nil {
		return func(y VectorStorePoint) float32 {
			pointB, okB := y.(*binaryQuantizedPoint)
			if !okX || !okB {
				log.Warn().Uint64("idX", x.Id()).Uint64("idY", y.Id()).Msg("point not found for distance calculation")
				return math.MaxFloat32
			}
			return bq.bitDistFn(pointX.BinaryVector, pointB.BinaryVector)
		}
	}
	// Fallback to original vector
	return func(y VectorStorePoint) float32 {
		pointB, okB := y.(*binaryQuantizedPoint)
		if !okX || !okB {
			log.Warn().Uint64("idX", x.Id()).Uint64("idY", y.Id()).Msg("point not found for distance calculation")
			return math.MaxFloat32
		}
		return bq.floatDistFn(pointX.Vector, pointB.Vector)
	}
}

func (bq *binaryQuantizer) Flush() error {
	if err := bq.items.Flush(); err != nil {
		return err
	}
	if len(bq.threshold) > 0 {
		return bq.storage.Put([]byte(binaryQuantizerThresholdKey), conversion.Float32ToBytes(bq.threshold))
	}
	return nil
}

// ---------------------------

type binaryQuantizedPoint struct {
	id           uint64
	Vector       []float32
	BinaryVector []uint64
	isDirty      bool
}

func (bqp *binaryQuantizedPoint) Id() uint64 {
	return bqp.id
}

func (bqp *binaryQuantizedPoint) IdFromKey(key []byte) (uint64, bool) {
	return conversion.NodeIdFromKey(key, 'v')
}

func (bqp *binaryQuantizedPoint) SizeInMemory() int64 {
	return int64(len(bqp.Vector)*4 + len(bqp.BinaryVector)*8)
}

func (bqp *binaryQuantizedPoint) CheckAndClearDirty() bool {
	// This case occurs usually after fitting the quantizer. That is we modify
	// the binary vectors and request that the points are rewritten.
	dirty := bqp.isDirty
	bqp.isDirty = false
	return dirty
}

func (bqp *binaryQuantizedPoint) ReadFrom(id uint64, storage storage.Storage) (point *binaryQuantizedPoint, err error) {
	point = &binaryQuantizedPoint{id: id}
	// ---------------------------
	binaryVecBytes := storage.Get(conversion.NodeKey(id, 'q'))
	if binaryVecBytes != nil {
		point.BinaryVector = conversion.BytesToEdgeList(binaryVecBytes)
		/* NOTE: We don't load the full vector if the quantised version exists.
		 * This is what saves memory. */
		return
	}
	// ---------------------------
	fullVecBytes := storage.Get(conversion.NodeKey(id, 'v'))
	if fullVecBytes == nil {
		err = cache.ErrNotFound
		return
	}
	point.Vector = conversion.BytesToFloat32(fullVecBytes)
	// ---------------------------
	return
}

func (bqp *binaryQuantizedPoint) WriteTo(id uint64, storage storage.Storage) error {
	if len(bqp.BinaryVector) != 0 {
		if err := storage.Put(conversion.NodeKey(id, 'q'), conversion.EdgeListToBytes(bqp.BinaryVector)); err != nil {
			return err
		}
		// We avoid writing the full vector if the quantised version exists.
		return nil
	}
	if len(bqp.Vector) != 0 {
		if err := storage.Put(conversion.NodeKey(id, 'v'), conversion.Float32ToBytes(bqp.Vector)); err != nil {
			return err
		}
	}
	return nil
}

func (bqp *binaryQuantizedPoint) DeleteFrom(id uint64, storage storage.Storage) error {
	if err := storage.Delete(conversion.NodeKey(id, 'v')); err != nil {
		return err
	}
	if err := storage.Delete(conversion.NodeKey(id, 'q')); err != nil {
		return err
	}
	return nil
}

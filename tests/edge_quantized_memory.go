// edge_quantized_mem.go
package edge

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type QuantizedMemoryBackend[V any, Q Quantization[V]] struct {
	vecs         []*V
	rng          *rand.Rand
	dim          int
	quantization Q
	mu           sync.RWMutex
}

var _ scannableBackend = &QuantizedMemoryBackend[Vector, NoQuantization]{}
var _ VectorBackend = &QuantizedMemoryBackend[Vector, NoQuantization]{}

func NewQuantizedMemoryBackend[V any, Q Quantization[V]](dimensions int, quantization Q) *QuantizedMemoryBackend[V, Q] {
	return &QuantizedMemoryBackend[V, Q]{
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
		dim:          dimensions,
		quantization: quantization,
		vecs:         make([]*V, 0),
	}
}

func (q *QuantizedMemoryBackend[V, Q]) Close() error {
	return nil
}

func (q *QuantizedMemoryBackend[V, Q]) PutVector(id ID, vector Vector) error {
	if len(vector) != q.dim {
		return errors.New("QuantizedMemoryBackend: vector dimension doesn't match")
	}

	lower, err := q.quantization.Lower(vector)
	if err != nil {
		return err
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	var v V
	// Type assertion to convert L (from Lower) to V
	lowerV, ok := any(lower).(V)
	if !ok {
		return errors.New("QuantizedMemoryBackend: Lower type assertion failed")
	}
	v = lowerV

	idx := int(id)
	if idx < len(q.vecs) {
		q.vecs[idx] = &v
	} else if idx == len(q.vecs) {
		q.vecs = append(q.vecs, &v)
	} else {
		// Fill the gap with nils
		for len(q.vecs) < idx {
			q.vecs = append(q.vecs, nil)
		}
		q.vecs = append(q.vecs, &v)
	}
	return nil
}

func (q *QuantizedMemoryBackend[V, Q]) ComputeSimilarity(vector Vector, targetID ID) (float32, error) {
	lower, err := q.quantization.Lower(vector)
	if err != nil {
		return 0, err
	}

	q.mu.RLock()
	defer q.mu.RUnlock()

	idx := int(targetID)
	if idx >= len(q.vecs) {
		return 0, ErrIDNotFound
	}
	targetPtr := q.vecs[idx]
	if targetPtr == nil {
		return 0, ErrIDNotFound
	}
	target := *targetPtr
	return q.quantization.Similarity(target, lower), nil
}

func (q *QuantizedMemoryBackend[V, Q]) Info() BackendInfo {
	return BackendInfo{
		HasIndexData: false,
		Dimensions:   q.dim,
	}
}

func (q *QuantizedMemoryBackend[V, Q]) Exists(id ID) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	i := int(id)
	if i >= len(q.vecs) {
		return false
	}
	return q.vecs[i] != nil
}

func (q *QuantizedMemoryBackend[V, Q]) GetVector(id ID) (V, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	i := int(id)
	if i >= len(q.vecs) {
		return *new(V), ErrIDNotFound
	}
	vPtr := q.vecs[i]
	if vPtr == nil {
		return *new(V), ErrIDNotFound
	}
	return *vPtr, nil
}

func (q *QuantizedMemoryBackend[V, Q]) RemoveVector(id ID) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	idx := int(id)
	if idx >= len(q.vecs) {
		return ErrIDNotFound
	}
	if q.vecs[idx] == nil {
		return ErrIDNotFound
	}
	// Set to nil to remove
	q.vecs[idx] = nil
	return nil
}

func (q *QuantizedMemoryBackend[V, Q]) ForEachVector(cb func(ID) error) error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	for i, vPtr := range q.vecs {
		if vPtr == nil {
			continue
		}
		err := cb(ID(i))
		if err != nil {
			return err
		}
	}
	return nil
}

// edge_mem.go
package edge

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/viterin/vek/vek32"
)

type MemoryBackend struct {
	vecs []Vector
	rng  *rand.Rand
	dim  int
	mu   sync.RWMutex
}

var _ scannableBackend = &MemoryBackend{}
var _ VectorBackend = &MemoryBackend{}

func NewMemoryBackend(dimensions int) *MemoryBackend {
	return &MemoryBackend{
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
		dim:  dimensions,
		vecs: make([]Vector, 0),
	}
}

func (mem *MemoryBackend) Close() error {
	return nil
}

func (mem *MemoryBackend) PutVector(id ID, vector Vector) error {
	if len(vector) != mem.dim {
		return errors.New("MemoryBackend: vector dimension doesn't match")
	}

	mem.mu.Lock()
	defer mem.mu.Unlock()

	idx := int(id)
	if idx < len(mem.vecs) {
		mem.vecs[idx] = vector
	} else if idx == len(mem.vecs) {
		mem.vecs = append(mem.vecs, vector)
	} else {
		// Fill the gap with zero vectors
		for len(mem.vecs) < idx {
			mem.vecs = append(mem.vecs, make(Vector, mem.dim))
		}
		mem.vecs = append(mem.vecs, vector)
	}
	return nil
}

func (mem *MemoryBackend) ComputeSimilarity(vector Vector, targetID ID) (float32, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()
	if int(targetID) >= len(mem.vecs) {
		return 0, ErrIDNotFound
	}
	target := mem.vecs[targetID]
	if target == nil || len(target) != mem.dim {
		return 0, ErrIDNotFound
	}
	return vek32.CosineSimilarity(target, vector), nil
}

func (mem *MemoryBackend) Info() BackendInfo {
	return BackendInfo{
		HasIndexData: false,
		Dimensions:   mem.dim,
	}
}

func (mem *MemoryBackend) Exists(id ID) bool {
	mem.mu.RLock()
	defer mem.mu.RUnlock()
	idx := int(id)
	if idx >= len(mem.vecs) {
		return false
	}
	return mem.vecs[idx] != nil && len(mem.vecs[idx]) == mem.dim
}

func (mem *MemoryBackend) GetVector(id ID) (Vector, error) {
	mem.mu.RLock()
	defer mem.mu.RUnlock()
	idx := int(id)
	if idx >= len(mem.vecs) {
		return nil, ErrIDNotFound
	}
	v := mem.vecs[idx]
	if v == nil || len(v) != mem.dim {
		return nil, ErrIDNotFound
	}
	return v, nil
}

func (mem *MemoryBackend) RemoveVector(id ID) error {
	mem.mu.Lock()
	defer mem.mu.Unlock()
	idx := int(id)
	if idx >= len(mem.vecs) {
		return ErrIDNotFound
	}
	if mem.vecs[idx] == nil {
		return ErrIDNotFound
	}
	// Set to nil to remove
	mem.vecs[idx] = nil
	return nil
}

func (mem *MemoryBackend) ForEachVector(cb func(ID) error) error {
	mem.mu.RLock()
	defer mem.mu.RUnlock()
	for i, v := range mem.vecs {
		if v == nil || len(v) != mem.dim {
			continue
		}
		err := cb(ID(i))
		if err != nil {
			return err
		}
	}
	return nil
}

// edge_disk.go
package edge

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/viterin/vek/vek32"
)

const defaultVecsPerFile = 200000

type DiskBackend struct {
	dir          string
	metadata     diskMetadata
	quantization Quantization[Vector]

	// Store vectors in memory for simplicity
	vectors []Vector
	mu      sync.RWMutex

	token uint64
}

type diskMetadata struct {
	Dimensions   int    `json:"dimensions"`
	Quantization string `json:"quantization"`
	VecsPerFile  int    `json:"vecs_per_file"`
	VecFiles     []int  `json:"vec_files"`
}

var _ scannableBackend = &DiskBackend{}
var _ VectorBackend = &DiskBackend{}

func NewDiskBackend(directory string, dimensions int, quantization Quantization[Vector]) (*DiskBackend, error) {
	be := &DiskBackend{
		dir: directory,
		metadata: diskMetadata{
			Dimensions:   dimensions,
			Quantization: quantization.Name(),
			VecsPerFile:  defaultVecsPerFile,
		},
		quantization: quantization,
		vectors:      make([]Vector, 0),
	}
	err := be.loadOrInitialize()
	if err != nil {
		return nil, err
	}
	return be, nil
}

func (d *DiskBackend) loadOrInitialize() error {
	err := os.MkdirAll(d.dir, 0755)
	if err != nil {
		return err
	}
	vectorsPath := filepath.Join(d.dir, "vectors.gz")
	if _, err := os.Stat(vectorsPath); errors.Is(err, os.ErrNotExist) {
		// No vectors yet
		return nil
	}
	f, err := os.Open(vectorsPath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	decoder := json.NewDecoder(gz)
	var loadedVectors []Vector
	err = decoder.Decode(&loadedVectors)
	if err != nil {
		return err
	}
	d.vectors = loadedVectors
	return nil
}

func (d *DiskBackend) Close() error {
	return d.Sync()
}

func (d *DiskBackend) Sync() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	vectorsPath := filepath.Join(d.dir, "vectors.gz")
	f, err := os.Create(vectorsPath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	defer gz.Close()
	encoder := json.NewEncoder(gz)
	return encoder.Encode(d.vectors)
}

func (d *DiskBackend) PutVector(id ID, v Vector) error {
	if len(v) != d.metadata.Dimensions {
		return fmt.Errorf("DiskBackend: vector dimension mismatch, expected %d got %d", d.metadata.Dimensions, len(v))
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	idx := int(id)
	if idx < len(d.vectors) {
		d.vectors[idx] = v
	} else if idx == len(d.vectors) {
		d.vectors = append(d.vectors, v)
	} else {
		// Fill the gap with zero vectors
		for len(d.vectors) < idx {
			d.vectors = append(d.vectors, make(Vector, d.metadata.Dimensions))
		}
		d.vectors = append(d.vectors, v)
	}
	return nil
}

func (d *DiskBackend) ComputeSimilarity(targetVector Vector, targetID ID) (float32, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if int(targetID) >= len(d.vectors) {
		return 0, ErrIDNotFound
	}
	target := d.vectors[targetID]
	if target == nil || len(target) != d.metadata.Dimensions {
		return 0, ErrIDNotFound
	}
	return vek32.CosineSimilarity(targetVector, target), nil
}

func (d *DiskBackend) Info() BackendInfo {
	vectorsPath := filepath.Join(d.dir, "vectors.gz")
	exists := false
	if _, err := os.Stat(vectorsPath); err == nil {
		exists = true
	}
	return BackendInfo{
		HasIndexData: exists,
		Dimensions:   d.metadata.Dimensions,
		Quantization: d.quantization.Name(),
	}
}

func (d *DiskBackend) Exists(id ID) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	idx := int(id)
	if idx >= len(d.vectors) {
		return false
	}
	return d.vectors[idx] != nil && len(d.vectors[idx]) == d.metadata.Dimensions
}

func (d *DiskBackend) GetVector(id ID) (Vector, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	idx := int(id)
	if idx >= len(d.vectors) {
		return nil, ErrIDNotFound
	}
	v := d.vectors[idx]
	if v == nil || len(v) != d.metadata.Dimensions {
		return nil, ErrIDNotFound
	}
	return v, nil
}

func (d *DiskBackend) RemoveVector(id ID) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	idx := int(id)
	if idx >= len(d.vectors) {
		return ErrIDNotFound
	}
	if d.vectors[idx] == nil {
		return ErrIDNotFound
	}
	// Option 1: Set to nil
	d.vectors[idx] = nil
	return nil
}

func (d *DiskBackend) ForEachVector(cb func(ID) error) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	for i, v := range d.vectors {
		if v == nil || len(v) != d.metadata.Dimensions {
			continue
		}
		err := cb(ID(i))
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DiskBackend) SaveBases(bases []Basis, token uint64) (uint64, error) {
	// Not needed, removed IndexBackend
	return d.token, nil
}

func (d *DiskBackend) LoadBases() ([]Basis, error) {
	// Not needed, removed IndexBackend
	return nil, nil
}

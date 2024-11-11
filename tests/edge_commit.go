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

// edge_commit.go
package edge

import (
	"fmt"
)

const defaultMaxSampling = 10000

type PrintfFunc func(string, ...any)

type VectorStore struct {
	logger        PrintfFunc
	backend       VectorBackend
	dimensions    int
	nbasis        int
	bases         []Basis
	preSpill      int
	lastSaveToken uint64
}

type VectorStoreOption func(vs *VectorStore) error

func WithPrespill(prespill int) VectorStoreOption {
	return func(vs *VectorStore) error {
		if prespill <= 0 {
			prespill = 1
		} else if prespill > vs.dimensions {
			prespill = vs.dimensions
		}
		vs.preSpill = prespill
		return nil
	}
}

func NewVectorStore(backend VectorBackend, nBasis int, opts ...VectorStoreOption) (*VectorStore, error) {
	info := backend.Info()
	v := &VectorStore{
		dimensions: info.Dimensions,
		nbasis:     nBasis,
		backend:    backend,
		bases:      make([]Basis, nBasis),
		preSpill:   1,
	}
	for _, o := range opts {
		err := o(v)
		if err != nil {
			return nil, err
		}
	}
	if info.HasIndexData {
		// Previously used to load bitmaps, now no-op
		err := v.loadFromBackend()
		if err != nil {
			return v, err
		}
	} else {
		err := v.makeBasis()
		if err != nil {
			return nil, err
		}
		err = v.Sync()
		if err != nil {
			return nil, err
		}
	}
	return v, nil
}

func (vs *VectorStore) Close() error {
	err := vs.Sync()
	if err != nil {
		return err
	}
	return vs.backend.Close()
}

func (vs *VectorStore) SetLogger(printf PrintfFunc) {
	vs.logger = printf
}

func (vs *VectorStore) log(s string, a ...any) {
	if vs.logger != nil {
		vs.logger(s, a...)
	}
}

func (vs *VectorStore) AddVector(id ID, v Vector) error {
	if vs.backend.Exists(id) {
		err := vs.backend.RemoveVector(id)
		if err != nil {
			return fmt.Errorf("failed to remove existing vector with ID %d: %v", id, err)
		}
	}
	err := vs.backend.PutVector(id, v)
	if err != nil {
		return err
	}
	return nil
}

func (vs *VectorStore) RemoveVector(id ID) error {
	if !vs.backend.Exists(id) {
		return ErrIDNotFound
	}
	return vs.backend.RemoveVector(id)
}

func (vs *VectorStore) AddVectorsWithOffset(offset ID, vecs []Vector) error {
	for i, v := range vecs {
		id := offset + ID(i)
		if vs.backend.Exists(id) {
			//
			err := vs.backend.RemoveVector(id)
			if err != nil {
				return fmt.Errorf("failed to remove existing vector with ID %d: %v", id, err)
			}
		}
		err := vs.backend.PutVector(id, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (vs *VectorStore) AddVectorsWithIDs(ids []ID, vecs []Vector) error {
	for i, v := range vecs {
		id := ids[i]
		if vs.backend.Exists(id) {
			//
			err := vs.backend.RemoveVector(id)
			if err != nil {
				return fmt.Errorf("failed to remove existing vector with ID %d: %v", id, err)
			}
		}
		err := vs.backend.PutVector(id, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (vs *VectorStore) FindNearest(vector Vector, k int, searchk int, spill int) (*ResultSet, error) {
	return FullTableScanSearch(vs.backend, vector, k)
}

func (vs *VectorStore) Sync() error {
	if syncBe, ok := vs.backend.(interface{ Sync() error }); ok {
		return syncBe.Sync()
	}
	return nil
}

func (vs *VectorStore) makeBasis() error {
	vs.log("Making basis set")
	for n := 0; n < vs.nbasis; n++ {
		basis := make(Basis, vs.dimensions)
		for i := 0; i < vs.dimensions; i++ {
			basis[i] = NewRandVector(vs.dimensions, nil)
		}
		for range [10]struct{}{} { // Range 10 times
			orthonormalize(basis)
		}
		vs.log("Completed basis %d", n)
		vs.bases[n] = basis
	}
	vs.log("Completed basis set generation")
	return nil
}

func (vs *VectorStore) loadFromBackend() error {
	// Previously loaded bitmaps; now do nothing
	return nil
}

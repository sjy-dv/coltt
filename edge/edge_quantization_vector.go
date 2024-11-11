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

package edge

import (
	"fmt"
	"sync"

	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/gomath"
)

type EdgeVectorQ struct {
	dimension      int
	vectors        map[uint64]float16Vec
	collectionName string
	distance       distance.Space
	quantization   Float16Quantization
	lock           sync.RWMutex
}

type QuantizedEdgeVectors struct {
	Edges map[string]*EdgeVectorQ
	lock  sync.RWMutex
}

var quantizedEdgeV *QuantizedEdgeVectors

func NewQuantizedEdgeVectorCollection() {
	quantizedEdgeV = &QuantizedEdgeVectors{
		Edges: make(map[string]*EdgeVectorQ),
	}
}

func (qx *QuantizedEdgeVectors) CreateCollection(config CollectionConfig) error {
	qx.lock.RLock()
	_, ok := qx.Edges[config.collectionName]
	qx.lock.RUnlock()
	if ok {
		return fmt.Errorf(ErrCollectionExists, config.collectionName)
	}
	qx.lock.Lock()
	qx.Edges[config.collectionName].collectionName = config.collectionName
	qx.Edges[config.collectionName].dimension = config.dimension
	qx.Edges[config.collectionName].distance = func() distance.Space {
		if config.distance == COSINE {
			return distance.NewCosine()
		} else if config.distance == EUCLIDEAN {
			return distance.NewEuclidean()
		}
		return distance.NewCosine()
	}()
	qx.Edges[config.collectionName].quantization = Float16Quantization{}
	qx.Edges[config.collectionName].vectors = make(map[uint64]float16Vec)
	qx.lock.Unlock()
	return nil
}

func (qx *QuantizedEdgeVectors) DropCollection(collectionName string) error {
	qx.lock.RLock()
	_, ok := qx.Edges[collectionName]
	qx.lock.RUnlock()
	if !ok {
		return nil
	}
	qx.lock.Lock()
	delete(qx.Edges, collectionName)
	qx.lock.Unlock()
	return nil
}

func (qx *QuantizedEdgeVectors) InsertVector(collectionName string, commitId uint64, vector gomath.Vector) error {
	qx.lock.RLock()
	basis, ok := qx.Edges[collectionName]
	qx.lock.RUnlock()
	if !ok {
		return fmt.Errorf(ErrCollectionNotFound, collectionName)
	}
	lower, err := basis.quantization.Lower(vector)
	if err != nil {
		return fmt.Errorf(ErrQuantizedFailed, err)
	}
	basis.lock.Lock()
	basis.vectors[commitId] = lower
	basis.lock.Unlock()
	return nil
}

func (qx *QuantizedEdgeVectors) UpdateVector(collectionName string, id uint64, vector gomath.Vector) error {
	qx.lock.RLock()
	basis, ok := qx.Edges[collectionName]
	qx.lock.RUnlock()
	if !ok {
		return fmt.Errorf(ErrCollectionNotFound, collectionName)
	}
	lower, err := basis.quantization.Lower(vector)
	if err != nil {
		return fmt.Errorf(ErrQuantizedFailed, err)
	}
	basis.lock.Lock()
	basis.vectors[id] = lower
	basis.lock.Unlock()
	return nil
}

func (qx *QuantizedEdgeVectors) RemoveVector(collectionName string, id uint64) error {
	qx.lock.RLock()
	basis, ok := qx.Edges[collectionName]
	qx.lock.RUnlock()
	if !ok {
		return fmt.Errorf(ErrCollectionNotFound, collectionName)
	}
	basis.lock.Lock()
	delete(basis.vectors, id)
	basis.lock.Unlock()
	return nil
}

func (qx *QuantizedEdgeVectors) FullScan(collectionName string, target gomath.Vector, topK int,
) (*ResultSet, error) {
	rs := NewResultSet(topK)
	qx.lock.RLock()
	basis, ok := qx.Edges[collectionName]
	qx.lock.RUnlock()
	if !ok {
		return nil, fmt.Errorf(ErrCollectionNotFound, collectionName)
	}
	lower, err := basis.quantization.Lower(target)
	if err != nil {
		return nil, fmt.Errorf(ErrQuantizedFailed, err)
	}
	basis.lock.RLock()
	defer basis.lock.RUnlock()
	for index, qvec := range basis.vectors {
		sim := basis.quantization.Similarity(lower, qvec, basis.distance)
		rs.AddResult(ID(index), sim)
	}
	return rs, nil
}

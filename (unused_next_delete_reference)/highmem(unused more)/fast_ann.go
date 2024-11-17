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

package highmem

import (
	"errors"
	"fmt"
	"sync"
	// "github.com/sjy-dv/nnv/pkg/fasthnsw"
)

type Tensor struct {
	tensors    map[string]*fasthnsw.Index
	tensorLock sync.RWMutex
}

var tensorLinker *Tensor

func NewTensorLink() {
	tensorLinker = &Tensor{
		tensors: make(map[string]*fasthnsw.Index),
	}
}

func (xx *Tensor) existsTensor(collectionName string) bool {
	xx.tensorLock.RLock()
	_, exists := xx.tensors[collectionName]
	xx.tensorLock.RUnlock()
	return exists
}

func (xx *Tensor) getTensor(collectionName string) *fasthnsw.Index {
	xx.tensorLock.RLock()
	tensor, exists := xx.tensors[collectionName]
	xx.tensorLock.RUnlock()
	if exists {
		return tensor
	}
	return nil
}

func (xx *Tensor) CreateTensorIndex(collectionName string, cfg CollectionConfig) error {

	c := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- fmt.Errorf(panicr, r)
			}
		}()

		ok := xx.existsTensor(collectionName)
		if ok {
			c <- errors.New("already exists tensor")
			return
		}
		tcfg := fasthnsw.DefaultConfig(uint(cfg.Dim))
		tcfg.Connectivity = uint(cfg.Connectivity)
		tcfg.ExpansionAdd = uint(cfg.ExpansionAdd)
		tcfg.ExpansionSearch = uint(cfg.ExpansionSearch)
		tcfg.Multi = cfg.Multi
		if cfg.Quantization != "None" {
			tcfg.Quantization = func() fasthnsw.Quantization {
				switch cfg.Quantization {
				case "BF16":
					return fasthnsw.BF16
				case "F16":
					return fasthnsw.F16
				case "F32":
					return fasthnsw.F32
				case "F64":
					return fasthnsw.F64
				case "I8":
					return fasthnsw.I8
				case "B1":
					return fasthnsw.B1
				}
				return fasthnsw.F16
			}()
		}
		tcfg.Metric = func() fasthnsw.Metric {
			switch cfg.Distance {
			case "InnerProduct":
				return fasthnsw.InnerProduct
			case "Cosine":
				return fasthnsw.Cosine
			case "Haversine":
				return fasthnsw.Haversine
			case "Divergence":
				return fasthnsw.Divergence
			case "Pearson":
				return fasthnsw.Pearson
			case "Hamming":
				return fasthnsw.Hamming
			case "Tanimoto":
				return fasthnsw.Tanimoto
			case "Sorensen":
				return fasthnsw.Sorensen
			case "L2sq":
				return fasthnsw.L2sq
			}
			return fasthnsw.Cosine
		}()
		newTensor, err := fasthnsw.NewIndex(tcfg)
		if err != nil {
			c <- err
			return
		}
		xx.tensorLock.Lock()
		xx.tensors[collectionName] = newTensor
		xx.tensorLock.Unlock()
		c <- nil
	}()
	return <-c
}

func (xx *Tensor) DropTensorIndex(collectionName string) error {
	c := make(chan error)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- fmt.Errorf(panicr, r)
			}
		}()
		ok := xx.existsTensor(collectionName)
		if !ok {
			c <- nil
			return
		}
		xx.tensorLock.Lock()
		delete(xx.tensors, collectionName)
		xx.tensorLock.Unlock()
		c <- nil
	}()
	return <-c
}

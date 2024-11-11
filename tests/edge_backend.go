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

// edge_backend.go
package edge

import (
	"errors"
)

type VectorBackend interface {
	PutVector(id ID, v Vector) error
	ComputeSimilarity(targetVector Vector, targetID ID) (float32, error)
	RemoveVector(id ID) error
	Info() BackendInfo
	Exists(id ID) bool
	Close() error
}

type scannableBackend interface {
	VectorBackend
	ForEachVector(func(ID) error) error
}

type VectorGetter[T any] interface {
	GetVector(id ID) (T, error)
}

type BackendInfo struct {
	HasIndexData bool
	Dimensions   int
	Quantization string
}

func FullTableScanSearch(be VectorBackend, target Vector, k int) (*ResultSet, error) {
	rs := NewResultSet(k)
	b, ok := be.(scannableBackend)
	if !ok {
		return nil, errors.New("Backend is incompatible")
	}
	err := b.ForEachVector(func(id ID) error {
		sim, err := b.ComputeSimilarity(target, id)
		if err != nil {
			return err
		}
		rs.AddResult(id, sim)
		return nil
	})
	return rs, err
}

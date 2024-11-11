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

package index

import (
	"time"

	"github.com/sjy-dv/nnv/backup/store"
)

type Type int

const (
	SingleField Type = iota
)

type Info struct {
	Field string
	Type  Type
}

type Index interface {
	Add(docId string, v interface{}, ttl time.Duration) error
	Remove(docId string, v interface{}) error
	Iterate(reverse bool, onValue func(docId string) error) error
	Drop() error
	Type() Type
	Collection() string
	Field() string
}

type indexBase struct {
	collection, field string
}

func (idx *indexBase) Collection() string {
	return idx.collection
}

func (idx *indexBase) Field() string {
	return idx.field
}

type Query interface {
	Run(onValue func(docId string) error) error
}

func CreateIndex(collection, field string, idxType Type, tx store.Tx) Index {
	indexBase := indexBase{collection: collection, field: field}
	switch idxType {
	case SingleField:
		return &rangeIndex{
			indexBase: indexBase,
			tx:        tx,
		}
	}
	return nil
}

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

import "github.com/sjy-dv/coltt/pkg/wal"

type Indexer interface {
	Put(key []byte, position *wal.ChunkPosition) *wal.ChunkPosition

	Get(key []byte) *wal.ChunkPosition

	Delete(key []byte) (*wal.ChunkPosition, bool)

	Size() int

	Ascend(handleFn func(key []byte, position *wal.ChunkPosition) (bool, error))

	AscendRange(startKey, endKey []byte, handleFn func(key []byte, position *wal.ChunkPosition) (bool, error))

	AscendGreaterOrEqual(key []byte, handleFn func(key []byte, position *wal.ChunkPosition) (bool, error))

	Descend(handleFn func(key []byte, pos *wal.ChunkPosition) (bool, error))

	DescendRange(startKey, endKey []byte, handleFn func(key []byte, position *wal.ChunkPosition) (bool, error))

	DescendLessOrEqual(key []byte, handleFn func(key []byte, position *wal.ChunkPosition) (bool, error))
}

type IndexerType = byte

const (
	BTree IndexerType = iota
)

// Change the index type as you implement.
var indexType = BTree

func NewIndexer() Indexer {
	switch indexType {
	case BTree:
		return newBTree()
	default:
		panic("unexpected index type")
	}
}

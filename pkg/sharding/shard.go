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

package sharding

import (
	"encoding/binary"
	"hash/fnv"

	"github.com/google/uuid"
)

// c values .. server count or data shard count
func ShardTraffic(x uuid.UUID, c uint64) uint64 {
	res := ((binary.LittleEndian.Uint64(x[:8]) % c) +
		(binary.BigEndian.Uint64(x[8:]) % c))
	return res % c
}

func ShardVertex(x uint64, c uint64) uint64 {
	hasher := fnv.New64a()
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, x)
	hasher.Write(buf)
	hashValue := hasher.Sum64()
	return hashValue % c
}

func ShardVertexV2(s string, c uint64) uint64 {
	hasher := fnv.New64a()
	hasher.Write([]byte(s))
	hashValue := hasher.Sum64()
	return hashValue % c
}

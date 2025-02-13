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
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/sjy-dv/coltt/pkg/snowflake"
	"github.com/stretchr/testify/assert"
)

func TestShardTraffic(t *testing.T) {
	var c uint64 = 3
	res := make(map[uint64]int)
	for i := 0; i < 999; i++ {
		r := ShardTraffic(uuid.New(), c)
		res[r] += 1
	}
	fmt.Println(res)
	t.Log(res)
}

func TestShardVertex(t *testing.T) {
	gen, _ := snowflake.NewNode(0)
	res := make(map[uint64]int)

	for range 10000 {
		x := uint64(gen.Generate())
		slice := ShardVertex(x, 16)
		res[slice] += 1
	}
	fmt.Println(res)
}

func TestShardVertexAlwaysSame(t *testing.T) {
	dummyValue := 128545215

	staticVal := ShardVertex(uint64(dummyValue), 16)

	for range 10000 {
		//always same
		assert.Equal(t, staticVal, ShardVertex(uint64(dummyValue), 16))
	}
}

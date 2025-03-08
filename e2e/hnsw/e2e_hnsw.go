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

package main

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/sjy-dv/coltt/core/vectorindex"
	"github.com/sjy-dv/coltt/pkg/distance"
	"github.com/sjy-dv/coltt/pkg/gomath"
)

func main() {
	index := generateRandomIndex(128, 1000, distance.NewCosine())
	query := gomath.RandomUniformVector(128)
	c1, err := index.Search(context.Background(), query, 10)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("==========c1 search================")
	for _, c := range c1 {
		fmt.Println(c.Id, c.Metadata, c.Score)
	}
	var buf bytes.Buffer
	err = index.Commit(&buf, true)
	if err != nil {
		log.Fatal(err)
	}
	copyindex := vectorindex.NewHnsw(128, distance.NewCosine())
	err = copyindex.Load(&buf, true)
	if err != nil {
		log.Fatal(err)
	}
	c2, err := copyindex.Search(context.Background(), query, 10)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("=========c2 search===========")
	for _, c := range c2 {
		fmt.Println(c.Id, c.Metadata, c.Score)
	}
}

func generateRandomIndex(dim, size int, dist distance.Space) *vectorindex.Hnsw {
	insertKeys := make(map[uint64]struct{})

	index := vectorindex.NewHnsw(uint(dim), dist)
	delOffset := int(size / 10)
	for i := 0; i < size; i++ {
		if i > delOffset && (gomath.RandomUniform() <= 0.2) {
			var key uint64
			for k := range insertKeys {
				key = k
				break
			}
			delete(insertKeys, key)
			index.Remove(key)
		} else {
			id := uint64(i)
			insertKeys[id] = struct{}{}
			index.Insert(id, gomath.RandomUniformVector(dim), vectorindex.Metadata{"foo": fmt.Sprintf("bar: %d", i), "id": fmt.Sprintf("%d", id)}, index.RandomLevel())
		}
	}
	return index
}

// ==========c1 search================
// 52 map[foo:bar: 52 id:52] 0.17748153
// 410 map[foo:bar: 410 id:410] 0.18120253
// 808 map[foo:bar: 808 id:808] 0.18551314
// 259 map[foo:bar: 259 id:259] 0.1857208
// 614 map[foo:bar: 614 id:614] 0.18677711
// 429 map[foo:bar: 429 id:429] 0.18750209
// 938 map[foo:bar: 938 id:938] 0.18993133
// 991 map[foo:bar: 991 id:991] 0.1916337
// 941 map[foo:bar: 941 id:941] 0.19311512
// 727 map[foo:bar: 727 id:727] 0.19370127
// =========c2 search===========
// 52 map[foo:bar: 52 id:52] 0.17748153
// 410 map[foo:bar: 410 id:410] 0.18120253
// 808 map[foo:bar: 808 id:808] 0.18551314
// 259 map[foo:bar: 259 id:259] 0.1857208
// 614 map[foo:bar: 614 id:614] 0.18677711
// 429 map[foo:bar: 429 id:429] 0.18750209
// 938 map[foo:bar: 938 id:938] 0.18993133
// 991 map[foo:bar: 991 id:991] 0.1916337
// 941 map[foo:bar: 941 id:941] 0.19311512
// 727 map[foo:bar: 727 id:727] 0.19370127

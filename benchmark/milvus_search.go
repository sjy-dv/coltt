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
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

func main() {
	collectionName := "benchmark_flat"
	client, err := client.NewClient(context.Background(), client.Config{
		Address: ":19530",
	})
	if err != nil {
		log.Fatal(err)
	}
	has, err := client.HasCollection(context.Background(), collectionName)
	fmt.Println(has, err)
	idx, err := entity.NewIndexFlat(entity.L2)
	if err != nil {
		log.Fatal(err)
	}
	c, _ := client.GetCollectionStatistics(context.Background(), collectionName)
	fmt.Println(c)
	err = client.CreateIndex(context.Background(), collectionName, "embeddings", idx, false)
	fmt.Println(err)
	rt := time.Now()
	err = client.LoadCollection(context.Background(), collectionName, false)
	if err != nil {
		log.Fatal(err)
	}
	ot := time.Since(rt)
	fmt.Println("release time : ", ot.Seconds())
	timeChan := make(chan time.Duration, 100)
	for i := 0; i < 100; i++ {
		sp, _ := entity.NewIndexFlatSearchParam()
		start := time.Now()
		//	defer wg.Done()
		_, err := client.Search(context.Background(), collectionName, nil, "",
			[]string{"ID"}, []entity.Vector{
				entity.FloatVector(generateRandomVector(128)),
			}, "embeddings",
			entity.L2, 10, sp)
		if err != nil {
			log.Fatal(err.Error())
		}
		elapsed := time.Since(start)
		timeChan <- elapsed

	}
	close(timeChan)
	var totalTime time.Duration
	for t := range timeChan {
		totalTime += t
	}
	fmt.Printf("search average time : %.2f", (totalTime.Seconds() / 100))
}

func generateRandomVector(dim int) []float32 {
	vec := make([]float32, dim)
	for i := 0; i < dim; i++ {
		vec[i] = rand.Float32()
	}
	return vec
}

// release time :  7.8202766
// search average time : 0.02

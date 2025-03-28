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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"playground.benchmark.coltt/edgeproto"
)

func main() {
	collectionName := "benchmark_flat"
	conn, err := grpc.Dial(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	dclient := edgeproto.NewEdgeRpcClient(conn)
	rt := time.Now()
	resp, err := dclient.LoadCollection(context.Background(), &edgeproto.CollectionName{
		CollectionName: collectionName,
	})
	if !resp.Status || err != nil {
		log.Fatal(resp.Error.ErrorMessage, err.Error())
	}
	ot := time.Since(rt)
	fmt.Println("release time : ", ot.Seconds())
	timeChan := make(chan time.Duration, 100)
	// var wg sync.WaitGroup
	// wg.Add(1000)
	for i := 0; i < 100; i++ {
		start := time.Now()
		//	defer wg.Done()
		resp, err := dclient.VectorSearch(context.Background(), &edgeproto.SearchReq{
			CollectionName:        collectionName,
			Vector:                generateRandomVector(128),
			TopK:                  10,
			HighResourceAvaliable: true,
		})
		if !resp.Status || err != nil {
			log.Fatal(resp.Error.ErrorMessage, err.Error())
		}
		elapsed := time.Since(start)
		timeChan <- elapsed

	}
	//wg.Wait()
	close(timeChan)
	var totalTime time.Duration
	for t := range timeChan {
		totalTime += t
	}
	fmt.Println(totalTime.Nanoseconds() / 100)
	fmt.Printf("search average time : %d", (totalTime.Milliseconds() / 100))
	fmt.Printf("search average time : %.2f", (totalTime.Seconds() / 100))
}

func generateTestVectors(num, dim int) [][]float32 {
	vectors := make([][]float32, num)
	for i := 0; i < num; i++ {
		vectors[i] = generateRandomVector(dim)
	}
	return vectors
}

func generateRandomVector(dim int) []float32 {
	vec := make([]float32, dim)
	for i := 0; i < dim; i++ {
		vec[i] = rand.Float32()
	}
	return vec
}

// release time :  0.004064
// search average time : 0.34

//0.87 ms / 0.00087 sec

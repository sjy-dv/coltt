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
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"playground.benchmark.coltt/edgeproto"
)

func main() {
	collectionName := "nil_collect"
	conn, err := grpc.Dial(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	dclient := edgeproto.NewEdgeRpcClient(conn)

	_, err = dclient.CreateCollection(context.Background(), &edgeproto.Collection{
		CollectionName: collectionName,
		Dim:            128,
		Distance:       edgeproto.Distance_Euclidean,
		Quantization:   edgeproto.Quantization_None,
	})

	if err != nil {
		log.Fatal(err)
	}

	_, err = dclient.Insert(context.Background(), &edgeproto.ModifyDataset{
		Id:             strconv.Itoa(0),
		Vector:         []float32{},
		CollectionName: collectionName,
	})
	if err != nil {
		log.Fatal(err)
	}
	resp, err := dclient.Flush(context.Background(), &edgeproto.CollectionName{CollectionName: collectionName})
	if !resp.Status || err != nil {
		log.Fatal(resp.Error.ErrorMessage, err)
	}
	resp2, err := dclient.VectorSearch(context.Background(), &edgeproto.SearchReq{
		CollectionName: collectionName,
		Vector:         generateRandomVector(128),
		TopK:           10,
	})
	fmt.Println(resp2, err)
	if !resp2.Status || err != nil {
		log.Fatal(resp.Error.ErrorMessage, err.Error())
	}
}

func generateRandomVector(dim int) []float32 {
	vec := make([]float32, dim)
	for i := 0; i < dim; i++ {
		vec[i] = rand.Float32()
	}
	return vec
}

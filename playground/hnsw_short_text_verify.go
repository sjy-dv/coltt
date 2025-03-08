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
	"container/heap"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/sjy-dv/coltt/edge"
	"github.com/sjy-dv/coltt/pkg/hnswpq"
	"github.com/sjy-dv/coltt/pkg/models"
	"github.com/sjy-dv/coltt/pkg/queue"
)

type JsonReview struct {
	Review    string    `json:"review"`
	Embedding []float32 `json:"embedding"`
}

type ResultCompare struct {
	BaseReview    string   `json:"base_review"`
	SimilarReview []Review `json:"similar_review"`
	Latency       string   `json:"latency"`
}

type Review struct {
	Review   string  `json:"review"`
	Distance float32 `json:"distance"`
}

func Normalize(v []float32) []float32 {
	var norm float32
	out := make([]float32, len(v))
	for i := range v {
		norm += v[i] * v[i]
	}
	if norm == 0 {
		return out
	}

	norm = float32(math.Sqrt(float64(norm)))
	for i := range v {
		out[i] = v[i] / norm
	}

	return out
}

func normalise(vec []float32) []float32 {
	// var magnitude float32 = 0.0
	// for _, v := range vec {
	// 	magnitude += v * v
	// }
	// magnitude = float32(math.Sqrt(float64(magnitude)))
	// for i, v := range vec {
	// 	vec[i] = v / magnitude
	// }
	return vec
}

func main() {

	jsonf, err := os.ReadFile("short_text.json")
	if err != nil {
		log.Fatal(1, err)
	}
	jrs := make([]JsonReview, 0, 1_100)
	err = json.Unmarshal(jsonf, &jrs)
	if err != nil {
		log.Fatal(2, err)
	}
	fmt.Println("dataset is ready >> data length : ", len(jrs))

	qf, err := os.ReadFile("review_question.json")
	prepareQ := make([]JsonReview, 0, 2)
	err = json.Unmarshal(qf, &prepareQ)
	if err != nil {
		log.Fatal(3, err)
	}

	collection := "review_collection"

	vectorLen := 384
	pqParams := models.ProductQuantizerParameters{
		NumSubVectors:    192,
		NumCentroids:     256,
		TriggerThreshold: 100,
	}

	cfg := models.HnswConfig{
		Efconstruction: 200,
		M:              16,
		Mmax:           32,
		Mmax0:          32,
		Ml:             1 / math.Log(1.0*float64(16)),
		Ep:             0,
		MaxLevel:       0,
		Dim:            uint32(vectorLen),
		DistanceType:   edge.COSINE,
		Heuristic:      false,
		BucketName:     collection,
	}

	hnswpq.NewOffsetCounter(0)
	hnswPQ := hnswpq.NewProductQuantizationHnsw()
	err = hnswPQ.CreateCollection(collection, cfg, pqParams)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("initial hnsw ok")
	fmt.Println("start pretrained pq")
	err = hnswPQ.Collections[collection].PQ.
		PreTrainProductQuantizer(collection, vectorLen, 10_000)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("PQ pretrained clear")

	fmt.Println("create genesis node with pre-trained pq")
	hnswPQ.Genesis(collection, cfg)
	fmt.Println("genesis-node clear")

	// save compare review
	for i, data := range jrs {
		commitId := hnswpq.NextId()

		err := hnswPQ.Insert(collection, commitId, Normalize(data.Embedding))
		if err != nil {
			log.Fatalf("insert error: %v", err)
		}
		if i%100 == 0 {
			fmt.Printf("Inserted %d vectors...\n", i)
		}
	}

	// check for pretrained search result
	saveJson := make([]ResultCompare, 0)
	for _, data := range prepareQ {
		start := time.Now()
		topCandidates := &queue.PriorityQueue{Order: false, Items: []*queue.Item{}}
		heap.Init(topCandidates)
		err := hnswPQ.Search(collection, Normalize(data.Embedding), topCandidates, 5, 16)
		if err != nil {
			log.Fatalf("pq search error: %v", err)
		}
		elapsed := time.Since(start)
		resultset := ResultCompare{}
		resultset.BaseReview = data.Review
		resultset.Latency = fmt.Sprintf("latency: %d ms", elapsed.Milliseconds())
		for _, out := range topCandidates.Items {
			resultset.SimilarReview = append(resultset.SimilarReview, Review{
				Review:   jrs[out.NodeID-1].Review,
				Distance: out.Distance,
			})
		}
		saveJson = append(saveJson, resultset)
	}
	w, err := os.Create("norm-pre-trained-short-text-verification.json")
	if err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(w)
	enc.Encode(saveJson)
	w.Close()

	fmt.Println("fit all vectors new kmeans cluster")
	err = hnswPQ.Collections[collection].PQ.Fit()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("clear fit pq")

	fmt.Println("copy fit new migrate hnsw")
	hnswpq.NewOffsetCounter(0)
	fitPQ := hnswpq.NewProductQuantizationHnsw()
	err = fitPQ.CreateCollection(collection, cfg, pqParams)
	if err != nil {
		log.Fatal(err)
	}
	fitPQ.Collections[collection].PQ = hnswPQ.Collections[collection].PQ
	fitPQ.Genesis(collection, cfg)

	fmt.Println("new fit insert point")
	for i, data := range jrs {
		commitId := hnswpq.NextId()

		err := fitPQ.Insert(collection, commitId, Normalize(data.Embedding))
		if err != nil {
			log.Fatalf("insert error: %v", err)
		}
		if i%100 == 0 {
			fmt.Printf("Inserted %d vectors...\n", i)
		}
	}
	fmt.Println("after fit search")
	saveJson = make([]ResultCompare, 0)
	for _, data := range prepareQ {
		start := time.Now()
		topCandidates := &queue.PriorityQueue{Order: false, Items: []*queue.Item{}}
		heap.Init(topCandidates)
		err := fitPQ.Search(collection, Normalize(data.Embedding), topCandidates, 5, 16)
		if err != nil {
			log.Fatalf("pq search error: %v", err)
		}
		elapsed := time.Since(start)
		resultset := ResultCompare{}
		resultset.BaseReview = data.Review
		resultset.Latency = fmt.Sprintf("latency: %d ms", elapsed.Milliseconds())
		for _, out := range topCandidates.Items {
			resultset.SimilarReview = append(resultset.SimilarReview, Review{
				Review:   jrs[out.NodeID-1].Review,
				Distance: out.Distance,
			})
		}
		saveJson = append(saveJson, resultset)
	}
	w, err = os.Create("norm-all-vector-fit-short-text-verification.json")
	if err != nil {
		log.Fatal(err)
	}
	enc = json.NewEncoder(w)
	enc.Encode(saveJson)
	w.Close()

	fmt.Println("Just in case, initialize all normal vectors")
	for i := range fitPQ.Collections[collection].NodeList.Nodes {
		fitPQ.Collections[collection].NodeList.Nodes[i].Vectors = nil
	}
	fmt.Println("We will conduct a final verification to see if it can be searched using pq.")
	saveJson = make([]ResultCompare, 0)
	for _, data := range prepareQ {
		start := time.Now()
		topCandidates := &queue.PriorityQueue{Order: false, Items: []*queue.Item{}}
		heap.Init(topCandidates)
		err := fitPQ.Search(collection, Normalize(data.Embedding), topCandidates, 5, 16)
		if err != nil {
			log.Fatalf("pq search error: %v", err)
		}
		elapsed := time.Since(start)
		resultset := ResultCompare{}
		resultset.BaseReview = data.Review
		resultset.Latency = fmt.Sprintf("latency: %d ms", elapsed.Milliseconds())
		for _, out := range topCandidates.Items {
			resultset.SimilarReview = append(resultset.SimilarReview, Review{
				Review:   jrs[out.NodeID-1].Review,
				Distance: out.Distance,
			})
		}
		saveJson = append(saveJson, resultset)
	}
	w, err = os.Create("norm-vector-nil-clear-fit-short-text-verification.json")
	if err != nil {
		log.Fatal(err)
	}
	enc = json.NewEncoder(w)
	enc.Encode(saveJson)
	w.Close()
}

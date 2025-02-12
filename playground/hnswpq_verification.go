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

// use dataset.csv => https://www.kaggle.com/datasets/lakshmi25npathi/imdb-dataset-of-50k-movie-reviews
// dataset process => data_process.py

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
	Review string `json:"review"`
}

func main() {

	jsonf, err := os.ReadFile("dataset.json")
	if err != nil {
		log.Fatal(err)
	}
	jrs := make([]JsonReview, 0, 50_000)
	err = json.Unmarshal(jsonf, &jrs)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("dataset is ready >> data length : ", len(jrs))

	collection := "review_collection"

	vectorLen := 384
	pqParams := models.ProductQuantizerParameters{
		NumSubVectors:    32,
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
		PreTrainProductQuantizer(collection, vectorLen, 20_000)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("PQ pretrained clear")

	fmt.Println("create genesis node with pre-trained pq")
	hnswPQ.Genesis(collection, cfg)
	fmt.Println("genesis-node clear")

	// save compare review
	compareReview := make(map[uint64]JsonReview)
	for i, data := range jrs {
		commitId := hnswpq.NextId()
		if i == 500 || i == 1500 || i == 15000 || i == 25000 || i == 43000 {
			compareReview[commitId] = data
		}
		err := hnswPQ.Insert(collection, commitId, data.Embedding)
		if err != nil {
			log.Fatalf("insert error: %v", err)
		}
		if i%100 == 0 {
			fmt.Printf("Inserted %d vectors...\n", i)
		}
	}

	// check for pretrained search result
	saveJson := make([]ResultCompare, 0)
	for _, data := range compareReview {
		start := time.Now()
		topCandidates := &queue.PriorityQueue{Order: false, Items: []*queue.Item{}}
		heap.Init(topCandidates)
		err := hnswPQ.Search(collection, data.Embedding, topCandidates, 5, 200)
		if err != nil {
			log.Fatalf("pq search error: %v", err)
		}
		elapsed := time.Since(start)
		resultset := ResultCompare{}
		resultset.BaseReview = data.Review
		resultset.Latency = fmt.Sprintf("latency: %d ms", elapsed.Milliseconds())
		for _, out := range topCandidates.Items {
			resultset.SimilarReview = append(resultset.SimilarReview, Review{
				Review: jrs[out.NodeID-1].Review,
			})
		}
		saveJson = append(saveJson, resultset)
	}
	w, err := os.Create("pre-trained-verification.json")
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

	fmt.Println("after fit search")
	saveJson = make([]ResultCompare, 0)
	for _, data := range compareReview {
		start := time.Now()
		topCandidates := &queue.PriorityQueue{Order: false, Items: []*queue.Item{}}
		heap.Init(topCandidates)
		err := hnswPQ.Search(collection, data.Embedding, topCandidates, 5, 200)
		if err != nil {
			log.Fatalf("pq search error: %v", err)
		}
		elapsed := time.Since(start)
		resultset := ResultCompare{}
		resultset.BaseReview = data.Review
		resultset.Latency = fmt.Sprintf("latency: %d ms", elapsed.Milliseconds())
		for _, out := range topCandidates.Items {
			resultset.SimilarReview = append(resultset.SimilarReview, Review{
				Review: jrs[out.NodeID-1].Review,
			})
		}
		saveJson = append(saveJson, resultset)
	}
	w, err = os.Create("all-vector-fit-verification.json")
	if err != nil {
		log.Fatal(err)
	}
	enc = json.NewEncoder(w)
	enc.Encode(saveJson)
	w.Close()

	fmt.Println("Just in case, initialize all normal vectors")
	for i := range hnswPQ.Collections[collection].NodeList.Nodes {
		hnswPQ.Collections[collection].NodeList.Nodes[i].Vectors = nil
	}
	fmt.Println("We will conduct a final verification to see if it can be searched using pq.")
	saveJson = make([]ResultCompare, 0)
	for _, data := range compareReview {
		start := time.Now()
		topCandidates := &queue.PriorityQueue{Order: false, Items: []*queue.Item{}}
		heap.Init(topCandidates)
		err := hnswPQ.Search(collection, data.Embedding, topCandidates, 5, 200)
		if err != nil {
			log.Fatalf("pq search error: %v", err)
		}
		elapsed := time.Since(start)
		resultset := ResultCompare{}
		resultset.BaseReview = data.Review
		resultset.Latency = fmt.Sprintf("latency: %d ms", elapsed.Milliseconds())
		for _, out := range topCandidates.Items {
			resultset.SimilarReview = append(resultset.SimilarReview, Review{
				Review: jrs[out.NodeID-1].Review,
			})
		}
		saveJson = append(saveJson, resultset)
	}
	w, err = os.Create("vector-nil-clear-fit-verification.json")
	if err != nil {
		log.Fatal(err)
	}
	enc = json.NewEncoder(w)
	enc.Encode(saveJson)
	w.Close()
}

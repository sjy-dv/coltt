package main

import (
	"container/heap"
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/sjy-dv/nnv/edge"
	"github.com/sjy-dv/nnv/pkg/hnsw"
	"github.com/sjy-dv/nnv/pkg/hnswpq"
)

func main() {

	collection := "test_col"

	vectorLen := 384

	pqParams := hnswpq.ProductQuantizerParameters{
		NumSubVectors:    16,
		NumCentroids:     256,
		TriggerThreshold: 1000,
	}

	cfg := hnsw.HnswConfig{
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
		EmptyNodes:     make([]uint32, 0),
	}
	hnswpq.NewOffsetCounter(0)
	hnswPQ := hnswpq.NewProductQuantizationHnsw()
	err := hnswPQ.CreateCollection(collection, cfg, pqParams)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("initial hnsw ok")

	//pretrained pq
	fmt.Println("start pretrained pq")
	err = hnswPQ.Collections[collection].PQ.PreTrainProductQuantizer(collection, vectorLen, 1000)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("PQ pretrained clear")
	hnswPQ.Genesis(collection, cfg)
	numVecs := 10_000
	fmt.Println("insert 10_000 vectors start")

	// save vector for result
	saveVector := make(map[uint64][]float32)
	for i := 1; i < numVecs; i++ {
		vec := genRandVec(vectorLen)
		commitId := hnswpq.NextId()
		if i == 500 || i == 1500 || i == 5000 {
			saveVector[commitId] = vec
		}
		err := hnswPQ.Insert(collection, commitId, vec)
		if err != nil {
			log.Fatalf("insert error : %v", err)
		}
		if i%1000 == 0 {
			fmt.Printf("Inserted %d vectors...\n", i)
		}
	}

	fmt.Println("vector insert Complete")
	topK := 1
	efSearch := 200

	// check for original pq search
	for uid, vec := range saveVector {
		topCandidates := &hnswpq.PriorityQueue{Order: false, Items: []*hnswpq.Item{}}
		heap.Init(topCandidates)
		err := hnswPQ.Search(collection, vec, topCandidates, topK, efSearch)
		if err != nil {
			log.Fatalf("pq search error: %v", err)
		}
		gets := topCandidates.Items[0]
		fmt.Printf("is same? expect: %d, got: %d\n", uid, gets.NodeID)
	}

	// fitting pq
	fmt.Println("fit all vectors product-quantization")
	err = hnswPQ.Collections[collection].PQ.Fit()
	if err != nil {
		fmt.Println("pq fit error : ", err.Error())
	}
	fmt.Println("clear fit pq")

	fmt.Println("after fit search going")
	for uid, vec := range saveVector {
		topCandidates := &hnswpq.PriorityQueue{Order: false, Items: []*hnswpq.Item{}}
		heap.Init(topCandidates)
		err := hnswPQ.Search(collection, vec, topCandidates, topK, efSearch)
		if err != nil {
			log.Fatalf("pq search error: %v", err)
		}
		gets := topCandidates.Items[0]
		fmt.Printf("is same? expect: %d, got: %d\n", uid, gets.NodeID)
	}
	fmt.Println("Just in case, initialize all normal vectors")
	for i := range hnswPQ.Collections[collection].NodeList.Nodes {
		hnswPQ.Collections[collection].NodeList.Nodes[i].Vectors = nil
	}
	fmt.Println("We will conduct a final verification to see if it can be searched using pq.")
	for uid, vec := range saveVector {
		topCandidates := &hnswpq.PriorityQueue{Order: false, Items: []*hnswpq.Item{}}
		heap.Init(topCandidates)
		err := hnswPQ.Search(collection, vec, topCandidates, topK, efSearch)
		if err != nil {
			log.Fatalf("pq search error: %v", err)
		}
		gets := topCandidates.Items[0]
		fmt.Printf("is same? expect: %d, got: %d\n", uid, gets.NodeID)
	}
}

func genRandVec(dim int) []float32 {
	vec := make([]float32, dim)
	for i := 0; i < dim; i++ {
		vec[i] = rand.Float32()
	}
	return vec
}

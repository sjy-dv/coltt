package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sjy-dv/nnv/core/vectorindex"
	"github.com/sjy-dv/nnv/edge"
	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/snowflake"
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

	vectorLen := 384

	nhnsw := vectorindex.NewHnsw(uint(vectorLen),
		distance.NewCosine(),
		vectorindex.HnswSearchAlgorithm(vectorindex.HnswSearchHeuristic))

	idgen, _ := snowflake.NewNode(0)

	for i, data := range jrs {
		rm := map[string]interface{}{
			"review": data.Review,
		}
		err = nhnsw.Insert(uint64(idgen.Generate()),
			edge.Vector(data.Embedding), rm, nhnsw.RandomLevel())

		if i%100 == 0 {
			fmt.Printf("Inserted %d vectors...\n", i)
		}
	}
	saveJson := make([]ResultCompare, 0)
	for _, data := range prepareQ {
		start := time.Now()
		result, err := nhnsw.Search(context.Background(), edge.Vector(data.Embedding), 5)
		if err != nil {
			log.Fatal(err)
		}
		elapsed := time.Since(start)
		resultset := ResultCompare{}
		resultset.BaseReview = data.Review
		resultset.Latency = fmt.Sprintf("latency: %d ms", elapsed.Milliseconds())
		for _, out := range result {
			resultset.SimilarReview = append(resultset.SimilarReview, Review{
				Review:   out.Metadata["review"].(string),
				Distance: out.Score,
			})
		}
		saveJson = append(saveJson, resultset)
	}
	w, err := os.Create("advanced_hnsw_test.json")
	if err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(w)
	enc.Encode(saveJson)
	w.Close()
}

// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"math/rand/v2"
// 	"strconv"
// 	"time"

// 	"github.com/milvus-io/milvus-sdk-go/v2/client"
// 	"github.com/milvus-io/milvus-sdk-go/v2/entity"
// 	// "github.com/milvus-io/milvus-sdk-go/v2/client"
// 	// "github.com/milvus-io/milvus-sdk-go/v2/entity"
// )

// func main() {
// 	collectionName := "benchmark_flat"
// 	client, err := client.NewClient(context.Background(), client.Config{
// 		Address: ":19530",
// 	})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	schema := entity.NewSchema().WithName(collectionName).
// 		WithField(entity.NewField().WithName("ID").WithDataType(entity.FieldTypeVarChar).
// 			WithMaxLength(8).WithIsPrimaryKey(true)).
// 		WithField(entity.NewField().WithName("embeddings").WithDataType(entity.FieldTypeFloatVector).WithDim(128))

// 	err = client.CreateCollection(
// 		context.Background(),
// 		schema,
// 		16,
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	idx, err := entity.NewIndexFlat(entity.COSINE)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// client.CreateIndex(context.Background(), collectionName, "embeddings", idx, false)
// 	client.LoadCollection(context.Background(), collectionName, false)

// 	test_vecs := generateTestVectors(1_000_000, 128)
// 	//not-batch each insert
// 	startTime := time.Now()
// 	timer := make(map[int]string)
// 	for idx, vec := range test_vecs {
// 		_, err := client.Insert(context.Background(), collectionName, "",
// 			entity.NewColumnVarChar("ID", []string{strconv.Itoa(idx)}),
// 			entity.NewColumnFloatVector("embeddings", 128, [][]float32{vec}))
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		if idx == 1000 || idx == 5000 || idx == 10000 || idx == 50000 || idx == 100_000 ||
// 			idx == 300_000 || idx == 500_000 || idx == 700_000 || idx == 999_999 {
// 			timer[idx] = fmt.Sprintf("%.2f s", time.Since(startTime).Seconds())
// 		}
// 		if idx%100 == 0 {
// 			fmt.Println("cur idx : ", idx)
// 		}
// 	}
// 	err = client.Flush(context.Background(), collectionName, false)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	timer[-1] = fmt.Sprintf("%.2f s", time.Since(startTime).Seconds())
// 	for k, v := range timer {
// 		fmt.Println(k, v)
// 	}
// }

// func generateTestVectors(num, dim int) [][]float32 {
// 	vectors := make([][]float32, num)
// 	for i := 0; i < num; i++ {
// 		vectors[i] = generateRandomVector(dim)
// 	}
// 	return vectors
// }

// func generateRandomVector(dim int) []float32 {
// 	vec := make([]float32, dim)
// 	for i := 0; i < dim; i++ {
// 		vec[i] = rand.Float32()
// 	}
// 	return vec
// }

// // 700000 2863.72 s
// // 999999 4225.28 s
// // 1000 3.46 s
// // 5000 16.94 s
// // 10000 33.70 s
// // 500000 1998.60 s
// // 50000 170.54 s
// // 100000 357.51 s
// // 300000 1197.36 s
// // -1 4228.34 s

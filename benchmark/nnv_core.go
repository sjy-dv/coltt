package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"playground.benchmark.nnv/coreproto"
)

func main() {
	collectionName := "benchmark_hnsw"

	conn, err := grpc.Dial(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	dclient := coreproto.NewCoreRpcClient(conn)

	_, err = dclient.CreateCollection(context.Background(), &coreproto.CollectionSpec{
		CollectionName:    collectionName,
		VectorDimension:   128,
		Distance:          coreproto.Distance_Euclidean,
		CompressionHelper: coreproto.Quantization_None,
		CollectionConfig: &coreproto.HnswConfig{
			SearchAlgorithm:           coreproto.SearchAlgorithm_Heuristic,
			LevelMultiplier:           -1,
			Ef:                        20,
			EfConstruction:            200,
			M:                         16,
			MMax:                      -1,
			MMax0:                     -1,
			HeuristicExtendCandidates: false,
			HeuristicKeepPruned:       true,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	testVecs := generateTestVectors(1_000_000, 128)
	startTime := time.Now()
	timer := make(map[int]string)
	for idx, vec := range testVecs {

		_, err := dclient.Insert(context.TODO(), &coreproto.DatasetChange{
			Id:               strconv.Itoa(idx),
			Vector:           vec,
			CollectionName:   collectionName,
			IndexChangeTypes: coreproto.IndexChangeTypes_INSERT,
		})
		if err != nil {
			log.Fatal(err)
		}
		if idx == 1000 || idx == 5000 || idx == 10000 || idx == 50000 || idx == 100_000 ||
			idx == 300_000 || idx == 500_000 || idx == 700_000 || idx == 999_999 {
			timer[idx] = fmt.Sprintf("%.2f s", time.Since(startTime).Seconds())
		}
		if idx%100 == 0 {
			fmt.Println("cur idx : ", idx)
		}
	}
	timer[-1] = fmt.Sprintf("%.2f s", time.Since(startTime).Seconds())
	for k, v := range timer {
		fmt.Println(k, v)
	}
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

// 300000 729.83 s
// 700000 1914.14 s
// 999999 2897.50 s
// -1 2897.50 s
// 10000 12.80 s
// 5000 5.53 s
// 50000 92.67 s
// 100000 209.00 s
// 500000 1300.40 s
// 1000 0.93 s

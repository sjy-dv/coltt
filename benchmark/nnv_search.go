package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
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
	rt := time.Now()
	resp, err := dclient.LoadCollection(context.Background(), &coreproto.CollectionName{
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
		resp, err := dclient.VectorSearch(context.Background(), &coreproto.SearchRequest{
			CollectionName: collectionName,
			Vector:         generateRandomVector(128),
			TopK:           10,
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

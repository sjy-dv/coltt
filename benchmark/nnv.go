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
	"google.golang.org/protobuf/types/known/structpb"
	"playground.benchmark.nnv/edgeproto"
)

func main() {
	collectionName := "benchmark_flat"
	conn, err := grpc.Dial(":50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	dclient := edgeproto.NewEdgeRpcClient(conn)

	_, err = dclient.CreateCollection(context.Background(), &edgeproto.Collection{
		CollectionName: collectionName,
		Dim:            128,
		Distance:       edgeproto.Distance_Cosine,
		Quantization:   edgeproto.Quantization_None,
	})

	if err != nil {
		log.Fatal(err)
	}
	test_vecs := generateTestVectors(1_000_000, 128)
	startTime := time.Now()
	timer := make(map[int]string)
	for idx, vec := range test_vecs {
		m, _ := structpb.NewStruct(map[string]interface{}{
			"_id": strconv.Itoa(idx),
		})
		_, err := dclient.Insert(context.Background(), &edgeproto.ModifyDataset{
			Id:             strconv.Itoa(idx),
			Vector:         vec,
			Metadata:       m,
			CollectionName: collectionName,
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
	resp, err := dclient.Flush(context.Background(), &edgeproto.CollectionName{CollectionName: collectionName})
	if !resp.Status || err != nil {
		log.Fatal(resp.Error.ErrorMessage, err)
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

// cache
// 999999 583.83 s
// -1 627.91 s
// 50000 31.64 s
// 5000 3.42 s
// 10000 6.81 s
// 100000 61.25 s
// 300000 174.69 s
// 500000 290.88 s
// 700000 402.35 s
// 1000 0.65 s

// disk
// 10000 6.77 s
// 50000 32.53 s
// 300000 200.83 s
// 500000 344.25 s
// -1 740.98 s
// 1000 0.75 s
// 100000 64.12 s
// 700000 486.86 s
// 999999 704.01 s
// 5000 3.52 s

// 100000 75.66 s
// 300000 223.49 s
// 500000 370.56 s
// -1 720.44 s
// 50000 36.59 s
// 5000 3.32 s
// 10000 6.77 s
// 700000 522.80 s
// 999999 716.43 s
// 1000 0.69 s

// 5000 3.43 s
// 10000 6.72 s
// 500000 315.05 s
// 700000 432.60 s
// -1 616.60 s
// 1000 0.74 s
// 50000 32.45 s
// 100000 66.61 s
// 300000 195.11 s
// 999999 612.41 s

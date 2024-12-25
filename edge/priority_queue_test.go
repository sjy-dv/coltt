package edge

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleVecSpace(t *testing.T) {
	cosStore := newSimpleVectorstore(CollectionConfig{
		Dimension:      3196,
		CollectionName: "test",
		Distance:       COSINE,
		Quantization:   NONE_QAUNTIZATION,
	})
	simrank := map[uint64][]float32{}

	randVecs := generateTestVectors(10000, 3196)

	for i, vec := range randVecs {
		if i%1444 == 0 {
			simrank[uint64(i)] = vec
		}
		cosStore.InsertVector("test", uint64(i), ENode{
			Vector: vec,
			Metadata: map[string]any{
				"meta": i,
			},
		})
	}

	for i, vec := range simrank {
		rs, _ := cosStore.FullScan("test", vec, 15, false)
		assert.Equal(t, i, rs[0].Id)
		t.Log(rs[0].Score)
		fmt.Println("============Candidates=============", ">>>", i)
		for _, r := range rs {
			fmt.Printf("ID: %d SCORE: %v\n", r.Id, ((2-r.Score)/2)*100)
		}
		fmt.Println()
		fmt.Println()
	}

	euStore := newSimpleVectorstore(CollectionConfig{
		Dimension:      3196,
		CollectionName: "test",
		Distance:       COSINE,
		Quantization:   NONE_QAUNTIZATION,
	})
	simrank = map[uint64][]float32{}

	for i, vec := range randVecs {
		if i%1444 == 0 {
			simrank[uint64(i)] = vec
		}
		euStore.InsertVector("test", uint64(i), ENode{
			Vector: vec,
			Metadata: map[string]any{
				"meta": i,
			},
		})
	}

	for i, vec := range simrank {
		rs, _ := euStore.FullScan("test", vec, 15, false)
		assert.Equal(t, i, rs[0].Id)
		t.Log(rs[0].Score)
		fmt.Println("============Candidates=============", ">>>", i)
		for _, r := range rs {
			fmt.Printf("ID: %d SCORE: %v\n", r.Id, r.Score)
		}
		fmt.Println()
		fmt.Println()
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

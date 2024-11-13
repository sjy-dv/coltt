package hnswpq

import (
	"fmt"
	"math/rand"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/nnv/pkg/gomath"
)

func (xx *productQuantizer) PreTrainProductQuantizer(collectionName string, dim, learning_spec int) error {

	dummyVals := generatePreTrainVectors(learning_spec, dim)
	log.Debug().Msgf("Inserting %d dummy vectors for PQ training...", learning_spec)
	for i, vec := range dummyVals {
		nodeId := uint64(i + 1)
		_, err := xx.Set(nodeId, vec)
		if err != nil {
			return fmt.Errorf("failed to set vector %d: %v", nodeId, err)
		}
	}
	fmt.Println("Fitting Product Quantizer with dummy vectors...")
	err := xx.Fit()
	if err != nil {
		return fmt.Errorf("failed to fit Product Quantizer: %v", err)
	}
	fmt.Println("Product Quantizer fitted successfully.")
	return nil
}

func generatePreTrainVectors(num, dim int) []gomath.Vector {
	vectors := make([]gomath.Vector, num)
	for i := 0; i < num; i++ {
		vectors[i] = generateRandomVector(dim)
	}
	return vectors
}

func generateRandomVector(dim int) gomath.Vector {
	vec := make(gomath.Vector, dim)
	for i := 0; i < dim; i++ {
		vec[i] = rand.Float32()
	}
	return vec
}

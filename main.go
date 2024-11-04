package main

import (
	"fmt"

	"github.com/sjy-dv/nnv/pkg/fasthnsw"
)

func main() {

	// Create Index
	vectorSize := 3
	vectorsCount := 100
	conf := fasthnsw.DefaultConfig(uint(vectorSize))
	index, err := fasthnsw.NewIndex(conf)
	if err != nil {
		panic("Failed to create Index")
	}
	defer index.Destroy()

	// Add to Index
	err = index.Reserve(uint(vectorsCount * 10))
	for i := 0; i < vectorsCount; i++ {
		err = index.Add(fasthnsw.Key(i), []float32{float32(i), float32(i + 1), float32(i + 2)})
		if err != nil {
			panic("Failed to add")
		}
	}
	keys, distances, err := index.Search([]float32{0.0, 1.0, 2.0}, 3)
	if err != nil {
		panic("Failed to search")
	}
	fmt.Println(keys, distances)
	fmt.Println("111111111")
	err = index.Save("index.tensor")
	if err != nil {
		panic("Failed to save index")
	}

	// index.Destroy()
	index, err = fasthnsw.NewIndex(conf)
	if err != nil {
		panic("111")
	}
	fmt.Println("22222222222222222")
	err = index.Reserve(uint(vectorsCount * 10))
	err = index.Load("index.tensor")
	if err != nil {
		panic("Failed to save index")
	}
	fmt.Println("33333333333333333333")
	keys, distances, err = index.Search([]float32{0.0, 1.0, 2.0}, 3)
	if err != nil {
		panic("Failed to search")
	}
	fmt.Println(keys, distances)
	fmt.Println("44444444444444444444444")
	// err = index.Reserve(uint(vectorsCount * 10))

	for i := 0; i < vectorsCount; i++ {
		err = index.Add(fasthnsw.Key(i*30+8124), []float32{float32(i), float32(i + 1), float32(i + 2)})
		if err != nil {
			fmt.Println(err)
			panic("Failed to add")
		}
	}
	// Search
	keys, distances, err = index.Search([]float32{0.0, 1.0, 2.0}, 3)
	if err != nil {
		panic("Failed to search")
	}
	fmt.Println(keys, distances)
}

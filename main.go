package main

import (
	"fmt"
	"slices"
)

type Candidate struct {
	Name  string
	Score float64
}

func main() {
	candidates := []Candidate{
		{Name: "Alice", Score: 85.5},
		{Name: "Bob", Score: 92.3},
		{Name: "Charlie", Score: 78.6},
		{Name: "Dave", Score: 99.1},
		{Name: "Eve", Score: 88.7},
	}

	slices.SortFunc(candidates, func(i, j Candidate) int {
		if i.Score > j.Score {
			return -1
		} else if i.Score < j.Score {
			return 1
		}
		return 0
	})

	fmt.Println("Sorted candidates by score (descending):")
	for _, candidate := range candidates {
		fmt.Printf("Name: %s, Score: %.2f\n", candidate.Name, candidate.Score)
	}
}

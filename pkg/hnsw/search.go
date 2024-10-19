package hnsw

import "github.com/google/uuid"

type SearchResult []SearchResultItem

type SearchResultItem struct {
	Id       uuid.UUID
	Metadata Metadata
	Score    float32
}

func (this SearchResult) Len() int {
	return len(this)
}

func (this SearchResult) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this SearchResult) Less(i, j int) bool {
	return this[i].Score < this[j].Score
}

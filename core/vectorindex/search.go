package vectorindex

type SearchResult []SearchResultItem

type SearchResultItem struct {
	Id       uint64
	Metadata map[string]any
	Score    float32
}

func (xx SearchResult) Len() int {
	return len(xx)
}

func (xx SearchResult) Swap(i, j int) {
	xx[i], xx[j] = xx[j], xx[i]
}

func (xx SearchResult) Less(i, j int) bool {
	return xx[i].Score < xx[j].Score
}

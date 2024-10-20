package hnsw

// ------------------------------
// Priority Queue Implementation
// ------------------------------

// Item represents an item in the priority queue.
type Item struct {
	id    uint64
	score float32
	index int
}

// PriorityQueue implements a max-heap based on similarity scores.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

// We want a max-heap, so higher scores have higher priority.
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].score > pq[j].score
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push adds an item to the heap.
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

// Pop removes and returns the highest priority item from the heap.
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	if n == 0 {
		return nil
	}
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// IsEmpty checks if the priority queue is empty.
func (pq PriorityQueue) IsEmpty() bool {
	return len(pq) == 0
}

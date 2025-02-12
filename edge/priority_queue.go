package edge

import (
	"sort"
	"sync"

	"github.com/sjy-dv/coltt/edge/priorityqueue"
)

type SearchResultItem struct {
	Id       uint64
	Metadata map[string]any
	Score    float32
}

type PriorityQueue struct {
	queue   priorityqueue.PriorityQueue
	maxSize int
	mutex   sync.Mutex
}

func NewPriorityQueue(maxSize int) *PriorityQueue {
	return &PriorityQueue{
		queue:   priorityqueue.NewMaxPriorityQueue(),
		maxSize: maxSize,
	}
}

func (pq *PriorityQueue) Add(item *SearchResultItem) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	priorityItem := priorityqueue.NewPriorityQueueItem(item.Score, item)
	pq.queue.Push(priorityItem)
	if pq.queue.Len() > pq.maxSize {
		pq.queue.Pop()
	}
}

func (pq *PriorityQueue) ToSlice() []*SearchResultItem {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	items := pq.queue.ToSlice()
	result := make([]*SearchResultItem, len(items))
	for i, item := range items {
		result[i] = item.Value().(*SearchResultItem)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score < result[j].Score
	})
	return result
}

func (pq *PriorityQueue) Len() int {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	return pq.queue.Len()
}

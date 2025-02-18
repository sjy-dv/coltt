package experimental

import (
	"sort"
	"sync"

	"github.com/sjy-dv/coltt/edge/priorityqueue"
)

// multivector는 이미 점수를 계산하고
// 높은 점수대로 가기 때문에 Min이 맞음

// 기존 대로 낮은 점수가 유사도가 높은 경우에 Max
type shardNeighbor struct {
	NN []*NearestNeighbor
}

type NearestNeighbor struct {
	Id       string
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
		queue:   priorityqueue.NewMinPriorityQueue(),
		maxSize: maxSize,
	}
}

func (pq *PriorityQueue) Add(item *NearestNeighbor) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	priorityItem := priorityqueue.NewPriorityQueueItem(item.Score, item)
	pq.queue.Push(priorityItem)
	if pq.queue.Len() > pq.maxSize {
		pq.queue.Pop()
	}
}

func (pq *PriorityQueue) ToSlice() []*NearestNeighbor {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	items := pq.queue.ToSlice()
	result := make([]*NearestNeighbor, len(items))
	for i, item := range items {
		result[i] = item.Value().(*NearestNeighbor)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})
	return result
}

func (pq *PriorityQueue) Len() int {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()
	return pq.queue.Len()
}

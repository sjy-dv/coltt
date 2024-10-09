package internal

import "container/heap"

type PriorityQueue interface {
	Len() int
	Push(*PriorityQueueItem)
	Pop() *PriorityQueueItem
	Peek() *PriorityQueueItem
	Reverse() PriorityQueue
	ToSlice() []*PriorityQueueItem
	ToIterator() <-chan *PriorityQueueItem
	Values() []interface{}
}

type priorityQueue struct {
	queue   heap.Interface
	maxSize int
}

type PriorityQueueItem struct {
	priority float32
	value    interface{}
}

func NewPriorityQueueItem(priority float32, value interface{}) *PriorityQueueItem {
	return &PriorityQueueItem{priority, value}
}

func NewMinPriorityQueue(items ...*PriorityQueueItem) PriorityQueue {
	queue := make(minPriorityQueue, 0)
	return initializePriorityQueue(&queue, items...)
}

func NewMaxPriorityQueue(items ...*PriorityQueueItem) PriorityQueue {
	queue := make(maxPriorityQueue, 0)
	return initializePriorityQueue(&queue, items...)
}

func initializePriorityQueue(queue heap.Interface, items ...*PriorityQueueItem) PriorityQueue {
	heap.Init(queue)
	pq := &priorityQueue{
		queue: queue,
	}

	for _, item := range items {
		pq.Push(item)
	}

	return pq
}

// Item
func (item *PriorityQueueItem) Priority() float32 {
	return item.priority
}

func (item *PriorityQueueItem) Value() interface{} {
	return item.value
}

// Priority Queue
func (pq *priorityQueue) Len() int {
	return pq.queue.Len()
}

func (pq *priorityQueue) Push(item *PriorityQueueItem) {
	if item.priority < 0 {
		panic("Negative priority")
	}

	heap.Push(pq.queue, item)
}

func (pq *priorityQueue) Pop() *PriorityQueueItem {
	if pq.Len() == 0 {
		panic("Empty priority queue")
	}

	return heap.Pop(pq.queue).(*PriorityQueueItem)
}

func (pq *priorityQueue) Peek() *PriorityQueueItem {
	if pq.Len() == 0 {
		panic("Empty priority queue")
	}

	return pq.ToSlice()[0]
}

func (pq *priorityQueue) Reverse() PriorityQueue {
	switch pq.queue.(type) {
	case *minPriorityQueue:
		queue := maxPriorityQueue(*pq.queue.(*minPriorityQueue))

		return initializePriorityQueue(&queue)
	case *maxPriorityQueue:
		queue := minPriorityQueue(*pq.queue.(*maxPriorityQueue))

		return initializePriorityQueue(&queue)
	default:
		panic("Invalid queue type")
	}
}

func (pq *priorityQueue) ToIterator() <-chan *PriorityQueueItem {
	resultChan := make(chan *PriorityQueueItem)
	go func() {
		for pq.Len() > 0 {
			resultChan <- pq.Pop()
		}
		close(resultChan)
	}()
	return resultChan
}

func (pq *priorityQueue) ToSlice() []*PriorityQueueItem {
	switch pq.queue.(type) {
	case *minPriorityQueue:
		return *pq.queue.(*minPriorityQueue)
	case *maxPriorityQueue:
		return *pq.queue.(*maxPriorityQueue)
	default:
		panic("Invalid queue type")
	}
}

func (pq *priorityQueue) Values() []interface{} {
	queue := pq.ToSlice()
	result := make([]interface{}, len(queue))
	for i, item := range queue {
		result[i] = item.Value()
	}
	return result
}

type minPriorityQueue []*PriorityQueueItem

type maxPriorityQueue []*PriorityQueueItem

func (pq minPriorityQueue) Len() int { return len(pq) }

func (pq minPriorityQueue) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority
}

func (pq minPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *minPriorityQueue) Pop() interface{} {
	queue := *pq
	item := queue[len(queue)-1]
	*pq = queue[0 : len(queue)-1]
	return item
}

func (pq *minPriorityQueue) Push(val interface{}) {
	*pq = append(*pq, val.(*PriorityQueueItem))
}

func (pq maxPriorityQueue) Len() int { return len(pq) }

func (pq maxPriorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

func (pq maxPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *maxPriorityQueue) Pop() interface{} {
	queue := *pq
	item := queue[len(queue)-1]
	*pq = queue[0 : len(queue)-1]
	return item
}

func (pq *maxPriorityQueue) Push(val interface{}) {
	*pq = append(*pq, val.(*PriorityQueueItem))
}

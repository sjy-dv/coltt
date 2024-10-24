package hnsw

import (
	"container/heap"
)

// An Item is something we manage in a priority queue.
type Item struct {
	Node     uint32  // The value of the item; arbitrary.
	Distance float32 // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	Index    int // The index of the item in the heap.
	Metadata map[string]interface{}
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue struct {
	Order bool
	Items []*Item
}

func (pq PriorityQueue) Len() int { return len(pq.Items) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the lowest distance.
	if pq.Order == false {
		return pq.Items[i].Distance < pq.Items[j].Distance
	} else {
		return pq.Items[i].Distance > pq.Items[j].Distance
	}

}

func (pq PriorityQueue) Swap(i, j int) {
	pq.Items[i], pq.Items[j] = pq.Items[j], pq.Items[i]
	pq.Items[i].Index = i
	pq.Items[j].Index = j
}

// Add a new element to our queue
func (pq *PriorityQueue) Push(x any) {
	n := len((*pq).Items)
	item := x.(*Item)
	item.Index = n
	(*pq).Items = append((*pq).Items, item)
}

// Remove the top element (e.g smallest in a min-heap, largest in a max-heap)
func (pq *PriorityQueue) Pop() any {
	old := (*pq).Items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.Index = -1 // for safety
	(*pq).Items = old[0 : n-1]
	return item
}

// Return the top element (either max, or min, depending on heap type)
func (pq *PriorityQueue) Top() any {
	item := (*pq).Items[0]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, node uint32, distance float32) {
	item.Node = node
	item.Distance = distance
	heap.Fix(pq, item.Index)
}

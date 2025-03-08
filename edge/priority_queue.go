// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
		queue:   priorityqueue.NewMinPriorityQueue(),
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

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

package queue

import "container/heap"

type Item struct {
	NodeID   uint64
	Distance float32
	Index    int
}

type PriorityQueue struct {
	Order bool
	Items []*Item
}

func (pq PriorityQueue) Len() int { return len(pq.Items) }

func (pq PriorityQueue) Less(i, j int) bool {
	if !pq.Order {
		return pq.Items[i].Distance < pq.Items[j].Distance
	}
	return pq.Items[i].Distance > pq.Items[j].Distance
}

func (pq PriorityQueue) Swap(i, j int) {
	pq.Items[i], pq.Items[j] = pq.Items[j], pq.Items[i]
	pq.Items[i].Index = i
	pq.Items[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(pq.Items)
	item := x.(*Item)
	item.Index = n
	pq.Items = append(pq.Items, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := pq.Items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	pq.Items = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) Top() interface{} {
	if len(pq.Items) == 0 {
		return nil
	}
	return pq.Items[0]
}

func (pq *PriorityQueue) update(item *Item, node uint64, distance float32) {
	item.NodeID = node
	item.Distance = distance
	heap.Fix(pq, item.Index)
}

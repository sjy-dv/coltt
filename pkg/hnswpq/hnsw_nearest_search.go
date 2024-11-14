package hnswpq

import (
	"container/heap"

	"github.com/sjy-dv/nnv/pkg/bitset"
	"github.com/sjy-dv/nnv/pkg/gomath"
	"github.com/sjy-dv/nnv/pkg/queue"
)

func (xx *Hnsw) findEp(vec gomath.Vector, curObj *Node, layer int16) (match Node, curDist float32, err error) {
	curDist = xx.DistFn(vec, curObj.Vectors)
	for level := xx.MaxLevel; level > 0; level-- {
		scan := true

		for scan {
			scan = false

			for _, nodeId := range xx.getConnection(curObj, level) {
				node := xx.NodeList.Nodes[nodeId]
				nodeDist := xx.DistFn(node.Vectors, vec)
				if nodeDist < curDist {
					match = node
					curDist = nodeDist
					scan = true
				}
			}
		}
	}
	return match, curDist, nil
}

func (xx *Hnsw) searchLayer(vec gomath.Vector, ep *queue.Item, topCandidates *queue.PriorityQueue, ef int, level uint) error {
	var visited bitset.BitSet

	candidates := &queue.PriorityQueue{Order: false, Items: []*queue.Item{}}
	heap.Init(candidates)
	heap.Push(candidates, ep)

	topCandidates.Order = true
	heap.Init(topCandidates)
	heap.Push(topCandidates, ep)

	for candidates.Len() > 0 {

		lowerBound := topCandidates.Top().(*queue.Item).Distance
		candidate := heap.Pop(candidates).(*queue.Item)

		if candidate.Distance > lowerBound {
			break
		}
		for _, nodeId := range xx.NodeList.Nodes[candidate.NodeID].LinkNodes[level] {
			if !visited.Test(uint(nodeId)) {
				visited.Set(uint(nodeId))
				node := xx.NodeList.Nodes[nodeId]
				nodeDist := xx.PQ.DistanceFromCentroidIDs(vec, node.Centroids)
				item := &queue.Item{
					Distance: nodeDist,
					NodeID:   node.Id,
				}
				topDistance := topCandidates.Top().(*queue.Item).Distance

				if topCandidates.Len() < ef {
					if node.Id != ep.NodeID {
						heap.Push(topCandidates, item)
					}
					heap.Push(candidates, item)
				} else if topDistance > nodeDist {
					heap.Push(topCandidates, item)
					heap.Pop(topCandidates)
					heap.Push(candidates, item)
				}
			}
		}
	}
	return nil
}

func (xx *Hnsw) SelectNeighboursSimple(topCandidates *queue.PriorityQueue, M int) {
	for topCandidates.Len() > M {
		_ = heap.Pop(topCandidates).(*queue.Item)
	}
}

func (xx *Hnsw) SelectNeighboursHeuristic(topCandidates *queue.PriorityQueue, M int, order bool) {
	if topCandidates.Len() < M {
		return
	}

	newCandidates := &queue.PriorityQueue{Order: order, Items: []*queue.Item{}}
	heap.Init(newCandidates)

	items := make([]*queue.Item, 0, M)

	if !order {
		for topCandidates.Len() > 0 {
			item := heap.Pop(topCandidates).(*queue.Item)
			heap.Push(newCandidates, item)
		}
	} else {
		newCandidates = topCandidates
	}

	for newCandidates.Len() > 0 {
		if len(items) >= M {
			break
		}
		item := heap.Pop(newCandidates).(*queue.Item)

		hit := true

		for _, v := range items {
			nodeDist := xx.DistFn(
				xx.NodeList.Nodes[v.NodeID].Vectors,
				xx.NodeList.Nodes[item.NodeID].Vectors,
			)

			if nodeDist < item.Distance {
				hit = false
				break
			}
		}

		if hit {
			items = append(items, item)
		} else {
			heap.Push(newCandidates, item)
		}
	}

	for len(items) < M && newCandidates.Len() > 0 {
		item := heap.Pop(newCandidates).(*queue.Item)
		items = append(items, item)
	}

	for _, item := range items {
		heap.Push(topCandidates, item)
	}
}

func (xx *Hnsw) addConnections(neighbourNode uint64, newNode uint64, level int) {
	var maxConnections int

	if level == 0 {
		maxConnections = xx.Mmax0
	} else {
		maxConnections = xx.Mmax
	}

	xx.NodeList.Nodes[neighbourNode].LinkNodes[level] = append(
		xx.NodeList.Nodes[neighbourNode].LinkNodes[level], newNode,
	)

	curConnections := len(xx.NodeList.Nodes[neighbourNode].LinkNodes[level])

	if curConnections > maxConnections {
		switch xx.Heuristic {
		case false:
			topCandidates := &queue.PriorityQueue{Order: true, Items: []*queue.Item{}}
			heap.Init(topCandidates)

			for i := 0; i < curConnections; i++ {
				connectedNode := xx.NodeList.Nodes[neighbourNode].LinkNodes[level][i]
				distanceBetweenNodes := xx.DistFn(
					xx.NodeList.Nodes[neighbourNode].Vectors,
					xx.NodeList.Nodes[connectedNode].Vectors,
				)
				heap.Push(topCandidates, &queue.Item{
					NodeID:   connectedNode,
					Distance: distanceBetweenNodes,
				})
			}

			xx.SelectNeighboursSimple(topCandidates, maxConnections)

			xx.NodeList.Nodes[neighbourNode].LinkNodes[level] = make([]uint64, maxConnections)

			for i := maxConnections - 1; i >= 0; i-- {
				node := heap.Pop(topCandidates).(*queue.Item)
				xx.NodeList.Nodes[neighbourNode].LinkNodes[level][i] = node.NodeID
			}
		case true:
			topCandidates := &queue.PriorityQueue{Order: false, Items: []*queue.Item{}}
			heap.Init(topCandidates)

			for i := 0; i < curConnections; i++ {
				connectedNode := xx.NodeList.Nodes[neighbourNode].LinkNodes[level][i]
				distanceBetweenNodes := xx.DistFn(
					xx.NodeList.Nodes[neighbourNode].Vectors,
					xx.NodeList.Nodes[connectedNode].Vectors,
				)
				heap.Push(topCandidates, &queue.Item{
					NodeID:   connectedNode,
					Distance: distanceBetweenNodes,
				})
			}

			xx.SelectNeighboursSimple(topCandidates, maxConnections)
			xx.NodeList.Nodes[neighbourNode].LinkNodes[level] = make([]uint64, maxConnections)

			for i := 0; i < maxConnections; i++ {
				node := heap.Pop(topCandidates).(*queue.Item)
				xx.NodeList.Nodes[neighbourNode].LinkNodes[level][i] = node.NodeID
			}
		}
	}
}

func (xx *Hnsw) getConnection(ep *Node, level int) []uint64 {
	return ep.LinkNodes[level]
}

// func (xx *Hnsw) removeConnection(nodeId uint64) error {
// 	node := &xx.NodeList.Nodes[nodeId]
// 	if node.Id == 0 {
// 		return errors.New("node not found")
// 	}

// 	for level := 0; level <= xx.MaxLevel; level++ {
// 		xx.NodeList.lock.Lock()
// 		connections := node.LinkNodes[level]
// 		for _, neighbourId := range connections {
// 			neighbor := &xx.NodeList.Nodes[neighbourId]
// 			newLinks := []uint64{}
// 			for _, link := range neighbor.LinkNodes[level] {
// 				if link != nodeId {
// 					newLinks = append(newLinks, link)
// 				}
// 			}
// 			neighbor.LinkNodes[level] = newLinks
// 		}
// 		xx.NodeList.lock.Unlock()
// 	}

// 	xx.NodeList.lock.Lock()
// 	xx.NodeList.Nodes[nodeId] = Node{
// 		Id:      nodeId,
// 		IsEmpty: true,
// 	}
// 	xx.NodeList.lock.Unlock()
// 	xx.hlock.Lock()
// 	xx.EmptyNodes = append(xx.EmptyNodes, nodeId)
// 	xx.hlock.Unlock()
// 	return nil
// }

package hnsw

import (
	"container/heap"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/sjy-dv/nnv/pkg/gomath"
	matchdbgo "github.com/sjy-dv/nnv/pkg/match_db.go"
	"github.com/sjy-dv/nnv/pkg/nnlogdb"
)

// nnlogdb using insert & delete
/*
	update => delete(nnlog record) + insert(nnlogrecord)

	colum => 1  aaaa
	update => 1 bbbb
		- internal flow
			1 aaaa <- delete (nnlog record to {delete-event, 1, aaaa})
			1 bbbb <- insert (nnlog record to {insert-event, 1, bbbb})
	when recovery
		- internal flow
			1 aaaa <- old log
			<- delete event : drop column
			<- insert event : add bbbb column but update must required full data, don t worry
			&& this hnsw first set node location is empty Nodes
	search => not necessary

*/

// isud => insert update delete... hahaha
// metadata include unserNodeId
func (xx *HnswBucket) Insert(bucketName string, userNodeId string, vec gomath.Vector, metadata map[string]interface{}) (uint32, error) {
	if !xx.BucketGroup[bucketName] {
		return 0, fmt.Errorf("not exists bucket : %s", bucketName)
	}
	if xx.Buckets[bucketName].Dim != uint32(vec.Len()) {
		return 0, fmt.Errorf("bucket expect dim: %d\ngot dim: %d\n dimension must be samed",
			xx.Buckets[bucketName].Dim, vec.Len())
	}
	node := Node{}
	node.Vectors = make(gomath.Vector, vec.Len())
	node.Vectors = vec
	node.Metadata = make(map[string]interface{})
	node.Metadata = metadata
	node.Timestamp = uint64(0)
	node.Timestamp = uint64(time.Now().UnixNano())
	curObj := &xx.Buckets[bucketName].NodeList.Nodes[xx.Buckets[bucketName].Ep]
	curDist := xx.Buckets[bucketName].Space.Distance(curObj.Vectors, vec)

	xx.Buckets[bucketName].NodeList.rmu.Lock()

	node.Layer = int(math.Floor(-math.Log(rand.Float64()) * xx.Buckets[bucketName].Ml))
	node.Id = uint32(len(xx.Buckets[bucketName].NodeList.Nodes))
	node.LinkNodes = make([][]uint32, xx.Buckets[bucketName].M+1)
	if len(xx.Buckets[bucketName].EmptyNodes) != 0 {
		emptyNodeId := xx.Buckets[bucketName].EmptyNodes[len(xx.Buckets[bucketName].EmptyNodes)-1]
		xx.Buckets[bucketName].EmptyNodes = xx.Buckets[bucketName].
			EmptyNodes[:len(xx.Buckets[bucketName].EmptyNodes)-1]
		node.Id = emptyNodeId
		xx.Buckets[bucketName].NodeList.Nodes[emptyNodeId] = node
	} else {
		xx.Buckets[bucketName].NodeList.Nodes = append(xx.Buckets[bucketName].NodeList.Nodes, node)
	}
	xx.Buckets[bucketName].NodeList.rmu.Unlock()

	pq := &PriorityQueue{}
	pq.Order = false
	heap.Init(pq)

	var topCandidates PriorityQueue
	topCandidates.Order = false

	for level := curObj.Layer; level > node.Layer; level-- {
		changed := true

		for changed {
			changed = false

			for _, nodeId := range xx.Buckets[bucketName].getConnection(curObj, level) {
				nodeDist := xx.Buckets[bucketName].Space.Distance(
					xx.Buckets[bucketName].NodeList.Nodes[nodeId].Vectors,
					vec,
				)
				if nodeDist < curDist {
					curObj = &xx.Buckets[bucketName].NodeList.Nodes[nodeId]
					curDist = nodeDist
					changed = true
				}
			}
		}
	}

	heap.Push(pq, &Item{Distance: curDist, Node: curObj.Id, Metadata: curObj.Metadata})

	for level := min(int(node.Layer),
		int(xx.Buckets[bucketName].MaxLevel)); level >= 0; level-- {
		err := xx.Buckets[bucketName].searchLayer(vec, &Item{
			Distance: curDist,
			Node:     curObj.Id,
		}, &topCandidates,
			int(xx.Buckets[bucketName].Efconstruction),
			uint(level))
		if err != nil {
			return 0, err
		}

		switch xx.Buckets[bucketName].Heuristic {
		case false:
			xx.Buckets[bucketName].SelectNeighboursSimple(&topCandidates, int(
				xx.Buckets[bucketName].M,
			))
		case true:
			xx.Buckets[bucketName].SelectNeighboursHeuristic(&topCandidates, int(
				xx.Buckets[bucketName].M,
			), false)
		}

		node.LinkNodes[level] = make([]uint32, topCandidates.Len())

		for i := topCandidates.Len() - 1; i >= 0; i-- {
			candidate := heap.Pop(&topCandidates).(*Item)
			node.LinkNodes[level][i] = candidate.Node
		}
	}

	xx.Buckets[bucketName].NodeList.rmu.Lock()
	xx.Buckets[bucketName].NodeList.Nodes[node.Id].LinkNodes = node.LinkNodes
	xx.Buckets[bucketName].NodeList.rmu.Unlock()

	for level := min(int(node.Layer), int(xx.Buckets[bucketName].MaxLevel)); level >= 0; level-- {
		xx.Buckets[bucketName].NodeList.rmu.Lock()
		for _, neighbourNode := range xx.Buckets[bucketName].NodeList.Nodes[node.Id].LinkNodes[level] {
			xx.Buckets[bucketName].addConnections(neighbourNode, node.Id, level)
		}
		xx.Buckets[bucketName].NodeList.rmu.Unlock()
	}

	if node.Layer > xx.Buckets[bucketName].MaxLevel {
		xx.Buckets[bucketName].rmu.Lock()
		xx.Buckets[bucketName].Ep = int64(node.Id)
		xx.Buckets[bucketName].MaxLevel = node.Layer
		xx.Buckets[bucketName].rmu.Unlock()
	}
	err := matchdbgo.Set(userNodeId, node.Id)
	if err != nil {
		rerr := xx.Buckets[bucketName].removeConnection(node.Id)
		if rerr != nil {
			return 0, fmt.Errorf("matchedDB.Set.Error: %v\nremovedError: %v", err, rerr)
		}
		return 0, err
	}
	// save-log
	err = nnlogdb.AddLogf(
		nnlogdb.PrintlF(
			userNodeId, "insert", bucketName, node.Id, node.Timestamp, node.Metadata, node.Vectors,
		),
	)
	if err != nil {
		rerr := xx.Buckets[bucketName].removeConnection(node.Id)
		if rerr != nil {
			return 0, fmt.Errorf("matchedDB.Set.Error: %v\nremovedError: %v", err, rerr)
		}
		return 0, err
	}
	// bytevec, err := msgpack.Marshal(node)
	// if err != nil {
	// 	rerr := xx.Buckets[bucketName].removeConnection(node.Id)
	// 	if rerr != nil {
	// 		return fmt.Errorf("msgpackMarshalError: %v\nremovedError: %v", err, rerr)
	// 	}
	// 	return err
	// }
	// err = xx.Storage.Put(
	// 	[]byte(fmt.Sprintf("%s_%s", bucketName, userNodeId)),
	// 	bytevec,
	// )
	// if err != nil {
	// 	rerr := xx.Buckets[bucketName].removeConnection(node.Id)
	// 	if rerr != nil {
	// 		return fmt.Errorf("storageError: %v\nremovedError: %v", err, rerr)
	// 	}
	// 	return err
	// }
	return node.Id, nil
}

// return old nodeid, new nodeid, error
func (xx *HnswBucket) Update(bucketName string, nodeId string,
	vec gomath.Vector, metadata map[string]interface{}) (
	uint32, uint32, map[string]interface{}, error,
) {
	// val, err := xx.Storage.Get([]byte(nodeId))
	// if err != nil {
	// 	return err
	// }
	// node := Node{}
	// err = msgpack.Unmarshal(val, &node)
	// if err != nil {
	// 	return err
	// }
	// err = xx.Storage.Delete([]byte(nodeId))
	// if err != nil {
	// 	return err
	// }

	val, err := matchdbgo.Get(nodeId)
	if err != nil {
		return 0, 0, nil, err
	}
	err = matchdbgo.Delete(nodeId)
	if err != nil {
		return 0, 0, nil, err
	}
	copyMeta := xx.Buckets[bucketName].NodeList.Nodes[val].Metadata
	err = xx.Buckets[bucketName].removeConnection(val)
	if err != nil {
		return 0, 0, nil, err
	}
	// insert record to nnlog
	newId, err := xx.Insert(bucketName, nodeId, vec, metadata)
	return val, newId, copyMeta, err
}

func (xx *HnswBucket) Delete(bucketName string, nodeId string) (
	uint32, map[string]interface{}, error) {
	val, err := matchdbgo.Get(nodeId)
	if err != nil {
		return 0, nil, err
	}
	err = matchdbgo.Delete(nodeId)
	if err != nil {
		return 0, nil, err
	}
	copyMeta := xx.Buckets[bucketName].NodeList.Nodes[val].Metadata
	err = xx.Buckets[bucketName].removeConnection(val)
	if err != nil {
		return 0, nil, err
	}
	err = nnlogdb.AddLogf(
		nnlogdb.PrintlF(
			nodeId, "delete", bucketName, val, 0, map[string]interface{}{"_id": nodeId}, []float32{0.0},
		),
	)
	if err != nil {
		return 0, nil, err
	}
	return val, copyMeta, err
}

func (xx *HnswBucket) Search(bucketName string, vec gomath.Vector, topCandidates *PriorityQueue, K int, efSearch int) (err error) {
	curObj := &xx.Buckets[bucketName].NodeList.Nodes[xx.Buckets[bucketName].Ep]

	match, curDist, err := xx.Buckets[bucketName].findEp(vec, curObj, 0)
	if err != nil {
		return err
	}

	err = xx.Buckets[bucketName].searchLayer(
		vec,
		&Item{Distance: curDist, Node: match.Id, Metadata: match.Metadata},
		topCandidates,
		efSearch,
		0,
	)
	if err != nil {
		return err
	}
	for topCandidates.Len() > K {
		_ = heap.Pop(topCandidates).(*Item)
	}
	return nil
}

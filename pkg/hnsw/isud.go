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
func (self *HnswBucket) Insert(bucketName string, userNodeId string, vec gomath.Vector, metadata map[string]interface{}) error {
	if !self.BucketGroup[bucketName] {
		return fmt.Errorf("not exists bucket : %s", bucketName)
	}
	if self.Buckets[bucketName].Dim != uint32(vec.Len()) {
		return fmt.Errorf("bucket expect dim: %d\ngot dim: %d\n dimension must be samed",
			self.Buckets[bucketName].Dim, vec.Len())
	}
	node := Node{}
	node.Vectors = make(gomath.Vector, vec.Len())
	node.Vectors = vec
	node.Metadata = make(map[string]interface{})
	node.Metadata = metadata
	node.Timestamp = uint64(0)
	node.Timestamp = uint64(time.Now().UnixNano())
	curObj := &self.Buckets[bucketName].NodeList.Nodes[self.Buckets[bucketName].Ep]
	curDist := self.Buckets[bucketName].Space.Distance(curObj.Vectors, vec)

	self.Buckets[bucketName].NodeList.rmu.Lock()

	node.Layer = int(math.Floor(-math.Log(rand.Float64()) * self.Buckets[bucketName].Ml))
	node.Id = uint32(len(self.Buckets[bucketName].NodeList.Nodes))
	node.LinkNodes = make([][]uint32, self.Buckets[bucketName].M+1)
	if len(self.Buckets[bucketName].EmptyNodes) != 0 {
		emptyNodeId := self.Buckets[bucketName].EmptyNodes[len(self.Buckets[bucketName].EmptyNodes)-1]
		self.Buckets[bucketName].EmptyNodes = self.Buckets[bucketName].
			EmptyNodes[:len(self.Buckets[bucketName].EmptyNodes)-1]
		node.Id = emptyNodeId
		self.Buckets[bucketName].NodeList.Nodes[emptyNodeId] = node
	} else {
		self.Buckets[bucketName].NodeList.Nodes = append(self.Buckets[bucketName].NodeList.Nodes, node)
	}
	self.Buckets[bucketName].NodeList.rmu.Unlock()

	pq := &PriorityQueue{}
	pq.Order = false
	heap.Init(pq)

	var topCandidates PriorityQueue
	topCandidates.Order = false

	for level := curObj.Layer; level > node.Layer; level-- {
		changed := true

		for changed {
			changed = false

			for _, nodeId := range self.Buckets[bucketName].getConnection(curObj, level) {
				nodeDist := self.Buckets[bucketName].Space.Distance(
					self.Buckets[bucketName].NodeList.Nodes[nodeId].Vectors,
					vec,
				)
				if nodeDist < curDist {
					curObj = &self.Buckets[bucketName].NodeList.Nodes[nodeId]
					curDist = nodeDist
					changed = true
				}
			}
		}
	}

	heap.Push(pq, &Item{Distance: curDist, Node: curObj.Id, Metadata: curObj.Metadata})

	for level := min(int(node.Layer),
		int(self.Buckets[bucketName].MaxLevel)); level >= 0; level-- {
		err := self.Buckets[bucketName].searchLayer(vec, &Item{
			Distance: curDist,
			Node:     curObj.Id,
		}, &topCandidates,
			int(self.Buckets[bucketName].Efconstruction),
			uint(level))
		if err != nil {
			return err
		}

		switch self.Buckets[bucketName].Heuristic {
		case false:
			self.Buckets[bucketName].SelectNeighboursSimple(&topCandidates, int(
				self.Buckets[bucketName].M,
			))
		case true:
			self.Buckets[bucketName].SelectNeighboursHeuristic(&topCandidates, int(
				self.Buckets[bucketName].M,
			), false)
		}

		node.LinkNodes[level] = make([]uint32, topCandidates.Len())

		for i := topCandidates.Len() - 1; i >= 0; i-- {
			candidate := heap.Pop(&topCandidates).(*Item)
			node.LinkNodes[level][i] = candidate.Node
		}
	}

	self.Buckets[bucketName].NodeList.rmu.Lock()
	self.Buckets[bucketName].NodeList.Nodes[node.Id].LinkNodes = node.LinkNodes
	self.Buckets[bucketName].NodeList.rmu.Unlock()

	for level := min(int(node.Layer), int(self.Buckets[bucketName].MaxLevel)); level >= 0; level-- {
		self.Buckets[bucketName].NodeList.rmu.Lock()
		for _, neighbourNode := range self.Buckets[bucketName].NodeList.Nodes[node.Id].LinkNodes[level] {
			self.Buckets[bucketName].addConnections(neighbourNode, node.Id, level)
		}
		self.Buckets[bucketName].NodeList.rmu.Unlock()
	}

	if node.Layer > self.Buckets[bucketName].MaxLevel {
		self.Buckets[bucketName].rmu.Lock()
		self.Buckets[bucketName].Ep = int64(node.Id)
		self.Buckets[bucketName].MaxLevel = node.Layer
		self.Buckets[bucketName].rmu.Unlock()
	}
	err := matchdbgo.Set(userNodeId, node.Id)
	if err != nil {
		rerr := self.Buckets[bucketName].removeConnection(node.Id)
		if rerr != nil {
			return fmt.Errorf("matchedDB.Set.Error: %v\nremovedError: %v", err, rerr)
		}
		return err
	}
	// save-log
	err = nnlogdb.AddLogf(
		nnlogdb.PrintlF(
			userNodeId, "insert", node.Id, node.Timestamp, node.Metadata, node.Vectors,
		),
	)
	if err != nil {
		rerr := self.Buckets[bucketName].removeConnection(node.Id)
		if rerr != nil {
			return fmt.Errorf("matchedDB.Set.Error: %v\nremovedError: %v", err, rerr)
		}
		return err
	}
	// bytevec, err := msgpack.Marshal(node)
	// if err != nil {
	// 	rerr := self.Buckets[bucketName].removeConnection(node.Id)
	// 	if rerr != nil {
	// 		return fmt.Errorf("msgpackMarshalError: %v\nremovedError: %v", err, rerr)
	// 	}
	// 	return err
	// }
	// err = self.Storage.Put(
	// 	[]byte(fmt.Sprintf("%s_%s", bucketName, userNodeId)),
	// 	bytevec,
	// )
	// if err != nil {
	// 	rerr := self.Buckets[bucketName].removeConnection(node.Id)
	// 	if rerr != nil {
	// 		return fmt.Errorf("storageError: %v\nremovedError: %v", err, rerr)
	// 	}
	// 	return err
	// }
	return nil
}

func (self *HnswBucket) Update(bucketName string, nodeId string, vec gomath.Vector, metadata map[string]interface{}) error {
	// val, err := self.Storage.Get([]byte(nodeId))
	// if err != nil {
	// 	return err
	// }
	// node := Node{}
	// err = msgpack.Unmarshal(val, &node)
	// if err != nil {
	// 	return err
	// }
	// err = self.Storage.Delete([]byte(nodeId))
	// if err != nil {
	// 	return err
	// }
	val, err := matchdbgo.Get(nodeId)
	if err != nil {
		return err
	}
	err = matchdbgo.Delete(nodeId)
	if err != nil {
		return err
	}
	err = self.Buckets[bucketName].removeConnection(val)
	if err != nil {
		return err
	}
	// insert record to nnlog
	return self.Insert(bucketName, nodeId, vec, metadata)
}

func (self *HnswBucket) Delete(bucketName string, nodeId string) error {
	val, err := matchdbgo.Get(nodeId)
	if err != nil {
		return err
	}
	// node := Node{}
	// err = msgpack.Unmarshal(val, &node)
	// if err != nil {
	// 	return err
	// }
	// err = self.Storage.Delete([]byte(nodeId))
	// if err != nil {
	// 	return err
	// }
	err = matchdbgo.Delete(nodeId)
	if err != nil {
		return err
	}
	err = self.Buckets[bucketName].removeConnection(val)
	if err != nil {
		return err
	}
	err = nnlogdb.AddLogf(
		nnlogdb.PrintlF(
			nodeId, "delete", val, 0, map[string]interface{}{"_id": nodeId}, []float32{0.0},
		),
	)
	if err != nil {
		return err
	}
	return err
}

func (self *HnswBucket) Search(bucketName string, vec gomath.Vector, topCandidates *PriorityQueue, K int, efSearch int) (err error) {
	curObj := &self.Buckets[bucketName].NodeList.Nodes[self.Buckets[bucketName].Ep]

	match, curDist, err := self.Buckets[bucketName].findEp(vec, curObj, 0)
	if err != nil {
		return err
	}

	err = self.Buckets[bucketName].searchLayer(
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

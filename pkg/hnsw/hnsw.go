package hnsw

import (
	"container/heap"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sjy-dv/nnv/pkg/bitset"
	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/gomath"
)

// func (xx *HnswBucket) Start(opts *kv.Options) error {

// 	xx.rmu.Lock()
// 	xx.BucketGroup = make(map[string]bool)
// 	xx.Buckets = make(map[string]*Hnsw)
// 	defer xx.rmu.Unlock()
// 	if opts == nil {
// 		opts = &kv.DefaultOptions
// 		opts.DirPath = "./data_dir/ann"
// 	}

// 	kvstore, err := kv.Open(*opts)
// 	if err != nil {
// 		log.Warn().Err(err).Msg("pkg.hnsw.hnsw.go(20) open kv file failed error")
// 		return err
// 	}

// 	xx.Storage = kvstore
// 	// reload hnsw config & node data
// 	iter, err := kvstore.NewIterator(kv.IteratorOptions{
// 		Reverse: false,
// 		Prefix:  []byte(BucketPrefix),
// 	})
// 	if err != nil {
// 		log.Warn().Err(err).Msg("pkg.hnsw.hnsw.go(30) kv iterator failed error")
// 		return err
// 	}

// 	for iter.Valid() {
// 		fmt.Println(iter.Value())
// 		err := xx.dataloader(string(iter.Value()))
// 		if err != nil {
// 			log.Warn().Err(err).Msg("pkg.hnsw.hnsw.go(38) dataloader.Fn failed error")
// 			return err
// 		}
// 		iter.Next()
// 	}

// 	if err := iter.Close(); err != nil {
// 		log.Warn().Err(err).Msg("pkg.hnsw.hnsw.go(38) kv iterator closed error")
// 		return err
// 	}
// 	return nil
// }

// func (xx *HnswBucket) dataloader(bucketName string) error {
// 	cfgbytes, err := xx.Storage.Get([]byte(fmt.Sprintf("%s%s", bucketName, BucketConfigPrefix)))
// 	if err != nil {
// 		log.Warn().Err(err).Msg(fmt.Sprintf("pkg.hnsw.hnsw.go(55) bucket %s config load failed error", bucketName))
// 		return err
// 	}
// 	var cfg HnswConfig
// 	err = msgpack.Unmarshal(cfgbytes, &cfg)
// 	if err != nil {
// 		log.Warn().Err(err).Msg(fmt.Sprintf("pkg.hnsw.hnsw.go(62) bucket %s config data msgpack.Unmarshal failed error", bucketName))
// 		return err
// 	}
// 	emptyNodebytes, err := xx.Storage.Get([]byte(fmt.Sprintf("%s%s", bucketName, BucketEmptyNodePrefix)))
// 	if err != nil {
// 		log.Warn().Err(err).Msg(fmt.Sprintf("pkg.hnsw.hnsw.go(70) bucket %s empty node data load failed error", bucketName))
// 		return err
// 	}
// 	var emptyNodes []uint32
// 	err = msgpack.Unmarshal(emptyNodebytes, &emptyNodes)
// 	if err != nil {
// 		log.Warn().Err(err).Msg(fmt.Sprintf("pkg.hnsw.hnsw.go(76) bucket %s empty node data msgpack.Unmarshal failed error", bucketName))
// 		return err
// 	}
// 	iter, err := xx.Storage.NewIterator(kv.IteratorOptions{
// 		Reverse: false,
// 		Prefix:  []byte(bucketName + "_"),
// 	})
// 	if err != nil {
// 		log.Warn().Err(err).Msg(fmt.Sprintf("pkg.hnsw.hnsw.go(6) bucket %s vector data loaded failed error", bucketName))
// 		return err
// 	}
// 	nodes := []Node{}
// 	for iter.Valid() {
// 		if !bytes.HasPrefix(iter.Key(), []byte(bucketName+"_")) {
// 			log.Warn().Msg(fmt.Sprintf("pkg.hnsw.hnsw.go(78) iterator key unmatched expected: %ssomething get: %s", bucketName+"_", string(iter.Key())))
// 			return errors.New("iterator key unmatched")
// 		}
// 		node := Node{}
// 		err := msgpack.Unmarshal(iter.Value(), &node)
// 		if err != nil {
// 			log.Warn().Err(err).Msg("pkg.hnsw.hnsw.go(62) bucket data msgpack.Unmarshal failed error")
// 			return err
// 		}
// 		nodes = append(nodes, node)
// 		iter.Next()
// 	}
// 	if err := iter.Close(); err != nil {
// 		log.Warn().Err(err).Msg("pkg.hnsw.hnsw.go(38) kv iterator closed error")
// 		return err
// 	}
// 	// ------find bitmap index---------

// 	// --------sort data----------
// 	// hnsw must order put node
// 	sort.Slice(nodes, func(i, j int) bool {
// 		return nodes[i].Id < nodes[j].Id
// 	})
// 	//------------------ transfer empty node-----------------
// 	xx.BucketGroup[bucketName] = true
// 	xx.Buckets[bucketName] = &Hnsw{
// 		Efconstruction: cfg.Efconstruction,
// 		M:              cfg.M,
// 		Mmax:           cfg.Mmax,
// 		Mmax0:          cfg.Mmax0,
// 		Ml:             cfg.Ml,
// 		Ep:             cfg.Ep,
// 		MaxLevel:       cfg.MaxLevel,
// 		Dim:            cfg.Dim,
// 		Heuristic:      cfg.Heuristic,
// 		Space:          cfg.Space,
// 		NodeList:       NodeList{Nodes: nodes},
// 		BucketName:     cfg.BucketName,
// 		EmptyNodes:     emptyNodes,
// 	}
// 	return nil
// }

func (xx *HnswBucket) NewHnswBucket(bucketName string, config HnswConfig, dist distance.Space) error {
	xx.rmu.RLock()
	defer xx.rmu.RUnlock()
	if ok := xx.BucketGroup[bucketName]; ok {
		return fmt.Errorf("bucket[%s] is already exists", bucketName)
	}

	xx.BucketGroup[bucketName] = true
	xx.Buckets[bucketName] = &Hnsw{
		Efconstruction: config.Efconstruction,
		M:              config.M,
		Mmax:           config.Mmax,
		Mmax0:          config.Mmax0,
		Ml:             config.Ml,
		Ep:             config.Ep,
		MaxLevel:       config.MaxLevel,
		Dim:            config.Dim,
		Heuristic:      config.Heuristic,
		Space:          dist,
		NodeList: NodeList{
			Nodes: make([]Node, 1),
		},
		BucketName: bucketName,
	}
	genesisNode := Node{}
	genesisNode.Id = 0
	genesisNode.Layer = 0
	genesisNode.Vectors = make(gomath.Vector, xx.Buckets[bucketName].Dim)
	genesisNode.LinkNodes = make([][]uint32, xx.Buckets[bucketName].Mmax0+1)
	genesisNode.Metadata = map[string]interface{}{
		"_id": uuid.New(),
	}
	genesisNode.Timestamp = uint64(time.Now().UnixNano())
	// genesisVal, err := msgpack.Marshal(genesisNode)
	// if err != nil {
	// 	log.Warn().Err(err).Msg("pkg.hnsw.hnsw.go(144) msgpackV5.Marshal failed error")
	// 	return err
	// }
	// node uint id is node.len, then delete the node data, duplicate id,...
	// serial := shortid.MustGenerate()
	// err = xx.Storage.Put([]byte(fmt.Sprintf("%s_%d_%s", bucketName, 0, serial)), genesisVal)
	// if err != nil {
	// 	log.Warn().Err(err).Msg("pkg.hnsw.hnsw.go(149) kv put genesis event failed error")
	// 	return err
	// }
	xx.Buckets[bucketName].NodeList.Nodes[0] = genesisNode
	return nil
}

func (xx *Hnsw) getConnection(ep *Node, level int) []uint32 {
	return ep.LinkNodes[level]
}

func (xx *Hnsw) removeConnection(nodeId uint32) error {
	node := &xx.NodeList.Nodes[nodeId]
	if node.Id == 0 && !node.IsEmpty {
		return errors.New("node not found")
	}

	for level := 0; level <= xx.MaxLevel; level++ {
		xx.NodeList.rmu.Lock()
		connections := node.LinkNodes[level]
		for _, neighbourId := range connections {
			neighbor := &xx.NodeList.Nodes[neighbourId]
			newLinks := []uint32{}
			for _, link := range neighbor.LinkNodes[level] {
				if link != nodeId {
					newLinks = append(newLinks, link)
				}
			}
			neighbor.LinkNodes[level] = newLinks
		}
		xx.NodeList.rmu.Unlock()
	}

	xx.NodeList.rmu.Lock()
	xx.NodeList.Nodes[nodeId] = Node{
		Id:      nodeId,
		IsEmpty: true,
	}
	xx.NodeList.rmu.Unlock()
	xx.rmu.Lock()
	xx.EmptyNodes = append(xx.EmptyNodes, nodeId)
	xx.rmu.Unlock()
	return nil
}

func (xx *Hnsw) searchLayer(vec gomath.Vector, ep *Item, topCandidates *PriorityQueue, ef int, level uint) error {
	var visited bitset.BitSet

	candidates := &PriorityQueue{}
	candidates.Order = false
	heap.Init(candidates)
	heap.Push(candidates, ep)

	topCandidates.Order = true
	heap.Init(topCandidates)
	heap.Push(topCandidates, ep)

	for candidates.Len() > 0 {

		lowerBound := topCandidates.Top().(*Item).Distance
		candidate := heap.Pop(candidates).(*Item)

		if candidate.Distance > lowerBound {
			break
		}
		for _, node := range xx.NodeList.Nodes[candidate.Node].LinkNodes[level] {
			if !visited.Test(uint(node)) {
				visited.Set(uint(node))
				nodeDist := xx.Space.Distance(vec, xx.NodeList.Nodes[node].Vectors)
				item := &Item{
					Distance: nodeDist,
					Node:     node,
				}
				topDistance := topCandidates.Top().(*Item).Distance

				if topCandidates.Len() < ef {
					if node != ep.Node {
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

func (xx *Hnsw) SelectNeighboursSimple(topCandidates *PriorityQueue, M int) {
	for topCandidates.Len() > M {
		_ = heap.Pop(topCandidates).(*Item)
	}
}

func (xx *Hnsw) SelectNeighboursHeuristic(topCandidates *PriorityQueue, M int, order bool) {
	if topCandidates.Len() < M {
		return
	}

	newCandidates := &PriorityQueue{}
	tmpCandidates := PriorityQueue{}
	tmpCandidates.Order = order
	heap.Init(&tmpCandidates)

	items := make([]*Item, 0, M)

	if !order {
		newCandidates.Order = order
		heap.Init(newCandidates)

		for topCandidates.Len() > 0 {
			item := heap.Pop(topCandidates).(*Item)

			heap.Push(newCandidates, item)
		}
	} else {
		newCandidates = topCandidates
	}

	for newCandidates.Len() > 0 {
		if len(items) >= M {
			break
		}
		item := heap.Pop(newCandidates).(*Item)

		hit := true

		for _, v := range items {

			nodeDist := xx.Space.Distance(
				xx.NodeList.Nodes[v.Node].Vectors,
				xx.NodeList.Nodes[item.Node].Vectors,
			)

			if nodeDist < item.Distance {
				hit = false
				break
			}
		}

		if hit {
			items = append(items, item)
		} else {
			heap.Push(&tmpCandidates, item)
		}
	}

	for len(items) < M && tmpCandidates.Len() > 0 {
		item := heap.Pop(&tmpCandidates).(*Item)
		items = append(items, item)
	}

	for _, item := range items {
		heap.Push(topCandidates, item)
	}
}

func (xx *Hnsw) addConnections(neighbourNode uint32, newNode uint32, level int) {
	var maxConnections int

	if level == 0 {
		maxConnections = int(xx.Mmax0)
	} else {
		maxConnections = int(xx.Mmax)
	}

	xx.NodeList.Nodes[neighbourNode].LinkNodes[level] = append(
		xx.NodeList.Nodes[neighbourNode].LinkNodes[level], newNode)

	curConnections := len(xx.NodeList.Nodes[neighbourNode].LinkNodes[level])

	if curConnections > maxConnections {
		switch xx.Heuristic {
		case false:
			topCandidates := &PriorityQueue{}
			topCandidates.Order = true
			heap.Init(topCandidates)

			for i := 0; i < curConnections; i++ {
				connectedNode := xx.NodeList.Nodes[neighbourNode].LinkNodes[level][i]
				distanceBetweenNodes := xx.Space.Distance(
					xx.NodeList.Nodes[neighbourNode].Vectors,
					xx.NodeList.Nodes[connectedNode].Vectors,
				)
				heap.Push(topCandidates, &Item{
					Node:     connectedNode,
					Distance: distanceBetweenNodes,
				})
			}

			xx.SelectNeighboursSimple(topCandidates, maxConnections)

			xx.NodeList.Nodes[neighbourNode].LinkNodes[level] = make([]uint32, maxConnections)

			for i := maxConnections - 1; i >= 0; i-- {
				node := heap.Pop(topCandidates).(*Item)
				xx.NodeList.Nodes[neighbourNode].LinkNodes[level][i] = node.Node
			}
		case true:
			topCandidates := &PriorityQueue{}
			topCandidates.Order = false
			heap.Init(topCandidates)

			for i := 0; i < curConnections; i++ {
				connectedNode := xx.NodeList.Nodes[neighbourNode].LinkNodes[level][i]
				distanceBetweenNodes := xx.Space.Distance(
					xx.NodeList.Nodes[neighbourNode].Vectors,
					xx.NodeList.Nodes[connectedNode].Vectors,
				)
				heap.Push(topCandidates, &Item{
					Node:     connectedNode,
					Distance: distanceBetweenNodes,
				})
			}

			xx.SelectNeighboursSimple(topCandidates, maxConnections)
			xx.NodeList.Nodes[neighbourNode].LinkNodes[level] = make([]uint32, maxConnections)

			for i := 0; i < maxConnections; i++ {
				node := heap.Pop(topCandidates).(*Item)
				xx.NodeList.Nodes[neighbourNode].LinkNodes[level][i] = node.Node
			}
		}
	}
}

func (xx *Hnsw) findEp(vec gomath.Vector, curObj *Node, layer int16) (match Node, curDist float32, err error) {
	curDist = xx.Space.Distance(vec, curObj.Vectors)
	for level := xx.MaxLevel; level > 0; level-- {
		scan := true

		for scan {
			scan = false

			for _, nodeId := range xx.getConnection(curObj, level) {
				nodeDist := xx.Space.Distance(xx.NodeList.Nodes[nodeId].Vectors, vec)
				if nodeDist < curDist {
					match = xx.NodeList.Nodes[nodeId]
					curDist = nodeDist
					scan = true
				}
			}
		}
	}
	return match, curDist, nil
}

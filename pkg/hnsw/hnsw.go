// hnsw/hnsw.go

package hnsw

import (
	"container/heap"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"

	"github.com/RoaringBitmap/roaring/roaring64"
)

// ------------------------------
// Node Struct
// ------------------------------

// Node represents a node in the HNSW graph.
type Node struct {
	ID        uint64
	Level     int
	Vector    []float32
	Neighbors [][]uint64 // Neighbors per level
}

// NewNode creates a new Node.
func NewNode(id uint64, vector []float32, level, M int) *Node {
	neighbors := make([][]uint64, level+1)
	for i := 0; i <= level; i++ {
		neighbors[i] = make([]uint64, 0, M)
	}
	return &Node{
		ID:        id,
		Vector:    vector,
		Level:     level,
		Neighbors: neighbors,
	}
}

// ------------------------------
// HNSW Struct
// ------------------------------

// MetricType defines the type of metric used.
type MetricType string

const (
	Cosine    MetricType = "cosine"
	Euclidean MetricType = "euclidean"
)

// SearchResult represents a single search result.
type SearchResult struct {
	ID    uint64
	Score float32
}

// HNSW represents the HNSW graph.
type HNSW struct {
	Metric         MetricType
	SimilarityFunc func(a, b []float32) float32
	D              int
	M              int
	EFConstruction int
	LevelMax       int
	EntryPointID   uint64           // 0 indicates no entry point
	Nodes          map[uint64]*Node // Map from node ID to Node
	Probs          []float32        // Probability distribution for level selection
	mu             sync.RWMutex     // To handle concurrent access
	nextID         uint64           // Next node ID to assign (starts from 1)
}

// NewHNSW creates a new HNSW instance.
func NewHNSW(M, efConstruction int, d int, metric MetricType) (*HNSW, error) {
	simFunc, err := getMetricFunction(metric)
	if err != nil {
		return nil, err
	}
	probs := setProbs(M, 1/math.Log(float64(M)))
	levelMax := len(probs) - 1
	return &HNSW{
		Metric:         metric,
		SimilarityFunc: simFunc,
		D:              d,
		M:              M,
		EFConstruction: efConstruction,
		LevelMax:       levelMax,
		EntryPointID:   0, // 0 indicates no entry point
		Nodes:          make(map[uint64]*Node),
		Probs:          probs,
		nextID:         1, // Start IDs from 1
	}, nil
}

// getMetricFunction returns the appropriate similarity function based on the metric.
func getMetricFunction(metric MetricType) (func(a, b []float32) float32, error) {
	switch metric {
	case Cosine:
		return CosineSimilarity, nil
	case Euclidean:
		return EuclideanSimilarity, nil
	default:
		return nil, errors.New("invalid metric")
	}
}

// setProbs initializes the probability distribution for level selection.
func setProbs(M int, levelMult float64) []float32 {
	var level int
	var probs []float32
	for {
		prob := math.Exp(-float64(level)/levelMult) * (1 - math.Exp(-1/levelMult))
		if prob < 1e-9 {
			break
		}
		probs = append(probs, float32(prob))
		level++
	}
	return probs
}

// SelectLevel randomly selects a level based on the probability distribution.
func (h *HNSW) SelectLevel() int {
	r := rand.Float32()
	for i, p := range h.Probs {
		if r < p {
			return i
		}
		r -= p
	}
	return len(h.Probs) - 1
}

// AddNodeToGraph adds a node to the HNSW graph.
// Assumes that the caller has already acquired the necessary lock.
func (h *HNSW) AddNodeToGraph(node *Node) {
	if h.EntryPointID == 0 { // 0 indicates no entry point
		h.EntryPointID = node.ID
		h.Nodes[node.ID] = node
		return
	}

	currentNode := h.Nodes[h.EntryPointID]
	closestNode := currentNode

	for level := h.LevelMax; level >= 0; level-- {
		for {
			nextNode, maxSim := h.findMostSimilar(node, currentNode, level)
			if nextNode != nil && maxSim > h.SimilarityFunc(node.Vector, closestNode.Vector) {
				currentNode = nextNode
				closestNode = currentNode
			} else {
				break
			}
		}
	}

	closestLevel := min(node.Level, closestNode.Level)

	for level := 0; level <= closestLevel; level++ {
		// Add node to closestNode's neighbors
		closestNode.Neighbors[level] = append(closestNode.Neighbors[level], node.ID)
		if len(closestNode.Neighbors[level]) > h.M {
			// Remove the farthest neighbor
			h.removeFarthest(closestNode, level, node.Vector)
		}

		// Add closestNode to node's neighbors
		node.Neighbors[level] = append(node.Neighbors[level], closestNode.ID)
		if len(node.Neighbors[level]) > h.M {
			h.removeFarthest(node, level, closestNode.Vector)
		}
	}

	h.Nodes[node.ID] = node

	if node.Level > h.LevelMax {
		h.LevelMax = node.Level

		// Extend all existing nodes' Neighbors slices to accommodate the new LevelMax
		for _, existingNode := range h.Nodes {
			for len(existingNode.Neighbors) <= h.LevelMax {
				existingNode.Neighbors = append(existingNode.Neighbors, []uint64{})
			}
		}

		// Update EntryPointID to the new node
		h.EntryPointID = node.ID
	}
}

// findMostSimilar finds the most similar neighbor at a given level.
func (h *HNSW) findMostSimilar(target *Node, current *Node, level int) (*Node, float32) {
	var bestNode *Node
	maxSim := float32(-math.MaxFloat32)

	neighbors := current.Neighbors[level]
	for _, neighborID := range neighbors {
		if neighborID == 0 {
			continue
		}
		neighbor, exists := h.Nodes[neighborID]
		if !exists {
			continue
		}
		sim := h.SimilarityFunc(target.Vector, neighbor.Vector)
		if sim > maxSim {
			maxSim = sim
			bestNode = neighbor
		}
	}
	return bestNode, maxSim
}

// removeFarthest removes the farthest neighbor based on similarity.
func (h *HNSW) removeFarthest(node *Node, level int, referenceVector []float32) {
	if len(node.Neighbors[level]) <= h.M {
		return
	}
	// Find the neighbor with the lowest similarity and remove it
	minSim := float32(math.MaxFloat32)
	minIdx := -1
	for i, neighborID := range node.Neighbors[level] {
		neighbor, exists := h.Nodes[neighborID]
		if !exists {
			continue
		}
		sim := h.SimilarityFunc(referenceVector, neighbor.Vector)
		if sim < minSim {
			minSim = sim
			minIdx = i
		}
	}
	if minIdx != -1 {
		node.Neighbors[level] = append(node.Neighbors[level][:minIdx], node.Neighbors[level][minIdx+1:]...)
	}
}

// AddPoint adds a new point to the HNSW graph with an automatically assigned ID.
func (h *HNSW) AddPoint(vector []float32) (uint64, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	id := h.nextID
	h.nextID++

	if h.D != 0 && len(vector) != h.D {
		return 0, fmt.Errorf("all vectors must be of the same dimension")
	}
	h.D = len(vector)

	level := h.SelectLevel()
	node := NewNode(id, vector, level, h.M)
	h.AddNodeToGraph(node)
	return id, nil
}

// DeletePoint removes a point from the HNSW index.
func (h *HNSW) DeletePoint(id uint64) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	node, exists := h.Nodes[id]
	if !exists {
		return fmt.Errorf("node with ID %d does not exist", id)
	}

	// Remove this node from its neighbors
	for lvl, neighbors := range node.Neighbors {
		for _, neighborID := range neighbors {
			if neighborID == 0 {
				continue
			}
			neighbor, exists := h.Nodes[neighborID]
			if !exists {
				continue
			}
			// Remove node.id from neighbor's neighbors at this level
			for j, nid := range neighbor.Neighbors[lvl] {
				if nid == id {
					neighbor.Neighbors[lvl] = append(neighbor.Neighbors[lvl][:j], neighbor.Neighbors[lvl][j+1:]...)
					break
				}
			}
		}
	}

	// Finally, remove the node
	delete(h.Nodes, id)

	// If the deleted node was the entry point, choose a new entry point
	if h.EntryPointID == id {
		if len(h.Nodes) > 0 {
			for newEntryID := range h.Nodes {
				h.EntryPointID = newEntryID
				break
			}
		} else {
			h.EntryPointID = 0 // 0 indicates no entry point
		}
	}

	return nil
}

// SearchKNN searches for the k nearest neighbors to the query vector, excluding nodes in the filter.
func (h *HNSW) SearchKNN(query []float32, k int, filter *roaring64.Bitmap) ([]SearchResult, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.EntryPointID == 0 {
		return nil, errors.New("the index is empty")
	}

	// Initialize the priority queue
	candidates := &PriorityQueue{}
	heap.Init(candidates)
	initialScore := h.SimilarityFunc(query, h.Nodes[h.EntryPointID].Vector)
	heap.Push(candidates, &Item{id: h.EntryPointID, score: initialScore})

	result := make([]SearchResult, 0, k)
	visited := make(map[uint64]bool)

	for candidates.Len() > 0 && len(result) < k {
		currentItem := heap.Pop(candidates).(*Item)
		currentID := currentItem.id

		if visited[currentID] {
			continue
		}
		visited[currentID] = true

		if filter != nil && filter.Contains(currentID) {
			continue // Skip nodes in the filter
		}

		currentNode := h.Nodes[currentID]
		similarity := h.SimilarityFunc(currentNode.Vector, query)

		if similarity > 0 {
			result = append(result, SearchResult{
				ID:    currentID,
				Score: similarity,
			})
		}

		if currentNode.Level == 0 {
			continue
		}

		for level := currentNode.Level; level >= 0; level-- {
			for _, neighborID := range currentNode.Neighbors[level] {
				if !visited[neighborID] && (filter == nil || !filter.Contains(neighborID)) {
					neighbor := h.Nodes[neighborID]
					heap.Push(candidates, &Item{
						id:    neighbor.ID,
						score: h.SimilarityFunc(query, neighbor.Vector),
					})
				}
			}
		}
	}

	// Sort the result by score in descending order
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	if len(result) > k {
		result = result[:k]
	}

	return result, nil
}

// Fit optimizes the HNSW index.
func (h *HNSW) Fit() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, node := range h.Nodes {
		for lvl := 0; lvl <= node.Level; lvl++ {
			sort.Slice(node.Neighbors[lvl], func(i, j int) bool {
				return node.Neighbors[lvl][i] < node.Neighbors[lvl][j]
			})
		}
	}

	return nil
}

// Flush saves the HNSW index to storage.
func (h *HNSW) Flush() error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// TODO
	fmt.Println("TODO FLUSH")
	return nil
}

// SizeInMemory returns the approximate memory size of the HNSW index.
func (h *HNSW) SizeInMemory() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var size int64
	for _, node := range h.Nodes {
		size += 8                           // ID (uint64)
		size += 4 * int64(len(node.Vector)) // Vector elements (float32)
		for _, neighbors := range node.Neighbors {
			size += 8 * int64(len(neighbors)) // Each neighbor is uint64
		}
	}
	return size
}

func (h *HNSW) UpdateStorage(storage interface{}) {
	fmt.Println("TODO")
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

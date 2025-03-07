package edge

import (
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/sjy-dv/coltt/gen/protoc/v4/edgepb"
	"github.com/sjy-dv/coltt/pkg/inverted"
)

type vectorspace interface {
	ChangedVertex(updateID string, Id uint64, edge ENode) error
	RemoveVertex(dropFilter map[string]interface{}) error
	VertexSearch(target Vector, topK int, highCpu bool) (
		[]*SearchResultItem, error)
	FilterableVertexSearch(filter *inverted.FilterExpression, target Vector, topK int, highCpu bool) (
		[]*SearchResultItem, error)
	SaveVertexMetadata() ([]byte, error)
	LoadVertexMetadata(collectionName string, data []byte) error
	SaveVertexInverted() ([]byte, error)
	LoadVertexInverted(data []byte) error
	SaveVertex() ([]byte, error)
	LoadVertex(data []byte) error
	Quantization() edgepb.Quantization
	Distance() edgepb.Distance
	Dim() uint32
	LoadSize() int64
	Indexer() map[string]IndexFeature
	Versional() bool
}

type Vectorstore struct {
	Space map[string]vectorspace
	slock sync.RWMutex
}

func NewVectorstore() *Vectorstore {
	return &Vectorstore{
		Space: make(map[string]vectorspace),
	}
}

func (vs *Vectorstore) CreateCollection(collectionName string, metadata Metadata) error {
	vs.slock.RLock()
	_, ok := vs.Space[collectionName]
	vs.slock.RUnlock()
	if ok {
		return fmt.Errorf(ErrCollectionExists, collectionName)
	}
	var vectorstore vectorspace
	if edgepb.Quantization(metadata.Quantization) == edgepb.Quantization_F8 {
		vectorstore = newF8Vectorstore(collectionName, metadata)
	} else if edgepb.Quantization(metadata.Quantization) == edgepb.Quantization_F16 {
		vectorstore = newF16Vectorstore(collectionName, metadata)
	} else if edgepb.Quantization(metadata.Quantization) == edgepb.Quantization_BF16 {
		vectorstore = newBF16Vectorstore(collectionName, metadata)
	} else if edgepb.Quantization(metadata.Quantization) == edgepb.Quantization_None {
		vectorstore = newNoneVectorstore(collectionName, metadata)
	} else {
		return errors.New("not support quantization type")
	}
	vs.slock.Lock()
	vs.Space[collectionName] = vectorstore
	vs.slock.Unlock()
	return nil
}

func (vs *Vectorstore) Quantization(collectionName string) edgepb.Quantization {
	return vs.Space[collectionName].Quantization()
}
func (vs *Vectorstore) Distance(collectionName string) edgepb.Distance {
	return vs.Space[collectionName].Distance()
}
func (vs *Vectorstore) Dim(collectionName string) uint32 {
	return vs.Space[collectionName].Dim()
}
func (vs *Vectorstore) LoadSize(collectionName string) int64 {
	return vs.Space[collectionName].LoadSize()
}
func (vs *Vectorstore) Indexer(collectionName string) map[string]IndexFeature {
	return vs.Space[collectionName].Indexer()
}

func (vs *Vectorstore) Versional(collectionName string) bool {
	return vs.Space[collectionName].Versional()
}

func (vs *Vectorstore) SavedMetadata(collectionName string) ([]byte, error) {
	return vs.Space[collectionName].SaveVertexMetadata()
}

func (vs *Vectorstore) SavedVertex(collectionName string) ([]byte, error) {
	return vs.Space[collectionName].SaveVertex()
}

func (vs *Vectorstore) SavedInverted(collectionName string) ([]byte, error) {
	return vs.Space[collectionName].SaveVertexInverted()
}

func (vs *Vectorstore) LoadedMetadata(collectionName string, data []byte) error {
	return vs.Space[collectionName].LoadVertexMetadata(collectionName, data)
}

func (vs *Vectorstore) LoadedVertex(collectionName string, data []byte) error {
	return vs.Space[collectionName].LoadVertex(data)
}

func (vs *Vectorstore) LoadedInverted(collectionName string, data []byte) error {
	return vs.Space[collectionName].LoadVertexInverted(data)
}

func (vs *Vectorstore) DestroySpace(collectionName string) {
	vs.slock.Lock()
	delete(vs.Space, collectionName)
	vs.slock.Unlock()
}

func (vs *Vectorstore) ChangedVertex(collectioName string, updateID string, Id uint64, metadata map[string]interface{}, vector Vector) error {
	newVertex := ENode{
		Vector:   vector,
		Metadata: metadata,
	}
	return vs.Space[collectioName].ChangedVertex(updateID, Id, newVertex)
}

func (vs *Vectorstore) RemoveVertex(collectionName string, dropfilter map[string]interface{}) error {
	return vs.Space[collectionName].RemoveVertex(dropfilter)
}

func (vs *Vectorstore) VertexSearch(collectioName string, topK uint64, vector Vector, highCpu bool) ([]*SearchResultItem, error) {
	return vs.Space[collectioName].VertexSearch(vector, int(topK), highCpu)
}

func (vs *Vectorstore) FilterableVertexSearch(collectioName string, filter *inverted.FilterExpression, topK uint64, vector Vector, highCpu bool) ([]*SearchResultItem, error) {
	return vs.Space[collectioName].FilterableVertexSearch(filter, vector, int(topK), highCpu)
}

func (vs *Vectorstore) FillEmpty(collectionName string) {
	vs.slock.Lock()
	vs.Space[collectionName] = &noneVecSpace{}
	vs.slock.Unlock()
}

func Normalize(v []float32) []float32 {
	var norm float32
	out := make([]float32, len(v))
	for i := range v {
		norm += v[i] * v[i]
	}
	if norm == 0 {
		return out
	}

	norm = float32(math.Sqrt(float64(norm)))
	for i := range v {
		out[i] = v[i] / norm
	}

	return out
}

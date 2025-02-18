package experimental

import (
	"errors"
	"fmt"
	"sync"

	"github.com/sjy-dv/coltt/edge"
	"github.com/sjy-dv/coltt/gen/protoc/v3/experimentalproto"
)

type vectorVertex interface {
	InsertVertex(collectionName string, Id string, edge VertexEdge) error
	UpdateVertex(collectionName string, Id string, edge VertexEdge) error
	RemoveVertex(collectionName string, Id string) error
	SaveVertexMetadata() ([]byte, error)
	LoadVertexMetadata(collectionName string, data []byte) error
	SaveVertex() ([]byte, error)
	LoadVertex(data []byte) error
	Quantization() experimentalproto.Quantization
	Distance() experimentalproto.Distance
	Dim() uint32
	LoadSize() int64
	Indexer() map[string]IndexFeature
	Versional() bool
}

type MultiVectorSpace struct {
	Space   map[string]vectorVertex
	mvsLock sync.RWMutex
}

func NewMultiVectorSpace() *MultiVectorSpace {
	return &MultiVectorSpace{
		Space: make(map[string]vectorVertex),
	}
}

func (multiSpace *MultiVectorSpace) CreateCollection(collectionName string, metadata Metadata) error {
	multiSpace.mvsLock.RLock()
	_, ok := multiSpace.Space[collectionName]
	multiSpace.mvsLock.RUnlock()
	if ok {
		return fmt.Errorf(edge.ErrCollectionExists, collectionName)
	}

	var initvectorVertex vectorVertex

	if experimentalproto.Quantization(metadata.Quantization) == experimentalproto.Quantization_None {
		initvectorVertex = newMultiVectorVertex(collectionName, metadata)
	} else {
		return errors.New("not support quantization type")
	}
	multiSpace.mvsLock.Lock()
	multiSpace.Space[collectionName] = initvectorVertex
	multiSpace.mvsLock.Unlock()
	return nil
}

func (multiSpace *MultiVectorSpace) Quantization(collectionName string) experimentalproto.Quantization {
	return multiSpace.Space[collectionName].Quantization()
}
func (multiSpace *MultiVectorSpace) Distance(collectionName string) experimentalproto.Distance {
	return multiSpace.Space[collectionName].Distance()
}
func (multiSpace *MultiVectorSpace) Dim(collectionName string) uint32 {
	return multiSpace.Space[collectionName].Dim()
}
func (multiSpace *MultiVectorSpace) LoadSize(collectionName string) int64 {
	return multiSpace.Space[collectionName].LoadSize()
}
func (multiSpace *MultiVectorSpace) Indexer(collectionName string) map[string]IndexFeature {
	return multiSpace.Space[collectionName].Indexer()
}

func (multiSpace *MultiVectorSpace) Versional(collectionName string) bool {
	return multiSpace.Space[collectionName].Versional()
}

func (multiSpace *MultiVectorSpace) SavedMetadata(collectionName string) ([]byte, error) {
	return multiSpace.Space[collectionName].SaveVertexMetadata()
}

func (multiSpace *MultiVectorSpace) SavedVertex(collectionName string) ([]byte, error) {
	return multiSpace.Space[collectionName].SaveVertex()
}

func (multiSpace *MultiVectorSpace) LoadedMetadata(collectionName string, data []byte) error {
	return multiSpace.Space[collectionName].LoadVertexMetadata(collectionName, data)
}

func (multiSpace *MultiVectorSpace) LoadedVertex(collectionName string, data []byte) error {
	return multiSpace.Space[collectionName].LoadVertex(data)
}

func (multiSpace *MultiVectorSpace) DestroySpace(collectionName string) {
	multiSpace.mvsLock.Lock()
	delete(multiSpace.Space, collectionName)
	multiSpace.mvsLock.Unlock()
}

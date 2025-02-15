package experimental

import (
	"sync"

	"github.com/sjy-dv/coltt/gen/protoc/v3/experimentalproto"
)

type metaStorage struct {
	mem  map[string]Metadata
	lock sync.RWMutex
}

func NewMetaStorage() *metaStorage {
	return &metaStorage{
		mem: make(map[string]Metadata),
	}
}

func (storage *metaStorage) Set(k string, v Metadata) {
	storage.lock.Lock()
	defer storage.lock.Unlock()
	storage.mem[k] = v
}

func (storage *metaStorage) Del(k string) {
	storage.lock.Lock()
	defer storage.lock.Unlock()
	delete(storage.mem, k)
}

func (storage *metaStorage) Dim(col string) uint32 {
	storage.lock.RLock()
	defer storage.lock.RUnlock()
	return storage.mem[col].dim
}

func (storage *metaStorage) Distance(col string) experimentalproto.Distance {
	storage.lock.RLock()
	defer storage.lock.RUnlock()
	return experimentalproto.Distance(storage.mem[col].distance)
}

func (storage *metaStorage) Quantization(col string) experimentalproto.Quantization {
	storage.lock.RLock()
	defer storage.lock.RUnlock()
	return experimentalproto.Quantization(storage.mem[col].quantization)
}

func (storage *metaStorage) IndexFeatures(col, index string) IndexFeature {
	storage.lock.RLock()
	copyIndexType := storage.mem[col].indexType[index]
	defer storage.lock.RUnlock()
	return copyIndexType
}

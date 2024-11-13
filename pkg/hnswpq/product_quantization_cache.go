package hnswpq

import (
	"fmt"
	"sync"

	"github.com/sjy-dv/nnv/edge"
)

type ProductQuantizerCache struct {
	caches map[uint64]*itemCacheElem
	cLock  sync.RWMutex
}

type itemCacheElem struct {
	value   *productQuantizedPoint
	IsDirty bool
}

func (xx *ProductQuantizerCache) Get(id uint64) (*productQuantizedPoint, error) {
	xx.cLock.RLock()
	defer xx.cLock.RUnlock()
	if item, ok := xx.caches[id]; ok {
		return item.value, nil
	}
	return nil, fmt.Errorf(edge.TargetIdNotFound, id)
}

func (xx *ProductQuantizerCache) Put(id uint64, point *productQuantizedPoint) {
	xx.cLock.Lock()
	defer xx.cLock.Unlock()
	xx.caches[id] = &itemCacheElem{value: point, IsDirty: true}
}

func (xx *ProductQuantizerCache) Delete(ids ...uint64) error {
	xx.cLock.Lock()
	defer xx.cLock.Unlock()
	for _, id := range ids {
		xx.caches[id] = nil
		delete(xx.caches, id)
	}
	return nil
}

func (xx *ProductQuantizerCache) Dirty(id uint64) {
	xx.cLock.Lock()
	xx.caches[id].IsDirty = true
	defer xx.cLock.Unlock()
	return
}

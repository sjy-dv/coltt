package cache

import (
	"errors"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/vemoo/storage"
)

var ErrNotFound = errors.New("not found")

type Storable[K comparable, T any] interface {
	Cachable
	IdFromKey(key []byte) (K, bool)
	CheckAndClearDirty() bool
	ReadFrom(id K, storage storage.Storage) (T, error)
	WriteTo(id K, storage storage.Storage) error
	DeleteFrom(id K, storage storage.Storage) error
}

type itemCacheElem[K comparable, V Storable[K, V]] struct {
	value     V
	IsDirty   bool
	IsDeleted bool
}

type ItemCache[K comparable, V Storable[K, V]] struct {
	items        map[K]*itemCacheElem[K, V]
	itemsMu      sync.Mutex
	isAllInCache bool
	storage      storage.Storage
}

func NewItemCache[K comparable, V Storable[K, V]](storage storage.Storage) *ItemCache[K, V] {
	ic := &ItemCache[K, V]{
		items:   make(map[K]*itemCacheElem[K, V]),
		storage: storage,
	}
	return ic
}

func (self *ItemCache[K, T]) SizeInMemory() int64 {
	self.itemsMu.Lock()
	defer self.itemsMu.Unlock()
	for _, item := range self.items {
		return int64(len(self.items)) * item.value.SizeInMemory()
	}
	return 0
}

func (self *ItemCache[K, T]) UpdateStorage(storage storage.Storage) {
	self.storage = storage
}

func (self *ItemCache[K, T]) read(id K) (T, error) {
	var dummyValue T
	value, err := dummyValue.ReadFrom(id, self.storage)
	if err != nil {
		return value, err
	}
	item := &itemCacheElem[K, T]{
		value: value,
	}
	self.items[id] = item
	return item.value, nil
}

func (self *ItemCache[K, V]) Get(id K) (value V, err error) {
	self.itemsMu.Lock()
	defer self.itemsMu.Unlock()
	if item, ok := self.items[id]; ok {
		if item.IsDeleted {
			err = ErrNotFound
			return
		}
		return item.value, nil
	}
	value, err = self.read(id)
	return
}

func (self *ItemCache[K, V]) GetMany(ids ...K) ([]V, error) {
	self.itemsMu.Lock()
	defer self.itemsMu.Unlock()
	values := make([]V, 0, len(ids))
	for _, id := range ids {
		if item, ok := self.items[id]; ok {
			if item.IsDeleted {
				continue
			}
			values = append(values, item.value)
			continue
		}
		if item, err := self.read(id); err != nil && err != ErrNotFound {
			return nil, err
		} else if err == nil {
			values = append(values, item)
		}
	}
	return values, nil
}

func (self *ItemCache[K, T]) Count() int {
	self.itemsMu.Lock()
	defer self.itemsMu.Unlock()
	bucketCount := 0
	err := self.storage.ForEach(func(key, value []byte) error {
		var dummyValue T
		id, ok := dummyValue.IdFromKey(key)
		if !ok {
			return nil
		}
		if _, ok := self.items[id]; !ok {
			bucketCount++
		}
		return nil
	})
	if err != nil {
		log.Warn().Err(err).Msg("error counting item cache items in bucket")
		return 0
	}
	cacheCount := 0
	for _, item := range self.items {
		if !item.IsDeleted {
			cacheCount++
		}
	}
	return cacheCount + bucketCount
}

func (self *ItemCache[K, T]) Put(id K, item T) {
	self.itemsMu.Lock()
	defer self.itemsMu.Unlock()
	self.items[id] = &itemCacheElem[K, T]{value: item, IsDirty: true}
}

func (self *ItemCache[K, T]) Delete(ids ...K) error {
	self.itemsMu.Lock()
	defer self.itemsMu.Unlock()
	for _, id := range ids {
		if elem, ok := self.items[id]; ok {
			elem.IsDeleted = true
			continue
		}
		_, err := self.read(id)
		if err == ErrNotFound {
			continue
		}
		if err != nil {
			return err
		}
		self.items[id].IsDeleted = true
	}
	return nil
}

func (self *ItemCache[K, T]) ForEach(fn func(id K, item T) error) error {
	self.itemsMu.Lock()
	defer self.itemsMu.Unlock()
	if !self.isAllInCache {
		err := self.storage.ForEach(func(key, value []byte) error {
			var dummyValue T
			id, ok := dummyValue.IdFromKey(key)
			if !ok {
				return nil
			}
			if _, ok := self.items[id]; ok {
				return nil
			}
			if _, err := self.read(id); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		self.isAllInCache = true
	}
	for id, item := range self.items {
		if item.IsDeleted {
			continue
		}
		if err := fn(id, item.value); err != nil {
			return err
		}
	}
	// ---------------------------
	return nil
}

func (self *ItemCache[K, T]) Flush() error {
	self.itemsMu.Lock()
	defer self.itemsMu.Unlock()
	for id, item := range self.items {
		if item.IsDeleted {
			if err := item.value.DeleteFrom(id, self.storage); err != nil {
				return err
			}
			delete(self.items, id)
			continue
		}
		if item.IsDirty || item.value.CheckAndClearDirty() {
			if err := item.value.WriteTo(id, self.storage); err != nil {
				return err
			}
			item.IsDirty = false
		}
	}
	return nil
}

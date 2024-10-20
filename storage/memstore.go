package storage

import (
	"bytes"
	"cmp"
	"fmt"
	"slices"
	"sync"
)

type memeStorage struct {
	data       map[string][]byte
	isReadOnly bool
}

func NewMemStorage(isReadOnly bool) Storage {
	return &memeStorage{
		data:       make(map[string][]byte),
		isReadOnly: isReadOnly,
	}
}

func (b *memeStorage) IsReadOnly() bool {
	return b.isReadOnly
}

func (b *memeStorage) Get(k []byte) []byte {
	return b.data[string(k)]
}

func (b *memeStorage) Put(k, v []byte) error {
	if b.isReadOnly {
		return fmt.Errorf("cannot put into read-only memory bucket")
	}
	b.data[string(k)] = v
	return nil
}

func (b *memeStorage) ForEach(f func(k, v []byte) error) error {
	for k, v := range b.data {
		if err := f([]byte(k), v); err != nil {
			return err
		}
	}
	return nil
}

func (b *memeStorage) PrefixScan(prefix []byte, f func(k, v []byte) error) error {
	for k, v := range b.data {
		if len(k) < len(prefix) {
			continue
		}
		if k[:len(prefix)] == string(prefix) {
			if err := f([]byte(k), v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *memeStorage) RangeScan(start, end []byte, inclusive bool, f func(k, v []byte) error) error {
	// The data neeself to be ordered first
	type pair struct {
		k string
		v []byte
	}
	orderedData := make([]pair, 0, len(b.data))
	for k, v := range b.data {
		orderedData = append(orderedData, pair{k, v})
	}
	slices.SortFunc(orderedData, func(a, b pair) int {
		return cmp.Compare(a.k, b.k)
	})
	for _, p := range orderedData {
		if start != nil {
			if inclusive {
				if bytes.Compare([]byte(p.k), start) < 0 {
					continue
				}
			} else {
				if bytes.Compare([]byte(p.k), start) <= 0 {
					continue
				}
			}
		}
		if end != nil {
			if inclusive {
				if bytes.Compare([]byte(p.k), end) > 0 {
					break
				}
			} else {
				if bytes.Compare([]byte(p.k), end) >= 0 {
					break
				}
			}
		}
		if err := f([]byte(p.k), p.v); err != nil {
			return err
		}
	}
	return nil
}

func (b *memeStorage) Delete(k []byte) error {
	if b.isReadOnly {
		return fmt.Errorf("cannot delete in a read-only memory bucket")
	}
	delete(b.data, string(k))
	return nil
}

type memeStorageManager struct {
	storages   map[string]map[string][]byte
	isReadOnly bool
	mu         sync.Mutex
}

func (self *memeStorageManager) Get(storageName string) (Storage, error) {
	self.mu.Lock()
	defer self.mu.Unlock()
	b, ok := self.storages[storageName]
	if !ok {
		b = make(map[string][]byte)
		self.storages[storageName] = b
	}
	mb := &memeStorage{
		data:       b,
		isReadOnly: self.isReadOnly,
	}
	return mb, nil
}

func (self *memeStorageManager) Delete(storageName string) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.isReadOnly {
		return fmt.Errorf("cannot delete %s in a read-only memory bucket manager", storageName)
	}
	delete(self.storages, storageName)
	return nil
}

type memDiskStore struct {
	storages map[string]map[string][]byte
	// This lock is used to give a consistent view of the store such that Write
	// does not interleave with any Read.
	mu sync.RWMutex
}

func newMemDiskStore() *memDiskStore {
	return &memDiskStore{
		storages: make(map[string]map[string][]byte),
	}
}

func (self *memDiskStore) Path() string {
	return "memory"
}

func (self *memDiskStore) Read(f func(StorageCoordinator) error) error {
	self.mu.RLock()
	defer self.mu.RUnlock()
	ms := &memeStorageManager{
		storages:   self.storages,
		isReadOnly: true,
	}
	return f(ms)
}

func (self *memDiskStore) Write(f func(StorageCoordinator) error) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	ms := &memeStorageManager{
		storages:   self.storages,
		isReadOnly: false,
	}
	return f(ms)
}

func (self *memDiskStore) BackupToFile(path string) error {
	return fmt.Errorf("not supported")
}

func (self *memDiskStore) SizeInBytes() (int64, error) {
	return 0, nil
}

func (self *memDiskStore) Close() error {
	clear(self.storages)
	return nil
}

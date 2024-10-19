package storage

import (
	"bytes"
	"cmp"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"slices"
	"sync"

	"github.com/sjy-dv/vemoo/pkg/flate"
)

var readOnlyConstraints error = errors.New("shard is read-only")

type compressionMemStore struct {
	data       map[string][]byte
	isReadOnly bool
}

func NewCompressionMemStore(isReadOnly bool) Storage {
	return &compressionMemStore{
		data:       make(map[string][]byte),
		isReadOnly: isReadOnly,
	}
}

func (self *compressionMemStore) IsReadOnly() bool {
	return self.isReadOnly
}

func (self *compressionMemStore) Get(k []byte) []byte {
	return self.data[string(k)]
}

func (self *compressionMemStore) Put(k, v []byte) error {
	if self.isReadOnly {
		return readOnlyConstraints
	}
	self.data[string(k)] = v
	return nil
}

func (self *compressionMemStore) ForEach(f func(k, v []byte) error) error {
	for k, v := range self.data {
		if err := f([]byte(k), v); err != nil {
			return err
		}
	}
	return nil
}

func (self *compressionMemStore) PrefixScan(prefix []byte, f func(k, v []byte) error) error {
	for k, v := range self.data {
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

func (self *compressionMemStore) RangeScan(start, end []byte, inclusive bool, f func(k, v []byte) error) error {
	type row struct {
		k string
		v []byte
	}
	aggregates := make([]row, 0, len(self.data))
	for k, v := range self.data {
		aggregates = append(aggregates, row{k, v})
	}
	slices.SortFunc(aggregates, func(x, y row) int {
		return cmp.Compare(x.k, y.k)
	})
	for _, p := range aggregates {
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

func (self *compressionMemStore) Delete(k []byte) error {
	if self.isReadOnly {
		return readOnlyConstraints
	}
	delete(self.data, string(k))
	return nil
}

type compressionMemStoreCoordinator struct {
	storages   map[string]map[string][]byte
	isReadOnly bool
	mu         sync.RWMutex
}

func (self *compressionMemStoreCoordinator) Get(storageName string) (Storage, error) {
	self.mu.Lock()
	defer self.mu.Unlock()
	storage, ok := self.storages[storageName]
	if !ok {
		storage = make(map[string][]byte)
		self.storages[storageName] = storage
	}
	cm := &compressionMemStore{
		data:       storage,
		isReadOnly: self.isReadOnly,
	}
	return cm, nil
}

func (self *compressionMemStoreCoordinator) Delete(storageName string) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.isReadOnly {
		return fmt.Errorf("coordinator is readonly constraints, delete failed %s", storageName)
	}
	delete(self.storages, storageName)
	return nil
}

type compressionCdat struct {
	storages map[string]map[string][]byte
	mu       sync.RWMutex
	path     string
}

func newCompressionCDat(path string) (*compressionCdat, error) {
	instance := &compressionCdat{
		storages: make(map[string]map[string][]byte),
		path:     path,
	}
	if path != "" {
		err := instance.loadFromCache(path)
		if err != nil {
			return nil, err
		}
	}
	return instance, nil
}

func (self *compressionCdat) Path() string {
	return self.path
}

func (self *compressionCdat) Read(f func(StorageCoordinator) error) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	cooridnator := &compressionMemStoreCoordinator{
		storages:   self.storages,
		isReadOnly: true,
	}
	return f(cooridnator)
}

func (self *compressionCdat) Write(f func(StorageCoordinator) error) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	coordinator := &compressionMemStoreCoordinator{
		storages:   self.storages,
		isReadOnly: false,
	}
	return f(coordinator)
}

func (self *compressionCdat) BackupToFile(path string) error {
	self.mu.RLock()
	defer self.mu.RUnlock()

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	compressor, err := flate.NewWriter(file, flate.BestCompression, nil)
	if err != nil {
		return err
	}
	defer compressor.Close()

	encoder := gob.NewEncoder(compressor)
	if err := encoder.Encode(self.storages); err != nil {
		return err
	}
	return nil
}

func (self *compressionCdat) Flush() error {
	return self.syncDisk()
}

func (self *compressionCdat) loadFromCache(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = self.syncDisk()
			if err != nil {
				return fmt.Errorf("failed to create new cache file: %v", err)
			}
		}
		return fmt.Errorf("failed to open cache file: %v", err)
	}
	defer file.Close()

	decompressor := flate.NewReader(file, nil)
	defer decompressor.Close()

	decoder := gob.NewDecoder(decompressor)
	if err := decoder.Decode(&self.storages); err != nil {
		return fmt.Errorf("failed to decode cache data: %v", err)
	}
	return nil
}

func (self *compressionCdat) SizeInBytes() (int64, error) {
	info, err := os.Stat(self.path)
	if err != nil {
		return 0, errors.New("SizeInBytes os.Stat function error : " + err.Error())
	}
	return info.Size(), nil
}

func (self *compressionCdat) Close() error {
	return self.syncDisk()
}

func (self *compressionCdat) syncDisk() error {
	if self.path == "" {
		return nil
	}
	file, err := os.Create(self.path)
	if err != nil {
		return err
	}
	defer file.Close()
	compressor, err := flate.NewWriter(file, flate.BestCompression, nil)
	if err != nil {
		return err
	}
	defer compressor.Close()

	encoder := gob.NewEncoder(compressor)
	if err := encoder.Encode(self.storages); err != nil {
		return err
	}
	return nil
}

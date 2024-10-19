package storage

import (
	"errors"
	"fmt"
	"time"

	"go.etcd.io/bbolt"
)

const defaultPath = "./.vcache"

type nullReadOnlyStorage struct{}

func (nullReadOnlyStorage) IsReadOnly() bool {
	return true
}

func (nullReadOnlyStorage) Get(k []byte) []byte {
	return nil
}

func (nullReadOnlyStorage) ForEach(f func(k, v []byte) error) error {
	return nil
}

func (nullReadOnlyStorage) PrefixScan(prefix []byte, f func(k, v []byte) error) error {
	return nil
}

func (nullReadOnlyStorage) RangeScan(start, end []byte, inclusive bool, f func(k, v []byte) error) error {
	return nil
}

func (nullReadOnlyStorage) Put(k, v []byte) error {
	return errors.New("cannot put into empty read-only storage")
}

func (nullReadOnlyStorage) Delete(k []byte) error {
	return errors.New("cannot delete from empty read-only storage")
}

type ReadOnlyStorage interface {
	Get([]byte) []byte
	ForEach(func(k, v []byte) error) error
	PrefixScan(prefix []byte, f func(k, v []byte) error) error
	RangeScan(start, end []byte, inclusive bool, f func(k, v []byte) error) error
}

type Storage interface {
	ReadOnlyStorage
	IsReadOnly() bool
	Put([]byte, []byte) error
	Delete([]byte) error
}

type StorageCoordinator interface {
	Get(storageName string) (Storage, error)
	Delete(storageName string) error
}

type StorageLayer interface {
	Path() string
	Read(f func(StorageCoordinator) error) error
	Write(f func(StorageCoordinator) error) error
	BackupToFile(path string) error
	SizeInBytes() (int64, error)
	Close() error
}

func Open(path string, stable bool) (StorageLayer, error) {
	if stable {
		db, err := bbolt.Open(path, 0644, &bbolt.Options{Timeout: 1 * time.Minute})
		if err != nil {
			return nil, fmt.Errorf("open db failed %s: %w", path, err)
		}
		return openDiskStore{db: db}, nil
	} else {
		return newCompressionCDat(path)
	}
}

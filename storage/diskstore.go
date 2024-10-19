package storage

import (
	"bytes"
	"fmt"
	"sync"

	"go.etcd.io/bbolt"
)

type diskStore struct {
	disk *bbolt.Bucket
}

func (self diskStore) IsReadOnly() bool {
	return !self.disk.Writable()
}

func (self diskStore) Get(k []byte) []byte {
	return self.disk.Get(k)
}

func (self diskStore) Put(k, v []byte) error {
	return self.disk.Put(k, v)
}

func (self diskStore) Delete(k []byte) error {
	return self.disk.Delete(k)
}

func (self diskStore) ForEach(f func(k, v []byte) error) error {
	return self.disk.ForEach(func(k, v []byte) error {
		return f(k, v)
	})
}

func (self diskStore) PrefixScan(prefix []byte, f func(k, v []byte) error) error {
	cursor := self.disk.Cursor()
	for k, v := cursor.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = cursor.Next() {
		if err := f(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (self diskStore) RangeScan(start, end []byte, inclusive bool, f func(k, v []byte) error) error {
	cursor := self.disk.Cursor()

	var k, v []byte
	if start == nil {
		k, v = cursor.First()
	} else {
		k, v = cursor.Seek(start)
		if !inclusive && bytes.Equal(k, start) {
			k, v = cursor.Next()
		}
	}

	for ; k != nil; k, v = cursor.Next() {
		if end != nil {
			if inclusive {
				if bytes.Compare(k, end) > 0 {
					break
				}
			} else {
				if bytes.Compare(k, end) >= 0 {
					break
				}
			}
		}
		if err := f(k, v); err != nil {
			return err
		}
	}
	return nil
}

type diskStoreCoordinator struct {
	tx         *bbolt.Tx
	isReadOnly bool
	mu         sync.Mutex
}

func (self *diskStoreCoordinator) Get(storageName string) (Storage, error) {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.isReadOnly {
		storage := self.tx.Bucket([]byte(storageName))
		if storage == nil {
			return nullReadOnlyStorage{}, nil
		}
		return diskStore{disk: storage}, nil
	}

	storage, err := self.tx.CreateBucketIfNotExists([]byte(storageName))
	if err != nil {
		return nil, fmt.Errorf("create storage failed(bucket) %s: %w", storageName, err)
	}
	return diskStore{disk: storage}, nil
}

func (self *diskStoreCoordinator) Delete(storageName string) error {
	self.mu.Lock()
	defer self.mu.Unlock()
	if self.isReadOnly {
		return fmt.Errorf("failed to delete storage (%s). read-only constraitns", storageName)
	}
	return self.tx.DeleteBucket([]byte(storageName))
}

type openDiskStore struct {
	db *bbolt.DB
}

func (self openDiskStore) Path() string {
	return self.db.Path()
}

func (self openDiskStore) Read(f func(StorageCoordinator) error) error {
	return self.db.View(func(tx *bbolt.Tx) error {
		coordinator := &diskStoreCoordinator{tx: tx, isReadOnly: true}
		return f(coordinator)
	})
}

func (self openDiskStore) Write(f func(StorageCoordinator) error) error {
	return self.db.Update(func(tx *bbolt.Tx) error {
		coordinator := &diskStoreCoordinator{tx: tx}
		return f(coordinator)
	})
}

func (self openDiskStore) BackupToFile(path string) error {
	return self.db.View(func(tx *bbolt.Tx) error {
		return tx.CopyFile(path, 0644)
	})
}

func (self openDiskStore) SizeInBytes() (int64, error) {
	var size int64
	err := self.db.View(func(tx *bbolt.Tx) error {
		size = tx.Size()
		return nil
	})
	return size, err
}

func (self openDiskStore) Close() error {
	return self.db.Close()
}

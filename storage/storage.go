package storage

const defaultPath = "./.vcache"

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

func Open(path string) (StorageLayer, error) {
	return newCompressionCDat(path)
}

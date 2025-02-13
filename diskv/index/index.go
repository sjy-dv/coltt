package index

import "github.com/sjy-dv/coltt/pkg/wal"

type Indexer interface {
	Put(key []byte, position *wal.ChunkPosition) *wal.ChunkPosition

	Get(key []byte) *wal.ChunkPosition

	Delete(key []byte) (*wal.ChunkPosition, bool)

	Size() int

	Ascend(handleFn func(key []byte, position *wal.ChunkPosition) (bool, error))

	AscendRange(startKey, endKey []byte, handleFn func(key []byte, position *wal.ChunkPosition) (bool, error))

	AscendGreaterOrEqual(key []byte, handleFn func(key []byte, position *wal.ChunkPosition) (bool, error))

	Descend(handleFn func(key []byte, pos *wal.ChunkPosition) (bool, error))

	DescendRange(startKey, endKey []byte, handleFn func(key []byte, position *wal.ChunkPosition) (bool, error))

	DescendLessOrEqual(key []byte, handleFn func(key []byte, position *wal.ChunkPosition) (bool, error))
}

type IndexerType = byte

const (
	BTree IndexerType = iota
)

// Change the index type as you implement.
var indexType = BTree

func NewIndexer() Indexer {
	switch indexType {
	case BTree:
		return newBTree()
	default:
		panic("unexpected index type")
	}
}

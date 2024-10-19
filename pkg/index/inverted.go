package index

import (
	"sync"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/sjy-dv/vemoo/storage"
)

type Invertable interface {
	uint64 | int64 | float64 | string
}

type setCacheItem struct {
	set     *roaring64.Bitmap
	isDirty bool
}

type IndexInverted[T Invertable] struct {
	setCache map[T]*setCacheItem
	storage  storage.Storage
	mu       sync.Mutex
}

func NewIndexInverted[T Invertable](storg storage.Storage) *IndexInverted[T] {
	inv := &IndexInverted[T]{
		setCache: make(map[T]*setCacheItem),
		storage:  storg,
	}
	return inv
}

func (inv *IndexInverted[T]) getSetCacheItem(value T, setBytes []byte) (*setCacheItem, error) {
    item, ok := inv.setCache[value]
    if !ok {
        key, err := 
    }
}
package index

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/sjy-dv/vemoo/pkg/models"
	"github.com/sjy-dv/vemoo/pkg/withcontext"
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
		key, err := toByteSortable(value)
		if err != nil {
			return nil, fmt.Errorf("error converting key to byte sortable: %w", err)
		}
		if setBytes == nil {
			setBytes = inv.storage.Get(key)
		}
		rSet := roaring64.New()
		if setBytes != nil {
			if _, err := rSet.ReadFrom(bytes.NewReader(setBytes)); err != nil {
				return nil, fmt.Errorf("error reading set from bytes: %w", err)
			}
		}
		item = &setCacheItem{
			set: rSet,
		}
		inv.setCache[value] = item
	}
	return item, nil
}

type IndexChange[T Invertable] struct {
	Id           uint64
	PreviousData *T
	CurrentData  *T
}

func (inv *IndexInverted[T]) InsertUpdateDelete(ctx context.Context, in <-chan IndexChange[T]) <-chan error {
	errC := make(chan error, 1)
	go func() {
		defer close(errC)
		inv.mu.Lock()
		defer inv.mu.Unlock()
		processErrC := withcontext.SinkWithContext(ctx, in, inv.processChange)
		if err := <-processErrC; err != nil {
			errC <- fmt.Errorf("error processing change: %w", err)
			return
		}
		errC <- inv.flush()
	}()
	return errC
}

func (inv *IndexInverted[T]) processChange(change IndexChange[T]) error {
	// ---------------------------
	switch {
	case change.PreviousData == nil && change.CurrentData == nil:
		// Blank change, nothing to do
	case change.PreviousData == nil && change.CurrentData != nil:
		// Insert
		set, err := inv.getSetCacheItem(*change.CurrentData, nil)
		if err != nil {
			return fmt.Errorf("error getting set cache item: %w", err)
		}
		set.isDirty = set.set.CheckedAdd(change.Id) || set.isDirty
	case change.PreviousData != nil && change.CurrentData == nil:
		// Delete
		set, err := inv.getSetCacheItem(*change.PreviousData, nil)
		if err != nil {
			return fmt.Errorf("error getting set cache item: %w", err)
		}
		set.isDirty = set.set.CheckedRemove(change.Id) || set.isDirty
	case *change.PreviousData != *change.CurrentData:
		// Update
		prevSet, err := inv.getSetCacheItem(*change.PreviousData, nil)
		if err != nil {
			return fmt.Errorf("error getting set cache item: %w", err)
		}
		prevSet.isDirty = prevSet.set.CheckedRemove(change.Id) || prevSet.isDirty
		currSet, err := inv.getSetCacheItem(*change.CurrentData, nil)
		if err != nil {
			return fmt.Errorf("error getting set cache item: %w", err)
		}
		currSet.isDirty = currSet.set.CheckedAdd(change.Id) || currSet.isDirty
	case *change.PreviousData == *change.CurrentData:
		// This case needs to be last not to get null pointer exception
	}
	return nil
}

func (inv *IndexInverted[T]) flush() error {
	// ---------------------------
	for term, item := range inv.setCache {
		if !item.isDirty {
			continue
		}
		// ---------------------------
		key, err := toByteSortable(term)
		if err != nil {
			return fmt.Errorf("error converting key to byte sortable: %w", err)
		}
		if item.set.IsEmpty() {
			if err := inv.storage.Delete(key); err != nil {
				return fmt.Errorf("error deleting term set from storage: %w", err)
			}
			continue
		}
		// ---------------------------
		setBytes, err := item.set.ToBytes()
		if err != nil {
			return fmt.Errorf("error converting term set to bytes: %w", err)
		}
		if err := inv.storage.Put(key, setBytes); err != nil {
			return fmt.Errorf("error putting term set to storage: %w", err)
		}
	}
	// ---------------------------
	return nil
}

func (inv *IndexInverted[T]) Search(query T, endQuery T, operator string) (*roaring64.Bitmap, error) {
	inv.mu.Lock()
	defer inv.mu.Unlock()
	// ---------------------------
	queryKey, err := toByteSortable(query)
	if err != nil {
		return nil, fmt.Errorf("error converting value %v to search: %w", query, err)
	}
	sets := make([]*roaring64.Bitmap, 0, 1)
	// ---------------------------
	var start, end []byte
	var inclusive bool
	// ---------------------------
	switch operator {
	// ---------------------------
	case models.OperatorEquals:
		item, err := inv.getSetCacheItem(query, nil)
		if err != nil {
			return nil, fmt.Errorf("error getting set cache item: %w", err)
		}
		return item.set, nil
	// ---------------------------
	case models.OperatorNotEquals:

		err := inv.storage.ForEach(func(k, v []byte) error {
			if bytes.Equal(k, queryKey) {
				return nil
			}
			var reverseKey T
			err := fromByteSortable(k, &reverseKey)
			if err != nil {
				return fmt.Errorf("error converting key to value: %w", err)
			}
			item, err := inv.getSetCacheItem(reverseKey, v)
			if err != nil {
				return fmt.Errorf("error getting set cache item: %w", err)
			}
			sets = append(sets, item.set)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error iterating over storage for inverted search: %w", err)
		}
	// ---------------------------
	case models.OperatorStartsWith:
		err := inv.storage.PrefixScan(queryKey, func(k, v []byte) error {
			var reverseKey T
			err := fromByteSortable(k, &reverseKey)
			if err != nil {
				return fmt.Errorf("error converting key to value: %w", err)
			}
			item, err := inv.getSetCacheItem(reverseKey, v)
			if err != nil {
				return fmt.Errorf("error getting set cache item: %w", err)
			}
			sets = append(sets, item.set)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error prefix scanning over storage for inverted search: %w", err)
		}
	// ---------------------------
	case models.OperatorGreaterThan:
		start = queryKey
		inclusive = false
	case models.OperatorGreaterOrEq:
		start = queryKey
		inclusive = true
	case models.OperatorLessThan:
		end = queryKey
		inclusive = false
	case models.OperatorLessOrEq:
		end = queryKey
		inclusive = true
	case models.OperatorInRange:
		start = queryKey
		endk, err := toByteSortable(endQuery)
		if err != nil {
			return nil, fmt.Errorf("error converting value %v to search: %w", endQuery, err)
		}
		end = endk
		inclusive = true
	// ---------------------------
	default:
		return nil, fmt.Errorf("unknown inverted search operator: %s", operator)
	}
	// ---------------------------
	if start != nil || end != nil {
		err := inv.storage.RangeScan(start, end, inclusive, func(k, v []byte) error {
			var reverseKey T
			err := fromByteSortable(k, &reverseKey)
			if err != nil {
				return fmt.Errorf("error converting key to value: %w", err)
			}
			item, err := inv.getSetCacheItem(reverseKey, v)
			if err != nil {
				return fmt.Errorf("error getting set cache item: %w", err)
			}
			sets = append(sets, item.set)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error range scanning over storage for inverted search: %w", err)
		}
	}
	// ---------------------------
	if len(sets) == 0 {
		return roaring64.New(), nil
	}
	if len(sets) == 1 {
		return sets[0], nil
	}
	// ---------------------------
	return roaring64.FastOr(sets...), nil
}

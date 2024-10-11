package wal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"os"
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	etcdRaft "go.etcd.io/raft/v3"
	etcdRaftpb "go.etcd.io/raft/v3/raftpb"
)

var (
	badgerRaftIdKey    []byte = []byte("raftid")
	cacheSnapshotKey   string = "snapshot"
	cacheFirstIndexKey string = "firstindex"
	cacheLastIndexKey  string = "lastindex"
)

var (
	ErrEntryNotFound  error = errors.New("Entry not found")
	ErrEmptyConfState error = errors.New("Empty ConfState")
)

type badgerWAL struct {
	groupId uuid.UUID
	db      *badger.DB
	cache   *sync.Map
}

func GetBadgerRaftId(db *badger.DB) (uint64, error) {
	var id uint64
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(badgerRaftIdKey)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			id = binary.BigEndian.Uint64(val)
			return nil
		})
	})
	return id, err
}

func SetBadgerRaftId(db *badger.DB, id uint64) error {
	return db.Update(func(txn *badger.Txn) error {
		var b [8]byte
		binary.BigEndian.PutUint64(b[:], id)
		return txn.Set(badgerRaftIdKey, b[:])
	})
}

func NewBadgerWal(db *badger.DB, groupId uuid.UUID) *badgerWAL {
	wal := &badgerWAL{
		groupId: groupId,
		db:      db,
		cache:   new(sync.Map),
	}

	_, err := wal.FirstIndex()
	if err == ErrEntryNotFound {
		wal.reset(make([]etcdRaftpb.Entry, 1))
	} else if err != nil {
		logrus.WithFields(logrus.Fields{
			"action": "New Badger Wal initialize",
		}).Error(err.Error())
		os.Exit(1)
	}
	return wal
}

func (bw *badgerWAL) InitialState() (
	etcdRaftpb.HardState, etcdRaftpb.ConfState, error) {
	hardstate, err := bw.HardState()
	if err != nil {
		return etcdRaftpb.HardState{}, etcdRaftpb.ConfState{}, err
	}

	snapshot, err := bw.Snapshot()
	if err != nil {
		return hardstate, etcdRaftpb.ConfState{}, err
	}
	return hardstate, snapshot.Metadata.ConfState, nil
}

func (bw *badgerWAL) HardState() (etcdRaftpb.HardState, error) {
	var hardState etcdRaftpb.HardState
	err := bw.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(bw.hardStateKey())
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return hardState.Unmarshal(val)
		})
	})
	if err == badger.ErrKeyNotFound {
		return hardState, nil
	}
	return hardState, err
}

func (bw *badgerWAL) hardStateKey() []byte {
	b := make([]byte, 18)
	copy(b[0:2], []byte("hs"))
	copy(b[2:18], bw.groupId[:])
	return b
}

func (bw *badgerWAL) Entries(lo, hi, maxSize uint64) (
	[]etcdRaftpb.Entry, error) {
	firstIndex, err := bw.FirstIndex()
	if err != nil {
		return nil, err
	}
	if lo < firstIndex {
		return nil, etcdRaft.ErrCompacted
	}

	lastIndex, err := bw.LastIndex()
	if err != nil {
		return nil, err
	}

	if hi > lastIndex+1 {
		return nil, etcdRaft.ErrUnavailable
	}
	return bw.getEntries(lo, hi, maxSize)
}

func (bw *badgerWAL) Term(idx uint64) (uint64, error) {
	firstIndex, err := bw.FirstIndex()
	if err != nil {
		return 0, err
	}
	if idx < firstIndex-1 {
		return 0, etcdRaft.ErrCompacted
	}

	var entry etcdRaftpb.Entry
	_, err = bw.seekEntry(&entry, idx, false)
	if err == ErrEntryNotFound {
		return 0, etcdRaft.ErrUnavailable
	} else if err != nil {
		return 0, err
	}

	if idx < entry.Index {
		return 0, etcdRaft.ErrCompacted
	}
	return entry.Term, nil
}

func (bw *badgerWAL) FirstIndex() (uint64, error) {
	if v, exists := bw.cache.Load(cacheSnapshotKey); exists {
		if snapshot, ok := v.(*etcdRaftpb.Snapshot); ok && !etcdRaft.IsEmptySnap(*snapshot) {
			return snapshot.Metadata.Index + 1, nil
		}
	}

	if v, exists := bw.cache.Load(cacheFirstIndexKey); exists {
		if idx, ok := v.(uint64); ok {
			return idx, nil
		}
	}

	index, err := bw.seekEntry(nil, 0, false)
	if err != nil {
		return 0, nil
	}
	bw.cache.Store(cacheFirstIndexKey, index+1)
	return index + 1, nil
}

func (bw *badgerWAL) LastIndex() (uint64, error) {
	if v, exists := bw.cache.Load(cacheLastIndexKey); exists {
		if idx, ok := v.(uint64); ok {
			return idx, nil
		}
	}
	return bw.seekEntry(nil, math.MaxUint64, true)
}

func (bw *badgerWAL) Save(hardState etcdRaftpb.HardState, entries []etcdRaftpb.Entry, snapshot etcdRaftpb.Snapshot) error {
	batch := bw.db.NewWriteBatch()
	defer batch.Cancel()

	if err := bw.writeEntries(batch, entries); err != nil {
		return err
	}
	if err := bw.writeHardState(batch, hardState); err != nil {
		return err
	}
	if !etcdRaft.IsEmptySnap(snapshot) {
		if err := bw.writeSnapshot(batch, snapshot); err != nil {
			return err
		}
		// Delete the log
		if err := bw.deleteEntriesFromIndex(batch, 0); err != nil {
			return err
		}
	}

	return batch.Flush()
}

func (bw *badgerWAL) CreateSnapshot(idx uint64, confState *etcdRaftpb.ConfState, data []byte) (etcdRaftpb.Snapshot, error) {
	var snapshot etcdRaftpb.Snapshot
	if confState == nil {
		return snapshot, ErrEmptyConfState
	}
	firstIndex, err := bw.FirstIndex()
	if err != nil {
		return snapshot, err
	}
	if idx < firstIndex {
		return snapshot, etcdRaft.ErrSnapOutOfDate
	}

	var entry etcdRaftpb.Entry
	if _, err := bw.seekEntry(&entry, idx, false); err != nil {
		return snapshot, err
	}
	if idx != entry.Index {
		return snapshot, ErrEntryNotFound
	}

	snapshot.Metadata.Index = entry.Index
	snapshot.Metadata.Term = entry.Term
	if confState != nil {
		snapshot.Metadata.ConfState = *confState
	}
	snapshot.Data = data

	batch := bw.db.NewWriteBatch()
	defer batch.Cancel()

	if err := bw.writeSnapshot(batch, snapshot); err != nil {
		return snapshot, err
	}
	if err := bw.deleteEntriesUntilIndex(batch, snapshot.Metadata.Index); err != nil {
		return snapshot, err
	}

	return snapshot, batch.Flush()
}

func (bw *badgerWAL) DeleteGroup() error {
	return bw.reset(nil)
}

func (bw *badgerWAL) seekEntry(
	entry *etcdRaftpb.Entry,
	seekTo uint64,
	reverse bool,
) (uint64, error) {
	var index uint64
	err := bw.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = bw.entryPrefix()
		opts.Reverse = reverse
		iter := txn.NewIterator(opts)
		defer iter.Close()

		iter.Seek(bw.entryKey(seekTo))
		if !iter.Valid() {
			return ErrEntryNotFound
		}

		item := iter.Item()
		index = bw.parseIndex(item.Key())
		if entry == nil {
			return nil
		}
		return item.Value(func(val []byte) error {
			return entry.Unmarshal(val)
		})
	})
	return index, err
}

func (bw *badgerWAL) entryPrefix() []byte {
	return bw.groupId[:]
}

func (bw *badgerWAL) entryKey(idx uint64) []byte {
	b := make([]byte, 24)
	copy(b[0:16], bw.entryPrefix())
	binary.BigEndian.PutUint64(b[16:24], idx)
	return b
}

func (bw *badgerWAL) parseIndex(key []byte) uint64 {
	return binary.BigEndian.Uint64(key[16:24])
}

func (bw *badgerWAL) Snapshot() (etcdRaftpb.Snapshot, error) {
	if v, exists := bw.cache.Load(cacheSnapshotKey); exists {
		if snapshot, ok := v.(*etcdRaftpb.Snapshot); ok &&
			!etcdRaft.IsEmptySnap(*snapshot) {
			return *snapshot, nil
		}
	}

	var snapshot etcdRaftpb.Snapshot
	err := bw.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(bw.snapShotKey())
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return snapshot.Unmarshal(val)
		})
	})
	if err == badger.ErrKeyNotFound {
		return snapshot, nil
	}
	return snapshot, err
}

func (bw *badgerWAL) snapShotKey() []byte {
	b := make([]byte, 18)
	copy(b[0:2], []byte("ss"))
	copy(b[2:18], bw.groupId[:])
	return b
}

func (bw *badgerWAL) reset(entries []etcdRaftpb.Entry) error {
	bw.cache = new(sync.Map)

	batch := bw.db.NewWriteBatch()
	defer batch.Cancel()

	if err := bw.deleteEntriesFromIndex(batch, 0); err != nil {
		return err
	}

	for _, entry := range entries {
		val, err := entry.Marshal()
		if err != nil {
			return err
		}
		err = batch.Set(bw.entryKey(entry.Index), val)
		if err != nil {
			return err
		}
	}
	return batch.Flush()
}

func (bw *badgerWAL) deleteEntriesFromIndex(batch *badger.WriteBatch, fromIdx uint64) error {
	var keys []string
	err := bw.db.View(func(txn *badger.Txn) error {
		startKey := bw.entryKey(fromIdx)

		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = bw.entryPrefix()

		iter := txn.NewIterator(opts)
		defer iter.Close()

		for iter.Seek(startKey); iter.Valid(); iter.Next() {
			keys = append(keys, string(iter.Item().Key()))
		}
		return nil
	})
	if err != nil {
		return err
	}
	return bw.deleteKeys(batch, keys)
}

func (bw *badgerWAL) deleteEntriesUntilIndex(batch *badger.WriteBatch, untilIdx uint64) error {
	var keys []string
	var index uint64
	err := bw.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Prefix = bw.entryPrefix()

		iter := txn.NewIterator(opts)
		defer iter.Close()

		startKey := bw.entryKey(0)
		first := true
		for iter.Seek(startKey); iter.Valid(); iter.Next() {
			item := iter.Item()
			index = bw.parseIndex(item.Key())
			if first {
				first = false
				if untilIdx <= index {
					return etcdRaft.ErrCompacted
				}
			}
			if index >= untilIdx {
				break
			}
			keys = append(keys, string(item.Key()))
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err := bw.deleteKeys(batch, keys); err != nil {
		return err
	}

	if v, ok := bw.cache.Load(cacheFirstIndexKey); ok {
		if v.(uint64) <= untilIdx {
			bw.cache.Store(cacheFirstIndexKey, untilIdx+1)
		}
	}
	return nil
}

func (bw *badgerWAL) deleteKeys(batch *badger.WriteBatch, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	for _, key := range keys {
		if err := batch.Delete([]byte(key)); err != nil {
			return err
		}
	}
	return nil
}

func (bw *badgerWAL) getEntries(lo, hi, maxSize uint64) ([]etcdRaftpb.Entry, error) {
	var entries []etcdRaftpb.Entry
	err := bw.db.View(func(txn *badger.Txn) error {
		if hi-lo == 1 {
			item, err := txn.Get(bw.entryKey(lo))
			if err != nil {
				return err
			}
			return item.Value(func(val []byte) error {
				var entry etcdRaftpb.Entry
				if err := entry.Unmarshal(val); err != nil {
					return err
				}
				entries = append(entries, entry)
				return nil
			})
		}

		iterOpt := badger.DefaultIteratorOptions
		iterOpt.PrefetchValues = false
		iterOpt.Prefix = bw.entryPrefix()

		iterator := txn.NewIterator(iterOpt)
		defer iterator.Close()

		startKey := bw.entryKey(lo)
		endKey := bw.entryKey(hi)

		var size uint64 = 0
		first := true
		for iterator.Seek(startKey); iterator.Valid(); iterator.Next() {
			item := iterator.Item()
			var entry etcdRaftpb.Entry
			err := item.Value(func(val []byte) error {
				return entry.Unmarshal(val)
			})
			if err != nil {
				return err
			}

			if bytes.Compare(item.Key(), endKey) >= 0 {
				break
			}
			size += uint64(entry.Size())
			if size > maxSize && !first {
				break
			}
			first = false
			entries = append(entries, entry)
		}
		return nil

	})

	return entries, err
}

func (bw *badgerWAL) writeEntries(batch *badger.WriteBatch, entries []etcdRaftpb.Entry) error {
	if len(entries) == 0 {
		return nil
	}

	firstIndex, err := bw.FirstIndex()
	if err != nil {
		return err
	}
	firstEntryIndex := entries[0].Index
	if firstEntryIndex+uint64(len(entries))-1 < firstIndex {
		return nil
	}
	if firstIndex > firstEntryIndex {
		entries = entries[(firstIndex - firstEntryIndex):]
	}

	lastIndex, err := bw.LastIndex()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryData, err := entry.Marshal()
		if err != nil {
			return err
		}
		err = batch.Set(bw.entryKey(entry.Index), entryData)
		if err != nil {
			return err
		}
	}

	lastEntryIndex := entries[len(entries)-1].Index
	bw.cache.Store(cacheLastIndexKey, lastEntryIndex)
	if lastIndex > lastEntryIndex {
		return bw.deleteEntriesFromIndex(batch, lastEntryIndex+1)
	}
	return nil
}

func (bw *badgerWAL) writeHardState(batch *badger.WriteBatch, hardState etcdRaftpb.HardState) error {
	if etcdRaft.IsEmptyHardState(hardState) {
		return nil
	}

	hardStateData, err := hardState.Marshal()
	if err != nil {
		return err
	}

	return batch.Set(bw.hardStateKey(), hardStateData)
}

func (bw *badgerWAL) writeSnapshot(batch *badger.WriteBatch, snapshot etcdRaftpb.Snapshot) error {
	if etcdRaft.IsEmptySnap(snapshot) {
		return nil
	}

	snapshotData, err := snapshot.Marshal()
	if err != nil {
		return err
	}
	err = batch.Set(bw.snapShotKey(), snapshotData)
	if err != nil {
		return err
	}

	entry := etcdRaftpb.Entry{Term: snapshot.Metadata.Term, Index: snapshot.Metadata.Index}
	entryData, err := entry.Marshal()
	if err != nil {
		return err
	}
	err = batch.Set(bw.entryKey(entry.Index), entryData)
	if err != nil {
		return err
	}

	if v, exists := bw.cache.Load(cacheLastIndexKey); exists {
		if v.(uint64) < entry.Index {
			bw.cache.Store(cacheLastIndexKey, entry.Index)
		}
	}

	bw.cache.Store(cacheSnapshotKey, &snapshot)

	return nil
}

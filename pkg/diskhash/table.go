// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package diskhash

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sjy-dv/nnv/pkg/fs"
	"github.com/sjy-dv/nnv/pkg/murmurV3"
)

const (
	primaryFileName  = "HASH.PRIMARY"
	overflowFileName = "HASH.OVERFLOW"
	metaFileName     = "HASH.META"
	slotsPerBucket   = 31
	nextOffLen       = 8
	hashLen          = 4
)

type MatchKeyFunc func(Slot) (bool, error)

type Table struct {
	primaryFile  fs.File
	overflowFile fs.File
	metaFile     fs.File // meta file stores the metadata of the hash table
	meta         *tableMeta
	mu           *sync.RWMutex // protect the table when multiple goroutines access it
	options      Options
}

type tableMeta struct {
	Level            uint8
	SplitBucketIndex uint32
	NumBuckets       uint32
	NumKeys          uint32
	SlotValueLength  uint32
	BucketSize       uint32
	FreeBuckets      []int64
}

func Open(options Options) (*Table, error) {
	if err := checkOptions(options); err != nil {
		return nil, err
	}

	t := &Table{
		mu:      new(sync.RWMutex),
		options: options,
	}

	// create data directory if not exist
	if _, err := os.Stat(options.DirPath); err != nil {
		if err := os.MkdirAll(options.DirPath, os.ModePerm); err != nil {
			return nil, err
		}
	}

	// open meta file and read metadata info
	if err := t.readMeta(); err != nil {
		return nil, err
	}

	// open and init primary file
	primaryFile, err := t.openFile(primaryFileName)
	if err != nil {
		return nil, err
	}
	// init first bucket if the primary file is empty
	if primaryFile.Size() == int64(t.meta.BucketSize) {
		if err := primaryFile.Truncate(int64(t.meta.BucketSize)); err != nil {
			_ = primaryFile.Close()
			return nil, err
		}
	}
	t.primaryFile = primaryFile

	// open overflow file
	overflowFile, err := t.openFile(overflowFileName)
	if err != nil {
		return nil, err
	}
	t.overflowFile = overflowFile

	return t, nil
}

func checkOptions(options Options) error {
	if options.DirPath == "" {
		return errors.New("dir path cannot be empty")
	}
	if options.SlotValueLength <= 0 {
		return errors.New("slot value length must be greater than 0")
	}
	if options.LoadFactor < 0 || options.LoadFactor > 1 {
		return errors.New("load factor must be between 0 and 1")
	}
	return nil
}

func (t *Table) readMeta() error {
	file, err := fs.Open(filepath.Join(t.options.DirPath, metaFileName), fs.OSFileSystem)
	if err != nil {
		return err
	}
	t.metaFile = file
	t.meta = &tableMeta{}

	// init meta file if not exist
	if file.Size() == 0 {
		t.meta.NumBuckets = 1
		t.meta.SlotValueLength = t.options.SlotValueLength
		t.meta.BucketSize = slotsPerBucket*(hashLen+t.meta.SlotValueLength) + nextOffLen
	} else {
		decoder := json.NewDecoder(t.metaFile)
		if err := decoder.Decode(t.meta); err != nil {
			return err
		}
		if t.meta.SlotValueLength != t.options.SlotValueLength {
			return errors.New("slot value length mismatch")
		}
	}

	return nil
}

// write the metadata info to the meta file in json format.
func (t *Table) writeMeta() error {
	encoder := json.NewEncoder(t.metaFile)
	return encoder.Encode(t.meta)
}

// Close closes the files of the hash table.
func (t *Table) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.writeMeta(); err != nil {
		return err
	}

	_ = t.primaryFile.Close()
	_ = t.overflowFile.Close()
	_ = t.metaFile.Close()
	return nil
}

// Sync flushes the data of the hash table to disk.
func (t *Table) Sync() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.primaryFile.Sync(); err != nil {
		return err
	}

	if err := t.overflowFile.Sync(); err != nil {
		return err
	}

	return nil
}

func (t *Table) Put(key, value []byte, matchKey MatchKeyFunc) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(value) != int(t.meta.SlotValueLength) {
		return errors.New("value length must be equal to the length specified in the options")
	}

	keyHash := getKeyHash(key)
	slot := &Slot{Hash: keyHash, Value: value}
	sw, err := t.getSlotWriter(slot.Hash, matchKey)
	if err != nil {
		return err
	}

	// write the new slot to the bucket
	if err = sw.insertSlot(*slot, t); err != nil {
		return err
	}
	if err := sw.writeSlots(); err != nil {
		return err
	}
	// if the slot already exists, no need to update meta
	// because the number of keys has not changed
	if sw.overwrite {
		return nil
	}

	t.meta.NumKeys++
	// split if the load factor is exceeded
	keyRatio := float64(t.meta.NumKeys) / float64(t.meta.NumBuckets*slotsPerBucket)
	if keyRatio > t.options.LoadFactor {
		if err := t.split(); err != nil {
			return err
		}
	}
	return nil
}

func (t *Table) getSlotWriter(keyHash uint32, matchKey MatchKeyFunc) (*slotWriter, error) {
	sw := &slotWriter{}
	bi := t.newBucketIterator(t.getKeyBucket(keyHash))

	for {
		b, err := bi.next()
		if err == io.EOF {
			return nil, errors.New("failed to put new slot")
		}
		if err != nil {
			return nil, err
		}

		sw.currentBucket = b

		for i, slot := range b.slots {
			if slot.Hash == 0 {
				sw.currentSlotIndex = i
				return sw, nil
			}
			if slot.Hash != keyHash {
				continue
			}
			match, err := matchKey(slot)
			if err != nil {
				return nil, err
			}
			// key already exists, overwrite the value
			if match {
				sw.currentSlotIndex, sw.overwrite = i, true
				return sw, nil
			}
		}

		if b.nextOffset == 0 {
			sw.currentSlotIndex = slotsPerBucket
			return sw, nil
		}
	}
}

func (t *Table) Get(key []byte, matchKey MatchKeyFunc) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	keyHash := getKeyHash(key)
	startBucket := t.getKeyBucket(keyHash)
	bi := t.newBucketIterator(startBucket)
	for {
		b, err := bi.next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		for _, slot := range b.slots {
			if slot.Hash == 0 {
				break
			}
			if slot.Hash != keyHash {
				continue
			}
			if match, err := matchKey(slot); match || err != nil {
				return err
			}
		}
	}
}

func (t *Table) Delete(key []byte, matchKey MatchKeyFunc) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// get the bucket according to the key hash
	keyHash := getKeyHash(key)
	bi := t.newBucketIterator(t.getKeyBucket(keyHash))
	for {
		b, err := bi.next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// the following code is similar to the Get method
		for i, slot := range b.slots {
			if slot.Hash == 0 {
				break
			}
			if slot.Hash != keyHash {
				continue
			}
			match, err := matchKey(slot)
			if err != nil {
				return err
			}
			if !match {
				continue
			}
			// now we find the slot to delete, remove it from the bucket
			b.removeSlot(i)
			if err := b.write(); err != nil {
				return err
			}
			t.meta.NumKeys--
			return nil
		}
	}
}

// Size returns the number of keys in the hash table.
func (t *Table) Size() uint32 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.meta.NumKeys
}

// get the hash value according to the key
func getKeyHash(key []byte) uint32 {
	return murmurV3.Sum32(key)
}

// get the bucket according to the key
func (t *Table) getKeyBucket(keyHash uint32) uint32 {
	bucketIndex := keyHash & ((1 << t.meta.Level) - 1)
	if bucketIndex < t.meta.SplitBucketIndex {
		return keyHash & ((1 << (t.meta.Level + 1)) - 1)
	}
	return bucketIndex
}

func (t *Table) openFile(name string) (fs.File, error) {
	file, err := fs.Open(filepath.Join(t.options.DirPath, name), fs.OSFileSystem)
	if err != nil {
		return nil, err
	}
	// init (dummy) file header
	// the first bucket size in the file is not used, so we just init it.
	if file.Size() == 0 {
		if err := file.Truncate(int64(t.meta.BucketSize)); err != nil {
			_ = file.Close()
			return nil, err
		}
	}
	return file, nil
}

// split the current bucket.
// it will create a new bucket, and rewrite all slots in the current bucket and the overflow buckets.
func (t *Table) split() error {
	// the bucket to be split
	splitBucketIndex := t.meta.SplitBucketIndex
	splitSlotWriter := &slotWriter{
		currentBucket: &bucket{
			file:       t.primaryFile,
			offset:     t.bucketOffset(splitBucketIndex),
			bucketSize: t.meta.BucketSize,
		},
	}

	// create a new bucket
	newSlotWriter := &slotWriter{
		currentBucket: &bucket{
			file:       t.primaryFile,
			offset:     t.primaryFile.Size(),
			bucketSize: t.meta.BucketSize,
		},
	}
	if err := t.primaryFile.Truncate(int64(t.meta.BucketSize)); err != nil {
		return err
	}

	// increase the split bucket index
	t.meta.SplitBucketIndex++
	// if the split bucket index is equal to 1 << level,
	// reset the split bucket index to 0, and increase the level.
	if t.meta.SplitBucketIndex == 1<<t.meta.Level {
		t.meta.Level++
		t.meta.SplitBucketIndex = 0
	}

	// iterate all slots in the split bucket and the overflow buckets,
	// rewrite all slots.
	var freeBuckets []int64
	bi := t.newBucketIterator(splitBucketIndex)
	for {
		b, err := bi.next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		for _, slot := range b.slots {
			if slot.Hash == 0 {
				break
			}
			var insertErr error
			if t.getKeyBucket(slot.Hash) == splitBucketIndex {
				insertErr = splitSlotWriter.insertSlot(slot, t)
			} else {
				insertErr = newSlotWriter.insertSlot(slot, t)
			}
			if insertErr != nil {
				return insertErr
			}
		}
		// if the splitBucket has overflow buckets, and these buckets are no longer used,
		// because all slots has been rewritten, so we can free these buckets.
		if b.nextOffset != 0 {
			freeBuckets = append(freeBuckets, b.nextOffset)
		}
	}

	// collect the free buckets
	if len(freeBuckets) > 0 {
		t.meta.FreeBuckets = append(t.meta.FreeBuckets, freeBuckets...)
	}

	// write all slots to the file
	if err := splitSlotWriter.writeSlots(); err != nil {
		return err
	}
	if err := newSlotWriter.writeSlots(); err != nil {
		return err
	}

	t.meta.NumBuckets++
	return nil
}

// create overflow bucket.
// If there are free buckets, it will reuse the free buckets.
// Otherwise, it will create a new bucket in the overflow file.
func (t *Table) createOverflowBucket() (*bucket, error) {
	var offset int64
	if len(t.meta.FreeBuckets) > 0 {
		offset = t.meta.FreeBuckets[0]
		t.meta.FreeBuckets = t.meta.FreeBuckets[1:]
	} else {
		offset = t.overflowFile.Size()
		err := t.overflowFile.Truncate(int64(t.meta.BucketSize))
		if err != nil {
			return nil, err
		}
	}

	return &bucket{
		file:       t.overflowFile,
		offset:     offset,
		bucketSize: t.meta.BucketSize,
	}, nil
}

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
	"encoding/binary"
	"io"

	"github.com/sjy-dv/coltt/pkg/fs"
)

type bucket struct {
	slots      [slotsPerBucket]Slot
	offset     int64
	nextOffset int64
	file       fs.File
	bucketSize uint32
}

type bucketIterator struct {
	currentFile  fs.File
	overflowFile fs.File
	offset       int64

	slotValueLen uint32
	bucketSize   uint32
}

type Slot struct {
	Hash  uint32
	Value []byte
}

type slotWriter struct {
	currentBucket    *bucket
	currentSlotIndex int
	prevBuckets      []*bucket
	overwrite        bool
}

func (t *Table) bucketOffset(bucketIndex uint32) int64 {
	return int64((bucketIndex + 1) * t.meta.BucketSize)
}

func (t *Table) newBucketIterator(startBucket uint32) *bucketIterator {
	return &bucketIterator{
		currentFile:  t.primaryFile,
		overflowFile: t.overflowFile,
		offset:       t.bucketOffset(startBucket),
		slotValueLen: t.options.SlotValueLength,
		bucketSize:   t.meta.BucketSize,
	}
}

func (bi *bucketIterator) next() (*bucket, error) {

	if bi.offset == 0 {
		return nil, io.EOF
	}

	bucket, err := bi.readBucket()
	if err != nil {
		return nil, err
	}

	bi.offset = bucket.nextOffset
	bi.currentFile = bi.overflowFile
	return bucket, nil
}

func (bi *bucketIterator) readBucket() (*bucket, error) {

	bucketBuf := make([]byte, bi.bucketSize)
	if _, err := bi.currentFile.ReadAt(bucketBuf, bi.offset); err != nil {
		return nil, err
	}

	b := &bucket{file: bi.currentFile, offset: bi.offset, bucketSize: bi.bucketSize}

	for i := 0; i < slotsPerBucket; i++ {
		_ = bucketBuf[hashLen+bi.slotValueLen]
		b.slots[i].Hash = binary.LittleEndian.Uint32(bucketBuf[:hashLen])
		if b.slots[i].Hash != 0 {
			b.slots[i].Value = bucketBuf[hashLen : hashLen+bi.slotValueLen]
		}
		bucketBuf = bucketBuf[hashLen+bi.slotValueLen:]
	}

	b.nextOffset = int64(binary.LittleEndian.Uint64(bucketBuf[:nextOffLen]))

	return b, nil
}

func (sw *slotWriter) insertSlot(sl Slot, t *Table) error {
	if sw.currentSlotIndex == slotsPerBucket {
		nextBucket, err := t.createOverflowBucket()
		if err != nil {
			return err
		}
		sw.currentBucket.nextOffset = nextBucket.offset
		sw.prevBuckets = append(sw.prevBuckets, sw.currentBucket)
		sw.currentBucket = nextBucket
		sw.currentSlotIndex = 0
	}

	sw.currentBucket.slots[sw.currentSlotIndex] = sl
	sw.currentSlotIndex++
	return nil
}

func (sw *slotWriter) writeSlots() error {
	for i := len(sw.prevBuckets) - 1; i >= 0; i-- {
		if err := sw.prevBuckets[i].write(); err != nil {
			return err
		}
	}
	return sw.currentBucket.write()
}

func (b *bucket) write() error {
	buf := make([]byte, b.bucketSize)
	var index = 0
	for i := 0; i < slotsPerBucket; i++ {
		slot := b.slots[i]

		binary.LittleEndian.PutUint32(buf[index:index+hashLen], slot.Hash)
		copy(buf[index+hashLen:index+hashLen+len(slot.Value)], slot.Value)

		index += hashLen + len(slot.Value)
	}

	binary.LittleEndian.PutUint64(buf[len(buf)-nextOffLen:], uint64(b.nextOffset))

	_, err := b.file.WriteAt(buf, b.offset)
	return err
}

func (b *bucket) removeSlot(slotIndex int) {
	i := slotIndex
	for ; i < slotsPerBucket-1; i++ {
		b.slots[i] = b.slots[i+1]
	}
	b.slots[i] = Slot{}
}

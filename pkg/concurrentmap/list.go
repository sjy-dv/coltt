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

package concurrentmap

import "sync/atomic"

const (
	notDeleted uint32 = iota
	deleted
)

// newListHead returns the new head of any list
func newListHead[K hashable, V any]() *element[K, V] {
	e := &element[K, V]{keyHash: 0, key: *new(K)}
	e.nextPtr.Store(nil)
	e.value.Store(new(V))
	return e
}

// a single node in the list
type element[K hashable, V any] struct {
	keyHash uintptr
	key     K
	// The next element in the list. If this pointer has the marked flag set it means THIS element, not the next one, is deleted.
	nextPtr atomicPointer[element[K, V]]
	value   atomicPointer[V]
	deleted uint32
}

// next returns the next element
// this also deletes all marked elements while traversing the list
func (self *element[K, V]) next() *element[K, V] {
	for nextElement := self.nextPtr.Load(); nextElement != nil; {
		// if our next element is itself deleted (by the same criteria) then we will just replace
		// it with its next() (which should be the first node behind it that isn't itself deleted) and then check again
		if nextElement.isDeleted() {
			self.nextPtr.CompareAndSwap(nextElement, nextElement.next()) // actual deletion happens here after nodes are marked deleted lazily
			nextElement = self.nextPtr.Load()
		} else {
			return nextElement
		}
	}
	return nil
}

// addBefore inserts an element before the specified element
func (self *element[K, V]) addBefore(allocatedElement, before *element[K, V]) bool {
	if self.next() != before {
		return false
	}
	allocatedElement.nextPtr.Store(before)
	return self.nextPtr.CompareAndSwap(before, allocatedElement)
}

// inject updates an existing value in the list if present or adds a new entry
func (self *element[K, V]) inject(c uintptr, key K, value *V) (*element[K, V], bool) {
	var (
		alloc             *element[K, V]
		left, curr, right = self.search(c, key)
	)
	if curr != nil {
		curr.value.Store(value)
		return curr, false
	}
	if left != nil {
		alloc = &element[K, V]{keyHash: c, key: key}
		alloc.value.Store(value)
		if left.addBefore(alloc, right) {
			return alloc, true
		}
	}
	return nil, false
}

// search for an element in the list and return left_element, searched_element and right_element respectively
func (self *element[K, V]) search(c uintptr, key K) (*element[K, V], *element[K, V], *element[K, V]) {
	var (
		left, right *element[K, V]
		curr        = self
	)
	for {
		if curr == nil {
			return left, curr, right
		}
		right = curr.next()
		if c < curr.keyHash {
			right = curr
			curr = nil
			return left, curr, right
		} else if c == curr.keyHash && key == curr.key {
			return left, curr, right
		}
		left = curr
		curr = left.next()
		right = nil
	}
}

// remove marks a node for deletion
// the node will be removed in the next iteration via `element.next()`
// CAS ensures each node can be marked for deletion exactly once
func (self *element[K, V]) remove() bool {
	return atomic.CompareAndSwapUint32(&self.deleted, notDeleted, deleted)
}

// if current element is deleted
func (self *element[K, V]) isDeleted() bool {
	return atomic.LoadUint32(&self.deleted) == deleted
}

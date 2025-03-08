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

import "iter"

func (m *Map[K, V]) Iterator() iter.Seq2[K, V] {
	return func(yield func(key K, value V) bool) {
		for item := m.listHead.next(); item != nil; item = item.next() {
			if !yield(item.key, *item.value.Load()) {
				return
			}
		}
	}
}

func (m *Map[K, _]) Keys() iter.Seq[K] {
	return func(yield func(key K) bool) {
		for item := m.listHead.next(); item != nil; item = item.next() {
			if !yield(item.key) {
				return
			}
		}
	}
}

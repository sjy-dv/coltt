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

package core

import "sync"

type autoMap[T any] struct {
	mem  map[string]T
	lock sync.RWMutex
}

func NewAutoMap[T any]() *autoMap[T] {
	return &autoMap[T]{
		mem: make(map[string]T),
	}
}

func (xx *autoMap[T]) Set(k string, v T) {
	xx.lock.Lock()
	defer xx.lock.Unlock()
	xx.mem[k] = v
}

func (xx *autoMap[T]) Del(k string) {
	xx.lock.Lock()
	defer xx.lock.Unlock()
	delete(xx.mem, k)
}

func (xx *autoMap[T]) Get(k string) T {
	xx.lock.RLock()
	defer xx.lock.RUnlock()
	return xx.mem[k]
}

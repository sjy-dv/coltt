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

import (
	"errors"
	"fmt"
	"sync"

	"github.com/sjy-dv/coltt/pkg/index"
)

type IndexGroup struct {
	indexes   map[string]*index.BitmapIndex
	indexLock sync.RWMutex
}

var indexdb *IndexGroup

func NewIndexDB() {
	indexdb = &IndexGroup{
		indexes: make(map[string]*index.BitmapIndex),
	}
}

func (xx *IndexGroup) existsIndex(collectionName string) bool {
	xx.indexLock.RLock()
	_, exists := xx.indexes[collectionName]
	xx.indexLock.RUnlock()
	return exists
}

func (xx *IndexGroup) CreateIndex(collectionName string) error {
	c := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- fmt.Errorf(panicr, r)
			}
		}()
		ok := xx.existsIndex(collectionName)
		if ok {
			c <- errors.New("already exists Index")
			return
		}
		xx.indexLock.Lock()
		xx.indexes[collectionName] = index.NewBitmapIndex()
		xx.indexLock.Unlock()
		c <- nil
	}()
	return <-c
}

func (xx *IndexGroup) DropIndex(collectionName string) error {
	c := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- fmt.Errorf(panicr, r)
			}
		}()
		ok := xx.existsIndex(collectionName)
		if !ok {
			c <- nil
			return
		}
		xx.indexLock.Lock()
		delete(xx.indexes, collectionName)
		xx.indexLock.Unlock()
		c <- nil
	}()
	return <-c
}

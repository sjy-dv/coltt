package core

import (
	"errors"
	"fmt"
	"sync"

	"github.com/sjy-dv/nnv/pkg/index"
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

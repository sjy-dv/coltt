package highspeedmemory

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	panicr = "panic %v"
)

type HighSpeedMem struct {
	Collections  map[string]*CollectionMem
	groupLock    sync.RWMutex
	commiter     *time.Ticker
	stopCommiter chan bool
}

type CollectionMem struct {
	Data            map[uint64]interface{}
	CollectionName  string
	Distance        string
	Quantization    string
	Dim             uint32
	Connectivity    uint32
	ExpansionAdd    uint32
	ExpansionSearch uint32
	Multi           bool
	Storage         string
	collectionLock  sync.RWMutex
}

type CollectionConfig struct {
	CollectionName  string
	Distance        string
	Quantization    string
	Dim             uint32
	Connectivity    uint32
	ExpansionAdd    uint32
	ExpansionSearch uint32
	Multi           bool
	Storage         string
}

func NewHighSpeedMemory() *HighSpeedMem {
	NewTensorLink()
	return &HighSpeedMem{
		Collections:  map[string]*CollectionMem{},
		stopCommiter: make(chan bool),
	}
}

// return nil not exists collection
func (xx *HighSpeedMem) getCollection(collectionName string) *CollectionMem {
	xx.groupLock.RLock()
	col, exists := xx.Collections[collectionName]
	xx.groupLock.RUnlock()
	if exists {
		return col
	}
	return nil
}

func (xx *HighSpeedMem) existsCollection(collectionName string) bool {
	xx.groupLock.RLock()
	_, exists := xx.Collections[collectionName]
	xx.groupLock.RUnlock()
	return exists
}

func (xx *HighSpeedMem) CreateCollection(collectionName string, cfg CollectionConfig) error {
	c := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- fmt.Errorf(panicr, r)
			}
		}()
		// check col
		ok := xx.existsCollection(collectionName)
		if ok {
			// already collection exists
			c <- errors.New("already exists collection")
			return
		}
		xx.groupLock.Lock()
		xx.Collections[collectionName] = &CollectionMem{
			Data:            make(map[uint64]interface{}),
			CollectionName:  collectionName,
			Distance:        cfg.Distance,
			Quantization:    cfg.Quantization,
			Dim:             cfg.Dim,
			ExpansionAdd:    cfg.ExpansionAdd,
			ExpansionSearch: cfg.ExpansionSearch,
			Multi:           cfg.Multi,
			Storage:         cfg.Storage,
		}
		xx.groupLock.Unlock()

		//=========== vector build============//
		err := tensorLinker.CreateTensorIndex(collectionName, cfg)
		if err != nil {
			xx.groupLock.Lock()
			delete(xx.Collections, collectionName)
			xx.groupLock.Unlock()
			c <- err
			return
		}
		c <- nil
	}()
	return <-c
}

func (xx *HighSpeedMem) DropCollection(collectionName string) error {
	c := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- fmt.Errorf(panicr, r)
			}
		}()
		ok := xx.existsCollection(collectionName)
		if !ok {
			c <- nil
			return
		}
		xx.groupLock.Lock()
		delete(xx.Collections, collectionName)
		xx.groupLock.Unlock()
		c <- nil
	}()
	return <-c
}

func (xx *HighSpeedMem) GetCollection(collectionName string) (CollectionConfig, error) {
	type cc struct {
		Result CollectionConfig
		Error  error
	}
	c := make(chan cc, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- cc{
					Error: fmt.Errorf(panicr, r),
				}
			}
		}()
		col := xx.getCollection(collectionName)
		if col == nil {
			c <- cc{
				Error: fmt.Errorf("not found collection %s", collectionName),
			}
			return
		}
		c <- cc{
			Result: CollectionConfig{
				CollectionName:  col.CollectionName,
				Distance:        col.Distance,
				Quantization:    col.Quantization,
				Dim:             col.Dim,
				ExpansionAdd:    col.ExpansionAdd,
				ExpansionSearch: col.ExpansionSearch,
				Multi:           col.Multi,
				Storage:         col.Storage,
			},
		}
	}()

	out := <-c
	return out.Result, out.Error
}

func (xx *HighSpeedMem) GetCollections() ([]CollectionConfig, error) {
	type cc struct {
		Result []CollectionConfig
		Error  error
	}
	c := make(chan cc, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- cc{
					Error: fmt.Errorf(panicr, r),
				}
			}
		}()
		ccs := make([]CollectionConfig, 0)
		xx.groupLock.RLock()
		for collectionName := range xx.Collections {
			col := xx.getCollection(collectionName)
			if col == nil {
				c <- cc{
					Error: fmt.Errorf("not found collection %s", collectionName),
				}
				return
			}
			ccs = append(ccs, CollectionConfig{
				CollectionName:  col.CollectionName,
				Distance:        col.Distance,
				Quantization:    col.Quantization,
				Dim:             col.Dim,
				ExpansionAdd:    col.ExpansionAdd,
				ExpansionSearch: col.ExpansionSearch,
				Multi:           col.Multi,
				Storage:         col.Storage,
			})
		}
		xx.groupLock.RUnlock()
		c <- cc{
			Result: ccs,
		}
	}()
	out := <-c
	return out.Result, out.Error
}

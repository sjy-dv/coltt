package edge

import (
	"compress/flate"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/gomath"
)

type f16vecSpace struct {
	dimension      int
	vectors        map[uint64]float16Vec
	collectionName string
	distance       distance.Space
	quantization   Float16Quantization
	lock           sync.RWMutex
}

func newF16Vectorstore(config CollectionConfig) *f16vecSpace {
	return &f16vecSpace{
		dimension:      config.Dimension,
		vectors:        make(map[uint64]float16Vec),
		collectionName: config.CollectionName,
		distance: func() distance.Space {
			if config.Distance == COSINE {
				return distance.NewCosine()
			} else if config.Distance == EUCLIDEAN {
				return distance.NewEuclidean()
			}
			return distance.NewCosine()
		}(),
		quantization: Float16Quantization{},
	}
}

func (qx *f16vecSpace) InsertVector(collectionName string, commitId uint64, vector gomath.Vector) error {

	lower, err := qx.quantization.Lower(vector)
	if err != nil {
		return fmt.Errorf(ErrQuantizedFailed, err)
	}
	qx.lock.Lock()
	qx.vectors[commitId] = lower
	qx.lock.Unlock()
	return nil
}

func (qx *f16vecSpace) UpdateVector(collectionName string, id uint64, vector gomath.Vector) error {

	lower, err := qx.quantization.Lower(vector)
	if err != nil {
		return fmt.Errorf(ErrQuantizedFailed, err)
	}
	qx.lock.Lock()
	qx.vectors[id] = lower
	qx.lock.Unlock()
	return nil
}

func (qx *f16vecSpace) RemoveVector(collectionName string, id uint64) error {
	qx.lock.Lock()
	delete(qx.vectors, id)
	qx.lock.Unlock()
	return nil
}

func (qx *f16vecSpace) FullScan(collectionName string, target gomath.Vector, topK int,
) (*ResultSet, error) {
	rs := NewResultSet(topK)

	lower, err := qx.quantization.Lower(target)
	if err != nil {
		return nil, fmt.Errorf(ErrQuantizedFailed, err)
	}
	qx.lock.RLock()
	defer qx.lock.RUnlock()
	for index, qvec := range qx.vectors {
		sim := qx.quantization.Similarity(lower, qvec, qx.distance)
		rs.AddResult(ID(index), sim)
	}
	return rs, nil
}

func (qx *f16vecSpace) Commit() error {
	_, err := os.Stat(fmt.Sprintf(edgeVector, qx.collectionName))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		os.Remove(fmt.Sprintf(edgeVector, qx.collectionName))

	}
	var iow io.Writer
	qx.lock.RLock()
	flushData := qx.vectors
	qx.lock.RUnlock()
	f, err := os.OpenFile(fmt.Sprintf(edgeVector, qx.collectionName), os.O_TRUNC|
		os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	iow, _ = flate.NewWriter(f, flate.BestCompression)
	enc := gob.NewEncoder(iow)
	if err := enc.Encode(flushData); err != nil {
		return err
	}
	if flusher, ok := iow.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			return err
		}
	}
	if err := iow.(io.Closer).Close(); err != nil {
		return err
	}
	return nil
}

func (qx *f16vecSpace) Load() error {
	_, err := os.Stat(fmt.Sprintf(edgeVector, qx.collectionName))
	if err != nil {
		if os.IsNotExist(err) {
			// for _, col := range collections {
			// 	if col == collectionName {
			// 		goto EmptyData
			// 	}
			// }
			if stateManager.checker.collections[qx.collectionName] {
				goto EmptyData
			}
			return fmt.Errorf("collection[vector]: %s is not defined [Not Found Collection Error]", qx.collectionName)
		}
		return err
	}
	goto ExistsData
EmptyData:
	qx.vectors = make(map[uint64]float16Vec)
	return nil
ExistsData:
	commitCdat, err := os.OpenFile(fmt.Sprintf(edgeVector, qx.collectionName), os.O_RDONLY, 0777)
	if err != nil {
		// cdat is damaged
		// after add recovery logic
		return err
	}
	cdat := make(map[uint64]float16Vec)

	var readIo io.Reader

	readIo = flate.NewReader(commitCdat)

	dataDec := gob.NewDecoder(readIo)
	err = dataDec.Decode(&cdat)
	if err != nil {
		// also cdat is damaged guess
		return err
	}
	err = readIo.(io.Closer).Close()
	if err != nil {
		return err
	}
	qx.lock.Lock()
	qx.vectors = cdat
	qx.lock.Unlock()
	return nil
}

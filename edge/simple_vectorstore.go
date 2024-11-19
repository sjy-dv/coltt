package edge

import (
	"sync"

	"github.com/sjy-dv/nnv/pkg/concurrentmap"
	"github.com/sjy-dv/nnv/pkg/distance"
)

type simplevecSpace struct {
	dimension int
	// vectors        map[uint64]Vector
	vectors        *concurrentmap.Map[uint64, Vector]
	collectionName string
	distance       distance.Space
	quantization   NoQuantization
	lock           sync.RWMutex
}

func newSimpleVectorstore(config CollectionConfig) *simplevecSpace {
	return &simplevecSpace{
		dimension:      config.Dimension,
		vectors:        concurrentmap.New[uint64, Vector](),
		collectionName: config.CollectionName,
		distance: func() distance.Space {
			if config.Distance == COSINE {
				return distance.NewCosine()
			} else if config.Distance == EUCLIDEAN {
				return distance.NewEuclidean()
			}
			return distance.NewCosine()
		}(),
		quantization: NoQuantization{},
	}
}

func (qx *simplevecSpace) InsertVector(collectionName string, commitId uint64, vector Vector) error {

	// qx.lock.Lock()
	// qx.vectors[commitId] = vector
	// qx.lock.Unlock()
	qx.vectors.Set(commitId, vector)
	return nil
}

func (qx *simplevecSpace) UpdateVector(collectionName string, id uint64, vector Vector) error {

	// qx.lock.Lock()
	// qx.vectors[id] = vector
	// qx.lock.Unlock()
	qx.vectors.Set(id, vector)
	return nil
}

func (qx *simplevecSpace) RemoveVector(collectionName string, id uint64) error {
	// qx.lock.Lock()
	// delete(qx.vectors, id)
	// qx.lock.Unlock()
	qx.vectors.Del(id)
	return nil
}

func (qx *simplevecSpace) FullScan(collectionName string, target Vector, topK int,
) (*ResultSet, error) {
	rs := NewResultSet(topK)

	// qx.lock.RLock()
	// defer qx.lock.RUnlock()
	// for index, qvec := range qx.vectors {
	// 	sim := qx.quantization.Similarity(target, qvec, qx.distance)
	// 	rs.AddResult(ID(index), sim)
	// }
	qx.vectors.ForEach(func(u uint64, v Vector) bool {
		sim := qx.quantization.Similarity(target, v, qx.distance)
		rs.AddResult(ID(u), sim)
		return true
	})
	return rs, nil
}

// func (qx *simplevecSpace) Commit() error {
// 	_, err := os.Stat(fmt.Sprintf(edgeVector, qx.collectionName))
// 	if err != nil {
// 		if !os.IsNotExist(err) {
// 			return err
// 		}
// 	} else {
// 		os.Remove(fmt.Sprintf(edgeVector, qx.collectionName))

// 	}
// 	var iow io.Writer
// 	qx.lock.RLock()
// 	flushData := qx.vectors
// 	qx.lock.RUnlock()
// 	f, err := os.OpenFile(fmt.Sprintf(edgeVector, qx.collectionName), os.O_TRUNC|
// 		os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	iow, _ = flate.NewWriter(f, flate.BestCompression)
// 	enc := gob.NewEncoder(iow)
// 	if err := enc.Encode(flushData); err != nil {
// 		return err
// 	}
// 	if flusher, ok := iow.(interface{ Flush() error }); ok {
// 		if err := flusher.Flush(); err != nil {
// 			return err
// 		}
// 	}
// 	if err := iow.(io.Closer).Close(); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (qx *simplevecSpace) Load() error {
// 	_, err := os.Stat(fmt.Sprintf(edgeVector, qx.collectionName))
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			// for _, col := range collections {
// 			// 	if col == collectionName {
// 			// 		goto EmptyData
// 			// 	}
// 			// }
// 			if stateManager.checker.collections[qx.collectionName] {
// 				goto EmptyData
// 			}
// 			return fmt.Errorf("collection[vector]: %s is not defined [Not Found Collection Error]", qx.collectionName)
// 		}
// 		return err
// 	}
// 	goto ExistsData
// EmptyData:
// 	qx.vectors = make(map[uint64]Vector)
// 	return nil
// ExistsData:
// 	commitCdat, err := os.OpenFile(fmt.Sprintf(edgeVector, qx.collectionName), os.O_RDONLY, 0777)
// 	if err != nil {
// 		// cdat is damaged
// 		// after add recovery logic
// 		return err
// 	}
// 	cdat := make(map[uint64]Vector)

// 	var readIo io.Reader

// 	readIo = flate.NewReader(commitCdat)

// 	dataDec := gob.NewDecoder(readIo)
// 	err = dataDec.Decode(&cdat)
// 	if err != nil {
// 		// also cdat is damaged guess
// 		return err
// 	}
// 	err = readIo.(io.Closer).Close()
// 	if err != nil {
// 		return err
// 	}
// 	qx.lock.Lock()
// 	qx.vectors = cdat
// 	qx.lock.Unlock()
// 	return nil
// }

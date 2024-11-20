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

package edge

import (
	"compress/flate"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/sjy-dv/nnv/gen/protoc/v2/phonyproto"
	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/gomath"
	"github.com/sjy-dv/nnv/pkg/index"
	"google.golang.org/protobuf/proto"
)

func init() {
	gob.Register(gomath.Vector{})
	gob.Register(map[uint64]gomath.Vector{})
	gob.Register(float16Vec{})
	gob.Register(map[uint64]float16Vec{})
	gob.Register(float8Vec{})
	gob.Register(map[uint64]float8Vec{})
	gob.Register(bfloat16Vec{})
	gob.Register(map[uint64]bfloat16Vec{})
	gob.Register(map[string]interface{}{})
	gob.Register(map[uint64]interface{}{})
	gob.Register([]interface{}{})

}

func (xx *Edge) LoadData(collectionName string, config CollectionConfig) error {
	if config.Quantization == F8_QUANTIZATION {
		vecspace := newF8Vectorstore(config)
		iserror := false
		sep := fmt.Sprintf("%s_", collectionName)
		var werr error
		xx.Disk.AscendKeys([]byte(sep), true, func(k []byte) (bool, error) {
			key, err := strconv.Atoi(strings.Split(string(k), sep)[1])
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}
			val, err := xx.Disk.Get(k)
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}
			phony := phonyproto.PhonyWrapper{}
			err = proto.Unmarshal(val, &phony)
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}
			err = vecspace.InsertVector(collectionName, uint64(key), phony.GetVector())
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}

			return true, nil
		})
		if iserror {
			return werr
		}
		xx.VectorStore.slock.Lock()
		defer xx.VectorStore.slock.Unlock()
		xx.VectorStore.Space[collectionName] = vecspace
	} else if config.Quantization == F16_QUANTIZATION {
		vecspace := newF16Vectorstore(config)
		iserror := false
		sep := fmt.Sprintf("%s_", collectionName)
		var werr error
		xx.Disk.AscendKeys([]byte(sep), true, func(k []byte) (bool, error) {
			key, err := strconv.Atoi(strings.Split(string(k), sep)[1])
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}
			val, err := xx.Disk.Get(k)
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}
			phony := phonyproto.PhonyWrapper{}
			err = proto.Unmarshal(val, &phony)
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}
			err = vecspace.InsertVector(collectionName, uint64(key), phony.GetVector())
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}
			// xx.Datas[collectionName].lock.Lock()
			// xx.Datas[collectionName].Data[uint64(key)] = phony.GetMetadata().AsMap()
			// xx.Datas[collectionName].lock.Unlock()
			return true, nil
		})
		if iserror {
			return werr
		}
		xx.VectorStore.slock.Lock()
		defer xx.VectorStore.slock.Unlock()
		xx.VectorStore.Space[collectionName] = vecspace
	} else if config.Quantization == BF16_QUANTIZATION {
		vecspace := newBF16Vectorstore(config)
		iserror := false
		sep := fmt.Sprintf("%s_", collectionName)
		var werr error
		xx.Disk.AscendKeys([]byte(sep), true, func(k []byte) (bool, error) {
			key, err := strconv.Atoi(strings.Split(string(k), sep)[1])
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}
			val, err := xx.Disk.Get(k)
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}
			phony := phonyproto.PhonyWrapper{}
			err = proto.Unmarshal(val, &phony)
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}
			err = vecspace.InsertVector(collectionName, uint64(key), phony.GetVector())
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}
			// xx.Datas[collectionName].lock.Lock()
			// xx.Datas[collectionName].Data[uint64(key)] = phony.GetMetadata().AsMap()
			// xx.Datas[collectionName].lock.Unlock()
			return true, nil
		})
		if iserror {
			return werr
		}
		xx.VectorStore.slock.Lock()
		defer xx.VectorStore.slock.Unlock()
		xx.VectorStore.Space[collectionName] = vecspace
	} else if config.Quantization == NONE_QAUNTIZATION {
		vecspace := newSimpleVectorstore(config)
		iserror := false
		sep := fmt.Sprintf("%s_", collectionName)
		var werr error
		xx.Disk.AscendKeys([]byte(sep), true, func(k []byte) (bool, error) {
			key, err := strconv.Atoi(strings.Split(string(k), sep)[1])
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}
			val, err := xx.Disk.Get(k)
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}
			phony := phonyproto.PhonyWrapper{}
			err = proto.Unmarshal(val, &phony)
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}
			err = vecspace.InsertVector(collectionName, uint64(key), phony.GetVector())
			if err != nil {
				iserror = true
				werr = err
				return false, err
			}
			// xx.Datas[collectionName].lock.Lock()
			// xx.Datas[collectionName].Data[uint64(key)] = phony.GetMetadata().AsMap()
			// xx.Datas[collectionName].lock.Unlock()
			return true, nil
		})
		if iserror {
			return werr
		}
		xx.VectorStore.slock.Lock()
		defer xx.VectorStore.slock.Unlock()
		xx.VectorStore.Space[collectionName] = vecspace
	} else {
		return errors.New("not support quantization type")
	}
	return nil
}

func (xx *Edge) CommitConfig(collectionName string) error {
	_, err := os.Stat(fmt.Sprintf(edgeConfig, collectionName))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		os.Remove(fmt.Sprintf(edgeConfig, collectionName))
	}
	conf := CollectionConfig{}
	conf.CollectionName = collectionName
	// xx.lock.RLock()
	// xx.Datas[collectionName].lock.RLock()
	cfg, ok := xx.Datas.Get(collectionName)
	if !ok {
		return errors.New("loss data in map")
	}
	conf.Dimension = int(cfg.dim)
	conf.Distance = cfg.distance
	conf.Quantization = cfg.quantization
	// xx.lock.RUnlock()
	// xx.Datas[collectionName].lock.RUnlock()
	f, err := os.Create(fmt.Sprintf(edgeConfig, collectionName))
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.Encode(conf)
	defer f.Close()
	return nil
}

func (xx *Edge) CommitIndex(collectionName string) error {
	_, err := os.Stat(fmt.Sprintf(edgeIndex, collectionName))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		os.Remove(fmt.Sprintf(edgeIndex, collectionName))

	}
	return indexdb.indexes[collectionName].SerializeBinary(fmt.Sprintf(edgeIndex, collectionName))
}

func (xx *Edge) CommitVector(collectionName string) error {
	_, err := os.Stat(fmt.Sprintf(edgeVector, collectionName))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		os.Remove(fmt.Sprintf(edgeVector, collectionName))

	}
	var iow io.Writer
	xx.lock.RLock()
	xx.VectorStore.slock.RLock()
	flushData := xx.VectorStore.Space[collectionName]
	xx.lock.RUnlock()
	xx.VectorStore.slock.RUnlock()

	f, err := os.OpenFile(fmt.Sprintf(edgeVector, collectionName), os.O_TRUNC|
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

func (xx *Edge) CommitCollection() error {
	_, err := os.Stat(collectionEdgeJson)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		os.Remove(collectionEdgeJson)
	}
	type collectionJsonF struct {
		Collections []string
	}
	c := collectionJsonF{}
	cols := make([]string, 0)
	for c, ok := range stateManager.checker.collections {
		if ok {
			cols = append(cols, c)
		}
	}
	c.Collections = cols
	f, err := os.Create(collectionEdgeJson)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.Encode(c)
	defer f.Close()
	return nil
}

func (xx *Edge) LoadCommitCollection() error {
	collectionJsonData, err := os.ReadFile(collectionEdgeJson)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	type collectionJsonF struct {
		Collections []string
	}
	c := collectionJsonF{}
	if err := json.Unmarshal(collectionJsonData, &c); err != nil {
		return err
	}
	for _, col := range c.Collections {
		stateManager.checker.collections[col] = true
	}
	return nil
}

func (xx *Edge) LoadCommitData(collectionName string) (map[uint64]interface{}, error) {
	_, err := os.Stat(fmt.Sprintf(edgeData, collectionName))
	if err != nil {
		if os.IsNotExist(err) {
			if stateManager.checker.collections[collectionName] {
				return make(map[uint64]interface{}), nil
			}
			return nil, fmt.Errorf("collection: %s is not defined [Not Found Collection Error]", collectionName)
		}
		return nil, err
	}
	commitCdat, err := os.OpenFile(fmt.Sprintf(edgeData, collectionName), os.O_RDONLY, 0777)
	if err != nil {
		// cdat is damaged
		// after add recovery logic
		return nil, err
	}
	cdat := make(map[uint64]interface{})

	var readIo io.Reader

	readIo = flate.NewReader(commitCdat)

	dataDec := gob.NewDecoder(readIo)
	err = dataDec.Decode(&cdat)
	if err != nil {
		return nil, err
	}
	err = readIo.(io.Closer).Close()
	if err != nil {
		return nil, err
	}
	return cdat, nil
}

func (xx *Edge) LoadCommitCollectionConifg(collectionName string) (
	CollectionConfig, error,
) {
	configJsonData, err := os.ReadFile(fmt.Sprintf(edgeConfig, collectionName))
	if err != nil {
		if os.IsNotExist(err) {
			if stateManager.checker.collections[collectionName] {
				//damaged data file
				// recovery logic add
				return CollectionConfig{}, fmt.Errorf("file %s.conf is damaged", collectionName)
			}

			return CollectionConfig{}, fmt.Errorf("collection Config: %s is not defined [Not Found Collection Error]", collectionName)
		}
		return CollectionConfig{}, err
	}
	cfg := CollectionConfig{}
	if err := json.Unmarshal(configJsonData, &cfg); err != nil {
		return CollectionConfig{}, err
	}
	return cfg, nil
}

func (xx *Edge) LoadCommitIndex(collectionName string) error {
	_, err := os.Stat(fmt.Sprintf(edgeIndex, collectionName))
	if err != nil {
		if os.IsNotExist(err) {
			if stateManager.checker.collections[collectionName] {
				goto EmptyIndex
			}

			return fmt.Errorf("collection BitmapIndex: %s is not defined [Not Found Collection Error]", collectionName)
		}
		return err
	}
	goto ExistsIndex
EmptyIndex:
	indexdb.indexLock.Lock()
	indexdb.indexes[collectionName] = index.NewBitmapIndex()
	defer indexdb.indexLock.Unlock()
	return nil
ExistsIndex:
	recoveryIndex := index.NewBitmapIndex()
	err = recoveryIndex.DeserializeBinary(fmt.Sprintf(edgeIndex, collectionName))
	if err != nil {
		// guess damaged file
		return err
	}
	err = recoveryIndex.ValidateIndex()
	if err != nil {
		// validation failed also damaged
		return err
	}
	indexdb.indexLock.Lock()
	indexdb.indexes[collectionName] = recoveryIndex
	defer indexdb.indexLock.Unlock()
	return nil
}

func (xx *Edge) LoadCommitNormalVector(collectionName string, cfg CollectionConfig) error {
	_, err := os.Stat(fmt.Sprintf(edgeVector, collectionName))
	if err != nil {
		if os.IsNotExist(err) {
			// for _, col := range collections {
			// 	if col == collectionName {
			// 		goto EmptyData
			// 	}
			// }
			if stateManager.checker.collections[collectionName] {
				goto EmptyData
			}
			return fmt.Errorf("collection[vector]: %s is not defined [Not Found Collection Error]", collectionName)
		}
		return err
	}
	goto ExistsData
EmptyData:
	normalEdgeV.lock.Lock()
	normalEdgeV.Edges[collectionName] = &EdgeVector{
		dimension:      cfg.Dimension,
		vectors:        make(map[uint64]gomath.Vector),
		collectionName: collectionName,
		distance: func() distance.Space {
			if cfg.Distance == COSINE {
				return distance.NewCosine()
			} else if cfg.Distance == EUCLIDEAN {
				return distance.NewEuclidean()
			}
			return distance.NewCosine()
		}(),
	}
	normalEdgeV.lock.Unlock()
	return nil
ExistsData:
	commitCdat, err := os.OpenFile(fmt.Sprintf(edgeVector, collectionName), os.O_RDONLY, 0777)
	if err != nil {
		// cdat is damaged
		// after add recovery logic
		return err
	}
	cdat := make(map[uint64]gomath.Vector)

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
	normalEdgeV.lock.Lock()
	normalEdgeV.Edges[collectionName] = &EdgeVector{
		dimension:      cfg.Dimension,
		vectors:        cdat,
		collectionName: collectionName,
		distance: func() distance.Space {
			if cfg.Distance == COSINE {
				return distance.NewCosine()
			} else if cfg.Distance == EUCLIDEAN {
				return distance.NewEuclidean()
			}
			return distance.NewCosine()
		}(),
	}
	normalEdgeV.lock.Unlock()
	return nil
}

func allremover(collectionName string) {
	os.Remove(fmt.Sprintf(edgeData, collectionName))
	os.Remove(fmt.Sprintf(edgeIndex, collectionName))
	os.Remove(fmt.Sprintf(edgeVector, collectionName))
	os.Remove(fmt.Sprintf(edgeConfig, collectionName))
}

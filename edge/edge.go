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
	"context"
	"fmt"
	"math"
	"slices"
	"sync"

	"github.com/sjy-dv/nnv/gen/protoc/v2/edgeproto"
	"google.golang.org/protobuf/types/known/structpb"
)

type Edge struct {
	Datas map[string]*EdgeData
	lock  sync.RWMutex
}

type EdgeData struct {
	Data         map[uint64]interface{}
	dim          int32
	distance     string
	quantization string
	lock         sync.RWMutex
}

func NewEdge() *Edge {
	return &Edge{
		Datas: make(map[string]*EdgeData),
	}
}

func existsCollection(collectionName string) bool {
	stateManager.checker.cecLock.RLock()
	exists := stateManager.checker.collections[collectionName]
	stateManager.checker.cecLock.RUnlock()
	return exists
}

func alreadyLoadCollection(collectionName string) bool {
	stateManager.auth.authLock.RLock()
	exists := stateManager.auth.collections[collectionName]
	stateManager.auth.authLock.RUnlock()
	return exists
}

func (xx *Edge) getQuantization(collectionName string) string {
	xx.lock.RLock()
	q := xx.Datas[collectionName].quantization
	defer xx.lock.RUnlock()
	return q
}

func (xx *Edge) getDist(collectionName string) string {
	xx.lock.RLock()
	dist := xx.Datas[collectionName].distance
	defer xx.lock.RUnlock()
	return dist
}

func (xx *Edge) CreateCollection(ctx context.Context, req *edgeproto.Collection) (
	*edgeproto.CollectionResponse, error) {
	type reply struct {
		Result *edgeproto.CollectionResponse
		Error  error
	}
	c := make(chan reply, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Error: fmt.Errorf(panicr, r),
				}
			}
		}()
		//scripts
		if existsCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.CollectionResponse{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionExists, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		dist := func() string {
			if req.GetDistance() == edgeproto.Distance_Cosine {
				return COSINE
			}
			return EUCLIDEAN
		}()
		q := func() string {
			if req.GetQuantization() == edgeproto.Quantization_F16 {
				return F16_QUANTIZATION
			}
			return NONE_QAUNTIZATION
		}()
		xx.lock.Lock()
		xx.Datas[req.GetCollectionName()] = &EdgeData{
			dim:          int32(req.GetDim()),
			distance:     dist,
			quantization: q,
			Data:         make(map[uint64]interface{}),
		}
		xx.lock.Unlock()
		//=========vector============
		cfg := CollectionConfig{
			Dimension:      int(req.GetDim()),
			CollectionName: req.GetCollectionName(),
			Distance:       dist,
			Quantization:   q,
		}
		if q == NONE_QAUNTIZATION {
			err := normalEdgeV.CreateCollection(cfg)
			if err != nil {
				xx.lock.Lock()
				delete(xx.Datas, req.GetCollectionName())
				xx.lock.Unlock()
				c <- reply{
					Result: &edgeproto.CollectionResponse{
						Status: false,
						Error:  &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR},
					},
				}
				return
			}
		} else {
			err := quantizedEdgeV.CreateCollection(cfg)
			if err != nil {
				xx.lock.Lock()
				delete(xx.Datas, req.GetCollectionName())
				xx.lock.Unlock()
				c <- reply{
					Result: &edgeproto.CollectionResponse{
						Status: false,
						Error:  &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR},
					},
				}
				return
			}
		}
		//bitmap
		err := indexdb.CreateIndex(req.GetCollectionName())
		if err != nil {
			xx.lock.Lock()
			delete(xx.Datas, req.GetCollectionName())
			xx.lock.Unlock()
			if q == NONE_QAUNTIZATION {
				normalEdgeV.DropCollection(req.GetCollectionName())
			} else {
				quantizedEdgeV.DropCollection(req.GetCollectionName())
			}
			c <- reply{
				Result: &edgeproto.CollectionResponse{
					Status: false,
					Error:  &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR},
				},
			}
			return
		}
		stateManager.checker.cecLock.Lock()
		stateManager.checker.collections[req.GetCollectionName()] = true
		stateManager.checker.cecLock.Unlock()
		stateManager.loadchecker.clcLock.Lock()
		stateManager.loadchecker.collections[req.GetCollectionName()] = true
		stateManager.loadchecker.clcLock.Unlock()
		stateManager.auth.authLock.Lock()
		stateManager.auth.collections[req.GetCollectionName()] = true
		stateManager.auth.authLock.Unlock()
		err = xx.CommitCollection()
		if err != nil {
			xx.lock.Lock()
			delete(xx.Datas, req.GetCollectionName())
			xx.lock.Unlock()
			if q == NONE_QAUNTIZATION {
				normalEdgeV.DropCollection(req.GetCollectionName())
			} else {
				quantizedEdgeV.DropCollection(req.GetCollectionName())
			}
			c <- reply{
				Result: &edgeproto.CollectionResponse{
					Status: false,
					Error:  &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR},
				},
			}
			return
		}
		c <- reply{
			Result: &edgeproto.CollectionResponse{
				Status: true,
				Collection: &edgeproto.Collection{
					CollectionName: req.GetCollectionName(),
					Distance:       req.Distance,
					Quantization:   req.Quantization,
					Dim:            req.GetDim(),
				},
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *Edge) DeleteCollection(ctx context.Context, req *edgeproto.CollectionName) (
	*edgeproto.DeleteCollectionResponse, error) {
	type reply struct {
		Result *edgeproto.DeleteCollectionResponse
		Error  error
	}

	c := make(chan reply, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Error: fmt.Errorf(panicr, r),
				}
			}
		}()
		if !existsCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.DeleteCollectionResponse{
					Status: true,
				},
			}
			return
		}
		xx.lock.Lock()
		q := xx.Datas[req.CollectionName].quantization
		delete(xx.Datas, req.GetCollectionName())
		xx.lock.Unlock()

		stateManager.auth.authLock.Lock()
		delete(stateManager.auth.collections, req.GetCollectionName())
		stateManager.auth.authLock.Unlock()

		stateManager.checker.cecLock.Lock()
		delete(stateManager.checker.collections, req.GetCollectionName())
		stateManager.checker.cecLock.Unlock()

		stateManager.loadchecker.clcLock.Lock()
		delete(stateManager.loadchecker.collections, req.GetCollectionName())
		stateManager.loadchecker.clcLock.Unlock()

		var err error
		if q == NONE_QAUNTIZATION {
			err = normalEdgeV.DropCollection(req.GetCollectionName())
		} else {
			err = quantizedEdgeV.DropCollection(req.GetCollectionName())
		}
		if err != nil {
			c <- reply{
				Result: &edgeproto.DeleteCollectionResponse{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		err = indexdb.DropIndex(req.GetCollectionName())
		if err != nil {
			c <- reply{
				Result: &edgeproto.DeleteCollectionResponse{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		allremover(req.GetCollectionName())
		c <- reply{
			Result: &edgeproto.DeleteCollectionResponse{
				Status: true,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *Edge) GetCollection(ctx context.Context, req *edgeproto.CollectionName) (
	*edgeproto.CollectionDetail, error) {
	type reply struct {
		Result *edgeproto.CollectionDetail
		Error  error
	}
	c := make(chan reply, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Error: fmt.Errorf(panicr, r),
				}
			}
		}()

		if !existsCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.CollectionDetail{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}

		c <- reply{
			Result: &edgeproto.CollectionDetail{
				Status: true,
				Collection: &edgeproto.Collection{
					CollectionName: req.GetCollectionName(),
				},
			},
		}
	}()
	out := <-c
	return out.Result, out.Error
}

func (xx *Edge) LoadCollection(ctx context.Context, req *edgeproto.CollectionName) (
	*edgeproto.CollectionDetail, error) {
	type reply struct {
		Result *edgeproto.CollectionDetail
		Error  error
	}
	c := make(chan reply, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Error: fmt.Errorf(panicr, r),
				}
			}
		}()
		if !existsCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.CollectionDetail{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		if alreadyLoadCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.CollectionDetail{
					Status: true,
					Collection: &edgeproto.Collection{
						CollectionName: req.GetCollectionName(),
					},
				},
			}
			return
		}
		loadData, err := xx.LoadCommitData(req.GetCollectionName())
		if err != nil {
			c <- reply{Result: &edgeproto.CollectionDetail{Status: false, Error: &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR}}}
			return
		}
		loadConfig, err := xx.LoadCommitCollectionConifg(req.GetCollectionName())
		if err != nil {
			c <- reply{Result: &edgeproto.CollectionDetail{Status: false, Error: &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR}}}
			return
		}
		err = xx.LoadCommitIndex(req.GetCollectionName())
		if err != nil {
			c <- reply{Result: &edgeproto.CollectionDetail{Status: false, Error: &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR}}}
			return
		}
		if loadConfig.Quantization == NONE_QAUNTIZATION {
			err = xx.LoadCommitNormalVector(req.GetCollectionName(), loadConfig)
		} else {
			err = xx.LoadCommitQuantizedVector(req.GetCollectionName(), loadConfig)
		}
		if err != nil {
			c <- reply{Result: &edgeproto.CollectionDetail{Status: false, Error: &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR}}}
			return
		}
		merge := &EdgeData{
			Data:         loadData,
			dim:          int32(loadConfig.Dimension),
			distance:     loadConfig.Distance,
			quantization: loadConfig.Quantization,
		}
		xx.lock.Lock()
		xx.Datas[req.GetCollectionName()] = &EdgeData{}
		xx.Datas[req.GetCollectionName()] = merge
		xx.lock.Unlock()
		stateManager.loadchecker.clcLock.Lock()
		stateManager.loadchecker.collections[req.GetCollectionName()] = true
		stateManager.loadchecker.clcLock.Unlock()
		stateManager.auth.authLock.Lock()
		stateManager.auth.collections[req.GetCollectionName()] = true
		stateManager.auth.authLock.Unlock()
		c <- reply{
			Result: &edgeproto.CollectionDetail{
				Status: true,
				Collection: &edgeproto.Collection{
					CollectionName: req.GetCollectionName(),
				},
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *Edge) ReleaseCollection(ctx context.Context, req *edgeproto.CollectionName) (
	*edgeproto.Response, error) {
	type reply struct {
		Result *edgeproto.Response
		Error  error
	}
	c := make(chan reply, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Error: fmt.Errorf(panicr, r),
				}
			}
		}()
		if !existsCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.Response{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		if !alreadyLoadCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.Response{
					Status: true,
				},
			}
			return
		}
		err := xx.CommitData(req.GetCollectionName())
		if err != nil {
			c <- reply{Result: &edgeproto.Response{Status: false, Error: &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR}}}
			return
		}
		err = xx.CommitConfig(req.GetCollectionName())
		if err != nil {
			c <- reply{Result: &edgeproto.Response{Status: false, Error: &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR}}}
			return
		}
		err = xx.CommitIndex(req.GetCollectionName())
		if err != nil {
			c <- reply{Result: &edgeproto.Response{Status: false, Error: &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR}}}
			return
		}
		q := xx.getQuantization(req.GetCollectionName())
		if q == NONE_QAUNTIZATION {
			err = xx.CommitNormalVector(req.GetCollectionName())
		} else {
			err = xx.CommitQuantizedVector(req.GetCollectionName())
		}
		if err != nil {
			c <- reply{Result: &edgeproto.Response{Status: false, Error: &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR}}}
			return
		}
		xx.lock.Lock()
		delete(xx.Datas, req.GetCollectionName())
		xx.lock.Unlock()

		indexdb.indexLock.Lock()
		delete(indexdb.indexes, req.GetCollectionName())
		indexdb.indexLock.Unlock()

		if q == NONE_QAUNTIZATION {
			normalEdgeV.lock.Lock()
			delete(normalEdgeV.Edges, req.GetCollectionName())
			normalEdgeV.lock.Unlock()
		} else {
			quantizedEdgeV.lock.Lock()
			delete(quantizedEdgeV.Edges, req.GetCollectionName())
			quantizedEdgeV.lock.Unlock()
		}
		stateManager.loadchecker.clcLock.Lock()
		stateManager.loadchecker.collections[req.GetCollectionName()] = false
		stateManager.loadchecker.clcLock.Unlock()
		stateManager.auth.authLock.Lock()
		stateManager.auth.collections[req.GetCollectionName()] = false
		stateManager.auth.authLock.Unlock()
		c <- reply{
			Result: &edgeproto.Response{
				Status: true,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *Edge) Flush(ctx context.Context, req *edgeproto.CollectionName) (
	*edgeproto.Response, error) {
	type reply struct {
		Result *edgeproto.Response
		Error  error
	}
	c := make(chan reply, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Error: fmt.Errorf(panicr, r),
				}
			}
		}()
		if !existsCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.Response{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		if !alreadyLoadCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.Response{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotLoad, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}

		err := xx.CommitData(req.GetCollectionName())
		if err != nil {
			c <- reply{Result: &edgeproto.Response{Status: false, Error: &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR}}}
			return
		}
		err = xx.CommitConfig(req.GetCollectionName())
		if err != nil {
			c <- reply{Result: &edgeproto.Response{Status: false, Error: &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR}}}
			return
		}
		err = xx.CommitIndex(req.GetCollectionName())
		if err != nil {
			c <- reply{Result: &edgeproto.Response{Status: false, Error: &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR}}}
			return
		}
		q := xx.getQuantization(req.GetCollectionName())
		fmt.Println(q)
		if q == NONE_QAUNTIZATION {
			err = xx.CommitNormalVector(req.GetCollectionName())
		} else {
			err = xx.CommitQuantizedVector(req.GetCollectionName())
		}
		if err != nil {
			c <- reply{Result: &edgeproto.Response{Status: false, Error: &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR}}}
			return
		}
		c <- reply{
			Result: &edgeproto.Response{
				Status: true,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *Edge) Insert(ctx context.Context, req *edgeproto.ModifyDataset) (
	*edgeproto.Response, error) {
	type reply struct {
		Result *edgeproto.Response
		Error  error
	}
	c := make(chan reply, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Error: fmt.Errorf(panicr, r),
				}
			}
		}()
		if !existsCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.Response{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		if !alreadyLoadCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.Response{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotLoad, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		autoID := autoCommitID()
		cloneMap := req.GetMetadata().AsMap()
		xx.Datas[req.GetCollectionName()].lock.Lock()
		xx.Datas[req.GetCollectionName()].Data[autoID] = cloneMap
		xx.Datas[req.GetCollectionName()].lock.Unlock()

		err := indexdb.indexes[req.GetCollectionName()].Add(autoID, cloneMap)
		if err != nil {
			c <- reply{Result: &edgeproto.Response{Status: false, Error: &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR}}}
			return
		}
		q := xx.getQuantization(req.GetCollectionName())
		if q == NONE_QAUNTIZATION {
			err = normalEdgeV.InsertVector(req.GetCollectionName(), autoID, req.GetVector())
		} else {
			err = quantizedEdgeV.InsertVector(req.GetCollectionName(), autoID, req.GetVector())
		}
		if err != nil {
			c <- reply{Result: &edgeproto.Response{Status: false, Error: &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR}}}
			return
		}
		c <- reply{
			Result: &edgeproto.Response{
				Status: true,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *Edge) Update(ctx context.Context, req *edgeproto.ModifyDataset) (
	*edgeproto.Response, error) {
	type reply struct {
		Result   *edgeproto.Response
		IsCreate bool
		Error    error
	}
	c := make(chan reply, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Error: fmt.Errorf(panicr, r),
				}
			}
		}()
		if !existsCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.Response{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		if !alreadyLoadCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.Response{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotLoad, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		getId := indexdb.indexes[req.GetCollectionName()].PureSearch(map[string]string{"_id": req.GetId()})
		if len(getId) == 0 {
			c <- reply{
				IsCreate: true,
			}
			return
		}
		xx.Datas[req.GetCollectionName()].lock.RLock()
		cloneMeta := xx.Datas[req.GetCollectionName()].Data[getId[0]]
		xx.Datas[req.GetCollectionName()].lock.RUnlock()

		xx.Datas[req.GetCollectionName()].lock.Lock()
		xx.Datas[req.GetCollectionName()].Data[getId[0]] = req.GetMetadata().AsMap()
		xx.Datas[req.GetCollectionName()].lock.Unlock()

		err := indexdb.indexes[req.GetCollectionName()].Remove(getId[0], cloneMeta.(map[string]interface{}))
		if err != nil {
			c <- reply{
				Result: &edgeproto.Response{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		err = indexdb.indexes[req.GetCollectionName()].Add(getId[0], req.GetMetadata().AsMap())
		if err != nil {
			c <- reply{
				Result: &edgeproto.Response{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		q := xx.getQuantization(req.GetCollectionName())
		if q == NONE_QAUNTIZATION {
			err = normalEdgeV.UpdateVector(req.GetCollectionName(), getId[0], req.GetVector())
		} else {
			err = quantizedEdgeV.UpdateVector(req.GetCollectionName(), getId[0], req.GetVector())
		}
		if err != nil {
			c <- reply{
				Result: &edgeproto.Response{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		c <- reply{
			Result: &edgeproto.Response{
				Status: true,
			},
		}
	}()

	res := <-c
	if res.IsCreate {
		return xx.Insert(ctx, req)
	}
	return res.Result, res.Error
}

func (xx *Edge) Delete(ctx context.Context, req *edgeproto.DeleteDataset) (
	*edgeproto.Response, error) {
	type reply struct {
		Result *edgeproto.Response
		Error  error
	}
	c := make(chan reply, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Error: fmt.Errorf(panicr, r),
				}
			}
		}()
		if !existsCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.Response{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		if !alreadyLoadCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.Response{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotLoad, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		getId := indexdb.indexes[req.GetCollectionName()].PureSearch(map[string]string{"_id": req.GetId()})
		if len(getId) == 0 {
			c <- reply{
				Result: &edgeproto.Response{
					Status: true,
				},
			}
			return
		}
		xx.Datas[req.GetCollectionName()].lock.RLock()
		cloneMeta := xx.Datas[req.GetCollectionName()].Data[getId[0]]
		xx.Datas[req.GetCollectionName()].lock.RUnlock()

		xx.Datas[req.GetCollectionName()].lock.Lock()
		delete(xx.Datas[req.GetCollectionName()].Data, getId[0])
		xx.Datas[req.GetCollectionName()].lock.Unlock()

		err := indexdb.indexes[req.GetCollectionName()].Remove(getId[0], cloneMeta.(map[string]interface{}))
		if err != nil {
			c <- reply{
				Result: &edgeproto.Response{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		q := xx.getQuantization(req.GetCollectionName())
		if q == NONE_QAUNTIZATION {
			err = normalEdgeV.RemoveVector(req.GetCollectionName(), getId[0])
		} else {
			err = quantizedEdgeV.RemoveVector(req.GetCollectionName(), getId[0])
		}
		if err != nil {
			c <- reply{
				Result: &edgeproto.Response{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		c <- reply{
			Result: &edgeproto.Response{
				Status: true,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *Edge) VectorSearch(ctx context.Context, req *edgeproto.SearchReq) (
	*edgeproto.SearchResponse, error) {
	type reply struct {
		Result *edgeproto.SearchResponse
		Error  error
	}
	c := make(chan reply, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Error: fmt.Errorf(panicr, r),
				}
			}
		}()
		if !existsCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.SearchResponse{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		if !alreadyLoadCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.SearchResponse{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotLoad, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		var (
			rs  *ResultSet
			err error
		)
		if xx.getQuantization(req.GetCollectionName()) == NONE_QAUNTIZATION {
			rs, err = normalEdgeV.FullScan(req.GetCollectionName(), req.GetVector(), int(req.GetTopK()))
		} else {
			rs, err = quantizedEdgeV.FullScan(req.GetCollectionName(), req.GetVector(), int(req.GetTopK()))
		}
		if err != nil {
			c <- reply{
				Result: &edgeproto.SearchResponse{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
		}
		dist := xx.getDist(req.GetCollectionName())
		retval := make([]*edgeproto.Candidates, 0, req.GetTopK())
		for rank, nodeId := range rs.ids {
			if dist == EUCLIDEAN {
				if rs.sims[rank] > 100 {
					continue
				}
			}
			xx.Datas[req.GetCollectionName()].lock.RLock()
			clone := xx.Datas[req.GetCollectionName()].Data[uint64(nodeId)]
			xx.Datas[req.GetCollectionName()].lock.RUnlock()

			candidate := new(edgeproto.Candidates)
			candidate.Id = clone.(map[string]interface{})["_id"].(string)
			candidate.Metadata, err = structpb.NewStruct(clone.(map[string]interface{}))
			if err != nil {
				c <- reply{
					Result: &edgeproto.SearchResponse{
						Status: false,
						Error: &edgeproto.Error{
							ErrorMessage: err.Error(),
							ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
						},
					},
				}
			}
			candidate.Score = func() float32 {
				if dist == COSINE {
					return ((rs.sims[rank] + 1) / 2) * 100
				}
				return float32(math.Max(0, float64(100-rs.sims[rank])))
			}()

			retval = append(retval, candidate)
		}
		c <- reply{
			Result: &edgeproto.SearchResponse{
				Status:     true,
				Candidates: retval,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *Edge) FilterSearch(ctx context.Context, req *edgeproto.SearchReq) (
	*edgeproto.SearchResponse, error) {
	type reply struct {
		Result *edgeproto.SearchResponse
		Error  error
	}
	c := make(chan reply, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Error: fmt.Errorf(panicr, r),
				}
			}
		}()
		if !existsCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.SearchResponse{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		if !alreadyLoadCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.SearchResponse{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotLoad, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		var err error
		indexdb.indexLock.RLock()
		candidates := indexdb.indexes[req.GetCollectionName()].PureSearch(req.GetFilter())
		indexdb.indexLock.RUnlock()
		retval := make([]*edgeproto.Candidates, 0, req.GetTopK())
		for _, nodeId := range candidates {
			xx.Datas[req.GetCollectionName()].lock.RLock()
			clone := xx.Datas[req.GetCollectionName()].Data[nodeId]
			xx.Datas[req.GetCollectionName()].lock.RUnlock()

			candidate := new(edgeproto.Candidates)
			candidate.Id = clone.(map[string]interface{})["_id"].(string)
			candidate.Metadata, err = structpb.NewStruct(clone.(map[string]interface{}))
			if err != nil {
				c <- reply{
					Result: &edgeproto.SearchResponse{
						Status: false,
						Error: &edgeproto.Error{
							ErrorMessage: err.Error(),
							ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
						},
					},
				}
			}
			candidate.Score = 100
			retval = append(retval, candidate)
		}
		c <- reply{
			Result: &edgeproto.SearchResponse{
				Status:     true,
				Candidates: retval,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *Edge) HybridSearch(ctx context.Context, req *edgeproto.SearchReq) (
	*edgeproto.SearchResponse, error) {
	type reply struct {
		Result *edgeproto.SearchResponse
		Error  error
	}
	c := make(chan reply, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Error: fmt.Errorf(panicr, r),
				}
			}
		}()
		if !existsCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.SearchResponse{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		if !alreadyLoadCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &edgeproto.SearchResponse{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: fmt.Sprintf(ErrCollectionNotLoad, req.GetCollectionName()),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		// step1. find vector (user request topK * 3)
		// step2. merge bitmap with vector candidates
		// sorting conditional
		// cosine => high score is more similar
		// euclidean => low score is more similar
		// score setup
		// cosine => 100 - (score * 100)
		// euclidean => 100 - score// when score > 100 going away //(0~ infinite)
		var (
			rs  *ResultSet
			err error
		)
		if xx.getQuantization(req.GetCollectionName()) == NONE_QAUNTIZATION {
			rs, err = normalEdgeV.FullScan(req.GetCollectionName(), req.GetVector(), int(req.GetTopK())*3)
		} else {
			rs, err = quantizedEdgeV.FullScan(req.GetCollectionName(), req.GetVector(), int(req.GetTopK())*3)
		}
		if err != nil {
			c <- reply{
				Result: &edgeproto.SearchResponse{
					Status: false,
					Error: &edgeproto.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
		}
		scores := make(map[uint64]float32)
		cvU64 := make([]uint64, 0, len(rs.ids))
		for index, candidate := range rs.ids {
			scores[uint64(candidate)] = rs.sims[index]
			cvU64 = append(cvU64, uint64(candidate))
		}
		dist := xx.getDist(req.GetCollectionName())
		indexdb.indexLock.RLock()
		mergeCandidates := indexdb.indexes[req.GetCollectionName()].SearchWitCandidates(cvU64, req.GetFilter())
		indexdb.indexLock.RUnlock()
		retval := make([]*edgeproto.Candidates, 0, len(mergeCandidates))
		for _, nodeId := range mergeCandidates {
			if dist == EUCLIDEAN {
				if scores[nodeId] > 100 {
					continue
				}
			}
			xx.Datas[req.GetCollectionName()].lock.RLock()
			clone := xx.Datas[req.GetCollectionName()].Data[nodeId]
			xx.Datas[req.GetCollectionName()].lock.RUnlock()
			candidate := new(edgeproto.Candidates)
			candidate.Id = clone.(map[string]interface{})["_id"].(string)
			candidate.Metadata, err = structpb.NewStruct(clone.(map[string]interface{}))
			if err != nil {
				c <- reply{
					Result: &edgeproto.SearchResponse{
						Status: false,
						Error: &edgeproto.Error{
							ErrorMessage: err.Error(),
							ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
						},
					},
				}
			}
			candidate.Score = func() float32 {
				if dist == COSINE {
					return ((scores[nodeId] + 1) / 2) * 100
				}
				return float32(math.Max(0, float64(100-scores[nodeId])))
			}()
			retval = append(retval, candidate)
		}
		slices.SortFunc(retval, func(i, j *edgeproto.Candidates) int {
			if i.Score > j.Score {
				return -1
			} else if i.Score < j.Score {
				return 1
			}
			return 0
		})
		if len(retval) > int(req.GetTopK()) {
			retval = retval[:req.GetTopK()]
		}
		c <- reply{
			Result: &edgeproto.SearchResponse{
				Status:     true,
				Candidates: retval,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

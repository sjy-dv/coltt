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
	"slices"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/coltt/diskv"
	"github.com/sjy-dv/coltt/gen/protoc/v2/phonyproto"
	"github.com/sjy-dv/coltt/gen/protoc/v3/edgeproto"
	"github.com/sjy-dv/coltt/pkg/concurrentmap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type Edge struct {
	// Datas       map[string]*EdgeData
	Datas       *concurrentmap.Map[string, *EdgeData]
	VectorStore *Vectorstore
	lock        sync.RWMutex
	Disk        *diskv.DB
}

type EdgeData struct {
	// Data         map[uint64]interface{}
	dim          int32
	distance     string
	quantization string
	lock         sync.RWMutex
}

func NewEdge() (*Edge, error) {

	diskdb, err := diskv.Open(diskv.Options{
		DirPath:           "./data_dir/",
		SegmentSize:       1 * diskv.GB,
		Sync:              false,
		BytesPerSync:      0,
		WatchQueueSize:    0,
		AutoMergeCronExpr: "",
	})
	if err != nil {
		return nil, err
	}
	return &Edge{
		Datas:       concurrentmap.New[string, *EdgeData](),
		VectorStore: NewVectorstore(),
		Disk:        diskdb,
	}, nil
}

func (erpc *Edge) Close() {
	if err := erpc.Disk.Close(); err != nil {
		log.Error().Err(err).Msg("diskv :> It did not shut down properly ")
		return
	}
	log.Info().Msg("database shut down successfully")
}

func (erpc *Edge) getDist(collectionName string) string {
	val, ok := erpc.Datas.Get(collectionName)
	if ok {
		return val.distance
	}
	return val.distance
}

func (erpc *Edge) CreateCollection(ctx context.Context, req *edgeproto.Collection) (
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
		failFn := func(errMsg string) reply {
			return reply{
				Result: &edgeproto.CollectionResponse{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		//scripts
		if hasCollection(req.GetCollectionName()) {
			c <- failFn(fmt.Sprintf(ErrCollectionExists, req.GetCollectionName()))
			return
		}
		dist, q := protoDistQuantizationHelper(req.GetDistance(), req.GetQuantization())
		erpc.Datas.Set(req.GetCollectionName(), &EdgeData{
			dim:          int32(req.GetDim()),
			distance:     dist,
			quantization: q,
		})
		//=========vector============
		cfg := CollectionConfig{
			Dimension:      int(req.GetDim()),
			CollectionName: req.GetCollectionName(),
			Distance:       dist,
			Quantization:   q,
		}
		err := erpc.VectorStore.CreateCollection(cfg)
		if err != nil {
			erpc.Datas.Del(req.GetCollectionName())
			c <- failFn(err.Error())
			return
		}

		//bitmap
		err = indexdb.CreateIndex(req.GetCollectionName())
		if err != nil {
			erpc.Datas.Del(req.GetCollectionName())
			erpc.VectorStore.DropCollection(req.GetCollectionName())
			c <- reply{
				Result: &edgeproto.CollectionResponse{
					Status: false,
					Error:  &edgeproto.Error{ErrorMessage: err.Error(), ErrorCode: edgeproto.ErrorCode_INTERNAL_FUNC_ERROR},
				},
			}
			return
		}
		err = erpc.CommitCollection(req.GetCollectionName())
		if err != nil {
			erpc.Datas.Del(req.GetCollectionName())
			erpc.VectorStore.DropCollection(req.GetCollectionName())
			c <- failFn(err.Error())
			return
		}
		err = erpc.CommitConfig(req.GetCollectionName())
		if err != nil {
			erpc.Datas.Del(req.GetCollectionName())
			erpc.VectorStore.DropCollection(req.GetCollectionName())
			c <- failFn(err.Error())
			return
		}
		stateTrueHelper(req.GetCollectionName())
		err = erpc.saveCollection(req.GetCollectionName())
		if err != nil {
			erpc.diskClear(req.GetCollectionName())
			c <- failFn(err.Error())
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

func (erpc *Edge) DeleteCollection(ctx context.Context, req *edgeproto.CollectionName) (
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
		successFn := func() reply {
			return reply{
				Result: &edgeproto.DeleteCollectionResponse{
					Status: true,
				},
			}
		}
		if !hasCollection(req.GetCollectionName()) {
			c <- successFn()
			return
		}
		stateDestroyHelper(req.GetCollectionName())
		erpc.diskClear(req.GetCollectionName())
		c <- reply{
			Result: &edgeproto.DeleteCollectionResponse{
				Status: true,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (erpc *Edge) GetCollection(ctx context.Context, req *edgeproto.CollectionName) (
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
		failFn := func(errMsg string) reply {
			return reply{
				Result: &edgeproto.CollectionDetail{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		if !hasCollection(req.GetCollectionName()) {
			c <- failFn(fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()))
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

func (erpc *Edge) LoadCollection(ctx context.Context, req *edgeproto.CollectionName) (
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
		failFn := func(errMsg string) reply {
			return reply{
				Result: &edgeproto.CollectionDetail{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		successFn := func() reply {
			return reply{
				Result: &edgeproto.CollectionDetail{
					Status: true,
					Collection: &edgeproto.Collection{
						CollectionName: req.GetCollectionName(),
					},
				},
			}
		}
		if !hasCollection(req.GetCollectionName()) {
			c <- failFn(fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()))
			return
		}
		if alreadyLoadCollection(req.GetCollectionName()) {
			c <- successFn()
			return
		}
		loadConfig, err := erpc.LoadCommitCollectionConifg(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = erpc.LoadCommitIndex(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		merge := &EdgeData{
			dim:          int32(loadConfig.Dimension),
			distance:     loadConfig.Distance,
			quantization: loadConfig.Quantization,
		}
		erpc.Datas.Set(req.GetCollectionName(), merge)
		err = erpc.LoadData(req.GetCollectionName(), loadConfig)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		stateTrueHelper(req.GetCollectionName())
		c <- successFn()
	}()
	res := <-c
	return res.Result, res.Error
}

func (erpc *Edge) ReleaseCollection(ctx context.Context, req *edgeproto.CollectionName) (
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
		failFn := func(errMsg string) reply {
			return reply{
				Result: &edgeproto.Response{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		successFn := func() reply {
			return reply{
				Result: &edgeproto.Response{
					Status: true,
				},
			}
		}
		if !hasCollection(req.GetCollectionName()) {
			c <- failFn(fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()))
			return
		}
		if !alreadyLoadCollection(req.GetCollectionName()) {
			c <- successFn()
			return
		}

		err := erpc.CommitConfig(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = erpc.CommitIndex(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}

		erpc.memFree(req.GetCollectionName())
		stateFalseHelper(req.GetCollectionName())
		c <- successFn()
	}()
	res := <-c
	return res.Result, res.Error
}

func (erpc *Edge) Flush(ctx context.Context, req *edgeproto.CollectionName) (
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
		failFn := func(errMsg string) reply {
			return reply{
				Result: &edgeproto.Response{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		err := collectionStatusHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = erpc.CommitConfig(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = erpc.CommitIndex(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}

		err = erpc.VectorStore.Commit(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
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

func (erpc *Edge) Insert(ctx context.Context, req *edgeproto.ModifyDataset) (
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
		failFn := func(errMsg string) reply {
			return reply{
				Result: &edgeproto.Response{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		err := collectionStatusHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}

		valid := erpc.ChkValidDimensionality(req.GetCollectionName(), int32(len(req.GetVector())))
		if valid != nil {
			c <- failFn(valid.Error())
			return
		}
		autoID := autoCommitID()
		cloneMap := req.GetMetadata().AsMap()

		err = indexdb.indexes[req.GetCollectionName()].Add(autoID, cloneMap)
		if err != nil {
			c <- failFn(err.Error())
			return
		}

		err = erpc.VectorStore.InsertVector(req.GetCollectionName(), autoID, ENode{
			Vector:   req.GetVector(),
			Metadata: cloneMap,
		})
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		phonywrap := phonyproto.PhonyWrapper{
			Id:       req.GetId(),
			Vector:   req.GetVector(),
			Metadata: req.GetMetadata(),
		}
		mapping, err := proto.Marshal(&phonywrap)
		if err != nil {
			indexdb.indexes[req.GetCollectionName()].Remove(autoID, cloneMap)
			erpc.VectorStore.RemoveVector(req.GetCollectionName(), autoID)
			c <- failFn(err.Error())
			return
		}
		err = erpc.Disk.Put([]byte(fmt.Sprintf("%s_%d", req.GetCollectionName(), autoID)), mapping)
		if err != nil {
			indexdb.indexes[req.GetCollectionName()].Remove(autoID, cloneMap)
			erpc.VectorStore.RemoveVector(req.GetCollectionName(), autoID)
			c <- failFn(err.Error())
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

func (erpc *Edge) Update(ctx context.Context, req *edgeproto.ModifyDataset) (
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
		failFn := func(errMsg string) reply {
			return reply{
				Result: &edgeproto.Response{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		err := collectionStatusHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		valid := erpc.ChkValidDimensionality(req.GetCollectionName(), int32(len(req.GetVector())))
		if valid != nil {
			c <- failFn(valid.Error())
			return
		}
		getId := indexdb.indexes[req.GetCollectionName()].PureSearch(map[string]string{"_id": req.GetId()})
		if len(getId) == 0 {
			c <- reply{
				IsCreate: true,
			}
			return
		}
		phonyD, err := erpc.Disk.Get([]byte(fmt.Sprintf("%s_%d", req.GetCollectionName(), getId[0])))
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		phonydec := phonyproto.PhonyWrapper{}
		err = proto.Unmarshal(phonyD, &phonydec)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = indexdb.indexes[req.GetCollectionName()].Remove(getId[0], phonydec.GetMetadata().AsMap())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = indexdb.indexes[req.GetCollectionName()].Add(getId[0], req.GetMetadata().AsMap())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = erpc.VectorStore.UpdateVector(req.GetCollectionName(), getId[0], ENode{
			Vector:   req.GetVector(),
			Metadata: req.GetMetadata().AsMap(),
		})
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		phonywrap := phonyproto.PhonyWrapper{
			Id:       req.GetId(),
			Vector:   req.GetVector(),
			Metadata: req.GetMetadata(),
		}
		mapping, err := proto.Marshal(&phonywrap)
		if err != nil {
			erpc.failIsDelete(getId[0], req.GetCollectionName(), req.GetMetadata().AsMap())
			c <- failFn(err.Error())
			return
		}
		err = erpc.Disk.Put([]byte(fmt.Sprintf("%s_%d", req.GetCollectionName(), getId[0])), mapping)
		if err != nil {
			erpc.failIsDelete(getId[0], req.GetCollectionName(), req.GetMetadata().AsMap())
			c <- failFn(err.Error())
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
		return erpc.Insert(ctx, req)
	}
	return res.Result, res.Error
}

func (erpc *Edge) Delete(ctx context.Context, req *edgeproto.DeleteDataset) (
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
		failFn := func(errMsg string) reply {
			return reply{
				Result: &edgeproto.Response{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		err := collectionStatusHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
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

		chunkKey := []byte(fmt.Sprintf("%s_%d", req.GetCollectionName(), getId[0]))
		phonyD, err := erpc.Disk.Get(chunkKey)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = erpc.Disk.Delete(chunkKey)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		phonydec := phonyproto.PhonyWrapper{}
		err = proto.Unmarshal(phonyD, &phonydec)
		if err != nil {
			c <- failFn(err.Error())
			return
		}

		err = indexdb.indexes[req.GetCollectionName()].Remove(getId[0], phonydec.GetMetadata().AsMap())
		if err != nil {
			c <- failFn(err.Error())
			return
		}

		err = erpc.VectorStore.RemoveVector(req.GetCollectionName(), getId[0])
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = erpc.Disk.Delete([]byte(fmt.Sprintf("%s_%d", req.GetCollectionName(), getId[0])))
		if err != nil {
			c <- failFn(err.Error())
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

func (erpc *Edge) VectorSearch(ctx context.Context, req *edgeproto.SearchReq) (
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
		failFn := func(errMsg string) reply {
			return reply{
				Result: &edgeproto.SearchResponse{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		err := collectionStatusHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		valid := erpc.ChkValidDimensionality(req.GetCollectionName(), int32(len(req.GetVector())))
		if valid != nil {
			c <- failFn(valid.Error())
			return
		}
		var (
			rs []*SearchResultItem
		)

		rs, err = erpc.VectorStore.FullScan(req.GetCollectionName(), req.GetVector(), int(req.GetTopK()), req.GetHighResourceAvaliable())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		dist := erpc.getDist(req.GetCollectionName())
		retval := make([]*edgeproto.Candidates, 0, req.GetTopK())
		for _, node := range rs {
			st, err := structpb.NewStruct(node.Metadata)
			if err != nil {
				c <- failFn(err.Error())
				return
			}
			candidate := new(edgeproto.Candidates)
			// candidate.Id = node.Metadata["_id"].(string)
			candidate.Metadata = st
			candidate.Score = scoreHelper(node.Score, dist)
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

func (erpc *Edge) FilterSearch(ctx context.Context, req *edgeproto.SearchReq) (
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
		failFn := func(errMsg string) reply {
			return reply{
				Result: &edgeproto.SearchResponse{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		err := collectionStatusHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		indexdb.indexLock.RLock()
		candidates := indexdb.indexes[req.GetCollectionName()].PureSearch(req.GetFilter())
		indexdb.indexLock.RUnlock()
		retval := make([]*edgeproto.Candidates, 0, req.GetTopK())
		for _, nodeId := range candidates {
			phonyD, err := erpc.Disk.Get([]byte(fmt.Sprintf("%s_%d", req.GetCollectionName(), nodeId)))
			if err != nil {
				c <- failFn(err.Error())
				return
			}
			phonydec := phonyproto.PhonyWrapper{}
			err = proto.Unmarshal(phonyD, &phonydec)
			if err != nil {
				c <- failFn(err.Error())
				return
			}
			candidate := new(edgeproto.Candidates)
			candidate.Id = phonydec.GetId()
			candidate.Metadata = phonydec.GetMetadata()
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

func (erpc *Edge) HybridSearch(ctx context.Context, req *edgeproto.SearchReq) (
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
		failFn := func(errMsg string) reply {
			return reply{
				Result: &edgeproto.SearchResponse{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		err := collectionStatusHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		valid := erpc.ChkValidDimensionality(req.GetCollectionName(), int32(len(req.GetVector())))
		if valid != nil {
			c <- failFn(valid.Error())
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
			rs []*SearchResultItem
		)
		// if erpc.getQuantization(req.GetCollectionName()) == NONE_QAUNTIZATION {
		// 	rs, err = normalEdgeV.FullScan(req.GetCollectionName(), req.GetVector(), int(req.GetTopK())*3)
		// } else {
		// 	rs, err = quantizedEdgeV.FullScan(req.GetCollectionName(), req.GetVector(), int(req.GetTopK())*3)
		// }
		rs, err = erpc.VectorStore.FullScan(req.GetCollectionName(), req.GetVector(), int(req.GetTopK()), req.GetHighResourceAvaliable())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		scores := make(map[uint64]*SearchResultItem)
		cvU64 := make([]uint64, 0, len(rs))
		for _, candidate := range rs {
			cvU64 = append(cvU64, candidate.Id)
			scores[candidate.Id] = candidate
		}
		dist := erpc.getDist(req.GetCollectionName())
		indexdb.indexLock.RLock()
		mergeCandidates := indexdb.indexes[req.GetCollectionName()].SearchWitCandidates(cvU64, req.GetFilter())
		indexdb.indexLock.RUnlock()
		retval := make([]*edgeproto.Candidates, 0, len(mergeCandidates))
		for _, nodeId := range mergeCandidates {
			if dist == EUCLIDEAN {
				if scores[nodeId].Score > 100 {
					continue
				}
			}
			st, err := structpb.NewStruct(scores[nodeId].Metadata)
			if err != nil {
				c <- failFn(err.Error())
				return
			}
			candidate := new(edgeproto.Candidates)
			candidate.Id = scores[nodeId].Metadata["_id"].(string)
			candidate.Metadata = st
			candidate.Score = scoreHelper(scores[nodeId].Score, dist)
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

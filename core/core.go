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
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/coltt/core/vectorindex"
	"github.com/sjy-dv/coltt/diskv"
	"github.com/sjy-dv/coltt/gen/protoc/v3/coreproto"
	"github.com/sjy-dv/coltt/gen/protoc/v3/diskproto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type Core struct {
	DataStore *autoMap[*vectorindex.Hnsw]
	CommitLog *diskv.DB
}

func NewCore() (*Core, error) {
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
	return &Core{
		DataStore: NewAutoMap[*vectorindex.Hnsw](),
		CommitLog: diskdb,
	}, nil
}

func (crpc *Core) Close() {
	if err := crpc.CommitLog.Close(); err != nil {
		log.Error().Err(err).Msg("diskv :> It did not shut down properly ")
	} else {
		log.Info().Msg("database shut down successfully")
	}
	if err := crpc.exitSnapshot(); err != nil {
		log.Error().Err(err).Msg("snapshot :> each collection is saved failed.")
		return
	}
	log.Info().Msg("all collection is saved success")
}

func (crpc *Core) CreateCollection(ctx context.Context,
	req *coreproto.CollectionSpec) (*coreproto.CollectionResponse, error) {
	type reply struct {
		Result *coreproto.CollectionResponse
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
				Result: &coreproto.CollectionResponse{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		if hasCollection(req.GetCollectionName()) {
			c <- failFn(fmt.Sprintf(ErrCollectionExists, req.GetCollectionName()))
			return
		}
		distFn, distFnName := protoDistHelper(req.GetDistance())
		searchAlgo, searchOpts := protoSearchAlgoHelper(req.GetCollectionConfig().GetSearchAlgorithm())

		// save config
		diskCol := diskproto.Collection{
			CollectionName:            req.GetCollectionName(),
			LevelMultiplier:           req.CollectionConfig.GetLevelMultiplier(),
			Ef:                        req.CollectionConfig.GetEf(),
			EfConstruction:            req.CollectionConfig.GetEfConstruction(),
			M:                         req.CollectionConfig.GetM(),
			MMax:                      req.CollectionConfig.GetMMax(),
			MMax0:                     req.CollectionConfig.GetMMax0(),
			HeuristicExtendCandidates: req.CollectionConfig.GetHeuristicExtendCandidates(),
			HeuristicKeepPruned:       req.CollectionConfig.GetHeuristicKeepPruned(),
			SearchAlgorithm:           searchAlgo,
			VectorDimension:           req.GetVectorDimension(),
			Distance:                  distFnName,
			Quantization:              "None", // after update
		}

		diskBytes, err := proto.Marshal(&diskCol)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		diskFlushName := fmt.Sprintf(diskRule0, req.GetCollectionName())
		err = crpc.CommitLog.Put([]byte(diskFlushName), diskBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		hnsw := vectorindex.NewHnsw(uint(req.GetVectorDimension()),
			distFn,
			searchOpts)
		crpc.DataStore.Set(req.GetCollectionName(), hnsw)
		err = indexdb.CreateIndex(req.GetCollectionName())
		if err != nil {
			crpc.diskClear(req.GetCollectionName())
			c <- failFn(err.Error())
			return
		}
		err = crpc.saveCollection(req.GetCollectionName())
		if err != nil {
			crpc.diskClear(req.GetCollectionName())
			c <- failFn(err.Error())
			return
		}
		stateTrueHelper(req.GetCollectionName())
		c <- reply{
			Result: &coreproto.CollectionResponse{
				Status: true,
				Spec:   req,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (crpc *Core) DropCollection(ctx context.Context, req *coreproto.CollectionName) (
	*coreproto.Response, error) {
	type reply struct {
		Result *coreproto.Response
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
				Result: &coreproto.Response{
					Status: true,
				},
			}
		}
		if !hasCollection(req.GetCollectionName()) {
			c <- successFn()
			return
		}
		crpc.diskClear(req.GetCollectionName())
		crpc.removeCollection(req.GetCollectionName())
		stateDestroyHelper(req.GetCollectionName())
		c <- successFn()
	}()
	res := <-c
	return res.Result, res.Error
}

func (crpc *Core) CollectionInfof(ctx context.Context, req *coreproto.CollectionName) (
	*coreproto.CollectionMsg, error) {
	type reply struct {
		Result *coreproto.CollectionMsg
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
				Result: &coreproto.CollectionMsg{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}

		if !hasCollection(req.GetCollectionName()) {
			c <- failFn(fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()))
			return
		}
		if !alreadyLoadCollection(req.GetCollectionName()) {
			c <- failFn(fmt.Sprintf(ErrCollectionNotLoad, req.GetCollectionName()))
			return
		}
		hnsw := crpc.DataStore.Get(req.GetCollectionName())

		c <- reply{
			Result: &coreproto.CollectionMsg{
				Status: true,
				Info: &coreproto.CollectionInfo{
					CollectionName:    req.GetCollectionName(),
					CollectionConfig:  reverseConfigHelper(hnsw.Config()),
					VectorDimension:   hnsw.Dim(),
					CollectionSize:    fmt.Sprintf("%d bytes", hnsw.BytesSize()),
					CollectionLength:  uint64(hnsw.Len()),
					Distance:          reverseprotoDistHelper(hnsw.Distance()),
					CompressionHelper: coreproto.Quantization_None,
				},
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (crpc *Core) LoadCollection(ctx context.Context, req *coreproto.CollectionName) (
	*coreproto.CollectionMsg, error) {
	type reply struct {
		Result *coreproto.CollectionMsg
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
				Result: &coreproto.CollectionMsg{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		if !hasCollection(req.GetCollectionName()) {
			c <- failFn(fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()))
			return
		}
		if alreadyLoadCollection(req.GetCollectionName()) {
			hnsw := crpc.DataStore.Get(req.GetCollectionName())
			c <- reply{
				Result: &coreproto.CollectionMsg{
					Status: true,
					Info: &coreproto.CollectionInfo{
						CollectionName:    req.GetCollectionName(),
						CollectionConfig:  reverseConfigHelper(hnsw.Config()),
						VectorDimension:   hnsw.Dim(),
						CollectionSize:    fmt.Sprintf("%d bytes", hnsw.BytesSize()),
						CollectionLength:  uint64(hnsw.Len()),
						Distance:          reverseprotoDistHelper(hnsw.Distance()),
						CompressionHelper: coreproto.Quantization_None,
					},
				},
			}
			return
		}

		loadcfg, err := crpc.CommitLog.Get([]byte(fmt.Sprintf(diskRule0, req.GetCollectionName())))
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		dp := diskproto.Collection{}
		err = proto.Unmarshal(loadcfg, &dp)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = crpc.snapShotHelper(req.GetCollectionName(), dp.GetVectorDimension(),
			reversesingleprotoDistHelper(dp.GetDistance()), reverseSearchAlgoHelper(dp.GetSearchAlgorithm()))
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = indexLoadHelper(req.GetCollectionName())
		if err != nil {
			crpc.memFree(req.GetCollectionName())
			c <- failFn(err.Error())
			return
		}
		stateTrueHelper(req.GetCollectionName())
		hnsw := crpc.DataStore.Get(req.GetCollectionName())
		c <- reply{
			Result: &coreproto.CollectionMsg{
				Status: true,
				Info: &coreproto.CollectionInfo{
					CollectionName:    req.GetCollectionName(),
					CollectionConfig:  reverseConfigHelper(hnsw.Config()),
					VectorDimension:   hnsw.Dim(),
					CollectionSize:    fmt.Sprintf("%d bytes", hnsw.BytesSize()),
					CollectionLength:  uint64(hnsw.Len()),
					Distance:          reverseprotoDistHelper(hnsw.Distance()),
					CompressionHelper: coreproto.Quantization_None,
				},
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (crpc *Core) ReleaseCollection(ctx context.Context, req *coreproto.CollectionName) (
	*coreproto.ResponseWithMessage, error) {
	type reply struct {
		Result *coreproto.ResponseWithMessage
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
				Result: &coreproto.ResponseWithMessage{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}

		if !hasCollection(req.GetCollectionName()) {
			c <- failFn(fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()))
			return
		}
		if !alreadyLoadCollection(req.GetCollectionName()) {
			c <- reply{
				Result: &coreproto.ResponseWithMessage{
					Status:  true,
					Message: "collection is already release",
				},
			}
			return
		}
		err := crpc.createSnapshotHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			crpc.memFree(req.GetCollectionName())
			return
		}
		err = indexSaveHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			crpc.memFree(req.GetCollectionName())
			return
		}
		stateFalseHelper(req.GetCollectionName())
		c <- reply{
			Result: &coreproto.ResponseWithMessage{
				Status: true,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (crpc *Core) Insert(ctx context.Context, req *coreproto.DatasetChange) (
	*coreproto.Response, error) {
	type reply struct {
		Result *coreproto.Response
		Error  error
	}
	c := make(chan reply, 1)
	autoId := autoCommitID()
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
				Result: &coreproto.Response{
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
		valid := crpc.chkValidDimensionality(req.GetCollectionName(), int32(len(req.GetVector())))
		if valid != nil {
			c <- failFn(valid.Error())
			return
		}
		cloneMap := req.GetMetadata().AsMap()
		err = indexdb.indexes[req.GetCollectionName()].Add(autoId, cloneMap)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		hnsw := crpc.DataStore.Get(req.GetCollectionName())
		err = hnsw.Insert(autoId, req.GetVector(), cloneMap, hnsw.RandomLevel())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		diskkv := diskproto.Dataset{}
		diskkv.CollectionUniqueId = autoId
		diskkv.Metadata = req.GetMetadata()
		diskkv.UserSpecificId = req.GetId()
		diskkv.Vector = req.GetVector()
		diskb, err := proto.Marshal(&diskkv)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = crpc.CommitLog.Put([]byte(fmt.Sprintf(diskRule1, req.GetCollectionName(), autoId)), diskb)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		c <- reply{
			Result: &coreproto.Response{Status: true},
		}
	}()

	res := <-c
	if !res.Result.Status || res.Error != nil {
		crpc.rollbackForConsistentHelper(req.GetCollectionName(), autoId, req.GetMetadata().AsMap())
	}
	return res.Result, res.Error
}

func (crpc *Core) Update(ctx context.Context, req *coreproto.DatasetChange) (
	*coreproto.Response, error) {
	type reply struct {
		Result   *coreproto.Response
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
		failFn := func(errMsg string, create bool) reply {
			return reply{
				Result: &coreproto.Response{
					Status: false,
					Error:  errorWrap(errMsg),
				},
				IsCreate: create,
			}
		}
		err := collectionStatusHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error(), false)
			return
		}
		valid := crpc.chkValidDimensionality(req.GetCollectionName(), int32(len(req.GetVector())))
		if valid != nil {
			c <- failFn(valid.Error(), false)
			return
		}
		getId := indexdb.indexes[req.GetCollectionName()].PureSearch(map[string]string{"_id": req.GetId()})
		if len(getId) == 0 {
			c <- failFn("", true)
			return
		}
		hnsw := crpc.DataStore.Get(req.GetCollectionName())
		vertex, err := hnsw.GetVertex(getId[0])
		if err != nil {
			c <- failFn(err.Error(), false)
			return
		}
		err = indexdb.indexes[req.GetCollectionName()].Remove(getId[0], vertex.Metadata())
		if err != nil {
			c <- failFn(err.Error(), false)
			return
		}
		err = hnsw.Remove(getId[0])
		if err != nil {
			c <- failFn(err.Error(), false)
			return
		}
		err = indexdb.indexes[req.GetCollectionName()].Add(getId[0], req.GetMetadata().AsMap())
		if err != nil {
			c <- failFn(err.Error(), false)
			return
		}
		err = hnsw.Insert(getId[0], req.GetVector(), req.GetMetadata().AsMap(), hnsw.RandomLevel())
		if err != nil {
			c <- failFn(err.Error(), false)
			return
		}
		diskkv := diskproto.Dataset{}
		diskkv.CollectionUniqueId = getId[0]
		diskkv.Metadata = req.GetMetadata()
		diskkv.UserSpecificId = req.GetId()
		diskkv.Vector = req.GetVector()
		diskb, err := proto.Marshal(&diskkv)
		if err != nil {
			c <- failFn(err.Error(), false)
			return
		}
		err = crpc.CommitLog.Put([]byte(fmt.Sprintf(diskRule1, req.GetCollectionName(), getId[0])), diskb)
		if err != nil {
			c <- failFn(err.Error(), false)
			return
		}
		c <- reply{
			Result: &coreproto.Response{
				Status: true,
			},
			IsCreate: false,
		}
	}()
	res := <-c
	if res.IsCreate {
		return crpc.Insert(ctx, req)
	}
	return res.Result, res.Error
}

func (crpc *Core) Delete(ctx context.Context, req *coreproto.DatasetChange) (
	*coreproto.Response, error) {
	type reply struct {
		Result *coreproto.Response
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
				Result: &coreproto.Response{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		successFn := func() reply {
			return reply{
				Result: &coreproto.Response{
					Status: true,
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
			c <- successFn()
			return
		}
		hnsw := crpc.DataStore.Get(req.GetCollectionName())
		vertex, err := hnsw.GetVertex(getId[0])
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = indexdb.indexes[req.GetCollectionName()].Remove(getId[0], vertex.Metadata())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = hnsw.Remove(getId[0])
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = crpc.CommitLog.Delete([]byte(fmt.Sprintf(diskRule1, req.GetCollectionName(), getId[0])))
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		c <- successFn()
	}()
	res := <-c
	return res.Result, res.Error
}

func (crpc *Core) VectorSearch(ctx context.Context, req *coreproto.SearchRequest) (
	*coreproto.SearchResponse, error) {
	type reply struct {
		Result *coreproto.SearchResponse
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
				Result: &coreproto.SearchResponse{
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
		valid := crpc.chkValidDimensionality(req.GetCollectionName(), int32(len(req.GetVector())))
		if valid != nil {
			c <- failFn(valid.Error())
			return
		}
		hnsw := crpc.DataStore.Get(req.GetCollectionName())
		candidates, err := hnsw.Search(context.TODO(), req.GetVector(), uint(req.GetTopK()))
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		resultSet := make([]*coreproto.Candidates, 0, req.GetTopK())
		for _, candidate := range candidates {
			n := new(coreproto.Candidates)
			n.Id = candidate.Metadata["_id"].(string)
			n.Metadata, err = structpb.NewStruct(candidate.Metadata)
			if err != nil {
				c <- failFn(err.Error())
				return
			}
			n.Score = scoreHelper(candidate.Score, hnsw.Distance())
			resultSet = append(resultSet, n)
		}
		c <- reply{
			Result: &coreproto.SearchResponse{
				Status:     true,
				Candidates: resultSet,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (crpc *Core) FilterSearch(ctx context.Context, req *coreproto.SearchRequest) (
	*coreproto.SearchResponse, error) {
	type reply struct {
		Result *coreproto.SearchResponse
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
				Result: &coreproto.SearchResponse{
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

		candidates := indexdb.indexes[req.GetCollectionName()].PureSearch(req.GetFilter())
		resultSet := make([]*coreproto.Candidates, 0, req.GetTopK())

		for _, id := range candidates {
			data, err := crpc.CommitLog.Get([]byte(fmt.Sprintf(diskRule1, req.GetCollectionName(), id)))
			if err != nil {
				c <- failFn(err.Error())
				return
			}
			dec := diskproto.Dataset{}
			err = proto.Unmarshal(data, &dec)
			if err != nil {
				c <- failFn(err.Error())
				return
			}
			n := new(coreproto.Candidates)
			n.Id = dec.GetUserSpecificId()
			n.Metadata = dec.GetMetadata()
			n.Score = 100
			resultSet = append(resultSet, n)
		}
		c <- reply{
			Result: &coreproto.SearchResponse{
				Status:     true,
				Candidates: resultSet,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (crpc *Core) HybridSearch(ctx context.Context, req *coreproto.SearchRequest) (
	*coreproto.SearchResponse, error) {
	type reply struct {
		Result *coreproto.SearchResponse
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
				Result: &coreproto.SearchResponse{
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
		valid := crpc.chkValidDimensionality(req.GetCollectionName(), int32(len(req.GetVector())))
		if valid != nil {
			c <- failFn(valid.Error())
			return
		}
		hnsw := crpc.DataStore.Get(req.GetCollectionName())
		candidates, err := hnsw.Search(context.TODO(), req.GetVector(), uint(req.GetTopK()*3))
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		vid := make([]uint64, 0, len(candidates))
		for _, cc := range candidates {
			vid = append(vid, cc.Id)
		}
		//
		mergeCandidates := indexdb.indexes[req.GetCollectionName()].SearchWitCandidates(vid, req.GetFilter())

		// find for in for => O(n^2)
		// in map => space complexity is grow but fast O(n)
		resultSet := make([]*coreproto.Candidates, 0, req.GetTopK())
		pos := 1
		chkmap := make(map[uint64]bool)
		for _, mc := range mergeCandidates {
			chkmap[mc] = true
		}
		for _, candidate := range candidates {
			if pos >= int(req.GetTopK()) {
				break
			}
			n := new(coreproto.Candidates)
			n.Id = candidate.Metadata["_id"].(string)
			n.Metadata, err = structpb.NewStruct(candidate.Metadata)
			if err != nil {
				c <- failFn(err.Error())
				return
			}
			n.Score = scoreHelper(candidate.Score, hnsw.Distance())
			resultSet = append(resultSet, n)
			pos++
		}
		c <- reply{
			Result: &coreproto.SearchResponse{
				Status:     true,
				Candidates: resultSet,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (crpc *Core) CompareDist(ctx context.Context, req *coreproto.CompXyDist) (
	*coreproto.XyDist, error) {
	type reply struct {
		Result *coreproto.XyDist
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

		provider, distname := protoDistHelper(req.GetDist())
		score := provider.Distance(req.GetVectorX(), req.GetVectorY())
		c <- reply{
			Result: &coreproto.XyDist{
				Score: scoreHelper(score, distname),
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

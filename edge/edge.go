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

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/coltt/gen/protoc/v4/edgepb"
	"github.com/sjy-dv/coltt/pkg/minio"
	"google.golang.org/protobuf/types/known/structpb"
)

type Edge struct {
	VectorStore *Vectorstore
	Storage     *minio.MinioAPI
}

func NewEdge() (*Edge, error) {
	minioStorage, err := minio.NewMinio("localhost:9000")
	if err != nil {
		return nil, err
	}
	return &Edge{
		VectorStore: NewVectorstore(),
		Storage:     minioStorage,
	}, nil
}

func (edge *Edge) Close() {
	for col, status := range stateManager.Load.collections {
		if status {
			metaBytes, err := edge.VectorStore.SavedMetadata(col)
			if err != nil {
				log.Error().Msgf("collection: %s saved metadata failed: %s", col, err.Error())
			}
			err = edge.saveMetadataHelper(col, metaBytes)
			if err != nil {
				log.Error().Msgf("collection: %s saved metadata to minio failed: %s", col, err.Error())
			}
			vertexBytes, err := edge.VectorStore.SavedVertex(col)
			if err != nil {
				log.Error().Msgf("collection: %s saved vertex data failed: %s", col, err.Error())
			}
			err = edge.saveVertexHelper(col, vertexBytes)
			if err != nil {
				log.Error().Msgf("collection: %s saved vertex data to minio failed: %s", col, err.Error())
			}
			indexBytes, err := edge.VectorStore.SavedInverted(col)
			if err != nil {
				log.Error().Msgf("collection: %s saved vertex inverted data failed: %s", col, err.Error())
			}
			err = edge.saveInvertedIndexHelper(col, indexBytes)
			if err != nil {
				log.Error().Msgf("collection: %s saved vertex inverted data to minio failed: %s", col, err.Error())
			}
		}
	}
	log.Info().Msg("database shut down successfully")
}

func (edge *Edge) CreateCollection(ctx context.Context,
	req *edgepb.Collection) (
	*edgepb.CollectionResponse, error) {
	type reply struct {
		Result *edgepb.CollectionResponse
		Error  error
		Clear  bool
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
				Result: &edgepb.CollectionResponse{
					Status: false,
					Error:  errorWrap(errMsg),
				},
				Clear: true,
			}
		}
		if hasCollection(req.GetCollectionName()) {
			wrap := failFn(fmt.Sprintf(ErrCollectionExists, req.GetCollectionName()))
			wrap.Clear = false
			c <- wrap
			return
		}
		err := edge.Storage.CreateBucket(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		if req.GetVersioning() {
			err := edge.Storage.Versioning(req.GetCollectionName())
			if err != nil {
				c <- failFn(err.Error())
				return
			}
		}
		err = edge.VectorStore.CreateCollection(req.GetCollectionName(), Metadata{
			Dim:          req.GetDim(),
			Distance:     int32(req.GetDistance()),
			Quantization: int32(req.GetQuantization()),
			IndexType:    indexDesignAnalyze(req.GetIndex()),
			Versioning:   req.GetVersioning(),
		})
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		metaBytes, err := edge.VectorStore.SavedMetadata(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = edge.saveMetadataHelper(req.GetCollectionName(), metaBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		vertexBytes, err := edge.VectorStore.SavedVertex(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = edge.saveVertexHelper(req.GetCollectionName(), vertexBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		indexBytes, err := edge.VectorStore.SavedInverted(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = edge.saveInvertedIndexHelper(req.GetCollectionName(), indexBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		newAuthorizationBucketHelper(req.GetCollectionName())
		c <- reply{
			Result: &edgepb.CollectionResponse{
				Status: true,
				Collection: &edgepb.Collection{
					CollectionName: req.GetCollectionName(),
					Index:          req.GetIndex(),
					Distance:       req.GetDistance(),
					Quantization:   req.GetQuantization(),
					Dim:            req.GetDim(),
					Versioning:     req.GetVersioning(),
				},
			},
		}
	}()
	res := <-c
	if !res.Result.Status || res.Error != nil {
		if res.Clear {
			edge.Storage.RemoveBucket(req.GetCollectionName())
			edge.VectorStore.DestroySpace(req.GetCollectionName())
			destroyBucketHelper(req.GetCollectionName())
		}
	}
	return res.Result, res.Error
}

func (edge *Edge) DeleteCollection(ctx context.Context, req *edgepb.CollectionName) (
	*edgepb.DeleteCollectionResponse, error) {
	type reply struct {
		Result *edgepb.DeleteCollectionResponse
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
				Result: &edgepb.DeleteCollectionResponse{
					Status: true,
				},
			}
		}
		failFn := func(errMsg string) reply {
			return reply{
				Result: &edgepb.DeleteCollectionResponse{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		if !hasCollection(req.GetCollectionName()) {
			c <- successFn()
			return
		}
		destroyBucketHelper(req.GetCollectionName())

		edge.VectorStore.DestroySpace(req.GetCollectionName())
		if err := edge.Storage.RemoveBucket(req.GetCollectionName()); err != nil {
			c <- failFn(err.Error())
			return
		}
		c <- successFn()
	}()
	res := <-c
	return res.Result, res.Error
}

func (edge *Edge) GetCollection(ctx context.Context,
	req *edgepb.CollectionName) (
	*edgepb.CollectionDetail, error,
) {
	type reply struct {
		Result *edgepb.CollectionDetail
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
				Result: &edgepb.CollectionDetail{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		if !hasCollection(req.GetCollectionName()) {
			c <- failFn(fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()))
			return
		}

		dataload := false
		if err := authorization(req.GetCollectionName()); err == nil {
			dataload = true
		}
		if dataload {
			c <- reply{
				Result: &edgepb.CollectionDetail{
					Status: true,
					Collection: &edgepb.Collection{
						CollectionName: req.GetCollectionName(),
						Index:          reverseIndexDesign(edge.VectorStore.Indexer(req.GetCollectionName())),
						Distance:       edge.VectorStore.Distance(req.GetCollectionName()),
						Quantization:   edge.VectorStore.Quantization(req.GetCollectionName()),
						Dim:            edge.VectorStore.Dim(req.GetCollectionName()),
						Versioning:     edge.VectorStore.Versional(req.GetCollectionName()),
					},
					CollectionSize:   uint32(edge.VectorStore.LoadSize(req.GetCollectionName())),
					CollectionMemory: uint64(edge.VectorStore.LoadSize(req.GetCollectionName())),
					Load:             true,
				},
			}
			return
		}
		c <- reply{
			Result: &edgepb.CollectionDetail{
				Status: true,
				Collection: &edgepb.Collection{
					CollectionName: req.GetCollectionName(),
				},
				Load: false,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (edge *Edge) LoadCollection(ctx context.Context,
	req *edgepb.CollectionName) (
	*edgepb.CollectionDetail, error,
) {
	type reply struct {
		Result *edgepb.CollectionDetail
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
				Result: &edgepb.CollectionDetail{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		successFn := func() reply {
			return reply{
				Result: &edgepb.CollectionDetail{
					Status: true,
					Collection: &edgepb.Collection{
						CollectionName: req.GetCollectionName(),
						Index:          reverseIndexDesign(edge.VectorStore.Indexer(req.GetCollectionName())),
						Distance:       edge.VectorStore.Distance(req.GetCollectionName()),
						Quantization:   edge.VectorStore.Quantization(req.GetCollectionName()),
						Dim:            edge.VectorStore.Dim(req.GetCollectionName()),
						Versioning:     edge.VectorStore.Versional(req.GetCollectionName()),
					},
					CollectionSize:   uint32(edge.VectorStore.LoadSize(req.GetCollectionName())),
					CollectionMemory: uint64(edge.VectorStore.LoadSize(req.GetCollectionName())),
					Load:             true,
				},
			}
		}
		if !hasCollection(req.GetCollectionName()) {
			c <- failFn(fmt.Sprintf(ErrCollectionNotFound, req.GetCollectionName()))
			return
		}
		if err := authorization(req.GetCollectionName()); err == nil {
			c <- successFn()
			return
		}
		metadata, err := edge.loadMetadataHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}

		quantization, err := convertBytesMetadata(metadata)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		edge.VectorStore.FillEmpty(req.GetCollectionName(), quantization)

		err = edge.VectorStore.LoadedMetadata(req.GetCollectionName(), metadata)
		if err != nil {
			c <- failFn(err.Error())
			return
		}

		ivertedIndex, err := edge.loadInvertedIndexHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = edge.VectorStore.LoadedInverted(req.GetCollectionName(), ivertedIndex)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		vertexdata, err := edge.loadVertexHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = edge.VectorStore.LoadedVertex(req.GetCollectionName(), vertexdata)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		newAuthorizationBucketHelper(req.GetCollectionName())
		edge.BucketLifeCycleJob(req.GetCollectionName())
		c <- successFn()
	}()
	res := <-c
	return res.Result, res.Error
}

func (edge *Edge) ReleaseCollection(ctx context.Context,
	req *edgepb.CollectionName) (
	*edgepb.Response, error,
) {
	type reply struct {
		Result *edgepb.Response
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
				Result: &edgepb.Response{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		successFn := func() reply {
			return reply{
				Result: &edgepb.Response{
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
		//순서상 이게 먼저되어야 오래걸려도 충돌 안발생
		eliminateBucketMemoryHelper(req.GetCollectionName())
		metaBytes, err := edge.VectorStore.SavedMetadata(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = edge.saveMetadataHelper(req.GetCollectionName(), metaBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		vertexBytes, err := edge.VectorStore.SavedVertex(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = edge.saveVertexHelper(req.GetCollectionName(), vertexBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}

		indexBytes, err := edge.VectorStore.SavedInverted(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = edge.saveInvertedIndexHelper(req.GetCollectionName(), indexBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		edge.VectorStore.DestroySpace(req.GetCollectionName())
		c <- successFn()
	}()
	res := <-c
	return res.Result, res.Error
}

func (edge *Edge) Flush(ctx context.Context,
	req *edgepb.CollectionName) (
	*edgepb.Response, error,
) {
	type reply struct {
		Result *edgepb.Response
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
				Result: &edgepb.Response{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		successFn := func() reply {
			return reply{
				Result: &edgepb.Response{
					Status: true,
				},
			}
		}
		if err := authorization(req.GetCollectionName()); err != nil {
			c <- failFn(err.Error())
			return
		}

		metaBytes, err := edge.VectorStore.SavedMetadata(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = edge.saveMetadataHelper(req.GetCollectionName(), metaBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		vertexBytes, err := edge.VectorStore.SavedVertex(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = edge.saveVertexHelper(req.GetCollectionName(), vertexBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		indexBytes, err := edge.VectorStore.SavedInverted(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = edge.saveInvertedIndexHelper(req.GetCollectionName(), indexBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		c <- successFn()
	}()
	res := <-c
	return res.Result, res.Error
}

func (edge *Edge) Index(ctx context.Context, req *edgepb.IndexChange) (
	*edgepb.Response, error) {
	type reply struct {
		Result *edgepb.Response
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
				Result: &edgepb.Response{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		successFn := func() reply {
			return reply{
				Result: &edgepb.Response{
					Status: true,
				},
			}
		}
		if err := authorization(req.GetCollectionName()); err != nil {
			c <- failFn(err.Error())
			return
		}
		switch req.GetChanged() {
		case edgepb.IndexChagedType_CHANGED:
			if err := edge.VectorStore.ChangedVertex(req.GetCollectionName(), req.GetPrimaryKey(), autoCommitID(), req.GetMetadata().AsMap(), req.GetVectors()); err != nil {
				c <- failFn(err.Error())
				return
			}
		case edgepb.IndexChagedType_DELETE:
			if err := edge.VectorStore.RemoveVertex(req.GetCollectionName(), req.GetMetadata().AsMap()); err != nil {
				c <- failFn(err.Error())
				return
			}
		default:
			c <- failFn("unsupported changed type")
			return
		}
		c <- successFn()
	}()
	res := <-c
	return res.Result, res.Error
}

func (edge *Edge) Search(ctx context.Context, req *edgepb.SearchIndex) (
	*edgepb.SearchResponse, error) {
	type reply struct {
		Result *edgepb.SearchResponse
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
				Result: &edgepb.SearchResponse{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		if err := authorization(req.GetCollectionName()); err != nil {
			c <- failFn(err.Error())
			return
		}
		expr, err := queryExprAnalyzer(req.GetFilterExpression())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		items := make([]*SearchResultItem, 0)
		switch expr == nil {
		case true:
			// non-filter
			recalls, err := edge.VectorStore.VertexSearch(req.GetCollectionName(), req.GetLimit()+req.GetOffset(), req.GetVector(), req.GetHighResourceAvaliable())
			if err != nil {
				c <- failFn(err.Error())
				return
			}
			items = recalls
		case false:
			recalls, err := edge.VectorStore.FilterableVertexSearch(req.GetCollectionName(), expr, req.GetLimit()+req.GetOffset(), req.GetVector(), req.GetHighResourceAvaliable())
			if err != nil {
				c <- failFn(err.Error())
				return
			}
			items = recalls
		default:
			c <- failFn("unsupported expr type")
			return
		}
		recallRpc := make([]*edgepb.Candidates, 0, len(items))
		dist := edge.VectorStore.Distance(req.GetCollectionName())
		for _, item := range items {
			st, err := structpb.NewStruct(item.Metadata)
			if err != nil {
				c <- failFn(err.Error())
				return
			}
			candidate := new(edgepb.Candidates)
			candidate.Metadata = st
			candidate.Score = scoreHelper(item.Score, func() string {
				if dist == edgepb.Distance_Cosine {
					return T_COSINE
				}
				return EUCLIDEAN
			}())
			recallRpc = append(recallRpc, candidate)
		}
		c <- reply{
			Result: &edgepb.SearchResponse{
				Status:     true,
				Candidates: recallRpc,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

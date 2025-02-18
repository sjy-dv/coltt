package experimental

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/coltt/edge"
	"github.com/sjy-dv/coltt/gen/protoc/v3/experimentalproto"
	"github.com/sjy-dv/coltt/pkg/minio"
	"google.golang.org/protobuf/types/known/structpb"
)

type ExperimentalMultiVector struct {
	Storage     *minio.MinioAPI
	VectorStore *MultiVectorSpace
}

func NewExperimentalMultiVector() (*ExperimentalMultiVector, error) {
	minioStorage, err := minio.NewMinio("localhost:9000")
	if err != nil {
		return nil, err
	}
	return &ExperimentalMultiVector{
		Storage:     minioStorage,
		VectorStore: NewMultiVectorSpace(),
	}, nil
}

func (emv *ExperimentalMultiVector) Close() {
	for col, status := range stateManager.Load.collections {
		if status {
			metaBytes, err := emv.VectorStore.SavedMetadata(col)
			if err != nil {
				log.Error().Msgf("collection: %s saved metadata failed: %s", col, err.Error())
			}
			err = emv.saveMetadataHelper(col, metaBytes)
			if err != nil {
				log.Error().Msgf("collection: %s saved metadata to minio failed: %s", col, err.Error())
			}
			vertexBytes, err := emv.VectorStore.SavedVertex(col)
			if err != nil {
				log.Error().Msgf("collection: %s saved vertex data failed: %s", col, err.Error())
			}
			err = emv.saveVertexHelper(col, vertexBytes)
			if err != nil {
				log.Error().Msgf("collection: %s saved vertex data to minio failed: %s", col, err.Error())
			}
		}
	}
}

func (emv *ExperimentalMultiVector) CreateCollection(ctx context.Context,
	req *experimentalproto.Collection) (
	*experimentalproto.CollectionResponse, error) {
	type reply struct {
		Result *experimentalproto.CollectionResponse
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
				Result: &experimentalproto.CollectionResponse{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		if hasCollection(req.GetCollectionName()) {
			c <- failFn(fmt.Sprintf(edge.ErrCollectionExists, req.GetCollectionName()))
			return
		}
		err := emv.Storage.CreateBucket(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		if req.GetVersioning() {
			err := emv.Storage.Versioning(req.GetCollectionName())
			if err != nil {
				c <- failFn(err.Error())
				return
			}
		}
		err = emv.VectorStore.CreateCollection(req.GetCollectionName(), Metadata{
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
		metaBytes, err := emv.VectorStore.SavedMetadata(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = emv.saveMetadataHelper(req.GetCollectionName(), metaBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		vertexBytes, err := emv.VectorStore.SavedVertex(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = emv.saveVertexHelper(req.GetCollectionName(), vertexBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		newAuthorizationBucketHelper(req.GetCollectionName())
		c <- reply{
			Result: &experimentalproto.CollectionResponse{
				Status: true,
				Collection: &experimentalproto.Collection{
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
		emv.Storage.RemoveBucket(req.GetCollectionName())
		emv.VectorStore.DestroySpace(req.GetCollectionName())
		destroyBucketHelper(req.GetCollectionName())
	}
	return res.Result, res.Error
}

func (emv *ExperimentalMultiVector) DeleteCollection(ctx context.Context,
	req *experimentalproto.CollectionName) (
	*experimentalproto.DeleteCollectionResponse, error,
) {
	type reply struct {
		Result *experimentalproto.DeleteCollectionResponse
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
				Result: &experimentalproto.DeleteCollectionResponse{
					Status: true,
				},
			}
		}
		failFn := func(errMsg string) reply {
			return reply{
				Result: &experimentalproto.DeleteCollectionResponse{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		if !hasCollection(req.GetCollectionName()) {
			c <- successFn()
			return
		}
		if err := authorization(req.GetCollectionName()); err == nil {
			destroyBucketHelper(req.GetCollectionName())
		}
		emv.VectorStore.DestroySpace(req.GetCollectionName())
		if err := emv.Storage.RemoveBucket(req.GetCollectionName()); err != nil {
			c <- failFn(err.Error())
			return
		}
		c <- successFn()
	}()
	res := <-c
	return res.Result, res.Error
}

func (emv *ExperimentalMultiVector) GetCollection(ctx context.Context,
	req *experimentalproto.CollectionName) (
	*experimentalproto.CollectionDetail, error,
) {
	type reply struct {
		Result *experimentalproto.CollectionDetail
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
				Result: &experimentalproto.CollectionDetail{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		if !hasCollection(req.GetCollectionName()) {
			c <- failFn(fmt.Sprintf(edge.ErrCollectionNotFound, req.GetCollectionName()))
			return
		}

		dataload := false
		if err := authorization(req.GetCollectionName()); err == nil {
			dataload = true
		}
		if dataload {
			c <- reply{
				Result: &experimentalproto.CollectionDetail{
					Status: true,
					Collection: &experimentalproto.Collection{
						CollectionName: req.GetCollectionName(),
						Index:          reverseIndexDesign(emv.VectorStore.Indexer(req.GetCollectionName())),
						Distance:       emv.VectorStore.Distance(req.GetCollectionName()),
						Quantization:   emv.VectorStore.Quantization(req.GetCollectionName()),
						Dim:            emv.VectorStore.Dim(req.GetCollectionName()),
						Versioning:     emv.VectorStore.Versional(req.GetCollectionName()),
					},
					CollectionSize:   uint32(emv.VectorStore.LoadSize(req.GetCollectionName())),
					CollectionMemory: uint64(emv.VectorStore.LoadSize(req.GetCollectionName())),
					Load:             true,
				},
			}
			return
		}
		c <- reply{
			Result: &experimentalproto.CollectionDetail{
				Status: true,
				Collection: &experimentalproto.Collection{
					CollectionName: req.GetCollectionName(),
				},
				Load: false,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (emv *ExperimentalMultiVector) LoadCollection(ctx context.Context,
	req *experimentalproto.CollectionName) (
	*experimentalproto.CollectionDetail, error,
) {
	type reply struct {
		Result *experimentalproto.CollectionDetail
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
				Result: &experimentalproto.CollectionDetail{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		successFn := func() reply {
			return reply{
				Result: &experimentalproto.CollectionDetail{
					Status: true,
					Collection: &experimentalproto.Collection{
						CollectionName: req.GetCollectionName(),
						Index:          reverseIndexDesign(emv.VectorStore.Indexer(req.GetCollectionName())),
						Distance:       emv.VectorStore.Distance(req.GetCollectionName()),
						Quantization:   emv.VectorStore.Quantization(req.GetCollectionName()),
						Dim:            emv.VectorStore.Dim(req.GetCollectionName()),
						Versioning:     emv.VectorStore.Versional(req.GetCollectionName()),
					},
					CollectionSize:   uint32(emv.VectorStore.LoadSize(req.GetCollectionName())),
					CollectionMemory: uint64(emv.VectorStore.LoadSize(req.GetCollectionName())),
					Load:             true,
				},
			}
		}
		if !hasCollection(req.GetCollectionName()) {
			c <- failFn(fmt.Sprintf(edge.ErrCollectionNotFound, req.GetCollectionName()))
			return
		}
		if err := authorization(req.GetCollectionName()); err == nil {
			c <- successFn()
			return
		}
		emv.VectorStore.FillEmpty(req.GetCollectionName())
		metadata, err := emv.loadMetadataHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = emv.VectorStore.LoadedMetadata(req.GetCollectionName(), metadata)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		vertexdata, err := emv.loadVertexHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = emv.VectorStore.LoadedVertex(req.GetCollectionName(), vertexdata)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		newAuthorizationBucketHelper(req.GetCollectionName())
		emv.BucketLifeCycleJob(req.GetCollectionName())
		c <- successFn()
	}()
	res := <-c
	return res.Result, res.Error
}

func (emv *ExperimentalMultiVector) ReleaseCollection(ctx context.Context,
	req *experimentalproto.CollectionName) (
	*experimentalproto.Response, error,
) {
	type reply struct {
		Result *experimentalproto.Response
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
				Result: &experimentalproto.Response{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		successFn := func() reply {
			return reply{
				Result: &experimentalproto.Response{
					Status: true,
				},
			}
		}
		if !hasCollection(req.GetCollectionName()) {
			c <- failFn(fmt.Sprintf(edge.ErrCollectionNotFound, req.GetCollectionName()))
			return
		}
		if !alreadyLoadCollection(req.GetCollectionName()) {
			c <- successFn()
			return
		}
		metaBytes, err := emv.VectorStore.SavedMetadata(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = emv.saveMetadataHelper(req.GetCollectionName(), metaBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		vertexBytes, err := emv.VectorStore.SavedVertex(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = emv.saveVertexHelper(req.GetCollectionName(), vertexBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		eliminateBucketMemoryHelper(req.GetCollectionName())
		emv.VectorStore.DestroySpace(req.GetCollectionName())
		c <- successFn()
	}()
	res := <-c
	return res.Result, res.Error
}

func (emv *ExperimentalMultiVector) Flush(ctx context.Context,
	req *experimentalproto.CollectionName) (
	*experimentalproto.Response, error,
) {
	type reply struct {
		Result *experimentalproto.Response
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
				Result: &experimentalproto.Response{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		successFn := func() reply {
			return reply{
				Result: &experimentalproto.Response{
					Status: true,
				},
			}
		}
		if err := authorization(req.GetCollectionName()); err != nil {
			c <- failFn(err.Error())
			return
		}

		metaBytes, err := emv.VectorStore.SavedMetadata(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = emv.saveMetadataHelper(req.GetCollectionName(), metaBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		vertexBytes, err := emv.VectorStore.SavedVertex(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = emv.saveVertexHelper(req.GetCollectionName(), vertexBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		c <- successFn()
	}()
	res := <-c
	return res.Result, res.Error
}

func (emv *ExperimentalMultiVector) Index(ctx context.Context,
	req *experimentalproto.IndexChange) (
	*experimentalproto.Response, error,
) {
	type reply struct {
		Result *experimentalproto.Response
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
				Result: &experimentalproto.Response{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		successFn := func() reply {
			return reply{
				Result: &experimentalproto.Response{
					Status: true,
				},
			}
		}
		if err := authorization(req.GetCollectionName()); err != nil {
			c <- failFn(err.Error())
			return
		}

		switch req.GetChanged() {
		case experimentalproto.IndexChagedType_CHANGED:
			if err := metadataAnalyzer(req.GetMetadata().AsMap(), emv.VectorStore.Indexer(req.GetCollectionName())); err != nil {
				c <- failFn(err.Error())
				return
			}
			if err := emv.VectorStore.ChangedVertex(req.GetCollectionName(), req.GetId(), req.GetMetadata().AsMap(), req.GetVectors()); err != nil {
				c <- failFn(err.Error())
				return
			}
		case experimentalproto.IndexChagedType_DELETE:
			if err := emv.VectorStore.RemoveVertex(req.GetCollectionName(), req.GetId()); err != nil {
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

func (emv *ExperimentalMultiVector) VectorSearch(ctx context.Context,
	req *experimentalproto.SearchMultiIndex) (
	*experimentalproto.SearchResponse, error,
) {
	type reply struct {
		Result *experimentalproto.SearchResponse
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
				Result: &experimentalproto.SearchResponse{
					Status: false,
					Error:  errorWrap(errMsg),
				},
			}
		}
		if err := authorization(req.GetCollectionName()); err != nil {
			c <- failFn(err.Error())
			return
		}
		if err := validateRatio(req.GetVector()); err != nil {
			c <- failFn(err.Error())
			return
		}
		recalls, err := emv.VectorStore.MultiVertexSearch(req.GetCollectionName(), req.GetTopK(), req.GetVector())
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		retval := make([]*experimentalproto.Candidates, 0, len(recalls))
		for _, recall := range recalls {
			st, err := structpb.NewStruct(recall.Metadata)
			if err != nil {
				c <- failFn(err.Error())
				return
			}
			candidate := new(experimentalproto.Candidates)
			candidate.Id = recall.Id
			candidate.Metadata = st
			candidate.Score = recall.Score
			retval = append(retval, candidate)
		}
		c <- reply{
			Result: &experimentalproto.SearchResponse{
				Status:     true,
				Candidates: retval,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

package core

import (
	"context"
	"fmt"

	"github.com/sjy-dv/nnv/core/vectorindex"
	"github.com/sjy-dv/nnv/diskv"
	"github.com/sjy-dv/nnv/gen/protoc/v3/coreproto"
	"github.com/sjy-dv/nnv/gen/protoc/v3/diskproto"
	"google.golang.org/protobuf/proto"
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

func (xx *Core) CreateCollection(ctx context.Context,
	req *coreproto.CollectionSpec) (*coreproto.CollectionMsg, error) {
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
		err = xx.CommitLog.Put([]byte(diskFlushName), diskBytes)
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		hnsw := vectorindex.NewHnsw(uint(req.GetVectorDimension()),
			distFn,
			searchOpts)
		xx.DataStore.Set(req.GetCollectionName(), hnsw)
		err = indexdb.CreateIndex(req.GetCollectionName())
		if err != nil {
			xx.diskClear(req.GetCollectionName())
			c <- failFn(err.Error())
			return
		}
		stateTrueHelper(req.GetCollectionName())
		c <- reply{
			Result: &coreproto.CollectionMsg{
				Status: true,
				Info: &coreproto.CollectionInfo{
					CollectionName:    req.GetCollectionName(),
					CollectionConfig:  req.GetCollectionConfig(),
					VectorDimension:   req.GetVectorDimension(),
					Distance:          req.GetDistance(),
					CompressionHelper: req.GetCompressionHelper(),
				},
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *Core) DropCollection(ctx context.Context, req *coreproto.CollectionName) (
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
		xx.diskClear(req.GetCollectionName())
		stateDestroyHelper(req.GetCollectionName())
		c <- successFn()
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *Core) CollectionInfof(ctx context.Context, req *coreproto.CollectionName) (
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
		hnsw := xx.DataStore.Get(req.GetCollectionName())

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

func (xx *Core) LoadCollection(ctx context.Context, req *coreproto.CollectionName) (
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
			hnsw := xx.DataStore.Get(req.GetCollectionName())
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

		loadcfg, err := xx.CommitLog.Get([]byte(fmt.Sprintf(diskRule0, req.GetCollectionName())))
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
		err = xx.snapShotHelper(req.GetCollectionName(), dp.GetVectorDimension(),
			reversesingleprotoDistHelper(dp.GetDistance()), reverseSearchAlgoHelper(dp.GetSearchAlgorithm()))
		if err != nil {
			c <- failFn(err.Error())
			return
		}
		err = indexLoadHelper(req.GetCollectionName())
		if err != nil {
			xx.memFree(req.GetCollectionName())
			c <- failFn(err.Error())
			return
		}
		stateTrueHelper(req.GetCollectionName())
		hnsw := xx.DataStore.Get(req.GetCollectionName())
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

func (xx *Core) ReleaseCollection(ctx context.Context, req *coreproto.CollectionName) (
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
		err := xx.createSnapshotHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			xx.memFree(req.GetCollectionName())
			return
		}
		err = indexSaveHelper(req.GetCollectionName())
		if err != nil {
			c <- failFn(err.Error())
			xx.memFree(req.GetCollectionName())
			return
		}
		c <- reply{
			Result: &coreproto.ResponseWithMessage{
				Status: true,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *Core) Insert(ctx context.Context, req *coreproto.DatasetChange) (
	*coreproto.Response, error) {
	return nil, nil
}

func (xx *Core) Update(ctx context.Context, req *coreproto.DatasetChange) (
	*coreproto.Response, error) {
	return nil, nil
}

func (xx *Core) Delete(ctx context.Context, req *coreproto.DatasetChange) (
	*coreproto.Response, error) {
	return nil, nil
}

func (xx *Core) VectorSearch(ctx context.Context, req *coreproto.SearchRequest) (
	*coreproto.SearchResponse, error) {
	return nil, nil
}

func (xx *Core) FilterSearch(ctx context.Context, req *coreproto.SearchRequest) (
	*coreproto.SearchResponse, error) {
	return nil, nil
}

func (xx *Core) HybridSearch(ctx context.Context, req *coreproto.SearchRequest) (
	*coreproto.SearchResponse, error) {
	return nil, nil
}

func (xx *Core) CompXyDist(ctx context.Context, req *coreproto.CompXyDist) (
	*coreproto.XyDist, error) {
	return nil, nil
}

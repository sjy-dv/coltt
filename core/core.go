package core

import (
	"context"
	"fmt"

	"github.com/sjy-dv/nnv/core/vectorindex"
	"github.com/sjy-dv/nnv/diskv"
	"github.com/sjy-dv/nnv/gen/protoc/v3/coreproto"
	"github.com/sjy-dv/nnv/gen/protoc/v3/diskproto"
	"github.com/sjy-dv/nnv/pkg/concurrentmap"
	"google.golang.org/protobuf/proto"
)

type Core struct {
	DataStore *concurrentmap.Map[string, *vectorindex.Hnsw]
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
		DataStore: concurrentmap.New[string, *vectorindex.Hnsw](),
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
		stateManager.checker.cecLock.Lock()
		defer stateManager.checker.cecLock.Unlock()
		stateManager.auth.authLock.Lock()
		defer stateManager.auth.authLock.Unlock()
		stateManager.checker.collections[req.GetCollectionName()] = true
		stateManager.auth.collections[req.GetCollectionName()] = true
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
		stateManager.auth.authLock.Lock()
		defer stateManager.auth.authLock.Unlock()
		stateManager.checker.cecLock.Lock()
		defer stateManager.checker.cecLock.Unlock()
		delete(stateManager.auth.collections, req.GetCollectionName())
		delete(stateManager.checker.collections, req.GetCollectionName())

		c <- successFn()
	}()
	res := <-c
	return res.Result, res.Error
}

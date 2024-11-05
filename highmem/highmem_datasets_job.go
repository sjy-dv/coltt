package highmem

import (
	"context"
	"fmt"
	"slices"

	"github.com/sjy-dv/nnv/gen/protoc/v2/dataCoordinatorV2"
	"google.golang.org/protobuf/types/known/structpb"
)

func (xx *HighMem) Insert(
	ctx context.Context,
	req *dataCoordinatorV2.ModifyDataset,
) (*dataCoordinatorV2.Response, error) {
	stateManager.auth.authLock.RLock()
	if stateManager.auth.collections[req.GetCollectionName()] {
		stateManager.auth.authLock.RUnlock()
		goto scripts
	}
	stateManager.auth.authLock.RUnlock()
	stateManager.checker.cecLock.RLock()
	if !stateManager.checker.collections[req.GetCollectionName()] {
		stateManager.checker.cecLock.RUnlock()
		return &dataCoordinatorV2.Response{
			Status: false,
			Error: &dataCoordinatorV2.Error{
				ErrorMessage: fmt.Sprintf(notFoundCollection, req.GetCollectionName()),
				ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	stateManager.checker.cecLock.RUnlock()
	stateManager.loadchecker.clcLock.RLock()
	if !stateManager.loadchecker.collections[req.GetCollectionName()] {
		stateManager.loadchecker.clcLock.RUnlock()
		return &dataCoordinatorV2.Response{
			Status: false,
			Error: &dataCoordinatorV2.Error{
				ErrorMessage: fmt.Sprintf(notLoadCollection, req.GetCollectionName()),
				ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	stateManager.loadchecker.clcLock.RUnlock()
	stateManager.auth.authLock.Lock()
	stateManager.auth.collections[req.GetCollectionName()] = true
	stateManager.auth.authLock.Unlock()
scripts:
	type reply struct {
		Result *dataCoordinatorV2.Response
		Error  error
	}

	c := make(chan reply, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Result: nil,
					Error:  fmt.Errorf(UncaughtPanicError, r),
				}
			}
		}()

		autoId := autoCommitID()
		// first add data
		cloneMap := req.GetMetadata().AsMap()
		xx.Collections[req.GetCollectionName()].collectionLock.Lock()
		xx.Collections[req.GetCollectionName()].Data[autoId] = cloneMap
		xx.Collections[req.GetCollectionName()].collectionLock.Unlock()
		//second build index
		err := indexdb.indexes[req.GetCollectionName()].Add(autoId, cloneMap)
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV2.Response{
					Status: false,
					Error: &dataCoordinatorV2.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		//last build vector index
		err = tensorLinker.tensors[req.GetCollectionName()].Add(autoId, req.GetVector())
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV2.Response{
					Status: false,
					Error: &dataCoordinatorV2.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		c <- reply{
			Result: &dataCoordinatorV2.Response{
				Status: true,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *HighMem) Update(
	ctx context.Context,
	req *dataCoordinatorV2.ModifyDataset,
) (*dataCoordinatorV2.Response, error) {
	stateManager.auth.authLock.RLock()
	if stateManager.auth.collections[req.GetCollectionName()] {
		stateManager.auth.authLock.RUnlock()
		goto scripts
	}
	stateManager.auth.authLock.RUnlock()
	stateManager.checker.cecLock.RLock()
	if !stateManager.checker.collections[req.GetCollectionName()] {
		stateManager.checker.cecLock.RUnlock()
		return &dataCoordinatorV2.Response{
			Status: false,
			Error: &dataCoordinatorV2.Error{
				ErrorMessage: fmt.Sprintf(notFoundCollection, req.GetCollectionName()),
				ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	stateManager.checker.cecLock.RUnlock()
	stateManager.loadchecker.clcLock.RLock()
	if !stateManager.loadchecker.collections[req.GetCollectionName()] {
		stateManager.loadchecker.clcLock.RUnlock()
		return &dataCoordinatorV2.Response{
			Status: false,
			Error: &dataCoordinatorV2.Error{
				ErrorMessage: fmt.Sprintf(notLoadCollection, req.GetCollectionName()),
				ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	stateManager.loadchecker.clcLock.RUnlock()
	stateManager.auth.authLock.Lock()
	stateManager.auth.collections[req.GetCollectionName()] = true
	stateManager.auth.authLock.Unlock()
scripts:

	type reply struct {
		Result   *dataCoordinatorV2.Response
		IsCreate bool
		Error    error
	}

	c := make(chan reply, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Result: nil,
					Error:  fmt.Errorf(UncaughtPanicError, r),
				}
			}
		}()
		// find using id in bitmap index
		_id := indexdb.indexes[req.GetCollectionName()].PureSearch(map[string]string{"_id": req.GetId()})
		if len(_id) == 0 {
			// create logic
			c <- reply{
				IsCreate: true,
			}
			return
		}

		xx.Collections[req.GetCollectionName()].collectionLock.RLock()
		copyMeta := xx.Collections[req.GetCollectionName()].Data[_id[0]]
		xx.Collections[req.GetCollectionName()].collectionLock.RUnlock()
		xx.Collections[req.GetCollectionName()].collectionLock.Lock()
		xx.Collections[req.GetCollectionName()].Data[_id[0]] = req.GetMetadata().AsMap()
		xx.Collections[req.GetCollectionName()].collectionLock.Unlock()
		//remove index & new index add
		err := indexdb.indexes[req.GetCollectionName()].Remove(_id[0], copyMeta.(map[string]interface{}))
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV2.Response{
					Status: false,
					Error: &dataCoordinatorV2.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		err = indexdb.indexes[req.GetCollectionName()].Add(_id[0], req.GetMetadata().AsMap())
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV2.Response{
					Status: false,
					Error: &dataCoordinatorV2.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		err = tensorLinker.tensors[req.GetCollectionName()].Remove(_id[0])
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV2.Response{
					Status: false,
					Error: &dataCoordinatorV2.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		err = tensorLinker.tensors[req.GetCollectionName()].Add(_id[0], req.GetVector())
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV2.Response{
					Status: false,
					Error: &dataCoordinatorV2.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		c <- reply{
			Result: &dataCoordinatorV2.Response{
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

func (xx *HighMem) Delete(
	ctx context.Context,
	req *dataCoordinatorV2.DeleteDataset,
) (*dataCoordinatorV2.Response, error) {

	stateManager.auth.authLock.RLock()
	if stateManager.auth.collections[req.GetCollectionName()] {
		stateManager.auth.authLock.RUnlock()
		goto scripts
	}
	stateManager.auth.authLock.RUnlock()
	stateManager.checker.cecLock.RLock()
	if !stateManager.checker.collections[req.GetCollectionName()] {
		stateManager.checker.cecLock.RUnlock()
		return &dataCoordinatorV2.Response{
			Status: false,
			Error: &dataCoordinatorV2.Error{
				ErrorMessage: fmt.Sprintf(notFoundCollection, req.GetCollectionName()),
				ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	stateManager.checker.cecLock.RUnlock()
	stateManager.loadchecker.clcLock.RLock()
	if !stateManager.loadchecker.collections[req.GetCollectionName()] {
		stateManager.loadchecker.clcLock.RUnlock()
		return &dataCoordinatorV2.Response{
			Status: false,
			Error: &dataCoordinatorV2.Error{
				ErrorMessage: fmt.Sprintf(notLoadCollection, req.GetCollectionName()),
				ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	stateManager.loadchecker.clcLock.RUnlock()
	stateManager.auth.authLock.Lock()
	stateManager.auth.collections[req.GetCollectionName()] = true
	stateManager.auth.authLock.Unlock()
scripts:
	type reply struct {
		Result *dataCoordinatorV2.Response
		Error  error
	}
	c := make(chan reply, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Result: nil,
					Error:  fmt.Errorf(UncaughtPanicError, r),
				}
			}
		}()
		_id := indexdb.indexes[req.GetCollectionName()].PureSearch(map[string]string{"_id": req.GetId()})
		if len(_id) == 0 {
			c <- reply{
				Result: &dataCoordinatorV2.Response{
					Status: true,
				},
			}
			return
		}
		xx.Collections[req.GetCollectionName()].collectionLock.RLock()
		copyMeta := xx.Collections[req.GetCollectionName()].Data[_id[0]]
		xx.Collections[req.GetCollectionName()].collectionLock.RUnlock()
		xx.Collections[req.GetCollectionName()].collectionLock.Lock()
		delete(xx.Collections[req.GetCollectionName()].Data, _id[0])
		xx.Collections[req.GetCollectionName()].collectionLock.Unlock()
		err := indexdb.indexes[req.GetCollectionName()].Remove(_id[0], copyMeta.(map[string]interface{}))
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV2.Response{
					Status: false,
					Error: &dataCoordinatorV2.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		err = tensorLinker.tensors[req.GetCollectionName()].Remove(_id[0])
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV2.Response{
					Status: false,
					Error: &dataCoordinatorV2.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		c <- reply{
			Result: &dataCoordinatorV2.Response{
				Status: true,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *HighMem) VectorSearch(
	ctx context.Context,
	req *dataCoordinatorV2.SearchReq,
) (*dataCoordinatorV2.SearchResponse, error) {

	stateManager.auth.authLock.RLock()
	if stateManager.auth.collections[req.GetCollectionName()] {
		stateManager.auth.authLock.RUnlock()
		goto scripts
	}
	stateManager.auth.authLock.RUnlock()
	stateManager.checker.cecLock.RLock()
	if !stateManager.checker.collections[req.GetCollectionName()] {
		stateManager.checker.cecLock.RUnlock()
		return &dataCoordinatorV2.SearchResponse{
			Status: false,
			Error: &dataCoordinatorV2.Error{
				ErrorMessage: fmt.Sprintf(notFoundCollection, req.GetCollectionName()),
				ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	stateManager.checker.cecLock.RUnlock()
	stateManager.loadchecker.clcLock.RLock()
	if !stateManager.loadchecker.collections[req.GetCollectionName()] {
		stateManager.loadchecker.clcLock.RUnlock()
		return &dataCoordinatorV2.SearchResponse{
			Status: false,
			Error: &dataCoordinatorV2.Error{
				ErrorMessage: fmt.Sprintf(notLoadCollection, req.GetCollectionName()),
				ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	stateManager.loadchecker.clcLock.RUnlock()
	stateManager.auth.authLock.Lock()
	stateManager.auth.collections[req.GetCollectionName()] = true
	stateManager.auth.authLock.Unlock()
scripts:
	type reply struct {
		Result *dataCoordinatorV2.SearchResponse
		Error  error
	}

	c := make(chan reply, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Result: nil,
					Error:  fmt.Errorf(UncaughtPanicError, r),
				}
			}
		}()
		tensorLinker.tensorLock.RLock()
		candidates, distances, err := tensorLinker.tensors[req.GetCollectionName()].
			Search(req.GetVector(), uint(req.GetTopK()))
		tensorLinker.tensorLock.RUnlock()
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV2.SearchResponse{
					Status: false,
					Error: &dataCoordinatorV2.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		retval := make([]*dataCoordinatorV2.Candidates, 0, req.GetTopK())
		for rank, nodeId := range candidates {
			xx.Collections[req.GetCollectionName()].collectionLock.RLock()
			copyMeta := xx.Collections[req.GetCollectionName()].Data[nodeId]
			xx.Collections[req.GetCollectionName()].collectionLock.RUnlock()

			candidate := new(dataCoordinatorV2.Candidates)
			candidate.Id = copyMeta.(map[string]interface{})["id"].(string)
			candidate.Metadata, err = structpb.NewStruct(copyMeta.(map[string]interface{}))
			if err != nil {
				c <- reply{
					Result: &dataCoordinatorV2.SearchResponse{
						Status: false,
						Error: &dataCoordinatorV2.Error{
							ErrorMessage: err.Error(),
							ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
						},
					},
				}
				return
			}
			candidate.Score = distances[rank]
			retval = append(retval, candidate)
		}
		c <- reply{
			Result: &dataCoordinatorV2.SearchResponse{
				Status:     true,
				Candidates: retval,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *HighMem) FilterSearch(
	ctx context.Context,
	req *dataCoordinatorV2.SearchReq,
) (*dataCoordinatorV2.SearchResponse, error) {
	stateManager.auth.authLock.RLock()
	if stateManager.auth.collections[req.GetCollectionName()] {
		stateManager.auth.authLock.RUnlock()
		goto scripts
	}
	stateManager.auth.authLock.RUnlock()
	stateManager.checker.cecLock.RLock()
	if !stateManager.checker.collections[req.GetCollectionName()] {
		stateManager.checker.cecLock.RUnlock()
		return &dataCoordinatorV2.SearchResponse{
			Status: false,
			Error: &dataCoordinatorV2.Error{
				ErrorMessage: fmt.Sprintf(notFoundCollection, req.GetCollectionName()),
				ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	stateManager.checker.cecLock.RUnlock()
	stateManager.loadchecker.clcLock.RLock()
	if !stateManager.loadchecker.collections[req.GetCollectionName()] {
		stateManager.loadchecker.clcLock.RUnlock()
		return &dataCoordinatorV2.SearchResponse{
			Status: false,
			Error: &dataCoordinatorV2.Error{
				ErrorMessage: fmt.Sprintf(notLoadCollection, req.GetCollectionName()),
				ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	stateManager.loadchecker.clcLock.RUnlock()
	stateManager.auth.authLock.Lock()
	stateManager.auth.collections[req.GetCollectionName()] = true
	stateManager.auth.authLock.Unlock()
scripts:
	type reply struct {
		Result *dataCoordinatorV2.SearchResponse
		Error  error
	}

	c := make(chan reply, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Result: nil,
					Error:  fmt.Errorf(UncaughtPanicError, r),
				}
			}
		}()
		var err error
		indexdb.indexLock.RLock()
		candidates := indexdb.indexes[req.GetCollectionName()].PureSearch(req.GetFilter())
		indexdb.indexLock.RUnlock()
		retval := make([]*dataCoordinatorV2.Candidates, 0, req.GetTopK())
		for _, nodeId := range candidates {
			xx.Collections[req.GetCollectionName()].collectionLock.RLock()
			copyMeta := xx.Collections[req.GetCollectionName()].Data[nodeId]
			xx.Collections[req.GetCollectionName()].collectionLock.RUnlock()

			candidate := new(dataCoordinatorV2.Candidates)
			candidate.Id = copyMeta.(map[string]interface{})["id"].(string)
			candidate.Metadata, err = structpb.NewStruct(copyMeta.(map[string]interface{}))
			if err != nil {
				c <- reply{
					Result: &dataCoordinatorV2.SearchResponse{
						Status: false,
						Error: &dataCoordinatorV2.Error{
							ErrorMessage: err.Error(),
							ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
						},
					},
				}
				return
			}
			candidate.Score = 100
			retval = append(retval, candidate)
		}
		c <- reply{
			Result: &dataCoordinatorV2.SearchResponse{
				Status:     true,
				Candidates: retval,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *HighMem) HybridSearch(
	ctx context.Context,
	req *dataCoordinatorV2.SearchReq,
) (*dataCoordinatorV2.SearchResponse, error) {
	stateManager.auth.authLock.RLock()
	if stateManager.auth.collections[req.GetCollectionName()] {
		stateManager.auth.authLock.RUnlock()
		goto scripts
	}
	stateManager.auth.authLock.RUnlock()
	stateManager.checker.cecLock.RLock()
	if !stateManager.checker.collections[req.GetCollectionName()] {
		stateManager.checker.cecLock.RUnlock()
		return &dataCoordinatorV2.SearchResponse{
			Status: false,
			Error: &dataCoordinatorV2.Error{
				ErrorMessage: fmt.Sprintf(notFoundCollection, req.GetCollectionName()),
				ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	stateManager.checker.cecLock.RUnlock()
	stateManager.loadchecker.clcLock.RLock()
	if !stateManager.loadchecker.collections[req.GetCollectionName()] {
		stateManager.loadchecker.clcLock.RUnlock()
		return &dataCoordinatorV2.SearchResponse{
			Status: false,
			Error: &dataCoordinatorV2.Error{
				ErrorMessage: fmt.Sprintf(notLoadCollection, req.GetCollectionName()),
				ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	stateManager.loadchecker.clcLock.RUnlock()
	stateManager.auth.authLock.Lock()
	stateManager.auth.collections[req.GetCollectionName()] = true
	stateManager.auth.authLock.Unlock()
scripts:
	type reply struct {
		Result *dataCoordinatorV2.SearchResponse
		Error  error
	}

	c := make(chan reply, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Result: nil,
					Error:  fmt.Errorf(UncaughtPanicError, r),
				}
			}
		}()
		// find vector topK*3 (why? filtering in bitmap)
		tensorLinker.tensorLock.RLock()
		candidates, distances, err := tensorLinker.tensors[req.GetCollectionName()].
			Search(req.GetVector(), uint(req.GetTopK()*3))
		tensorLinker.tensorLock.RUnlock()
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV2.SearchResponse{
					Status: false,
					Error: &dataCoordinatorV2.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}

		scores := make(map[uint64]float32)
		for index, candidate := range candidates {
			scores[candidate] = distances[index]
		}
		indexdb.indexLock.RLock()
		mergeCandidates := indexdb.indexes[req.GetCollectionName()].SearchWitCandidates(candidates, req.GetFilter())
		indexdb.indexLock.RUnlock()
		retval := make([]*dataCoordinatorV2.Candidates, 0, len(mergeCandidates))
		for _, nodeId := range mergeCandidates {
			xx.Collections[req.GetCollectionName()].collectionLock.RLock()
			cloneMap := xx.Collections[req.GetCollectionName()].Data[nodeId]
			xx.Collections[req.GetCollectionName()].collectionLock.RUnlock()
			candidate := new(dataCoordinatorV2.Candidates)
			candidate.Id = cloneMap.(map[string]interface{})["id"].(string)
			candidate.Metadata, err = structpb.NewStruct(cloneMap.(map[string]interface{}))
			if err != nil {
				c <- reply{
					Result: &dataCoordinatorV2.SearchResponse{
						Status: false,
						Error: &dataCoordinatorV2.Error{
							ErrorMessage: err.Error(),
							ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
						},
					},
				}
				return
			}
			candidate.Score = scores[nodeId]
			retval = append(retval, candidate)
		}
		slices.SortFunc(retval, func(i, j *dataCoordinatorV2.Candidates) int {
			if i.Score < j.Score {
				return -1
			} else {
				return 1
			}
		})
		if len(retval) > int(req.GetTopK()) {
			retval = retval[:req.GetTopK()]
		}
		c <- reply{
			Result: &dataCoordinatorV2.SearchResponse{
				Status:     true,
				Candidates: retval,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

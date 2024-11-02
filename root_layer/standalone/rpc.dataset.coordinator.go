package standalone

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/sjy-dv/nnv/gen/protoc/v1/dataCoordinatorV1"
	"github.com/sjy-dv/nnv/pkg/hnsw"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (xx *datasetCoordinator) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (xx *datasetCoordinator) Insert(
	ctx context.Context,
	req *dataCoordinatorV1.ModifyDataset) (
	*dataCoordinatorV1.Response,
	error,
) {
	type reply struct {
		Result *dataCoordinatorV1.Response
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
		metadata := make(map[string]interface{})
		err := msgpack.Unmarshal(req.GetMetadata(), &metadata)
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV1.Response{
					Status: false,
					Error: &dataCoordinatorV1.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV1.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		nodeId, err := roots.VBucket.Insert(
			req.GetBucketName(), req.GetId(),
			req.GetVector(), metadata)
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV1.Response{
					Status: false,
					Error: &dataCoordinatorV1.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV1.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		//Logic needs to be upgraded when index addition fails
		if err := roots.BitmapIndex.Add(nodeId, metadata); err != nil {
			c <- reply{
				Result: &dataCoordinatorV1.Response{
					Status: false,
					Error: &dataCoordinatorV1.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV1.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
		}
		c <- reply{
			Result: &dataCoordinatorV1.Response{
				Status: true,
				Error:  nil,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *datasetCoordinator) Update(
	ctx context.Context,
	req *dataCoordinatorV1.ModifyDataset) (
	*dataCoordinatorV1.Response,
	error,
) {
	type reply struct {
		Result *dataCoordinatorV1.Response
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
		metadata := make(map[string]interface{})
		err := msgpack.Unmarshal(req.GetMetadata(), &metadata)
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV1.Response{
					Status: false,
					Error: &dataCoordinatorV1.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV1.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		delId, newNodeId, copyMeta, err := roots.VBucket.Update(
			req.GetBucketName(), req.GetId(),
			req.GetVector(), metadata)
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV1.Response{
					Status: false,
					Error: &dataCoordinatorV1.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV1.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}

		// bitmap delete & insert
		if err := roots.BitmapIndex.Remove(delId, copyMeta); err != nil {
			c <- reply{
				Result: &dataCoordinatorV1.Response{
					Status: false,
					Error: &dataCoordinatorV1.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV1.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		if err := roots.BitmapIndex.Add(newNodeId, metadata); err != nil {
			c <- reply{
				Result: &dataCoordinatorV1.Response{
					Status: false,
					Error: &dataCoordinatorV1.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV1.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		c <- reply{
			Result: &dataCoordinatorV1.Response{
				Status: true,
				Error:  nil,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *datasetCoordinator) Delete(
	ctx context.Context,
	req *dataCoordinatorV1.DeleteDataset) (
	*dataCoordinatorV1.Response,
	error,
) {
	type reply struct {
		Result *dataCoordinatorV1.Response
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
		nodeId, copyMeta, err := roots.VBucket.Delete(req.GetBucketName(), req.GetId())
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV1.Response{
					Status: false,
					Error: &dataCoordinatorV1.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV1.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}

		if err := roots.BitmapIndex.Remove(nodeId, copyMeta); err != nil {
			c <- reply{
				Result: &dataCoordinatorV1.Response{
					Status: false,
					Error: &dataCoordinatorV1.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV1.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		c <- reply{
			Result: &dataCoordinatorV1.Response{
				Status: true,
				Error:  nil,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

// experimental api
func (xx *datasetCoordinator) VectorSearch(
	ctx context.Context,
	req *dataCoordinatorV1.SearchReq) (
	*dataCoordinatorV1.SearchResponse,
	error,
) {
	type reply struct {
		Result *dataCoordinatorV1.SearchResponse
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
		topCandidates := hnsw.PriorityQueue{}
		if err := roots.VBucket.Search(
			req.GetBucketName(), req.GetVector(),
			&topCandidates, int(req.GetTopK()), int(req.GetEfSearch())); err != nil {
			c <- reply{
				Result: &dataCoordinatorV1.SearchResponse{
					Status: false,
					Error: &dataCoordinatorV1.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV1.ErrorCode_INTERNAL_FUNC_ERROR,
					},
					Candidates: nil,
				},
			}
			return
		}
		retval := make([]*dataCoordinatorV1.Candidates, 0, req.GetTopK())
		for _, candidate := range topCandidates.Items {
			if candidate.Node == 0 {
				continue
			}
			vmeta, err := msgpack.Marshal(roots.VBucket.Buckets[req.GetBucketName()].NodeList.Nodes[candidate.Node].Metadata)
			if err != nil {
				c <- reply{
					Result: &dataCoordinatorV1.SearchResponse{
						Status: false,
						Error: &dataCoordinatorV1.Error{
							ErrorMessage: err.Error(),
							ErrorCode:    dataCoordinatorV1.ErrorCode_INTERNAL_FUNC_ERROR,
						},
						Candidates: nil,
					},
				}
				return
			}
			retval = append(retval, &dataCoordinatorV1.Candidates{
				Id:       "",
				Metadata: vmeta,
				Score: float32(math.
					Round(float64(
						100-(candidate.Distance*100))*10) / 10),
			})
		}
		sort.Slice(retval, func(i, j int) bool {
			return retval[i].Score > retval[j].Score
		})
		c <- reply{
			Result: &dataCoordinatorV1.SearchResponse{
				Status:     true,
				Error:      nil,
				Candidates: retval,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *datasetCoordinator) FilterSearch(
	ctx context.Context,
	req *dataCoordinatorV1.SearchReq) (
	*dataCoordinatorV1.SearchResponse,
	error,
) {
	type reply struct {
		Result *dataCoordinatorV1.SearchResponse
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
		matchIds := roots.BitmapIndex.PureSearch(req.GetFilter())
		retval := make([]*dataCoordinatorV1.Candidates, 0, req.GetTopK())
		for num, id := range matchIds {
			//6 num 6 0~5 => 6
			if num >= int(req.GetTopK()) {
				break
			}
			vmeta, err := msgpack.Marshal(roots.VBucket.Buckets[req.GetBucketName()].NodeList.Nodes[id].Metadata)
			if err != nil {
				c <- reply{
					Result: &dataCoordinatorV1.SearchResponse{
						Status: false,
						Error: &dataCoordinatorV1.Error{
							ErrorMessage: err.Error(),
							ErrorCode:    dataCoordinatorV1.ErrorCode_INTERNAL_FUNC_ERROR,
						},
						Candidates: nil,
					},
				}
				return
			}
			retval = append(retval, &dataCoordinatorV1.Candidates{
				Id:       "",
				Metadata: vmeta,
				Score:    100,
			})
		}
		c <- reply{
			Result: &dataCoordinatorV1.SearchResponse{
				Status:     true,
				Error:      nil,
				Candidates: retval,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (xx *datasetCoordinator) HybridSearch(
	ctx context.Context,
	req *dataCoordinatorV1.SearchReq) (
	*dataCoordinatorV1.SearchResponse,
	error,
) {
	type reply struct {
		Result *dataCoordinatorV1.SearchResponse
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
		//Vector search searches in multiples of 3.
		topCandidates := hnsw.PriorityQueue{}
		if err := roots.VBucket.Search(
			req.GetBucketName(), req.GetVector(),
			&topCandidates, int(req.GetTopK()*3), int(req.GetEfSearch())); err != nil {
			c <- reply{
				Result: &dataCoordinatorV1.SearchResponse{
					Status: false,
					Error: &dataCoordinatorV1.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV1.ErrorCode_INTERNAL_FUNC_ERROR,
					},
					Candidates: nil,
				},
			}
			return
		}
		saveScore := make(map[uint32]float32)
		vecMatchIds := make([]uint32, 0, req.GetTopK()*3)
		for _, candidate := range topCandidates.Items {
			if candidate.Node == 0 {
				continue
			}
			vecMatchIds = append(vecMatchIds, candidate.Node)
			saveScore[candidate.Node] = candidate.Distance
		}
		finalIds := roots.BitmapIndex.SearchWitCandidates(vecMatchIds, req.GetFilter())
		retval := make([]*dataCoordinatorV1.Candidates, 0, req.GetTopK())
		for _, id := range finalIds {
			vmeta, err := msgpack.Marshal(roots.VBucket.Buckets[req.GetBucketName()].NodeList.Nodes[id].Metadata)
			if err != nil {
				c <- reply{
					Result: &dataCoordinatorV1.SearchResponse{
						Status: false,
						Error: &dataCoordinatorV1.Error{
							ErrorMessage: err.Error(),
							ErrorCode:    dataCoordinatorV1.ErrorCode_INTERNAL_FUNC_ERROR,
						},
						Candidates: nil,
					},
				}
				return
			}
			retval = append(retval, &dataCoordinatorV1.Candidates{
				Id:       "",
				Metadata: vmeta,
				Score: float32(math.
					Round(float64(
						100-(saveScore[id]*100))*10) / 10),
			})
		}
		c <- reply{
			Result: &dataCoordinatorV1.SearchResponse{
				Status:     true,
				Error:      nil,
				Candidates: retval,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

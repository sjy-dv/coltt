package rootlayer

import (
	"context"
	"fmt"
	"math"

	"github.com/sjy-dv/nnv/gen/protoc/v1/dataCoordinatorV1"
	"github.com/sjy-dv/nnv/pkg/hnsw"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (self *datasetCoordinator) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (self *datasetCoordinator) Insert(
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
		if err := roots.VBucket.Insert(
			req.GetBucketName(), req.GetId(),
			req.GetVector(), metadata); err != nil {
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
func (self *datasetCoordinator) Search(
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
			vmeta, err := msgpack.Marshal(candidate.Metadata)
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
				Id:       candidate.Metadata["_id"].(string),
				Metadata: vmeta,
				Score: float32(math.
					Round(float64(
						100-(candidate.Distance*100))*10) / 10),
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

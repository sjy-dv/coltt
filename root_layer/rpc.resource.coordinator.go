package rootlayer

import (
	"context"
	"fmt"

	"github.com/sjy-dv/nnv/gen/protoc/v1/resourceCoordinatorV1"
	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/hnsw"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (self *resourceCoordinator) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (self *resourceCoordinator) CreateBucket(
	ctx context.Context,
	req *resourceCoordinatorV1.Bucket) (
	*resourceCoordinatorV1.BucketResponse,
	error,
) {
	type reply struct {
		Result *resourceCoordinatorV1.BucketResponse
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

		var dist distance.Space
		if req.GetSpace() == resourceCoordinatorV1.Space_Cosine {
			dist = distance.NewCosine()
		} else if req.GetSpace() == resourceCoordinatorV1.Space_Manhattan {
			dist = distance.NewManhattan()
		} else {
			dist = distance.NewEuclidean()
		}

		config := hnsw.DefaultConfig(req.GetDim(), req.GetBucketName(), dist)
		if err := self.rootClone.VBucket.NewHnswBucket(req.GetBucketName(), config); err != nil {
			c <- reply{
				Result: &resourceCoordinatorV1.BucketResponse{
					Bucket: nil,
					Status: false,
					Error: &resourceCoordinatorV1.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    resourceCoordinatorV1.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
		}
		c <- reply{
			Result: &resourceCoordinatorV1.BucketResponse{
				Bucket: &resourceCoordinatorV1.Bucket{
					Efconstruction: int32(config.Efconstruction),
					M:              int32(config.M),
					Mmax:           int32(config.Mmax),
					Mmax0:          int32(config.Mmax),
					Ml:             config.Ml,
					Ep:             config.Ep,
					MaxLevel:       int32(config.MaxLevel),
					Dim:            config.Dim,
					Heuristic:      config.Heuristic,
					Space:          req.GetSpace(),
					BucketName:     config.BucketName,
					Filter:         config.Filter,
				},
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

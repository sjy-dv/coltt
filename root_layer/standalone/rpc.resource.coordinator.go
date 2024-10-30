package standalone

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
		var distType string
		if req.GetSpace() == resourceCoordinatorV1.Space_Cosine {
			dist = distance.NewCosine()
			distType = "cosine"
		} else if req.GetSpace() == resourceCoordinatorV1.Space_Manhattan {
			dist = distance.NewManhattan()
			distType = "manhattan"
		} else {
			dist = distance.NewEuclidean()
			distType = "euclidean"
		}

		config := hnsw.DefaultConfig(req.GetDim(), req.GetBucketName(), distType)
		if err := roots.VBucket.NewHnswBucket(req.GetBucketName(), config, dist); err != nil {
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
			return
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

func (self *resourceCoordinator) DeleteBucket(
	ctx context.Context,
	req *resourceCoordinatorV1.BucketName) (
	*resourceCoordinatorV1.DeleteBucketResponse,
	error,
) {
	type reply struct {
		Result *resourceCoordinatorV1.DeleteBucketResponse
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
		if ok := roots.VBucket.BucketGroup[req.GetBucketName()]; !ok {
			c <- reply{
				Result: &resourceCoordinatorV1.DeleteBucketResponse{
					Status: true,
				},
			}
			return
		}
		roots.VBucket.BucketGroup[req.GetBucketName()] = false
		for _, node := range roots.VBucket.Buckets[req.GetBucketName()].NodeList.Nodes {
			roots.VBucket.Delete(req.GetBucketName(), node.Metadata["_id"].(string))
		}
		roots.VBucket.Buckets[req.GetBucketName()] = &hnsw.Hnsw{}
		c <- reply{
			Result: &resourceCoordinatorV1.DeleteBucketResponse{
				Status: true,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (self *resourceCoordinator) GetBucket(
	ctx context.Context,
	req *resourceCoordinatorV1.BucketName) (
	*resourceCoordinatorV1.BucketDetail,
	error,
) {
	type reply struct {
		Result *resourceCoordinatorV1.BucketDetail
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
		node := roots.VBucket.Buckets[req.GetBucketName()]
		retval := &resourceCoordinatorV1.BucketDetail{}
		retval.Status = true
		retval.BucketSize = 0
		retval.BucketMemory = 0
		retval.Bucket = &resourceCoordinatorV1.Bucket{
			Efconstruction: int32(node.Efconstruction),
			M:              int32(node.M),
			Mmax:           int32(node.Mmax),
			Mmax0:          int32(node.Mmax0),
			Ml:             node.Ml,
			Ep:             node.Ep,
			MaxLevel:       int32(node.MaxLevel),
			Dim:            node.Dim,
			Heuristic:      node.Heuristic,
			Space: func() resourceCoordinatorV1.Space {
				if node.DistanceType == "cosine" {
					return resourceCoordinatorV1.Space_Cosine
				} else if node.DistanceType == "euclidean" {
					return resourceCoordinatorV1.Space_Euclidean
				} else if node.DistanceType == "manhattan" {
					return resourceCoordinatorV1.Space_Manhattan
				}
				return resourceCoordinatorV1.Space_Cosine
			}(),
			BucketName: node.BucketName,
			Filter:     node.Filter,
		}
		c <- reply{
			Result: retval,
			Error:  nil,
		}
	}()
	res := <-c
	return res.Result, res.Error
}

func (self *resourceCoordinator) GetAllBuckets(
	ctx context.Context,
	req *resourceCoordinatorV1.GetBuckets) (
	*resourceCoordinatorV1.BucketsList,
	error,
) {
	type reply struct {
		Result *resourceCoordinatorV1.BucketsList
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
		retval := &resourceCoordinatorV1.BucketsList{}
		buckets := make([]*resourceCoordinatorV1.BucketList, 0)
		for bucketName, active := range roots.VBucket.BucketGroup {
			if !active {
				continue
			}
			node := roots.VBucket.Buckets[bucketName]
			buckets = append(buckets, &resourceCoordinatorV1.BucketList{
				Bucket: &resourceCoordinatorV1.Bucket{
					Efconstruction: int32(node.Efconstruction),
					M:              int32(node.M),
					Mmax:           int32(node.Mmax),
					Mmax0:          int32(node.Mmax0),
					Ml:             node.Ml,
					Ep:             node.Ep,
					MaxLevel:       int32(node.MaxLevel),
					Dim:            node.Dim,
					Heuristic:      node.Heuristic,
					Space: func() resourceCoordinatorV1.Space {
						if node.DistanceType == "cosine" {
							return resourceCoordinatorV1.Space_Cosine
						} else if node.DistanceType == "euclidean" {
							return resourceCoordinatorV1.Space_Euclidean
						} else if node.DistanceType == "manhattan" {
							return resourceCoordinatorV1.Space_Manhattan
						}
						return resourceCoordinatorV1.Space_Cosine
					}(),
					BucketName: node.BucketName,
					Filter:     node.Filter,
				},
				BucketSize:   0,
				BucketMemory: 0,
			})
		}
		retval.Status = true
		retval.Buckets = buckets
		c <- reply{
			Result: retval,
		}
	}()
	res := <-c
	return res.Result, res.Error
}

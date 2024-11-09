package standalone

import (
	"context"

	"github.com/sjy-dv/nnv/gen/protoc/v2/resourceCoordinatorV2"
	"github.com/sjy-dv/nnv/highmem"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (xx *resourceCoordinator) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (xx *resourceCoordinator) CreateCollection(
	ctx context.Context,
	req *resourceCoordinatorV2.Collection) (
	*resourceCoordinatorV2.CollectionResponse,
	error,
) {
	err := roots.HighMem.CreateCollection(req.GetCollectionName(), highmem.CollectionConfig{
		CollectionName:  req.GetCollectionName(),
		Dim:             req.GetDim(),
		Connectivity:    req.GetConnectivity(),
		ExpansionAdd:    req.GetExpansionAdd(),
		ExpansionSearch: req.GetExpansionSearch(),
		Multi:           req.GetMulti(),
		Distance: func() string {
			switch req.GetDistance() {
			case resourceCoordinatorV2.Distance_Ip:
				return "InnerProduct"
			case resourceCoordinatorV2.Distance_L2sq:
				return "L2sq"
			case resourceCoordinatorV2.Distance_Cosine:
				return "Cosine"
			case resourceCoordinatorV2.Distance_Haversine:
				return "Haversine"
			case resourceCoordinatorV2.Distance_Divergence:
				return "Divergence"
			case resourceCoordinatorV2.Distance_Pearson:
				return "Pearson"
			case resourceCoordinatorV2.Distance_Hamming:
				return "Hamming"
			case resourceCoordinatorV2.Distance_Tanimoto:
				return "Tanimoto"
			case resourceCoordinatorV2.Distance_Sorensen:
				return "Sorensen"
			default:
				panic("Drop distance Params!")
			}
		}(),
		Quantization: func() string {
			switch req.GetQuantization() {
			case resourceCoordinatorV2.Quantization_BF16:
				return "BF16"
			case resourceCoordinatorV2.Quantization_F16:
				return "F16"
			case resourceCoordinatorV2.Quantization_F32:
				return "F32"
			case resourceCoordinatorV2.Quantization_F64:
				return "F64"
			case resourceCoordinatorV2.Quantization_I8:
				return "I8"
			case resourceCoordinatorV2.Quantization_B1:
				return "B1"
			default:
				return "None"
			}
		}(),
		Storage: func() string {
			if req.GetStorage() == resourceCoordinatorV2.StorageType_highspeed_memory {
				return "highmem"
			}
			return "not support storage"
		}(),
	})
	if err != nil {
		return &resourceCoordinatorV2.CollectionResponse{
			Status: false,
			Error: &resourceCoordinatorV2.Error{
				ErrorMessage: err.Error(),
				ErrorCode:    resourceCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	return &resourceCoordinatorV2.CollectionResponse{
		Status: true,
		Collection: &resourceCoordinatorV2.Collection{
			CollectionName:  req.GetCollectionName(),
			Distance:        req.GetDistance(),
			Quantization:    req.GetQuantization(),
			Dim:             req.GetDim(),
			Connectivity:    req.GetConnectivity(),
			ExpansionAdd:    req.GetExpansionAdd(),
			ExpansionSearch: req.GetExpansionSearch(),
			Multi:           req.GetMulti(),
			Storage:         req.GetStorage(),
		},
	}, nil
}

func (xx *resourceCoordinator) DeleteCollection(
	ctx context.Context,
	req *resourceCoordinatorV2.CollectionName,
) (
	*resourceCoordinatorV2.DeleteCollectionResponse,
	error,
) {
	err := roots.HighMem.DropCollection(req.GetCollectionName())
	if err != nil {
		return &resourceCoordinatorV2.DeleteCollectionResponse{
			Status: false,
			Error: &resourceCoordinatorV2.Error{
				ErrorMessage: err.Error(),
				ErrorCode:    resourceCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	return &resourceCoordinatorV2.DeleteCollectionResponse{
		Status: true,
	}, nil
}

// after release
func (xx *resourceCoordinator) GetCollection(
	ctx context.Context,
	req *resourceCoordinatorV2.CollectionName,
) (
	*resourceCoordinatorV2.CollectionDetail,
	error,
) {
	config, err := roots.HighMem.GetCollection(req.GetCollectionName())
	if err != nil {
		return &resourceCoordinatorV2.CollectionDetail{
			Status: false,
			Error: &resourceCoordinatorV2.Error{
				ErrorMessage: err.Error(),
				ErrorCode:    resourceCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	return &resourceCoordinatorV2.CollectionDetail{
		Status: true,
		Collection: &resourceCoordinatorV2.Collection{
			CollectionName: config.CollectionName,
		},
		CollectionSize:   0,
		CollectionMemory: 0,
	}, nil
}

func (xx *resourceCoordinator) GetAllCollections(
	ctx context.Context,
	req *resourceCoordinatorV2.GetCollections,
) (
	*resourceCoordinatorV2.CollectionLists,
	error,
) {
	panic("unimplement yet")
}

// after release
func (xx *resourceCoordinator) LoadCollection(
	ctx context.Context,
	req *resourceCoordinatorV2.CollectionName,
) (
	*resourceCoordinatorV2.CollectionDetail,
	error,
) {
	cinf, err := roots.HighMem.LoadCollection(req.GetCollectionName())
	if err != nil {
		return &resourceCoordinatorV2.CollectionDetail{
			Status: false,
			Error: &resourceCoordinatorV2.Error{
				ErrorMessage: err.Error(),
				ErrorCode:    resourceCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	return &resourceCoordinatorV2.CollectionDetail{
		Status: true,
		Collection: &resourceCoordinatorV2.Collection{
			CollectionName: cinf.CollectionName,
		},
		CollectionSize: uint32(cinf.DataSize),
	}, nil
}

func (xx *resourceCoordinator) ReleaseCollection(
	ctx context.Context,
	req *resourceCoordinatorV2.CollectionName,
) (
	*resourceCoordinatorV2.Response,
	error,
) {
	err := roots.HighMem.ReleaseCollection(req.GetCollectionName())
	if err != nil {
		return &resourceCoordinatorV2.Response{
			Status: false,
			Error: &resourceCoordinatorV2.Error{
				ErrorMessage: err.Error(),
				ErrorCode:    resourceCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
			},
		}, nil
	}
	return &resourceCoordinatorV2.Response{
		Status: true,
	}, nil
}

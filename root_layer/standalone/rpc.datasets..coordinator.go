package standalone

import (
	"context"

	"github.com/sjy-dv/nnv/gen/protoc/v2/dataCoordinatorV2"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (xx *datasetCoordinator) Ping(ctx context.Context,
	req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (xx *datasetCoordinator) Insert(
	ctx context.Context,
	req *dataCoordinatorV2.ModifyDataset,
) (
	*dataCoordinatorV2.Response,
	error,
) {
	return roots.HighMem.Insert(ctx, req)
}

func (xx *datasetCoordinator) Update(
	ctx context.Context,
	req *dataCoordinatorV2.ModifyDataset,
) (
	*dataCoordinatorV2.Response,
	error,
) {
	return roots.HighMem.Update(ctx, req)
}

func (xx *datasetCoordinator) Delete(
	ctx context.Context,
	req *dataCoordinatorV2.DeleteDataset,
) (*dataCoordinatorV2.Response, error) {
	return roots.HighMem.Delete(ctx, req)
}

func (xx *datasetCoordinator) VectorSearch(
	ctx context.Context,
	req *dataCoordinatorV2.SearchReq,
) (*dataCoordinatorV2.SearchResponse, error) {
	return roots.HighMem.VectorSearch(ctx, req)
}

func (xx *datasetCoordinator) FilterSearch(
	ctx context.Context,
	req *dataCoordinatorV2.SearchReq,
) (*dataCoordinatorV2.SearchResponse, error) {
	return roots.HighMem.FilterSearch(ctx, req)
}

func (xx *datasetCoordinator) HybridSearch(
	ctx context.Context,
	req *dataCoordinatorV2.SearchReq,
) (*dataCoordinatorV2.SearchResponse, error) {
	return roots.HighMem.HybridSearch(ctx, req)
}

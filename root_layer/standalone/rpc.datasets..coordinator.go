package standalone

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (xx *datasetCoordinator) Ping(ctx context.Context,
	req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

package edgelite

import (
	"context"

	"github.com/sjy-dv/coltt/gen/protoc/v4/edgepb"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (*edgeProtoConn) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (*edgeProtoConn) CreateCollection(ctx context.Context, req *edgepb.Collection) (
	*edgepb.CollectionResponse, error) {
	return edgelites.Edge.CreateCollection(ctx, req)
}

func (*edgeProtoConn) DeleteCollection(ctx context.Context, req *edgepb.CollectionName) (
	*edgepb.DeleteCollectionResponse, error) {
	return edgelites.Edge.DeleteCollection(ctx, req)
}

func (*edgeProtoConn) GetCollection(ctx context.Context, req *edgepb.CollectionName) (
	*edgepb.CollectionDetail, error) {
	return edgelites.Edge.GetCollection(ctx, req)
}

func (*edgeProtoConn) LoadCollection(ctx context.Context, req *edgepb.CollectionName) (
	*edgepb.CollectionDetail, error) {
	return edgelites.Edge.LoadCollection(ctx, req)
}

func (*edgeProtoConn) ReleaseCollection(ctx context.Context, req *edgepb.CollectionName) (
	*edgepb.Response, error) {
	return edgelites.Edge.ReleaseCollection(ctx, req)
}

func (*edgeProtoConn) Flush(ctx context.Context, req *edgepb.CollectionName) (
	*edgepb.Response, error) {
	return edgelites.Edge.Flush(ctx, req)
}

func (*edgeProtoConn) Index(ctx context.Context, req *edgepb.IndexChange) (
	*edgepb.Response, error) {
	return edgelites.Edge.Index(ctx, req)
}

func (*edgeProtoConn) Search(ctx context.Context, req *edgepb.SearchIndex) (
	*edgepb.SearchResponse, error) {
	return edgelites.Edge.Search(ctx, req)
}

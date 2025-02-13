package root

import (
	"context"

	"github.com/sjy-dv/coltt/gen/protoc/v3/coreproto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (xx *coreProtoConn) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (xx *coreProtoConn) CreateCollection(ctx context.Context, req *coreproto.CollectionSpec) (
	*coreproto.CollectionResponse, error) {
	return rc.Core.CreateCollection(ctx, req)
}

func (xx *coreProtoConn) DropCollection(ctx context.Context, req *coreproto.CollectionName) (
	*coreproto.Response, error) {
	return rc.Core.DropCollection(ctx, req)
}

func (xx *coreProtoConn) CollectionInfof(ctx context.Context, req *coreproto.CollectionName) (
	*coreproto.CollectionMsg, error) {
	return rc.Core.CollectionInfof(ctx, req)
}

func (xx *coreProtoConn) LoadCollection(ctx context.Context, req *coreproto.CollectionName) (
	*coreproto.CollectionMsg, error) {
	return rc.Core.LoadCollection(ctx, req)
}

func (xx *coreProtoConn) ReleaseCollection(ctx context.Context, req *coreproto.CollectionName) (
	*coreproto.ResponseWithMessage, error) {
	return rc.Core.ReleaseCollection(ctx, req)
}

func (xx *coreProtoConn) Insert(ctx context.Context, req *coreproto.DatasetChange) (
	*coreproto.Response, error) {
	return rc.Core.Insert(ctx, req)
}

func (xx *coreProtoConn) Update(ctx context.Context, req *coreproto.DatasetChange) (
	*coreproto.Response, error) {
	return rc.Core.Update(ctx, req)
}

func (xx *coreProtoConn) Delete(ctx context.Context, req *coreproto.DatasetChange) (
	*coreproto.Response, error) {
	return rc.Core.Delete(ctx, req)
}

func (xx *coreProtoConn) VectorSearch(ctx context.Context, req *coreproto.SearchRequest) (
	*coreproto.SearchResponse, error) {
	return rc.Core.VectorSearch(ctx, req)
}

func (xx *coreProtoConn) FilterSearch(ctx context.Context, req *coreproto.SearchRequest) (
	*coreproto.SearchResponse, error) {
	return rc.Core.FilterSearch(ctx, req)
}

func (xx *coreProtoConn) HybridSearch(ctx context.Context, req *coreproto.SearchRequest) (
	*coreproto.SearchResponse, error) {
	return rc.Core.HybridSearch(ctx, req)
}

func (xx *coreProtoConn) CompareDist(ctx context.Context, req *coreproto.CompXyDist) (
	*coreproto.XyDist, error) {
	return rc.Core.CompareDist(ctx, req)
}

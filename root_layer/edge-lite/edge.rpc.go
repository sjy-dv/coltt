package edgelite

import (
	"context"

	"github.com/sjy-dv/coltt/gen/protoc/v3/edgeproto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (xx *edgeProtoConn) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (xx *edgeProtoConn) CreateCollection(ctx context.Context, req *edgeproto.Collection) (
	*edgeproto.CollectionResponse, error) {
	return edgelites.Edge.CreateCollection(ctx, req)
}

func (xx *edgeProtoConn) DeleteCollection(ctx context.Context, req *edgeproto.CollectionName) (
	*edgeproto.DeleteCollectionResponse, error) {
	return edgelites.Edge.DeleteCollection(ctx, req)
}

func (xx *edgeProtoConn) GetCollection(ctx context.Context, req *edgeproto.CollectionName) (
	*edgeproto.CollectionDetail, error) {
	return edgelites.Edge.GetCollection(ctx, req)
}

func (xx *edgeProtoConn) LoadCollection(ctx context.Context, req *edgeproto.CollectionName) (
	*edgeproto.CollectionDetail, error) {
	return edgelites.Edge.LoadCollection(ctx, req)
}

func (xx *edgeProtoConn) ReleaseCollection(ctx context.Context, req *edgeproto.CollectionName) (
	*edgeproto.Response, error) {
	return edgelites.Edge.ReleaseCollection(ctx, req)
}

func (xx *edgeProtoConn) Flush(ctx context.Context, req *edgeproto.CollectionName) (
	*edgeproto.Response, error) {
	return edgelites.Edge.Flush(ctx, req)
}

func (xx *edgeProtoConn) Insert(ctx context.Context, req *edgeproto.ModifyDataset) (
	*edgeproto.Response, error) {
	return edgelites.Edge.Insert(ctx, req)
}

func (xx *edgeProtoConn) Update(ctx context.Context, req *edgeproto.ModifyDataset) (
	*edgeproto.Response, error) {
	return edgelites.Edge.Update(ctx, req)
}

func (xx *edgeProtoConn) Delete(ctx context.Context, req *edgeproto.DeleteDataset) (
	*edgeproto.Response, error) {
	return edgelites.Edge.Delete(ctx, req)
}

func (xx *edgeProtoConn) VectorSearch(ctx context.Context, req *edgeproto.SearchReq) (
	*edgeproto.SearchResponse, error) {
	return edgelites.Edge.VectorSearch(ctx, req)
}

func (xx *edgeProtoConn) FilterSearch(ctx context.Context, req *edgeproto.SearchReq) (
	*edgeproto.SearchResponse, error) {
	return edgelites.Edge.FilterSearch(ctx, req)
}

func (xx *edgeProtoConn) HybridSearch(ctx context.Context, req *edgeproto.SearchReq) (
	*edgeproto.SearchResponse, error) {
	return edgelites.Edge.HybridSearch(ctx, req)
}

package experimentalLayer

import (
	"context"

	"github.com/sjy-dv/coltt/gen/protoc/v3/experimentalproto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (protoconn *ProtoConn) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (protoconn *ProtoConn) CreateCollection(ctx context.Context, req *experimentalproto.Collection) (
	*experimentalproto.CollectionResponse, error) {
	return administrator.Engine.CreateCollection(ctx, req)
}

func (protoconn *ProtoConn) DeleteCollection(ctx context.Context, req *experimentalproto.CollectionName) (
	*experimentalproto.DeleteCollectionResponse, error) {
	return administrator.Engine.DeleteCollection(ctx, req)
}

func (protoconn *ProtoConn) GetCollection(ctx context.Context, req *experimentalproto.CollectionName) (
	*experimentalproto.CollectionDetail, error) {
	return administrator.Engine.GetCollection(ctx, req)
}

func (protoconn *ProtoConn) LoadCollection(ctx context.Context, req *experimentalproto.CollectionName) (
	*experimentalproto.CollectionDetail, error) {
	return administrator.Engine.LoadCollection(ctx, req)
}

func (protoconn *ProtoConn) Flush(ctx context.Context, req *experimentalproto.CollectionName) (*experimentalproto.Response, error) {
	return administrator.Engine.Flush(ctx, req)
}

// Index implements experimentalproto.ExperimentalMultiVectorRpcServer.
func (protoconn *ProtoConn) Index(ctx context.Context, req *experimentalproto.IndexChange) (*experimentalproto.Response, error) {
	return administrator.Engine.Index(ctx, req)
}

// ReleaseCollection implements experimentalproto.ExperimentalMultiVectorRpcServer.
func (protoconn *ProtoConn) ReleaseCollection(ctx context.Context, req *experimentalproto.CollectionName) (*experimentalproto.Response, error) {
	return administrator.Engine.ReleaseCollection(ctx, req)
}

// VectorSearch implements experimentalproto.ExperimentalMultiVectorRpcServer.
func (protoconn *ProtoConn) VectorSearch(ctx context.Context, req *experimentalproto.SearchMultiIndex) (*experimentalproto.SearchResponse, error) {
	return administrator.Engine.VectorSearch(ctx, req)
}

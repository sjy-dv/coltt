// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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

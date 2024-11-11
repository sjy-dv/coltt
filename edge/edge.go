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

package edge

import (
	"context"
	"sync"

	"github.com/sjy-dv/nnv/gen/protoc/v2/edgeproto"
)

type Edge struct {
	Datas map[string]*EdgeData
	lock  sync.RWMutex
}

type EdgeData struct {
	Data         map[uint64]interface{}
	dim          int32
	distance     string
	quantization string
	lock         sync.RWMutex
}

func NewEdge() *Edge {
	return &Edge{
		Datas: make(map[string]*EdgeData),
	}
}

func (xx *Edge) CreateCollection(ctx context.Context, req *edgeproto.Collection) (
	*edgeproto.CollectionResponse, error) {

}

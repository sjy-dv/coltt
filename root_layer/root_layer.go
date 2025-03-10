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

package rootlayer

import (
	"context"

	edgelite "github.com/sjy-dv/coltt/root_layer/edge-lite"
	"github.com/sjy-dv/coltt/root_layer/experimentalLayer"
	"github.com/sjy-dv/coltt/root_layer/root"
)

// after code refactoring

func NewRootLayer(mode string) error {
	if mode == "edge" {
		return edgelite.NewEdgeLite()
	} else if mode == "experimental" {
		return experimentalLayer.NewExperimentalLayer()
	}
	return root.NewRoot()
}

func StableRelease(ctx context.Context, mode string) error {
	if mode == "edge" {
		return edgelite.StableRelease(ctx)
	} else if mode == "experimental" {
		return experimentalLayer.StableRelease(ctx)
	}
	return root.StableRelease(ctx)
}

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

package data_access_layer

import (
	"encoding/gob"

	"github.com/RoaringBitmap/roaring"
	"github.com/google/uuid"
	"github.com/sjy-dv/nnv/pkg/hnsw"
)

func init() {
	//prevent `gob: type not registered for interface: uuid.UUID` Error
	gob.Register(uuid.UUID{})
	gob.Register(hnsw.Node{})
	gob.Register([]hnsw.Node{})
	gob.Register(hnsw.HnswConfig{})
	gob.Register(map[string]interface{}{})
	gob.Register(map[string][]hnsw.Node{})
	gob.Register(map[string]hnsw.HnswConfig{})
	gob.Register(roaring.Bitmap{})
	gob.Register(map[string]map[string]*roaring.Bitmap{})
}

type BackupHnswBucket struct {
	DataNodes    []hnsw.Node
	BucketConfig hnsw.HnswConfig
	BucketName   string
}

type BackupNodes map[string][]hnsw.Node
type BackupConfig map[string]hnsw.HnswConfig
type BackupBucketList []string

type SerdeBitmap struct {
	Data string `json:"data"`
}

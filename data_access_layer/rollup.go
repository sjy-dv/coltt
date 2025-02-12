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
	"compress/flate"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/coltt/pkg/distance"
	"github.com/sjy-dv/coltt/pkg/hnsw"
	"github.com/sjy-dv/coltt/pkg/index"
)

// Recovery logic needs to be added in the future.
func Rollup() (*hnsw.HnswBucket, *index.BitmapIndex, error) {
	dirs, err := os.ReadDir(parentDir)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_ReadDir failed")
		return nil, nil, err
	}
	// find recent dirs
	var dirsList []string
	if len(dirs) == 0 {
		return nil, nil, nil
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			if dir.Name() == "matchid" || dir.Name() == "backup-log" {
				continue
			}
			dirsList = append(dirsList, dir.Name())
		}
	}
	if len(dirsList) == 0 {
		return nil, nil, nil
	}
	sort.Strings(dirsList)
	rollUpDirs := filepath.Join(parentDir, dirsList[len(dirsList)-1])
	fmt.Println("rollup", rollUpDirs)
	// =======config decode key===========
	// decodeKey, err := base64.StdEncoding.DecodeString(config.Config.CacheKey)
	// if err != nil {
	// 	log.Warn().Err(err).Msg("data_access_layer.Rollup_base64_decode_string failed")
	// 	return nil, err
	// }
	// ----------load nodes data---------------

	nodesCdat, err := os.OpenFile(fmt.Sprintf("%s/nodes.cdat", rollUpDirs), os.O_RDONLY, 0777)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_nodesCdat_OpenFile failed")
		return nil, nil, err
	}

	loadNodes := make(map[string][]hnsw.Node)

	var ncIo io.Reader

	ncIo = flate.NewReader(nodesCdat)

	nodeDec := gob.NewDecoder(ncIo)
	err = nodeDec.Decode(&loadNodes)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_NodeCdat Decode failed")
		return nil, nil, err
	}
	err = ncIo.(io.Closer).Close()
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_Abnormaly nodesCdat Io Closing failed")
		return nil, nil, err
	}

	//============load nodes config ==================
	loadCofigs := make(map[string]hnsw.HnswConfig)

	nodesConf, err := os.OpenFile(fmt.Sprintf("%s/nodes_config.cdat", rollUpDirs), os.O_RDONLY, 0777)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_nodesConfig_OpenFile failed")
		return nil, nil, err
	}

	var ncfgIo io.Reader

	ncfgIo = flate.NewReader(nodesConf)
	nodeCfgDec := gob.NewDecoder(ncfgIo)
	err = nodeCfgDec.Decode(&loadCofigs)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_NodesConfig Decode failed")
		return nil, nil, err
	}
	err = ncfgIo.(io.Closer).Close()
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_Abnormaly config Io Closing failed")
		return nil, nil, err
	}
	//=========load buckets ===================
	loadBuckets := make([]string, 0)

	nodesBucket, err := os.OpenFile(fmt.Sprintf("%s/buckets.cdat", rollUpDirs), os.O_RDONLY, 0777)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_Bucket_OpenFile failed")
		return nil, nil, err
	}

	var bucketIo io.Reader

	bucketIo = flate.NewReader(nodesBucket)
	bucketDec := gob.NewDecoder(bucketIo)
	err = bucketDec.Decode(&loadBuckets)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_Bucket Decode failed")
		return nil, nil, err
	}
	err = bucketIo.(io.Closer).Close()
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_Abnormaly Bucket Io Closing failed")
		return nil, nil, err
	}
	recoveryBuckets := &hnsw.HnswBucket{
		Buckets:     make(map[string]*hnsw.Hnsw),
		BucketGroup: make(map[string]bool),
	}
	for _, bucket := range loadBuckets {
		recoveryBuckets.Buckets[bucket] = new(hnsw.Hnsw)
		recoveryBuckets.BucketGroup[bucket] = true
		recoveryBuckets.Buckets[bucket].BucketName = bucket
		recoveryBuckets.Buckets[bucket].Efconstruction = loadCofigs[bucket].Efconstruction
		recoveryBuckets.Buckets[bucket].M = loadCofigs[bucket].M
		recoveryBuckets.Buckets[bucket].Mmax = loadCofigs[bucket].Mmax
		recoveryBuckets.Buckets[bucket].Mmax0 = loadCofigs[bucket].Mmax0
		recoveryBuckets.Buckets[bucket].Ml = loadCofigs[bucket].Ml
		recoveryBuckets.Buckets[bucket].Ep = loadCofigs[bucket].Ep
		recoveryBuckets.Buckets[bucket].MaxLevel = loadCofigs[bucket].MaxLevel
		recoveryBuckets.Buckets[bucket].Dim = loadCofigs[bucket].Dim
		recoveryBuckets.Buckets[bucket].Heuristic = loadCofigs[bucket].Heuristic
		recoveryBuckets.Buckets[bucket].DistanceType = loadCofigs[bucket].DistanceType
		recoveryBuckets.Buckets[bucket].Space = func() distance.Space {
			switch loadCofigs[bucket].DistanceType {
			case "cosine":
				return distance.NewCosine()
			case "euclidean":
				return distance.NewEuclidean()
			case "manhattan":
				return distance.NewManhattan()
			default:
				log.Warn().Err(err).Msg("data_access_layer.Rollup_recovery distance failed using cosine setup")
				return distance.NewCosine()
			}
		}()
		// index add

		recoveryBuckets.Buckets[bucket].EmptyNodes = loadCofigs[bucket].EmptyNodes
		recoveryBuckets.Buckets[bucket].NodeList.Nodes = loadNodes[bucket]
	}

	recoveryIndex := index.NewBitmapIndex()
	err = recoveryIndex.DeserializeBinary(fmt.Sprintf("%s/index.bin", rollUpDirs))
	if err != nil {
		log.Warn().Err(err).Msg("bitmap index recovery failed")
		return nil, nil, err
	}
	err = recoveryIndex.ValidateIndex()
	if err != nil {
		log.Warn().Err(err).Msg("bitmap index validation test failed")
		return nil, nil, err
	}
	return recoveryBuckets, recoveryIndex, nil
}

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
	"github.com/sjy-dv/nnv/pkg/distance"
	"github.com/sjy-dv/nnv/pkg/hnsw"
)

// Recovery logic needs to be added in the future.
func Rollup() (*hnsw.HnswBucket, error) {
	dirs, err := os.ReadDir(parentDir)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_ReadDir failed")
		return nil, err
	}
	// find recent dirs
	var dirsList []string
	if len(dirs) == 0 {
		return nil, nil
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			if dir.Name() == "matchid" || dir.Name() == "backup" {
				continue
			}
			dirsList = append(dirsList, dir.Name())
		}
	}
	if len(dirsList) == 0 {
		return nil, nil
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
		return nil, err
	}

	loadNodes := make(map[string][]hnsw.Node)

	var ncIo io.Reader

	ncIo = flate.NewReader(nodesCdat)

	nodeDec := gob.NewDecoder(ncIo)
	err = nodeDec.Decode(&loadNodes)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_NodeCdat Decode failed")
		return nil, err
	}
	err = ncIo.(io.Closer).Close()
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_Abnormaly nodesCdat Io Closing failed")
		return nil, err
	}

	//============load nodes config ==================
	loadCofigs := make(map[string]hnsw.HnswConfig)

	nodesConf, err := os.OpenFile(fmt.Sprintf("%s/nodes_config.cdat", rollUpDirs), os.O_RDONLY, 0777)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_nodesConfig_OpenFile failed")
		return nil, err
	}

	var ncfgIo io.Reader

	ncfgIo = flate.NewReader(nodesConf)
	nodeCfgDec := gob.NewDecoder(ncfgIo)
	err = nodeCfgDec.Decode(&loadCofigs)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_NodesConfig Decode failed")
		return nil, err
	}
	err = ncfgIo.(io.Closer).Close()
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_Abnormaly config Io Closing failed")
		return nil, err
	}
	//=========load buckets ===================
	loadBuckets := make([]string, 0)

	nodesBucket, err := os.OpenFile(fmt.Sprintf("%s/buckets.cdat", rollUpDirs), os.O_RDONLY, 0777)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_Bucket_OpenFile failed")
		return nil, err
	}

	var bucketIo io.Reader

	bucketIo = flate.NewReader(nodesBucket)
	bucketDec := gob.NewDecoder(bucketIo)
	err = bucketDec.Decode(&loadBuckets)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_Bucket Decode failed")
		return nil, err
	}
	err = bucketIo.(io.Closer).Close()
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Rollup_Abnormaly Bucket Io Closing failed")
		return nil, err
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
	return recoveryBuckets, nil
}

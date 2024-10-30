package data_access_layer

import (
	"compress/flate"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/sjy-dv/nnv/pkg/hnsw"
)

const parentDir string = "./data_dir"

func Commit(nodeTrees *hnsw.HnswBucket) error {
	metaTime := time.Now().UnixNano()
	dirs, err := os.ReadDir(parentDir)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Commit_ReadDir failed")
		return err
	}

	var dirsList []string
	for _, dir := range dirs {
		if dir.IsDir() {
			if dir.Name() == "matchid" || dir.Name() == "backup" {
				continue
			}
			dirsList = append(dirsList, dir.Name())
		}
	}

	sort.Strings(dirsList)
	var nextNumber int
	if len(dirsList) > 0 {
		lastDir := dirsList[len(dirsList)-1]
		lastNumber, err := strconv.Atoi(lastDir)
		if err != nil {
			log.Warn().Err(err).Msg("data_access_layer.convert dir name string to int failed")
			return err
		}
		nextNumber = lastNumber + 1
		// if not first file
		// second file delete
		if len(dirsList) > 1 {
			oldDir := dirsList[len(dirsList)-2]
			err = os.RemoveAll(filepath.Join(parentDir, oldDir))
			if err != nil {
				log.Warn().Err(err).Msg("data_access_layer.OldDir Remove failed")
				return err
			}
		}
	} else {
		nextNumber = 1
	}

	newDirName := fmt.Sprintf("%08d", nextNumber)
	newDirPath := filepath.Join(parentDir, newDirName)
	err = os.Mkdir(newDirPath, os.ModePerm)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Mkdir failed")
		return err
	}

	var nodesOut io.Writer
	var nodesCofingOut io.Writer
	var bucketsOut io.Writer

	backupNodes := make(map[string][]hnsw.Node)
	backupConfig := make(map[string]hnsw.HnswConfig)
	backupBuckets := make([]string, 0)
	for _, nodes := range nodeTrees.Buckets {
		backupNodes[nodes.BucketName] = nodes.NodeList.Nodes
		backupConfig[nodes.BucketName] = hnsw.HnswConfig{
			Efconstruction: nodes.Efconstruction,
			M:              nodes.M,
			Mmax:           nodes.Mmax,
			Mmax0:          nodes.Mmax0,
			Ml:             nodes.Ml,
			Ep:             nodes.Ep,
			MaxLevel:       nodes.MaxLevel,
			Dim:            nodes.Dim,
			DistanceType:   nodes.DistanceType,
			Heuristic:      nodes.Heuristic,
			BucketName:     nodes.BucketName,
			Filter:         nodes.Filter,
			EmptyNodes:     nodes.EmptyNodes,
		}
		backupBuckets = append(backupBuckets, nodes.BucketName)
	}
	//=========write backup nodes =====================//
	f, err := os.OpenFile(fmt.Sprintf("%s/nodes.cdat", newDirPath), os.O_TRUNC|
		os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Commit_nodesCdat_OpenFile failed")
		return err
	}

	nodesOut, _ = flate.NewWriter(f, flate.BestCompression)
	enc := gob.NewEncoder(nodesOut)
	if err := enc.Encode(backupNodes); err != nil {
		log.Warn().Err(err).Msg("data_access_layer.write nodes data.cdat failed")
		return err
	}
	if flusher, ok := nodesOut.(interface{ Flush() error }); ok {
		if err := flusher.Flush(); err != nil {
			log.Warn().Err(err).Msg("Failed to flush data")
			return err
		}
	}

	if err := nodesOut.(io.Closer).Close(); err != nil {
		log.Warn().Err(err).Msg("data_access_layer.write nodes data close failed")
		return err
	}
	// enc := gob.NewEncoder(nodesOut)
	// enc.Encode(backupNodes)
	// err = nodesOut.(io.Closer).Close()
	// if err != nil {
	// 	log.Warn().Err(err).Msg("data_access_layer.write nodes data.cdat failed")
	// 	return err
	// }

	//============write backup hnsw nodes config =============
	cf, err := os.OpenFile(fmt.Sprintf("%s/nodes_config.cdat", newDirPath), os.O_TRUNC|
		os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Commit_configCdat_OpenFile failed")
		return err
	}
	nodesCofingOut, _ = flate.NewWriter(cf, flate.BestCompression)

	cenc := gob.NewEncoder(nodesCofingOut)
	cenc.Encode(backupConfig)
	err = nodesCofingOut.(io.Closer).Close()
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.write nodes_config data.cdat failed")
		return err
	}

	//==================write bucket list=======================
	bf, err := os.OpenFile(fmt.Sprintf("%s/buckets.cdat", newDirPath), os.O_TRUNC|
		os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.Commit_bucketCdat_OpenFile failed")
		return err
	}
	bucketsOut, _ = flate.NewWriter(bf, flate.BestCompression)

	benc := gob.NewEncoder(bucketsOut)
	benc.Encode(backupBuckets)
	err = bucketsOut.(io.Closer).Close()
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.write buckets data.cdat failed")
		return err
	}

	//================write lasttimestamp meta.json================
	type meta struct {
		Timestamp int64 `json:"timestamp"`
	}
	metaJson := meta{
		Timestamp: metaTime,
	}
	mf, err := os.Create(fmt.Sprintf("%s/meta.json", newDirPath))
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.create meta timestamp file failed")
		return err
	}
	jenc := json.NewEncoder(mf)
	jenc.Encode(metaJson)
	err = mf.Close()
	if err != nil {
		log.Warn().Err(err).Msg("data_access_layer.write meta timestamp file failed")
		return err
	}
	return nil
}

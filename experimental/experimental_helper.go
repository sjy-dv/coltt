package experimental

import (
	"bytes"
	"fmt"
	"math"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/coltt/edge"
	"github.com/sjy-dv/coltt/gen/protoc/v3/experimentalproto"
)

func (emv *ExperimentalMultiVector) LoadAuthorizationBuckets() error {
	authorizationBuckets, err := emv.Storage.LoadBucketList()
	if err != nil {
		return err
	}
	stateManager.Exists.Lock.Lock()
	defer stateManager.Exists.Lock.Unlock()
	for _, bucket := range authorizationBuckets {
		stateManager.Exists.collections[bucket] = true
	}
	return nil
}

func errorWrap(errMsg string) *experimentalproto.Error {
	return &experimentalproto.Error{
		ErrorMessage: errMsg,
		ErrorCode:    experimentalproto.ErrorCode_INTERNAL_FUNC_ERROR,
	}
}

func newAuthorizationBucketHelper(collectionName string) {
	stateManager.Exists.Lock.Lock()
	stateManager.Exists.collections[collectionName] = true
	stateManager.Exists.Lock.Unlock()
	stateManager.Load.Lock.Lock()
	stateManager.Load.collections[collectionName] = true
	stateManager.Load.Lock.Unlock()
}

func eliminateBucketMemoryHelper(collectionName string) {
	stateManager.Load.Lock.Lock()
	delete(stateManager.Load.collections, collectionName)
	stateManager.Load.Lock.Unlock()
}

func destroyBucketHelper(collectionName string) {
	stateManager.Exists.Lock.Lock()
	delete(stateManager.Exists.collections, collectionName)
	stateManager.Exists.Lock.Unlock()
	stateManager.Load.Lock.Lock()
	delete(stateManager.Load.collections, collectionName)
	stateManager.Load.Lock.Unlock()
}

func authorization(collectionName string) error {
	if !hasCollection(collectionName) {
		return fmt.Errorf(edge.ErrCollectionNotFound, collectionName)
	}
	if !alreadyLoadCollection(collectionName) {
		return fmt.Errorf(edge.ErrCollectionNotLoad, collectionName)
	}
	return nil
}

func (emv *ExperimentalMultiVector) saveMetadataHelper(collectionName string, data []byte) error {
	return emv.Storage.PutObject(collectionName, fmt.Sprintf("%s.meta.json", collectionName), bytes.NewReader(data), int64(len(data)))
}

func (emv *ExperimentalMultiVector) saveVertexHelper(collectionName string, data []byte) error {
	return emv.Storage.PutObject(collectionName, fmt.Sprintf("%s.vertex", collectionName), bytes.NewReader(data), int64(len(data)))
}

func (emv *ExperimentalMultiVector) BucketLifeCycleJob(collectionName string) {
	versioning, err := emv.Storage.IsVersionBucket(collectionName)
	if err != nil {
		log.Error().Msgf("bucket [%s] version check error: %s", collectionName, err.Error())
	}
	if versioning {
		emv.Storage.VersionCleanUp(collectionName)
	}
}

func (emv *ExperimentalMultiVector) loadMetadataHelper(collectionName string) ([]byte, error) {
	return emv.Storage.GetObject(collectionName, fmt.Sprintf("%s.meta.json", collectionName))
}

func (emv *ExperimentalMultiVector) loadVertexHelper(collectionName string) ([]byte, error) {
	return emv.Storage.GetObject(collectionName, fmt.Sprintf("%s.vertex", collectionName))
}

func indexDesignAnalyze(indexDesign []*experimentalproto.Index) map[string]IndexFeature {
	features := make(map[string]IndexFeature)
	for _, column := range indexDesign {
		features[column.IndexName] = IndexFeature{
			IndexName:  column.IndexName,
			IndexType:  int32(column.IndexType),
			EnableNull: column.EnableNull,
		}
	}
	return features
}

func reverseIndexDesign(features map[string]IndexFeature) []*experimentalproto.Index {
	design := make([]*experimentalproto.Index, 0)
	for _, column := range features {
		design = append(design, &experimentalproto.Index{
			IndexName:  column.IndexName,
			IndexType:  experimentalproto.IndexType(column.IndexType),
			EnableNull: column.EnableNull,
		})
	}
	return design
}

func scoreHelper(score float32, dist string) float32 {
	if dist == T_COSINE {
		return ((2 - score) / 2) * 100
	}
	return float32(math.Max(0, float64(100-score)))
}

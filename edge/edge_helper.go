package edge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/coltt/gen/protoc/v4/edgepb"
	"github.com/sjy-dv/coltt/pkg/inverted"
)

func (helper *Edge) LoadAuthorizationBuckets() error {
	authorizationBuckets, err := helper.Storage.LoadBucketList()
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

func errorWrap(errMsg string) *edgepb.Error {
	return &edgepb.Error{
		ErrorMessage: errMsg,
		ErrorCode:    edgepb.ErrorCode_INTERNAL_FUNC_ERROR,
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
		return fmt.Errorf(ErrCollectionNotFound, collectionName)
	}
	if !alreadyLoadCollection(collectionName) {
		return fmt.Errorf(ErrCollectionNotLoad, collectionName)
	}
	return nil
}

func (helper *Edge) saveMetadataHelper(collectionName string, data []byte) error {
	return helper.Storage.PutObject(collectionName, fmt.Sprintf("%s.meta.json", collectionName), bytes.NewReader(data), int64(len(data)))
}

func (helper *Edge) saveVertexHelper(collectionName string, data []byte) error {
	return helper.Storage.PutObject(collectionName, fmt.Sprintf("%s.vertex", collectionName), bytes.NewReader(data), int64(len(data)))
}

func (helper *Edge) saveInvertedIndexHelper(collectionName string, data []byte) error {
	return helper.Storage.PutObject(collectionName, fmt.Sprintf("%s.inverted.raw", collectionName), bytes.NewReader(data), int64(len(data)))
}

func (helper *Edge) BucketLifeCycleJob(collectionName string) {
	versioning, err := helper.Storage.IsVersionBucket(collectionName)
	if err != nil {
		log.Error().Msgf("bucket [%s] version check error: %s", collectionName, err.Error())
	}
	if versioning {
		helper.Storage.VersionCleanUp(collectionName)
	}
}

func (helper *Edge) loadMetadataHelper(collectionName string) ([]byte, error) {
	return helper.Storage.GetObject(collectionName, fmt.Sprintf("%s.meta.json", collectionName))
}

func (helper *Edge) loadVertexHelper(collectionName string) ([]byte, error) {
	return helper.Storage.GetObject(collectionName, fmt.Sprintf("%s.vertex", collectionName))
}

func (helper *Edge) loadInvertedIndexHelper(collectionName string) ([]byte, error) {
	return helper.Storage.GetObject(collectionName, fmt.Sprintf("%s.inverted.raw", collectionName))
}

func indexDesignAnalyze(indexDesign []*edgepb.Index) map[string]IndexFeature {
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

func reverseIndexDesign(features map[string]IndexFeature) []*edgepb.Index {
	design := make([]*edgepb.Index, 0)
	for _, column := range features {
		design = append(design, &edgepb.Index{
			IndexName:  column.IndexName,
			IndexType:  edgepb.IndexType(column.IndexType),
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

func convertProtoOp(op edgepb.Op) inverted.FilterOp {
	switch op {
	case edgepb.Op_EQ:
		return inverted.OpEqual
	case edgepb.Op_NEQ:
		return inverted.OpNotEqual
	case edgepb.Op_GT:
		return inverted.OpGreaterThan
	case edgepb.Op_GTE:
		return inverted.OpGreaterThanEqual
	case edgepb.Op_LT:
		return inverted.OpLessThan
	case edgepb.Op_LTE:
		return inverted.OpLessThanEqual
	default:
		return inverted.OpEqual
	}
}

func convertProtoLogicalOperator(op edgepb.LogicalOperator) inverted.LogicalOp {
	if op == edgepb.LogicalOperator_OR {
		return inverted.LogicalOr
	}
	return inverted.LogicalAnd
}

func convertBytesMetadata(data []byte) (edgepb.Quantization, error) {
	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return edgepb.Quantization_None, err
	}
	return edgepb.Quantization(metadata.Quantization), nil
}

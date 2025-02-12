package core

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"

	"github.com/sjy-dv/coltt/core/vectorindex"
	"github.com/sjy-dv/coltt/gen/protoc/v3/coreproto"
	"github.com/sjy-dv/coltt/pkg/distance"
	"github.com/sjy-dv/coltt/pkg/index"
	"github.com/vmihailenco/msgpack/v5"
)

func errorWrap(errMsg string) *coreproto.Error {
	return &coreproto.Error{
		ErrorMessage: errMsg,
		ErrorCode:    coreproto.ErrorCode_INTERNAL_FUNC_ERROR,
	}
}

func (xx *Core) diskClear(collectionName string) {
	//delete disk config
	configKey := fmt.Sprintf(diskRule0, collectionName)
	xx.DataStore.Del(collectionName)
	xx.CommitLog.Delete([]byte(configKey))
	xx.CommitLog.AscendKeys([]byte(fmt.Sprintf(diskRule2, collectionName)),
		true, func(k []byte) (bool, error) {
			err := xx.CommitLog.Delete(k)
			if err != nil {
				// after code
			}
			return true, nil
		})
}

func (xx *Core) memFree(collectionName string) {
	xx.DataStore.Del(collectionName)

	indexdb.indexLock.Lock()
	delete(indexdb.indexes, collectionName)
	indexdb.indexLock.Unlock()

	stateManager.auth.authLock.Lock()
	stateManager.auth.collections[collectionName] = false
	stateManager.auth.authLock.Unlock()
}

func (xx *Core) createSnapshotHelper(collectionName string) error {
	var buf bytes.Buffer
	index := xx.DataStore.Get(collectionName)
	err := index.Commit(&buf, true)
	if err != nil {
		return err
	}
	return os.WriteFile(fmt.Sprintf(noQuantizationRule, collectionName), buf.Bytes(), 0644)
}

func (xx *Core) snapShotHelper(collectionName string, dim uint32, dist distance.Space, searchOpts vectorindex.HnswOption) error {
	data, err := os.ReadFile(fmt.Sprintf(noQuantizationRule, collectionName))
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(data)
	hnsw := vectorindex.NewHnsw(uint(dim), dist, searchOpts)
	err = hnsw.Load(buf, true)
	if err != nil {
		return err
	}
	xx.DataStore.Set(collectionName, hnsw)
	return nil
}

func protoDistHelper(dist coreproto.Distance) (distance.Space, string) {
	if dist == coreproto.Distance_Cosine {
		return distance.NewCosine(), COSINE
	}
	return distance.NewEuclidean(), EUCLIDEAN
}

func reverseprotoDistHelper(dist string) coreproto.Distance {
	if dist == COSINE {
		return coreproto.Distance_Cosine
	}
	return coreproto.Distance_Euclidean
}

func reversesingleprotoDistHelper(dist string) distance.Space {
	if dist == COSINE {
		return distance.NewCosine()
	}
	return distance.NewEuclidean()
}

func protoSearchAlgoHelper(algo coreproto.SearchAlgorithm) (string, vectorindex.HnswOption) {
	if algo == coreproto.SearchAlgorithm_Simple {
		return "simple", vectorindex.HnswSearchAlgorithm(vectorindex.HnswSearchSimple)
	}
	return "heuristic", vectorindex.HnswSearchAlgorithm(vectorindex.HnswSearchHeuristic)
}

func reverseSearchAlgoHelper(alg string) vectorindex.HnswOption {
	if alg == "simple" {
		return vectorindex.HnswSearchAlgorithm(vectorindex.HnswSearchSimple)
	}
	return vectorindex.HnswSearchAlgorithm(vectorindex.HnswSearchHeuristic)
}

func reverseConfigHelper(config vectorindex.ProtoConfig) *coreproto.HnswConfig {
	return &coreproto.HnswConfig{
		SearchAlgorithm: func() coreproto.SearchAlgorithm {
			if config.SearchAlgorithm == "simple" {
				return coreproto.SearchAlgorithm_Simple
			}
			return coreproto.SearchAlgorithm_Heuristic
		}(),
		LevelMultiplier:           config.LevelMultiplier,
		Ef:                        int32(config.Ef),
		EfConstruction:            int32(config.EfConstruction),
		M:                         int32(config.M),
		MMax:                      int32(config.MMax),
		MMax0:                     int32(config.MMax0),
		HeuristicExtendCandidates: config.HeuristicExtendCandidates,
		HeuristicKeepPruned:       config.HeuristicKeepPruned,
	}
}

func indexSaveHelper(collectionName string) error {
	_, err := os.Stat(fmt.Sprintf(indexRule, collectionName))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		os.Remove(fmt.Sprintf(indexRule, collectionName))

	}
	return indexdb.indexes[collectionName].SerializeBinary(fmt.Sprintf(indexRule, collectionName))
}

func indexLoadHelper(collectionName string) error {
	_, err := os.Stat(fmt.Sprintf(indexRule, collectionName))
	if err != nil {
		if os.IsNotExist(err) {
			if stateManager.checker.collections[collectionName] {
				goto EmptyIndex
			}

			return fmt.Errorf("collection BitmapIndex: %s is not defined [Not Found Collection Error]", collectionName)
		}
		return err
	}
	goto ExistsIndex
EmptyIndex:
	indexdb.indexLock.Lock()
	indexdb.indexes[collectionName] = index.NewBitmapIndex()
	defer indexdb.indexLock.Unlock()
	return nil
ExistsIndex:
	recoveryIndex := index.NewBitmapIndex()
	err = recoveryIndex.DeserializeBinary(fmt.Sprintf(indexRule, collectionName))
	if err != nil {
		// guess damaged file
		return err
	}
	err = recoveryIndex.ValidateIndex()
	if err != nil {
		// validation failed also damaged
		return err
	}
	indexdb.indexLock.Lock()
	indexdb.indexes[collectionName] = recoveryIndex
	defer indexdb.indexLock.Unlock()
	return nil
}

func stateTrueHelper(collectionName string) {
	stateManager.checker.cecLock.Lock()
	defer stateManager.checker.cecLock.Unlock()
	stateManager.auth.authLock.Lock()
	defer stateManager.auth.authLock.Unlock()
	stateManager.checker.collections[collectionName] = true
	stateManager.auth.collections[collectionName] = true
}

func stateRegistHelper(collectionName string) {
	stateManager.checker.cecLock.Lock()
	defer stateManager.checker.cecLock.Unlock()
	stateManager.checker.collections[collectionName] = true
}

func stateFalseHelper(collectionName string) {
	stateManager.auth.authLock.Lock()
	defer stateManager.auth.authLock.Unlock()
	stateManager.auth.collections[collectionName] = false
}

func stateDestroyHelper(collectionName string) {
	stateManager.auth.authLock.Lock()
	defer stateManager.auth.authLock.Unlock()
	stateManager.checker.cecLock.Lock()
	defer stateManager.checker.cecLock.Unlock()
	delete(stateManager.auth.collections, collectionName)
	delete(stateManager.checker.collections, collectionName)

}

func collectionStatusHelper(collectionName string) error {
	if !hasCollection(collectionName) {
		return fmt.Errorf(ErrCollectionNotFound, collectionName)
	}
	if !alreadyLoadCollection(collectionName) {
		return fmt.Errorf(ErrCollectionNotLoad, collectionName)
	}
	return nil
}

func (xx *Core) rollbackForConsistentHelper(collectionName string, commitId uint64, metadata map[string]interface{}) {
	if err := xx.CommitLog.Delete([]byte(fmt.Sprintf(diskRule1, collectionName, commitId))); err != nil {
		// using log after
	}
	hnsw := xx.DataStore.Get(collectionName)
	if err := hnsw.Remove(commitId); err != nil {
		//
	}
	if err := indexdb.indexes[collectionName].Remove(commitId, metadata); err != nil {
		//
	}
}

func scoreHelper(score float32, dist string) float32 {
	if dist == COSINE {
		return ((2 - score) / 2) * 100
	}
	return float32(math.Max(0, float64(100-score)))
}

func (xx *Core) saveCollection(collectionName string) error {
	ok, err := xx.CommitLog.Exist([]byte(diskColList))
	if err != nil {
		return err
	}
	if ok {
		gb, err := xx.CommitLog.Get([]byte(diskColList))
		if err != nil {
			return err
		}
		data := make([]string, 0)
		err = msgpack.Unmarshal(gb, &data)
		if err != nil {
			return err
		}
		data = append(data, collectionName)
		byts, err := msgpack.Marshal(data)
		if err != nil {
			return err
		}
		err = xx.CommitLog.Put([]byte(diskColList), byts)
		if err != nil {
			return err
		}
	} else {
		data := []string{collectionName}
		bytes, err := msgpack.Marshal(data)
		if err != nil {
			return err
		}
		err = xx.CommitLog.Put([]byte(diskColList), bytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func (xx *Core) removeCollection(collectionName string) error {
	gb, err := xx.CommitLog.Get([]byte(diskColList))
	if err != nil {
		return err
	}
	data := make([]string, 0)
	err = msgpack.Unmarshal(gb, &data)
	if err != nil {
		return err
	}

	newData := make([]string, 0)
	for _, d := range data {
		if d != collectionName {
			newData = append(newData, d)
		}
	}
	byts, err := msgpack.Marshal(newData)
	if err != nil {
		return err
	}
	err = xx.CommitLog.Put([]byte(diskColList), byts)
	if err != nil {
		return err
	}
	return nil
}

func (xx *Core) RegistCollectionStManager() error {
	ok, err := xx.CommitLog.Exist([]byte(diskColList))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	gb, err := xx.CommitLog.Get([]byte(diskColList))
	if err != nil {
		return err
	}
	data := make([]string, 0)
	err = msgpack.Unmarshal(gb, &data)
	if err != nil {
		return err
	}
	for _, col := range data {
		stateRegistHelper(col)
	}
	return nil
}

func (xx *Core) exitSnapshot() error {
	for col, ok := range stateManager.auth.collections {
		if ok {
			xx.ReleaseCollection(context.Background(), &coreproto.CollectionName{
				CollectionName: col,
			})
		}
	}
	return nil
}

func (xx *Core) chkValidDimensionality(collectionName string, dim int32) error {
	collection := xx.DataStore.Get(collectionName)
	if collection.Dim() != uint32(dim) {
		return fmt.Errorf("Err Collection %s expects %d dimensions, but has %d dimensions", collectionName, collection.Dim(), dim)
	}
	return nil
}

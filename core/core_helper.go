package core

import (
	"bytes"
	"fmt"
	"os"

	"github.com/sjy-dv/nnv/core/vectorindex"
	"github.com/sjy-dv/nnv/gen/protoc/v3/coreproto"
	"github.com/sjy-dv/nnv/pkg/distancer"
	"github.com/sjy-dv/nnv/pkg/index"
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

func (xx *Core) snapShotHelper(collectionName string, dim uint32, dist distancer.Provider, searchOpts vectorindex.HnswOption) error {
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

func protoDistHelper(dist coreproto.Distance) (distancer.Provider, string) {
	if dist == coreproto.Distance_Cosine {
		return distancer.NewCosineDistanceProvider(), COSINE
	}
	return distancer.NewL2SquaredProvider(), EUCLIDEAN
}

func reverseprotoDistHelper(dist string) coreproto.Distance {
	if dist == COSINE {
		return coreproto.Distance_Cosine
	}
	return coreproto.Distance_Euclidean
}

func reversesingleprotoDistHelper(dist string) distancer.Provider {
	if dist == COSINE {
		return distancer.NewCosineDistanceProvider()
	}
	return distancer.NewL2SquaredProvider()
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

package core

import (
	"fmt"

	"github.com/sjy-dv/nnv/core/vectorindex"
	"github.com/sjy-dv/nnv/gen/protoc/v3/coreproto"
	"github.com/sjy-dv/nnv/pkg/distancer"
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

func protoSearchAlgoHelper(algo coreproto.SearchAlgorithm) (string, vectorindex.HnswOption) {
	if algo == coreproto.SearchAlgorithm_Simple {
		return "simple", vectorindex.HnswSearchAlgorithm(vectorindex.HnswSearchSimple)
	}
	return "heuristic", vectorindex.HnswSearchAlgorithm(vectorindex.HnswSearchHeuristic)
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

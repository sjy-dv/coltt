package vectorindex

import (
	"fmt"

	"github.com/sjy-dv/nnv/pkg/gomath"
)

var hnswSearchAlgorithmNames = [...]string{
	"Simple",
	"Heuristic",
}

type hnswSearchAlgorithm int

const (
	HnswSearchSimple hnswSearchAlgorithm = iota
	HnswSearchHeuristic
)

func (a hnswSearchAlgorithm) String() string {
	return hnswSearchAlgorithmNames[a]
}

// Options
type HnswOption interface {
	apply(*hnswConfig)
}

type hnswOption struct {
	applyFunc func(*hnswConfig)
}

func (opt *hnswOption) apply(config *hnswConfig) {
	opt.applyFunc(config)
}

func HnswLevelMultiplier(value float32) HnswOption {
	return &hnswOption{func(config *hnswConfig) {
		config.levelMultiplier = value
	}}
}

func HnswEf(value int) HnswOption {
	return &hnswOption{func(config *hnswConfig) {
		config.ef = value
	}}
}

func HnswEfConstruction(value int) HnswOption {
	return &hnswOption{func(config *hnswConfig) {
		config.efConstruction = value
	}}
}

func HnswM(value int) HnswOption {
	return &hnswOption{func(config *hnswConfig) {
		config.m = value
	}}
}

func HnswMmax(value int) HnswOption {
	return &hnswOption{func(config *hnswConfig) {
		config.mMax = value
	}}
}

func HnswMmax0(value int) HnswOption {
	return &hnswOption{func(config *hnswConfig) {
		config.mMax0 = value
	}}
}

func HnswSearchAlgorithm(value hnswSearchAlgorithm) HnswOption {
	return &hnswOption{func(config *hnswConfig) {
		config.searchAlgorithm = value
	}}
}

func HnswHeuristicExtendCandidates(value bool) HnswOption {
	return &hnswOption{func(config *hnswConfig) {
		config.heuristicExtendCandidates = value
	}}
}

func HnswHeuristicKeepPruned(value bool) HnswOption {
	return &hnswOption{func(config *hnswConfig) {
		config.heuristicKeepPruned = value
	}}
}

type hnswConfig struct {
	searchAlgorithm           hnswSearchAlgorithm
	levelMultiplier           float32
	ef                        int
	efConstruction            int
	m                         int
	mMax                      int
	mMax0                     int
	heuristicExtendCandidates bool
	heuristicKeepPruned       bool
}

func newHnswConfig(options []HnswOption) *hnswConfig {
	config := &hnswConfig{
		searchAlgorithm:           HnswSearchSimple,
		levelMultiplier:           -1,
		ef:                        20,
		efConstruction:            200,
		m:                         16,
		mMax:                      -1,
		mMax0:                     -1,
		heuristicExtendCandidates: false,
		heuristicKeepPruned:       true,
	}
	for _, option := range options {
		option.apply(config)
	}

	if config.levelMultiplier == -1 {
		config.levelMultiplier = 1.0 / gomath.Log(float32(config.m))
	}
	if config.mMax == -1 {
		config.mMax = config.m
	}
	if config.mMax0 == -1 {
		config.mMax0 = 2 * config.m
	}

	return config
}

func (this *hnswConfig) String() string {
	return fmt.Sprintf(
		"searchAlgorithm: %s, ef: %d, efConstruction: %d, m: %d, mMax: %d, mMax0: %d, levelMultiplier: %.4f, extendCandidates: %t, keepPruned: %t",
		this.searchAlgorithm,
		this.ef,
		this.efConstruction,
		this.m,
		this.mMax,
		this.mMax0,
		this.levelMultiplier,
		this.heuristicExtendCandidates,
		this.heuristicKeepPruned,
	)
}

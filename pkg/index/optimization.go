package index

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

func (idx *BitmapIndex) StartOptimization(interval time.Duration) {
	if idx.optimizationTicker != nil {
		return
	}

	idx.optimizationTicker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-idx.optimizationTicker.C:

			case <-idx.stopOptimization:
				idx.optimizationTicker.Stop()
				return
			}
		}
	}()
}

func (idx *BitmapIndex) StopOptimization() {
	if idx.optimizationTicker == nil {
		return
	}
	idx.stopOptimization <- true
}

func (idx *BitmapIndex) optimize() {
	log.Info().Msg("Starting index optimization...")
	start := time.Now()

	idx.shardLock.RLock()
	defer idx.shardLock.RUnlock()

	var wg sync.WaitGroup

	for _, shard := range idx.Shards {
		wg.Add(1)
		go func(s *IndexShard) {
			defer wg.Done()
			s.rmu.Lock()
			// after
			s.rmu.Unlock()
		}(shard)
	}
	wg.Wait()
	elapsed := time.Since(start)
	log.Info().Msgf("Index optimization completed in %v\n", elapsed)
}

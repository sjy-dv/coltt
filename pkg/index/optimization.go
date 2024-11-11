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

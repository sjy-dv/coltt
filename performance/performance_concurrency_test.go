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

package performance_test

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

type MemDB struct {
	Memory []int
	lock   sync.Mutex
}

var itertime int = 1_000_000

func BenchmarkPureJob(b *testing.B) {
	memdb := MemDB{}
	runtime.GC()
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)
	start := time.Now()
	b.ReportAllocs()

	for i := 0; i < itertime; i++ {
		memdb.Memory = append(memdb.Memory, i)
	}
	if itertime != len(memdb.Memory) {
		b.Fail()
	}
	end := time.Since(start)
	runtime.ReadMemStats(&mAfter)
	memUsed := mAfter.HeapAlloc - mBefore.HeapAlloc
	memUsedMB := float64(memUsed) / (1024 * 1024)
	b.Logf("Estimated memory used : %.2f mb\n", memUsedMB)
	b.Logf("process time: %.2f sec", end.Seconds())
}

// goos: windows
// goarch: amd64
// pkg: github.com/sjy-dv/coltt/performance
// cpu: Intel(R) Core(TM) Ultra 9 185H
// BenchmarkPureJob-22    	1000000000	         0.007214 ns/op	       0 B/op	       0 allocs/op
// --- BENCH: BenchmarkPureJob-22
//     performance_concurrency_test.go:37: Estimated memory used : 0.01 mb
//     performance_concurrency_test.go:38: process time: 0.02 sec
//     performance_concurrency_test.go:37: Estimated memory used : 14.49 mb
//     performance_concurrency_test.go:38: process time: 0.01 sec
//     performance_concurrency_test.go:37: Estimated memory used : 14.50 mb
//     performance_concurrency_test.go:38: process time: 0.01 sec
//     performance_concurrency_test.go:37: Estimated memory used : 14.50 mb
//     performance_concurrency_test.go:38: process time: 0.01 sec
//     performance_concurrency_test.go:37: Estimated memory used : 0.00 mb
//     performance_concurrency_test.go:38: process time: 0.01 sec
// 	... [output truncated]
// PASS
// ok  	github.com/sjy-dv/coltt/performance	0.476s

func BenchmarkGoroutineJob(b *testing.B) {
	memdb := MemDB{}
	runtime.GC()
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)
	start := time.Now()
	b.ReportAllocs()
	eg := errgroup.Group{}
	workerpool := runtime.GOMAXPROCS(0) * 2
	b.Logf("ready workerpool : %d", workerpool)
	sema := make(chan struct{}, workerpool)
	for i := 0; i < itertime; i++ {
		sema <- struct{}{}
		copyi := i
		eg.Go(func() error {
			defer func() { <-sema }()

			memdb.lock.Lock()
			memdb.Memory = append(memdb.Memory, copyi)
			memdb.lock.Unlock()
			return nil
		})
	}
	err := eg.Wait()
	if err != nil {
		b.Fatal(err)
	}
	if itertime != len(memdb.Memory) {
		b.Fail()
	}
	end := time.Since(start)
	runtime.ReadMemStats(&mAfter)
	memUsed := mAfter.HeapAlloc - mBefore.HeapAlloc
	memUsedMB := float64(memUsed) / (1024 * 1024)
	b.Logf("Estimated memory used : %.2f mb\n", memUsedMB)
	b.Logf("process time: %.2f sec", end.Seconds())
}

// goos: windows
// goarch: amd64
// pkg: github.com/sjy-dv/coltt/performance
// cpu: Intel(R) Core(TM) Ultra 9 185H
// BenchmarkGoroutineJob-22    	1000000000	         0.5354 ns/op	       0 B/op	       0 allocs/op
// --- BENCH: BenchmarkGoroutineJob-22
//     performance_concurrency_test.go:70: ready workerpool : 44
//     performance_concurrency_test.go:95: Estimated memory used : 23.15 mb
//     performance_concurrency_test.go:96: process time: 0.56 sec
//     performance_concurrency_test.go:70: ready workerpool : 44
//     performance_concurrency_test.go:95: Estimated memory used : 22.88 mb
//     performance_concurrency_test.go:96: process time: 0.58 sec
//     performance_concurrency_test.go:70: ready workerpool : 44
//     performance_concurrency_test.go:95: Estimated memory used : 22.86 mb
//     performance_concurrency_test.go:96: process time: 0.56 sec
//     performance_concurrency_test.go:70: ready workerpool : 44
// 	... [output truncated]
// PASS
// ok  	github.com/sjy-dv/coltt/performance	16.446s

package performance_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/huandu/skiplist"
	"github.com/sjy-dv/coltt/pkg/concurrentmap"
)

func BenchmarkIterGoMap(b *testing.B) {
	b.N = 1_000_000
	b.ReportAllocs()
	bucket := make(map[uint64][]float32)

	for i := 0; i < b.N; i++ {
		bucket[uint64(i)] = generateRandomVector(128)
	}

	runtime.GC()
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)
	start := time.Now()
	for k, v := range bucket {
		_ = k
		_ = v
	}
	end := time.Since(start)
	runtime.ReadMemStats(&mAfter)
	memUsed := mAfter.HeapAlloc - mBefore.HeapAlloc
	memUsedMB := float64(memUsed) / (1024 * 1024)
	b.Logf("Estimated memory used by myMap: %.2f mb\n", memUsedMB)
	b.Logf("iter time: %.2f sec", end.Seconds())
}

//  performance_iter_test.go:34: iter time: 0.01 sec

func BenchmarkIterGoArray(b *testing.B) {
	b.N = 1_000_000
	b.ReportAllocs()
	bucket := make([][]float32, 0)

	for i := 0; i < b.N; i++ {
		bucket = append(bucket, generateRandomVector(128))
	}

	runtime.GC()
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)
	start := time.Now()
	for k, v := range bucket {
		_ = k
		_ = v
	}
	end := time.Since(start)
	runtime.ReadMemStats(&mAfter)
	memUsed := mAfter.HeapAlloc - mBefore.HeapAlloc
	memUsedMB := float64(memUsed) / (1024 * 1024)
	b.Logf("Estimated memory used by myMap: %.2f mb\n", memUsedMB)
	b.Logf("iter time: %.2f sec", end.Seconds())
}

//  performance_iter_test.go:61: iter time: 0.00 sec

func BenchmarkIterConcurrentMap(b *testing.B) {
	b.N = 1_000_000
	b.ReportAllocs()
	bucket := concurrentmap.New[uint64, []float32]()

	for i := 0; i < b.N; i++ {
		bucket.Set(uint64(i), generateRandomVector(128))
	}

	runtime.GC()
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)
	start := time.Now()
	bucket.ForEach(func(u uint64, f []float32) bool {
		_ = u
		_ = f
		return true
	})
	end := time.Since(start)
	runtime.ReadMemStats(&mAfter)
	memUsed := mAfter.HeapAlloc - mBefore.HeapAlloc
	memUsedMB := float64(memUsed) / (1024 * 1024)
	b.Logf("Estimated memory used by myMap: %.2f mb\n", memUsedMB)
	b.Logf("iter time: %.2f sec", end.Seconds())
}

// performance_iter_test.go:94: iter time: 0.16 sec

func BenchmarkIterSkiplist(b *testing.B) {
	b.N = 1_000_000
	b.ReportAllocs()
	bucket := skiplist.New(skiplist.Uint64)

	for i := 0; i < b.N; i++ {
		bucket.Set(uint64(i), generateRandomVector(128))
	}

	runtime.GC()
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)
	start := time.Now()
	for i := 0; i < bucket.Len(); i++ {
		key := bucket.Find(uint64(i))
		_ = key.Value.([]float32)
	}
	end := time.Since(start)
	runtime.ReadMemStats(&mAfter)
	memUsed := mAfter.HeapAlloc - mBefore.HeapAlloc
	memUsedMB := float64(memUsed) / (1024 * 1024)
	b.Logf("Estimated memory used by myMap: %.2f mb\n", memUsedMB)
	b.Logf("iter time: %.2f sec", end.Seconds())
}

//performance_iter_test.go:116: iter time: 0.19 sec

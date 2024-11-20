package performance_test

import (
	"math/rand/v2"
	"runtime"
	"testing"

	"github.com/huandu/skiplist"
	"github.com/sjy-dv/nnv/pkg/concurrentmap"
)

func BenchmarkInsertGoMap(b *testing.B) {
	runtime.GC()
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)
	b.N = 1_000_000
	b.ReportAllocs()
	bucket := make(map[uint64][]float32)

	for i := 0; i < b.N; i++ {
		bucket[uint64(i)] = generateRandomVector(128)
	}
	runtime.ReadMemStats(&mAfter)
	memUsed := mAfter.HeapAlloc - mBefore.HeapAlloc
	memUsedMB := float64(memUsed) / (1024 * 1024)
	b.Logf("Estimated memory used by myMap: %.2f mb\n", memUsedMB)
}

// BenchmarkInsertGoMap-22    	 1000000	      1074 ns/op	     674 B/op	       1 allocs/op
// --- BENCH: BenchmarkInsertGoMap-22
//     performance_test.go:23: Estimated memory used by myMap: 601.95 bytes

func BenchmarkInsertConcurrentMap(b *testing.B) {
	runtime.GC()
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)
	b.N = 1_000_000
	b.ReportAllocs()
	bucket := concurrentmap.New[uint64, []float32]()

	for i := 0; i < b.N; i++ {
		bucket.Set(uint64(i), generateRandomVector(128))
	}
	runtime.ReadMemStats(&mAfter)
	memUsed := mAfter.HeapAlloc - mBefore.HeapAlloc
	memUsedMB := float64(memUsed) / (1024 * 1024)
	b.Logf("Estimated memory used by myMap: %.2f mb\n", memUsedMB)
}

// BenchmarkInsertConcurrentMap-22    	 1000000	      1534 ns/op	     617 B/op	       3 allocs/op
// --- BENCH: BenchmarkInsertConcurrentMap-22
//     performance_test.go:46: Estimated memory used by myMap: 573.03 mb

func BenchmarkSkiplist(b *testing.B) {
	runtime.GC()
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)
	b.N = 1_000_000
	b.ReportAllocs()
	bucket := skiplist.New(skiplist.Uint64)

	for i := 0; i < b.N; i++ {
		bucket.Set(uint64(i), generateRandomVector(128))
	}
	runtime.ReadMemStats(&mAfter)
	memUsed := mAfter.HeapAlloc - mBefore.HeapAlloc
	memUsedMB := float64(memUsed) / (1024 * 1024)
	b.Logf("Estimated memory used by myMap: %.2f mb\n", memUsedMB)
}

// BenchmarkSkiplist-22    	 1000000	      1014 ns/op	     656 B/op	       4 allocs/op
// --- BENCH: BenchmarkSkiplist-22
//     performance_test.go:108: Estimated memory used by myMap: 626.01 mb

func generateRandomVector(dim int) []float32 {
	vec := make([]float32, dim)
	for i := 0; i < dim; i++ {
		vec[i] = rand.Float32()
	}
	return vec
}

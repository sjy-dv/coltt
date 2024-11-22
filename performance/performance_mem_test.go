package performance_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/types/known/structpb"
)

func BenchmarkProtoMessage(b *testing.B) {
	runtime.GC()
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)
	mustkeep := "test msg"
	m := map[string]interface{}{"id": mustkeep}
	pm := make(map[uint64]*structpb.Struct)
	start := time.Now()
	for i := 0; i < 1_000_000; i++ {
		val, err := structpb.NewStruct(m)
		if err != nil {
			b.Fatal(err)
		}
		pm[uint64(i)] = val
	}
	end := time.Since(start)
	runtime.ReadMemStats(&mAfter)
	memUsed := mAfter.HeapAlloc - mBefore.HeapAlloc
	memUsedMB := float64(memUsed) / (1024 * 1024)
	b.Logf("Estimated memory used to insert: %.2f mb\n", memUsedMB)
	b.Logf("Insert time: %.2f sec", end.Seconds())

	runtime.GC()
	runtime.ReadMemStats(&mBefore)
	start = time.Now()
	for i := 1; i < 1_000_000; i++ {
		get := pm[uint64(i)]
		msg := get.AsMap()["id"].(string)
		if msg != mustkeep {
			b.Fail()
		}
	}
	end = time.Since(start)
	runtime.ReadMemStats(&mAfter)
	memUsed = mAfter.HeapAlloc - mBefore.HeapAlloc
	memUsedMB = float64(memUsed) / (1024 * 1024)
	b.Logf("Estimated memory used to read: %.2f mb\n", memUsedMB)
	b.Logf("Get time: %d ns", end.Nanoseconds())
}

// --- BENCH: BenchmarkProtoMessage-22
//     performance_mem_test.go:30: Estimated memory used to insert: 426.31 mb
//     performance_mem_test.go:31: Insert time: 0.53 sec
//     performance_mem_test.go:47: Estimated memory used to read: 55.22 mb
//     performance_mem_test.go:48: Get time: 3854_68300 ns

func BenchmarkMsgPackBytes(b *testing.B) {
	runtime.GC()
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)
	mustkeep := "test msg"
	m := map[string]interface{}{"id": mustkeep}
	pm := make(map[uint64][]byte)
	start := time.Now()
	for i := 0; i < 1_000_000; i++ {
		msgp, err := msgpack.Marshal(m)
		if err != nil {
			b.Fatal(err)
		}
		pm[uint64(i)] = msgp
	}
	end := time.Since(start)
	runtime.ReadMemStats(&mAfter)
	memUsed := mAfter.HeapAlloc - mBefore.HeapAlloc
	memUsedMB := float64(memUsed) / (1024 * 1024)
	b.Logf("Estimated memory used to insert: %.2f mb\n", memUsedMB)
	b.Logf("Insert time: %.2f sec", end.Seconds())

	runtime.GC()
	runtime.ReadMemStats(&mBefore)
	start = time.Now()
	for i := 1; i < 1_000_000; i++ {
		get := pm[uint64(i)]
		am := make(map[string]interface{})
		err := msgpack.Unmarshal(get, &am)
		if err != nil {
			b.Fatal(err)
		}
		msg := am["id"].(string)
		if msg != mustkeep {
			b.Fail()
		}
	}
	end = time.Since(start)
	runtime.ReadMemStats(&mAfter)
	memUsed = mAfter.HeapAlloc - mBefore.HeapAlloc
	memUsedMB = float64(memUsed) / (1024 * 1024)
	b.Logf("Estimated memory used to read: %.2f mb\n", memUsedMB)
	b.Logf("Get time: %d ns", end.Nanoseconds())
}

// --- BENCH: BenchmarkMsgPackBytes-22
//     performance_mem_test.go:83: Estimated memory used to insert: 181.53 mb
//     performance_mem_test.go:84: Insert time: 0.36 sec
//     performance_mem_test.go:105: Estimated memory used to read: 83.64 mb
//     performance_mem_test.go:106: Get time: 4246_29400 ns

func BenchmarkPureMap(b *testing.B) {
	runtime.GC()
	var mBefore, mAfter runtime.MemStats
	runtime.ReadMemStats(&mBefore)
	mustkeep := "test msg"
	m := map[string]interface{}{"id": mustkeep}
	pm := make(map[uint64]map[string]interface{})
	start := time.Now()
	for i := 0; i < 1_000_000; i++ {
		pm[uint64(i)] = m
	}
	end := time.Since(start)
	runtime.ReadMemStats(&mAfter)
	memUsed := mAfter.HeapAlloc - mBefore.HeapAlloc
	memUsedMB := float64(memUsed) / (1024 * 1024)
	b.Logf("Estimated memory used to insert: %.2f mb\n", memUsedMB)
	b.Logf("Insert time: %.2f sec", end.Seconds())

	runtime.GC()
	runtime.ReadMemStats(&mBefore)
	start = time.Now()
	for i := 1; i < 1_000_000; i++ {
		get := pm[uint64(i)]
		msg := get["id"].(string)
		if msg != mustkeep {
			b.Fail()
		}
	}
	end = time.Since(start)
	runtime.ReadMemStats(&mAfter)
	memUsed = mAfter.HeapAlloc - mBefore.HeapAlloc
	memUsedMB = float64(memUsed) / (1024 * 1024)
	b.Logf("Estimated memory used to read: %.2f mb\n", memUsedMB)
	b.Logf("Get time: %d ns", end.Nanoseconds())
}

// --- BENCH: BenchmarkPureMap-22
//     performance_mem_test.go:124: Estimated memory used to insert: 60.06 mb
//     performance_mem_test.go:125: Insert time: 0.10 sec
//     performance_mem_test.go:141: Estimated memory used to read: 0.00 mb
//     performance_mem_test.go:142: Get time: 717_14500 ns

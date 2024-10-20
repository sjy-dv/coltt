package conversion

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/rs/zerolog/log"
)

var Float32ToBytes func([]float32) []byte = float32ToBytesSafe
var BytesToFloat32 func([]byte) []float32 = bytesToFloat32Safe

func init() {
	// Determine native endianness
	var i uint16 = 0xABCD
	isLittleEndian := *(*byte)(unsafe.Pointer(&i)) == 0xCD
	log.Info().Bool("isLittleEndian", isLittleEndian).Msg("Endianness")
	if isLittleEndian {
		Float32ToBytes = float32ToBytesRaw
		BytesToFloat32 = bytesToFloat32Raw
	}
}

func Uint64ToBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	return b
}

func BytesToUint64(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}

func SingleFloat32ToBytes(f float32) []byte {
	b := [4]byte{}
	binary.LittleEndian.PutUint32(b[:], math.Float32bits(f))
	return b[:]
}

func BytesToSingleFloat32(b []byte) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(b))
}

func float32ToBytesSafe(f []float32) []byte {
	b := make([]byte, len(f)*4)
	for i, v := range f {
		binary.LittleEndian.PutUint32(b[i*4:], math.Float32bits(v))
	}
	return b
}

func bytesToFloat32Safe(b []byte) []float32 {
	// We allocate a new slice because the original byte slice may be disposed.
	// Most likely this byte slices comes from a BoltDB transaction.
	f := make([]float32, len(b)/4)
	for i := range f {
		f[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return f
}

func float32ToBytesRaw(f []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&f[0])), len(f)*4)
}

func bytesToFloat32Raw(b []byte) []float32 {
	f := make([]float32, len(b)/4)
	copy(f, unsafe.Slice((*float32)(unsafe.Pointer(&b[0])), len(b)/4))
	return f
	// If we know the values are never used outside a transaction then we can
	// shortcut like below. Use with caution.
	// return unsafe.Slice((*float32)(unsafe.Pointer(&b[0])), len(b)/4)
}

func EdgeListToBytes(edges []uint64) []byte {
	b := make([]byte, len(edges)*8)
	for i, e := range edges {
		binary.LittleEndian.PutUint64(b[i*8:], e)
	}
	return b
}

func BytesToEdgeList(b []byte) []uint64 {
	edges := make([]uint64, len(b)/8)
	for i := range edges {
		edges[i] = binary.LittleEndian.Uint64(b[i*8:])
	}
	return edges
}

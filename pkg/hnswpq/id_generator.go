package hnswpq

import "sync/atomic"

type OffsetCounter struct {
	count uint64
}

var offsetCounter *OffsetCounter

func NewOffsetCounter(lastoffset uint64) {
	offsetCounter = new(OffsetCounter)
	offsetCounter = &OffsetCounter{count: lastoffset}
}

func NextId() uint64 {
	return atomic.AddUint64(&offsetCounter.count, 1)
}

func GetCurId() uint64 {
	return atomic.LoadUint64(&offsetCounter.count)
}

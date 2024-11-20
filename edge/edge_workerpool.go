package edge

import (
	"context"
	"runtime"
	"sync"
)

type Job struct {
	Vectorspace vectorspace
	Target      Vector
	TopK        int
	ResultChan  chan *ResultSet
	CancelFunc  context.CancelFunc
}

type KResult struct {
	ResultSet *ResultSet
	Error     error
}

var (
	workerPoolSize = runtime.NumCPU()
	jobQueue       = make(chan Job, workerPoolSize*2)
	once           sync.Once
)

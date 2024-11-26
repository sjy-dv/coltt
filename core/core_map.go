package core

import "sync"

type autoMap[T any] struct {
	mem  map[string]T
	lock sync.RWMutex
}

func NewAutoMap[T any]() *autoMap[T] {
	return &autoMap[T]{
		mem: make(map[string]T),
	}
}

func (xx *autoMap[T]) Set(k string, v T) {
	xx.lock.Lock()
	defer xx.lock.Unlock()
	xx.mem[k] = v
}

func (xx *autoMap[T]) Del(k string) {
	xx.lock.Lock()
	defer xx.lock.Unlock()
	delete(xx.mem, k)
}

func (xx *autoMap[T]) Get(k string) T {
	xx.lock.RLock()
	defer xx.lock.RUnlock()
	return xx.mem[k]
}

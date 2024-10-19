package cache

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

type Cachable interface {
	SizeInMemory() int64
}

type sharedCacheElem struct {
	item       Cachable
	lastAccess time.Time
	mu         sync.RWMutex
	scrapped   bool
}

type Manager struct {
	maxSize      int64
	sharedCaches map[string]*sharedCacheElem
	mu           sync.Mutex
}

func NewManager(maxSize int64) *Manager {
	return &Manager{
		sharedCaches: make(map[string]*sharedCacheElem),
		maxSize:      maxSize,
	}
}

func (self *Manager) Release(name string) {
	self.mu.Lock()
	defer self.mu.Unlock()
	delete(self.sharedCaches, name)
	log.Debug().Str("action", "cache.manager.ReleaseFn").
		Str("name", name).
		Int("numCaches", len(self.sharedCaches)).Msg("release cache")
}

func (self *Manager) checkAndPrune() {
	if self.maxSize == -1 {
		return
	}
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.maxSize == 0 {
		clear(self.sharedCaches)
		return
	}
	type cacheElem struct {
		name       string
		lastAccess time.Time
		size       int64
	}
	caches := make([]cacheElem, 0, len(self.sharedCaches))
	totalSize := int64(0)
	for n, s := range self.sharedCaches {
		ssize := s.item.SizeInMemory()
		caches = append(caches, cacheElem{name: n, size: ssize, lastAccess: s.lastAccess})
		totalSize += ssize
	}

	if totalSize <= self.maxSize {
		return
	}

	slices.SortFunc(caches, func(a, b cacheElem) int {
		return a.lastAccess.Compare(b.lastAccess)
	})

	for _, s := range caches {
		if totalSize <= self.maxSize {
			break
		}
		delete(self.sharedCaches, s.name)
		log.Debug().Str("action", "cache.manager.checkAndPruneFn").
			Str("name", s.name).Msg("pruning cache")
		totalSize -= s.size
	}

}

type Transaction struct {
	writtenCaches map[string]*sharedCacheElem
	mu            sync.Mutex
	manager       *Manager
	failed        atomic.Bool
}

func (self *Manager) NewTransaction() *Transaction {
	return &Transaction{
		writtenCaches: make(map[string]*sharedCacheElem),
		manager:       self,
	}
}

func (self *Transaction) With(
	name string, readOnly bool,
	createFn func() (Cachable, error),
	f func(cacheToUse Cachable) error) error {
	if self.failed.Load() {
		return errors.New("tx has already failed")
	}

	self.manager.mu.Lock()
	if exCache, ok := self.manager.sharedCaches[name]; ok {
		exCache.lastAccess = time.Now()
		self.manager.mu.Unlock()

		cacheToUse := exCache
		if readOnly {
			self.mu.Lock()
			_, ok := self.writtenCaches[name]
			self.mu.Unlock()
			if !ok {
				if exCache.mu.TryRLock() {
					defer exCache.mu.RUnlock()
				} else {
					log.Debug().Str("action", "cache.Transaction.WithFn").
						Str("name", name).
						Msg("create read only cold cache")
					freshCachable, err := createFn()
					if err != nil {
						self.failed.Store(true)
						return fmt.Errorf("error while creating fresh read only cold cache: %w", err)
					}
					cacheToUse = &sharedCacheElem{
						item:       freshCachable,
						lastAccess: time.Now(),
					}
				}
			}
		} else {
			self.mu.Lock()
			if _, ok := self.writtenCaches[name]; !ok {
				exCache.mu.Lock()
				self.writtenCaches[name] = exCache
			}
			self.mu.Unlock()
		}
		if cacheToUse.scrapped {
			log.Debug().Str("action", "cache.Transaction.WithFn").
				Str("name", name).Bool("read-only", readOnly).Msg("cache is scrapped, using temporary new cache")
			freshCachable, err := createFn()
			if err != nil {
				self.failed.Store(true)
				return fmt.Errorf("error while creating fresh cold temporary cache: %w", err)
			}
			cacheToUse = &sharedCacheElem{
				item:       freshCachable,
				lastAccess: time.Now(),
			}
		}
		if cacheToUse == exCache {
			log.Debug().Str("action", "cache.Transaction.WithFn").
				Str("name", name).Bool("readOnly", readOnly).Msg("reusing cache")
			defer self.manager.checkAndPrune()
		}
		if err := f(cacheToUse.item); err != nil {
			self.failed.Store(true)
			cacheToUse.scrapped = true
			self.manager.mu.Lock()
			delete(self.manager.sharedCaches, name)
			self.manager.mu.Unlock()
			return fmt.Errorf("error while executing cache operation: %w", err)
		}
		return nil
	}
	log.Debug().Str("action", "cache.Transaction.WithFn").
		Str("name", name).Bool("readOnly", readOnly).Msg("Creating new cache")
	freshCachable, err := createFn()
	if err != nil {
		self.failed.Store(true)
		self.manager.mu.Unlock()
		return fmt.Errorf("error while creating fresh cache: %w", err)
	}
	s := &sharedCacheElem{
		item:       freshCachable,
		lastAccess: time.Now(),
	}
	if self.manager.maxSize != 0 {
		self.manager.sharedCaches[name] = s
		defer self.manager.checkAndPrune()
	}
	if readOnly {
		s.mu.RLock()
		defer s.mu.RUnlock()
	} else {
		s.mu.Lock()
		self.mu.Lock()
		self.writtenCaches[name] = s
		self.mu.Unlock()
	}
	self.manager.mu.Unlock()
	if err := f(s.item); err != nil {
		self.failed.Store(true)
		s.scrapped = true
		self.manager.mu.Lock()
		delete(self.manager.sharedCaches, name)
		self.manager.mu.Unlock()
		return fmt.Errorf("error while executing on new cache operation: %w", err)
	}
	return nil
}

func (self *Transaction) Commit(fail bool) {
	self.mu.Lock()
	defer self.mu.Unlock()
	if len(self.writtenCaches) == 0 {
		return
	}
	self.manager.mu.Lock()
	defer self.manager.mu.Unlock()
	failed := self.failed.Load() || fail
	for name, s := range self.writtenCaches {
		if failed {
			s.scrapped = true
			delete(self.manager.sharedCaches, name)
		}
		log.Debug().Str("action", "cache.Transaction.CommitFn").
			Str("name", name).Bool("failed", failed).Msg("commit cache")
		s.mu.Unlock()
	}
}

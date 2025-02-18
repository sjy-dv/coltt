package experimental

import (
	"sync"
)

var stateManager *collectionCoordinator

func NewStateManager() {
	stateManager = &collectionCoordinator{
		Exists: &collectionExistChecker{
			collections: make(map[string]bool),
		},
		Load: &authorizationCollection{
			collections: make(map[string]bool),
		},
	}
}

type collectionCoordinator struct {
	Exists *collectionExistChecker
	Load   *authorizationCollection
}

type collectionExistChecker struct {
	Lock        sync.RWMutex
	collections map[string]bool
}

type authorizationCollection struct {
	collections map[string]bool
	Lock        sync.RWMutex
}

func hasCollection(collectionName string) bool {
	stateManager.Exists.Lock.RLock()
	exists := stateManager.Exists.collections[collectionName]
	stateManager.Exists.Lock.RUnlock()
	return exists
}

func alreadyLoadCollection(collectionName string) bool {
	stateManager.Load.Lock.RLock()
	exists := stateManager.Load.collections[collectionName]
	stateManager.Load.Lock.RUnlock()
	return exists
}

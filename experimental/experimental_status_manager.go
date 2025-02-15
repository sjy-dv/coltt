package experimental

import "sync"

var stateManager *collectionCoordinator

func NewStateManager() {
	stateManager = &collectionCoordinator{
		checker: &collectionExistChecker{
			collections: make(map[string]bool),
		},
		auth: &authorizationCollection{
			collections: make(map[string]bool),
		},
	}
}

type collectionCoordinator struct {
	checker *collectionExistChecker
	auth    *authorizationCollection
}

type collectionExistChecker struct {
	cecLock     sync.RWMutex
	collections map[string]bool
}

type authorizationCollection struct {
	collections map[string]bool
	authLock    sync.RWMutex
}

func hasCollection(collectionName string) bool {
	stateManager.checker.cecLock.RLock()
	exists := stateManager.checker.collections[collectionName]
	stateManager.checker.cecLock.RUnlock()
	return exists
}

func alreadyLoadCollection(collectionName string) bool {
	stateManager.auth.authLock.RLock()
	exists := stateManager.auth.collections[collectionName]
	stateManager.auth.authLock.RUnlock()
	return exists
}

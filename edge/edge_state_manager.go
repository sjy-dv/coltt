// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package edge

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

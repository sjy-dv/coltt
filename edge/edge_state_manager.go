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

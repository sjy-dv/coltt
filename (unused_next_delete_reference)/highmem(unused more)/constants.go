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

package highmem

import (
	"errors"
	"sync"
)

var (
	fLinkCdat            = "./data_dir/%s.cdat"
	backupfLinkCdat      = "./data_dir/%s-backup.cdat"
	indexBin             = "./data_dir/%s.bin"
	backupIndexBin       = "./data_dir/%s-backup.bin"
	tensorLink           = "./data_dir/%s.tensor"
	backupTensorLink     = "./data_dir/%s-backup.tensor"
	confJson             = "./data_dir/%s_conf.json"
	backupConfJson       = "./data_dir/%s_conf-backup.json"
	metaJson             = "./data_dir/meta.json"
	panicr               = "panic %v"
	collectionJson       = "./data_dir/collection.json"
	backupCollectionJson = "./data_dir/collection-backup.json"
	commitLog            = "./data_dir/commit-log"
	commitCollection     = "back-log"
	fatalCommit          = "fatal-log"
)

var errUnrecoverable = errors.New("unrecoverable error")
var UncaughtPanicError = "uncaught panic error: %v"
var notFoundCollection = "collection: %s is not defined [NOT FOUND COLLECTION]"
var notLoadCollection = "collection: %s is not load. please try to `LoadCollection` [NOT FOUND COLLECTION IN MEMORY]"
var stateManager *collectionCoordinator

func NewStateManager() {
	stateManager = &collectionCoordinator{
		loadchecker: &collectionLoadChecker{
			collections: make(map[string]bool),
		},
		checker: &collectionExistChecker{
			collections: make(map[string]bool),
		},
		auth: &authorizationCollection{
			collections: make(map[string]bool),
		},
	}
}

type collectionCoordinator struct {
	loadchecker *collectionLoadChecker
	checker     *collectionExistChecker
	auth        *authorizationCollection
}
type collectionLoadChecker struct {
	clcLock     sync.RWMutex
	collections map[string]bool
}
type collectionExistChecker struct {
	cecLock     sync.RWMutex
	collections map[string]bool
}

type authorizationCollection struct {
	collections map[string]bool
	authLock    sync.RWMutex
}

var tensorCapacity uint = 0

type functionAttempt int

const (
	retryBinaryDo functionAttempt = iota
	retryCommitLogDo
	scaleUpCapacity
)

type event int

const (
	INSERT event = iota
	UPDATE
	DELETE
)

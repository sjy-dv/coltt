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

package kv

import "errors"

var (
	ErrKeyIsEmpty                    = errors.New("the key is empty")
	ErrKeyNotFound                   = errors.New("key not found in database")
	ErrDatabaseIsUsing               = errors.New("the database directory is used by another process")
	ErrReadOnlyBatch                 = errors.New("the batch is read only")
	ErrBatchCommitted                = errors.New("the batch is committed")
	ErrDBClosed                      = errors.New("the database is closed")
	ErrDBDirectoryISEmpty            = errors.New("the database directory path can not be empty")
	ErrWaitMemtableSpaceTimeOut      = errors.New("wait memtable space timeout, try again later")
	ErrDBIteratorUnsupportedTypeHASH = errors.New("hash index does not support iterator")
)

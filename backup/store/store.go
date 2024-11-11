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

package store

type Store interface {
	Begin(update bool) (Tx, error)
	Close() error
}

type Tx interface {
	Set(key, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
	Cursor(forward bool) (Cursor, error)
	Commit() error
	Rollback() error
}

type Cursor interface {
	Seek(key []byte) error
	Next()
	Valid() bool
	Item() (Item, error)
	Close() error
}

type Item struct {
	Key, Value []byte
}

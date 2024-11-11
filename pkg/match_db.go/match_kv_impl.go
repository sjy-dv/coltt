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

package matchdbgo

import (
	"strconv"

	"github.com/sjy-dv/nnv/kv"
)

var mdb *kv.DB

func Open() error {
	opts := kv.DefaultOptions
	opts.DirPath = "./data_dir/matchid"
	db, err := kv.Open(opts)
	if err != nil {
		return err
	}
	mdb = db
	return nil
}

func Close() error {
	return mdb.Close()
}

func Get(key string) (uint32, error) {

	nodeId, err := mdb.Get([]byte(key))
	if err != nil {
		return 0, err
	}
	uintId, err := strconv.ParseUint(string(nodeId), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(uintId), nil
}

func Set(key string, val uint32) error {
	bk := []byte(key)
	bv := []byte(strconv.FormatUint(uint64(val), 10))
	return mdb.Put(bk, bv)
}

func Delete(key string) error {
	return mdb.Delete([]byte(key))
}

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

package main

import (
	"fmt"
	"time"

	"github.com/sjy-dv/nnv/kv"
)

func main() {
	opts := kv.DefaultOptions
	opts.DirPath = "./data_dir/tmp"

	db, err := kv.Open(opts)
	if err != nil {
		panic(err)
	}

	opts.DirPath = "./data_dir/tmp2"
	db2, err := kv.Open(opts)
	if err != nil {
		panic(err)
	}
	time.Sleep(10 * time.Second)
	if err := db.Close(); err != nil {
		fmt.Println(err)
	}
	if err := db2.Close(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("closing")
}

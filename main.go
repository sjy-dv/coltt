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
	"reflect"

	"github.com/sjy-dv/nnv/backup"
	"github.com/sjy-dv/nnv/backup/document"
	"github.com/sjy-dv/nnv/backup/query"
)

func main() {
	tdb, err := backup.Open("./tmp/test")
	fmt.Println(err)
	log := document.NewDocument()
	tdb.CreateCollection("test")
	log.Set("metadata", map[string]interface{}{
		"event": "how",
		"value": 10.0,
	})
	log.Set("vector", []float32{0.1, 0.2, 0.4})
	tdb.Insert("test", log)

	getL, err := tdb.FindFirst(query.NewQuery("test"))
	fmt.Println(err)
	recoverMap := getL.Get("metadata").(map[string]interface{})
	fmt.Println(recoverMap, reflect.TypeOf(recoverMap))
	recoverVec := getL.Get("vector").([]interface{})
	newFloat := make([]float32, len(recoverVec))
	for i, v := range recoverVec {
		newFloat[i] = float32(v.(float64))
	}
	fmt.Println(newFloat, reflect.TypeOf(newFloat))
	defer tdb.Close()
}

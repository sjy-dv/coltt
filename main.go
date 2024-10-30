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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
)

func main() {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Fatal(err)
	}
	base64Key := base64.StdEncoding.EncodeToString(key)
	fmt.Println("Generated 32-byte base64 key:", base64Key)

	// logdb, err := backup.NewStorage(backup.WithTimestampPrecision(backup.Nanoseconds))
	// fmt.Println(err)
	// defer logdb.Close()

	// err = logdb.InsertRows([]backup.Row{
	// 	{Metric: "test", DataPoint: backup.DataPoint{
	// 		Value:     1,
	// 		Timestamp: time.Now().UnixNano(),
	// 	},
	// 		Labels: []backup.Label{
	// 			backup.Label{
	// 				Name:  "bucketok?",
	// 				Value: "maybe ok..",
	// 			},
	// 		},
	// 	},
	// })
	// fmt.Println(err)
	// p, err := logdb.Select("test", nil, 0, time.Now().UnixNano())
	// fmt.Println(err)
	// fmt.Println(p)
}

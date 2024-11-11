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

package nnlogdb

import (
	"github.com/google/uuid"
	"github.com/sjy-dv/nnv/backup"
	"github.com/sjy-dv/nnv/backup/document"
	"github.com/sjy-dv/nnv/backup/query"
)

var nnlogdb *backup.DB
var partition string

/*
	node-log

	id
	private_id => user put specifiy nodeid
	vector
	partition
	metadata
*/

const nodeLogCollection = "node-log"

func Open() error {

	logdb, err := backup.Open("./data_dir/backup-log")
	if err != nil {
		return err
	}
	partition = uuid.New().String()
	pass, err := logdb.HasCollection(nodeLogCollection)
	if err != nil {
		return err
	}
	if !pass {
		err := logdb.CreateCollection(nodeLogCollection)
		if err != nil {
			return err
		}
		err = logdb.CreateIndex(nodeLogCollection, "partition")
		if err != nil {
			return err
		}
		err = logdb.CreateIndex(nodeLogCollection, "bucket")
		if err != nil {
			return err
		}
	}
	nnlogdb = logdb
	return nil
}

func PrintlF(privateId, event, bucket string, nodeId uint32, timestamp uint64, metadata map[string]interface{}, vector []float32) *document.Document {
	printer := document.NewDocument()
	printer.Set("private_id", privateId)
	printer.Set("node_id", nodeId)
	printer.Set("bucket", bucket)
	printer.Set("timestamp", timestamp)
	printer.Set("metadata", metadata)
	printer.Set("vector", vector)
	printer.Set("event", event)
	printer.Set("partition", partition)
	return printer
}

func AddLogf(logs *document.Document) error {
	_, err := nnlogdb.InsertOne(nodeLogCollection, logs)
	return err
}

type RecoveryLog struct {
	PrivateId string
	Bucket    string
	Event     string
	NodeId    uint32
	Timestamp uint64
	Metadata  map[string]interface{}
	Vector    []float32
}

func CurPartitionLabel() string {
	return partition
}

func GetPartition(partition string) (*[]RecoveryLog, error) {
	logs, err := nnlogdb.FindAll(query.NewQuery(nodeLogCollection).Where(
		query.Field("partition").Eq(partition),
	).Sort(query.SortOption{Field: "timestamp", Direction: -1}))
	if err != nil {
		return nil, err
	}
	recoveryLogs := make([]RecoveryLog, 0, len(logs))
	for _, log := range logs {
		retry := RecoveryLog{}
		retry.Bucket = log.Get("bucket").(string)
		retry.Event = log.Get("event").(string)
		retry.Metadata = log.Get("metadata").(map[string]interface{})
		retry.NodeId = log.Get("node_id").(uint32)
		retry.Timestamp = log.Get("timestamp").(uint64)
		retry.PrivateId = log.Get("private_id").(string)
		retry.Vector = func() []float32 {
			vec := log.Get("vector").([]interface{})
			newVec := make([]float32, len(vec))
			for loc, f := range vec {
				newVec[loc] = float32(f.(float64))
			}
			return newVec
		}()
		recoveryLogs = append(recoveryLogs, retry)
	}
	return &recoveryLogs, nil
}

func NewPartition() {
	partition = uuid.New().String()
}

func DeleteLastPartition(partition string) error {
	return nnlogdb.Delete(
		query.NewQuery(nodeLogCollection).Where(
			query.Field("partition").Eq(partition),
		),
	)
}

func DropBucketLogs(bucket string) error {
	return nnlogdb.Delete(
		query.NewQuery(nodeLogCollection).Where(
			query.Field("bucket").Eq(bucket),
		),
	)
}

func Close() error {
	return nnlogdb.Close()
}

package nnlogdb

import (
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/sjy-dv/nnv/backup"
	"github.com/vmihailenco/msgpack/v5"
)

var nnlogdb backup.Storage

func Open() error {

	logdb, err := backup.NewStorage(
		backup.WithDataPath("./data_dir/backup"),
		backup.WithPartitionDuration(6*time.Hour),
		backup.WithTimestampPrecision(backup.Nanoseconds),
		// backup data saved 1-week
		backup.WithRetention(time.Hour*168),
		backup.WithLogger(&zerolog.Logger{}),
	)
	if err != nil {
		return err
	}
	nnlogdb = logdb
	return nil
}

func AddLogf(bucketName, event, userNodeId string,
	vector []float32, metadata map[string]interface{},
	eventTime int64, nodeID uint32) error {

	vecBytes, err := msgpack.Marshal(vector)
	if err != nil {
		return err
	}
	metaBytes, err := msgpack.Marshal(metadata)
	if err != nil {
		return err
	}
	err = nnlogdb.InsertRows([]backup.Row{
		{
			Metric: bucketName,
			Labels: []backup.Label{
				{
					Name:  "event",
					Value: event,
				},
				{
					Name:  "nodeID",
					Value: strconv.FormatUint(uint64(nodeID), 10),
				},
				{
					Name:  "privateID",
					Value: userNodeId,
				},
				{
					Name:  "vector",
					Value: string(vecBytes),
				},
				{
					Name:  "metadata",
					Value: string(metaBytes),
				},
			},
			DataPoint: backup.DataPoint{
				Value:     1,
				Timestamp: eventTime,
			},
		},
	})
	return err
}

// fix nodes fault data
// func GetAllPartitionLogf(bucketName string) ([]*Logf, error) {
// 	backups, err := nnlogdb.Select(bucketName, nil, 0, time.Now().UnixNano())
//     if err != nil {
//         return nil, err
//     }
//     backupLogs := make([]*Logf, 0, len(backups))
//     for _, backdata := range backups {
//         backdata.
//     }
// }

// when nodes add empty node space

type Logf struct {
	Event     string
	NodeId    uint32
	PrivateId string
}

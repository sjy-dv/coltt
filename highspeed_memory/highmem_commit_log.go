package highspeedmemory

import (
	"time"

	"github.com/google/uuid"
	"github.com/sjy-dv/nnv/backup"
	"github.com/sjy-dv/nnv/backup/document"
	"github.com/sjy-dv/nnv/backup/query"
	"github.com/sjy-dv/nnv/pkg/snowflake"
)

var autogen *snowflake.Node

func autoCommitID() uint64 {
	x := autogen.Generate()
	if x.Int64() < 0 {
		return uint64(-x.Int64())
	}
	return uint64(x.Int64())
}

func NewIdGenerator() error {
	gen, err := snowflake.NewNode(0)
	if err != nil {
		return err
	}
	autogen = gen
	return nil
}

type CommitLogger struct {
	commitDB     *backup.DB
	diskGCTicker *time.Ticker
	stopDiskGC   chan bool
	partition    string
}

type CommitLog struct {
	PrivateId      string
	NodeId         uint64
	CollectionName string
	Metadata       map[string]interface{}
	Vector         []float32
	Partition      string
	E              event
}

var commitLogger *CommitLogger

func StartCommitLogger() error {
	commitdb, err := backup.Open(commitLog)
	if err != nil {
		return err
	}
	commitLogger = &CommitLogger{
		commitDB:   commitdb,
		stopDiskGC: make(chan bool),
		partition:  uuid.New().String(),
	}
	return nil
}

func (xx *CommitLogger) Commit(data *document.Document) error {
	_, err := xx.commitDB.InsertOne(commitCollection, data)
	return err
}

func (xx *CommitLogger) printlf(
	privateId, collectionName string,
	nodeId uint64, metadata map[string]interface{},
	vector []float32, e event,
) *document.Document {
	printer := document.NewDocument()
	printer.Set("private_id", privateId)
	printer.Set("node_id", nodeId)
	printer.Set("collection", collectionName)
	printer.Set("metadata", metadata)
	printer.Set("vector", vector)
	printer.Set("partition", xx.partition)
	printer.Set("event", int(e))
	printer.Set("timestamp", time.Now().UnixNano())
	return printer
}

func (xx *CommitLogger) GetPartition(partition string) ([]CommitLog, error) {
	partitionLogs, err := xx.commitDB.FindAll(query.NewQuery(commitCollection).Where(
		query.Field("partition").Eq(xx.partition),
	).Sort(query.SortOption{Field: "timestamp", Direction: -1}))
	if err != nil {
		return []CommitLog{}, err
	}

	recoveryCommitLogs := make([]CommitLog, 0, len(partitionLogs))
	for _, log := range partitionLogs {
		recovery := CommitLog{}
		recovery.CollectionName = log.Get("collection").(string)
		recovery.Metadata = log.Get("metadata").(map[string]interface{})
		recovery.NodeId = log.Get("node_id").(uint64)
		recovery.E = event(log.Get("event").(int))
		recovery.PrivateId = log.Get("private_id").(string)
		recovery.Partition = log.Get("partition").(string)
		recovery.Vector = func() []float32 {
			vec := log.Get("vector").([]interface{})
			newVec := make([]float32, len(vec))
			for loc, f := range vec {
				newVec[loc] = float32(f.(float64))
			}
			return newVec
		}()
		recoveryCommitLogs = append(recoveryCommitLogs, recovery)
	}
	return recoveryCommitLogs, nil
}

func (xx *CommitLogger) ReleasePartition(partition string) error {
	return xx.commitDB.Delete(
		query.NewQuery(commitCollection).Where(
			query.Field("partition").Eq(partition),
		),
	)
}

func (xx *CommitLogger) ReleaseCollectionPartition(collection string) error {
	return xx.commitDB.Delete(
		query.NewQuery(commitCollection).Where(
			query.Field("collection").Eq(collection),
		),
	)
}

func (xx *CommitLogger) Close() error {
	return xx.commitDB.Close()
}

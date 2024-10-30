package data_access_layer

import (
	"encoding/gob"

	"github.com/google/uuid"
	"github.com/sjy-dv/nnv/pkg/hnsw"
)

func init() {
	//prevent `gob: type not registered for interface: uuid.UUID` Error
	gob.Register(uuid.UUID{})
	gob.Register(hnsw.Node{})
	gob.Register([]hnsw.Node{})
	gob.Register(hnsw.HnswConfig{})
	gob.Register(map[string]interface{}{})
	gob.Register(map[string][]hnsw.Node{})
	gob.Register(map[string]hnsw.HnswConfig{})
}

type BackupHnswBucket struct {
	DataNodes    []hnsw.Node
	BucketConfig hnsw.HnswConfig
	BucketName   string
}

type BackupNodes map[string][]hnsw.Node
type BackupConfig map[string]hnsw.HnswConfig
type BackupBucketList []string

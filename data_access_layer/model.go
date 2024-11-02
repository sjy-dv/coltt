package data_access_layer

import (
	"encoding/gob"

	"github.com/RoaringBitmap/roaring"
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
	gob.Register(roaring.Bitmap{})
	gob.Register(map[string]map[string]*roaring.Bitmap{})
}

type BackupHnswBucket struct {
	DataNodes    []hnsw.Node
	BucketConfig hnsw.HnswConfig
	BucketName   string
}

type BackupNodes map[string][]hnsw.Node
type BackupConfig map[string]hnsw.HnswConfig
type BackupBucketList []string

type SerdeBitmap struct {
	Data string `json:"data"`
}

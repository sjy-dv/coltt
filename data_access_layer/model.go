package data_access_layer

import "github.com/sjy-dv/nnv/pkg/hnsw"

type BackupHnswBucket struct {
	DataNodes    []hnsw.Node
	BucketConfig hnsw.HnswConfig
	BucketName   string
}

type BackupNodes map[string][]hnsw.Node
type BackupConfig map[string]hnsw.HnswConfig
type BackupBucketList []string

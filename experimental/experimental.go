package experimental

import "github.com/sjy-dv/coltt/pkg/minio"

type ExperimentalMultiVector struct {
	Collection *metaStorage
	Storage    *minio.MinioAPI
}

type Metadata struct {
	dim          uint32
	distance     int32
	quantization int32
	indexType    map[string]IndexFeature
}

type IndexFeature struct {
	IndexName  string
	IndexType  int32
	PrimaryKey bool
	EnableNull bool
	Fulltext   bool
	Filterable bool
}

func NewExperimentalMultiVector() (*ExperimentalMultiVector, error) {
	minioStorage, err := minio.NewMinio("localhost:9000")
	if err != nil {
		return nil, err
	}
	return &ExperimentalMultiVector{
		Collection: NewMetaStorage(),
		Storage:    minioStorage,
	}, nil
}

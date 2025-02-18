package experimental

import (
	"github.com/sjy-dv/coltt/gen/protoc/v3/experimentalproto"
)

type Metadata struct {
	Dim          uint32                  `json:"dim"`
	Distance     int32                   `json:"distance"`
	Quantization int32                   `json:"quantization"`
	IndexType    map[string]IndexFeature `json:"index_type"`
	Versioning   bool                    `json:"versioning"`
}

type IndexFeature struct {
	IndexName  string `json:"index_name"`
	IndexType  int32  `json:"index_type"`
	PrimaryKey bool   `json:"primary_key"`
	EnableNull bool   `json:"enable_null"`
}

func (metadata *Metadata) Dimensional() uint32 {
	return metadata.Dim
}

func (metadata *Metadata) Distancer() experimentalproto.Distance {
	return experimentalproto.Distance(metadata.Distance)
}

func (metadata *Metadata) Quantizationer() experimentalproto.Quantization {
	return experimentalproto.Quantization(metadata.Quantization)
}

func (metadata *Metadata) IndexFeatures(index string) IndexFeature {
	return metadata.IndexType[index]
}

func (metadata *Metadata) Versional() bool {
	return metadata.Versioning
}

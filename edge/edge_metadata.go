package edge

import "github.com/sjy-dv/coltt/gen/protoc/v4/edgepb"

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
	EnableNull bool   `json:"enable_null"`
	PrimaryKey bool   `json:"primary_key"`
}

func (metadata *Metadata) Dimensional() uint32 {
	return metadata.Dim
}

func (metadata *Metadata) Distancer() edgepb.Distance {
	return edgepb.Distance(metadata.Distance)
}

func (metadata *Metadata) Quantizationer() edgepb.Quantization {
	return edgepb.Quantization(metadata.Quantization)
}

func (metadata *Metadata) IndexFeatures(index string) IndexFeature {
	return metadata.IndexType[index]
}

func (metadata *Metadata) Versional() bool {
	return metadata.Versioning
}

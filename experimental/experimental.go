package experimental

type ExperimentalMultiVector struct {
	Collection *metaStorage
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

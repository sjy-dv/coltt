package core

var (
	ErrCollectionNotFound = "collection: %s not found"
	panicr                = "panic %v"
	ErrCollectionExists   = "collection: %s is already exists"
	ErrCollectionNotLoad  = "collection: %s is not loaded in memory"
)

const (
	COSINE            = "cosine-dot"
	EUCLIDEAN         = "euclidean"
	NONE_QAUNTIZATION = "none"
	F16_QUANTIZATION  = "f16"
	F8_QUANTIZATION   = "f8"
	BF16_QUANTIZATION = "bf16"
	PQ_QUANTIZATION   = "productQuantization"
	BQ_QUANTIZATION   = "binaryQuantization"
)

var (
	diskRule0 = "%s_archive"
	diskRule1 = "%s_%d" // save data segments
	diskRule2 = "%s_"   // find all collection segments data
)

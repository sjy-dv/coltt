package core

var (
	ErrCollectionNotFound = "collection: %s not found"
	panicr                = "panic %v"
	ErrCollectionExists   = "collection: %s is already exists"
	ErrCollectionNotLoad  = "collection: %s is not loaded in memory"
	ErrAlreadyRelease     = "collection: %s is already release"
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
	diskRule0   = "%s_archive"
	diskRule1   = "%s_%d" // save data segments
	diskRule2   = "%s_"   // find all collection segments data
	diskColList = "collections"
)

var (
	noQuantizationRule   = "./data_dir/%s.raw"
	f8QuantizationRule   = "./data_dir/%s.f8.raw"
	bf16QuantizationRule = "./data_dir/%s.bf16.raw"
	f16QuantizationRule  = "./data_dir/%s.f16.raw"
)

var (
	indexRule = "./data_dir/%s.bin"
)

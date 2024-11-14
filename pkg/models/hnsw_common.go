package models

type ProductQuantizerParameters struct {
	// Number of centroids to quantize to, this is the k* parameter in the paper
	// and is often set to 255 giving 256 centroids (including 0). We are
	// limiting this to maximum of 256 (uint8) to keep the overhead of this
	// process tractable.
	NumCentroids int `json:"numCentroids" binding:"required,min=2,max=256"`
	// Number of subvectors / segments / subquantizers to use, this is the m
	// parameter in the paper and is often set to 8.
	NumSubVectors int `json:"numSubVectors" binding:"required,min=2"`
	// Number of points to use to train the quantizer, it will automatically trigger training
	// when this number of points is reached.
	TriggerThreshold int `json:"triggerThreshold" binding:"required,min=1000,max=10000"`
}

type HnswConfig struct {
	Efconstruction int
	M              int
	Mmax           int
	Mmax0          int
	Ml             float64
	Ep             int64
	MaxLevel       int
	Dim            uint32
	DistanceType   string
	Heuristic      bool
	BucketName     string // using seperate vector or find prefix kv
	EmptyNodes     []uint32
}

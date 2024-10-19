package models

type Quantizer struct {
	Type    string                      `json:"type" binding:"required,oneof=none binary product"`
	Binary  *BinaryQuantizerParamaters  `json:"binary,omitempty"`
	Product *ProductQuantizerParameters `json:"product,omitempty"`
}

type BinaryQuantizerParamaters struct {
	Threshold        *float32 `json:"threshold"`
	TriggerThreshold int      `json:"triggerThreshold" binding:"min=0,max=50000"`
	DistanceMetric   string   `json:"distanceMetric" binding:"required,oneof=hamming jaccard"`
}

type ProductQuantizerParameters struct {
	NumCentroids     int `json:"numCentroids" binding:"required,min=2,max=256"`
	NumSubVectors    int `json:"numSubVectors" binding:"required,min=2"`
	TriggerThreshold int `json:"triggerThreshold" binding:"required,min=1000,max=10000"`
}

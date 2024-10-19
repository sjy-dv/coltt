package models

type IndexSchema map[string]IndexOptions

type IndexOptions struct {
	Type        string                      `json:"type" binding:"required,oneof=vectorFlat vectorVamana vectorHnsw text string integer float stringArray"`
	VectorFlat  *IndexVectorParameters      `json:"vectorFlat,omitempty"`
	VectorHnsw  *IndexVectorParameters      `json:"vectorHnsw,omitempty"`
	Text        *IndexTextParameters        `json:"text,omitempty"`
	String      *IndexStringParameters      `json:"string,omitempty"`
	StringArray *IndexStringArrayParameters `json:"stringArray,omitempty"`
}

type IndexVectorParameters struct {
	VectorSize     uint       `json:"vectorSize" binding:"required,min=1,max=4096"`
	DistanceMetric string     `json:"distanceMetric" binding:"required,oneof=euclidean cosine dot hamming jaccard haversine"`
	Quantizer      *Quantizer `json:"quantizer,omitempty"`
}

type IndexTextParameters struct {
	Analyser string `json:"analyser" binding:"required,oneof=standard"`
}

type IndexStringParameters struct {
	CaseSensitive bool `json:"caseSensitive"`
}

type IndexStringArrayParameters struct {
	IndexStringParameters
}

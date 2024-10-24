// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package models

type IndexSchema map[string]IndexOptions

type IndexOptions struct {
	Type        string                      `json:"type" binding:"required,oneof=vectorFlat vectorVamana vectorHnsw text string integer float stringArray"`
	VectorFlat  *IndexVectorFlatParameters  `json:"vectorFlat,omitempty"`
	VectorHnsw  *IndexVectorHnswParameters  `json:"vectorHnsw,omitempty"`
	Text        *IndexTextParameters        `json:"text,omitempty"`
	String      *IndexStringParameters      `json:"string,omitempty"`
	StringArray *IndexStringArrayParameters `json:"stringArray,omitempty"`
}

type IndexVectorFlatParameters struct {
	VectorSize     uint       `json:"vectorSize" binding:"required,min=1,max=4096"`
	DistanceMetric string     `json:"distanceMetric" binding:"required,oneof=euclidean cosine dot hamming jaccard haversine"`
	Quantizer      *Quantizer `json:"quantizer,omitempty"`
}

type IndexVectorHnswParameters struct {
	VectorSize     uint       `json:"vectorSize" binding:"required,min=1,max=4096"`
	DistanceMetric string     `json:"distanceMetric" binding:"required,oneof=euclidean cosine"`
	Quantizer      *Quantizer `json:"quantizer,omitempty"`
	//Maximum Number of Connections per Node
	M              uint `json:"m" binding:"required,min=1,max=100"`
	EfConstruction uint `json:"efConstruction" binding:"required,min=1,max=1000"`
}
type IndexVectorVamanaParameters struct {
	VectorSize     uint       `json:"vectorSize" binding:"required,min=1,max=4096"`
	DistanceMetric string     `json:"distanceMetric" binding:"required,oneof=euclidean cosine dot hamming jaccard haversine"`
	SearchSize     int        `json:"searchSize" binding:"min=25,max=75"`
	DegreeBound    int        `json:"degreeBound" binding:"min=32,max=64"`
	Alpha          float32    `json:"alpha" binding:"min=1.1,max=1.5"`
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

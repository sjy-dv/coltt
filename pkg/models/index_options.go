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

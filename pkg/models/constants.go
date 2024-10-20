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

const (
	DistanceEuclidean = "euclidean"
	DistanceCosine    = "cosine"
	DistanceDot       = "dot"
	DistanceHamming   = "hamming"
	DistanceJaccard   = "jaccard"
	DistanceHaversine = "haversine"
)

// ---------------------------

const (
	IndexTypeVectorFlat   = "vectorFlat"
	IndexTypeVectorVamana = "vectorVamana"
	IndexTypeText         = "text"
	IndexTypeString       = "string"
	IndexTypeInteger      = "integer"
	IndexTypeFloat        = "float"
	IndexTypeStringArray  = "stringArray"
)

// ---------------------------

const (
	OperatorContainsAll = "containsAll"
	OperatorContainsAny = "containsAny"
	OperatorEquals      = "equals"
	OperatorNotEquals   = "notEquals"
	OperatorStartsWith  = "startsWith"
	OperatorGreaterThan = "greaterThan"
	OperatorGreaterOrEq = "greaterThanOrEquals"
	OperatorLessThan    = "lessThan"
	OperatorLessOrEq    = "lessThanOrEquals"
	OperatorInRange     = "inRange"
)

// ---------------------------

const (
	QuantizerNone    = "none"
	QuantizerBinary  = "binary"
	QuantizerProduct = "product"
)

// ---------------------------

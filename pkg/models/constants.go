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

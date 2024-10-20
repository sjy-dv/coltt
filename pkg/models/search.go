package models

import (
	"fmt"

	"github.com/google/uuid"
)

type SearchRequest struct {
	Query  Query        `json:"query" binding:"required"`
	Select []string     `json:"select"`
	Sort   []SortOption `json:"sort" binding:"max=10,dive"`
	Offset int          `json:"offset" binding:"min=0"`
	Limit  int          `json:"limit" binding:"required,min=1,max=100"`
}

// ---------------------------

type Query struct {
	Property     string                     `json:"property" binding:"required"`
	VectorFlat   *SearchVectorFlatOptions   `json:"vectorFlat"`
	VectorVamana *SearchVectorVamanaOptions `json:"vectorVamana"`
	Text         *SearchTextOptions         `json:"text"`
	String       *SearchStringOptions       `json:"string"`
	Integer      *SearchIntegerOptions      `json:"integer"`
	Float        *SearchFloatOptions        `json:"float"`
	StringArray  *SearchStringArrayOptions  `json:"stringArray"`
	And          []Query                    `json:"_and" binding:"dive"`
	Or           []Query                    `json:"_or" binding:"dive"`
}

func (q Query) Validate(schema IndexSchema) error {
	// Handle recursive case
	switch q.Property {
	case "_and":
		for _, subQuery := range q.And {
			if err := subQuery.Validate(schema); err != nil {
				return err
			}
		}
		return nil
	case "_or":
		for _, subQuery := range q.Or {
			if err := subQuery.Validate(schema); err != nil {
				return err
			}
		}
		return nil
	case "_id":
		// Either string with Equals operator or stringArray with ContainsAny operator
		switch {
		case q.String != nil:
			if q.String.Operator != OperatorEquals {
				return fmt.Errorf("invalid operator %s for %s, expected %s", q.String.Operator, q.Property, OperatorEquals)
			}
			if _, err := uuid.Parse(q.String.Value); err != nil {
				return fmt.Errorf("invalid UUID %v for %s, %v", q.String.Value, q.Property, err)
			}
		case q.StringArray != nil:
			if q.StringArray.Operator != OperatorContainsAny {
				return fmt.Errorf("invalid operator %s for %s, expected %s", q.StringArray.Operator, q.Property, OperatorContainsAny)
			}
			for _, v := range q.StringArray.Value {
				if _, err := uuid.Parse(v); err != nil {
					return fmt.Errorf("invalid UUID %s for %s, %v", v, q.Property, err)
				}
			}
		default:
			return fmt.Errorf("invalid query for _id, expected string or stringArray")
		}
		return nil
	}
	// Handle base case
	value, ok := schema[q.Property]
	if !ok {
		return fmt.Errorf("property %s not found in index schema, cannot query", q.Property)
	}
	// Are the options given correctly?
	switch value.Type {
	case IndexTypeVectorFlat:
		if q.VectorFlat == nil {
			return fmt.Errorf("vectorFlat query options not provided for property %s", q.Property)
		}
		if len(q.VectorFlat.Vector) != int(value.VectorFlat.VectorSize) {
			return fmt.Errorf("vectorFlat query vector length mismatch for property %s, expected %d got %d", q.Property, value.VectorFlat.VectorSize, len(q.VectorFlat.Vector))
		}
		if q.VectorFlat.Filter != nil {
			if err := q.VectorFlat.Filter.Validate(schema); err != nil {
				return err
			}
		}
	case IndexTypeText:
		if q.Text == nil {
			return fmt.Errorf("text query options not provided for property %s", q.Property)
		}
		if q.Text.Filter != nil {
			if err := q.Text.Filter.Validate(schema); err != nil {
				return err
			}
		}
	case IndexTypeString:
		if q.String == nil {
			return fmt.Errorf("string query options not provided for property %s", q.Property)
		}
	case IndexTypeStringArray:
		if q.StringArray == nil {
			return fmt.Errorf("stringArray query options not provided for property %s", q.Property)
		}
	case IndexTypeInteger:
		if q.Integer == nil {
			return fmt.Errorf("integer query options not provided for property %s", q.Property)
		}
	case IndexTypeFloat:
		if q.Float == nil {
			return fmt.Errorf("float query options not provided for property %s", q.Property)
		}
	default:
		return fmt.Errorf("unknown index type %s", value.Type)
	}
	return nil
}

// Shared search result struct for ordered search results
type SearchResult struct {
	Point
	// Used to transmit partially decoded data to the client
	DecodedData PointAsMap
	// Internal NodeId is not exposed to the client
	NodeId uint64 `json:"-" msgpack:"-"`
	// Pointers are used to differentiate between zero values and unset values.
	// A distance or score of 0 could be valid.  Computed from vector indices,
	// lower is better
	Distance *float32 `json:"_distance,omitempty" msgpack:"_distance,omitempty"`
	// Computed from generic indices, higher is better
	Score *float32 `json:"_score,omitempty" msgpack:"_score,omitempty"`
	// Combined final score
	HybridScore float32 `json:"_hybridScore" msgpack:"_hybridScore"`
}

// ---------------------------

type SortOption struct {
	Property   string `json:"property" binding:"required"`
	Descending bool   `json:"descending"`
}

type SearchVectorVamanaOptions struct {
	Vector     []float32 `json:"vector" binding:"required,max=4096"`
	Operator   string    `json:"operator" binding:"required,oneof=near"`
	SearchSize int       `json:"searchSize" binding:"required,min=25,max=75"`
	Limit      int       `json:"limit" binding:"required,min=1,max=75"`
	Filter     *Query    `json:"filter"`
	Weight     *float32  `json:"weight"`
}

type SearchVectorFlatOptions struct {
	Vector   []float32 `json:"vector" binding:"required,max=4096"`
	Operator string    `json:"operator" binding:"required,oneof=near"`
	Limit    int       `json:"limit" binding:"required,min=1,max=75"`
	Filter   *Query    `json:"filter"`
	Weight   *float32  `json:"weight"`
}

type SearchTextOptions struct {
	Value    string   `json:"value" binding:"required"`
	Operator string   `json:"operator" binding:"required,oneof=containsAll containsAny"`
	Limit    int      `json:"limit" binding:"required,min=1,max=75"`
	Filter   *Query   `json:"filter"`
	Weight   *float32 `json:"weight"`
}

type SearchStringOptions struct {
	Value    string `json:"value" binding:"required"`
	Operator string `json:"operator" binding:"required,oneof=equals notEquals startsWith greaterThan greaterThanOrEquals lessThan lessThanOrEquals inRange"`
	// Used for range queries
	EndValue string `json:"endValue"`
}

type SearchIntegerOptions struct {
	Value    int64  `json:"value" binding:"required"`
	Operator string `json:"operator" binding:"required,oneof=equals notEquals greaterThan greaterThanOrEquals lessThan lessThanOrEquals inRange"`
	EndValue int64  `json:"endValue"`
}

type SearchFloatOptions struct {
	Value    float64 `json:"value" binding:"required"`
	Operator string  `json:"operator" binding:"required,oneof=equals notEquals greaterThan greaterThanOrEquals lessThan lessThanOrEquals inRange"`
	EndValue float64 `json:"endValue"`
}

type SearchStringArrayOptions struct {
	Value    []string `json:"value" binding:"required"`
	Operator string   `json:"operator" binding:"required,oneof=containsAll containsAny"`
}

package models

import (
	"bytes"
	"fmt"

	"github.com/google/uuid"
	"github.com/vmihailenco/msgpack/v5"
)

type PointAsMap map[string]any

func (p PointAsMap) ExtractIdField(createNew bool) (uuid.UUID, error) {
	id, ok := p["_id"]
	if !ok {
		if createNew {
			return uuid.New(), nil
		}
		return uuid.Nil, fmt.Errorf("missing _id field")
	}
	stringId, ok := id.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid id type, expected string got %T", id)
	}
	parsedId, err := uuid.Parse(stringId)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid id format, %s", err.Error())
	}
	// We remove the _id field because it is internal and not part of the
	// actual point data. But to keep the API simpler, we combine the user
	// request to include the _id field.
	delete(p, "_id")
	return parsedId, nil
}

// This is the internal data type for a point that the system uses. We extract
// the id field for quick access and encode all the data into a byte slice.
type Point struct {
	Id   uuid.UUID
	Data []byte
}

func (p *Point) GetField(name string) (any, error) {
	dec := msgpack.NewDecoder(bytes.NewReader(p.Data))
	queryResult, err := dec.Query(name)
	if err != nil {
		return nil, fmt.Errorf("failed to query point field %s: %w", name, err)
	}
	if len(queryResult) == 0 {
		// This is not an error, it just means the field is not present
		return nil, nil
	}
	return queryResult[0], nil
}

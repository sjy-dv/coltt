package index

import (
	"github.com/sjy-dv/vemoo/pkg/models"
	"github.com/sjy-dv/vemoo/storage"
)

type IndexInvertedString struct {
	inner  *IndexInverted[string]
	params models.IndexStringParameters
}

func NewIndexInvertedString(storg storage.Storage,
	params models.IndexStringParameters) *IndexInvertedString {
	inv := NewIndexInverted[string](storg)
	return &IndexInvertedString{inner: inv, params: params}
}


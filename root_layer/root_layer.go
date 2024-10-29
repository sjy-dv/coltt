package rootlayer

import (
	"context"

	"github.com/sjy-dv/nnv/root_layer/standalone"
)

// after code refactoring

func NewRootLayer() error {
	return standalone.NewRootLayer()
}

func StableRelease(ctx context.Context) error {
	return standalone.StableRelease(ctx)
}

package rootlayer

import (
	"errors"
	"runtime"

	"github.com/nats-io/nats.go"
	"github.com/sjy-dv/nnv/config"
	"github.com/sjy-dv/nnv/kv"
	"github.com/sjy-dv/nnv/pkg/hnsw"
	"google.golang.org/grpc"
)

var roots = &RootLayer{}

func NewRootLayer() error {
	roots = &RootLayer{
		VBucket:     &hnsw.HnswBucket{},
		Bucket:      &kv.DB{},
		S:           &grpc.Server{},
		StreamLayer: &nats.Conn{},
	}
	err := roots.VBucket.Start(nil)
	if err != nil {
		roots.Log.Warn().Err(err).Msg("root_layer.root.go(23) vBucket start failed")
		return err
	}
	kvOpts := &kv.DefaultOptions
	kvOpts.DirPath = "./data_dir/kv"
	roots.Bucket, err = kv.Open(*kvOpts)
	if err != nil {
		roots.Log.Warn().Err(err).Msg("root_layer.root.go(27) kv bucket start failed")
		return err
	}
	// single instance
	// only supported - 2024.10.25
	if config.Config.Standalone {
		go func() {
			if err := roots.gRpcStart(); err != nil {
				runtime.Goexit()
			}
		}()
	} else {
		// cluster mode
		return errors.New("not supported mode")
	}
	return nil
}

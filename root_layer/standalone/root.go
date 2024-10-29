package standalone

import (
	"context"
	"os"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/nnv/kv"
	"github.com/sjy-dv/nnv/pkg/hnsw"
	"google.golang.org/grpc"
)

var roots = &RootLayer{}

func NewRootLayer() error {
	log.Info().Msg("wait for rootlayer create..")
	roots = &RootLayer{
		VBucket:     &hnsw.HnswBucket{},
		Bucket:      &kv.DB{},
		S:           &grpc.Server{},
		StreamLayer: &nats.Conn{},
	}
	log.Info().Msg("rootlayer mount vector database")
	err := roots.VBucket.Start(nil)
	if err != nil {
		log.Warn().Err(err).Msg("root_layer.root.go(23) vBucket start failed")
		return err
	}
	log.Info().Msg("rootlayer mount key-value database")
	kvOpts := kv.DefaultOptions
	kvOpts.DirPath = "./data_dir/kv"
	roots.Bucket, err = kv.Open(kvOpts)
	if err != nil {
		log.Warn().Err(err).Msg("root_layer.root.go(27) kv bucket start failed")
		return err
	}
	// single instance
	// only supported - 2024.10.25
	// if config.Config.Standalone {
	if err := gRpcStart(); err != nil {
		log.Warn().Err(err).Msg("root_layer.root.go(38) grpc start failed")
		os.Exit(1)
	}
	// } else {
	// 	// cluster mode
	// 	return errors.New("not supported mode")
	// }
	return nil
}

func StableRelease(ctx context.Context) error {
	if roots.S != nil {
		stopped := make(chan struct{})
		go func() {
			roots.S.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
			log.Debug().Msg("gRPC server shut down successfully")
		case <-ctx.Done():
			roots.S.Stop()
			log.Debug().Msg("gRPC server forced shutdown due to timeout")
		}
	}
	if roots.Bucket != nil {
		log.Debug().Msg("Attempting to close KV store")
		bkStopped := make(chan struct{})
		go func() {
			if err := roots.Bucket.Close(); err != nil {
				log.Warn().Err(err).Msg("kv-store closed failed")
			} else {
				log.Debug().Msg("kv-store closed successfully")
			}
			close(bkStopped)
		}()
		select {
		case <-bkStopped:
		case <-ctx.Done():
			log.Warn().Msg("kv-store close forced shutdown due to timeout")
		}
	}
	if roots.VBucket != nil {
		log.Debug().Msg("Attempting to close vector store")
		vbStopped := make(chan struct{})
		go func() {
			if err := roots.VBucket.Storage.Close(); err != nil {
				log.Warn().Err(err).Msg("vector-store closed failed")
			} else {
				log.Debug().Msg("vector-store closed successfully")
			}
			close(vbStopped)
		}()
		select {
		case <-vbStopped:
		case <-ctx.Done():
			log.Warn().Msg("vector-store close forced shutdown due to timeout")
		}
	}

	return nil
}

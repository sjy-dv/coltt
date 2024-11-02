package standalone

import (
	"context"
	"os"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/nnv/data_access_layer"
	"github.com/sjy-dv/nnv/pkg/hnsw"
	"github.com/sjy-dv/nnv/pkg/index"
	matchdbgo "github.com/sjy-dv/nnv/pkg/match_db.go"
	"github.com/sjy-dv/nnv/pkg/nnlogdb"
	"google.golang.org/grpc"
)

var roots = &RootLayer{}

func NewRootLayer() error {
	log.Info().Msg("wait for rootlayer create..")
	roots = &RootLayer{
		VBucket: &hnsw.HnswBucket{
			Buckets:     make(map[string]*hnsw.Hnsw),
			BucketGroup: make(map[string]bool),
		},
		BitmapIndex: &index.BitmapIndex{},
		S:           &grpc.Server{},
		StreamLayer: &nats.Conn{},
	}
	log.Info().Msg("rootlayer mount match k/v database")
	err := matchdbgo.Open()
	if err != nil {
		log.Warn().Err(err).Msg("root_layer.root.go(280) matchdb open failed")
		return err
	}
	log.Info().Msg("rootlayer mount nnlogdb")
	err = nnlogdb.Open()
	if err != nil {
		log.Warn().Err(err).Msg("root_layer.root.go(36) nnlogdb open failed")
		return err
	}
	log.Info().Msg("rootlayer mount vector database")
	// hnsw bucket loaded
	loadbuckets, bidx, err := data_access_layer.Rollup()
	if err != nil {
		log.Warn().Err(err).Msg("root_layer.root.go(36) vector-data rollup failed")
		return err
	}
	if loadbuckets != nil {
		roots.VBucket = loadbuckets
	}
	if bidx != nil {
		roots.BitmapIndex = bidx
	} else {
		// when first start
		roots.BitmapIndex = index.NewBitmapIndex()
	}
	// err := roots.VBucket.Start(nil)
	// if err != nil {
	// 	log.Warn().Err(err).Msg("root_layer.root.go(23) vBucket start failed")
	// 	return err
	// }
	// log.Info().Msg("rootlayer mount key-value database")
	// kvOpts := kv.DefaultOptions
	// kvOpts.DirPath = "./data_dir/kv"
	// roots.Bucket, err = kv.Open(kvOpts)
	// if err != nil {
	// 	log.Warn().Err(err).Msg("root_layer.root.go(27) kv bucket start failed")
	// 	return err
	// }
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
	log.Debug().Msg("Attempting to close match-database")
	err := matchdbgo.Close()
	if err != nil {
		log.Warn().Err(err).Msg("match-database closed failed")
	}
	if roots.VBucket != nil {
		log.Debug().Msg("Attempting to close vector store")
		vbStopped := make(chan struct{})
		go func() {
			err := data_access_layer.Commit(roots.VBucket, roots.BitmapIndex)
			if err != nil {
				log.Warn().Err(err).Msg("data_access_layer stable saved data failed")
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

package experimentalLayer

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/coltt/experimental"
)

var administrator = &ExperimentalLayer{}

func NewExperimentalLayer() error {
	log.Info().Msg("wait for experimental layer create...")
	administrator = &ExperimentalLayer{
		Engine: &experimental.ExperimentalMultiVector{},
	}

	experimental.NewStateManager()
	log.Info().Msg("experimental.state.manager init")

	layer, err := experimental.NewExperimentalMultiVector()
	if err != nil {
		log.Error().Err(err).Msg("create failed layer")
		return err
	}
	administrator.Engine = layer
	if err := administrator.Engine.LoadAuthorizationBuckets(); err != nil {
		log.Error().Err(err).Msg("cannot find authorization buckets")
		return err
	}
	log.Info().Msg("find authorization bucket, ready for check")

	if err := gRpcStart(); err != nil {
		log.Warn().Err(err).Msg("gRPC server start failed")
		os.Exit(1)
	}
	return nil
}

func StableRelease(ctx context.Context) error {
	if administrator.gRPC != nil {
		stopped := make(chan struct{})
		go func() {
			administrator.gRPC.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
			log.Debug().Msg("gRPC server shut down successfully")
		case <-ctx.Done():
			administrator.gRPC.Stop()
			log.Debug().Msg("gRPC server forced shutdown due to timeout")
		}
	}
	administrator.Engine.Close()
	return nil
}

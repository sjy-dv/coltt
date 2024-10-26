package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	rootlayer "github.com/sjy-dv/nnv/root_layer"
)

func main() {
	log.Info().Msg("rootlayer start")
	go func() {
		err := rootlayer.NewRootLayer()
		if err != nil {
			log.Warn().Err(err).Msg("rootlayer start failed")
			os.Exit(1)
		}
	}()
	// sigChan := make(chan os.Signal, 1)
	// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// go func() {
	// 	<-sigChan
	// 	log.Debug().Msg("received shutdown signal")

	// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// 	defer cancel()

	// 	rootlayer.StableRelease(ctx)

	// 	log.Debug().Msg("shutdown complete")
	// }()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Debug().Msg("received shutdown signal")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := rootlayer.StableRelease(shutdownCtx)
	if err != nil {
		log.Debug().Msgf("info stable release >> %s", err.Error())
	}
	log.Debug().Msg("shutdown complete")
}

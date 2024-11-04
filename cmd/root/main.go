package main

import (
	"context"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/nnv/config"
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
	if config.Config.RootLayer.ProfAddr != "" {
		go func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/debug/pprof/", pprof.Index)
			mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

			err := http.ListenAndServe(config.Config.RootLayer.ProfAddr, mux)
			if err != nil {
				log.Error().Err(err).Msg("profile server crashed!")
			}
		}()
	}
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

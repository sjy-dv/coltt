// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"context"
	"flag"
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

var (
	mode string
)

func main() {
	flag.StringVar(&mode, "mode", "root", "mode select")
	flag.Parse()
	log.Info().Msgf("user select mode : %s", mode)
	log.Info().Msg("setup directory..")
	dirPath := "./data_dir"
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.Mkdir(dirPath, os.ModePerm)
		if err != nil {
			log.Error().Err(err).Msg("create directory failed")
			os.Exit(1)
		}
	}
	log.Info().Msg("setup complete directory")
	log.Info().Msg("rootlayer start")
	go func() {
		err := rootlayer.NewRootLayer(mode)
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
	err := rootlayer.StableRelease(shutdownCtx, mode)
	if err != nil {
		log.Debug().Msgf("info stable release >> %s", err.Error())
	}
	log.Debug().Msg("shutdown complete")
}

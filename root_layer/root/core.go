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

package root

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/coltt/core"
)

var rc = &RootCore{}

func NewRoot() error {
	log.Info().Msg("wait for core-root layer create...")
	rc = &RootCore{
		Core: &core.Core{},
	}
	//-----------------------------------------------//
	core.NewStateManager()
	log.Info().Msg("core-root.stateManager init")
	//-----------------------------------------------//
	err := core.NewIdGenerator()
	if err != nil {
		log.Error().Err(err).Msg("core-root.id-generator init failed")
		return err
	}
	log.Info().Msg("core-root.id-generator init")

	//-----------------------------------------------//
	cr, err := core.NewCore()
	if err != nil {
		log.Error().Err(err).Msg("core-root.disk open failed")
		return err
	}
	rc.Core = cr
	log.Info().Msg("core-root.edge-space init")
	//-----------------------------------------------//
	err = rc.Core.RegistCollectionStManager()
	if err != nil {
		log.Error().Err(err).Msg("core-root.stmanager failed")
		return err
	}
	log.Info().Msg("core-root.stmanager init")
	//-----------------------------------------------//
	core.NewIndexDB()
	log.Info().Msg("core-root.indexdb init")
	if err := gRpcStart(); err != nil {
		log.Warn().Err(err).Msg("core-root.root.go(50) grpc start failed")
		os.Exit(1)
	}

	return nil
}
func StableRelease(ctx context.Context) error {
	if rc.S != nil {
		stopped := make(chan struct{})
		go func() {
			rc.S.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
			log.Debug().Msg("gRPC server shut down successfully")
		case <-ctx.Done():
			rc.S.Stop()
			log.Debug().Msg("gRPC server forced shutdown due to timeout")
		}
	}
	rc.Core.Close()
	return nil
}

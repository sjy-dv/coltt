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

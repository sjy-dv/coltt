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

package edgelite

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/coltt/edge"
)

var edgelites = &EdgeLite{}

func NewEdgeLite() error {
	log.Info().Msg("wait for edge-lite layer create...")
	edgelites = &EdgeLite{
		Edge: &edge.Edge{},
	}
	//-----------------------------------------------//
	edge.NewStateManager()
	log.Info().Msg("edge-lite.stateManager init")
	//-----------------------------------------------//
	err := edge.NewIdGenerator()
	if err != nil {
		log.Error().Err(err).Msg("edge-lite.id-generator init failed")
		return err
	}
	log.Info().Msg("edge-lite.id-generator init")
	//-----------------------------------------------//
	e, err := edge.NewEdge()
	if err != nil {
		log.Error().Err(err).Msg("edge-lite.disk open failed")
		return err
	}
	edgelites.Edge = e
	log.Info().Msg("edge-lite.edge-space init")
	//-----------------------------------------------//
	err = edgelites.Edge.LoadAuthorizationBuckets()
	if err != nil {
		log.Error().Err(err).Msg("edge-lit.authorization bucket load failed")
		return err
	}
	log.Info().Msg("edge-lit.authorization bucket load")

	if err := gRpcStart(); err != nil {
		log.Warn().Err(err).Msg("edge-lite.root.go(50) grpc start failed")
		os.Exit(1)
	}

	return nil
}

func StableRelease(ctx context.Context) error {
	if edgelites.S != nil {
		stopped := make(chan struct{})
		go func() {
			edgelites.S.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
			log.Debug().Msg("gRPC server shut down successfully")
		case <-ctx.Done():
			edgelites.S.Stop()
			log.Debug().Msg("gRPC server forced shutdown due to timeout")
		}
	}
	edgelites.Edge.Close()
	return nil
}

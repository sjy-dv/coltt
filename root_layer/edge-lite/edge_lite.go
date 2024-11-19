package edgelite

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/nnv/edge"
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
	err = edgelites.Edge.LoadCommitCollection()
	if err != nil {
		log.Error().Err(err).Msg("edge-lite.loadcommitcollection failed")
		return err
	}
	log.Info().Msg("edge-lite.loadcommitcollection init")
	//-----------------------------------------------//
	e, err := edge.NewEdge()
	if err != nil {
		log.Error().Err(err).Msg("edge-lite.disk open failed")
		return err
	}
	edgelites.Edge = e
	log.Info().Msg("edge-lite.edge-space init")
	//-----------------------------------------------//
	edge.NewIndexDB()
	log.Info().Msg("edge-lite.indexdb init")
	//-----------------------------------------------//
	edge.NewEdgeVectorCollection()
	log.Info().Msg("edge-lite.normalize-vector-store init")
	//-----------------------------------------------//
	edge.NewQuantizedEdgeVectorCollection()
	log.Info().Msg("edge-lite.quantized-vector-store init")
	//-----------------------------------------------//
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
	return nil
}

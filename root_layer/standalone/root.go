package standalone

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/nnv/highmem"
)

var roots = &RootLayer{}

func NewRootLayer() error {
	log.Info().Msg("wait for rootlayer create..")
	roots = &RootLayer{
		HighMem: &highmem.HighMem{},
	}
	//-----------------------------------------------//
	highmem.NewStateManager()
	log.Info().Msg("root_layer.stateManager init")
	//-----------------------------------------------//
	err := highmem.NewIdGenerator()
	if err != nil {
		log.Error().Err(err).Msg("root_layer.id-generator init failed")
		return err
	}
	log.Info().Msg("root_layer.id-generator init")
	//-----------------------------------------------//
	err = highmem.StartCommitLogger()
	if err != nil {
		log.Error().Err(err).Msg("root_layer.startcommitlogger init failed")
	}
	log.Info().Msg("root_layer.startcommitlogger init")
	//-----------------------------------------------//
	go highmem.BackLogging()
	log.Info().Msg("root_layer.backlogging init")
	//-----------------------------------------------//
	err = roots.HighMem.LoadCommitCollection()
	if err != nil {
		log.Error().Err(err).Msg("root_layer.loadcommitcollection failed")
		return err
	}
	log.Info().Msg("root_layer.loadcommitcollection init")
	//-----------------------------------------------//
	hmem := highmem.NewHighMemory()
	roots.HighMem = hmem
	log.Info().Msg("root_layer.highmemory init")
	//-----------------------------------------------//
	highmem.NewIndexDB()
	log.Info().Msg("root_layer.indexdb init")
	//-----------------------------------------------//
	highmem.NewTensorLink()
	log.Info().Msg("root_layer.tensorlink init")
	//-----------------------------------------------//
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

/*
 */
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

	//-----------------------------------------------//
	log.Debug().Msg("Attempting to close HighMemory")

	roots.HighMem.CommitAll()
	//-----------------------------------------------//

	err := highmem.BorrowCommitLogger().Close()
	if err != nil {
		log.Warn().Err(err).Msg("commit-logger closed failed")
	}
	log.Info().Msg("commit-logger closed success")
	//-----------------------------------------------//

	// log.Debug().Msg("Attempting to close match-database")
	// err := matchdbgo.Close()
	// if err != nil {
	// 	log.Warn().Err(err).Msg("match-database closed failed")
	// }
	// if roots.VBucket != nil {
	// 	log.Debug().Msg("Attempting to close vector store")
	// 	vbStopped := make(chan struct{})
	// 	go func() {
	// 		err := data_access_layer.Commit(roots.VBucket, roots.BitmapIndex)
	// 		if err != nil {
	// 			log.Warn().Err(err).Msg("data_access_layer stable saved data failed")
	// 		}
	// 		close(vbStopped)
	// 	}()
	// 	select {
	// 	case <-vbStopped:
	// 	case <-ctx.Done():
	// 		log.Warn().Msg("vector-store close forced shutdown due to timeout")
	// 	}
	// }

	return nil
}

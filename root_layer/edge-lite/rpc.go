package edgelite

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	rpclog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/rs/zerolog/log"
	"github.com/sjy-dv/nnv/config"
	"github.com/sjy-dv/nnv/gen/protoc/v2/edgeproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

func gRpcStart() error {
	logger := rpclog.NewLogfmtLogger(os.Stderr)
	rpcLogger := rpclog.With(logger, "service", "gRPC/server", "component", "nnv-rpc")
	lis, err := net.Listen("tcp", config.Config.RootLayer.BindAddress)
	if err != nil {
		return err
	}
	rpcOpts := []grpc.ServerOption{}
	if config.Config.RootLayer.KeepAliveTime == 0 {
		log.Warn().Msg("keep-alive cannot reach 0 seconds. It is set to the default of 60 seconds.")
		config.Config.RootLayer.KeepAliveTime = 60
	}
	if config.Config.RootLayer.KeepAliveTimeOut == 0 {
		log.Warn().Msg("keep-alive-timeout cannot reach 0 seconds. It is set to the default of 10 seconds.")
		config.Config.RootLayer.KeepAliveTimeOut = 10
	}
	rpcOpts = append(rpcOpts, grpc.KeepaliveParams(keepalive.ServerParameters{
		Time:    time.Second * time.Duration(config.Config.RootLayer.KeepAliveTime),
		Timeout: time.Second * time.Duration(config.Config.RootLayer.KeepAliveTimeOut),
	}))
	log.Debug().Msg("The keep-alive and keep-alive-timeout values should be the same for ingress and client.")

	if config.Config.RootLayer.MaxRecvMsgSize == 0 {
		log.Warn().Msg("MaxRecvMsgSize cannot reach 0. Is is set to the default of 10MB")
		config.Config.RootLayer.MaxRecvMsgSize = 10 * MB
	}
	if config.Config.RootLayer.MaxSendMsgSize == 0 {
		log.Warn().Msg("MaxSendMsgSize cannot reach 0. Is is set to the default of 10MB")
		config.Config.RootLayer.MaxSendMsgSize = 10 * MB
	}
	rpcOpts = append(rpcOpts, []grpc.ServerOption{
		grpc.MaxRecvMsgSize(config.Config.RootLayer.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(config.Config.RootLayer.MaxSendMsgSize),
	}...)
	log.Debug().Msg("max-recv & max-send msg size needs to be synchronized with clients")

	if config.Config.RootLayer.EnforcementPolicyMinTime == 0 {
		log.Warn().Msg("keep-alive-enforcement-policy cannot reach 0 seconds. It is set to the default of 5 seconds.")
		config.Config.RootLayer.EnforcementPolicyMinTime = 5
	}
	rpcOpts = append(rpcOpts, grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
		MinTime:             time.Second * time.Duration(config.Config.RootLayer.EnforcementPolicyMinTime),
		PermitWithoutStream: true,
	}))
	log.Debug().Msg("Be careful not to conflict with the client settings. Incorrect configuration can lead to the error [transport] Client received GoAway with error code ENHANCE_YOUR_CALM and debug data equal to ASCII 'too_many_pings'")
	if config.Config.RootLayer.PemFile != "" && config.Config.RootLayer.KeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(
			config.Config.RootLayer.PemFile,
			config.Config.RootLayer.KeyFile,
		)
		if err != nil {
			log.Warn().Err(err).Msg("tls configured error")
			return err
		}
		rpcOpts = append(rpcOpts, grpc.Creds(creds))
	}
	rpcOpts = append(rpcOpts, grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(rpcPanicHandler)),
		logging.UnaryServerInterceptor(interceptorLogger(rpcLogger)),
	), grpc.ChainStreamInterceptor(
		recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(rpcPanicHandler)),
		logging.StreamServerInterceptor(interceptorLogger(rpcLogger)),
	))
	edgelites.S = grpc.NewServer(rpcOpts...)
	rpcLayer := rpcLayer{}
	rpcLayer.Eproto = &edgeProtoConn{rpcLayer: rpcLayer}
	//rpcLayer.rootClone = xx
	edgeproto.RegisterEdgeRpcServer(edgelites.S, rpcLayer.Eproto)
	log.Debug().Msgf("grpc_startup bind_addr : %s", config.Config.RootLayer.BindAddress)
	if err := edgelites.S.Serve(lis); err != nil {
		log.Warn().Err(err).Msg("grpc_startup failed")
		return err
	}
	return nil
}

func rpcPanicHandler(p any) (err error) {
	return status.Errorf(codes.Internal, "%s", p)
}

func interceptorLogger(l rpclog.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		largs := append([]any{"msg", msg}, fields...)
		switch lvl {
		case logging.LevelDebug:
			_ = level.Debug(l).Log(largs...)
		case logging.LevelInfo:
			_ = level.Info(l).Log(largs...)
		case logging.LevelWarn:
			_ = level.Warn(l).Log(largs...)
		case logging.LevelError:
			_ = level.Error(l).Log(largs...)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

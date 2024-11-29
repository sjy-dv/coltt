package root

import (
	"context"
	"time"

	"github.com/sjy-dv/nnv/core"
	"google.golang.org/grpc"
)

type RootCore struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	Core   *core.Core
	S      *grpc.Server
}

type rpcLayer struct {
	Cproto *coreProtoConn
}

type coreProtoConn struct {
	rpcLayer
}

const (
	B  = 1
	KB = 1024
	MB = 1024 * 1024
	GB = 1024 * 1024 * 1024
)

const DefaultMsgSize = 104858000 // 10mb
const DefaultKeepAliveTimeout = 10 * time.Second
const DefaultKeepAlive = 60 * time.Second
const DefaultEnforcementPolicyMinTime = 5 * time.Second

var UncaughtPanicError = "uncaught panic error: %v"

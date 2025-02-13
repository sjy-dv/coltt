package edgelite

import (
	"context"
	"time"

	"github.com/sjy-dv/coltt/edge"
	"google.golang.org/grpc"
)

type EdgeLite struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	Edge   *edge.Edge
	S      *grpc.Server
}

type rpcLayer struct {
	Eproto *edgeProtoConn
}

type edgeProtoConn struct {
	rpcLayer
}

const (
	b  = 1
	kb = 1024
	mb = 1024 * 1024
	gb = 1024 * 1024 * 1024

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

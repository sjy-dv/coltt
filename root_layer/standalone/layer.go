package standalone

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sjy-dv/nnv/gen/protoc/v1/dataCoordinatorV1"
	"github.com/sjy-dv/nnv/gen/protoc/v1/resourceCoordinatorV1"
	"github.com/sjy-dv/nnv/pkg/hnsw"
	"google.golang.org/grpc"
)

type RootLayer struct {
	Ctx    context.Context
	Cancel context.CancelFunc

	StreamLayer    *nats.Conn
	StreamLayerCtx nats.JetStreamContext

	VBucket *hnsw.HnswBucket // vector store
	S       *grpc.Server
}

type rpcLayer struct {
	X1 *datasetCoordinator
	X2 *resourceCoordinator
	// rootClone *RootLayer
}

type datasetCoordinator struct {
	dataCoordinatorV1.UnimplementedDatasetCoordinatorServer
	rpcLayer
}

type resourceCoordinator struct {
	resourceCoordinatorV1.UnimplementedResourceCoordinatorServer
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

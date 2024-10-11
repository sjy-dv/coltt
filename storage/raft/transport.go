package raft

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/sjy-dv/vemoo/cluster"
	"github.com/sjy-dv/vemoo/proto/gen/v1/raftV1"
)

var (
	ErrGroupAlreadyExists error = errors.New("Group already exists")
	ErrGroupNotFound      error = errors.New("Group not found")
)

type RaftTransport struct {
	nodeId      uint64
	addr        string
	clusterConn *cluster.Conn

	nodeClients   map[uint64]raftV1.RaftTransportClient
	nodeClientsMu sync.RWMutex
	groups        map[uuid.UUID]*RaftGroup
	groupsMu      sync.RWMutex
}

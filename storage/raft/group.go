package raft

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/sjy-dv/vemoo/storage/wal"
	etcdRaft "go.etcd.io/raft/v3"
	etcdRaftpb "go.etcd.io/raft/v3/raftpb"
)

const snapshotOffset uint64 = 5000

var (
	ErrProcessFnAlreadyRegist  error = errors.New("ProcessFn already registerd")
	ErrShapshotFnAlreadyRegist error = errors.New("SnapshotFn already registerd")
)

type ProcessFn func([]byte) error
type SnapshotFn func() ([]byte, error)

type RaftGroup struct {
	id                uuid.UUID
	transport         *RaftTransport
	ctx               context.Context
	ctxCancel         context.CancelFunc
	processFn         ProcessFn
	processSnapshotFn ProcessFn
	snapshotFn        SnapshotFn

	raft          etcdRaft.Node
	raftConfState *etcdRaftpb.ConfState
	raftLeaderId  uint64

	wal wal.WAL
	log *logrus.Entry
}

func startRaftNode()

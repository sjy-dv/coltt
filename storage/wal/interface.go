package wal

import (
	etcdRaft "go.etcd.io/raft/v3"
	etcdRaftpb "go.etcd.io/raft/v3/raftpb"
)

type WAL interface {
	etcdRaft.Storage
	Save(etcdRaftpb.HardState, []etcdRaftpb.Entry, etcdRaftpb.Snapshot) error
	CreateSnapshot(uint64, *etcdRaftpb.ConfState, []byte) (etcdRaftpb.Snapshot, error)
	DeleteGroup() error
}

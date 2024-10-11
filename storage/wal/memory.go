package wal

import (
	etcdRaft "go.etcd.io/raft/v3"
	etcdRaftpb "go.etcd.io/raft/v3/raftpb"
)

type memoryWAL struct {
	*etcdRaft.MemoryStorage
}

func NewMemoryWAL() *memoryWAL {
	return &memoryWAL{MemoryStorage: etcdRaft.NewMemoryStorage()}
}

func (this *memoryWAL) Save(
	hs etcdRaftpb.HardState,
	es []etcdRaftpb.Entry,
	snap etcdRaftpb.Snapshot) error {
	if err := this.Append(es); err != nil {
		return err
	}
	if err := this.SetHardState(hs); err != nil {
		return err
	}
	if !etcdRaft.IsEmptySnap(snap) {
		if err := this.ApplySnapshot(snap); err != nil {
			return err
		}
	}
	return nil
}

func (this *memoryWAL) DeleteGroup() error {
	this = &memoryWAL{MemoryStorage: etcdRaft.NewMemoryStorage()}
	return nil
}

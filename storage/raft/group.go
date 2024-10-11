package raft

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

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

func startRaftNode(id uint64, nodeIds []uint64, storage wal.WAL, logger *logrus.Entry) (etcdRaft.Node, error) {
	raftCfg := &etcdRaft.Config{
		ID:              id,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
		Logger:          logger,
	}

	if len(nodeIds) > 0 {
		var peers []etcdRaft.Peer
		for _, nodeId := range nodeIds {
			peers = append(peers, etcdRaft.Peer{ID: nodeId})
		}
		return etcdRaft.StartNode(raftCfg, peers), nil
	} else {
		return etcdRaft.RestartNode(raftCfg), nil
	}
}

func NewRaftGroup(id uuid.UUID, nodeIds []uint64, storage wal.WAL, transport *RaftTransport) (*RaftGroup, error) {
	logger := logrus.WithFields(logrus.Fields{
		"node_id":  fmt.Sprintf("%16x", transport.nodeId),
		"group_id": id.String(),
	})

	ctx, cancel := context.WithCancel(context.Background())
	raftNode, err := startRaftNode(transport.NodeId(), nodeIds, storage, logger)
	if err != nil {
		return nil, err
	}

	g := &RaftGroup{
		id:                id,
		transport:         transport,
		ctx:               ctx,
		ctxCancel:         cancel,
		processFn:         nil,
		processSnapshotFn: nil,
		snapshotFn:        nil,
		raft:              raftNode,
		wal:               storage,
		log:               logger,
	}

	if err := transport.addGroup(g); err != nil {
		return nil, err
	}
	return g, nil
}

func (instance *RaftGroup) Start() error {
	snap, err := instance.wal.Snapshot()
	if err != nil {
		return err
	}
	if !etcdRaft.IsEmptySnap(snap) {
		if err := instance.processSnapshotFn(snap.Data); err != nil {
			return err
		}
	}

	return nil
}

func (instance *RaftGroup) Stop() {
	instance.raft.Stop()
	instance.ctxCancel()

	if err := instance.transport.removeGroup(instance.id); err != nil {
		instance.log.Error(err)
	}
}

func (instance *RaftGroup) RegisterProcessFn(fn ProcessFn) error {
	if instance.processFn != nil {
		return ErrProcessFnAlreadyRegist
	}
	instance.processFn = fn
	return nil
}

func (instance *RaftGroup) RegisterProcessSnapshotFn(fn ProcessFn) error {
	if instance.processSnapshotFn != nil {
		return ErrProcessFnAlreadyRegist
	}
	instance.processSnapshotFn = fn
	return nil
}

func (instance *RaftGroup) RegisterSnapshotFn(fn SnapshotFn) error {
	if instance.snapshotFn != nil {
		return ErrShapshotFnAlreadyRegist
	}
	instance.snapshotFn = fn
	return nil
}

func (instance *RaftGroup) LeaderId() uint64 {
	return instance.raftLeaderId
}

func (instance *RaftGroup) Propose(ctx context.Context, data []byte) error {
	return instance.raft.Propose(ctx, data)
}

func (instance *RaftGroup) ProposeJoin(nodeId uint64, address string) error {
	var cc etcdRaftpb.ConfChange
	cc.Type = etcdRaftpb.ConfChangeAddNode
	cc.NodeID = nodeId
	cc.Context = []byte(address)

	return instance.raft.ProposeConfChange(instance.ctx, cc)
}

func (instance *RaftGroup) ProposeLeave(nodeId uint64) error {
	var cc etcdRaftpb.ConfChange
	cc.Type = etcdRaftpb.ConfChangeRemoveNode
	cc.NodeID = nodeId

	return instance.raft.ProposeConfChange(instance.ctx, cc)
}

func (instance *RaftGroup) trySnapshot(lastCommittedIdx, skip uint64) error {
	if instance.snapshotFn == nil {
		return nil
	}

	existing, err := instance.wal.Snapshot()
	if err != nil {
		return err
	}
	if lastCommittedIdx <= existing.Metadata.Index+skip {
		// Not enough new log entries to create snapshot
		return nil
	}

	startAt := time.Now()
	snapshotData, err := instance.snapshotFn()
	if err != nil {
		return err
	}

	_, err = instance.wal.CreateSnapshot(lastCommittedIdx, instance.raftConfState, snapshotData)
	if err == nil {
		instance.log.WithFields(logrus.Fields{
			"action": "raft.group.(fn).try.snapshot.CreateSnapshot()",
		}).Infof("Complete Snapshot => SnapShot.Size: %d bytes. Duration: %s", len(snapshotData), time.Since(startAt))

	}
	return err
}

func (instance *RaftGroup) bootstrap() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	snapshotTicker := time.NewTicker(10 * time.Second)
	defer snapshotTicker.Stop()

	var lastAppliedIdx uint64 = 0
	for {
		select {
		case <-snapshotTicker.C:
			if err := instance.trySnapshot(lastAppliedIdx, snapshotOffset); err != nil {
				instance.log.WithFields(logrus.Fields{
					"action": "snapshotTicker.trySnapshot",
				}).Error(err.Error())
				continue
			}
		case <-ticker.C:
			instance.raft.Tick()
		case ready := <-instance.raft.Ready():
			if ready.SoftState != nil {
				instance.raftLeaderId = atomic.LoadUint64(&ready.SoftState.Lead)
			}
			if instance.isLeader() {
				instance.transport.Send(instance.ctx, instance, ready.Messages)
			}
			instance.raft.Advance()
		case <-instance.ctx.Done():
			instance.log.WithFields(logrus.Fields{
				"action": "bootstrap.raft.group.ctx.done",
			}).Info("raft stopping..")
			return
		}
	}
}

func (instance *RaftGroup) isLeader() bool {
	return instance.LeaderId() == instance.transport.NodeId()
}

func (instance *RaftGroup) receive(message etcdRaftpb.Message) error {
	return instance.raft.Step(instance.ctx, message)
}

func (instance *RaftGroup) reportUnreachable(nodeId uint64) {
	instance.raft.ReportUnreachable(nodeId)
}

func (instance *RaftGroup) reportSnapshot(nodeId uint64, status etcdRaft.SnapshotStatus) {
	instance.raft.ReportSnapshot(nodeId, status)
}

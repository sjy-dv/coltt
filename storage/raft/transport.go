package raft

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/sjy-dv/vemoo/cluster"
	"github.com/sjy-dv/vemoo/proto/gen/v1/raftV1"
	etcdRaft "go.etcd.io/raft/v3"
	etcdRaftpb "go.etcd.io/raft/v3/raftpb"
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

func (rt *RaftTransport) NodeId() uint64 {
	return rt.nodeId
}

func (rt *RaftTransport) GetAddr() string {
	return rt.addr
}

func (rt *RaftTransport) addGroup(group *RaftGroup) error {
	rt.groupsMu.Lock()
	defer rt.groupsMu.Unlock()

	if _, exists := rt.groups[group.id]; exists {
		return ErrGroupAlreadyExists
	}

	rt.groups[group.id] = group
	return nil
}

func (rt *RaftTransport) Send(ctx context.Context, group *RaftGroup, messages []etcdRaftpb.Message) {
	for _, m := range messages {
		mBytes, err := m.Marshal()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"action": "raft.transport.send",
			}).Error(err.Error())
			continue
		}

		client, err := rt.getNodeRaftTransportClient(m.To)
		if err != nil {
			group.reportUnreachable(m.To)
			if m.Type == etcdRaftpb.MsgSnap {
				group.reportSnapshot(m.To, etcdRaft.SnapshotFailure)
			}
			continue
		}

		payload := &raftV1.RaftLog{
			GroupId: group.id[:],
			Message: mBytes,
		}

		sendCtx, _ := context.WithTimeout(ctx, 500*time.Millisecond)
		if _, err := client.Receive(sendCtx, payload); err != nil {
			group.reportUnreachable(m.To)
			if m.Type == etcdRaftpb.MsgSnap {
				group.reportSnapshot(m.To, etcdRaft.SnapshotFailure)
			}
			continue
		}

		if m.Type == etcdRaftpb.MsgSnap {
			group.reportSnapshot(m.To, etcdRaft.SnapshotFinish)
		}
	}
}

func (rt *RaftTransport) addNodeAddress(nodeId uint64, address string) {
	rt.clusterConn.ProvisioningNode(nodeId, address)
}

func (rt *RaftTransport) removeNodeAddress(nodeId uint64) {
	rt.clusterConn.DeProvisioningNode(nodeId)
}

func (rt *RaftTransport) removeGroup(id uuid.UUID) error {
	rt.groupsMu.Lock()
	defer rt.groupsMu.Unlock()

	if _, exists := rt.groups[id]; !exists {
		return ErrGroupNotFound
	}

	delete(rt.groups, id)
	return nil
}

func (rt *RaftTransport) getGroup(id uuid.UUID) (*RaftGroup, error) {
	rt.groupsMu.RLock()
	defer rt.groupsMu.RUnlock()

	group, exists := rt.groups[id]
	if !exists {
		return nil, ErrGroupNotFound
	}
	return group, nil
}

func (rt *RaftTransport) getNodeRaftTransportClient(nodeId uint64) (raftV1.RaftTransportClient, error) {
	rt.nodeClientsMu.RLock()
	if client, exists := rt.nodeClients[nodeId]; exists {
		rt.nodeClientsMu.RUnlock()
		return client, nil
	}
	rt.nodeClientsMu.RUnlock()

	conn, err := rt.clusterConn.NewDial(nodeId)
	if err != nil {
		return nil, err
	}

	rt.nodeClientsMu.Lock()
	defer rt.nodeClientsMu.Unlock()

	rt.nodeClients[nodeId] = raftV1.NewRaftTransportClient(conn)
	return rt.nodeClients[nodeId], nil
}

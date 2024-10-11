package raft

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
	"github.com/sjy-dv/vemoo/cluster"
	"github.com/sjy-dv/vemoo/proto/gen/v1/clusterV1"
)

type NodesCoordinator struct {
	cluster   *cluster.Conn
	zeroGroup *RaftGroup
}

func NewNodesCoordinator(conn *cluster.Conn, zeroGroup *RaftGroup) *NodesCoordinator {
	return &NodesCoordinator{
		cluster:   conn,
		zeroGroup: zeroGroup,
	}
}

func (coordinator *NodesCoordinator) Join(ctx context.Context, addrs []string) error {
	for idx, addr := range addrs {
		err := coordinator.tryJoin(ctx, addr)
		if err != nil {
			if idx < len(addrs)-1 {
				logrus.WithFields(logrus.Fields{
					"action": "NodesCoordinator.Join.Attempt",
				}).Errorf("Failed Join : %v", err)
				continue
			} else {
				return errors.New(fmt.Sprintf("NodesCoordinator.Join.Cluster.Failed.Error >> %v", err))
			}
		}
		break
	}
	return nil
}

func (coordinator *NodesCoordinator) ListNodes() map[uint64]string {
	return coordinator.cluster.GetNodes()
}

func (coordinator *NodesCoordinator) AddNode(id uint64, addr string) (map[uint64]string, error) {
	if err := coordinator.zeroGroup.ProposeJoin(id, addr); err != nil {
		return nil, err
	}
	nodes := coordinator.cluster.GetNodes()
	nodes[id] = addr
	return nodes, nil
}

func (coordinator *NodesCoordinator) RemoveNode(id uint64) error {
	return coordinator.zeroGroup.ProposeLeave(id)
}

func (coordinator *NodesCoordinator) tryJoin(ctx context.Context, address string) error {
	conn, err := coordinator.cluster.DialAddrs(address)
	if err != nil {
		return err
	}
	defer conn.Close()

	nodesStream, err := clusterV1.NewNodesCoordinatorClient(conn).AddNode(ctx, &clusterV1.Node{
		Id:      coordinator.cluster.GetId(),
		Address: coordinator.cluster.GetAddr(),
	})
	if err != nil {
		return err
	}

	for {
		node, err := nodesStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		coordinator.cluster.ProvisioningNode(node.GetId(), node.GetAddress())
	}
	return nil
}

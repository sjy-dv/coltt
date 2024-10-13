// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package services

import (
	"context"

	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/sjy-dv/vemoo/proto/gen/v1/clusterV1"
	"github.com/sjy-dv/vemoo/proto/gen/v1/coreV1"
	"github.com/sjy-dv/vemoo/storage/raft"
)

type nodesCoordinatorServer struct {
	nodesCoordinator *raft.NodesCoordinator
}

func NewNodesCoordinatorServer(nodesCoordinator *raft.NodesCoordinator) *nodesCoordinatorServer {
	return &nodesCoordinatorServer{
		nodesCoordinator: nodesCoordinator,
	}
}

func (this *nodesCoordinatorServer) ListNodes(req *coreV1.EmptyMessage,
	stream clusterV1.NodesCoordinator_ListNodesServer) error {
	for id, addr := range this.nodesCoordinator.ListNodes() {
		if err := stream.Send(&clusterV1.Node{
			Id:      id,
			Address: addr,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (this *nodesCoordinatorServer) AddNode(req *clusterV1.Node,
	stream clusterV1.NodesCoordinator_AddNodeServer) error {
	nodes, err := this.nodesCoordinator.AddNode(req.GetId(), req.GetAddress())
	if err != nil {
		return err
	}
	for nodeId, addr := range nodes {
		if err := stream.Send(&clusterV1.Node{
			Id:      nodeId,
			Address: addr,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (this *nodesCoordinatorServer) RemoveNode(ctx context.Context, node *clusterV1.Node) (*coreV1.EmptyMessage, error) {
	if err := this.nodesCoordinator.RemoveNode(node.GetId()); err != nil {
		return nil, err
	}
	return &coreV1.EmptyMessage{}, nil
}

func (this *nodesCoordinatorServer) LoadMetrics(ctx context.Context, req *coreV1.EmptyMessage) (*clusterV1.NodeMetrics, error) {
	hostStat, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, err
	}
	vMemStat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, err
	}
	avgStat, err := load.AvgWithContext(ctx)
	if err != nil {
		return nil, err
	}

	return &clusterV1.NodeMetrics{
		Uptime:         hostStat.Uptime,
		CpuLoad1:       avgStat.Load1,
		CpuLoad5:       avgStat.Load5,
		CpuLoad15:      avgStat.Load15,
		MemTotal:       vMemStat.Total,
		MemAvailable:   vMemStat.Available,
		MemUsed:        vMemStat.Used,
		MemFree:        vMemStat.Free,
		MemUsedPercent: vMemStat.UsedPercent,
	}, nil
}

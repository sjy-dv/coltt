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

package cluster

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrNodeAddrNotFound = "Node [%16x] Addr not found."
)

type nodesProvisionType uint

const (
	NewNodesProvisioning nodesProvisionType = iota
	NodesDeProvisioning
)

type nodesProvision struct {
	Type   nodesProvisionType
	NodeId uint64
}

type Conn struct {
	id              uint64
	addr            string
	nodesAddrs      map[uint64]string
	nodesAddrsMu    sync.RWMutex
	nodesConns      map[uint64]*grpc.ClientConn
	nodesConnsMu    sync.RWMutex
	notifications   []chan *nodesProvision
	notificationsMu *sync.RWMutex

	transportCredentials credentials.TransportCredentials

	log *logrus.Entry
}

func NewConn(id uint64, addr string, tlsCert string) (*Conn, error) {
	c := &Conn{
		id:              id,
		addr:            addr,
		nodesAddrs:      make(map[uint64]string),
		nodesAddrsMu:    sync.RWMutex{},
		nodesConns:      make(map[uint64]*grpc.ClientConn),
		nodesConnsMu:    sync.RWMutex{},
		notifications:   make([]chan *nodesProvision, 0),
		notificationsMu: &sync.RWMutex{},
		log: logrus.WithFields(logrus.Fields{
			"node_id": fmt.Sprintf("%16x", id),
		}),
	}

	if tlsCert != "" {
		var err error
		c.transportCredentials, err = credentials.NewClientTLSFromFile(tlsCert, "devJ-Cluster")
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (this *Conn) GetId() uint64 {
	return this.id
}

func (this *Conn) GetAddr() string {
	return this.addr
}

func (this *Conn) Close() {
	this.nodesConnsMu.Lock()
	defer this.nodesAddrsMu.Unlock()

	for _, conn := range this.nodesConns {
		if err := conn.Close(); err != nil {
			logrus.WithFields(logrus.Fields{
				"action": "Error: node connection closing",
			}).Error(err.Error())
		}
	}

	this.notificationsMu.Lock()
	defer this.notificationsMu.Unlock()
	for _, noti := range this.notifications {
		close(noti)
	}
}

func (this *Conn) NodeStatusNotifications() <-chan *nodesProvision {
	this.notificationsMu.Lock()
	defer this.notificationsMu.Unlock()

	c := make(chan *nodesProvision, 10)
	this.notifications = append(this.notifications, c)
	return c
}

func (this *Conn) GetNodes() map[uint64]string {
	this.nodesAddrsMu.RLock()
	defer this.nodesAddrsMu.RUnlock()

	nodes := make(map[uint64]string)
	for id, addr := range this.nodesAddrs {
		nodes[id] = addr
	}
	return nodes
}

func (this *Conn) GetNodeIds() []uint64 {
	this.nodesAddrsMu.RLock()
	defer this.nodesAddrsMu.RUnlock()

	ids := make([]uint64, 0, len(this.nodesAddrs))
	for id := range this.nodesAddrs {
		ids = append(ids, id)
	}
	return ids
}

func (this *Conn) ProvisioningNode(id uint64, addr string) {
	this.nodesAddrsMu.Lock()
	defer this.nodesAddrsMu.Unlock()

	if _, exists := this.nodesAddrs[id]; !exists {
		this.nodesAddrs[id] = addr
		this.sendNodesUpdateNotify(&nodesProvision{
			Type:   NewNodesProvisioning,
			NodeId: id,
		})
		this.log.WithFields(logrus.Fields{
			"action": "new node provisioning!",
		}).Infof("Provisioning Node-%16x", id)
	}
}

func (this *Conn) DeProvisioningNode(id uint64) {
	this.nodesAddrsMu.Lock()
	defer this.nodesAddrsMu.Unlock()
	this.nodesConnsMu.Lock()
	defer this.nodesConnsMu.Unlock()

	if _, exists := this.nodesAddrs[id]; exists {
		delete(this.nodesAddrs, id)
		if conn, exists := this.nodesConns[id]; exists {
			if err := conn.Close(); err != nil {
				this.log.WithFields(logrus.Fields{
					"action": "grpc connection closing error!",
				}).Error(err.Error())
			}
			delete(this.nodesConns, id)
		}
		this.sendNodesUpdateNotify(&nodesProvision{
			Type:   NodesDeProvisioning,
			NodeId: id,
		})
		this.log.WithFields(logrus.Fields{
			"action": "deprovisioning node",
		}).Infof("Deprovisioning Node-%16x", id)
	}
}

func (this *Conn) NewDial(id uint64) (*grpc.ClientConn, error) {
	conn := this.loadCacheConn(id)
	if conn != nil {
		return conn, nil
	}
	addr, err := this.findAddr(id)
	if err != nil {
		return nil, err
	}
	conn, err = grpc.NewClient(addr, this.grpcDefaultDialOpts()...)
	if err != nil {
		return nil, err
	}

	this.nodesConnsMu.Lock()
	defer this.nodesConnsMu.Unlock()
	if existsConn, exists := this.nodesConns[id]; exists {
		conn.Close()
		return existsConn, nil
	}
	this.nodesConns[id] = conn
	return conn, nil
}

func (this *Conn) DialAddrs(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, this.grpcDefaultDialOpts()...)
}

func (this *Conn) loadCacheConn(id uint64) *grpc.ClientConn {
	this.nodesConnsMu.RLock()
	defer this.nodesConnsMu.RUnlock()
	if conn, exists := this.nodesConns[id]; exists {
		return conn
	}
	return nil
}

func (this *Conn) findAddr(id uint64) (string, error) {
	this.nodesAddrsMu.RLock()
	defer this.nodesAddrsMu.RUnlock()

	if addr, exists := this.nodesAddrs[id]; exists {
		return addr, nil
	}
	return "", fmt.Errorf(ErrNodeAddrNotFound, id)
}

func (this *Conn) grpcDefaultDialOpts() []grpc.DialOption {
	opts := make([]grpc.DialOption, 0)
	if this.transportCredentials != nil {
		opts = append(opts, grpc.WithTransportCredentials(this.transportCredentials))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	return opts
}

func (this *Conn) sendNodesUpdateNotify(n *nodesProvision) {
	this.notificationsMu.RLock()
	defer this.notificationsMu.RUnlock()

	for _, c := range this.notifications {
		c <- n
	}
}

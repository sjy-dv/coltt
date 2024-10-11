package raft

import (
	"context"

	"github.com/sjy-dv/vemoo/proto/gen/v1/raftV1"
)

type sharedGroup struct {
	group   *RaftGroup
	proxies map[string]*sharedGroupProxy
}

func NewSharedGroup(group *RaftGroup) (*sharedGroup, error) {
	sg := &sharedGroup{
		group:   group,
		proxies: make(map[string]*sharedGroupProxy),
	}

	if err := group.RegisterProcessFn(sg.process); err != nil {
		return nil, err
	}
	if err := group.RegisterProcessSnapshotFn(sg.processSnapshot); err != nil {
		return nil, err
	}
	if err := group.RegisterSnapshotFn(sg.snapshot); err != nil {
		return nil, err
	}
	return sg, nil
}

func (instance sharedGroup) Get(name string) *sharedGroupProxy {
	if proxy, exists := instance.proxies[name]; exists {
		return proxy
	}

	instance.proxies[name] = newSharedGroupProxy(instance.group, name)
	return instance.proxies[name]
}

type sharedGroupProxy struct {
	name  string
	group *RaftGroup

	processFn         ProcessFn
	processSnapshotFn ProcessFn
	snapshotFn        SnapshotFn
}

func (instance *sharedGroup) process(data []byte) error {
	proposal := new(raftV1.GroupProposalAction)
	if err := proposal.Unmarshal(data); err != nil {
		return err
	}

	if proxy, exists := instance.proxies[proposal.GetProxyName()]; exists {
		return proxy.processFn(proposal.GetData())
	}
	return nil
}

func (instance *sharedGroup) processSnapshot(data []byte) error {
	snapshot := new(raftV1.GroupStateSnapshot)
	if err := snapshot.Unmarshal(data); err != nil {
		return err
	}

	for proxyName, proxySnapshot := range snapshot.GetProxySnapshots() {
		proxy := instance.proxies[proxyName]
		if err := proxy.processSnapshotFn(proxySnapshot); err != nil {
			return err
		}
	}
	return nil
}

func (instance *sharedGroup) snapshot() ([]byte, error) {
	var err error
	proxySnapshots := make(map[string][]byte)
	for _, proxy := range instance.proxies {
		if proxy.snapshotFn != nil {
			proxySnapshots[proxy.name], err = proxy.snapshotFn()
			if err != nil {
				return nil, err
			}
		}
	}

	return (&raftV1.GroupStateSnapshot{ProxySnapshots: proxySnapshots}).Marshal()
}

func newSharedGroupProxy(group *RaftGroup, name string) *sharedGroupProxy {
	return &sharedGroupProxy{
		name:              name,
		group:             group,
		processFn:         nil,
		processSnapshotFn: nil,
		snapshotFn:        nil,
	}
}

func (instance *sharedGroupProxy) RegisterProcessFn(fn ProcessFn) error {
	if instance.processFn != nil {
		return ErrProcessFnAlreadyRegist
	}
	instance.processFn = fn
	return nil
}

func (instance *sharedGroupProxy) RegisterProcessSnapshotFn(fn ProcessFn) error {
	if instance.processSnapshotFn != nil {
		return ErrProcessFnAlreadyRegist
	}
	instance.processSnapshotFn = fn
	return nil
}

func (instance *sharedGroupProxy) RegisterSnapshotFn(fn SnapshotFn) error {
	if instance.snapshotFn != nil {
		return ErrShapshotFnAlreadyRegist
	}
	instance.snapshotFn = fn
	return nil
}

func (instance *sharedGroupProxy) LeaderId() uint64 {
	return instance.group.LeaderId()
}

func (instance *sharedGroupProxy) Propose(ctx context.Context, data []byte) error {
	proposal := &raftV1.GroupProposalAction{
		ProxyName: instance.name,
		Data:      data,
	}
	proposalData, err := proposal.Marshal()
	if err != nil {
		return err
	}
	return instance.group.Propose(ctx, proposalData)
}

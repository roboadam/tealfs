// Copyright (C) 2025 Adam Hess
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU Affero General Public License as published by the Free
// Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License
// for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package cluster

import (
	"context"
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
)

type Cluster struct {
	Mapper      model.NodeConnectionMapper
	Incoming    <-chan IncomingConn
	Distributer *dist.MirrorDistributer
}

type IncomingConn struct {
	Iam    model.IAm
	ConnId model.ConnId
}

func NewCluster() *Cluster {
	return &Cluster{
		Mapper: *model.NewNodeConnectionMapper(),
	}
}

func (c *Cluster) Start(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case incoming := <-c.Incoming:
		c.handleIam(&incoming)
	}
}

func (c *Cluster) handleIam(incoming *IncomingConn) {
	// c.Mapper.SetAll(incoming.ConnId, incoming.Iam.Address, incoming.Iam.NodeId)
	// // TODO: make Connection status in ui just read from mapper
	// _ = m.addNodeToCluster(*p, i.ConnId)
	// for _, d := range p.Disks {
	// 	diskStatus := model.UiDiskStatus{
	// 		Localness:     model.Remote,
	// 		Availableness: model.Available,
	// 		Node:          p.NodeId,
	// 		Id:            d.Id,
	// 		Path:          d.Path,
	// 	}
	// 	chanutil.Send(m.ctx, m.MgrUiDiskStatuses, diskStatus, "mgr: handleReceives: ui disk status")
	// }
	// syncNodes := m.syncNodesPayloadToSend()
	// connections := m.nodeConnMapper.Connections()
	// for _, connId := range connections.GetValues() {
	// 	mcs := model.MgrConnsSend{
	// 		ConnId:  connId,
	// 		Payload: &syncNodes,
	// 	}
	// 	chanutil.Send(m.ctx, m.MgrConnsSends, mcs, "mgr: handleReceives: sync nodes")
	// }
}

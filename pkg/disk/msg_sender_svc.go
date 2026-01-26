// Copyright (C) 2026 Adam Hess
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

package disk

import (
	"context"
	"tealfs/pkg/model"

	"github.com/sirupsen/logrus"
)

type MsgSenderSvc struct {
	InAddDiskMsg   <-chan model.AddDiskMsg
	InDiskAddedMsg <-chan model.DiskAddedMsg

	OutRemote chan<- model.SendPayloadMsg

	NodeId      model.NodeId
	NodeConnMap *model.NodeConnectionMapper
}

func (m *MsgSenderSvc) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-m.InAddDiskMsg:
			m.sendAddDiskMsg(req)
		case req := <-m.InDiskAddedMsg:
			m.sendDiskAddedMsg(req)
		}
	}
}

func (m *MsgSenderSvc) sendDiskAddedMsg(msg model.DiskAddedMsg) {
	connections := m.NodeConnMap.Connections()
	for _, conn := range connections.GetValues() {
		m.OutRemote <- model.SendPayloadMsg{
			ConnId:  conn,
			Payload: &msg,
		}
	}
}

func (m *MsgSenderSvc) sendAddDiskMsg(msg model.AddDiskMsg) {
	if conn, ok := m.NodeConnMap.ConnForNode(msg.NodeId); ok {
		m.OutRemote <- model.SendPayloadMsg{
			ConnId:  conn,
			Payload: &msg,
		}
	} else {
		logrus.Panic("Not connected")
	}
}

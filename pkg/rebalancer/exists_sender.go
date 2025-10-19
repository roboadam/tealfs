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

package rebalancer

import (
	"context"
	"tealfs/pkg/model"

	"github.com/sirupsen/logrus"
)

type ExistsSender struct {
	InExistsReq        <-chan ExistsReq
	InExistsResp       <-chan ExistsResp
	OutLocalExistsReq  chan<- ExistsReq
	OutLocalExistsResp chan<- ExistsResp
	OutRemote          chan<- model.MgrConnsSend
	NodeId             model.NodeId
	NodeConnMap        *model.NodeConnectionMapper
}

func (e *ExistsSender) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-e.InExistsReq:
			e.sendExistsReq(req)
		case resp := <-e.InExistsResp:
			e.sendExistsResp(resp)
		}
	}
}

func (e *ExistsSender) sendExistsReq(req ExistsReq) {
	if req.DestNodeId == e.NodeId {
		e.OutLocalExistsReq <- req
	} else if conn, ok := e.NodeConnMap.ConnForNode(req.DestNodeId); ok {
		e.OutRemote <- model.MgrConnsSend{
			ConnId:  conn,
			Payload: &req,
		}
	} else {
		logrus.Error("Not connected")
	}
}

func (e *ExistsSender) sendExistsResp(resp ExistsResp) {
	if resp.Req.Caller == e.NodeId {
		e.OutLocalExistsResp <- resp
	} else if conn, ok := e.NodeConnMap.ConnForNode(resp.Req.DestNodeId); ok {
		e.OutRemote <- model.MgrConnsSend{
			ConnId:  conn,
			Payload: &resp,
		}
	} else {
		logrus.Error("Not connected")
	}
}

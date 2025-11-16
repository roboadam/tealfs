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

type MsgSender struct {
	InExistsReq   <-chan ExistsReq
	InExistsResp  <-chan ExistsResp
	InStoreItCmd  <-chan StoreItCmd
	InStoreItReq  <-chan StoreItReq
	InStoreItResp <-chan StoreItResp

	OutExistsReq   chan<- ExistsReq
	OutExistsResp  chan<- ExistsResp
	OutStoreItCmd  chan<- StoreItCmd
	OutStoreItReq  chan<- StoreItReq
	OutStoreItResp chan<- StoreItResp

	OutRemote chan<- model.MgrConnsSend

	NodeId      model.NodeId
	NodeConnMap *model.NodeConnectionMapper
}

func (e *MsgSender) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-e.InExistsReq:
			e.sendExistsReq(req)
		case resp := <-e.InExistsResp:
			e.sendExistsResp(resp)
		case req := <-e.InStoreItReq:
			e.sendStoreItReq(req)
		case cmd := <-e.InStoreItCmd:
			e.sendStoreItCmd(cmd)
		case resp := <-e.InStoreItResp:
			e.sendStoreItResp(resp)
		}
	}
}

func (e *MsgSender) sendStoreItResp(resp StoreItResp) {
	if resp.Req.Cmd.DestNodeId == e.NodeId {
		e.OutStoreItResp <- resp
	} else if conn, ok := e.NodeConnMap.ConnForNode(resp.Req.Cmd.DestNodeId); ok {
		e.OutRemote <- model.MgrConnsSend{
			ConnId:  conn,
			Payload: &resp,
		}
	} else {
		logrus.Error("Not connected")
	}
}

func (e *MsgSender) sendExistsReq(req ExistsReq) {
	if req.DestNodeId == e.NodeId {
		e.OutExistsReq <- req
	} else if conn, ok := e.NodeConnMap.ConnForNode(req.DestNodeId); ok {
		e.OutRemote <- model.MgrConnsSend{
			ConnId:  conn,
			Payload: &req,
		}
	} else {
		logrus.Error("Not connected")
	}
}

func (e *MsgSender) sendExistsResp(resp ExistsResp) {
	if resp.Req.Caller == e.NodeId {
		e.OutExistsResp <- resp
	} else if conn, ok := e.NodeConnMap.ConnForNode(resp.Req.Caller); ok {
		e.OutRemote <- model.MgrConnsSend{
			ConnId:  conn,
			Payload: &resp,
		}
	} else {
		logrus.Error("Not connected")
	}
}

func (e *MsgSender) sendStoreItCmd(cmd StoreItCmd) {
	if cmd.DestNodeId == e.NodeId {
		e.OutStoreItCmd <- cmd
	} else if conn, ok := e.NodeConnMap.ConnForNode(cmd.DestNodeId); ok {
		e.OutRemote <- model.MgrConnsSend{
			ConnId:  conn,
			Payload: &cmd,
		}
	} else {
		logrus.Error("Not connected")
	}
}

func (e *MsgSender) sendStoreItReq(req StoreItReq) {
	if req.NodeId == e.NodeId {
		e.OutStoreItReq <- req
	} else if conn, ok := e.NodeConnMap.ConnForNode(req.NodeId); ok {
		e.OutRemote <- model.MgrConnsSend{
			ConnId:  conn,
			Payload: &req,
		}
	} else {
		logrus.Error("Not connected")
	}
}

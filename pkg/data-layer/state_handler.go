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

package datalayer

import (
	"context"
	"sync"
	"tealfs/pkg/model"

	"github.com/sirupsen/logrus"
)

type StateHandler struct {
	OutSaveRequest   chan<- SaveRequest
	OutDeleteRequest chan<- DeleteRequest
	OutSends         chan<- model.SendPayloadMsg

	state state
	mux   sync.Mutex

	MainNodeId  model.NodeId
	MyNodeId    model.NodeId
	NodeConnMap *model.NodeConnectionMapper
}

func (s *StateHandler) Start(ctx context.Context) {
	if s.MainNodeId == s.MyNodeId {
		s.OutSaveRequest = s.state.outSaveRequest
	} else {
		saveRequests := make(chan SaveRequest)
		s.state.outSaveRequest = saveRequests
		for {
			select {
			case <-ctx.Done():
				return
			case req := <-saveRequests:
				var connId model.ConnId = -1
				for _, dest := range req.from {
					if dest.nodeId == s.MyNodeId {
						s.OutSaveRequest <- req
						continue
					}
					if foundConn, ok := s.NodeConnMap.ConnForNode(dest.nodeId); ok {
						connId = foundConn
					}
				}
				if connId == -1 {
					logrus.Panic("No connection found")
				}
				s.OutSends <- model.SendPayloadMsg{
					ConnId:  connId,
					Payload: req,
				}
			}
		}
	}
}

type SetDiskSpaceParams struct {
	d     dest
	space int
}

func (s *StateHandler) SetDiskSpace(d dest, space int) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.MainNodeId == s.MyNodeId {
		s.state.setDiskSpace(d, space)
	} else {
		if conn, ok := s.NodeConnMap.ConnForNode(s.MainNodeId); ok {
			params := SetDiskSpaceParams{d: d, space: space}
			s.OutSends <- model.SendPayloadMsg{
				ConnId:  conn,
				Payload: params,
			}
		}
	}
}

type SavedParams struct {
	blockId model.BlockId
	d       dest
}

func (s *StateHandler) Saved(blockId model.BlockId, d dest) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.MainNodeId == s.MyNodeId {
		s.state.saved(blockId, d)
	} else {
		if conn, ok := s.NodeConnMap.ConnForNode(s.MainNodeId); ok {
			params := SavedParams{blockId: blockId, d: d}
			s.OutSends <- model.SendPayloadMsg{
				ConnId:  conn,
				Payload: params,
			}
		}
	}
}

type DeletedParams struct {
	b model.BlockId
	d dest
}

func (s *StateHandler) Deleted(b model.BlockId, d dest) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.MainNodeId == s.MyNodeId {
		s.state.deleted(b, d)
	} else {
		if conn, ok := s.NodeConnMap.ConnForNode(s.MainNodeId); ok {
			params := DeletedParams{b: b, d: d}
			s.OutSends <- model.SendPayloadMsg{
				ConnId:  conn,
				Payload: params,
			}
		}
	}
}

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

	log "github.com/sirupsen/logrus"
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
		s.state.outSaveRequest = s.OutSaveRequest
		s.state.outDeleteRequest = s.OutDeleteRequest
	} else {
		saveRequests := make(chan SaveRequest, 1)
		deleteRequests := make(chan DeleteRequest, 1)
		s.state.outSaveRequest = saveRequests
		s.state.outDeleteRequest = deleteRequests

		for {
			select {
			case <-ctx.Done():
				return
			case req := <-saveRequests:
				var connId model.ConnId = -1
				for _, dest := range req.From {
					if dest.NodeId == s.MyNodeId {
						s.OutSaveRequest <- req
						continue
					}
					if foundConn, ok := s.NodeConnMap.ConnForNode(dest.NodeId); ok {
						connId = foundConn
					}
				}
				if connId == -1 {
					log.Panic("No connection found")
				}
				s.OutSends <- model.SendPayloadMsg{
					ConnId:  connId,
					Payload: req,
				}
			case req := <-deleteRequests:
				if req.dest.NodeId == s.MyNodeId {
					s.OutDeleteRequest <- req
					continue
				}
				if foundConn, ok := s.NodeConnMap.ConnForNode(req.dest.NodeId); ok {
					s.OutSends <- model.SendPayloadMsg{
						ConnId:  foundConn,
						Payload: req,
					}
				} else {
					log.Panic("No connection found")
				}
			}
		}
	}
}

type SetDiskSpaceParams struct {
	d     Dest
	space int
}

func (s *StateHandler) SetDiskSpace(d Dest, space int) {
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
	d       Dest
}

func (s *StateHandler) Saved(blockId model.BlockId, d Dest) {
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
	d Dest
}

func (s *StateHandler) Deleted(b model.BlockId, d Dest) {
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

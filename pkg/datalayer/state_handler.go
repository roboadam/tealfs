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

	state        state
	mux          sync.Mutex
	waitingDests map[model.BlockId]chan DestsForBlock

	MyNodeId    model.NodeId
	NodeConnMap *model.NodeConnectionMapper
}

func (s *StateHandler) Start(ctx context.Context) {
	if s.NodeConnMap.MainNode(s.MyNodeId) != s.MyNodeId {
		return
	}

	s.waitingDests = make(map[model.BlockId]chan DestsForBlock)
	saveRequests := make(chan SaveRequest, 1)
	deleteRequests := make(chan DeleteRequest, 1)
	s.state.outSaveRequest = saveRequests
	s.state.outDeleteRequest = deleteRequests

	go s.listen(ctx, saveRequests, deleteRequests)
}

type SetDiskSpaceParams struct {
	D     Dest
	Space int
}

func (s *StateHandler) SetDiskSpace(d Dest, space int) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.NodeConnMap.MainNode(s.MyNodeId) == s.MyNodeId {
		s.state.setDiskSpace(d, space)
	} else {
		if conn, ok := s.NodeConnMap.ConnForNode(s.NodeConnMap.MainNode(s.MyNodeId)); ok {
			params := SetDiskSpaceParams{D: d, Space: space}
			s.OutSends <- model.SendPayloadMsg{
				ConnId:  conn,
				Payload: params,
			}
		}
	}
}

type SavedParams struct {
	B model.BlockId
	D Dest
}

func (s *StateHandler) Saved(blockId model.BlockId, d Dest) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.NodeConnMap.MainNode(s.MyNodeId) == s.MyNodeId {
		s.state.saved(blockId, d)
	} else {
		if conn, ok := s.NodeConnMap.ConnForNode(s.NodeConnMap.MainNode(s.MyNodeId)); ok {
			params := SavedParams{B: blockId, D: d}
			s.OutSends <- model.SendPayloadMsg{
				ConnId:  conn,
				Payload: params,
			}
		}
	}
}

type DeletedParams struct {
	B model.BlockId
	D Dest
}

func (s *StateHandler) Deleted(b model.BlockId, d Dest) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.NodeConnMap.MainNode(s.MyNodeId) == s.MyNodeId {
		s.state.deleted(b, d)
	} else {
		if conn, ok := s.NodeConnMap.ConnForNode(s.NodeConnMap.MainNode(s.MyNodeId)); ok {
			params := DeletedParams{B: b, D: d}
			s.OutSends <- model.SendPayloadMsg{
				ConnId:  conn,
				Payload: params,
			}
		}
	}
}

type DestsForBlockParams struct {
	BlockId        model.BlockId
	PreferedNodeId model.NodeId
	Caller         model.NodeId
}

func (s *StateHandler) DestsForBlock(blockId model.BlockId, preferedNodeId model.NodeId, caller model.NodeId) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.NodeConnMap.MainNode(s.MyNodeId) == s.MyNodeId {
		s.state.destsForBlock(blockId, preferedNodeId, caller)
		return
	}

	if conn, ok := s.NodeConnMap.ConnForNode(s.NodeConnMap.MainNode(s.MyNodeId)); ok {
		params := DestsForBlockParams{BlockId: blockId, PreferedNodeId: preferedNodeId, Caller: caller}
		s.OutSends <- model.SendPayloadMsg{
			ConnId:  conn,
			Payload: params,
		}
	}
}

func (s *StateHandler) listen(ctx context.Context, saveRequests chan SaveRequest, deleteRequests chan DeleteRequest) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-saveRequests:
			local, connId := s.whereToSendSaveRequest(req)
			if local {
				s.OutSaveRequest <- req
			} else {
				s.OutSends <- model.SendPayloadMsg{
					ConnId:  connId,
					Payload: req,
				}
			}
		case req := <-deleteRequests:
			if req.Dest.NodeId == s.MyNodeId {
				s.OutDeleteRequest <- req
				continue
			}
			if foundConn, ok := s.NodeConnMap.ConnForNode(req.Dest.NodeId); ok {
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

func (s *StateHandler) whereToSendSaveRequest(req SaveRequest) (local bool, connId model.ConnId) {
	found := false
	for _, dest := range req.From {
		if dest.NodeId == s.MyNodeId {
			local = true
			return
		}
		if foundConn, ok := s.NodeConnMap.ConnForNode(dest.NodeId); ok {
			connId = foundConn
			found = true
			local = false
		}
	}
	if !found {
		log.Panic("didn't find conn")
	}
	return
}

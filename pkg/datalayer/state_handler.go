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
	OutPayload       chan<- model.Payload2

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
	deleteRequests := make(chan DeleteRequest, 1)
	s.state.outDeleteRequest = deleteRequests

	go s.listen(ctx, deleteRequests)
}

type SetDiskSpaceParams struct {
	D          model.NodeDisk
	Space      int
	MainNodeId model.NodeId
}

func (s *SetDiskSpaceParams) Destination() model.NodeId {
	return s.MainNodeId
}

func (s *StateHandler) SetDiskSpace(d model.NodeDisk, space int) {
	s.mux.Lock()
	defer s.mux.Unlock()

	mainNodeId := s.NodeConnMap.MainNode(s.MyNodeId)
	if mainNodeId == s.MyNodeId {
		s.state.setDiskSpace(d, space)
	} else {
		s.OutPayload <- &SetDiskSpaceParams{D: d, Space: space, MainNodeId: mainNodeId}
	}
}

type SavedParams struct {
	B          model.BlockId
	D          model.NodeDisk
	MainNodeId model.NodeId
}

func (s *SavedParams) Destination() model.NodeId {
	return s.MainNodeId
}

func (s *StateHandler) Saved(blockId model.BlockId, d model.NodeDisk) {
	s.mux.Lock()
	defer s.mux.Unlock()

	mainNodeId := s.NodeConnMap.MainNode(s.MyNodeId)
	if mainNodeId == s.MyNodeId {
		s.state.saved(blockId, d)
	} else {
		s.OutPayload <- &SavedParams{B: blockId, D: d, MainNodeId: mainNodeId}
	}
}

type DeletedParams struct {
	B          model.BlockId
	D          model.NodeDisk
	MainNodeId model.NodeId
}

func (d *DeletedParams) Destination() model.NodeId {
	return d.MainNodeId
}

func (s *StateHandler) Deleted(b model.BlockId, d model.NodeDisk) {
	s.mux.Lock()
	defer s.mux.Unlock()

	mainNodeId := s.NodeConnMap.MainNode(s.MyNodeId)
	if mainNodeId == s.MyNodeId {
		s.state.deleted(b, d)
	} else {
		s.OutPayload <- &DeletedParams{B: b, D: d, MainNodeId: mainNodeId}
	}
}

type DestsForBlockParams struct {
	BlockId        model.BlockId
	PreferedNodeId model.NodeId
	Caller         model.NodeId
	MainNodeId     model.NodeId
}

func (d *DestsForBlockParams) Destination() model.NodeId {
	return d.MainNodeId
}

type NotMainNodeErr struct{}

func (e NotMainNodeErr) Error() string {
	return "This is no the main node"
}

func (s *StateHandler) FetchBlockReqToCmd(req *model.FetchBlockReq) (*model.FetchBlockCmd, error) {
	if s.NodeConnMap.MainNode(s.MyNodeId) != s.MyNodeId {
		return nil, NotMainNodeErr{}
	}

	dests := s.state.destsForBlock(req.BlockId, req.Caller)
	cmd := model.FetchBlockCmd{
		Sources: dests,
		BlockId: req.BlockId,
		Caller:  req.Caller,
		Id:      req.Id,
	}
	return &cmd, nil
}

func (s *StateHandler) DestsForBlock(blockId model.BlockId, preferedNodeId model.NodeId, caller model.NodeId) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.NodeConnMap.MainNode(s.MyNodeId) == s.MyNodeId {
		s.state.destsForBlock(blockId, preferedNodeId)
		return
	}

	params := DestsForBlockParams{BlockId: blockId, PreferedNodeId: preferedNodeId, Caller: caller}
	s.OutPayload <- &params
}

func (s *StateHandler) listen(ctx context.Context, deleteRequests chan DeleteRequest) {
	for {
		select {
		case <-ctx.Done():
			return
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

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
	"sync"
	"tealfs/pkg/model"

	log "github.com/sirupsen/logrus"
)

type StateHandler struct {
	OutPayload chan<- model.Payload2

	state state
	mux   sync.Mutex

	MyNodeId    model.NodeId
	NodeConnMap *model.NodeConnectionMapper
}

func (s *StateHandler) Start() {
	if s.OutPayload == nil || s.MyNodeId == "" || s.NodeConnMap == nil {
		log.Panic("Invalid inputs for state handler")
	}
	if s.NodeConnMap.MainNode(s.MyNodeId) != s.MyNodeId {
		return
	}

	s.state.outPayload = s.OutPayload
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

type NotMainNodeErr struct{}

func (e NotMainNodeErr) Error() string {
	return "This is no the main node"
}

func (s *StateHandler) FetchBlockReqToCmd(req *model.FetchBlockReq) (model.Payload2, error) {
	if s.NodeConnMap.MainNode(s.MyNodeId) != s.MyNodeId {
		return nil, NotMainNodeErr{}
	}

	dests := s.state.destsForBlock(req.BlockId, req.Caller)

	if len(dests) == 0 {
		resp := model.FetchBlockResp{
			Caller:  req.Caller,
			Block:   model.Block{},
			Id:      req.Id,
			Success: false,
			Msg:     "block not found",
		}
		return &resp, nil
	}

	cmd := model.FetchBlockCmd{
		Sources: dests,
		BlockId: req.BlockId,
		Caller:  req.Caller,
		Id:      req.Id,
	}
	return &cmd, nil
}

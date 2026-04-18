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
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type SaveRequestHandler struct {
	InSaveRequests        <-chan SaveRequest
	InDataForSaveRequest  <-chan DataForSaveRequest
	Disks                 *set.Set[disk.Disk]
	NodeId                model.NodeId
	OutDataForSaveRequest chan<- DataForSaveRequest
	OutSends              chan<- model.SendPayloadMsg
	NodeConnMap           *model.NodeConnectionMapper
	StateHandler          Saver
}

type Saver interface {
	Saved(blockId model.BlockId, d model.NodeDisk)
}

type DataForSaveRequest struct {
	SaveRequest SaveRequest
	Data        []byte
}

func (s *SaveRequestHandler) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-s.InSaveRequests:
			s.handleSaveRequest(req)
		case req := <-s.InDataForSaveRequest:
			s.handleDataForSaveRequest(req)
		}
	}
}

func (s *SaveRequestHandler) handleDataForSaveRequest(req DataForSaveRequest) {
	for _, disk := range s.Disks.GetValues() {
		if req.SaveRequest.To.DiskId == disk.DiskId() {
			s.saveDataToDisk(disk, req)
			return
		}
	}
}

func (s *SaveRequestHandler) saveDataToDisk(disk disk.Disk, req DataForSaveRequest) {
	if disk.Save(req.Data, req.SaveRequest.BlockId) {
		s.StateHandler.Saved(req.SaveRequest.BlockId, req.SaveRequest.To)
	}
}

func (s *SaveRequestHandler) handleSaveRequest(req SaveRequest) {
	for _, dest := range req.From {
		s.handleSaveRequestForDest(req, dest)
	}
}

func (s *SaveRequestHandler) handleSaveRequestForDest(req SaveRequest, dest model.NodeDisk) {
	d, ok := s.hasDisk(dest.NodeId, dest.DiskId)
	if !ok {
		return
	}
	data, ok := d.Get(req.BlockId)
	if !ok {
		return
	}
	s.routeData(req.To, DataForSaveRequest{SaveRequest: req, Data: data})
}

func (s *SaveRequestHandler) routeData(to model.NodeDisk, outReq DataForSaveRequest) {
	if to.NodeId == s.NodeId {
		s.OutDataForSaveRequest <- outReq
		return
	}
	conn, ok := s.NodeConnMap.ConnForNode(to.NodeId)
	if ok {
		s.OutSends <- model.SendPayloadMsg{ConnId: conn, Payload: outReq}
	}
}

func (s *SaveRequestHandler) hasDisk(nodeId model.NodeId, diskId model.DiskId) (disk.Disk, bool) {
	if s.NodeId != nodeId {
		return disk.Disk{}, false
	}
	for _, d := range s.Disks.GetValues() {
		if d.DiskId() == diskId {
			return d, true
		}
	}
	return disk.Disk{}, false
}

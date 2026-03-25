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
	// "tealfs/pkg/disk"
	"tealfs/pkg/disk"
	"tealfs/pkg/model"
	"tealfs/pkg/set"

	log "github.com/sirupsen/logrus"
)

type SaveRequestHandler struct {
	InSaveRequests        <-chan SaveRequest
	Disks                 *set.Set[disk.Disk]
	NodeId                model.NodeId
	OutDataforSaveRequest chan<- DataForSaveRequest
	OutSends              chan<- model.SendPayloadMsg
	NodeConnMap           *model.NodeConnectionMapper
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
			for _, dest := range req.From {
				if d, ok := s.HasDisk(dest.NodeId, dest.DiskId); ok {
					if data, ok := d.Get(req.BlockId); ok {
						outReq := DataForSaveRequest{
							SaveRequest: req,
							Data:        data,
						}
						if req.To.NodeId == s.NodeId {
							s.OutDataforSaveRequest <- outReq
						} else if conn, ok := s.NodeConnMap.ConnForNode(req.To.NodeId); ok {
							s.OutSends <- model.SendPayloadMsg{
								ConnId:  conn,
								Payload: outReq,
							}
						} else {
							log.Panic("no connection")
						}
					}
				}
			}
		}
	}
}

func (s *SaveRequestHandler) HasDisk(nodeId model.NodeId, diskId model.DiskId) (disk.Disk, bool) {
	if s.NodeId != nodeId {
		return disk.Disk{}, false
	}
	for _, d := range s.Disks.GetValues() {
		return d, true
	}
	return disk.Disk{}, false
}

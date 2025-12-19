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

package disk

import (
	"context"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type IamSender struct {
	InIamDiskUpdate  <-chan struct{}
	OutSends         chan<- model.SendPayloadMsg
	Mapper           *model.NodeConnectionMapper
	LocalDiskSvcList *set.Set[Disk]
	NodeId           model.NodeId
	Address          string
}

func (i *IamSender) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-i.InIamDiskUpdate:
			iam := model.IAm{
				NodeId:  i.NodeId,
				Disks:   i.diskInfo(),
				Address: i.Address,
			}
			conns := i.Mapper.Connections()
			for _, conn := range conns.GetValues() {
				i.OutSends <- model.SendPayloadMsg{
					ConnId:  conn,
					Payload: &iam,
				}
			}

		}
	}
}

func (i *IamSender) diskInfo() []model.DiskInfo {
	result := []model.DiskInfo{}
	for _, d := range i.LocalDiskSvcList.GetValues() {
		result = append(result, model.DiskInfo{
			DiskId: d.diskId,
			Path:   d.path.String(),
			NodeId: i.NodeId,
		})
	}
	return result
}

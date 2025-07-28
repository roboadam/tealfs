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
)

type IamSender struct {
	InIamDiskUpdate <-chan []model.AddDiskReq
	OutSends        chan<- model.MgrConnsSend
	Mapper          *model.NodeConnectionMapper
	NodeId          model.NodeId
	Address         string
}

func (i *IamSender) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case disks := <-i.InIamDiskUpdate:
			iam := model.IAm{
				NodeId:  i.NodeId,
				Disks:   disks,
				Address: i.Address,
			}
			conns := i.Mapper.Connections()
			for _, conn := range conns.GetValues() {
				i.OutSends <- model.MgrConnsSend{
					ConnId:  conn,
					Payload: &iam,
				}
			}

		}
	}
}

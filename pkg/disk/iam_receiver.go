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
	"tealfs/pkg/disk/dist"
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type IamReceiver struct {
	InIam       <-chan model.IAm
	OutSave     chan<- struct{}
	Distributer *dist.MirrorDistributer
	AllDiskIds  *set.Set[model.AddDiskReq]
}

func (i *IamReceiver) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case iam := <-i.InIam:
			for _, d := range iam.Disks {
				i.Distributer.SetWeight(d.NodeId, d.DiskId, 1)
				i.AllDiskIds.Add(d)
			}
			i.OutSave <- struct{}{}
		}
	}
}

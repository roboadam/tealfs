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

package conns

import (
	"tealfs/pkg/model"
	"tealfs/pkg/set"
)

type IamSender struct {
	NodeId      model.NodeId
	Address     string
	Disks       *set.Set[model.DiskInfo]
	NodeConnMap *model.NodeConnectionMapper
}

func (i *IamSender) generateIam() *model.IAm {
	disks := i.Disks.GetValues()
	return &model.IAm{
		Node: model.IamNodeAddress{
			NodeId:  i.NodeId,
			Address: i.Address,
		},
		Disks:    disks,
		Siblings: i.siblings(),
		Dest:     "",
	}
}

func (i *IamSender) siblings() []model.IamNodeAddress {
	result := []model.IamNodeAddress{}
	sibs := i.NodeConnMap.NodesWithAddress()
	for _, node := range sibs {
		result = append(result, model.IamNodeAddress{
			NodeId:  node.J,
			Address: node.K,
		})
	}
	return result
}

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

package model

type IAm struct {
	NodeId    NodeId
	Disks     []DiskIdPath
	Address   string
	FreeBytes uint32
}

func (i *IAm) Type() PayloadType {
	return IAmType
}

func NewIam(nodeId NodeId, disks []DiskIdPath, address string, freeBytes uint32) IAm {
	return IAm{
		NodeId:    nodeId,
		Disks:     disks,
		Address:   address,
		FreeBytes: freeBytes,
	}
}

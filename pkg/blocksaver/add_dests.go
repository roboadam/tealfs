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

package blocksaver

import "tealfs/pkg/model"

func (b *BlockSaver) destsFor(req model.PutBlockReq) []Dest {
	dests := make([]Dest, 0, 2)
	ptrs := b.Distributer.WritePointersForId(req.Block.Id)
	for _, ptr := range ptrs {
		dests = append(dests, Dest{
			NodeId: ptr.NodeId,
			DiskId: ptr.Disk,
		})
	}
	return dests
}

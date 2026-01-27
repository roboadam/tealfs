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
	"cmp"
	"slices"
	"sync"
	"tealfs/pkg/model"
	"tealfs/pkg/set"

	log "github.com/sirupsen/logrus"
)

type State struct {
	blockMap      set.OtM[model.DiskId, model.BlockId]
	diskFreeSpace []diskFreeSpace
	mux           sync.RWMutex
}

type diskFreeSpace struct {
	diskId    model.DiskId
	freeSpace int
}

func (s *State) SetDiskFreeSpace(diskId model.DiskId, freeSpace int) {
	s.mux.Lock()
	defer s.mux.Unlock()
	defer slices.SortFunc(s.diskFreeSpace, sortDesc)

	for i := range s.diskFreeSpace {
		if s.diskFreeSpace[i].diskId == diskId {
			s.diskFreeSpace[i].freeSpace = freeSpace
			return
		}
	}

	s.diskFreeSpace = append(s.diskFreeSpace, diskFreeSpace{diskId: diskId, freeSpace: freeSpace})
}

func sortDesc(a diskFreeSpace, b diskFreeSpace) int {
	return cmp.Compare(b.freeSpace, a.freeSpace)
}

func (s *State) AddBlock(blockId model.BlockId) model.DiskId {
	if diskId, ok := s.blockMap.GetKey(blockId); ok {
		return diskId
	}
}

func (s *State) emptiestDisks(count int) []model.DiskId {
	s.mux.RLock()
	defer s.mux.Unlock()

	if len(s.diskFreeSpace) == 0 {
		log.Panic("No disks")
	}


}

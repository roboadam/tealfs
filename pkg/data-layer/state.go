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

	log "github.com/sirupsen/logrus"
)

type State struct {
	diskBlockMap  map[model.DiskId][]model.BlockId
	blockDiskMap  map[model.BlockId][]model.DiskId
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
	if _, exists := s.diskBlockMap[diskId]; !exists {
		s.diskBlockMap[diskId] = make([]model.BlockId, 0)
	}
}

func sortDesc(a diskFreeSpace, b diskFreeSpace) int {
	return cmp.Compare(b.freeSpace, a.freeSpace)
}

func (s *State) AddBlock(blockId model.BlockId) []model.DiskId {
	s.mux.Lock()
	defer s.mux.Unlock()

	if diskId, ok := s.blockDiskMap[blockId]; ok {
		return diskId
	}

	disks := s.emptiestDisks(2)
	s.blockDiskMap[blockId] = disks
	for _, disk := range disks {
		s.diskBlockMap[disk] = append(s.diskBlockMap[disk], blockId)
	}
	return disks
}

func (s *State) emptiestDisks(count int) []model.DiskId {
	s.mux.RLock()
	defer s.mux.Unlock()

	if len(s.diskFreeSpace) == 0 {
		log.Panic("No disks")
	}

	result := make([]model.DiskId, 0, count)
	for _, dfs := range s.diskFreeSpace[:count] {
		result = append(result, dfs.diskId)
	}
	return result
}

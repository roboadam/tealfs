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
	"sync"
	"tealfs/pkg/model"
)

type State struct {
	diskBlockMapFuture  map[model.DiskId]map[model.BlockId]struct{}
	blockDiskMapFuture  map[model.BlockId]map[model.DiskId]struct{}
	diskBlockMapCurrent map[model.DiskId]map[model.BlockId]struct{}
	blockDiskMapCurrent map[model.BlockId]map[model.DiskId]struct{}

	OutSaveRequest   saveRequest
	OutDeleteRequest deleteRequest

	diskSpace []diskSpace
	mux       sync.RWMutex
}

type saveRequest struct {
	to      model.DiskId
	from    []model.DiskId
	blockId model.BlockId
}

func (s *saveRequest) Type() model.PayloadType {
	return model.StateSaveRequest
}

type deleteRequest struct {
	diskId  model.DiskId
	blockId model.BlockId
}

func (s *deleteRequest) Type() model.PayloadType {
	return model.StateDeleteRequest
}

type diskSpace struct {
	diskId model.DiskId
	space  int
}

func (s *State) SetDiskSpace(diskId model.DiskId, space int) {
	s.mux.Lock()
	defer s.mux.Unlock()

	for i := range s.diskSpace {
		if s.diskSpace[i].diskId == diskId {
			s.diskSpace[i].space = space
			return
		}
	}

	s.diskSpace = append(s.diskSpace, diskSpace{diskId: diskId, space: space})

	if _, exists := s.diskBlockMapFuture[diskId]; !exists {
		s.diskBlockMapFuture[diskId] = make(map[model.BlockId]struct{})
	}
	if _, exists := s.diskBlockMapFuture[diskId]; !exists {
		s.diskBlockMapFuture[diskId] = make(map[model.BlockId]struct{})
	}
}

func (s *State) Saved(blockId model.BlockId, diskId model.DiskId) {
	s.mux.Lock()
	defer s.mux.Unlock()

	addBlockAndDisk(s.diskBlockMapCurrent, s.blockDiskMapCurrent, blockId, diskId)
}

func addBlockAndDisk(
	diskBlockMap map[model.DiskId]map[model.BlockId]struct{},
	blockDiskMap map[model.BlockId]map[model.DiskId]struct{},
	blockId model.BlockId,
	diskId model.DiskId,
) {
	if _, ok := blockDiskMap[blockId]; !ok {
		blockDiskMap[blockId] = make(map[model.DiskId]struct{})
	}
	blockDiskMap[blockId][diskId] = struct{}{}

	if _, ok := diskBlockMap[diskId]; !ok {
		diskBlockMap[diskId] = make(map[model.BlockId]struct{})
	}
	diskBlockMap[diskId][blockId] = struct{}{}
}

func (s *State) Deleted(blockId model.BlockId, diskId model.DiskId) {}

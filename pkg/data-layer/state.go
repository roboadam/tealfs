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

	OutSaveRequest   chan<- saveRequest
	OutDeleteRequest chan<- deleteRequest

	diskSpace []diskSpace
	mux       sync.RWMutex
}

func (s *State) emptiestDisks() []model.DiskId {
	var disk1 model.DiskId
	var disk2 model.DiskId

	var ratio1 float32 = 0
	var ratio2 float32 = 0

	for _, ds := range s.diskSpace {
		ratio := float32(ds.space) / (float32(len(s.diskBlockMapFuture[ds.diskId]) + 1))
		if ratio > ratio1 {
			disk2 = disk1
			ratio2 = ratio1
			disk1 = ds.diskId
			ratio1 = ratio
		} else if ratio > ratio2 {
			disk2 = ds.diskId
			ratio2 = ratio
		}
	}
	if disk1 == "" {
		return []model.DiskId{}
	}
	if disk2 == "" {
		return []model.DiskId{disk1}
	}
	return []model.DiskId{disk1, disk2}
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
	if futureDisks, ok := s.blockDiskMapFuture[blockId]; ok {
		if currentDisks, ok := s.blockDiskMapCurrent[blockId]; ok {
			if setEqual(futureDisks, currentDisks) {
				return
			}
			needToSave := minus(futureDisks, currentDisks)
			if len(needToSave) > 0 {
				for toSaveTo := range needToSave {
					s.OutSaveRequest <- saveRequest{
						to:      toSaveTo,
						from:    toSlice(currentDisks),
						blockId: blockId,
					}
				}
			} else {
				needToDelete := minus(currentDisks, futureDisks)
				for toDeleteFrom := range needToDelete {
					s.OutDeleteRequest <- deleteRequest{
						diskId:  toDeleteFrom,
						blockId: blockId,
					}
				}
			}
		}
	}
	for _, emptyDisk := range s.emptiestDisks() {
		addBlockAndDisk(s.diskBlockMapFuture, s.blockDiskMapFuture, blockId, emptyDisk)
		if emptyDisk != diskId {
			s.OutSaveRequest <- saveRequest{
				to:      emptyDisk,
				from:    []model.DiskId{diskId},
				blockId: blockId,
			}
		}
	}
}

func minus[K comparable](first, second map[K]struct{}) map[K]struct{} {
	result := make(map[K]struct{})
	for k := range first {
		if _, ok := second[k]; !ok {
			result[k] = struct{}{}
		}
	}
	return result
}

func setEqual[K comparable](map1, map2 map[K]struct{}) bool {
	if len(map1) != len(map2) {
		return false
	}

	for key := range map1 {
		if _, exists := map2[key]; !exists {
			return false
		}
	}

	return true
}

func toSlice[K comparable](set map[K]struct{}) []K {
	result := make([]K, 0, len(set))
	for k := range set {
		result = append(result, k)
	}
	return result
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

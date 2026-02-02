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
	diskBlockMapFuture  map[dest]map[model.BlockId]struct{}
	blockDiskMapFuture  map[model.BlockId]map[dest]struct{}
	diskBlockMapCurrent map[dest]map[model.BlockId]struct{}
	blockDiskMapCurrent map[model.BlockId]map[dest]struct{}

	OutSaveRequest   chan<- saveRequest
	OutDeleteRequest chan<- deleteRequest

	diskSpace []diskSpace
	mux       sync.RWMutex
}

func (s *State) emptiestDisks() []dest {
	var dest1 *dest
	var dest2 *dest

	var ratio1 float32 = 0
	var ratio2 float32 = 0

	for _, ds := range s.diskSpace {
		ratio := float32(ds.space) / (float32(len(s.diskBlockMapFuture[ds.dest]) + 1))
		if ratio > ratio1 {
			dest2 = dest1
			ratio2 = ratio1
			dest1 = &ds.dest
			ratio1 = ratio
		} else if ratio > ratio2 {
			dest2 = &ds.dest
			ratio2 = ratio
		}
	}
	if dest1 == nil {
		return []dest{}
	}
	if dest2 == nil {
		return []dest{*dest1}
	}
	return []dest{*dest1, *dest2}
}

type saveRequest struct {
	to      dest
	from    []dest
	blockId model.BlockId
}

func (s *saveRequest) Type() model.PayloadType {
	return model.StateSaveRequest
}

type deleteRequest struct {
	dest    dest
	blockId model.BlockId
}

func (s *deleteRequest) Type() model.PayloadType {
	return model.StateDeleteRequest
}

type diskSpace struct {
	dest  dest
	space int
}

type dest struct {
	diskId model.DiskId
	nodeId model.NodeId
}

func (s *State) SetDiskSpace(d dest, space int) {
	s.mux.Lock()
	defer s.mux.Unlock()

	for i := range s.diskSpace {
		if s.diskSpace[i].dest == d {
			s.diskSpace[i].space = space
			return
		}
	}

	s.diskSpace = append(s.diskSpace, diskSpace{dest: d, space: space})

	if _, exists := s.diskBlockMapFuture[d]; !exists {
		s.diskBlockMapFuture[d] = make(map[model.BlockId]struct{})
	}
	if _, exists := s.diskBlockMapFuture[d]; !exists {
		s.diskBlockMapFuture[d] = make(map[model.BlockId]struct{})
	}
}

func (s *State) Saved(blockId model.BlockId, d dest) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.addBlockToCurrent(blockId, d)
	if futureDisks, ok := s.blockDiskMapFuture[blockId]; ok {
		if currentDisks, ok := s.blockDiskMapCurrent[blockId]; ok {
			if setEqual(futureDisks, currentDisks) {
				return
			}
			needToSave := minus(futureDisks, currentDisks)
			if len(needToSave) > 0 {
				s.sendSaveMsgs(needToSave, currentDisks, blockId)
			} else {
				needToDelete := minus(currentDisks, futureDisks)
				s.sendDeleteMsgs(needToDelete, blockId)
			}
		}
	}
	for _, emptyDisk := range s.emptiestDisks() {
		s.addBlockToFuture(blockId, emptyDisk)
		if emptyDisk != d {
			s.OutSaveRequest <- saveRequest{
				to:      emptyDisk,
				from:    []dest{d},
				blockId: blockId,
			}
		}
	}
}

func (s *State) Deleted(b model.BlockId, d dest) {
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.blockDiskMapCurrent, b)
	delete(s.diskBlockMapCurrent, d)
	if _, ok := s.blockDiskMapFuture[b][d]; ok {
		sources := toSlice(s.blockDiskMapCurrent[b])
		s.OutSaveRequest <- saveRequest{
			to:      d,
			from:    sources,
			blockId: b,
		}
	}
}

func (s *State) addBlockToFuture(blockId model.BlockId, emptyDisk dest) {
	addBlockAndDisk(s.diskBlockMapFuture, s.blockDiskMapFuture, blockId, emptyDisk)
}

func (s *State) sendDeleteMsgs(needToDelete map[dest]struct{}, blockId model.BlockId) {
	for toDeleteFrom := range needToDelete {
		s.OutDeleteRequest <- deleteRequest{
			dest:    toDeleteFrom,
			blockId: blockId,
		}
	}
}

func (s *State) sendSaveMsgs(needToSave map[dest]struct{}, currentDisks map[dest]struct{}, blockId model.BlockId) {
	for toSaveTo := range needToSave {
		s.OutSaveRequest <- saveRequest{
			to:      toSaveTo,
			from:    toSlice(currentDisks),
			blockId: blockId,
		}
	}
}

func (s *State) addBlockToCurrent(blockId model.BlockId, d dest) {
	addBlockAndDisk(s.diskBlockMapCurrent, s.blockDiskMapCurrent, blockId, d)
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
	diskBlockMap map[dest]map[model.BlockId]struct{},
	blockDiskMap map[model.BlockId]map[dest]struct{},
	b model.BlockId,
	d dest,
) {
	if _, ok := blockDiskMap[b]; !ok {
		blockDiskMap[b] = make(map[dest]struct{})
	}
	blockDiskMap[b][d] = struct{}{}

	if _, ok := diskBlockMap[d]; !ok {
		diskBlockMap[d] = make(map[model.BlockId]struct{})
	}
	diskBlockMap[d][b] = struct{}{}
}

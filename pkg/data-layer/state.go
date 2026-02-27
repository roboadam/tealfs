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
	"tealfs/pkg/model"
)

type state struct {
	diskBlockMapFuture   map[Dest]map[model.BlockId]struct{}
	blockDiskMapFuture   map[model.BlockId]map[Dest]struct{}
	diskBlockMapCurrent  map[Dest]map[model.BlockId]struct{}
	blockDiskMapCurrent  map[model.BlockId]map[Dest]struct{}
	diskBlockMapInFlight map[Dest]map[model.BlockId]struct{}
	blockDiskMapInFlight map[model.BlockId]map[Dest]struct{}

	outSaveRequest   chan<- SaveRequest
	outDeleteRequest chan<- DeleteRequest

	diskSpace []diskSpace
}

func (s *state) emptiestDisks() []Dest {
	var dest1 *Dest
	var dest2 *Dest

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
		return []Dest{}
	}
	if dest2 == nil {
		return []Dest{*dest1}
	}
	return []Dest{*dest1, *dest2}
}

type SaveRequest struct {
	To      Dest
	From    []Dest
	BlockId model.BlockId
}

type DeleteRequest struct {
	Dest    Dest
	BlockId model.BlockId
}

type diskSpace struct {
	dest  Dest
	space int
}

type Dest struct {
	DiskId model.DiskId
	NodeId model.NodeId
}

func (s *state) setDiskSpace(d Dest, space int) {
	s.init()
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
}

func (s *state) init() {
	if s.diskBlockMapCurrent == nil {
		s.diskBlockMapCurrent = make(map[Dest]map[model.BlockId]struct{})
		s.diskBlockMapFuture = make(map[Dest]map[model.BlockId]struct{})
		s.diskBlockMapInFlight = make(map[Dest]map[model.BlockId]struct{})
		s.blockDiskMapCurrent = make(map[model.BlockId]map[Dest]struct{})
		s.blockDiskMapFuture = make(map[model.BlockId]map[Dest]struct{})
		s.blockDiskMapInFlight = make(map[model.BlockId]map[Dest]struct{})
	}
}

func (s *state) saved(blockId model.BlockId, d Dest) {
	s.init()
	s.addBlockToCurrent(blockId, d)
	s.removeBlockFromInFlight(blockId, d)
	if futureDisks, ok := s.blockDiskMapFuture[blockId]; ok {
		if currentDisks, ok := s.blockDiskMapCurrent[blockId]; ok {
			if setEqual(futureDisks, currentDisks) {
				return
			}

			needToDelete := minus(currentDisks, futureDisks)
			s.sendDeleteMsgs(needToDelete, blockId)
		}
	}
	for _, emptyDisk := range s.emptiestDisks() {
		s.addBlockToFuture(blockId, emptyDisk)
		if emptyDisk != d && !s.saveAlreadySent(blockId, emptyDisk) {
			s.addBlockToInFlight(blockId, emptyDisk)
			s.outSaveRequest <- SaveRequest{
				To:      emptyDisk,
				From:    []Dest{d},
				BlockId: blockId,
			}
		}
	}
}

func (s *state) deleted(b model.BlockId, d Dest) {
	s.init()
	s.removeBlockFromCurrent(b, d)
	if _, ok := s.blockDiskMapFuture[b][d]; ok {
		sources := toSlice(s.blockDiskMapCurrent[b])
		s.addBlockToInFlight(b, d)
		s.outSaveRequest <- SaveRequest{
			To:      d,
			From:    sources,
			BlockId: b,
		}
	}
}

func (s *state) addBlockToFuture(blockId model.BlockId, emptyDisk Dest) {
	addBlockAndDisk(s.diskBlockMapFuture, s.blockDiskMapFuture, blockId, emptyDisk)
}

func (s *state) sendDeleteMsgs(needToDelete map[Dest]struct{}, blockId model.BlockId) {
	for toDeleteFrom := range needToDelete {
		s.outDeleteRequest <- DeleteRequest{
			Dest:    toDeleteFrom,
			BlockId: blockId,
		}
	}
}

func (s *state) sendSaveMsgs(needToSave map[Dest]struct{}, currentDisks map[Dest]struct{}, blockId model.BlockId) {
	for toSaveTo := range needToSave {
		if !s.saveAlreadySent(blockId, toSaveTo) {
			s.addBlockToInFlight(blockId, toSaveTo)
			s.outSaveRequest <- SaveRequest{
				To:      toSaveTo,
				From:    toSlice(currentDisks),
				BlockId: blockId,
			}
		}
	}
}

func (s *state) addBlockToCurrent(blockId model.BlockId, d Dest) {
	addBlockAndDisk(s.diskBlockMapCurrent, s.blockDiskMapCurrent, blockId, d)
}
func (s *state) removeBlockFromCurrent(blockId model.BlockId, d Dest) {
	removeBlockAndDisk(s.diskBlockMapCurrent, s.blockDiskMapCurrent, blockId, d)
}

func (s *state) saveAlreadySent(blockId model.BlockId, d Dest) bool {
	if inFlightDests, ok := s.blockDiskMapInFlight[blockId]; ok {
		_, ok := inFlightDests[d]
		if ok {
			return true
		}
	}
	if currentDests, ok := s.blockDiskMapCurrent[blockId]; ok {
		_, ok := currentDests[d]
		return ok

	}
	return false
}

func (s *state) addBlockToInFlight(blockId model.BlockId, d Dest) {
	addBlockAndDisk(s.diskBlockMapInFlight, s.blockDiskMapInFlight, blockId, d)
}

func (s *state) removeBlockFromInFlight(blockId model.BlockId, d Dest) {
	removeBlockAndDisk(s.diskBlockMapInFlight, s.blockDiskMapInFlight, blockId, d)
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
	diskBlockMap map[Dest]map[model.BlockId]struct{},
	blockDiskMap map[model.BlockId]map[Dest]struct{},
	b model.BlockId,
	d Dest,
) {
	if _, ok := blockDiskMap[b]; !ok {
		blockDiskMap[b] = make(map[Dest]struct{})
	}
	blockDiskMap[b][d] = struct{}{}

	if _, ok := diskBlockMap[d]; !ok {
		diskBlockMap[d] = make(map[model.BlockId]struct{})
	}
	diskBlockMap[d][b] = struct{}{}
}

func removeBlockAndDisk(
	diskBlockMap map[Dest]map[model.BlockId]struct{},
	blockDiskMap map[model.BlockId]map[Dest]struct{},
	b model.BlockId,
	d Dest,
) {
	if _, ok := diskBlockMap[d]; ok {
		delete(diskBlockMap[d], b)
		if len(diskBlockMap[d]) == 0 {
			delete(diskBlockMap, d)
		}
	}
	if _, ok := blockDiskMap[b]; ok {
		delete(blockDiskMap[b], d)
		if len(blockDiskMap[b]) == 0 {
			delete(blockDiskMap, b)
		}
	}
}

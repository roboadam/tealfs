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

package dist

import (
	"encoding/binary"
	"errors"
	"hash"
	"hash/crc32"
	"maps"
	"tealfs/pkg/model"
)

type XorDistributer struct {
	weights  map[model.NodeId]int
	checksum hash.Hash32
}

func NewXorDistributer() XorDistributer {
	return XorDistributer{
		weights:  make(map[model.NodeId]int),
		checksum: crc32.NewIEEE(),
	}
}

func (d *XorDistributer) RawDataForBlocks(block1 model.Block, block2 model.Block) ([]model.RawData, error) {
	id1 := block1.Id
	id2 := block2.Id
	node1, node2, parity, err := d.generateNodeIds(id1, id2)
	if err != nil {
		return []model.RawData{}, err
	}

	ptr1 := model.DiskPointer{
		NodeId:   node1,
		FileName: string(id1),
	}
	ptr2 := model.DiskPointer{
		NodeId:   node2,
		FileName: string(id2),
	}
	parityPointer := model.DiskPointer{
		NodeId:   parity,
		FileName: string(id1) + "." + string(id2),
	}

	raw1 := model.RawData{
		Ptr:  ptr1,
		Data: block1.Data,
	}
	raw2 := model.RawData{
		Ptr:  ptr2,
		Data: block2.Data,
	}
	rawP := model.RawData{
		Ptr:  parityPointer,
		Data: xor(block1.Data, block2.Data),
	}

	return []model.RawData{raw1, raw2, rawP}, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func xor(data1 []byte, data2 []byte) []byte {
	maxLen := max(len(data1), len(data2))

	result := make([]byte, maxLen)
	for i := range maxLen {
		var d1 byte = 0
		var d2 byte = 0
		if i < len(data1) {
			d1 = data1[i]
		} else if i < len(data2) {
			d2 = data2[i]
		}

		result[i] = d1 ^ d2
	}
	return result
}

func (d *XorDistributer) generateNodeIds(id1 model.BlockId, id2 model.BlockId) (node1 model.NodeId, node2 model.NodeId, parity model.NodeId, err error) {
	if len(d.weights) < 3 {
		return "", "", "", errors.New("not enough nodes to generate parity")
	}

	weights := maps.Clone(d.weights)

	idb := []byte(id1)
	checksum := d.Checksum(idb)
	intHash := int(binary.BigEndian.Uint32(checksum))

	node1 = d.nodeIdForHashAndWeights(intHash, weights)

	if len(d.weights) == 1 {
		for key := range d.weights {
			node2 = key
			parity = key
			err = nil
			return
		}
	}

	delete(weights, node1)

	idb = []byte(id2)
	checksum = d.Checksum(idb)
	intHash = int(binary.BigEndian.Uint32(checksum))

	node2 = d.nodeIdForHashAndWeights(intHash, weights)

	if len(weights) == 1 {
		for key := range d.weights {
			parity = key
			err = nil
			return
		}
	}
	delete(weights, node2)

	idb = []byte(id1 + id2)
	checksum = d.Checksum(idb)
	intHash = int(binary.BigEndian.Uint32(checksum))

	parity = d.nodeIdForHashAndWeights(intHash, weights)
	err = nil
	return
}

func (d *XorDistributer) nodeIdForHashAndWeights(hash int, weights map[model.NodeId]int) model.NodeId {
	total := totalWeight(weights)
	randomNum := hash % total

	cumulativeWeight := 0
	keys := sortedKeys(weights)
	for _, key := range keys {
		cumulativeWeight += weights[key]
		if randomNum < cumulativeWeight {
			return key
		}
	}
	panic("should never get here")
}

func (d *XorDistributer) Checksum(data []byte) []byte {
	d.checksum.Reset()
	d.checksum.Write(data)
	return d.checksum.Sum(nil)
}

func (d *XorDistributer) SetWeight(id model.NodeId, weight int) {
	d.weights[id] = weight
}

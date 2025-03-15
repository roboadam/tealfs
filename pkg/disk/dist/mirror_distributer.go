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
	"hash"
	"hash/crc32"
	"maps"
	"sort"
	"tealfs/pkg/model"
)

type MirrorDistributer struct {
	weights map[model.NodeId]int
	hasher  hash.Hash32
}

func NewMirrorDistributer() MirrorDistributer {
	return MirrorDistributer{
		weights: make(map[model.NodeId]int),
		hasher:  crc32.NewIEEE(),
	}
}

func (d *MirrorDistributer) PointersForId(id model.BlockId) []model.DiskPointer {
	nodeIds := d.generateNodeIds(id)
	data := []model.DiskPointer{}
	for _, nodeId := range nodeIds {
		data = append(data, model.NewDiskPointer(nodeId.n, nodeId.d, string(id)))
	}
	return data
}

type nodeAndDisk struct {
	n model.NodeId
	d model.DiskId
}

func (d *MirrorDistributer) generateNodeIds(id model.BlockId) []nodeAndDisk {
	if len(d.weights) == 0 {
		return []model.NodeId{}
	}

	idb := []byte(id)
	checksum := d.checksum(idb)
	intHash := int(binary.BigEndian.Uint32(checksum))

	node1 := d.nodeIdForHashAndWeights(intHash, d.weights)

	if len(d.weights) == 1 {
		return []model.NodeId{node1}
	}

	weights2 := maps.Clone(d.weights)
	delete(weights2, node1)
	node2 := d.nodeIdForHashAndWeights(intHash, weights2)
	delete(weights2, node2)

	result := []model.NodeId{node1, node2}

	for nodeId := range weights2 {
		result = append(result, nodeId)
	}

	return result
}

func (d *MirrorDistributer) nodeIdForHashAndWeights(hash int, weights map[model.NodeId]int) model.NodeId {
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

func sortedKeys(m map[model.NodeId]int) []model.NodeId {
	stringKeys := make([]string, len(m))
	for k := range m {
		stringKeys = append(stringKeys, string(k))
	}
	sort.Strings(stringKeys)
	keys := make([]model.NodeId, len(stringKeys))
	for _, k := range stringKeys {
		keys = append(keys, model.NodeId(k))
	}
	return keys
}

func totalWeight(weights map[model.NodeId]int) int {
	total := 0
	for _, weight := range weights {
		total += weight
	}
	return total
}

func (d *MirrorDistributer) checksum(data []byte) []byte {
	d.hasher.Reset()
	d.hasher.Write(data)
	return d.hasher.Sum(nil)
}

func (d *MirrorDistributer) SetWeight(id model.NodeId, weight int) {
	if weight > 0 {
		d.weights[id] = weight
	} else {
		delete(d.weights, id)
	}
}

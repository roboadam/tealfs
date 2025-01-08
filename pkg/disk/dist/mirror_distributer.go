// Copyright (C) 2024 Adam Hess
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
	"math/rand"
	"sort"
	"tealfs/pkg/model"
)

type MirrorDistributer struct {
	weights  map[model.NodeId]int
	checksum hash.Hash32
}

func NewMirrorDistributer() MirrorDistributer {
	return MirrorDistributer{
		weights:  make(map[model.NodeId]int),
		checksum: crc32.NewIEEE(),
	}
}

func (d *MirrorDistributer) KeyForId(id model.BlockKeyId) model.BlockKey {
	nodeIds := d.generateNodeIds(id)
	data := []model.DiskPointer{}
	for _, nodeId := range nodeIds {
		data = append(data, model.DiskPointer{NodeId: nodeId, FileName: string(id)})
	}
	return model.BlockKey{
		Id:   id,
		Type: model.Mirrored,
		Data: data,
	}
}

func (d *MirrorDistributer) generateNodeIds(id model.BlockKeyId) []model.NodeId {
	if len(d.weights) == 0 {
		return []model.NodeId{}
	}

	idb := []byte(id)
	checksum := d.Checksum(idb)
	intHash := int(binary.BigEndian.Uint32(checksum))

	node1 := d.nodeIdForHashAndWeights(intHash, d.weights)

	if len(d.weights) == 1 {
		return []model.NodeId{node1}
	}

	weights2 := d.weights
	delete(weights2, node1)
	node2 := d.nodeIdForHashAndWeights(intHash, d.weights)

	return []model.NodeId{node1, node2}
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
	for k := range stringKeys {
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

func randWeight(max int) int {
	return rand.Intn(max)
}

func (d *MirrorDistributer) Checksum(data []byte) []byte {
	d.checksum.Reset()
	d.checksum.Write(data)
	return d.checksum.Sum(nil)
}

func (d *MirrorDistributer) SetWeight(id model.NodeId, weight int) {
	d.weights[id] = weight
}

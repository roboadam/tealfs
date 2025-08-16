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
	"strings"
	"tealfs/pkg/model"
)

type MirrorDistributer struct {
	weights map[string]int
	hasher  hash.Hash32
	nodeId  model.NodeId
}

func NewMirrorDistributer(nodeId model.NodeId) MirrorDistributer {
	return MirrorDistributer{
		weights: make(map[string]int),
		hasher:  crc32.NewIEEE(),
		nodeId:  nodeId,
	}
}

func (d *MirrorDistributer) WritePointersForId(id model.BlockId) []model.DiskPointer {
	ptrs := d.ReadPointersForId(id)
	if len(ptrs) < 2 {
		return ptrs
	}
	return ptrs[:2]
}

func (d *MirrorDistributer) ReadPointersForId(id model.BlockId) []model.DiskPointer {
	nodeIds := d.generateNodeIds(id)
	data := []model.DiskPointer{}
	for _, nodeId := range nodeIds {
		data = append(data, model.DiskPointer{NodeId: nodeId.n, Disk: nodeId.d, FileName: string(id)})
	}
	return data
}

type nodeAndDisk struct {
	n model.NodeId
	d model.DiskId
}

func (n nodeAndDisk) string() string {
	return string(n.n) + "|" + string(n.d)
}

func fromString(val string) nodeAndDisk {
	raw := strings.Split(val, "|")
	return nodeAndDisk{
		n: model.NodeId(raw[0]),
		d: model.DiskId(raw[1]),
	}
}

func (d *MirrorDistributer) generateNodeIds(id model.BlockId) []nodeAndDisk {
	if len(d.weights) == 0 {
		return []nodeAndDisk{}
	}

	idb := []byte(id)
	checksum := d.checksum(idb)
	intHash := int(binary.BigEndian.Uint32(checksum))

	node1 := d.nodeIdForHashAndWeights(intHash, d.weights)

	if len(d.weights) == 1 {
		return []nodeAndDisk{fromString(node1)}
	}

	weights2 := maps.Clone(d.weights)
	delete(weights2, node1)
	node2 := d.nodeIdForHashAndWeights(intHash, weights2)
	delete(weights2, node2)

	result := []nodeAndDisk{fromString(node1), fromString(node2)}
	if result[1].n == d.nodeId {
		result = []nodeAndDisk{result[1], result[0]}
	}

	for nodeId := range weights2 {
		result = append(result, fromString(nodeId))
	}

	return result
}

func (d *MirrorDistributer) nodeIdForHashAndWeights(hash int, weights map[string]int) string {
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

func sortedKeys(m map[string]int) []string {
	stringKeys := make([]string, len(m))
	for k := range m {
		stringKeys = append(stringKeys, string(k))
	}
	sort.Strings(stringKeys)
	return stringKeys
}

func totalWeight(weights map[string]int) int {
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

func (d *MirrorDistributer) SetWeight(node model.NodeId, disk model.DiskId, weight int) {
	nad := nodeAndDisk{n: node, d: disk}
	id := nad.string()
	if weight > 0 {
		d.weights[id] = weight
	} else {
		delete(d.weights, id)
	}
}

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
	"hash/crc32"
	"maps"
	"sort"
	"strings"
	"sync"
	"tealfs/pkg/model"
)

type MirrorDistributer struct {
	weights map[string]int
	mux     sync.RWMutex
}

func NewMirrorDistributer() MirrorDistributer {
	return MirrorDistributer{
		weights: make(map[string]int),
		mux:     sync.RWMutex{},
	}
}

func (d *MirrorDistributer) WritePointersForId(id model.BlockId) []model.DiskPointer {
	d.mux.RLock()
	defer d.mux.RUnlock()
	ptrs := d.readPointersForId(id)
	if len(ptrs) < 2 {
		return ptrs
	}
	return ptrs[:2]
}

func (d *MirrorDistributer) readPointersForId(id model.BlockId) []model.DiskPointer {
	nodeIds := d.generateNodeIds(id)
	data := []model.DiskPointer{}
	for _, nodeId := range nodeIds {
		data = append(data, model.NewDiskPointer(nodeId.n, nodeId.d, id))
	}
	return data
}

func (d *MirrorDistributer) ReadPointersForId(id model.BlockId) []model.DiskPointer {
	d.mux.RLock()
	defer d.mux.RUnlock()
	return d.readPointersForId(id)
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
	sum := checksum(idb)
	intHash := int(binary.BigEndian.Uint32(sum))

	node1 := nodeIdForHashAndWeights(intHash, d.weights)

	if len(d.weights) == 1 {
		return []nodeAndDisk{fromString(node1)}
	}

	weights2 := maps.Clone(d.weights)
	delete(weights2, node1)
	node2 := nodeIdForHashAndWeights(intHash, weights2)
	delete(weights2, node2)

	result := []nodeAndDisk{fromString(node1), fromString(node2)}

	for nodeId := range weights2 {
		result = append(result, fromString(nodeId))
	}

	return result
}

func nodeIdForHashAndWeights(hash int, weights map[string]int) string {
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

func checksum(data []byte) []byte {
	hasher := crc32.NewIEEE()
	hasher.Write(data)
	return hasher.Sum(nil)
}

func (d *MirrorDistributer) SetWeight(node model.NodeId, disk model.DiskId, weight int) {
	d.mux.Lock()
	defer d.mux.Unlock()
	nad := nodeAndDisk{n: node, d: disk}
	id := nad.string()
	if weight > 0 {
		d.weights[id] = weight
	} else {
		delete(d.weights, id)
	}
}

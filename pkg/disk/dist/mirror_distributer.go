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
	dist1    map[key]model.NodeId
	dist2    map[key]model.NodeId
	weights  map[model.NodeId]int
	checksum hash.Hash32
}

func NewMirrorDistributer() MirrorDistributer {
	return MirrorDistributer{
		dist1:    make(map[key]model.NodeId),
		dist2:    make(map[key]model.NodeId),
		weights:  make(map[model.NodeId]int),
		checksum: crc32.NewIEEE(),
	}
}

func (d *MirrorDistributer) KeyForId(id model.BlockKeyId) model.BlockKey {
}

func (d *MirrorDistributer) randomNodeId(id model.BlockKeyId) model.NodeId {
	idb := []byte(id)
	checksum := d.Checksum(idb)
	total := totalWeight(d.weights)
	randomNum := int(binary.BigEndian.Uint32(checksum)) % total

	cumulativeWeight := 0
	keys := sortedKeys(d.weights)
	for _, key := range keys {
		cumulativeWeight += d.weights[key]
		if randomNum < cumulativeWeight {
			return items[i]
		}
	}
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

// func (d *Distributer) applyWeights() {
// 	paths := d.sortedIds()
// 	if len(paths) == 0 {
// 		return
// 	}
// 	pathIdx := 0
// 	slotsLeft := d.numSlotsForPath(get(paths, pathIdx))

// 	for i := 0; i <= 255; i++ {
// 		d.dist[key{byte(i)}] = get(paths, pathIdx)
// 		slotsLeft--
// 		if slotsLeft == 0 {
// 			pathIdx++
// 			slotsLeft = d.numSlotsForPath(get(paths, pathIdx))
// 		}
// 	}
// }

// func get(paths Slice, idx int) model.NodeId {
// 	if len(paths) <= 0 {
// 		return ""
// 	}

// 	if idx >= len(paths) {
// 		return paths[len(paths)-1]
// 	}

// 	return paths[idx]
// }

// func (d *Distributer) numSlotsForPath(p model.NodeId) int {
// 	weight := d.weights[p]
// 	totalWeight := d.totalWeights()
// 	return weight * 256 / totalWeight
// }

// func (d *Distributer) totalWeights() int {
// 	total := 0
// 	for _, weight := range d.weights {
// 		total += weight
// 	}
// 	return total
// }

// func (d *Distributer) sortedIds() Slice {
// 	ids := make(Slice, 0)
// 	for k := range d.weights {
// 		ids = append(ids, k)
// 	}
// 	sort.Sort(ids)
// 	return ids
// }

// type key struct {
// 	value byte
// }

// func (k key) next() (bool, key) {
// 	if k.value == 0xFF {
// 		return false, key{}
// 	}
// 	return true, key{k.value + 1}
// }

// type Slice []model.NodeId

// func (p Slice) Len() int           { return len(p) }
// func (p Slice) Less(i, j int) bool { return p[i] < p[j] }
// func (p Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

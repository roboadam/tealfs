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
	"hash"
	"hash/crc32"
	"sort"
	"tealfs/pkg/model"
)

type Distributer struct {
	dist     map[key]model.Id
	weights  map[model.Id]int
	checksum hash.Hash32
}

func New() Distributer {
	return Distributer{
		dist:     make(map[key]model.Id),
		weights:  make(map[model.Id]int),
		checksum: crc32.NewIEEE(),
	}
}

func (d *Distributer) NodeIdForStoreId(id model.BlockId) model.Id {
	idb := []byte(id)
	sum := d.Checksum(idb)
	k := key{value: sum[0]}
	return d.dist[k]
}

func (d *Distributer) Checksum(data []byte) []byte {
	d.checksum.Reset()
	d.checksum.Write(data)
	return d.checksum.Sum(nil)
}

func (d *Distributer) SetWeight(id model.Id, weight int) {
	d.weights[id] = weight
	d.applyWeights()
}

func (d *Distributer) PrintDist() {
	for i := 0; i <= 255; i++ {
		println("byteIdx:", i, ", nodeId:", d.dist[key{byte(i)}])
	}
}

func (d *Distributer) applyWeights() {
	paths := d.sortedIds()
	if len(paths) == 0 {
		return
	}
	pathIdx := 0
	slotsLeft := d.numSlotsForPath(get(paths, pathIdx))

	for i := 0; i <= 255; i++ {
		d.dist[key{byte(i)}] = get(paths, pathIdx)
		slotsLeft--
		if slotsLeft == 0 {
			pathIdx++
			slotsLeft = d.numSlotsForPath(get(paths, pathIdx))
		}
	}
}

func get(paths Slice, idx int) model.Id {
	if len(paths) <= 0 {
		return ""
	}

	if idx >= len(paths) {
		return paths[len(paths)-1]
	}

	return paths[idx]
}

func (d *Distributer) numSlotsForPath(p model.Id) int {
	weight := d.weights[p]
	totalWeight := d.totalWeights()
	return weight * 256 / totalWeight
}

func (d *Distributer) totalWeights() int {
	total := 0
	for _, weight := range d.weights {
		total += weight
	}
	return total
}

func (d *Distributer) sortedIds() Slice {
	ids := make(Slice, 0)
	for k := range d.weights {
		ids = append(ids, k)
	}
	sort.Sort(ids)
	return ids
}

type key struct {
	value byte
}

func (k key) next() (bool, key) {
	if k.value == 0xFF {
		return false, key{}
	}
	return true, key{k.value + 1}
}

type Slice []model.Id

func (p Slice) Len() int           { return len(p) }
func (p Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

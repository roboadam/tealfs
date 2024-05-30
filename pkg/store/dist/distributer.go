package dist

import (
	"hash"
	"hash/crc32"
	"sort"
	"tealfs/pkg/nodes"
	"tealfs/pkg/store"
)

type Distributer struct {
	dist     map[key]nodes.Id
	weights  map[nodes.Id]int
	checksum hash.Hash32
}

func New() Distributer {
	return Distributer{
		dist:     make(map[key]nodes.Id),
		weights:  make(map[nodes.Id]int),
		checksum: crc32.NewIEEE(),
	}
}

func (d *Distributer) NodeIdForStoreId(id store.Id) nodes.Id {
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

func (d *Distributer) SetWeight(id nodes.Id, weight int) {
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

func get(paths nodes.Slice, idx int) nodes.Id {
	if len(paths) <= 0 {
		return ""
	}

	if idx >= len(paths) {
		return paths[len(paths)-1]
	}

	return paths[idx]
}

func (d *Distributer) numSlotsForPath(p nodes.Id) int {
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

func (d *Distributer) sortedIds() nodes.Slice {
	ids := make(nodes.Slice, 0)
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

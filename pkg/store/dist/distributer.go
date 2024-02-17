package dist

import (
	"sort"
	h "tealfs/pkg/hash"
	"tealfs/pkg/nodes"
)

type Distributer struct {
	dist    map[key]nodes.Id
	weights map[nodes.Id]int
}

func NewDistributer() *Distributer {
	return &Distributer{
		dist:    make(map[key]nodes.Id),
		weights: make(map[nodes.Id]int),
	}
}

func (d *Distributer) NodeIdForHash(hash h.Hash) nodes.Id {
	k := key{value: hash.Value[0]}
	return d.dist[k]
}

func (d *Distributer) SetWeight(id nodes.Id, weight int) {
	d.weights[id] = weight
	d.applyWeights()
}

func (d *Distributer) PrintDist() {
	for i := 0; i <= 255; i++ {
		println("byteIdx:", i, ", nodeId:", d.dist[key{byte(i)}].String())
	}
}

func (d *Distributer) applyWeights() {
	paths := d.sortedPaths()
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
		return nodes.Id{}
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

func (d *Distributer) sortedPaths() nodes.Slice {
	paths := make(nodes.Slice, 0)
	for k := range d.weights {
		paths = append(paths, k)
	}
	sort.Sort(paths)
	return paths
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

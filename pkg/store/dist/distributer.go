package dist

import (
	"sort"
	h "tealfs/pkg/hash"
	"tealfs/pkg/model/node"
)

type Distributer struct {
	dist    map[key]node.Id
	weights map[node.Id]int
}

func NewDistributer() *Distributer {
	return &Distributer{
		dist:    make(map[key]node.Id),
		weights: make(map[node.Id]int),
	}
}

func (d *Distributer) NodeIdForHash(hash h.Hash) node.Id {
	k := key{value: hash.Value[0]}
	return d.dist[k]
}

func (d *Distributer) SetWeight(id node.Id, weight int) {
	d.weights[id] = weight
	d.applyWeights()
}

func (d *Distributer) applyWeights() {
	paths := d.sortedPaths()
	if len(paths) == 0 {
		return
	}
	pathIdx := 0
	slotsLeft := d.numSlotsForPath(get(paths, pathIdx))

	for i := byte(0); i <= byte(255); i++ {
		d.dist[key{i}] = get(paths, pathIdx)
		slotsLeft--
		if slotsLeft == 0 {
			pathIdx++
			slotsLeft = d.numSlotsForPath(get(paths, pathIdx))
		}
	}
}

func get(paths node.Slice, idx int) node.Id {
	if len(paths) <= 0 {
		return node.Id{}
	}

	if idx >= len(paths) {
		return paths[len(paths)-1]
	}

	return paths[idx]
}

func (d *Distributer) numSlotsForPath(p node.Id) byte {
	weight := d.weights[p]
	totalWeight := d.totalWeights()
	return byte(weight * 256 / totalWeight)
}

func (d *Distributer) totalWeights() int {
	total := 0
	for _, weight := range d.weights {
		total += weight
	}
	return total
}

func (d *Distributer) sortedPaths() node.Slice {
	paths := make(node.Slice, len(d.weights))
	for key, i := range d.weights {
		paths[i] = key
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

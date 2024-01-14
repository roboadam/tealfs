package dist

import (
	"sort"
	"tealfs/pkg/store"
)

type Distributer struct {
	dist    map[key]store.PathId
	weights map[store.PathId]int
}

func (d *Distributer) SetWeight(id store.PathId, weight int) {
	d.weights[id] = weight
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

func get(paths store.PathSlice, idx int) store.PathId {
	if len(paths) <= 0 {
		return store.PathId{}
	}

	if idx >= len(paths) {
		return paths[len(paths)-1]
	}

	return paths[idx]
}

func (d *Distributer) numSlotsForPath(p store.PathId) byte {
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

func (d *Distributer) sortedPaths() store.PathSlice {
	paths := make(store.PathSlice, len(d.weights))
	for key := range d.weights {
		paths = append(paths, key)
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

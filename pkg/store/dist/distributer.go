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

func (d *Distributer) applyWeights(id store.PathId, weight int) {
	numSlotAssinments := make(map[store.PathId]int)
	totalWeight := d.totalWeights()
	weightsLeft := d.totalWeights()
	paths := d.sortedPaths()

	for _, p := range paths {
		slots :=
	}
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
	value [2]byte
}

func (k key) next() (bool, key) {
	if k.value[1] != 0xFF {
		return false, key{value: [2]byte{k.value[0], k.value[1] + 0x01}}
	}
	if k.value[0] != 0xFF {
		return false, key{value: [2]byte{k.value[0] + 0x01, 0x00}}
	}
	return true, key{}
}

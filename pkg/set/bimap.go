package set

type Bimap[K comparable, J comparable] struct {
	dataKj map[K]J
	dataJk map[J]K
}

func NewBimap[K comparable, J comparable]() Bimap[K, J] {
	return Bimap[K, J]{
		dataKj: make(map[K]J),
		dataJk: make(map[J]K),
	}
}

func (b *Bimap[K, J]) Add(item1 K, item2 J) {
	b.dataKj[item1] = item2
	b.dataJk[item2] = item1
}

func (b *Bimap[K, J]) Remove1(item K) {
	delete(b.dataKj, item)
}

func (b *Bimap[K, J]) Remove2(item J) {
	delete(b.dataJk, item)
}

func (b *Bimap[K, J]) Get1(item K) (J, bool) {
	value, ok := b.dataKj[item]
	return value, ok
}

func (b *Bimap[K, J]) Get2(item J) (K, bool) {
	value, ok := b.dataJk[item]
	return value, ok
}

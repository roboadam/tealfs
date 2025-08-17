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

package set

import "maps"

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

func NewBimapFromMap[K comparable, J comparable](input map[K]J) Bimap[K, J] {
	result := NewBimap[K, J]()

	for k, j := range input {
		result.Add(k, j)
	}

	return result
}

func (b *Bimap[K, J]) Len() int {
	return len(b.dataKj)
}

func (b *Bimap[K, J]) ToMap() map[K]J {
	result := make(map[K]J)
	maps.Copy(result, b.dataKj)
	return result
}

func (b *Bimap[K, J]) Clear() {
	b.dataJk = make(map[J]K)
	b.dataKj = make(map[K]J)
}

func (b *Bimap[K, J]) Add(item1 K, item2 J) {
	b.Remove1(item1)
	b.dataKj[item1] = item2
	b.dataJk[item2] = item1
}

func (b *Bimap[K, J]) Remove1(item K) {
	item2 := b.dataKj[item]
	delete(b.dataKj, item)
	delete(b.dataJk, item2)
}

func (b *Bimap[K, J]) Remove2(item J) {
	item2 := b.dataJk[item]
	delete(b.dataJk, item)
	delete(b.dataKj, item2)
}

func (b *Bimap[K, J]) Get1(item K) (J, bool) {
	value, ok := b.dataKj[item]
	return value, ok
}

func (b *Bimap[K, J]) Get2(item J) (K, bool) {
	value, ok := b.dataJk[item]
	return value, ok
}

func (b *Bimap[K, J]) AllValues() []struct {
	K K
	J J
} {
	result := []struct {
		K K
		J J
	}{}
	for k, j := range b.dataKj {
		result = append(result, struct {
			K K
			J J
		}{K: k, J: j})
	}
	return result
}

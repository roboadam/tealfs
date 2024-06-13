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

// Copyright (C) 2026 Adam Hess
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

import "sync"

type OtM[K comparable, J comparable] struct {
	mux       *sync.RWMutex
	oneToMany map[K]Set[J]
	manyToOne map[J]K
}

func NewOtM[K comparable, J comparable]() OtM[K, J] {
	return OtM[K, J]{
		mux:       &sync.RWMutex{},
		oneToMany: make(map[K]Set[J]),
		manyToOne: make(map[J]K),
	}
}

func (b *OtM[K, J]) Add(k1 K, j1 J) {
	b.mux.Lock()
	defer b.mux.Unlock()

	if _, ok := b.oneToMany[k1]; !ok {
		b.oneToMany[k1] = NewSet[J]()
	}

	s := b.oneToMany[k1]
	s.Add(j1)
	b.oneToMany[k1] = s

	b.manyToOne[j1] = k1
}

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
	mux       sync.Mutex
	oneToMany map[K]Set[J]
	manyToOne map[J]K
}

func (o *OtM[K, J]) initOtM() {
	if o.oneToMany == nil {
		o.oneToMany = make(map[K]Set[J])
	}
	if o.manyToOne == nil {
		o.manyToOne = make(map[J]K)
	}
}

func (o *OtM[K, J]) GetKey(j J) (K, bool) {
	o.mux.Lock()
	defer o.mux.Unlock()

	k, ok := o.manyToOne[j]
	return k, ok
}

func (o *OtM[K, J]) Add(k1 K, j1 J) {
	o.mux.Lock()
	defer o.mux.Unlock()
	o.initOtM()

	if _, ok := o.oneToMany[k1]; !ok {
		o.oneToMany[k1] = NewSet[J]()
	}

	s := o.oneToMany[k1]
	s.Add(j1)
	o.oneToMany[k1] = s

	o.manyToOne[j1] = k1
}

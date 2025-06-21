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

type Set[K comparable] struct {
	data map[K]bool
}

func NewSet[K comparable]() Set[K] {
	return Set[K]{data: make(map[K]bool)}
}

func NewSetFromSlice[K comparable](input []K) Set[K] {
	result := NewSet[K]()

	for _, item := range input {
		result.Add(item)
	}

	return result
}

func NewSetFromMapKeys[K comparable, J any](input map[K]J) Set[K] {
	result := Set[K]{data: make(map[K]bool)}

	for k := range input {
		result.Add(k)
	}

	return result
}

func (s *Set[K]) ToSlice() []K {
	result := make([]K, len(s.data))
	i := 0
	for k := range s.data {
		result[i] = k
		i++
	}
	return result
}

func (s *Set[K]) Add(item K) {
	s.data[item] = true
}

func (s *Set[K]) Remove(item K) {
	delete(s.data, item)
}

func (s *Set[K]) Equal(b *Set[K]) bool {
	if len(s.data) != len(b.data) {
		return false
	}

	for key := range s.data {
		if !b.Contains(key) {
			return false
		}
	}

	return true
}

func (s *Set[K]) Contains(k K) bool {
	_, exists := s.data[k]
	return exists
}

func (s *Set[K]) GetValues() []K {
	result := make([]K, len(s.data))
	i := 0
	for k := range s.data {
		result[i] = k
		i++
	}
	return result
}

func (s *Set[K]) Minus(o *Set[K]) *Set[K] {
	result := NewSet[K]()
	for k := range s.data {
		if !o.Contains(k) {
			result.Add(k)
		}
	}
	return &result
}

func (s *Set[K]) Len() int {
	return len(s.data)
}

func (s *Set[K]) Clone() Set[K] {
	return Set[K]{
		data: maps.Clone(s.data),
	}
}

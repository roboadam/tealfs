package set

import "maps"

type Set[K comparable] struct {
	data map[K]bool
}

func NewSet[K comparable]() Set[K] {
	return Set[K]{data: make(map[K]bool)}
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
		if !b.Exists(key) {
			return false
		}
	}

	return true
}

func (s *Set[K]) Exists(k K) bool {
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
		if !o.Exists(k) {
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

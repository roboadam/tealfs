package util

type Set[K comparable] struct {
	data map[K]bool
}

func NewSet[K comparable]() Set[K] {
	return Set[K]{data: make(map[K]bool)}
}

func (s *Set[K]) Add(item K) {
	s.data[item] = true
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

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

func (m *Set[K]) GetValues() []K {
	result := make([]K, len(m.data))
	i := 0
	for k := range m.data {
		result[i] = k
		i++
	}
	return result
}

func (m *Set[K]) Minus(o *Set[K]) *Set[K] {
	result := NewSet[K]()
	for k := range m.data {
		if !o.Exists(k) {
			result.Add(k)
		}
	}
	return &result
}

func (m *Set[K]) Len() int {
	return len(m.data)
}

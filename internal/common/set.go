package common

type Set[T comparable] struct {
	values map[T]struct{}
}

func (s *Set[T]) Contains(v T) bool {
	_, ok := s.values[v]
	return ok
}

func NewSet[T comparable](values []T) *Set[T] {
	s := &Set[T]{
		values: make(map[T]struct{}, len(values)),
	}
	for _, v := range values {
		s.values[v] = struct{}{}
	}
	return s
}

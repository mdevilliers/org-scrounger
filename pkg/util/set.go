package util

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

func NewSet[T constraints.Ordered]() Set[T] {
	return Set[T]{}
}

type Set[T constraints.Ordered] map[T]int

func (s Set[T]) Add(c T) {
	_, found := s[c]
	if found {
		s[c] = s[c] + 1
	} else {
		s[c] = 1
	}
}

func (s Set[T]) Keys() []T {
	r := make([]T, len(s))
	for k, _ := range s {
		r = append(r, k)
	}
	return r
}

func (s Set[T]) OrderedKeys() []T {
	r := s.Keys()
	slices.Sort(r)
	return r
}

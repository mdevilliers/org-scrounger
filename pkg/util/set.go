package util

import (
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

// NewSet returns a Set
func NewSet[T constraints.Ordered]() Set[T] {
	return Set[T]{}
}

// Set tracks items added keeping the total
// number of each individual item added.
type Set[T constraints.Ordered] map[T]int

// Add
func (s Set[T]) Add(c T) {
	_, found := s[c]
	if found {
		s[c]++
	} else {
		s[c] = 1
	}
}

func (s Set[T]) Keys() []T {
	r := []T{}
	for k := range s {
		r = append(r, k)
	}
	return r
}

func (s Set[T]) OrderedKeys() []T {
	r := s.Keys()
	slices.Sort(r)
	return r
}

func (s Set[T]) Join(x Set[T]) Set[T] {

	for k := range x {
		_, found := s[k]
		if found {
			s[k] += x[k]
		} else {
			s[k] = x[k]
		}
	}
	return s
}

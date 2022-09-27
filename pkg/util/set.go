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

// Add an item, incrementing the count
func (s Set[T]) Add(c T) {
	_, found := s[c]
	if found {
		s[c]++
	} else {
		s[c] = 1
	}
}

// Keys returns a slice of keys
func (s Set[T]) Keys() []T {
	r := []T{}
	for k := range s {
		r = append(r, k)
	}
	return r
}

// OrderedKeys returns the keys ordered in ascending order
func (s Set[T]) OrderedKeys() []T {
	r := s.Keys()
	slices.Sort(r)
	return r
}

// Join returns the superset of two Sets
func (s Set[T]) Join(x Set[T]) Set[T] {

	dst := NewSet[T]()
	for k := range s {
		dst[k] = s[k]
	}

	for k := range x {
		_, found := dst[k]
		if found {
			dst[k] += x[k]
		} else {
			dst[k] = x[k]
		}
	}
	return dst
}

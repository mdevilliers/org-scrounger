package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Set(t *testing.T) {

	s := NewSet[string]()
	s.AddWithValue("foo", 123)
	s.AddWithValue("bar", 456)
	s.AddWithValue("bar", 1)

	v, n := s.TopValue()

	require.Equal(t, 457, n)
	require.Equal(t, "bar", v)

}

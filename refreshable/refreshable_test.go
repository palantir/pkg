package refreshable

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpand(t *testing.T) {
	strs := New([]string{"hello", "world"})
	expanded := Expand(strs, func(s string) []byte { return []byte(s) })

	// Verify we get 2 refreshables
	require.Len(t, expanded, 2)

	// Verify initial content
	require.Equal(t, []byte("hello"), expanded[0].Current())
	require.Equal(t, []byte("world"), expanded[1].Current())

	// Update the slice
	strs.Update([]string{"foo", "bar"})
	require.Equal(t, []byte("foo"), expanded[0].Current())
	require.Equal(t, []byte("bar"), expanded[1].Current())
}

type ExpandWrapper struct {
	Refreshables []Refreshable[[]byte]
}

// Expand takes a Refreshable of a slice and a mapper function, returning a slice of Refreshables,
// one for each element, that updates when the input slice changes.
func Expand(strSlice Refreshable[[]string], mapFn func(string) []byte) []Refreshable[[]byte] {
	current := strSlice.Current()
	result := make([]Refreshable[[]byte], len(current))
	outputs := make([]*defaultRefreshable[[]byte], len(current))

	for i := range current {
		outputs[i] = newZero[[]byte]()
		result[i] = outputs[i].readOnly()
	}

	strSlice.Subscribe(func(strs []string) {
		for i := range outputs {
			if i < len(strs) {
				outputs[i].Update(mapFn(strs[i]))
			} else {
				outputs[i].Update(nil)
			}
		}
	})

	return result
}

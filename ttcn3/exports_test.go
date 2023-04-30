package ttcn3

import (
	"github.com/nokia/ntt/ttcn3/syntax"
)

// SliceAt returns the slice of nodes at the given position.
func (t *Tree) SliceAt(pos int) []syntax.Node {
	return t.sliceAt(pos)
}

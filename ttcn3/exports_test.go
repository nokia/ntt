package ttcn3

import (
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/ttcn3/syntax"
)

// SliceAt returns the slice of nodes at the given position.
func (t *Tree) SliceAt(pos loc.Pos) []syntax.Node {
	return t.sliceAt(pos)
}

func (t *Tree) FileSet() *loc.FileSet {
	return t.fset
}

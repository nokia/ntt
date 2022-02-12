package ttcn3_test

import (
	"testing"

	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/stretchr/testify/assert"
)

func TestSliceAt(t *testing.T) {
	tree := parseFile(t, t.Name(), `module M {
function func<type T>(T x) {
	{
		T
	}
}

}`)

	s := tree.SliceAt(tree.Pos(4, 3))
	assert.Equal(t, "T", s[0].(ast.Token).Lit)

}

package ttcn3

import (
	"testing"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/stretchr/testify/assert"
)

func TestSliceAt(t *testing.T) {
	tree := parse(`module M {
function func<T>(T x) {
	{
		T
	}
}

}`)

	s := tree.SliceAt(tree.Pos(4, 3))
	assert.Equal(t, "T", s[0].(ast.Token).Lit)

}

func parse(src string) *Tree {
	file := "test://test.ttcn3"
	fs.Open(file).SetBytes([]byte(src))
	return ParseFile(file)
}

package ttcn3_test

import (
	"testing"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/stretchr/testify/assert"
)

func TestSliceAt(t *testing.T) {
	tree := testParse(`module M {
function func<T>(T x) {
	{
		T
	}
}

}`)

	s := tree.SliceAt(tree.Pos(4, 3))
	assert.Equal(t, "T", s[0].(ast.Token).Lit)

}

func testParse(src string) *ttcn3.Tree {
	file := "test://test.ttcn3"
	fs.SetContent(file, []byte(src))
	return ttcn3.ParseFile(file)
}

package ttcn3

import (
	"fmt"
	"testing"

	"github.com/nokia/ntt/internal/fs"
)

func TestSliceAt(t *testing.T) {
	tree := parse(`module M {
function func<T>(T x) {
	{
		T
	}
}

}`)
	t1 := tree.SliceAt(4, 3)
	for i := range t1.nodes {
		fmt.Printf("--------->%#v\n\n\n", t1.nodes[i])
	}

}

func parse(src string) *Tree {
	file := "test://test.ttcn3"
	fs.Open(file).SetBytes([]byte(src))
	return ParseFile(file)
}

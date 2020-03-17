package ntt_test

import (
	"testing"

	"github.com/nokia/ntt/internal/ntt"
)

func TestDefine(t *testing.T) {
	suite := &ntt.Suite{}
	suite.AddSources("a.ttcn3")
	src := suite.File("a.ttcn3")
	src.SetBytes([]byte(`
module foo {
	const integer C := 23

	function f() return integer {
		var integer x := C
		log(C, x)
	}
}
	`))

	tree, fset, _ := suite.Parse(src)
	_ = suite.Define(fset, tree)
}

/*
	var x<a,b>[-].bla v;
*/

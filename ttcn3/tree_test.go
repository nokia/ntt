package ttcn3_test

import (
	"fmt"
	"testing"

	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/stretchr/testify/assert"
)

func TestSliceAt(t *testing.T) {
	input := `module M
		  {
			function func<type T>(T x)
			{
				{
					¶T := x;
				}
			}
		  }`

	cursor, source := extractCursor(input)
	tree := parseFile(t, t.Name(), source)

	var actual []string
	for _, n := range tree.SliceAt(cursor) {
		s := fmt.Sprintf("%T", n)
		if n := ast.Name(n); n != "" {
			s += fmt.Sprintf("(%s)", n)
		}
		actual = append(actual, s)
	}

	expected := []string{
		"*ast.Ident(T)",
		"*ast.BinaryExpr",
		"*ast.ExprStmt",
		"*ast.BlockStmt",
		"*ast.BlockStmt",
		"*ast.FuncDecl(func)",
		"*ast.ModuleDef(func)",
		"*ast.Module(M)",
		"ast.NodeList",
	}
	assert.Equal(t, expected, actual)
}

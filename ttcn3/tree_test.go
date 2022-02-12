package ttcn3_test

import (
	"testing"

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
		actual = append(actual, nodeDesc(n))
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

func TestExprAt(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "<nil>"},
		{"x", "<nil>"},
		{"¶", "<nil>"},
		{"x¶", "<nil>"},
		{"¶x", "*ast.Ident(x)"},
		{"¶x[-]", "*ast.Ident(x)"},
		{"¶x[-][-]", "*ast.Ident(x)"},
		{"¶x()[-]", "*ast.Ident(x)"},
		{"¶x.y.z", "*ast.Ident(x)"},
		{"x¶.y.z", "<nil>"},
		{"x.¶y.z", "*ast.SelectorExpr(x.y)"},
		{"x.y.¶z", "*ast.SelectorExpr(x.y.z)"},
		{"x[-].¶y.z", "*ast.SelectorExpr(.y)"},
		{"foo(23, ¶x:= 1)", "<nil>"}, // not supported yet
		{"a := { ¶x:= 1}", "<nil>"},  // not supported yet
		{"a := { [¶x]:= 1}", "*ast.Ident(x)"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cursor, source := extractCursor(tt.input)
			tree := parseFile(t, t.Name(), source)
			actual := nodeDesc(tree.ExprAt(cursor))
			assert.Equal(t, tt.want, actual)
		})
	}
}

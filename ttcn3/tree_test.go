package ttcn3_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceAt(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{
			input: `module M2 {} module M1 {import from ¶M2 all}`,
			want: []string{
				"*ast.Ident(M2)",
				"*ast.ImportDecl",
				"*ast.ModuleDef",
				"*ast.Module(M1)",
				"ast.NodeList",
			}},
		{

			input: `module M {
				  function func<type T>(T x) {
				    while (true) { ¶T := x; }
				  }
		  		}`,
			want: []string{
				"*ast.Ident(T)",
				"*ast.BinaryExpr",
				"*ast.ExprStmt",
				"*ast.BlockStmt",
				"*ast.WhileStmt",
				"*ast.BlockStmt",
				"*ast.FuncDecl(func)",
				"*ast.ModuleDef(func)",
				"*ast.Module(M)",
				"ast.NodeList",
			}},
	}

	for _, tt := range tests {
		cursor, source := extractCursor(tt.input)
		tree := parseFile(t, t.Name(), source)

		var actual []string
		for _, n := range tree.SliceAt(cursor) {
			actual = append(actual, nodeDesc(n))
		}

		assert.Equal(t, tt.want, actual)
	}
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

package syntax_test

import (
	"testing"

	"github.com/nokia/ntt/internal/ntttest"
	"github.com/nokia/ntt/ttcn3/syntax"
)

func TestFindChildOf(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "", want: ""},
		{input: "¶", want: ""},
		{input: "¶x", want: "x"},
		{input: "x¶", want: ""},
		{input: "¶x,y", want: "x"},
		{input: "x,¶y", want: "y"},
		{input: "x,y,z¶", want: ""},
		{input: "x,y,¶z", want: "z"},
		{input: "x,y¶,z", want: ""},
		{input: "x,¶y,z", want: "y"},
		{input: "x¶,y,z", want: ""},
		{input: "¶x,y,z", want: "x"},
	}
	for _, tt := range tests {
		input, cursor := ntttest.CutCursor(tt.input)
		root, _, _ := syntax.Parse([]byte(input), syntax.WithFilename(tt.input))
		actual := printNode(syntax.FindChildOf(root, cursor))
		if actual != tt.want {
			t.Errorf("FindChildOf(%q) = %q, want %q", tt.input, actual, tt.want)
		}
	}
}

func printNode(n syntax.Node) string {
	switch n := n.(type) {
	case *syntax.ExprStmt:
		return printNode(n.Expr)
	default:
		return syntax.Name(n)
	}
}

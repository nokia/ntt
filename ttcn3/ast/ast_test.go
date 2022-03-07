package ast_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
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
		cursor, input := extractCursor(tt.input)
		tree := parseFile(t, "test", input)
		actual := printNode(ast.FindChildOf(tree.Root, cursor))
		if actual != tt.want {
			t.Errorf("FindChildOf(%q) = %q, want %q", tt.input, actual, tt.want)
		}
	}
}

const CURSOR = "¶"

func extractCursor(input string) (loc.Pos, string) {
	return loc.Pos(strings.Index(input, CURSOR) + 1), strings.Replace(input, CURSOR, "", 1)
}

func parseFile(t *testing.T, name string, input string) *ttcn3.Tree {
	t.Helper()
	file := fmt.Sprintf("%s.ttcn3", name)
	fs.SetContent(file, []byte(input))
	tree := ttcn3.ParseFile(file)
	if tree.Err != nil {
		t.Fatalf("%s", tree.Err.Error())
	}
	return tree
}

func printNode(n ast.Node) string {
	switch n := n.(type) {
	case *ast.ExprStmt:
		return printNode(n.Expr)
	default:
		return ast.Name(n)
	}
}

package ttcn3_test

import (
	"testing"

	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
)

// TestScope verifies that scopes are built and populated correctly.
func TestScopes(t *testing.T) {
	tests := []struct {
		input string
		names []string
	}{
		{`{var int x := x}`, []string{"x"}},
		{`{var int x := x; {var int y := x}}`, []string{"x"}},
		{`template int t<type T>(int x) := 1`, []string{"T", "x"}},
		{`function f<type T>(int p) {var int x}`, []string{"T", "p", "x"}},
		{`type record { int x } R<type T>`, []string{"x", "T"}},
		{`type record { record { int y } x } R<type T>`, []string{"x", "T"}},
		{`type record of record { int x } R`, []string{}},
	}

	for _, tt := range tests {
		tree := ttcn3.Parse(tt.input)
		scp := ttcn3.NewScope(unwrapFirst(tree.Root), tree)
		if scp == nil {
			t.Errorf("%q: scope is nil", tt.input)
			continue
		}
		if actual := nameSlice(scp); !equal(actual, tt.names) {
			t.Errorf("%q: expected %v names, got %v", tt.input, tt.names, actual)
		}
	}
}

// Unwrap first node from NodeLists
func unwrapFirst(n ast.Node) ast.Node {
	switch n := n.(type) {
	case ast.NodeList:
		if len(n) == 0 {
			return nil
		}
		return unwrapFirst(n[0])
	case *ast.ExprStmt:
		return unwrapFirst(n.Expr)
	case *ast.DeclStmt:
		return unwrapFirst(n.Decl)
	case *ast.ModuleDef:
		return unwrapFirst(n.Def)
	default:
		return n
	}
}

// equal returns true if a and b are equal, order is ignored.
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int, len(a))
	for i := range a {
		m[a[i]]++
		m[b[i]]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}

func nameSlice(scp *ttcn3.Scope) []string {
	s := make([]string, 0, len(scp.Names))
	for k := range scp.Names {
		s = append(s, k)
	}
	return s
}

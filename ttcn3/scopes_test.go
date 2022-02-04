package ttcn3_test

import (
	"testing"

	"github.com/nokia/ntt/ttcn3"
	"github.com/nokia/ntt/ttcn3/ast"
)

func TestScopes(t *testing.T) {
	tests := []struct {
		input string
		names []string
	}{
		{`{var int x := x}`, []string{"x"}},
		{`{var int x := x; {var int y := x}}`, []string{"x"}},
	}
	for _, tt := range tests {
		tree := ttcn3.Parse(tt.input)
		scp := ttcn3.NewScope(unwrapFirst(tree.Root))
		if scp == nil {
			t.Errorf("%q: scope is nil", tt.input)
			continue
		}
		if !equal(nameKeys(scp.Names), tt.names) {
			t.Errorf("Expected %d names, got %d", len(tt.names), len(scp.Names))
		}
	}
}

// Unwrap first node from NodeLists
func unwrapFirst(n ast.Node) ast.Node {
	if n, ok := n.(ast.NodeList); ok && len(n) > 0 {
		return n[0]
	}
	return n
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

func nameKeys(m map[string]*ast.Ident) []string {
	s := make([]string, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	return s
}

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
		{`signature S<type T>(int x) return A exception(B)`, []string{"T", "x"}},
		{`type union U<type T>{ int x }`, []string{"T", "x"}},
		{`type enumerated E<type T>{ E1, E2 }`, []string{"T", "E1", "E2"}},
		{`type function F<type T>(int p) runs on C return X`, []string{"T", "p"}},
		{`type port P<type T}> message { inout T; map param(int p)}`, []string{"T"}},
		{`type component C<type T> extends D {var T x}`, []string{"T", "x"}},
		{`for (var int i;true;i:=i+1) {var int x}`, []string{"i", "x"}},
		{`while (true) {var int x}`, []string{"x"}},
		{`do {var int x} while (true){`, []string{"x"}},
		{`if (true) {var int x} else {var int y}`, []string{}},
		{`
			module M {
				import from foo all;
				friend module bar;
				group G { group G2 {
				var int x;
				}}
				control {};
				type enumerated E<type T> {E1}
			}`, []string{"foo", "x", "control", "E", "E1"}},
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

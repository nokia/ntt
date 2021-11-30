package types_test

import (
	"testing"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/parser"
	"github.com/nokia/ntt/types"
)

func TestValueDecl(t *testing.T) {
	tests := []struct {
		input string
		want  types.Object
	}{
		{"var integer x", &types.Var{Name: "x", Type: nil}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			scp := makeScope(t, tt.input)
			if got := scp.last(); types.Equal(got, tt.want) != true {
				t.Fatalf("InsertTree(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func makeScope(t *testing.T, input string) *testScope {
	t.Helper()

	scp := testScope{t: t}

	info := &types.Info{}
	if err := info.InsertTree(parse(t, input), &scp); err != nil {
		t.Fatalf("InsertTree: %v", err)
	}

	return &scp

}

func parse(t *testing.T, input string) ast.Node {
	t.Helper()

	fset := loc.NewFileSet()
	root, err := parser.Parse(fset, "", input)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return root
}

type pair struct {
	name string
	obj  types.Object
}

type testScope struct {
	t     *testing.T
	pairs []pair
	names map[string]pair
}

func (s *testScope) Lookup(name string) types.Object {
	if p, ok := s.names[name]; ok {
		return p.obj
	}
	return nil
}

func (s *testScope) Insert(name string, obj types.Object) types.Object {
	if s.names == nil {
		s.names = make(map[string]pair)
	}
	if _, ok := s.names[name]; ok {
		s.t.Fatalf("%s: already defined", name)
	}

	p := pair{name, obj}
	s.names[name] = p
	s.pairs = append(s.pairs, p)
	return nil
}

func (s *testScope) Names() []string {
	var names []string
	for _, p := range s.pairs {
		names = append(names, p.name)
	}
	return names
}

func (s *testScope) EnclosingScope() types.Scope {
	return nil
}

func (s *testScope) last() types.Object {
	if len(s.pairs) == 0 {
		return nil
	}
	return s.pairs[len(s.pairs)-1].obj
}

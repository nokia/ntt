package types_test

import (
	"testing"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/parser"
	"github.com/nokia/ntt/types"
)

func makeScope(t *testing.T, input string) (*orderedScope, *types.Info, error) {
	t.Helper()

	fset := loc.NewFileSet()
	root, err := parser.Parse(fset, "", input)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	scp := &orderedScope{t: t}
	info := &types.Info{}
	return scp, info, info.InsertTree(root, scp)
}

type orderedScope struct {
	t     *testing.T
	pairs []pair
	names map[string]pair
}

func (scp *orderedScope) Lookup(name string) types.Object {
	if p, ok := scp.names[name]; ok {
		return p.obj
	}
	scp.t.Fatalf("%scp: not found", name)
	return nil
}

func (scp *orderedScope) Insert(name string, obj types.Object) types.Object {
	if scp.names == nil {
		scp.names = make(map[string]pair)
	}
	if _, ok := scp.names[name]; ok {
		scp.t.Fatalf("%scp: already defined", name)
	}

	p := pair{name, obj}
	scp.names[name] = p
	scp.pairs = append(scp.pairs, p)
	return nil
}

func (scp *orderedScope) Names() []string {
	var names []string
	for _, p := range scp.pairs {
		names = append(names, p.name)
	}
	return names
}

func (scp *orderedScope) EnclosingScope() types.Scope {
	return nil
}

func (scp *orderedScope) last() types.Object {
	if len(scp.pairs) == 0 {
		return nil
	}
	return scp.pairs[len(scp.pairs)-1].obj
}

func (scp *orderedScope) Objects() []types.Object {
	var objs []types.Object
	for _, p := range scp.pairs {
		objs = append(objs, p.obj)
	}
	return objs
}

type pair struct {
	name string
	obj  types.Object
}

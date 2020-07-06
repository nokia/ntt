package types

import (
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

// Module describes a Module.
type Module struct {
	object
	scope

	Imports []string
}

func NewModule(n *ast.Module, name string) *Module {
	return &Module{
		object: object{
			node: n,
			name: name,
		},
	}
}

func (m *Module) Lookup(name string) Object {
	// m.scope.Lookup does not climb up scope chains. When obj != nil we know
	// the scope is m.scope.
	// However we must return m to make sure clients can use type assertions, like
	// 		scp.(*ntt.Module).Name()
	if obj := m.scope.Lookup(name); obj != nil {
		return obj
	}
	return Universe.Lookup(name)
}

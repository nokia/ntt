package types

import (
	"bytes"
	"fmt"

	"github.com/nokia/ntt/internal/ttcn3/ast"
)

type ComponentType struct {
	scope
	object
	Vars  []*Var
	Ports []*Port
}

func NewComponentType(n *ast.ComponentTypeDecl, name string) *ComponentType {
	return &ComponentType{
		object: object{
			node: n,
			name: name,
		},
	}
}

func (c *ComponentType) Insert(obj Object) Object {
	if alt := c.scope.Insert(obj); alt != nil {
		return alt
	}

	switch obj := obj.(type) {
	case *Var:
		c.Vars = append(c.Vars, obj)
	case *Port:
		c.Ports = append(c.Ports, obj)
	default:
		// TODO(5nord) Add error
	}
	return nil
}

func (c *ComponentType) Underlying() Type { return c }
func (c *ComponentType) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "component %s", c.name)
	return buf.String()
}

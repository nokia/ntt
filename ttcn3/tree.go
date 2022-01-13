package ttcn3

import (
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

// Tree represents the TTCN-3 syntax tree, usually of a file.
type Tree struct {
	FileSet *loc.FileSet
	Root    ast.NodeList
	Err     error
}

func (t *Tree) Modules() []*ast.Module {
	var nodes []*ast.Module
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.Module); ok {
			nodes = append(nodes, n)
			return false
		}
		return true
	})
	return nodes
}

func (t *Tree) Funcs() []*ast.FuncDecl {
	var nodes []*ast.FuncDecl
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.FuncDecl); ok {
			nodes = append(nodes, n)
			return false
		}
		return true
	})
	return nodes
}

func (t *Tree) Imports() []*ast.ImportDecl {
	var nodes []*ast.ImportDecl
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.ImportDecl); ok {
			nodes = append(nodes, n)
			return false
		}
		return true
	})
	return nodes
}

func (t *Tree) Ports() []*ast.PortTypeDecl {
	var nodes []*ast.PortTypeDecl
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.PortTypeDecl); ok {
			nodes = append(nodes, n)
			return false
		}
		return true
	})
	return nodes
}

func (t *Tree) Components() []*ast.ComponentTypeDecl {
	var nodes []*ast.ComponentTypeDecl
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.ComponentTypeDecl); ok {
			nodes = append(nodes, n)
			return false
		}
		return true
	})
	return nodes
}

func (t *Tree) Controls() []*ast.ControlPart {
	var nodes []*ast.ControlPart
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.ControlPart); ok {
			nodes = append(nodes, n)
			return false
		}
		return true
	})
	return nodes
}

func (t *Tree) ModulePars() []*ast.Declarator {
	var nodes []*ast.Declarator
	ast.Inspect(t.Root, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.Module, *ast.ModuleDef, *ast.GroupDecl, *ast.ModuleParameterGroup:
			return true

		case *ast.ValueDecl:
			if n.Kind.Kind != token.MODULEPAR && n.Kind.Kind != token.ILLEGAL {
				return false
			}
			return true

		case *ast.Declarator:
			nodes = append(nodes, n)
		}
		return false
	})
	return nodes
}

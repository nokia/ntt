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

// Pos encodes a line and column tuple into a offset-based Pos tag. If file nas
// not been parsed yet, Pos will return loc.NoPos.
func (tree *Tree) Pos(line int, column int) loc.Pos {
	if tree.FileSet == nil {
		return loc.NoPos
	}

	// We asume every FileSet has only one file, to make this work.
	return loc.Pos(int(tree.FileSet.File(loc.Pos(1)).LineStart(line)) + column - 1)
}

func (tree *Tree) SliceAt(line, col int) []ast.Node {
	pos := tree.Pos(line, col)
	return slice(tree.Root, pos)
}

func slice(n ast.Node, pos loc.Pos) []ast.Node {
	var (
		path  []ast.Node
		found bool
	)

	ast.Inspect(n, func(n ast.Node) bool {
		if found {
			return false
		}

		if n == nil {
			path = path[:len(path)-1]
			return false
		}

		path = append(path, n)

		// We don't need to descend any deeper if we're not near desired
		// position.
		if n.End() < pos || pos < n.Pos() {
			return false
		}

		if n.Pos() <= pos && pos <= n.End() {
			found = true
		}

		return !found
	})

	if len(path) == 0 {
		return nil
	}

	// Reverse path so leaf is first element.
	for i := 0; i < len(path)/2; i++ {
		path[i], path[len(path)-1-i] = path[len(path)-1-i], path[i]
	}
	return path
}

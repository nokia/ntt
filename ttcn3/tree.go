package ttcn3

import (
	"reflect"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

// Tree represents the TTCN-3 syntax tree, usually of a file.
type Tree struct {
	FileSet *loc.FileSet
	Root    ast.NodeList
	Names   map[string]bool
	Err     error

	filename string
	parents  map[ast.Node]ast.Node
}

// Filename returns the filename of the file that was parsed.
func (t *Tree) Filename() string {
	return t.filename
}

// ParentOf returns the parent of the given node.
func (t *Tree) ParentOf(n ast.Node) ast.Node {
	if t.parents == nil {
		t.parents = make(map[ast.Node]ast.Node)
		var visit func(n ast.Node)
		visit = func(n ast.Node) {
			for _, c := range ast.Children(n) {
				switch c.(type) {
				case ast.Token, ast.NodeList:
				default:
					t.parents[c] = n
				}
				visit(c)
			}
		}
		visit(t.Root)
	}

	switch n.(type) {
	case ast.Token, ast.NodeList:
		return nil
	}
	return t.parents[n]
}

// Lookup returns the definitions of the given expression. For handling imports
// and multiple modules, use LookupWithDB.
func (tree *Tree) Lookup(n ast.Expr) []*Definition {
	f := &finder{DB: &DB{}, v: make(map[ast.Node]bool)}
	return f.lookup(n, tree)
}

// LookupWithDB returns the definitions of the given expression, but uses the database for import resoltion.
func (tree *Tree) LookupWithDB(n ast.Expr, db *DB) []*Definition {
	f := &finder{DB: db, v: make(map[ast.Node]bool)}
	return f.lookup(n, tree)

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

	// LineStart panics sometimes. We don't know why, yet. So we just log the
	// error and return loc.NoPos to prevent the language server from crashing.
	defer func() {
		if r := recover(); r != nil {
			log.Println("Getting file position failed:", r)
		}
	}()

	// We assume every FileSet has only one file, to make this work.
	return loc.Pos(int(tree.FileSet.File(loc.Pos(1)).LineStart(line)) + column - 1)
}

// Position return a human readable position of the node.
func (tree *Tree) Position(pos loc.Pos) loc.Position {
	if tree.FileSet == nil {
		return loc.Position{}
	}
	return tree.FileSet.Position(pos)
}

// ExprAt returns the primary expression at the given position.
func (tree *Tree) ExprAt(pos loc.Pos) ast.Expr {
	s := tree.SliceAt(pos)
	if len(s) == 0 {
		log.Debugf("%s: no expression at cursor position.\n", tree.Position(pos))
		return nil
	}

	id, ok := s[0].(*ast.Ident)
	if !ok {
		log.Debugf("%s: no identifier at cursor position.\n", tree.Position(pos))
		return nil
	}

	parent := tree.ParentOf(id)
	switch p := parent.(type) {
	case *ast.SelectorExpr:
		if id == p.Sel {
			return p
		}
	case *ast.BinaryExpr:
		if id == p.X && p.Op.Kind == token.ASSIGN {
			q := tree.ParentOf(p)
			switch q := q.(type) {
			case *ast.CompositeLiteral:
				log.Debugf("%s: field assignment not supported.\n",
					tree.Position(pos))
				return nil
			case *ast.ParenExpr:
				if _, ok := tree.ParentOf(q).(*ast.CallExpr); ok {
					log.Debugf("%s: field assignment not supported.\n",
						tree.Position(pos))
					return nil
				}
			}

		}
	}

	return id
}

// SliceAt returns the slice of nodes at the given position.
func (tree *Tree) SliceAt(pos loc.Pos) []ast.Node {
	var (
		path  []ast.Node
		visit func(n ast.Node)
	)

	visit = func(n ast.Node) {
		if v := reflect.ValueOf(n); v.Kind() == reflect.Ptr && v.IsNil() || n == nil {
			return
		}

		if _, ok := n.(ast.Token); ok {
			return
		}

		if inside := n.Pos() <= pos && pos < n.End(); inside {
			path = append(path, n)
			for _, child := range ast.Children(n) {
				visit(child)
			}
		}

	}
	visit(tree.Root)

	// Reverse path so leaf is first element.
	for i := 0; i < len(path)/2; i++ {
		path[i], path[len(path)-1-i] = path[len(path)-1-i], path[i]
	}

	return path
}

type finder struct {
	*DB
	v map[ast.Node]bool
}

func (f *finder) lookup(n ast.Expr, tree *Tree) []*Definition {
	if n == nil {
		return nil
	}

	switch n := n.(type) {
	case *ast.SelectorExpr:
		return f.dot(n, tree)

	case *ast.Ident:
		return f.globals(n, tree)

	case *ast.IndexExpr:
		return f.index(n, tree)

	case *ast.CallExpr:
		return nil
	}

	log.Debugf("%s: Unsupported node type: %T\n", tree.Position(n.Pos()), n)
	return nil
}

func (f *finder) globals(id *ast.Ident, tree *Tree) []*Definition {
	parents := ast.Parents(id, tree.Root)
	if len(parents) > 0 {
		if _, ok := parents[len(parents)-1].(ast.NodeList); ok {
			parents = parents[:len(parents)-1]
		}
	}

	var defs []*Definition
	// Find definitions in current file by walking up the scopes.
	for _, n := range parents {
		f.v[n] = true
		found := Definitions(id.String(), n, tree)
		defs = append(defs, found...)
	}

	// Find definitions in visible files.
	if mod, ok := parents[len(parents)-1].(*ast.Module); ok {
		for _, m := range f.VisibleModules(id.String(), mod) {
			if id.String() == m.Ident.String() {
				defs = append(defs, m)
			}
			if m.Node == mod {
				continue
			}
			defs = append(defs, Definitions(id.String(), m.Node, m.Tree)...)
		}
	}

	return defs
}

func (f *finder) dot(n *ast.SelectorExpr, tree *Tree) []*Definition {
	var result []*Definition
	candidates := f.lookup(n.X, tree)
	for _, c := range candidates {
		for _, t := range f.typeOf(c) {
			if defs := Definitions(ast.Name(n.Sel), t.Node, t.Tree); len(defs) > 0 {
				result = append(result, defs...)
			}
		}
	}
	return result
}

func (f *finder) index(n *ast.IndexExpr, tree *Tree) []*Definition {
	var result []*Definition
	candidates := f.lookup(n.X, tree)
	for _, c := range candidates {
		for _, t := range f.typeOf(c) {
			//fmt.Printf("XXXXX %T\n", t.Node)
			switch n := t.Node.(type) {
			case *ast.Field:
				if l, ok := n.Type.(*ast.ListSpec); ok {
					result = append(result, &Definition{Node: l.ElemType, Tree: t.Tree})
				}
			}
		}
	}
	return result
}
func (f *finder) typeOf(def *Definition) []*Definition {
	if t := def.Type(); t != nil {
		def = t
	}

	if x, ok := def.Node.(ast.Expr); ok {
		return f.lookup(x, def.Tree)
	}

	return []*Definition{def}
}

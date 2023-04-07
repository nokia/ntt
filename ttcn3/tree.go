package ttcn3

import (
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

// Tree represents the TTCN-3 syntax tree, usually of a file.
type Tree struct {
	FileSet *loc.FileSet
	Root    ast.Node
	Names   map[string]bool
	Uses    map[string]bool
	Err     error

	filename  string
	parents   map[ast.Node]ast.Node
	parentsMu sync.Mutex
	scopes    map[ast.Node]*Scope
	scopesMu  sync.Mutex
}

// Filename returns the filename of the file that was parsed.
func (t *Tree) Filename() string {
	return t.filename
}

// ParentOf returns the parent of the given node.
func (t *Tree) ParentOf(n ast.Node) ast.Node {
	t.parentsMu.Lock()
	defer t.parentsMu.Unlock()
	if t.parents == nil {
		t.parents = make(map[ast.Node]ast.Node)
	}
	if p, ok := t.parents[n]; ok {
		return p
	}
	parents := ast.Parents(n, t.Root)
	if len(parents) == 0 {
		t.parents[n] = nil
		return nil
	}
	for _, p := range parents {
		if _, ok := t.parents[n]; ok {
			break
		}
		t.parents[n] = p
		n = p
	}
	return parents[0]
}

// Returns the qualified name of the given node.
func (t *Tree) QualifiedName(n ast.Node) string {
	if name := ast.Name(n); name != "" {
		if mod := t.ModuleOf(n); mod != nil {
			return fmt.Sprintf("%s.%s", ast.Name(mod.Name), name)
		}
		return name
	}
	return ""
}

// ModuleOf returns the module of the given node, by walking up the tree.
func (t *Tree) ModuleOf(n ast.Node) *ast.Module {
	for n := n; n != nil; n = t.ParentOf(n) {
		if m, ok := n.(*ast.Module); ok {
			return m
		}
	}
	return nil
}

func (t *Tree) Modules() []*Node {
	var defs []*Node
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.Module); ok {
			defs = append(defs, &Node{Ident: n.Name, Node: n, Tree: t})
			return false
		}
		return true
	})
	return defs
}

func (t *Tree) Funcs() []*Node {
	var defs []*Node
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.FuncDecl); ok {
			defs = append(defs, &Node{Ident: n.Name, Node: n, Tree: t})
			return false
		}
		return true
	})
	return defs
}

func (t *Tree) Tests() []*Node {
	var defs []*Node
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.FuncDecl); ok && n.IsTest() {
			defs = append(defs, &Node{Ident: n.Name, Node: n, Tree: t})
			return false
		}
		return true
	})
	return defs
}

func (t *Tree) Imports() []*Node {
	var defs []*Node
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.ImportDecl); ok {
			defs = append(defs, &Node{Node: n, Tree: t})
			return false
		}
		return true
	})
	return defs
}

func (t *Tree) Ports() []*Node {
	var defs []*Node
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.PortTypeDecl); ok {
			defs = append(defs, &Node{Node: n, Tree: t})
			return false
		}
		return true
	})
	return defs
}

func (t *Tree) Components() []*Node {
	var defs []*Node
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.ComponentTypeDecl); ok {
			defs = append(defs, &Node{Node: n, Tree: t})
			return false
		}
		return true
	})
	return defs
}

func (t *Tree) Controls() []*Node {
	var defs []*Node
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.ControlPart); ok {
			defs = append(defs, &Node{Ident: n.Name, Node: n, Tree: t})
			return false
		}
		return true
	})
	return defs
}

func (t *Tree) ModulePars() []*Node {
	var defs []*Node
	ast.Inspect(t.Root, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.Module, *ast.ModuleDef, *ast.GroupDecl, *ast.ModuleParameterGroup:
			return true

		case *ast.ValueDecl:
			if n.Kind.Kind() != token.MODULEPAR && n.Kind.Kind() != token.ILLEGAL {
				return false
			}
			return true

		case *ast.Declarator:
			defs = append(defs, &Node{Node: n, Tree: t})
		}
		return false
	})
	return defs
}

// Pos encodes a line and column tuple into a offset-based Pos tag. If file nas
// not been parsed yet, Pos will return loc.NoPos.
func (t *Tree) Pos(line int, column int) loc.Pos {
	if t.FileSet == nil {
		return loc.NoPos
	}

	// We assume every FileSet has only one file, to make this work.
	return loc.Pos(int(t.FileSet.File(loc.Pos(1)).LineStart(line)) + column - 1)
}

// Position return a human readable position of the node.
func (t *Tree) Position(pos loc.Pos) loc.Position {
	if t.FileSet == nil {
		return loc.Position{}
	}
	return t.FileSet.Position(pos)
}

// ExprAt returns the primary expression at the given position.
func (t *Tree) ExprAt(pos loc.Pos) ast.Expr {
	s := t.SliceAt(pos)
	if len(s) == 0 {
		log.Debugf("%s: no expression at cursor position.\n", t.Position(pos))
		return nil
	}

	id, ok := s[0].(*ast.Ident)
	if !ok {
		log.Debugf("%s: no identifier at cursor position.\n", t.Position(pos))
		return nil
	}

	// Return the most left selector subtree (SelectorExpr is left-associative).
	if p, ok := t.ParentOf(id).(*ast.SelectorExpr); ok && id == p.Sel {
		return p
	}

	return id
}

// SliceAt returns the slice of nodes at the given position.
func (t *Tree) SliceAt(pos loc.Pos) []ast.Node {
	var (
		path  []ast.Node
		visit func(n ast.Node)
	)

	visit = func(n ast.Node) {
		if ast.IsNil(n) {
			return
		}

		if _, ok := n.(ast.Token); ok {
			return
		}

		path = append(path, n)
		if child := ast.FindChildOf(n, pos); !ast.IsNil(child) {
			visit(child)
		}

	}
	visit(t.Root)

	// Reverse path so leaf is first element.
	for i := 0; i < len(path)/2; i++ {
		path[i], path[len(path)-1-i] = path[len(path)-1-i], path[i]
	}

	return path
}

// Lookup returns the definitions of the given expression. For handling imports
// and multiple modules, use LookupWithDB.
func (t *Tree) Lookup(n ast.Expr) []*Node {
	return newFinder(&DB{}).lookup(n, t)
}

// LookupWithDB returns the definitions of the given expression, but uses the database for import resoltion.
func (t *Tree) LookupWithDB(n ast.Expr, db *DB) []*Node {
	return newFinder(db).lookup(n, t)

}

func newFinder(db *DB) *finder {
	return &finder{DB: db, cache: make(map[ast.Node][]*Node)}
}

type finder struct {
	*DB
	cache map[ast.Node][]*Node
}

func (f *finder) lookup(n ast.Expr, tree *Tree) []*Node {
	if n == nil {
		return nil
	}

	if results, ok := f.cache[n]; ok {
		return results
	}
	f.cache[n] = nil

	var results []*Node
	switch n := n.(type) {
	case *ast.Ident:
		results = f.ident(n, tree)

	case *ast.SelectorExpr:
		results = f.dot(n, tree)

	case *ast.IndexExpr:
		results = f.index(n, tree)

	case *ast.CallExpr:
		results = f.call(n, tree)

	default:
		log.Debugf("%s: unsupported node type: %T\n", tree.Position(n.Pos()), n)
	}

	f.cache[n] = results
	return results
}

func (f *finder) ident(id *ast.Ident, tree *Tree) []*Node {
	if p, ok := tree.ParentOf(id).(*ast.BinaryExpr); ok && id == p.X && p.Op.Kind() == token.ASSIGN {
		switch pp := tree.ParentOf(p).(type) {
		case *ast.CompositeLiteral:
			var results []*Node
			for _, c := range f.typeOf(&Node{Node: pp, Tree: tree}) {
				results = append(results, Definitions(id.String(), c.Node, c.Tree)...)
			}
			return results
		case *ast.ParenExpr:
			if ppp, ok := tree.ParentOf(pp).(*ast.CallExpr); ok {
				var results []*Node
				for _, c := range f.lookup(ppp.Fun, tree) {
					results = append(results, Definitions(id.String(), c.Node, c.Tree)...)
					// Bellow switch is required to handle behvaiour types references.
					// See tests TestLookup/parameters#01 and TestLookup/parameters#02.
					//
					// This brute force solution is not ideal, but it works for now.
					switch c.Node.(type) {
					case *ast.Field, *ast.RefSpec:
						for _, t := range f.typeOf(c) {
							results = append(results, Definitions(id.String(), t.Node, t.Tree)...)
						}
					}
				}
				return results
			}
		}

	}
	return f.globals(id, tree)
}

func (f *finder) globals(id *ast.Ident, tree *Tree) []*Node {
	var defs, q []*Node

	// Traverse parent scopes (P+) and collect imports scopes (I*)
	parents := ast.Parents(id, tree.Root)
	for _, n := range parents {
		switch n := n.(type) {
		case *ast.FuncDecl:
			if n.RunsOn != nil {
				q = append(q, &Node{Node: n.RunsOn.Comp, Tree: tree})
			}
			if n.System != nil {
				q = append(q, &Node{Node: n.System.Comp, Tree: tree})
			}
			if n.Mtc != nil {
				q = append(q, &Node{Node: n.Mtc.Comp, Tree: tree})
			}
		case *ast.BehaviourSpec:
			if n.RunsOn != nil {
				q = append(q, &Node{Node: n.RunsOn.Comp, Tree: tree})
			}
			if n.System != nil {
				q = append(q, &Node{Node: n.System.Comp, Tree: tree})
			}
		case *ast.BehaviourTypeDecl:
			if n.RunsOn != nil {
				q = append(q, &Node{Node: n.RunsOn.Comp, Tree: tree})
			}
			if n.System != nil {
				q = append(q, &Node{Node: n.System.Comp, Tree: tree})
			}
		}
		found := Definitions(id.String(), n, tree)
		defs = append(defs, found...)
	}

	// Traverse import scopes (I*)
	for len(q) > 0 {
		def := q[0]
		q = q[1:]

		n, ok := def.Node.(ast.Expr)
		if !ok {
			continue
		}

		for _, d := range f.lookup(n, def.Tree) {
			defs = append(defs, Definitions(id.String(), d.Node, d.Tree)...)
			if c, ok := d.Node.(*ast.ComponentTypeDecl); ok {
				for _, e := range c.Extends {
					q = append(q, &Node{Node: e, Tree: d.Tree})
				}
			}
		}
	}

	// Traver visible module scopes (I*)
	if mod := tree.ModuleOf(id); mod != nil {
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

func (f *finder) dot(n *ast.SelectorExpr, tree *Tree) []*Node {
	var result []*Node
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

func (f *finder) index(n *ast.IndexExpr, tree *Tree) []*Node {
	var result []*Node
	candidates := f.lookup(n.X, tree)
	for _, c := range candidates {
		for _, t := range f.typeOf(c) {
			if l, ok := t.Node.(*ast.ListSpec); ok {
				result = append(result, &Node{Node: l.ElemType, Tree: t.Tree})
			} else {
				result = append(result, c)
			}
		}
	}
	return result
}

func (f *finder) call(n *ast.CallExpr, tree *Tree) []*Node {
	var result []*Node
	candidates := f.lookup(n.Fun, tree)
	for _, c := range candidates {
		for _, t := range f.typeOf(c) {
			switch n := t.Node.(type) {
			case *ast.BehaviourTypeDecl:
				if n.Return != nil {
					result = append(result, &Node{Node: n.Return.Type, Tree: t.Tree})
				}
			case *ast.FuncDecl:
				if n.Return != nil {
					result = append(result, &Node{Node: n.Return.Type, Tree: t.Tree})
				}
			case *ast.SignatureDecl:
				if n.Return != nil {
					result = append(result, &Node{Node: n.Return.Type, Tree: t.Tree})
				}
			default:
				result = append(result, t)
			}
		}
	}
	return result
}

func (f *finder) typeOf(def *Node) []*Node {
	var result []*Node

	q := []*Node{def}

	for len(q) > 0 {

		def := q[0]
		q = q[1:]

		switch n := def.Node.(type) {
		case *ast.CompositeLiteral:
			q = append(q, &Node{Node: def.ParentOf(n), Tree: def.Tree})

		case *ast.BinaryExpr:
			q = append(q, &Node{Node: n.X, Tree: def.Tree})

		case *ast.TemplateDecl:
			q = append(q, &Node{Node: n.Type, Tree: def.Tree})

		case *ast.ValueDecl:
			q = append(q, &Node{Node: n.Type, Tree: def.Tree})

		case *ast.FormalPar:
			q = append(q, &Node{Node: n.Type, Tree: def.Tree})

		case *ast.Field:
			q = append(q, &Node{Node: n.Type, Tree: def.Tree})

		case *ast.SubTypeDecl:
			if n.Field != nil {
				q = append(q, &Node{Node: n.Field.Type, Tree: def.Tree})
			}

		case *ast.FuncDecl:
			if n.Return != nil {
				q = append(q, &Node{Node: n.Return.Type, Tree: def.Tree})
			}

		case *ast.SignatureDecl:
			if n.Return != nil {
				q = append(q, &Node{Node: n.Return.Type, Tree: def.Tree})
			}

		case ast.Expr:
			q = append(q, f.lookup(n, def.Tree)...)

		case *ast.RefSpec:
			q = append(q, f.lookup(n.X, def.Tree)...)

		case *ast.BehaviourSpec,
			*ast.BehaviourTypeDecl,
			*ast.ComponentTypeDecl,
			*ast.EnumSpec,
			*ast.EnumTypeDecl,
			*ast.Module,
			*ast.PortTypeDecl,
			*ast.ListSpec,
			*ast.StructSpec,
			*ast.StructTypeDecl:
			result = append(result, &Node{Node: n, Tree: def.Tree})
		}
	}
	return result
}

// Inspect the AST and return a list of all the definitions found by fn.
func Inspect(files []string, fn func(*Tree) []*Node) ([]*Node, error) {
	var (
		result []*Node
		wg     sync.WaitGroup
		mu     sync.Mutex
		err    *multierror.Error
	)

	wg.Add(len(files))
	for _, file := range files {
		go func(file string) {
			defer wg.Done()
			tree := ParseFile(file)
			if tree.Err != nil {
				err = multierror.Append(err, tree.Err)
			}
			defs := fn(tree)
			mu.Lock()
			defer mu.Unlock()
			result = append(result, defs...)
		}(file)
	}

	wg.Wait()
	return result, err
}

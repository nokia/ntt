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
				if _, ok := c.(ast.Token); !ok {
					t.parents[c] = n
				}
				visit(c)
			}
		}
		visit(t.Root)
	}

	if _, ok := n.(ast.Token); ok {
		return nil
	}
	return t.parents[n]
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

// Lookup returns the definitions of the given expression. For handling imports
// and multiple modules, use LookupWithDB.
func (tree *Tree) Lookup(n ast.Expr) []*Definition {
	f := &finder{DB: &DB{}, cache: make(map[ast.Node][]*Definition)}
	return f.lookup(n, tree)
}

// LookupWithDB returns the definitions of the given expression, but uses the database for import resoltion.
func (tree *Tree) LookupWithDB(n ast.Expr, db *DB) []*Definition {
	f := &finder{DB: db, cache: make(map[ast.Node][]*Definition)}
	return f.lookup(n, tree)

}

func (t *Tree) Modules() []*Definition {
	var defs []*Definition
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.Module); ok {
			defs = append(defs, &Definition{Ident: n.Name, Node: n, Tree: t})
			return false
		}
		return true
	})
	return defs
}

func (t *Tree) Funcs() []*Definition {
	var defs []*Definition
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.FuncDecl); ok {
			defs = append(defs, &Definition{Ident: n.Name, Node: n, Tree: t})
			return false
		}
		return true
	})
	return defs
}

func (t *Tree) Imports() []*Definition {
	var defs []*Definition
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.ImportDecl); ok {
			defs = append(defs, &Definition{Node: n, Tree: t})
			return false
		}
		return true
	})
	return defs
}

func (t *Tree) Ports() []*Definition {
	var defs []*Definition
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.PortTypeDecl); ok {
			defs = append(defs, &Definition{Node: n, Tree: t})
			return false
		}
		return true
	})
	return defs
}

func (t *Tree) Components() []*Definition {
	var defs []*Definition
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.ComponentTypeDecl); ok {
			defs = append(defs, &Definition{Node: n, Tree: t})
			return false
		}
		return true
	})
	return defs
}

func (t *Tree) Controls() []*Definition {
	var defs []*Definition
	ast.Inspect(t.Root, func(n ast.Node) bool {
		if n, ok := n.(*ast.ControlPart); ok {
			defs = append(defs, &Definition{Ident: n.Name, Node: n, Tree: t})
			return false
		}
		return true
	})
	return defs
}

func (t *Tree) ModulePars() []*Definition {
	var defs []*Definition
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
			defs = append(defs, &Definition{Node: n, Tree: t})
		}
		return false
	})
	return defs
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
	visit(tree.Root)

	// Reverse path so leaf is first element.
	for i := 0; i < len(path)/2; i++ {
		path[i], path[len(path)-1-i] = path[len(path)-1-i], path[i]
	}

	return path
}

type finder struct {
	*DB
	cache map[ast.Node][]*Definition
}

func (f *finder) lookup(n ast.Expr, tree *Tree) []*Definition {
	if n == nil {
		return nil
	}

	if results := f.cache[n]; results != nil {
		return results
	}

	var results []*Definition
	switch n := n.(type) {
	case *ast.SelectorExpr:
		results = f.dot(n, tree)

	case *ast.Ident:
		results = f.globals(n, tree)

	case *ast.IndexExpr:
		results = f.index(n, tree)

	case *ast.CallExpr:
		log.Debugf("%s: not implemented yet: %T\n", tree.Position(n.Pos()), n)
	default:
		log.Debugf("%s: Unsupported node type: %T\n", tree.Position(n.Pos()), n)
	}

	f.cache[n] = results
	return results
}

func (f *finder) globals(id *ast.Ident, tree *Tree) []*Definition {
	parents := ast.Parents(id, tree.Root)

	var defs, q []*Definition
	// Find definitions in current file by walking up the scopes.
	for _, n := range parents {
		switch n := n.(type) {
		case *ast.FuncDecl:
			if n.RunsOn != nil {
				q = append(q, &Definition{Node: n.RunsOn.Comp, Tree: tree})
			}
			if n.System != nil {
				q = append(q, &Definition{Node: n.System.Comp, Tree: tree})
			}
			if n.Mtc != nil {
				q = append(q, &Definition{Node: n.Mtc.Comp, Tree: tree})
			}
		case *ast.BehaviourSpec:
			if n.RunsOn != nil {
				q = append(q, &Definition{Node: n.RunsOn.Comp, Tree: tree})
			}
			if n.System != nil {
				q = append(q, &Definition{Node: n.System.Comp, Tree: tree})
			}
		case *ast.BehaviourTypeDecl:
			if n.RunsOn != nil {
				q = append(q, &Definition{Node: n.RunsOn.Comp, Tree: tree})
			}
			if n.System != nil {
				q = append(q, &Definition{Node: n.System.Comp, Tree: tree})
			}
		}
		found := Definitions(id.String(), n, tree)
		defs = append(defs, found...)
	}

	// Traverse alternate scope hierarchies (runs on, extends, etc.)
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
					q = append(q, &Definition{Node: e, Tree: d.Tree})
				}
			}
		}
	}

	// Find definitions in visible files.
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
			if l, ok := t.Node.(*ast.ListSpec); ok {
				result = append(result, &Definition{Node: l.ElemType, Tree: t.Tree})
			} else {
				result = append(result, c)
			}
		}
	}
	return result
}
func (f *finder) typeOf(def *Definition) []*Definition {
	var result []*Definition

	q := []*Definition{def}

	for len(q) > 0 {

		def := q[0]
		q = q[1:]

		switch n := def.Node.(type) {
		case *ast.TemplateDecl:
			q = append(q, &Definition{Node: n.Type, Tree: def.Tree})

		case *ast.ValueDecl:
			q = append(q, &Definition{Node: n.Type, Tree: def.Tree})

		case *ast.FormalPar:
			q = append(q, &Definition{Node: n.Type, Tree: def.Tree})

		case *ast.Field:
			q = append(q, &Definition{Node: n.Type, Tree: def.Tree})

		case *ast.SubTypeDecl:
			if n.Field != nil {
				q = append(q, &Definition{Node: n.Field.Type, Tree: def.Tree})
			}

		case *ast.FuncDecl:
			if n.Return != nil {
				q = append(q, &Definition{Node: n.Return.Type, Tree: def.Tree})
			}

		case *ast.SignatureDecl:
			if n.Return != nil {
				q = append(q, &Definition{Node: n.Return.Type, Tree: def.Tree})
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
			result = append(result, &Definition{Node: n, Tree: def.Tree})
		}
	}
	return result
}

// Inspect the AST and return a list of all the definitions found by fn.
func Inspect(files []string, fn func(*Tree) []*Definition) ([]*Definition, error) {
	var (
		result []*Definition
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

package ttcn3

import (
	"fmt"
	"sync"
	"time"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

// DB implements database for querying TTCN-3 source code bases.
type DB struct {
	// Names is a map from symbol name to a list of files that contain the symbol.
	Names map[string]map[string]bool

	// Modules maps from module name to file path.
	Modules map[string]map[string]bool

	mu sync.Mutex
}

// Index parses TTCN-3 source files and adds names and dependencies to the database.
func (db *DB) Index(files ...string) {
	if db.Names == nil {
		db.Names = make(map[string]map[string]bool)
	}
	if db.Modules == nil {
		db.Modules = make(map[string]map[string]bool)
	}

	var (
		syms  int
		wg    sync.WaitGroup
		start = time.Now()
	)
	wg.Add(len(files))
	for _, path := range files {
		go func(path string) {
			defer wg.Done()
			tree := ParseFile(path)
			db.mu.Lock()
			for _, n := range tree.Modules() {
				syms++
				db.addModule(path, ast.Name(n))
			}

			for k := range tree.Names {
				syms++
				db.addDefinition(path, k)
			}
			db.mu.Unlock()

		}(path)
	}
	wg.Wait()
	log.Debugf("Cache built in %v: %d symbols in %d files.\n", time.Since(start), syms, len(files))
}

func (db *DB) LookupAt(file string, line int, col int) []*Definition {
	start := time.Now()
	log.Debugf("%s:%d:%d: Lookup started...\n", file, line, col)
	defer log.Debugf("%s:%d:%d: Lookup took %s\n", file, line, col, time.Since(start))

	tree := ParseFile(file)
	n := tree.ExprAt(tree.Pos(line, col))
	if n == nil {
		log.Debugf("%s:%d:%d: No symbol at cursor position.", file, line, col)
		return nil
	}
	parents := ast.Parents(n, tree.Root)
	return db.FindDefinitions(n.(ast.Expr), tree, parents...)

}

func (db *DB) FindDefinitions(n ast.Expr, tree *Tree, parents ...ast.Node) []*Definition {
	return db.findDefinitions(make(map[ast.Node]bool), n, tree, parents...)
}

func (db *DB) findDefinitions(visited map[ast.Node]bool, n ast.Expr, tree *Tree, parents ...ast.Node) []*Definition {
	if len(parents) > 1 {
		if _, ok := parents[len(parents)-1].(ast.NodeList); ok {
			parents = parents[:len(parents)-1]
		}
	}
	switch n := n.(type) {
	case *ast.SelectorExpr:
		return db.findTypes(visited, n, tree, parents...)

	case *ast.Ident:
		var defs []*Definition
		id := n

		// Find definitions in current file by walking up the scopes.
		for _, n := range parents {
			visited[n] = true
			found := NewScope(n, tree).Lookup(id.String())
			defs = append(defs, found...)
		}

		if mod, ok := parents[len(parents)-1].(*ast.Module); ok {
			// TTCN-3 standard requires, that all global definition may have a module prefix.
			if id.String() == ast.Name(mod) {
				defs = append(defs, &Definition{
					Ident: mod.Name,
					Node:  mod,
					Tree:  tree,
				})
			}
			// Find defintions of files of the same module
			for file := range db.Modules[ast.Name(mod)] {
				tree := ParseFile(file)
				for _, m := range tree.Modules() {
					if !visited[m] && ast.Name(m) == ast.Name(mod) {
						found := NewScope(m, tree).Lookup(id.String())
						for _, d := range found {
							if _, ok := d.Node.(*ast.ImportDecl); !ok {
								defs = append(defs, d)
							}
						}
					}
				}
			}

			// Find definitions in imported files.
			for _, m := range db.FindImportedDefinitions(ast.Name(id), mod) {
				defs = append(defs, db.findDefinitions(visited, id, m.Tree, m.Node)...)
			}
		}

		return defs

	default:
		log.Debugf("%s: Unsupported node type: %T\n", tree.Position(n.Pos()), n)
		return nil
	}

}

// findType returns all type definitions refered by expression n.
func (db *DB) findTypes(visited map[ast.Node]bool, n ast.Expr, tree *Tree, parents ...ast.Node) []*Definition {

	var result []*Definition
	switch n := n.(type) {
	case *ast.SelectorExpr:
		candidates := db.findTypes(visited, n.X, tree, parents...)
		for _, c := range candidates {
			for _, t := range db.typeOf(visited, c, parents...) {
				if defs := NewScope(t.Node, t.Tree).Lookup(ast.Name(n.Sel)); len(defs) > 0 {
					result = append(result, defs...)
				}
			}
		}
		return result

	case *ast.IndexExpr:
		return db.findTypes(visited, n.X, tree, append(parents, n)...)

	case *ast.CallExpr:
		return db.findTypes(visited, n.Fun, tree, append(parents, n)...)

	case *ast.Ident:
		result = db.findDefinitions(visited, n, tree, parents...)
		return result
	}

	log.Debugf("%s: Unsupported node type: %T\n", tree.Position(n.Pos()), n)
	return nil
}

func (db *DB) typeOf(visited map[ast.Node]bool, def *Definition, parents ...ast.Node) []*Definition {
	if t := def.Type(); t != nil {
		def = t
	}

	if x, ok := def.Node.(ast.Expr); ok {
		return db.findTypes(visited, x, def.Tree, parents...)
	}

	return []*Definition{def}
}

// FindImportedDefinitions returns a list of modules that may contain the given
// symbol. First parameter id specifies the symbol to look for and second
// parameter module specifies where the imports come from.
func (db *DB) FindImportedDefinitions(id string, module *ast.Module) []*Definition {
	importedModules := make(map[string]bool)
	importedFiles := make(map[string]bool)

	// Only use imports from the current module.
	ast.WalkModuleDefs(func(n *ast.ModuleDef) bool {
		if n, ok := n.Def.(*ast.ImportDecl); ok {
			imported := ast.Name(n.Module)

			// Ignore self-imports
			if imported == ast.Name(module) {
				return false
			}
			importedModules[imported] = true
			for file := range db.Modules[imported] {
				importedFiles[file] = true
			}
		}
		return true
	}, module)

	// Find all files that contain the symbol.
	var candidates []string
	for file := range db.Names[id] {
		if importedFiles[file] {

			candidates = append(candidates, file)
		}
	}

	// Parse all imported modules that contain the symbol.
	var mods []*Definition
	for _, file := range candidates {
		tree := ParseFile(file)
		for _, mod := range tree.Modules() {
			if importedModules[ast.Name(mod)] && mod != module {
				mods = append(mods, &Definition{Ident: mod.Name, Node: mod, Tree: tree})
			}
		}
	}
	return mods
}

func (db *DB) addModule(file string, name string) {
	if db.Modules[name] == nil {
		db.Modules[name] = make(map[string]bool)
	}
	db.Modules[name][file] = true
}

func (db *DB) addDefinition(file string, name string) {
	if db.Names[name] == nil {
		db.Names[name] = make(map[string]bool)
	}
	db.Names[name][file] = true
}

// parentNodes returns the symbol name and the stack of nodes that lead to the symbol.
func parentNodes(tree *Tree, pos loc.Pos) (ast.Expr, []ast.Node) {
	// First two elements on the parents are expected to be the ID-Token and the
	// Identifier node.
	if s := tree.SliceAt(pos); len(s) >= 2 {
		if tok, ok := s[0].(ast.Token); ok && tok.Kind == token.IDENT {
			return s[1].(ast.Expr), s[2:]
		}
	}

	return nil, nil
}

// A selector expression requires a different context to resolve a symbol.
func isSelectorID(id ast.Node, parents []ast.Node) bool {
	if len(parents) > 0 {
		if x, ok := parents[0].(*ast.SelectorExpr); ok {
			return id == x.Sel
		}
	}
	return false
}

// isFieldID returns true if the ID is the left hand side of a field assignment.
func isFieldID(n ast.Node, parents []ast.Node) bool {
	if len(parents) < 2 {
		return false
	}

	switch parents[1].(type) {
	case *ast.CompositeLiteral:
		return false
	}
	if x, ok := parents[0].(*ast.BinaryExpr); ok {
		return n == x.X
	}
	return false
}

func nodes(nodes []ast.Node) []string {
	var types []string
	for _, n := range nodes {
		types = append(types, fmt.Sprintf("%T", n))
	}
	return types
}

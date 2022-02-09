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
	n, stack := parentNodes(tree, tree.Pos(line, col))
	if n == nil {
		log.Debugf("%s:%d:%d: No symbol at cursor position. Stack: %v\n", file, line, col, nodes(stack))
	}

	if isFieldID(n, stack) {
		log.Debugf("Field assignment not implemented. Stack: %v", nodes(stack))
		return nil
	}

	if isSelectorID(n, stack) {
		n, stack = stack[0].(*ast.SelectorExpr), stack[1:]
	}
	return db.findDefinitions(n.(ast.Expr), tree, stack...)

}

func (db *DB) findDefinitions(n ast.Expr, tree *Tree, stack ...ast.Node) []*Definition {
	switch n := n.(type) {
	case *ast.SelectorExpr:
		return db.findTypes(n, tree, stack...)

	case *ast.Ident:
		var defs []*Definition
		id := n

		// Find definitions in current file by walking up the scopes.
		for _, n := range stack {
			defs = append(defs, NewScope(n, tree).Lookup(id.String())...)
		}

		// Find definitions in imported files.
		if mod, ok := stack[len(stack)-1].(*ast.Module); ok {
			for _, m := range db.FindImportedDefinitions(ast.Name(id), ast.Name(mod)) {
				defs = append(defs, db.findDefinitions(id, m.Tree, m.Node)...)
			}
		}

		return defs
	default:
		log.Debugf("%s: Unsupported node type: %T\n", tree.Position(n.Pos()), n)
		return nil
	}

}

// findType returns all type definitions refered by expression n.
func (db *DB) findTypes(n ast.Expr, tree *Tree, stack ...ast.Node) []*Definition {
	switch n := n.(type) {
	case *ast.SelectorExpr:

		var result []*Definition
		n, stack := stack[0].(*ast.SelectorExpr), stack[1:]
		for _, t := range db.findTypes(n.X, tree, stack...) {
			// Convert every found type definition to a scope and
			// try looking up the field. Non-Scope types are ignored.
			defs := NewScope(t.Node, t.Tree).Lookup(ast.Name(n.Sel))
			result = append(result, defs...)

		}
		return result
		//for _, t := range db.findTypes(n.X, tree, stack...) {
		//	defs := NewScope(t.Node, t.Tree).Lookup(ast.Name(n.Sel))
		//	result = append(result, defs...)

		//}
	case *ast.IndexExpr:
		return db.findTypes(n.X, tree, append(stack, n)...)

	case *ast.CallExpr:
		return db.findTypes(n.Fun, tree, append(stack, n)...)

	case *ast.Ident:
		for _, def := range db.findDefinitions(n, tree, stack...) {
			if t := def.Type(); t != nil {
				// Type references need further resolving.
				if _, ok := t.Node.(ast.Expr); ok {
					//types = append(types, findTypes()...)
				} else {
					//types = append(types, t)
				}
			}
		}
		return nil
	}

	log.Debugf("%s: Unsupported node type: %T\n", tree.Position(n.Pos()), n)
	return nil
}

// FindImportedDefinitions returns a list of modules that may contain the given
// symbol. First parameter id specifies the symbol to look for and second
// parameter module specifies where the imports come from.
func (db *DB) FindImportedDefinitions(id, module string) []*Definition {
	importedModules := make(map[string]bool)
	importedFiles := make(map[string]bool)

	for file := range db.Modules[module] {
		tree := ParseFile(file)
		for _, mod := range tree.Modules() {
			if ast.Name(mod) != module {
				continue
			}
			ast.WalkModuleDefs(func(n *ast.ModuleDef) bool {
				if n, ok := n.Def.(*ast.ImportDecl); ok {
					name := ast.Name(n.Module)
					importedModules[name] = true
					for file := range db.Modules[name] {
						importedFiles[file] = true
					}
				}
				return true
			}, mod)
		}
	}

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
			if importedModules[ast.Name(mod)] {
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
	// First two elements on the stack are expected to be the ID-Token and the
	// Identifier node.
	if s := tree.SliceAt(pos); len(s) >= 2 {
		if tok, ok := s[0].(ast.Token); ok && tok.Kind == token.IDENT {
			return s[1].(ast.Expr), s[2:]
		}
	}

	return nil, nil
}

// A selector expression requires a different context to resolve a symbol.
func isSelectorID(id ast.Node, stack []ast.Node) bool {
	if len(stack) > 0 {
		if x, ok := stack[0].(*ast.SelectorExpr); ok {
			return id == x.Sel
		}
	}
	return false
}

// isFieldID returns true if the ID is the left hand side of a field assignment.
func isFieldID(n ast.Node, stack []ast.Node) bool {
	if len(stack) < 2 {
		return false
	}

	switch stack[1].(type) {
	case *ast.CompositeLiteral:
		return false
	}
	if x, ok := stack[0].(*ast.BinaryExpr); ok {
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

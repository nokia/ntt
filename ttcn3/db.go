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

func (db *DB) ResolveAt(file string, line int, col int) []*Definition {
	start := time.Now()
	log.Debugf("%s:%d:%d: Resolving...\n", file, line, col)
	defer log.Debugf("%s:%d:%d: Resolve took %s\n", file, line, col, time.Since(start))

	tree := ParseFile(file)
	id, stack := getResolveStack(tree, tree.Pos(line, col))
	if id == nil {
		log.Debugf("%s:%d:%d: No symbol to resolve. Stack: %v\n", file, line, col, nodes(stack))
		return nil
	}

	if isSelectorID(id, stack) || isFieldID(id, stack) {
		log.Debugf("Resolve %v not implemented.", nodes(stack))
		return nil
	}

	var ret []*Definition
	if defs := db.findLocals(ast.Name(id), tree, stack...); len(defs) > 0 {
		ret = append(ret, defs...)
	}

	if mod, ok := stack[len(stack)-1].(*ast.Module); ok {
		ret = append(ret, db.findGlobals(ast.Name(id), mod)...)
	}

	return ret
}

func (db *DB) findLocals(name string, tree *Tree, stack ...ast.Node) []*Definition {
	var defs []*Definition
	for _, n := range stack {
		if scope := NewScope(n, tree); scope != nil {
			if def, ok := scope.Names[name]; ok {
				for {
					defs = append(defs, def)
					if def.Next == nil {
						break
					}
					def = def.Next
				}
			}
		}
	}
	return defs
}

func (db *DB) findGlobals(name string, mod *ast.Module) []*Definition {
	modules, files := db.importMaps(mod)
	candidates := db.candidates(name, files)

	var result []*Definition
	for _, file := range candidates {
		tree := ParseFile(file)
		for _, mod := range tree.Modules() {
			if modules[ast.Name(mod)] {
				if defs := db.findLocals(name, tree, mod); len(defs) > 0 {
					result = append(result, defs...)
				}
			}
		}
	}
	return result
}

func (db *DB) importMaps(n ast.Node) (mods, files map[string]bool) {
	mods = make(map[string]bool)
	files = make(map[string]bool)
	ast.WalkModuleDefs(func(n *ast.ModuleDef) bool {
		if n, ok := n.Def.(*ast.ImportDecl); ok {
			name := ast.Name(n.Module)
			mods[name] = true
			for file := range db.Modules[name] {
				files[file] = true
			}
		}
		return true
	}, n)
	return mods, files
}

func (db *DB) candidates(name string, imported map[string]bool) []string {
	db.mu.Lock()
	defer db.mu.Unlock()

	var candidates []string
	for file := range db.Names[name] {
		if imported[file] {
			candidates = append(candidates, file)
		}
	}
	return candidates
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

// getResolveStack returns the symbol name and the stack of nodes that lead to the symbol.
func getResolveStack(tree *Tree, pos loc.Pos) (ast.Node, []ast.Node) {
	// First two elements on the stack are expected to be the ID-Token and the
	// Identifier node.
	if s := tree.SliceAt(pos); len(s) >= 2 {
		if tok, ok := s[0].(ast.Token); ok && tok.Kind == token.IDENT {
			return s[1], s[2:]
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

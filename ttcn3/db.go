package ttcn3

import (
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

func (db *DB) ResolveAt(file string, line int, col int) []*Definition {
	start := time.Now()
	log.Debugf("%s:%d:%d: Resolving...\n", file, line, col)
	defer log.Debugf("%s:%d:%d: Resolve took %s\n", file, line, col, time.Since(start))

	tree := ParseFile(file)
	name, stack := getResolveStack(tree, tree.Pos(line, col))
	if name == "" {
		log.Printf("%s:%d:%d: No symbol to resolve.\n", file, line, col)
		return nil
	}

	var defs []*Definition
	for len(stack) > 0 {
		n := stack[0]
		stack = stack[1:]

		if scope := NewScope(n); scope != nil {
			if def, ok := scope.Names[name]; ok {
				defs = append(defs, def)
			}
		}

	}
	return defs
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
				db.addModule(path, ast.Name(n))
			}

			for k := range tree.Names {
				db.addDefinition(path, k)
			}
			db.mu.Unlock()

		}(path)
	}
	wg.Wait()
	log.Debugf("Cache built in %v: %d symbols in %d files.\n", time.Since(start), len(db.Names), len(files))
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
func getResolveStack(tree *Tree, pos loc.Pos) (string, []ast.Node) {
	// First two elements on the stack are expected to be the ID-Token and the
	// Identifier node.
	if s := tree.SliceAt(pos); len(s) >= 2 {
		if tok, ok := s[0].(*ast.Token); ok && tok.Kind == token.IDENT {
			return tok.Lit, s[2:]
		}
	}

	return "", nil
}

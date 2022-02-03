package ttcn3

import (
	"sync"
	"time"

	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/ttcn3/ast"
)

// DB implements database for querying TTCN-3 source code bases.
type DB struct {
	// Names is a map from symbol name to a list of files that contain the symbol.
	Names map[string]map[string]bool

	// Dependencies is a map from module name to a list of modules that the module depends on.
	Dependencies map[string]map[string]bool
}

// Index parses TTCN-3 source files and adds names and dependencies to the database.
func (db *DB) Index(files ...string) {
	if db.Names == nil {
		db.Names = make(map[string]map[string]bool)
	}
	if db.Dependencies == nil {
		db.Dependencies = make(map[string]map[string]bool)
	}

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	start := time.Now()
	wg.Add(len(files))
	for _, path := range files {
		go func(path string) {
			defer wg.Done()

			tree := ParseFile(path)
			for _, m := range tree.Modules() {
				ast.WalkModuleDefs(func(n *ast.ModuleDef) bool {
					mu.Lock()
					switch n := n.Def.(type) {
					case *ast.ValueDecl:
						for _, d := range n.Decls {
							db.addDefinition(path, ast.Name(d))
						}
					case *ast.ImportDecl:
						db.addDependency(ast.Name(m), ast.Name(n.Module))
					default:
						db.addDefinition(path, ast.Name(n))
					}
					mu.Unlock()
					return true
				}, m)
			}
		}(path)
	}
	wg.Wait()
	log.Debugf("Cache built in %v: %d symbols in %d files.\n", time.Since(start), len(db.Names), len(files))
}

func (db *DB) addDependency(from, to string) {
	if db.Dependencies[from] == nil {
		db.Dependencies[from] = make(map[string]bool)
	}
	db.Dependencies[from][to] = true
}

func (db *DB) addDefinition(file string, name string) {
	if db.Names[name] == nil {
		db.Names[name] = make(map[string]bool)
	}
	db.Names[name][file] = true
}

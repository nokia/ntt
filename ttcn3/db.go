package ttcn3

import (
	"sync"
	"time"

	"github.com/nokia/ntt/internal/log"
	"github.com/nokia/ntt/ttcn3/ast"
)

// DB implements database for querying TTCN-3 source code bases.
type DB struct {
	Names   map[string][]string
	namesMu sync.Mutex

	Dependencies map[string][]string
	depsMu       sync.Mutex
}

// Index fi
func (db *DB) Index(files ...string) {
	if db.Names == nil {
		db.Names = make(map[string][]string)
	}
	if db.Dependencies == nil {
		db.Dependencies = make(map[string][]string)
	}

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(len(files))
	for _, path := range files {
		go func(path string) {
			defer wg.Done()
			db.cacheFile(path)
		}(path)
	}
	wg.Wait()
	log.Debugf("Cache built in %v: %d symbols in %d files.\n", time.Since(start), len(db.Names), len(files))
}

func (db *DB) cacheFile(path string) {
	tree := ParseFile(path)
	for _, m := range tree.Modules() {
		ast.WalkModuleDefs(func(n *ast.ModuleDef) bool {
			switch n := n.Def.(type) {
			case *ast.ValueDecl:
				for _, d := range n.Decls {
					db.addDefinition(path, tree, m, d)
				}
			case *ast.ImportDecl:
				db.addDependency(ast.Name(m), ast.Name(n.Module))
			default:
				db.addDefinition(path, tree, m, n)
			}
			return true
		}, m)
	}
}

func (db *DB) addDependency(from, to string) {
	db.depsMu.Lock()
	defer db.depsMu.Unlock()
	db.Dependencies[from] = append(db.Dependencies[from], to)
}

func (db *DB) addDefinition(file string, tree *Tree, m *ast.Module, n ast.Node) {
	name := ast.Name(n)
	db.namesMu.Lock()
	defer db.namesMu.Unlock()
	db.Names[name] = append(db.Names[name], file)
}

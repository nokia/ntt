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

// VisibleModules returns a list of modules that may contain the given
// symbol. First parameter id specifies the symbol to look for and second
// parameter module specifies where the imports come from.
func (db *DB) VisibleModules(id string, mod *ast.Module) []*Definition {
	importedModules := make(map[string]bool)
	importedFiles := make(map[string]bool)

	addImport := func(moduleName string) {
		importedModules[moduleName] = true
		for file := range db.Modules[moduleName] {
			importedFiles[file] = true
		}
	}

	// TTCN-3 standard requires, that all global definition may have a
	// module prefix. We handle this by "self-importing" the current
	// module.
	addImport(ast.Name(mod))

	// Only use imports from the current module.
	ast.WalkModuleDefs(func(n *ast.ModuleDef) bool {
		if n, ok := n.Def.(*ast.ImportDecl); ok {
			addImport(ast.Name(n.Module))
		}
		return true
	}, mod)

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
		for _, m := range tree.Modules() {
			if importedModules[ast.Name(m)] {
				mods = append(mods, &Definition{Ident: m.Name, Node: m, Tree: tree})
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

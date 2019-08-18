package loader

import (
	"fmt"
	gort "runtime"
	"sync"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/runtime"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/doc"
	"github.com/nokia/ntt/internal/ttcn3/parser"
)

type ttcn3Suite struct {
	fset       *loc.FileSet
	mods       []*runtime.Module
	tests      []*runtime.Test
	syntax     []*ast.Module
	mode       parser.Mode
	numWorkers int

	Config
}

type definition struct {
	obj    interface{}
	syntax ast.Node
}

func NewSuite(conf Config) *ttcn3Suite {
	suite := &ttcn3Suite{
		Config:     conf,
		fset:       loc.NewFileSet(),
		numWorkers: gort.NumCPU(),
		mods:       make([]*runtime.Module, 0, len(conf.Sources)),
		tests:      make([]*runtime.Test, 0, len(conf.Sources)),
	}
	return suite
}

func (suite *ttcn3Suite) load() (*ttcn3Suite, error) {
	var wg sync.WaitGroup

	paths := make(chan string)
	go func() {
		defer close(paths)
		for _, src := range suite.Sources {
			paths <- src
		}
	}()

	defs := make(chan definition)
	wg.Add(suite.numWorkers)
	for i := 0; i < suite.numWorkers; i++ {
		go func() {
			for path := range paths {
				mods, err := parser.ParseModules(suite.fset, path, nil, suite.mode)

				// TODO(5nord) implement proper error propagation
				if err != nil {
					fmt.Println(err.Error())
					return
				}

				for _, modSyn := range mods {
					mod := runtime.NewModule(
						modSyn.Name.String(),
						path,
						doc.FindAllTags(modSyn.Tok.Comments()),
					)
					defs <- definition{mod, modSyn}

					ast.WalkModuleDefs(func(n ast.Node) bool {
						switch d := n.(type) {
						case *ast.FuncDecl:
							if !d.IsTest() {
								break
							}
							test := runtime.NewTest(
								mod,
								d.Name.String(),
								doc.FindAllTags(d.Kind.Comments()),
							)
							defs <- definition{test, d}

						case *ast.ImportDecl:
							imp := runtime.NewImport(
								mod,
								d.Module.String(),
								doc.FindAllTags(d.ImportTok.Comments()),
							)
							mod.AddImport(imp)
						}
						return true
					}, modSyn)
				}
			}
			wg.Done()
		}()
	}

	// End of Pipeline
	go func() {
		wg.Wait()
		close(defs)
	}()

	// Pipeline sink
	for d := range defs {
		switch x := d.obj.(type) {
		case *runtime.Module:
			suite.mods = append(suite.mods, x)
			suite.syntax = append(suite.syntax, d.syntax.(*ast.Module))
		case *runtime.Test:
			suite.tests = append(suite.tests, x)
		}
	}

	return suite, nil
}

func (suite *ttcn3Suite) Modules() []*runtime.Module { return suite.mods }
func (suite *ttcn3Suite) Tests() []*runtime.Test     { return suite.tests }
func (suite *ttcn3Suite) Syntax() []*ast.Module      { return suite.syntax }

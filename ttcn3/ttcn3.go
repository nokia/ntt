// Package ttcn3 provides routines for evaluating TTCN-3 source code.
//
// This package is in alpha stage, as we are still figuring out requirements and interfaces.
package ttcn3

import (
	"context"
	"runtime"

	"github.com/nokia/ntt/internal/fs"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/memoize"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/parser"
)

var (
	// cache stores various (expensive) calculation
	cache = memoize.Store{}

	// Limits the number of parallel parser calls per process.
	parseLimit = make(chan struct{}, runtime.NumCPU())
)

// Tree represents the TTCN-3 syntax tree, usually of a file.
type Tree struct {
	fset  *loc.FileSet
	nodes []ast.Node
	err   error
}

// ParseFile parses a file and returns a syntax tree.
func ParseFile(path string) *Tree {
	f := fs.Open(path)
	f.Handle = cache.Bind(f.ID(), func(ctx context.Context) interface{} {
		b, err := f.Bytes()
		if err != nil {
			return &Tree{err: err}
		}

		parseLimit <- struct{}{}
		defer func() { <-parseLimit }()

		fset := loc.NewFileSet()
		mods, err := parser.ParseModules(fset, path, b, parser.AllErrors)

		tree := &Tree{fset: fset, err: err}
		for _, n := range mods {
			tree.nodes = append(tree.nodes, n)
		}
		return tree
	})

	return f.Handle.Get(context.TODO()).(*Tree)
}

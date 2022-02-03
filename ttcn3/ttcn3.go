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
	"github.com/nokia/ntt/ttcn3/parser"
)

var (
	// cache stores various (expensive) calculation
	cache = memoize.Store{}

	// Limits the number of parallel parser calls per process.
	parseLimit = make(chan struct{}, runtime.NumCPU())
)

// Parse parses a string and returns a syntax tree.
func Parse(src string) *Tree {
	parseLimit <- struct{}{}
	defer func() { <-parseLimit }()

	fset := loc.NewFileSet()
	root, err := parser.Parse(fset, "", src)
	return &Tree{FileSet: fset, Root: root, Err: err}
}

// ParseFile parses a file and returns a syntax tree.
func ParseFile(path string) *Tree {
	f := fs.Open(path)
	f.Handle = cache.Bind(f.ID(), func(ctx context.Context) interface{} {
		b, err := f.Bytes()
		if err != nil {
			return &Tree{Err: err}
		}

		parseLimit <- struct{}{}
		defer func() { <-parseLimit }()

		fset := loc.NewFileSet()
		root, err := parser.Parse(fset, path, b)
		return &Tree{FileSet: fset, Root: root, Err: err}
	})

	return f.Handle.Get(context.TODO()).(*Tree)
}

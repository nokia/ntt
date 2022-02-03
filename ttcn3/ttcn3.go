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
	return parse("", []byte(src))
}

// ParseFile parses a file and returns a syntax tree.
func ParseFile(path string) *Tree {
	f := fs.Open(path)
	f.Handle = cache.Bind(f.ID(), func(ctx context.Context) interface{} {
		return parse(path, nil)
	})

	return f.Handle.Get(context.TODO()).(*Tree)
}

func parse(path string, input []byte) *Tree {
	// Without parseLimit we may end up with too many open files.
	parseLimit <- struct{}{}
	defer func() { <-parseLimit }()

	if input == nil {
		b, err := fs.Content(path)
		if err != nil {
			return &Tree{Err: err}
		}
		input = b
	}

	fset := loc.NewFileSet()
	root, err := parser.Parse(fset, "", input)
	return &Tree{FileSet: fset, Root: root, Err: err}
}

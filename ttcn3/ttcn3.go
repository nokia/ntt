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

// Pos encodes a line and column tuple into a offset-based Pos tag. If file nas
// not been parsed yet, Pos will return loc.NoPos.
func (tree *Tree) Pos(line int, column int) loc.Pos {
	if tree.fset == nil {
		return loc.NoPos
	}

	// We asume every FileSet has only one file, to make this work.
	return loc.Pos(int(tree.fset.File(loc.Pos(1)).LineStart(line)) + column - 1)
}

func (tree *Tree) SliceAt(line, col int) *Tree {
	pos := tree.Pos(line, col)
	for i := range tree.nodes {
		if path := slice(tree.nodes[i], pos); path != nil {
			// A manual deep copy is fragile. We keep it here, until he have something better.
			// TODO(5nord): Find something better.
			return &Tree{
				fset:  tree.fset,
				nodes: path,
				err:   tree.err,
			}
		}
	}
	return nil
}

func slice(n ast.Node, pos loc.Pos) []ast.Node {
	var (
		path  []ast.Node
		found bool
	)

	ast.Inspect(n, func(n ast.Node) bool {
		if found {
			return false
		}

		if n == nil {
			path = path[:len(path)-1]
			return false
		}

		path = append(path, n)

		// We don't need to descend any deeper if we're not near desired
		// position.
		if n.End() < pos || pos < n.Pos() {
			return false
		}

		if n.Pos() <= pos && pos <= n.End() {
			found = true
		}

		return !found
	})

	if len(path) == 0 {
		return nil
	}

	// Reverse path so leaf is first element.
	for i := 0; i < len(path)/2; i++ {
		path[i], path[len(path)-1-i] = path[len(path)-1-i], path[i]
	}
	return path
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

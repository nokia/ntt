package ntt

import (
	"context"
	"fmt"
	"runtime"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/parser"
)

// Limits the number of parallel parser calls per process.
var parseLimit = make(chan struct{}, runtime.NumCPU())

func (suite *Suite) Parse(f *File) (*ast.Module, *loc.FileSet, error) {
	type parseData struct {
		mod  *ast.Module
		fset *loc.FileSet
		err  error
	}

	f.handle = suite.store.Bind(f.ID(), func(ctx context.Context) interface{} {
		data := parseData{}

		b, err := f.Bytes()
		if err != nil {
			data.err = err
			return &data
		}

		parseLimit <- struct{}{}
		defer func() { <-parseLimit }()

		var mods []*ast.Module
		data.fset = loc.NewFileSet()
		mods, data.err = parser.ParseModules(data.fset, f.Path(), b, 0)

		// It's easier to support only one module per file.
		if len(mods) == 1 {
			data.mod = mods[0]
		} else if len(mods) > 1 {
			data.err = fmt.Errorf("file %q contains more than one module.", f.Path())
		}
		return &data
	})

	v := f.handle.Get(context.TODO())
	data := v.(*parseData)
	return data.mod, data.fset, data.err
}

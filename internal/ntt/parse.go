package ntt

import (
	"context"
	"runtime"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/parser"
)

// Limits the number of parallel parser calls per process.
var parseLimit = make(chan struct{}, runtime.NumCPU())

func (suite *Suite) Parse(f *File) ([]*ast.Module, *loc.FileSet, error) {
	type parseData struct {
		mods []*ast.Module
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

		data.fset = loc.NewFileSet()
		data.mods, data.err = parser.ParseModules(data.fset, f.Path(), b, 0)
		return &data
	})

	v := f.handle.Get(context.TODO())
	data := v.(*parseData)
	return data.mods, data.fset, data.err
}

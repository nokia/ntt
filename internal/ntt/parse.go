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

type ParseInfo struct {
	Module  *ast.Module
	FileSet *loc.FileSet
	Err     error
}

func (suite *Suite) Parse(f *File) *ParseInfo {
	f.handle = suite.store.Bind(f.ID(), func(ctx context.Context) interface{} {
		data := ParseInfo{}

		b, err := f.Bytes()
		if err != nil {
			data.Err = err
			return &data
		}

		parseLimit <- struct{}{}
		defer func() { <-parseLimit }()

		var mods []*ast.Module
		data.FileSet = loc.NewFileSet()
		mods, data.Err = parser.ParseModules(data.FileSet, f.Path(), b, 0)

		// It's easier to support only one module per file.
		if len(mods) == 1 {
			data.Module = mods[0]
		} else if len(mods) > 1 {
			data.Err = fmt.Errorf("file %q contains more than one module.", f.Path())
		}
		return &data
	})

	v := f.handle.Get(context.TODO())
	return v.(*ParseInfo)
}

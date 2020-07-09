package ntt

import (
	"context"

	"github.com/nokia/ntt/internal/ttcn3/types"
)

// Symbols
func (suite *Suite) symbols(syntax *ParseInfo) *types.Info {
	syntax.handle = suite.store.Bind(syntax.ID(), func(ctx context.Context) interface{} {

		info := &types.Info{
			Fset:  syntax.FileSet,
			Error: func(err error) {},
		}

		info.Define(syntax.Module)
		info.Resolve()

		return info
	})

	v := syntax.handle.Get(context.TODO())
	return v.(*types.Info)
}

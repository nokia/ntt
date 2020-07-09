package types

import (
	"fmt"

	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/ast"
)

// Info holds various information about TTCN-3 symbols and types.
type Info struct {

	// Scopes associates identifiers with their enclosing scope.
	Scopes map[*ast.Ident]Scope

	// Types associcates expressions with their evaluated type.
	Types map[ast.Expr]Type

	//
	Fset *loc.FileSet

	//
	Modules map[string]*Module

	//
	Error func(error)

	Import func(name string) error

	currScope Scope
	currMod   *Module
}

// error records errors during definition phase, such like ErrRedefined, ...
func (info *Info) error(err error) {
	if info.Error == nil {
		panic(err)
	}
	info.Error(err)
}

func (info *Info) unknownIdentifierError(n ast.Node) {
	info.error(&UnknownIdentifierError{
		Pos:  info.Fset.Position(n.Pos()),
		Name: identName(n),
	})
}

func (info *Info) noFieldError(typ Type, field ast.Expr, pos loc.Pos) {
	info.error(&NoFieldError{
		Pos:   info.Fset.Position(pos),
		Type:  typ.String(),
		Field: identName(field),
	})
}

func identName(n ast.Node) string {
	switch n := n.(type) {
	case *ast.Ident:
		return n.String()
	case *ast.ParametrizedIdent:
		return n.Ident.String()
	default:
		panic(fmt.Sprintf("unexpected type %t", n))
	}
}

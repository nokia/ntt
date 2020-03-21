package ast

import (
	"github.com/nokia/ntt/internal/errors"
	"github.com/nokia/ntt/internal/loc"
	"github.com/nokia/ntt/internal/ttcn3/token"
)

type visitor struct {
	name   Node
	arrays []Expr
	value  Expr

	errs *errors.ErrorList
	fset *loc.FileSet
}

func (v *visitor) Visit(n Node) Visitor {
	if n == nil {
		return v
	}

	switch n := n.(type) {
	case *BinaryExpr:
		if n.Op.Kind == token.ASSIGN {
			// RHS is the initialization value.
			v.value = n.Y

			// Check if LHS is okay.
			switch n := n.X.(type) {
			case *Ident, *ParametrizedIdent, *IndexExpr:
				Walk(v, n)
				return nil
			}
		}

	case *Ident, *ParametrizedIdent:
		v.name = n
		return nil

	case *IndexExpr:
		v.arrays = append(v.arrays, n.Index)
		Walk(v, n.X)
		return nil
	}

	v.errs.Add(v.fset.Position(n.Pos()), "unexpected token")
	return nil
}

func Declarators(decl *ValueDecl, fset *loc.FileSet, f func(decl Expr, name Node, arrays []Expr, value Expr)) error {
	var errs errors.ErrorList
	for _, d := range decl.Decls {
		v := visitor{
			fset: fset,
			errs: &errs,
		}
		Walk(&v, d)
		if v.name != nil {
			// Reverse the array slice
			for i, l := 0, len(v.arrays); i < l/2; i++ {
				v.arrays[i], v.arrays[l-1-i] = v.arrays[l-1-i], v.arrays[i]
			}
			f(d, v.name, v.arrays, v.value)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

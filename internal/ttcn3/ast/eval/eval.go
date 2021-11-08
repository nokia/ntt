package eval

import (
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/token"
	"github.com/nokia/ntt/runtime"
)

func Eval(n ast.Node, env *runtime.Env) runtime.Object {
	switch n := n.(type) {
	case *ast.ValueLiteral:
		return evalLiteral(n, env)
	case *ast.UnaryExpr:
		return evalUnary(n, env)
	}
	return nil
}

func evalLiteral(n *ast.ValueLiteral, env *runtime.Env) runtime.Object {
	switch n.Tok.Kind {
	case token.INT:
		return runtime.NewInt(n.Tok.Lit)
	case token.TRUE:
		return runtime.NewBool(true)
	case token.FALSE:
		return runtime.NewBool(false)
	}
	return nil
}

func evalUnary(n *ast.UnaryExpr, env *runtime.Env) runtime.Object {
	val := Eval(n.X, env)
	switch n.Op.Kind {
	case token.ADD:
		if _, ok := val.(runtime.Int); ok {
			return val
		}
	case token.SUB:
		if x, ok := val.(runtime.Int); ok {
			return runtime.Int{Int: x.Neg(x.Value())}
		}
	case token.NOT:
		if b, ok := val.(runtime.Bool); ok {
			return !b
		}
	}
	return nil
}

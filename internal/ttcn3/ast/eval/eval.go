package eval

import (
	"math/big"

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
	case *ast.BinaryExpr:
		return evalBinary(n, env)
	case *ast.ParenExpr:
		// can be template `x := (1,2,3)`, but also artihmetic expression: `1*(2+3)`.
		// For now, we assume it's an arithmetic expression, when there's only one child.
		if len(n.List) == 1 {
			return Eval(n.List[0], env)
		}
	case *ast.BlockStmt:
		var result runtime.Object
		for _, stmt := range n.Stmts {
			result = Eval(stmt, env)
		}
		return result

	case *ast.ExprStmt:
		return Eval(n.Expr, env)

	case *ast.IfStmt:
		val := Eval(n.Cond, env)
		if val == nil {
			break
		}
		b, ok := val.(runtime.Bool)
		if !ok {
			break
		}
		if b {
			return Eval(n.Then, env)
		} else {
			return Eval(n.Else, env)
		}

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

func evalBinary(n *ast.BinaryExpr, env *runtime.Env) runtime.Object {
	op := n.Op.Kind
	x := Eval(n.X, env)
	y := Eval(n.Y, env)

	switch op {
	case token.ADD:
		x, ok := x.(runtime.Int)
		if !ok {
			return nil
		}
		y, ok := y.(runtime.Int)
		if !ok {
			return nil
		}
		return runtime.Int{Int: new(big.Int).Add(x.Int, y.Int)}

	case token.SUB:
		x, ok := x.(runtime.Int)
		if !ok {
			return nil
		}
		y, ok := y.(runtime.Int)
		if !ok {
			return nil
		}
		return runtime.Int{Int: new(big.Int).Sub(x.Int, y.Int)}

	case token.MUL:
		x, ok := x.(runtime.Int)
		if !ok {
			return nil
		}
		y, ok := y.(runtime.Int)
		if !ok {
			return nil
		}
		return runtime.Int{Int: new(big.Int).Mul(x.Int, y.Int)}

	case token.DIV:
		x, ok := x.(runtime.Int)
		if !ok {
			return nil
		}
		y, ok := y.(runtime.Int)
		if !ok {
			return nil
		}
		return runtime.Int{Int: new(big.Int).Div(x.Int, y.Int)}

	case token.REM:
		x, ok := x.(runtime.Int)
		if !ok {
			return nil
		}
		y, ok := y.(runtime.Int)
		if !ok {
			return nil
		}
		return runtime.Int{Int: new(big.Int).Rem(x.Int, y.Int)}

	case token.MOD:
		x, ok := x.(runtime.Int)
		if !ok {
			return nil
		}
		y, ok := y.(runtime.Int)
		if !ok {
			return nil
		}
		return runtime.Int{Int: new(big.Int).Mod(x.Int, y.Int)}

	case token.LT:
		x, ok := x.(runtime.Int)
		if !ok {
			return nil
		}
		y, ok := y.(runtime.Int)
		if !ok {
			return nil
		}
		if x.Cmp(y.Int) < 0 {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)

	case token.LE:
		x, ok := x.(runtime.Int)
		if !ok {
			return nil
		}
		y, ok := y.(runtime.Int)
		if !ok {
			return nil
		}
		if x.Cmp(y.Int) <= 0 {
			return runtime.NewBool(true)
		}

	case token.GT:
		x, ok := x.(runtime.Int)
		if !ok {
			return nil
		}
		y, ok := y.(runtime.Int)
		if !ok {
			return nil
		}
		if x.Cmp(y.Int) > 0 {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)

	case token.GE:
		x, ok := x.(runtime.Int)
		if !ok {
			return nil
		}
		y, ok := y.(runtime.Int)
		if !ok {
			return nil
		}
		if x.Cmp(y.Int) >= 0 {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)

	case token.EQ:
		switch x := x.(type) {
		case runtime.Int:
			y, ok := y.(runtime.Int)
			if !ok {
				return nil
			}
			if x.Cmp(y.Int) == 0 {
				return runtime.NewBool(true)
			}
			return runtime.NewBool(false)

		case runtime.Bool:
			y, ok := y.(runtime.Bool)
			if !ok {
				return nil
			}
			if x == y {
				return runtime.NewBool(true)
			}
			return runtime.NewBool(false)
		}

	case token.NE:
		switch x := x.(type) {
		case runtime.Int:
			y, ok := y.(runtime.Int)
			if !ok {
				return nil
			}
			if x.Cmp(y.Int) != 0 {
				return runtime.NewBool(true)
			}
			return runtime.NewBool(false)

		case runtime.Bool:
			y, ok := y.(runtime.Bool)
			if !ok {
				return nil
			}
			if x != y {
				return runtime.NewBool(true)
			}
			return runtime.NewBool(false)
		}
	}
	return nil
}

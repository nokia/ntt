package eval

import (
	"math/big"

	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/token"
	"github.com/nokia/ntt/runtime"
)

func Eval(n ast.Node, env *runtime.Env) runtime.Object {
	return unwrap(eval(n, env))
}

func eval(n ast.Node, env *runtime.Env) runtime.Object {
	switch n := n.(type) {
	case *ast.ModuleDef:
		return eval(n.Def, env)

	case *ast.ValueDecl:
		var result runtime.Object
		for _, decl := range n.Decls {
			result = eval(decl, env)
			if runtime.IsError(result) {
				return result
			}
		}
		return result

	case *ast.Declarator:
		val := eval(n.Value, env)
		if runtime.IsError(val) {
			return val
		}
		env.Set(n.Name.String(), val)
		return nil

	case ast.NodeList:
		var result runtime.Object
		for _, stmt := range n {
			result = eval(stmt, env)
			if needBreak(result) {
				return result
			}
		}
		return result

	case *ast.Ident:
		name := n.String()
		if val, ok := env.Get(name); ok {
			return val
		}
		return runtime.Errorf("identifier not found: %s", name)

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
			return eval(n.List[0], env)
		}
	case *ast.BlockStmt:
		var result runtime.Object
		for _, stmt := range n.Stmts {
			result = eval(stmt, env)
			if needBreak(result) {
				return result
			}
		}
		return result

	case *ast.ExprStmt:
		return eval(n.Expr, env)

	case *ast.IfStmt:
		val := eval(n.Cond, env)
		if runtime.IsError(val) {
			return val
		}

		b, ok := val.(runtime.Bool)
		if !ok {
			return runtime.Errorf("boolean expression expected. Got %s (%s)", val.Type(), val.Inspect())
		}

		switch {
		case b == true:
			return eval(n.Then, env)

		case n.Else != nil:
			return eval(n.Else, env)

		default:
			return nil
		}

	case *ast.ReturnStmt:
		val := eval(n.Result, env)
		if runtime.IsError(val) {
			return val
		}
		return &runtime.ReturnValue{Value: val}

	case *ast.FuncDecl:
		f := &runtime.Function{
			Env:    env,
			Params: n.Params,
			Body:   n.Body,
		}
		env.Set(n.Name.String(), f)
		return nil
	}

	return runtime.Errorf("unknown syntax node type: %T (%+v)", n, n)
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
	return runtime.Errorf("unknown literal kind %q (%s)", n.Tok.Kind, n.Tok.Lit)
}

func evalUnary(n *ast.UnaryExpr, env *runtime.Env) runtime.Object {
	val := eval(n.X, env)
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

	return runtime.Errorf("unknown operator: %s%s", n.Op.Kind, val.Inspect())
}

func evalBinary(n *ast.BinaryExpr, env *runtime.Env) runtime.Object {
	op := n.Op.Kind
	x := eval(n.X, env)
	y := eval(n.Y, env)

	switch {
	case x.Type() == runtime.INTEGER && y.Type() == runtime.INTEGER:
		return evalArithBinary(x.(runtime.Int), y.(runtime.Int), op, env)
	case x.Type() == runtime.BOOL && y.Type() == runtime.BOOL:
		return evalBoolBinary(bool(x.(runtime.Bool)), bool(y.(runtime.Bool)), op, env)
	case x.Type() != y.Type():
		return runtime.Errorf("type mismatch: %s %s %s", x.Type(), op, y.Type())
	}

	return runtime.Errorf("unknown operator: %s %s %s", x.Inspect(), op, y.Inspect())
}

func evalBoolBinary(x bool, y bool, op token.Kind, env *runtime.Env) runtime.Object {
	switch op {
	case token.EQ:
		return runtime.NewBool(x == y)
	case token.NE:
		return runtime.NewBool(x != y)
	case token.AND:
		return runtime.NewBool(x && y)
	case token.OR:
		return runtime.NewBool(x || y)
	case token.XOR:
		return runtime.NewBool(x && !y || !x && y)
	}

	return runtime.Errorf("unknown operator: boolean %s boolean", op)
}

func evalArithBinary(x runtime.Int, y runtime.Int, op token.Kind, env *runtime.Env) runtime.Object {
	switch op {
	case token.ADD:
		return runtime.Int{Int: new(big.Int).Add(x.Int, y.Int)}

	case token.SUB:
		return runtime.Int{Int: new(big.Int).Sub(x.Int, y.Int)}

	case token.MUL:
		return runtime.Int{Int: new(big.Int).Mul(x.Int, y.Int)}

	case token.DIV:
		return runtime.Int{Int: new(big.Int).Div(x.Int, y.Int)}

	case token.REM:
		return runtime.Int{Int: new(big.Int).Rem(x.Int, y.Int)}

	case token.MOD:
		return runtime.Int{Int: new(big.Int).Mod(x.Int, y.Int)}

	case token.LT:
		if x.Cmp(y.Int) < 0 {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)

	case token.LE:
		if x.Cmp(y.Int) <= 0 {
			return runtime.NewBool(true)
		}

	case token.GT:
		if x.Cmp(y.Int) > 0 {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)

	case token.GE:
		if x.Cmp(y.Int) >= 0 {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)

	case token.EQ:
		if x.Cmp(y.Int) == 0 {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)

	case token.NE:
		if x.Cmp(y.Int) != 0 {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)
	}
	return runtime.Errorf("unknown operator: integer %s integer", op)
}

func needBreak(v interface{}) bool {
	switch v.(type) {
	case *runtime.ReturnValue:
		return true
	case *runtime.Error:
		return true
	default:
		return false
	}
}

func unwrap(obj runtime.Object) runtime.Object {
	if ret, ok := obj.(*runtime.ReturnValue); ok {
		return ret.Value
	}
	return obj
}

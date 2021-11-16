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

	case *ast.DeclStmt:
		return eval(n.Decl, env)

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
		var val runtime.Object = runtime.Undefined
		if n.Value != nil {
			val = eval(n.Value, env)
			if runtime.IsError(val) {
				return val
			}
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
		if builtin, ok := runtime.Builtins[name]; ok {
			return builtin
		}
		return runtime.Errorf("identifier not found: %s", name)

	case *ast.CompositeLiteral:
		objs := evalExprList(n.List, env)
		if len(objs) == 1 && runtime.IsError(objs[0]) {
			return objs[0]
		}

		return &runtime.List{Elements: objs}

	case *ast.ValueLiteral:
		return evalLiteral(n, env)

	case *ast.UnaryExpr:
		return evalUnary(n, env)

	case *ast.BinaryExpr:
		return evalBinary(n, env)

	case *ast.IndexExpr:
		left := eval(n.X, env)
		if runtime.IsError(left) {
			return left
		}
		index := eval(n.Index, env)
		if runtime.IsError(index) {
			return index
		}
		switch {
		case left.Type() == runtime.LIST && index.Type() == runtime.INTEGER:
			list := left.(*runtime.List)
			i := index.(runtime.Int).Int64()
			if i < 0 || i >= int64(len(list.Elements)) {
				return runtime.Undefined
			}
			return list.Elements[i]
		}
		return runtime.Errorf("index operator not supported: %s", left.Type())

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
		if n, ok := n.Expr.(*ast.BinaryExpr); ok && n.Op.Kind == token.ASSIGN {
			return evalAssign(n.X, n.Y, env)
		}
		return eval(n.Expr, env)

	case *ast.IfStmt:
		b, err := evalBoolExpr(n.Cond, env)
		if runtime.IsError(err) {
			return err
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

	case *ast.CallExpr:
		f := eval(n.Fun, env)
		if runtime.IsError(f) {
			return f
		}

		args := evalExprList(n.Args.List, env)
		if len(args) == 1 && runtime.IsError(args[0]) {
			return args[0]
		}

		return apply(f, args)

	case *ast.WhileStmt:
		var result runtime.Object
		for {
			cond, err := evalBoolExpr(n.Cond, env)
			if runtime.IsError(err) {
				return err
			}
			if cond == false {
				break
			}
			result = eval(n.Body, env)
			if runtime.IsError(result) {
				break
			}
		}
		return result

	case *ast.DoWhileStmt:
		var result runtime.Object
		for {
			result = eval(n.Body, env)
			if runtime.IsError(result) {
				break
			}

			cond, err := evalBoolExpr(n.Cond, env)
			if runtime.IsError(err) {
				return err
			}
			if cond == false {
				break
			}
		}
		return result

	case *ast.ForStmt:
		if n.Init != nil {
			val := eval(n.Init, env)
			if runtime.IsError(val) {
				return val
			}
		}

		var result runtime.Object
		for {
			cond, err := evalBoolExpr(n.Cond, env)
			if runtime.IsError(err) {
				return err
			}
			if cond == false {
				break
			}

			result = eval(n.Body, env)
			if runtime.IsError(result) {
				break
			}

			result = eval(n.Post, env)
			if runtime.IsError(result) {
				break
			}

		}
		return result

	}

	return runtime.Errorf("unknown syntax node type: %T (%+v)", n, n)
}

func evalLiteral(n *ast.ValueLiteral, env *runtime.Env) runtime.Object {
	switch n.Tok.Kind {
	case token.INT:
		return runtime.NewInt(n.Tok.Lit)
	case token.FLOAT:
		return runtime.NewFloat(n.Tok.Lit)
	case token.TRUE:
		return runtime.NewBool(true)
	case token.FALSE:
		return runtime.NewBool(false)
	case token.NONE:
		return runtime.NoneVerdict
	case token.PASS:
		return runtime.PassVerdict
	case token.INCONC:
		return runtime.InconcVerdict
	case token.FAIL:
		return runtime.FailVerdict
	case token.ERROR:
		return runtime.ErrorVerdict
	case token.STRING:
		s, err := token.Unquote(n.Tok.Lit)
		if err != nil {
			return runtime.Errorf("%s", err.Error())
		}
		return &runtime.String{Value: s}
	case token.BSTRING:
		b, err := runtime.NewBitstring(n.Tok.Lit)
		if err != nil {
			return runtime.Errorf("%s", err.Error())
		}
		return b
	}
	return runtime.Errorf("unknown literal kind %q (%s)", n.Tok.Kind, n.Tok.Lit)
}

func evalUnary(n *ast.UnaryExpr, env *runtime.Env) runtime.Object {
	val := eval(n.X, env)
	if runtime.IsError(val) {
		return val
	}

	switch n.Op.Kind {
	case token.ADD:
		switch val := val.(type) {
		case runtime.Int:
			return val
		case runtime.Float:
			return val
		}
	case token.SUB:
		switch val := val.(type) {
		case runtime.Int:
			return runtime.Int{Int: val.Neg(val.Int)}
		case runtime.Float:
			return runtime.Float{Float: val.Neg(val.Float)}

		}
	case token.NOT:
		if b, ok := val.(runtime.Bool); ok {
			return !b
		}
	case token.NOT4B:
		if b, ok := val.(*runtime.Bitstring); ok {
			return &runtime.Bitstring{Value: new(big.Int).Abs(new(big.Int).Not(b.Value)), Unit: b.Unit}
		}
	}

	return runtime.Errorf("unknown operator: %s%s", n.Op.Kind, val.Inspect())
}

func evalBinary(n *ast.BinaryExpr, env *runtime.Env) runtime.Object {
	op := n.Op.Kind
	x := eval(n.X, env)
	if runtime.IsError(x) {
		return x
	}

	y := eval(n.Y, env)
	if runtime.IsError(y) {
		return y
	}

	switch {
	case x.Type() == runtime.INTEGER && y.Type() == runtime.INTEGER:
		return evalIntBinary(x.(runtime.Int), y.(runtime.Int), op, env)

	case x.Type() == runtime.FLOAT && y.Type() == runtime.FLOAT:
		return evalFloatBinary(x.(runtime.Float), y.(runtime.Float), op, env)

	case x.Type() == runtime.BOOL && y.Type() == runtime.BOOL:
		return evalBoolBinary(bool(x.(runtime.Bool)), bool(y.(runtime.Bool)), op, env)

	case x.Type() == runtime.STRING && y.Type() == runtime.STRING:
		return evalStringBinary(x.(*runtime.String).Value, y.(*runtime.String).Value, op, env)

	case x.Type() == runtime.BITSTRING && y.Type() == runtime.BITSTRING:
		return evalBitstringBinary(x.(*runtime.Bitstring), y.(*runtime.Bitstring), op, env)

	case x.Type() != y.Type():
		return runtime.Errorf("type mismatch: %s %s %s", x.Type(), op, y.Type())
	}

	return runtime.Errorf("unknown operator: %s %s %s", x.Inspect(), op, y.Inspect())
}

func evalIntBinary(x runtime.Int, y runtime.Int, op token.Kind, env *runtime.Env) runtime.Object {
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

func evalFloatBinary(x runtime.Float, y runtime.Float, op token.Kind, env *runtime.Env) runtime.Object {
	switch op {
	case token.ADD:
		return runtime.Float{Float: new(big.Float).Add(x.Float, y.Float)}

	case token.SUB:
		return runtime.Float{Float: new(big.Float).Sub(x.Float, y.Float)}

	case token.MUL:
		return runtime.Float{Float: new(big.Float).Mul(x.Float, y.Float)}

	case token.DIV:
		return runtime.Float{Float: new(big.Float).Quo(x.Float, y.Float)}

	case token.LT:
		if x.Cmp(y.Float) < 0 {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)

	case token.LE:
		if x.Cmp(y.Float) <= 0 {
			return runtime.NewBool(true)
		}

	case token.GT:
		if x.Cmp(y.Float) > 0 {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)

	case token.GE:
		if x.Cmp(y.Float) >= 0 {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)

	case token.EQ:
		if x.Cmp(y.Float) == 0 {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)

	case token.NE:
		if x.Cmp(y.Float) != 0 {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)
	}
	return runtime.Errorf("unknown operator: float %s float", op)
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

func evalStringBinary(x string, y string, op token.Kind, env *runtime.Env) runtime.Object {
	if op == token.CONCAT {
		return &runtime.String{Value: x + y}
	}
	return runtime.Errorf("unknown operator: charstring %s charstring", op)

}

func evalBitstringBinary(x *runtime.Bitstring, y *runtime.Bitstring, op token.Kind, env *runtime.Env) runtime.Object {
	switch op {
	case token.AND4B:
		return &runtime.Bitstring{Value: new(big.Int).And(x.Value, y.Value), Unit: x.Unit}

	case token.OR4B:
		return &runtime.Bitstring{Value: new(big.Int).Or(x.Value, y.Value), Unit: x.Unit}

	case token.XOR4B:
		return &runtime.Bitstring{Value: new(big.Int).Xor(x.Value, y.Value), Unit: x.Unit}

	case token.EQ:
		if x.Value.Cmp(y.Value) == 0 {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)

	case token.NE:
		if x.Value.Cmp(y.Value) != 0 {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)
	}
	return runtime.Errorf("unknown operator: bitstring %s bitstring", op)
}

func evalBoolExpr(n ast.Expr, env *runtime.Env) (bool, runtime.Object) {
	val := eval(n, env)
	if runtime.IsError(val) {
		return false, val
	}

	if b, ok := val.(runtime.Bool); ok {
		return b == true, nil
	}

	return false, runtime.Errorf("boolean expression expected. Got %s (%s)", val.Type(), val.Inspect())

}

func evalExprList(exprs []ast.Expr, env *runtime.Env) []runtime.Object {
	var result []runtime.Object
	for _, e := range exprs {
		val := eval(e, env)
		if runtime.IsError(val) {
			return []runtime.Object{val}
		}
		result = append(result, val)
	}
	return result
}

func evalAssign(lhs ast.Expr, rhs ast.Expr, env *runtime.Env) runtime.Object {
	val := eval(rhs, env)
	if runtime.IsError(val) {
		return val
	}

	id, ok := lhs.(*ast.Ident)
	if !ok {
		return runtime.Errorf("expected an identifier. not supported: %T (%+v)", lhs, lhs)
	}

	if _, ok := env.Get(id.String()); ok {
		env.Set(id.String(), val)
		return nil
	}

	return runtime.Errorf("identifier not found: %s", id.String())
}

func apply(obj runtime.Object, args []runtime.Object) runtime.Object {
	switch fn := obj.(type) {
	case *runtime.Function:
		fenv := runtime.NewEnv(fn.Env)
		for i, param := range fn.Params.List {
			fenv.Set(param.Name.String(), args[i])
		}
		return unwrap(eval(fn.Body, fenv))

	case *runtime.Builtin:
		return fn.Fn(args...)

	default:
		return runtime.Errorf("not a function: %s (%s)", obj.Type(), obj.Inspect())
	}

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

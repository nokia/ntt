package interpreter

import (
	"math/big"

	_ "github.com/nokia/ntt/builtins"
	"github.com/nokia/ntt/runtime"
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

func Eval(n ast.Node, env runtime.Scope) runtime.Object {
	return unwrap(eval(n, env))
}

func eval(n ast.Node, env runtime.Scope) runtime.Object {
	if n == nil {
		return nil
	}

	switch n := n.(type) {
	case *ast.Module:
		for _, d := range n.Defs {
			if ret := eval(d, env); runtime.IsError(ret) {
				return ret
			}
		}
		return nil

	case *ast.GroupDecl:
		for _, d := range n.Defs {
			if ret := eval(d, env); runtime.IsError(ret) {
				return ret
			}
		}
		return nil

	case *ast.ModuleDef:
		return eval(n.Def, env)

	case *ast.ControlPart:
		return eval(n.Body, env)

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

	case *ast.NodeList:
		var result runtime.Object
		for _, stmt := range n.Nodes {
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

	case *ast.CompositeLiteral:
		return evalComposite(n, env)

	case *ast.ValueLiteral:
		return evalLiteral(n, env)

	case *ast.UnaryExpr:
		return evalUnary(n, env)

	case *ast.BinaryExpr:
		return evalBinary(n, env)

	case *ast.SelectorExpr:
		left := eval(n.X, env)
		if runtime.IsError(left) {
			return left
		}

		env, ok := left.(runtime.Scope)
		if !ok {
			return runtime.Errorf("%s is not allowed for %s", ".", left.Type())
		}

		return eval(n.Sel, env)

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
		case left.Type() == runtime.MAP:
			m := left.(*runtime.Map)
			val, _ := m.Get(index)
			return val
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
		for {
			cond, err := evalBoolExpr(n.Cond, env)
			if runtime.IsError(err) {
				return err
			}
			if cond == false {
				return nil
			}

			result := eval(n.Body, env)
			switch {
			case runtime.IsError(result):
				return result
			case result == runtime.Break:
				return nil
			}
		}

	case *ast.DoWhileStmt:
		for {
			result := eval(n.Body, env)
			switch {
			case runtime.IsError(result):
				return result
			case result == runtime.Break:
				return nil
			}

			cond, err := evalBoolExpr(n.Cond, env)
			if runtime.IsError(err) {
				return err
			}
			if cond == false {
				return nil
			}
		}

	case *ast.ForStmt:
		if n.Init != nil {
			val := eval(n.Init, env)
			if runtime.IsError(val) {
				return val
			}
		}

		for {
			cond, err := evalBoolExpr(n.Cond, env)
			if runtime.IsError(err) {
				return err
			}
			if cond == false {
				return nil
			}

			result := eval(n.Body, env)
			switch {
			case runtime.IsError(result):
				return result
			case result == runtime.Break:
				return nil
			}

			result = eval(n.Post, env)
			if runtime.IsError(result) {
				return result
			}

		}

	case *ast.BranchStmt:
		switch n.Tok.Kind {
		case token.BREAK:
			return runtime.Break
		case token.CONTINUE:
			return runtime.Continue
		case token.LABEL:
			return nil
		case token.GOTO:
			return runtime.Errorf("goto statement not implemented")
		}
	}

	return runtime.Errorf("unknown syntax node type: %T (%+v)", n, n)
}

func evalLiteral(n *ast.ValueLiteral, env runtime.Scope) runtime.Object {
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
		return &runtime.String{Value: []rune(s)}
	case token.BSTRING:
		b, err := runtime.NewBitstring(n.Tok.Lit)
		if err != nil {
			return runtime.Errorf("%s", err.Error())
		}
		return b
	case token.MUL:
		return runtime.AnyOrNone
	case token.ANY:
		return runtime.Any

	}
	return runtime.Errorf("unknown literal kind %q (%s)", n.Tok.Kind, n.Tok.Lit)
}

func evalComposite(n *ast.CompositeLiteral, env runtime.Scope) runtime.Object {
	// An empty composite literal will evaluate as List.
	if len(n.List) == 0 {
		return evalValueList(n.List, env)
	}

	// The first element tells us, if we expect a value list or an assignment list.
	if first, ok := n.List[0].(*ast.BinaryExpr); ok && first.Op.Kind == token.ASSIGN {
		if _, ok := first.X.(*ast.Ident); ok {
			return evalRecordAssignmentList(n.List, env)
		} else {
			return evalMapAssignmentList(n.List, env)
		}
	}

	return evalValueList(n.List, env)
}

func evalValueList(s []ast.Expr, env runtime.Scope) runtime.Object {
	objs := evalExprList(s, env)
	if len(objs) == 1 && runtime.IsError(objs[0]) {
		return objs[0]
	}
	return runtime.NewList(objs...)
}

func evalMapAssignmentList(exprs []ast.Expr, env runtime.Scope) runtime.Object {
	m := runtime.NewMap()

	for _, expr := range exprs {

		n, ok := expr.(*ast.BinaryExpr)
		if !ok || n.Op.Kind != token.ASSIGN {
			return runtime.Errorf("missing key/value. got=%T", n)
		}

		val := eval(n.Y, env)
		if runtime.IsError(val) {
			return val
		}

		key := evalKeyExpr(n.X, env)
		if runtime.IsError(key) {
			return key
		}

		if ret := m.Set(key, val); runtime.IsError(ret) {
			return ret
		}

	}
	return m
}

func evalKeyExpr(n ast.Expr, env runtime.Scope) runtime.Object {
	switch n := n.(type) {
	case *ast.IndexExpr:
		if n.X == nil {
			return eval(n.Index, env)
		}
	}
	return runtime.Errorf("syntax error. Expecting a key expression. got=%T", n)
}

func evalRecordAssignmentList(exprs []ast.Expr, env runtime.Scope) runtime.Object {
	r := runtime.NewRecord()

	for _, expr := range exprs {

		n, ok := expr.(*ast.BinaryExpr)
		if !ok || n.Op.Kind != token.ASSIGN {
			return runtime.Errorf("missing key/value. got=%T", n)
		}

		val := eval(n.Y, env)
		if runtime.IsError(val) {
			return val
		}

		if ret := r.Set(n.X.(*ast.Ident).String(), val); runtime.IsError(ret) {
			return ret
		}

	}
	return r
}

func evalUnary(n *ast.UnaryExpr, env runtime.Scope) runtime.Object {
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
			return -val

		}
	case token.NOT:
		if b, ok := val.(runtime.Bool); ok {
			return !b
		}
	case token.NOT4B:
		if b, ok := val.(*runtime.Bitstring); ok {
			return &runtime.Bitstring{Value: new(big.Int).Abs(new(big.Int).Not(b.Value)), Unit: b.Unit, String: runtime.BigIntToBitstring(new(big.Int).Abs(new(big.Int).Not(b.Value)), b.Unit)}
		}
	}

	return runtime.Errorf("unknown operator: %s%s", n.Op.Kind, val.Inspect())
}

func evalBinary(n *ast.BinaryExpr, env runtime.Scope) runtime.Object {
	op := n.Op.Kind
	x := eval(n.X, env)
	if runtime.IsError(x) {
		return x
	}

	y := eval(n.Y, env)
	if runtime.IsError(y) {
		return y
	}

	if x.Type() != y.Type() {
		return runtime.Errorf("type mismatch: %s %s %s", x.Type(), op, y.Type())
	}

	switch {
	case op == token.EQ:
		return runtime.NewBool(x.Equal(y))

	case op == token.NE:
		return runtime.NewBool(!x.Equal(y))

	case x.Type() == runtime.INTEGER:
		return evalIntBinary(x.(runtime.Int), y.(runtime.Int), op, env)

	case x.Type() == runtime.FLOAT:
		return evalFloatBinary(x.(runtime.Float), y.(runtime.Float), op, env)

	case x.Type() == runtime.BOOL:
		return evalBoolBinary(bool(x.(runtime.Bool)), bool(y.(runtime.Bool)), op, env)

	case x.Type() == runtime.STRING:
		return evalStringBinary(string(x.(*runtime.String).Value), string(y.(*runtime.String).Value), op, env)

	case x.Type() == runtime.BITSTRING:
		return evalBitstringBinary(x.(*runtime.Bitstring), y.(*runtime.Bitstring), op, env)
	}

	return runtime.Errorf("unknown operator: %s %s %s", x.Inspect(), op, y.Inspect())
}

func evalIntBinary(x runtime.Int, y runtime.Int, op token.Kind, env runtime.Scope) runtime.Object {
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
	}
	return runtime.Errorf("unknown operator: integer %s integer", op)
}

func evalFloatBinary(x runtime.Float, y runtime.Float, op token.Kind, env runtime.Scope) runtime.Object {
	switch op {
	case token.ADD:
		return runtime.Float(x + y)

	case token.SUB:
		return runtime.Float(x - y)

	case token.MUL:
		return runtime.Float(x * y)

	case token.DIV:
		return runtime.Float(x / y)

	case token.LT:
		if x < y {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)

	case token.LE:
		if x <= y {
			return runtime.NewBool(true)
		}

	case token.GT:
		if x > y {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)

	case token.GE:
		if x >= y {
			return runtime.NewBool(true)
		}
		return runtime.NewBool(false)
	}
	return runtime.Errorf("unknown operator: float %s float", op)
}

func evalBoolBinary(x bool, y bool, op token.Kind, env runtime.Scope) runtime.Object {
	switch op {
	case token.AND:
		return runtime.NewBool(x && y)
	case token.OR:
		return runtime.NewBool(x || y)
	case token.XOR:
		return runtime.NewBool(x && !y || !x && y)
	}

	return runtime.Errorf("unknown operator: boolean %s boolean", op)
}

func evalStringBinary(x string, y string, op token.Kind, env runtime.Scope) runtime.Object {
	if op == token.CONCAT {
		return &runtime.String{Value: []rune(string(x) + string(y))}
	}
	return runtime.Errorf("unknown operator: charstring %s charstring", op)

}

func evalBitstringBinary(x *runtime.Bitstring, y *runtime.Bitstring, op token.Kind, env runtime.Scope) runtime.Object {
	switch op {
	case token.AND4B:
		return &runtime.Bitstring{Value: new(big.Int).And(x.Value, y.Value), Unit: x.Unit, String: runtime.BigIntToBitstring(new(big.Int).And(x.Value, y.Value), x.Unit)}

	case token.OR4B:
		return &runtime.Bitstring{Value: new(big.Int).Or(x.Value, y.Value), Unit: x.Unit, String: runtime.BigIntToBitstring(new(big.Int).Or(x.Value, y.Value), x.Unit)}

	case token.XOR4B:
		return &runtime.Bitstring{Value: new(big.Int).Xor(x.Value, y.Value), Unit: x.Unit, String: runtime.BigIntToBitstring(new(big.Int).Xor(x.Value, y.Value), x.Unit)}

	}
	return runtime.Errorf("unknown operator: bitstring %s bitstring", op)
}

func evalBoolExpr(n ast.Expr, env runtime.Scope) (bool, runtime.Object) {
	val := eval(n, env)
	if runtime.IsError(val) {
		return false, val
	}

	if b, ok := val.(runtime.Bool); ok {
		return b == true, nil
	}

	return false, runtime.Errorf("boolean expression expected. Got %s (%s)", val.Type(), val.Inspect())

}

func evalExprList(exprs []ast.Expr, env runtime.Scope) []runtime.Object {
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

func evalAssign(lhs ast.Expr, rhs ast.Expr, env runtime.Scope) runtime.Object {
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
		return v == runtime.Break || v == runtime.Continue
	}
}

func unwrap(obj runtime.Object) runtime.Object {
	if ret, ok := obj.(*runtime.ReturnValue); ok {
		return ret.Value
	}

	if obj == runtime.Break || obj == runtime.Continue {
		return runtime.Errorf("break or continue statements not allowed outside loops")
	}
	return obj
}

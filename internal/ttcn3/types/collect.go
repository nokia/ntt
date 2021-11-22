package types

import (
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/token"
)

func (info *Info) CollectInfo(n ast.Node) {
	if info.Types == nil {
		info.Types = make(map[ast.Expr]Type)
	}
	ast.Apply(n, info.enter, info.exit)
}

func (info *Info) enter(c *ast.Cursor) bool {
	return true
}

func (info *Info) exit(c *ast.Cursor) bool {
	switch n := c.Node().(type) {
	case *ast.ValueLiteral:
		info.collectLiteral(n)
	case *ast.UnaryExpr:
		info.collectUnaryExpr(n)
	}

	return true
}

func (info *Info) collectLiteral(n *ast.ValueLiteral) {
	switch n.Tok.Kind {
	case token.INT:
		info.Types[n] = Typ[Integer]

	case token.FLOAT:
		info.Types[n] = Typ[Float]

	case token.NAN:
		info.Types[n] = Typ[Float]

	case token.ANY, token.MUL:
		info.Types[n] = Typ[Template]

	case token.NULL:
		info.Types[n] = Typ[Component]

	case token.OMIT:
		info.Types[n] = Typ[Omit]

	case token.FALSE, token.TRUE:
		info.Types[n] = Typ[Boolean]

	case token.NONE, token.INCONC, token.PASS, token.FAIL, token.ERROR:
		info.Types[n] = Typ[Verdict]

	default:
		panic("unhandled literal")
	}
}

func (info *Info) collectUnaryExpr(n *ast.UnaryExpr) {
	switch n.Op.Kind {
	case token.NOT:
		info.assertType(n.X, Typ[Boolean])

	case token.ADD, token.SUB:
		info.assertType(n.X, Typ[Integer])
	}

	info.Types[n] = info.Types[n.X]
}

func (info *Info) assertType(n ast.Expr, expected Type) {
	if actual := info.Types[n]; actual != expected {
		info.invalidTypeError(n, actual, expected)
	}
}

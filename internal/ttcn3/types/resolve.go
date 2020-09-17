package types

import (
	"github.com/nokia/ntt/internal/ttcn3/ast"
	"github.com/nokia/ntt/internal/ttcn3/token"
)

// Resolve
func (info *Info) Resolve() {
	if info.Types == nil {
		info.Types = make(map[ast.Expr]Type)
	}
	for _, mod := range info.Modules {
		info.resolve(mod.node)
	}
}

func (info *Info) resolve(n ast.Node) {
	ast.Apply(n, nil, info.resolveExit)
}

func (info *Info) resolveExit(c *ast.Cursor) bool {
	switch n := c.Node().(type) {
	case *ast.ValueLiteral:
		typeOf := func(tok token.Kind) Type {
			switch tok {
			case token.INT:
				return Typ[Integer]

			case token.FLOAT:
				return Typ[Float]

			case token.NAN:
				return Typ[Float]

			case token.ANY, token.MUL:
				return Typ[Template]

			case token.NULL:
				return Typ[Component]

			case token.OMIT:
				return Typ[Omit]

			case token.FALSE, token.TRUE:
				return Typ[Boolean]

			case token.NONE, token.INCONC, token.PASS, token.FAIL, token.ERROR:
				return Typ[Verdict]

			case token.BSTRING:
				// TODO(5nord) Implement hexstring, octetstring, ...
				return Typ[Bitstring]

			case token.STRING:
				// TODO(5nord) Implement universal charstring
				return Typ[Charstring]

			default:
				return Typ[Invalid]
			}
		}
		info.Types[n] = typeOf(n.Tok.Kind)

	case *ast.Ident:
		scp := info.Scopes[n]

		// Identifier which to not have a scope are part of declarations and can
		// be skipped.
		if scp == nil {
			break
		}

		def := scp.Lookup(n.String())
		if def == nil {
			info.unknownIdentifierError(n)
			break
		}

		// In local scopes, check if declaration comes after
		if _, ok := def.Parent().(*LocalScope); ok {
			if def.End() >= n.Pos() {
				info.unknownIdentifierError(n)
			}
		}
		info.Types[n] = def.Type()

	case *ast.ParametrizedIdent:
		info.Types[n] = info.Types[n.Ident]

	case *ast.UnaryExpr:
		info.Types[n] = info.Types[n.X]

	case *ast.BinaryExpr:
		switch n.Op.Kind {
		case token.ASSIGN:
			info.Types[n] = info.Types[n.X]
		case token.COLON:
			info.Types[n] = info.Types[n.X]
		case token.RANGE:
			info.Types[n] = Typ[Integer]
		case token.OR:
			info.Types[n] = Typ[Boolean]
		case token.XOR:
			info.Types[n] = Typ[Boolean]
		case token.AND:
			info.Types[n] = Typ[Boolean]
		case token.NOT:
			info.Types[n] = Typ[Boolean]
		case token.EQ, token.NE:
			info.Types[n] = Typ[Boolean]
		case token.LT, token.LE, token.GT, token.GE:
			info.Types[n] = Typ[Boolean]
		case token.SHR, token.SHL, token.ROR, token.ROL:
			info.Types[n] = Typ[String]
		case token.OR4B:
			info.Types[n] = Typ[String]
		case token.XOR4B:
			info.Types[n] = Typ[String]
		case token.AND4B:
			info.Types[n] = Typ[String]
		case token.NOT4B:
			info.Types[n] = Typ[String]
		case token.CONCAT:
			info.Types[n] = Typ[String]
		case token.ADD, token.SUB:
			info.Types[n] = Typ[Numerical]
		case token.MUL, token.DIV, token.REM, token.MOD:
			info.Types[n] = Typ[Numerical]
		}

	case *ast.SelectorExpr:
		typ := info.Types[n.X]
		if typ == nil {
			return false
		}

		id := identName(n.Sel)

		scp, ok := typ.(Scope)
		if !ok {
			info.noFieldError(typ, n.Sel, n.Pos())
			return false
		}

		obj := scp.Lookup(id)
		if !ok {
			info.noFieldError(typ, n.Sel, n.Pos())
			return false
		}

		info.Types[n] = obj.Type()
	}
	return true
}

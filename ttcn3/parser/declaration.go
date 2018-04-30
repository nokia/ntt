package parser

import (
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

func (p *parser) parseDecl() ast.Decl {
	switch p.tok {
	case token.TEMPLATE:
		return p.parseTemplateDecl()
	case token.VAR, token.CONST, token.MODULEPAR,
		token.TIMER, token.PORT:
		return p.parseValueDecl()
	case token.FUNCTION, token.TESTCASE, token.ALTSTEP:
		return p.parseFuncDecl()
	default:
		p.errorExpected(p.pos, "declaration")
	}
	return nil
}

func (p *parser) parseTemplateDecl() *ast.ValueDecl {
	if p.trace {
		defer un(trace(p, "TemplateDecl"))
	}

	x := &ast.ValueDecl{DeclPos: p.pos, Kind: p.tok}
	p.next()

	if p.tok == token.LPAREN {
		p.next() // consume '('
		p.next() // consume omit/value/...
		p.expect(token.RPAREN)
	}

	if p.tok == token.MODIF {
		p.next()
	}

	x.Type = p.parseTypeRef()
	p.parseIdent()
	if p.tok == token.LPAREN {
		p.parseParameters()
	}
	if p.tok == token.MODIFIES {
		p.next()
		p.parseIdent()
	}
	p.expect(token.ASSIGN)
	p.parseExpr()

	p.expectSemi()
	return x
}

func (p *parser) parseValueDecl() *ast.ValueDecl {
	if p.trace {
		defer un(trace(p, "ValueDecl"))
	}

	x := &ast.ValueDecl{DeclPos: p.pos, Kind: p.tok}
	p.next()
	p.parseRestrictionSpec()

	if p.tok == token.MODIF {
		p.next()
	}

	if x.Kind != token.TIMER {
		x.Type = p.parseTypeRef()
	}
	x.Decls = p.parseExprList()

	p.expectSemi()
	return x
}

func (p *parser) parseRestrictionSpec() *ast.RestrictionSpec {
	switch p.tok {
	case token.TEMPLATE:
		x := &ast.RestrictionSpec{Kind: p.tok, KindPos: p.pos}
		p.next()
		if p.tok != token.LPAREN {
			return x
		}

		p.next()
		x.Kind = p.tok
		x.KindPos = p.pos
		p.next()
		p.expect(token.RPAREN)

	case token.OMIT, token.VALUE, token.PRESENT:
		x := &ast.RestrictionSpec{Kind: p.tok, KindPos: p.pos}
		p.next()
		return x
	}
	return nil
}

func (p *parser) parseFuncDecl() *ast.FuncDecl {
	if p.trace {
		defer un(trace(p, "FuncDecl"))
	}

	x := &ast.FuncDecl{FuncPos: p.pos, Kind: p.tok}
	p.next()
	x.Name = p.parseIdent()

	if p.tok == token.MODIF {
		p.next()
	}

	x.Params = p.parseParameters()
	if p.tok == token.RUNS {
		p.next()
		p.expect(token.ON)
		x.RunsOn = p.parseTypeRef()
	}
	if p.tok == token.MTC {
		p.next()
		x.Mtc = p.parseTypeRef()
	}
	if p.tok == token.SYSTEM {
		p.next()
		x.System = p.parseTypeRef()
	}
	if p.tok == token.RETURN {
		x.Return = p.parseReturn()
	}
	if p.tok == token.LBRACE {
		x.Body = p.parseBlockStmt()
	}

	p.expectSemi()
	return x
}

func (p *parser) parseReturn() ast.Expr {
	p.next()
	p.parseRestrictionSpec()
	if p.tok == token.MODIF {
		p.next()
	}
	return p.parseTypeRef()
}

func (p *parser) parseParameters() *ast.FieldList {
	x := &ast.FieldList{From: p.pos}
	p.expect(token.LPAREN)
	for p.tok != token.RPAREN {
		x.Fields = append(x.Fields, p.parseParameter())
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
	p.expect(token.RPAREN)
	return x
}

func (p *parser) parseParameter() *ast.Field {
	x := &ast.Field{}

	switch p.tok {
	case token.IN:
		p.next()
	case token.OUT:
		p.next()
	case token.INOUT:
		p.next()
	}

	p.parseRestrictionSpec()
	if p.tok == token.MODIF {
		p.next()
	}
	x.Type = p.parseTypeRef()
	x.Name = p.parseExpr()

	return x
}

func (p *parser) parseRefList() {
	for {
		p.parseTypeRef()
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
}

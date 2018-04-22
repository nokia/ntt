package parser

import (
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

func (p *parser) parseDecl() ast.Decl {
	switch p.tok {
	case token.VAR, token.CONST, token.MODULEPAR:
		return p.parseValueDecl()
	case token.FUNCTION, token.TESTCASE, token.ALTSTEP:
		return p.parseFuncDecl()
	}
	return nil
}

func (p *parser) parseValueDecl() *ast.ValueDecl {
	if p.trace {
		defer un(trace(p, "ValueDecl"))
	}

	x := &ast.ValueDecl{DeclPos: p.pos, Kind: p.tok}
	p.next()
	x.Type = p.parseIdent()
	x.Decls = p.parseExprList()

	p.expectSemi()
	return x
}

func (p *parser) parseFuncDecl() *ast.FuncDecl {
	if p.trace {
		defer un(trace(p, "FuncDecl"))
	}

	x := &ast.FuncDecl{FuncPos: p.pos, Kind: p.tok}
	p.next()
	x.Name = p.parseIdent()
	x.Params = p.parseParameters()
	if p.tok == token.RUNS {
		p.expect(token.ON)
		x.RunsOn = p.parseTypeRef()
	}
	if p.tok == token.MTC {
		x.Mtc = p.parseTypeRef()
	}
	if p.tok == token.SYSTEM {
		x.System = p.parseTypeRef()
	}
	if p.tok == token.RETURN {
		x.Return = p.parseTypeRef()
	}
	if p.tok == token.LBRACE {
		x.Body = p.parseBlockStmt()
	}

	p.expectSemi()
	return x
}

func (p *parser) parseParameters() *ast.FieldList {
	x := &ast.FieldList{From: p.pos}
	p.expect(token.LPAREN)
	for p.tok != token.EOF && p.tok != token.RPAREN {
		x.Fields = append(x.Fields, p.parseParameter())
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

	x.Type = p.parseTypeRef()
	x.Name = p.parseExpr()

	return x
}

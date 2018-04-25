package parser

import (
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

func (p *parser) parseType() ast.Decl {
	if p.trace {
		defer un(trace(p, "Type"))
	}
	pos := p.pos
	p.next()
	switch p.tok {
	case token.IDENT:
		x := p.parseSubType()
		x.TypePos = pos
		p.expectSemi()
		return x
	}
	return nil
}

func (p *parser) parseSubType() *ast.SubType {
	if p.trace {
		defer un(trace(p, "SubType"))
	}
	return &ast.SubType{
		Type: p.parseTypeRef(),
		Name: p.parseIdent(),
	}
}

func (p *parser) parseTypeRef() ast.Expr {
	if p.trace {
		defer un(trace(p, "TypeRef"))
	}
	x := p.parsePrimaryExpr()
	return x
}

package parser

import (
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

func (p *parser) parseType() ast.Expr {
	pos := p.pos
	p.next()
	switch p.tok {
	case token.IDENT:
		x := p.parseSubType()
		x.TypePos = pos
		return x
	}
	return nil
}

func (p *parser) parseSubType() *ast.SubType {
	x := &ast.SubType{}
	x.Type = p.parseTypeRef()
	x.Name = p.parseIdent()
	return x
}

func (p *parser) parseTypeRef() ast.Expr {
	x := p.parseIdent()
	p.resolve(x)
	return x
}

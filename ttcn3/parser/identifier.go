package parser

import (
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

func (p *parser) parseIdent() *ast.Ident {
	pos := p.pos
	name := "_"
	switch p.tok {
	case token.IDENT, token.ADDRESS:
		name = p.lit
		p.next()
	default:
		p.expect(token.IDENT) // use expect() error handling
	}
	return &ast.Ident{NamePos: pos, Name: name}
}

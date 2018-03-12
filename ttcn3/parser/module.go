package parser

import (
	"github.com/nokia/ntt/ttcn3/ast"
	"github.com/nokia/ntt/ttcn3/token"
)

func (p *parser) parseModule() *ast.Module {
	if p.trace {
		defer un(trace(p, "Module"))
	}

	pos := p.expect(token.MODULE)
	name := p.parseIdent()
	p.expect(token.LBRACE)

	var decls []ast.Decl
	for p.tok != token.RBRACE {
		decls = append(decls, p.parseModuleDef())
	}

	return &ast.Module{
		Module:   pos,
		Name:     name,
		Decls:    decls,
		Comments: p.comments,
	}
}

func (p *parser) parseModuleDef() ast.Decl {
	p.next()
	return nil
}

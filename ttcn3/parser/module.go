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
	p.expect(token.RBRACE)

	return &ast.Module{
		Module:   pos,
		Name:     name,
		Decls:    decls,
		Comments: p.comments,
	}
}

func (p *parser) parseModuleDef() ast.Decl {
	if p.trace {
		defer un(trace(p, "ModuleDef"))
	}
	p.next()
	return nil
}

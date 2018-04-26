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

	if p.tok == token.LANGUAGE {
		p.next()
		for {
			p.expect(token.STRING)
			if p.tok != token.COMMA {
				break
			}
			p.next()
		}
	}

	p.expect(token.LBRACE)

	p.openScope()

	var decls []ast.Decl
	for p.tok != token.RBRACE && p.tok != token.EOF {
		decls = append(decls, p.parseModuleDef())
	}
	p.expect(token.RBRACE)

	p.closeScope()

	return &ast.Module{
		Module:   pos,
		Name:     name,
		Decls:    decls,
		Comments: p.comments,
	}
}

func (p *parser) parseModuleDef() ast.Decl {
	switch p.tok {
	case token.IMPORT:
		return p.parseImport()
	case token.TYPE:
		return p.parseType()
	case token.VAR, token.CONST, token.MODULEPAR:
		return p.parseValueDecl()
	case token.FUNCTION, token.TESTCASE, token.ALTSTEP:
		return p.parseFuncDecl()
	default:
		p.errorExpected(p.pos, "module definition")
		p.next()
		return nil
	}
}

func (p *parser) parseImport() ast.Decl {
	if p.trace {
		defer un(trace(p, "Import"))
	}

	pos := p.pos
	p.next()
	if p.tok != token.FROM {
		p.errorExpected(p.pos, "from")
	}
	p.next()

	name := p.parseIdent()

	var specs []ast.ImportSpec
	if p.tok != token.ALL {
		p.errorExpected(p.pos, "all")
	}
	p.next()

	p.expectSemi()

	return &ast.ImportDecl{
		ImportPos:   pos,
		Module:      name,
		ImportSpecs: specs,
	}
}

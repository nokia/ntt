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

	// TODO: Language Specification

	p.expect(token.LBRACE)

	p.openScope()

	var decls []ast.Decl
	for p.tok != token.RBRACE && p.tok != token.EOF {
		var d ast.Decl
		switch p.tok {
		case token.IMPORT:
			d = p.parseImport()
		case token.TYPE:
			d = p.parseType()
		case token.VAR, token.CONST, token.MODULEPAR:
			d = p.parseValueDecl()
		case token.FUNCTION, token.TESTCASE, token.ALTSTEP:
			d = p.parseFuncDecl()

		/*
			TODO:
				* Module Definitions:
					- [ ] Templates
					- [ ] External Definitions

				* Package Management:
					- [ ] Import Statement
					- [ ] Visibility
					- [ ] Group Definitions
		*/

		default:
			p.errorExpected(p.pos, "module definition")
			p.next()
		}
		decls = append(decls, d)
	}
	p.expect(token.RBRACE)

	// TODO: With Statements

	p.closeScope()

	return &ast.Module{
		Module:   pos,
		Name:     name,
		Decls:    decls,
		Comments: p.comments,
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

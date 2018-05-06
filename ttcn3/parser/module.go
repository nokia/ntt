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
		p.parseLanguageSpec()
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

func (p *parser) parseLanguageSpec() {
	p.next()
	for {
		p.expect(token.STRING)
		if p.tok != token.COMMA {
			break
		}
		p.next()
	}
}

func (p *parser) parseModuleDef() ast.Decl {
	switch p.tok {
	case token.PRIVATE, token.PUBLIC, token.FRIEND:
		p.next()
	}

	switch p.tok {
	case token.IMPORT:
		p.parseImport()
	case token.GROUP:
		p.parseGroup()
	case token.TYPE:
		p.parseType()
	case token.TEMPLATE:
		p.parseTemplateDecl()
	case token.MODULEPAR:
		p.parseModulePar()
	case token.VAR, token.CONST:
		p.parseValueDecl()
	case token.SIGNATURE:
		p.parseSignatureDecl()
	case token.FUNCTION, token.TESTCASE, token.ALTSTEP:
		p.parseFuncDecl()
	case token.CONTROL:
		p.next()
		p.parseBlockStmt()
	default:
		p.errorExpected(p.pos, "module definition")
		p.next()
	}
	p.expectSemi()
	return nil
}

func (p *parser) parseImport() ast.Decl {
	if p.trace {
		defer un(trace(p, "Import"))
	}

	pos := p.pos
	p.next()
	p.expect(token.FROM)

	name := p.parseIdent()

	if p.tok == token.LANGUAGE {
		p.parseLanguageSpec()
	}

	var specs []ast.ImportSpec
	switch p.tok {
	case token.ALL:
		p.next()
		if p.tok == token.EXCEPT {
			p.parseExceptSpec()
		}
	case token.LBRACE:
		p.parseImportSpec()
	default:
		p.errorExpected(p.pos, "'all' or import spec")
	}

	p.parseWith()

	return &ast.ImportDecl{
		ImportPos:   pos,
		Module:      name,
		ImportSpecs: specs,
	}
}

func (p *parser) parseImportSpec() {
	p.expect(token.LBRACE)
	for p.tok != token.EOF && p.tok != token.RBRACE {
		p.parseImportStmt()
	}
	p.expect(token.RBRACE)
}

func (p *parser) parseImportStmt() {
	switch p.tok {
	case token.ALTSTEP, token.CONST, token.FUNCTION, token.MODULEPAR,
		token.SIGNATURE, token.TEMPLATE, token.TESTCASE, token.TYPE:
		p.next()
		if p.tok == token.ALL {
			p.next()
			if p.tok == token.EXCEPT {
				p.next()
				p.parseRefList()
			}
		} else {
			p.parseRefList()
		}
	case token.GROUP:
		p.next()
		for {
			p.parseTypeRef()
			if p.tok == token.EXCEPT {
				p.parseExceptSpec()
			}
			if p.tok != token.COMMA {
				break
			}
			p.next()
		}
	case token.IMPORT:
		p.next()
		p.expect(token.ALL)
	default:
		p.errorExpected(p.pos, "definition qualifier")
	}

	p.expectSemi()
}

func (p *parser) parseExceptSpec() {
	p.next()
	p.expect(token.LBRACE)
	for p.tok != token.EOF && p.tok != token.RBRACE {
		p.parseExceptStmt()
	}
	p.expect(token.RBRACE)
}

func (p *parser) parseExceptStmt() {
	switch p.tok {
	case token.ALTSTEP, token.CONST, token.FUNCTION, token.GROUP,
		token.IMPORT, token.MODULEPAR, token.SIGNATURE, token.TEMPLATE,
		token.TESTCASE, token.TYPE:
		p.next()
	default:
		p.errorExpected(p.pos, "definition qualifier")
	}

	if p.tok == token.ALL {
		p.next()
	} else {
		for {
			p.parseTypeRef()
			if p.tok != token.COMMA {
				break
			}
			p.next()
		}
	}
	p.expectSemi()
}

func (p *parser) parseGroup() {
	p.next()
	p.parseIdent()
	p.expect(token.LBRACE)

	var decls []ast.Decl
	for p.tok != token.RBRACE && p.tok != token.EOF {
		decls = append(decls, p.parseModuleDef())
	}
	p.expect(token.RBRACE)
	p.parseWith()
}
